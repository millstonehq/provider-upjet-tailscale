// Package posture contains configuration for Tailscale device posture resources.
package posture

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures device posture resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_posture_integration", func(r *config.Resource) {
		// Use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "posture"

		// Kind will be Integration
		r.Kind = "Integration"

		r.UseAsync = false
	})
}
