// Package tailnetkey contains configuration for Tailscale auth key resources.
package tailnetkey

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures the tailnet key (auth key) resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_tailnet_key", func(r *config.Resource) {
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
