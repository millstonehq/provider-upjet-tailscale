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
	TerraformProviderVersion = "0.18.0"
)

// TerraformSetupBuilder returns Terraform setup with provider config.
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, fmt.Errorf("no provider config referenced")
		}

		pc := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, fmt.Errorf("cannot get provider config: %w", err)
		}

		// Get credentials from the referenced secret
		creds, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, fmt.Errorf("cannot extract credentials: %w", err)
		}

		// Configure Terraform provider based on available credentials
		ps.Configuration = map[string]any{}

		// API Key authentication (preferred for simplicity)
		if apiKey, ok := creds[KeyAPIKey]; ok {
			ps.Configuration["api_key"] = string(apiKey)
		}

		// OAuth authentication (alternative)
		if clientID, ok := creds[KeyOAuthClientID]; ok {
			ps.Configuration["oauth_client_id"] = string(clientID)
		}
		if clientSecret, ok := creds[KeyOAuthClientSecret]; ok {
			ps.Configuration["oauth_client_secret"] = string(clientSecret)
		}

		// Tailnet (optional, can be inferred from API key)
		if tailnet, ok := creds[KeyTailnet]; ok {
			ps.Configuration["tailnet"] = string(tailnet)
		}

		return ps, nil
	}
}
