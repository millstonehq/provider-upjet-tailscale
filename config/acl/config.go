// Package acl contains configuration for Tailscale ACL resources.
package acl

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures the ACL resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_acl", func(r *config.Resource) {
		// ACL is a singleton resource in Tailscale - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "acl"

		// Kind will be ACL
		r.Kind = "ACL"

		// ACL resource supports HuJSON format
		r.UseAsync = false
	})
}
