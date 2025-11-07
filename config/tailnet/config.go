// Package tailnet contains configuration for Tailscale tailnet-level resources.
package tailnet

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures tailnet-level resources.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("tailscale_contacts", func(r *config.Resource) {
		// Contacts is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "tailnet"

		// Kind will be Contacts
		r.Kind = "Contacts"

		r.UseAsync = false
	})

	a.AddResourceConfigurator("tailscale_tailnet_settings", func(r *config.Resource) {
		// Tailnet settings is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "tailnet"

		// Kind will be Settings
		r.Kind = "Settings"

		r.UseAsync = false
	})
}
