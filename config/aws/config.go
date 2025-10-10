// Package aws contains configuration for Tailscale AWS integration resources.
package aws

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures AWS integration resources.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("tailscale_aws_external_id", func(r *config.Resource) {
		// AWS external ID is a singleton resource - use identifier from provider
		r.ExternalName = config.IdentifierFromProvider

		// Short group for CRD generation
		r.ShortGroup = "aws"

		// Kind will be ExternalID
		r.Kind = "ExternalID"

		r.UseAsync = false
	})
}
