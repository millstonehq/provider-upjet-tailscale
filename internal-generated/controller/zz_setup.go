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
	authorization "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/authorization"
	tags "github.com/millstonehq/provider-upjet-tailscale/internal/controller/device/tags"
	nameservers "github.com/millstonehq/provider-upjet-tailscale/internal/controller/dns/nameservers"
	providerconfig "github.com/millstonehq/provider-upjet-tailscale/internal/controller/providerconfig"
	key "github.com/millstonehq/provider-upjet-tailscale/internal/controller/tailnetkey/key"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		acl.Setup,
		authorization.Setup,
		tags.Setup,
		nameservers.Setup,
		providerconfig.Setup,
		key.Setup,
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
		authorization.SetupGated,
		tags.SetupGated,
		nameservers.SetupGated,
		providerconfig.SetupGated,
		key.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
