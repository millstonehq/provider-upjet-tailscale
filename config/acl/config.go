// Package acl contains configuration for Tailscale ACL resources.
package acl

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures the ACL resource.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("tailscale_acl", func(r *config.Resource) {
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
