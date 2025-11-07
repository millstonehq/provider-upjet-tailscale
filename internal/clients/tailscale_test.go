// Package clients contains the provider config setup.
package clients

import (
	"context"
	"strings"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/millstonehq/provider-upjet-tailscale/apis/v1beta1"
	tailnetkeyv1alpha1 "github.com/millstonehq/provider-upjet-tailscale/apis/tailnetkey/v1alpha1"
)

// newManagedWithProviderConfigRef returns a real managed resource from this provider
// with its ProviderConfigReference set. This avoids having to reimplement runtime
// interfaces that changed between v1 and v2.
func newManagedWithProviderConfigRef(name string) resource.Managed {
	mg := &tailnetkeyv1alpha1.Key{}
	mg.Spec.ResourceSpec = xpv1.ResourceSpec{
		ProviderConfigReference: &xpv1.Reference{Name: name},
	}
	return mg
}

func TestTerraformSetupBuilder(t *testing.T) {
	type args struct {
		version         string
		providerSource  string
		providerVersion string
		mg              resource.Managed
		kube            client.Client
	}
	type want struct {
		setup terraform.Setup
		err   error
	}

	providerConfigName := "test-provider-config"
	secretName := "test-secret"
	secretNamespace := "crossplane-system"
	apiKey := "tskey-api-test123456789"

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulSetupWithAPIKey": {
			reason: "Should successfully setup Terraform with API key credentials",
			args: args{
				version:         "v0.1.0",
				providerSource:  TerraformProviderSource,
				providerVersion: TerraformProviderVersion,
				mg: newManagedWithProviderConfigRef(providerConfigName),
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch o := obj.(type) {
						case *v1beta1.ProviderConfig:
							// Return a valid ProviderConfig
							*o = v1beta1.ProviderConfig{
								ObjectMeta: metav1.ObjectMeta{
									Name: providerConfigName,
								},
								Spec: v1beta1.ProviderConfigSpec{
									Credentials: v1beta1.ProviderCredentials{
										Source: "Secret",
										CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
											SecretRef: &xpv1.SecretKeySelector{
												SecretReference: xpv1.SecretReference{
													Name:      secretName,
													Namespace: secretNamespace,
												},
												Key: KeyAPIKey,
											},
										},
									},
								},
							}
							return nil
						case *corev1.Secret:
							// Return a valid secret with API key
							*o = corev1.Secret{
								ObjectMeta: metav1.ObjectMeta{
									Name:      secretName,
									Namespace: secretNamespace,
								},
								Data: map[string][]byte{
									KeyAPIKey: []byte(apiKey),
								},
							}
							return nil
						default:
							return errors.New("unexpected object type")
						}
					},
				},
			},
			want: want{
				setup: terraform.Setup{
					Version: "v0.1.0",
					Requirement: terraform.ProviderRequirement{
						Source:  TerraformProviderSource,
						Version: TerraformProviderVersion,
					},
					Configuration: map[string]any{
						"api_key": apiKey,
					},
				},
				err: nil,
			},
		},
		"MissingProviderConfigReference": {
			reason: "Should return error when no provider config is referenced",
			args: args{
				version:         "v0.1.0",
				providerSource:  TerraformProviderSource,
				providerVersion: TerraformProviderVersion,
				mg: &tailnetkeyv1alpha1.Key{}, // no ProviderConfigReference set
				kube: &test.MockClient{},
			},
			want: want{
				setup: terraform.Setup{
					Version: "v0.1.0",
					Requirement: terraform.ProviderRequirement{
						Source:  TerraformProviderSource,
						Version: TerraformProviderVersion,
					},
				},
				err: errors.New("no provider config referenced"),
			},
		},
		"ProviderConfigNotFound": {
			reason: "Should return error when provider config cannot be found",
			args: args{
				version:         "v0.1.0",
				providerSource:  TerraformProviderSource,
				providerVersion: TerraformProviderVersion,
				mg: newManagedWithProviderConfigRef(providerConfigName),
				kube: &test.MockClient{
					MockGet: func(_ context.Context, _ client.ObjectKey, _ client.Object) error {
						return errors.New("provider config not found")
					},
				},
			},
			want: want{
				setup: terraform.Setup{
					Version: "v0.1.0",
					Requirement: terraform.ProviderRequirement{
						Source:  TerraformProviderSource,
						Version: TerraformProviderVersion,
					},
				},
				err: errors.New("cannot get provider config: provider config not found"),
			},
		},
		"CredentialExtractionFailure": {
			reason: "Should return error when credentials cannot be extracted",
			args: args{
				version:         "v0.1.0",
				providerSource:  TerraformProviderSource,
				providerVersion: TerraformProviderVersion,
				mg: newManagedWithProviderConfigRef(providerConfigName),
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch o := obj.(type) {
						case *v1beta1.ProviderConfig:
							*o = v1beta1.ProviderConfig{
								ObjectMeta: metav1.ObjectMeta{
									Name: providerConfigName,
								},
								Spec: v1beta1.ProviderConfigSpec{
									Credentials: v1beta1.ProviderCredentials{
										Source: "Secret",
										CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
											SecretRef: &xpv1.SecretKeySelector{
												SecretReference: xpv1.SecretReference{
													Name:      secretName,
													Namespace: secretNamespace,
												},
												Key: KeyAPIKey,
											},
										},
									},
								},
							}
							return nil
						case *corev1.Secret:
							// Return error when fetching secret
							return errors.New("secret not found")
						default:
							return errors.New("unexpected object type")
						}
					},
				},
			},
			want: want{
				setup: terraform.Setup{
					Version: "v0.1.0",
					Requirement: terraform.ProviderRequirement{
						Source:  TerraformProviderSource,
						Version: TerraformProviderVersion,
					},
				},
				err: errors.New("cannot extract credentials"),
			},
		},
		"EmptyCredentials": {
			reason: "Should handle empty credentials gracefully",
			args: args{
				version:         "v0.1.0",
				providerSource:  TerraformProviderSource,
				providerVersion: TerraformProviderVersion,
				mg: newManagedWithProviderConfigRef(providerConfigName),
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch o := obj.(type) {
						case *v1beta1.ProviderConfig:
							*o = v1beta1.ProviderConfig{
								ObjectMeta: metav1.ObjectMeta{
									Name: providerConfigName,
								},
								Spec: v1beta1.ProviderConfigSpec{
									Credentials: v1beta1.ProviderCredentials{
										Source: "Secret",
										CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
											SecretRef: &xpv1.SecretKeySelector{
												SecretReference: xpv1.SecretReference{
													Name:      secretName,
													Namespace: secretNamespace,
												},
												Key: KeyAPIKey,
											},
										},
									},
								},
							}
							return nil
						case *corev1.Secret:
							// Return empty credentials
							*o = corev1.Secret{
								ObjectMeta: metav1.ObjectMeta{
									Name:      secretName,
									Namespace: secretNamespace,
								},
								Data: map[string][]byte{
									KeyAPIKey: []byte(""),
								},
							}
							return nil
						default:
							return errors.New("unexpected object type")
						}
					},
				},
			},
			want: want{
				setup: terraform.Setup{
					Version: "v0.1.0",
					Requirement: terraform.ProviderRequirement{
						Source:  TerraformProviderSource,
						Version: TerraformProviderVersion,
					},
					Configuration: map[string]any{},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			setupFn := TerraformSetupBuilder(tc.args.version, tc.args.providerSource, tc.args.providerVersion)
			got, err := setupFn(context.Background(), tc.args.kube, tc.args.mg)

			// Check error - be flexible about error wrapping
			if tc.want.err != nil {
				if err == nil {
					t.Errorf("\n%s\nTerraformSetupBuilder(...): expected error, got nil", tc.reason)
				} else if !strings.Contains(err.Error(), tc.want.err.Error()) {
					t.Errorf("\n%s\nTerraformSetupBuilder(...): error should contain %q, got %q", tc.reason, tc.want.err.Error(), err.Error())
				}
			} else if err != nil {
				t.Errorf("\n%s\nTerraformSetupBuilder(...): unexpected error: %v", tc.reason, err)
			}

			if diff := cmp.Diff(tc.want.setup, got, cmp.AllowUnexported(terraform.Setup{})); diff != "" {
				t.Errorf("\n%s\nTerraformSetupBuilder(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "KeyAPIKey constant",
			constant: KeyAPIKey,
			expected: "api_key",
		},
		{
			name:     "KeyOAuthClientID constant",
			constant: KeyOAuthClientID,
			expected: "oauth_client_id",
		},
		{
			name:     "KeyOAuthClientSecret constant",
			constant: KeyOAuthClientSecret,
			expected: "oauth_client_secret",
		},
		{
			name:     "KeyTailnet constant",
			constant: KeyTailnet,
			expected: "tailnet",
		},
		{
			name:     "TerraformProviderSource constant",
			constant: TerraformProviderSource,
			expected: "tailscale/tailscale",
		},
		{
			name:     "TerraformProviderVersion constant",
			constant: TerraformProviderVersion,
			expected: "0.22.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

// TestSetup is commented out as it requires a full manager setup
// which is complex to mock. The Setup function is covered by integration tests.
// func TestSetup(t *testing.T) { ... }
