// Package oauth contains configuration for Tailscale OAuth resources.
package oauth

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures OAuth resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_oauth_client", func(r *config.Resource) {
		// Use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "oauth"

		// Kind will be Client
		r.Kind = "Client"

		r.UseAsync = false
	})
}
