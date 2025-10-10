// Package providerconfig contains the ProviderConfig controller setup.
package providerconfig

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	"github.com/millstonehq/provider-upjet-tailscale/internal/clients"
)

// Setup sets up the ProviderConfig controller.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return clients.Setup(mgr, o)
}
