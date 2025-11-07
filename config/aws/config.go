// Package aws contains configuration for Tailscale AWS integration resources.
package aws

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// adder is a narrow interface to allow testing without a real Provider.
type adder interface {
	AddResourceConfigurator(name string, f config.ResourceConfiguratorFn)
}

// Configure configures AWS integration resources.
func Configure(p *config.Provider) {
	configureWithAdder(p)
}

// configureWithAdder is the testable entrypoint.
func configureWithAdder(a adder) {
	a.AddResourceConfigurator("tailscale_aws_external_id", func(r *config.Resource) {
		// AWS external ID is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "aws"

		// Kind will be ExternalID
		r.Kind = "ExternalID"

		r.UseAsync = false
	})
}
