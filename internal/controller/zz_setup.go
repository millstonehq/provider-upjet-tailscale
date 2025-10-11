/*
Copyright 2025 Millstone HQ.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	acl "github.com/millstonehq/provider-upjet-tailscale/internal/controller/acl/acl"
	externalid "github.com/millstonehq/provider-upjet-tailscale/internal/controller/aws/externalid"
	authorization "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/authorization"
	key "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/key"
	subnetroutes "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/subnetroutes"
	tags "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/tags"
	nameservers "github.com/millstonehq/provider-upjet-tailscale/internal/controller/dns/nameservers"
	preferences "github.com/millstonehq/provider-upjet-tailscale/internal/controller/dns/preferences"
	searchpaths "github.com/millstonehq/provider-upjet-tailscale/internal/controller/dns/searchpaths"
	splitnameservers "github.com/millstonehq/provider-upjet-tailscale/internal/controller/dns/splitnameservers"
	configuration "github.com/millstonehq/provider-upjet-tailscale/internal/controller/logstream/configuration"
	client "github.com/millstonehq/provider-upjet-tailscale/internal/controller/oauth/client"
	integration "github.com/millstonehq/provider-upjet-tailscale/internal/controller/posture/integration"
	providerconfig "github.com/millstonehq/provider-upjet-tailscale/internal/controller/providerconfig"
	contacts "github.com/millstonehq/provider-upjet-tailscale/internal/controller/tailnet/contacts"
	settings "github.com/millstonehq/provider-upjet-tailscale/internal/controller/tailnet/settings"
	keytailnetkey "github.com/millstonehq/provider-upjet-tailscale/internal/controller/tailnetkey/key"
	webhook "github.com/millstonehq/provider-upjet-tailscale/internal/controller/webhook/webhook"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		acl.Setup,
		externalid.Setup,
		authorization.Setup,
		key.Setup,
		subnetroutes.Setup,
		tags.Setup,
		nameservers.Setup,
		preferences.Setup,
		searchpaths.Setup,
		splitnameservers.Setup,
		configuration.Setup,
		client.Setup,
		integration.Setup,
		providerconfig.Setup,
		contacts.Setup,
		settings.Setup,
		keytailnetkey.Setup,
		webhook.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		acl.SetupGated,
		externalid.SetupGated,
		authorization.SetupGated,
		key.SetupGated,
		subnetroutes.SetupGated,
		tags.SetupGated,
		nameservers.SetupGated,
		preferences.SetupGated,
		searchpaths.SetupGated,
		splitnameservers.SetupGated,
		configuration.SetupGated,
		client.SetupGated,
		integration.SetupGated,
		providerconfig.SetupGated,
		contacts.SetupGated,
		settings.SetupGated,
		keytailnetkey.SetupGated,
		webhook.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
