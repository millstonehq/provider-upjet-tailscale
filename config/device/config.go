// Package device contains configuration for Tailscale device resources.
package device

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures device-related resources.
func Configure(p *config.Provider) {
	// Device tags resource
	p.AddResourceConfigurator("tailscale_device_tags", func(r *config.Resource) {
		// Use device_id as the external identifier
		r.ExternalName = config.ParameterAsIdentifier("device_id")

		// Short group for CRD generation
		r.ShortGroup = "device"

		// Kind will be Tags
		r.Kind = "Tags"

		r.UseAsync = false
	})

	// Device authorization resource
	p.AddResourceConfigurator("tailscale_device_authorization", func(r *config.Resource) {
		// Use device_id as the external identifier
		r.ExternalName = config.ParameterAsIdentifier("device_id")

		// Short group for CRD generation
		r.ShortGroup = "device"

		// Kind will be Authorization
		r.Kind = "Authorization"

		r.UseAsync = false
	})

	// Device key resource
	p.AddResourceConfigurator("tailscale_device_key", func(r *config.Resource) {
		// Use device_id as the external identifier
		r.ExternalName = config.ParameterAsIdentifier("device_id")

		// Short group for CRD generation
		r.ShortGroup = "device"

		// Kind will be Key
		r.Kind = "Key"

		r.UseAsync = false
	})

	// Device subnet routes resource
	p.AddResourceConfigurator("tailscale_device_subnet_routes", func(r *config.Resource) {
		// Use device_id as the external identifier
		r.ExternalName = config.ParameterAsIdentifier("device_id")

		// Short group for CRD generation
		r.ShortGroup = "device"

		// Kind will be SubnetRoutes
		r.Kind = "SubnetRoutes"

		r.UseAsync = false
	})
}
