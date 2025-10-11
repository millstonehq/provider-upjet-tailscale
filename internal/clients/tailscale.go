// Package clients contains the provider config setup.
package clients

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/millstonehq/provider-upjet-tailscale/apis/v1beta1"
)

const (
	// KeyAPIKey is the key for the Tailscale API key in credentials
	KeyAPIKey = "api_key"
	// KeyOAuthClientID is the key for OAuth client ID
	KeyOAuthClientID = "oauth_client_id"
	// KeyOAuthClientSecret is the key for OAuth client secret
	KeyOAuthClientSecret = "oauth_client_secret"
	// KeyTailnet is the key for the tailnet name
	KeyTailnet = "tailnet"

	// TerraformProviderSource is the source for the Terraform provider
	TerraformProviderSource = "tailscale/tailscale"
	// TerraformProviderVersion is the version of the Terraform provider
	TerraformProviderVersion = "0.22.0"
)

// TerraformSetupBuilder returns Terraform setup with provider config.
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, kube client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		// Resolve provider config reference (v2 API)
		var configRef *string
		switch m := mg.(type) {
		case resource.LegacyManaged:
			if ref := m.GetProviderConfigReference(); ref != nil {
				configRef = &ref.Name
			}
		case resource.ModernManaged:
			if ref := m.GetProviderConfigReference(); ref != nil {
				configRef = &ref.Name
			}
		default:
			return ps, fmt.Errorf("resource is neither LegacyManaged nor ModernManaged")
		}

		if configRef == nil {
			return ps, fmt.Errorf("no provider config referenced")
		}

		pc := &v1beta1.ProviderConfig{}
		if err := kube.Get(ctx, types.NamespacedName{Name: *configRef}, pc); err != nil {
			return ps, fmt.Errorf("cannot get provider config: %w", err)
		}

		// Get credentials from the referenced secret
		// In crossplane-runtime v2, CommonCredentialExtractor returns []byte
		credData, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, kube, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, fmt.Errorf("cannot extract credentials: %w", err)
		}

		// Configure Terraform provider based on available credentials
		ps.Configuration = map[string]any{}

		// The credentials are returned as raw bytes, so we treat them as the API key directly
		if len(credData) > 0 {
			ps.Configuration["api_key"] = string(credData)
		}

		return ps, nil
	}
}
