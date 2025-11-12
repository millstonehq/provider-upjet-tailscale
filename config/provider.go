// Package config contains the provider configuration.
package config

import (
	_ "embed"

	tjconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/millstonehq/provider-upjet-tailscale/config/acl"
	"github.com/millstonehq/provider-upjet-tailscale/config/aws"
	"github.com/millstonehq/provider-upjet-tailscale/config/device"
	"github.com/millstonehq/provider-upjet-tailscale/config/dns"
	"github.com/millstonehq/provider-upjet-tailscale/config/logstream"
	"github.com/millstonehq/provider-upjet-tailscale/config/oauth"
	"github.com/millstonehq/provider-upjet-tailscale/config/posture"
	"github.com/millstonehq/provider-upjet-tailscale/config/tailnet"
	"github.com/millstonehq/provider-upjet-tailscale/config/tailnetkey"
	"github.com/millstonehq/provider-upjet-tailscale/config/webhook"
)

const (
	resourcePrefix = "tailscale"
	modulePath     = "github.com/millstonehq/provider-upjet-tailscale"
)

//go:embed schema.json
var providerSchema []byte

// GetProvider returns provider configuration
func GetProvider() *tjconfig.Provider {
	pc := tjconfig.NewProvider(
		providerSchema,  // Schema extracted by OpenTofu
		resourcePrefix,
		modulePath,
		[]byte{},        // Empty metadata
		tjconfig.WithRootGroup("tailscale.upbound.io"),
		tjconfig.WithFeaturesPackage("internal/features"),
		tjconfig.WithIncludeList([]string{
			// ACL resources
			"tailscale_acl$",
			// AWS resources
			"tailscale_aws_external_id$",
			// Device resources
			"tailscale_device_authorization$",
			"tailscale_device_key$",
			"tailscale_device_subnet_routes$",
			"tailscale_device_tags$",
			// DNS resources
			"tailscale_dns_nameservers$",
			"tailscale_dns_preferences$",
			"tailscale_dns_search_paths$",
			"tailscale_dns_split_nameservers$",
			// Logstream resources
			"tailscale_logstream_configuration$",
			// OAuth resources
			"tailscale_oauth_client$",
			// Posture resources
			"tailscale_posture_integration$",
			// Tailnet resources
			"tailscale_contacts$",
			"tailscale_tailnet_key$",
			"tailscale_tailnet_settings$",
			// Webhook resources
			"tailscale_webhook$",
		}),
		tjconfig.WithDefaultResourceOptions(
			func(r *tjconfig.Resource) {
				r.ExternalName = tjconfig.NameAsIdentifier
			},
		),
	)

	// Configure individual resources
	for _, configure := range []func(*tjconfig.Provider){
		acl.Configure,
		aws.Configure,
		device.Configure,
		dns.Configure,
		logstream.Configure,
		oauth.Configure,
		posture.Configure,
		tailnet.Configure,
		tailnetkey.Configure,
		webhook.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
