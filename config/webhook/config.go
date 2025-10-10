// Package webhook contains configuration for Tailscale webhook resources.
package webhook

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures webhook resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_webhook", func(r *config.Resource) {
		// Use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "webhook"

		// Kind will be Webhook
		r.Kind = "Webhook"

		r.UseAsync = false
	})
}
