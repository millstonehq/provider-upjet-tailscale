// Package clients contains the provider configuration controller setup.
package clients

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/providerconfig"

	"github.com/millstonehq/provider-upjet-tailscale/apis/v1beta1"
)

// Setup sets up the ProviderConfig controller.
func Setup(mgr manager.Manager, o controller.Options) error {
	return providerconfig.Setup(mgr, o,
		providerconfig.WithReconciler(
			providerconfig.NewReconciler(mgr,
				providerconfig.WithLogger(o.Logger.WithValues("controller", "providerconfig")),
			),
		),
		providerconfig.WithGVK(&v1beta1.ProviderConfigGroupVersionKind),
	)
}
