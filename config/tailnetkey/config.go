// Package tailnetkey contains configuration for Tailscale auth key resources.
package tailnetkey

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures the tailnet key (auth key) resource.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("tailscale_tailnet_key", func(r *config.Resource) {
		// Auth keys are server-generated, use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "tailnetkey"

		// Kind will be Key
		r.Kind = "Key"

		r.UseAsync = false

		// Sensitive fields that should be marked as secret
		r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
			conn := map[string][]byte{}
			if key, ok := attr["key"].(string); ok {
				conn["key"] = []byte(key)
			}
			return conn, nil
		}
	})
}
