// Package dns contains configuration for Tailscale DNS resources.
package dns

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures DNS resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_dns_nameservers", func(r *config.Resource) {
		// DNS nameservers is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "dns"

		// Kind will be Nameservers
		r.Kind = "Nameservers"

		r.UseAsync = false
	})

	p.AddResourceConfigurator("tailscale_dns_preferences", func(r *config.Resource) {
		// DNS preferences is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "dns"

		// Kind will be Preferences
		r.Kind = "Preferences"

		r.UseAsync = false
	})

	p.AddResourceConfigurator("tailscale_dns_search_paths", func(r *config.Resource) {
		// DNS search paths is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "dns"

		// Kind will be SearchPaths
		r.Kind = "SearchPaths"

		r.UseAsync = false
	})

	p.AddResourceConfigurator("tailscale_dns_split_nameservers", func(r *config.Resource) {
		// Use domain as the external identifier
		r.ExternalName = config.ParameterAsIdentifier("domain")

		// Short group for CRD generation
		r.ShortGroup = "dns"

		// Kind will be SplitNameservers
		r.Kind = "SplitNameservers"

		r.UseAsync = false
	})
}
