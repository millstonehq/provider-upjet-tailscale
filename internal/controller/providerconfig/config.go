// Package providerconfig contains the ProviderConfig controller setup.
package providerconfig

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	"github.com/millstonehq/provider-upjet-tailscale/internal/clients"
)

// Setup sets up the ProviderConfig controller.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return clients.Setup(mgr, o)
}

// SetupGated sets up the ProviderConfig controller gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	// ProviderConfig doesn't use gated setup, just call regular Setup
	return Setup(mgr, o)
}
