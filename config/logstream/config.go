// Package logstream contains configuration for Tailscale log streaming resources.
package logstream

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures log streaming resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_logstream_configuration", func(r *config.Resource) {
		// Use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "logstream"

		// Kind will be Configuration
		r.Kind = "Configuration"

		r.UseAsync = false
	})
}
