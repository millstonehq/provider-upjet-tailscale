// Package oauth contains configuration for Tailscale OAuth resources.
package oauth

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures OAuth resources.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("tailscale_oauth_client", func(r *config.Resource) {
		// Use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "oauth"

		// Kind will be Client
		r.Kind = "Client"

		r.UseAsync = false

		// Configure connection details to match Tailscale operator expectations
		// Operator expects: client_id and client_secret (as files in mounted volume)
		r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
			conn := map[string][]byte{}

			// Extract client ID from the resource ID
			if id, ok := attr["id"].(string); ok {
				conn["client_id"] = []byte(id)
			}

			// Extract client secret from the key attribute
			if key, ok := attr["key"].(string); ok {
				conn["client_secret"] = []byte(key)
			}

			return conn, nil
		}
	})
}
