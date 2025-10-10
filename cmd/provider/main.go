// Package main is the entry point for the Tailscale Crossplane provider.
package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kingpin/v2"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/millstonehq/provider-upjet-tailscale/apis"
	"github.com/millstonehq/provider-upjet-tailscale/config"
	"github.com/millstonehq/provider-upjet-tailscale/internal/clients"
	"github.com/millstonehq/provider-upjet-tailscale/internal/controller"
	"github.com/millstonehq/provider-upjet-tailscale/internal/features"
)

func main() {
	var (
		app                    = kingpin.New(filepath.Base(os.Args[0]), "Tailscale support for Crossplane.").DefaultEnvars()
		debug                  = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncInterval           = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration()
		pollInterval           = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").Default("10m").Duration()
		leaderElection         = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").Envar("LEADER_ELECTION").Bool()
		maxReconcileRate       = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("10").Int()
		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for Management Policies.").Default("true").Envar("ENABLE_MANAGEMENT_POLICIES").Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-tailscale"))
	if *debug {
		ctrl.SetLogger(zl)
	}

	pollJitter := time.Duration(float64(*pollInterval) * 0.05)
	log.Debug("Starting", "sync-interval", syncInterval.String(),
		"poll-interval", pollInterval.String(), "poll-jitter", pollJitter, "max-reconcile-rate", *maxReconcileRate)

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	// Create scheme and register provider APIs
	scheme := runtime.NewScheme()
	kingpin.FatalIfError(apis.AddToScheme(scheme), "Cannot add provider APIs to scheme")
	kingpin.FatalIfError(corev1.AddToScheme(scheme), "Cannot add Kubernetes core API types to scheme")

	// Setup controller manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		LeaderElection:             *leaderElection,
		LeaderElectionID:           "crossplane-leader-election-provider-tailscale",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		Cache: cache.Options{
			SyncPeriod: syncInterval,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: "/webhook/certs",
		}),
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	// Initialize provider configuration
	providerConfig := config.GetProvider()
	if providerConfig == nil {
		kingpin.Fatalf("config.GetProvider() returned nil")
	}
	if providerConfig.Resources == nil {
		kingpin.Fatalf("providerConfig.Resources is nil")
	}
	if _, ok := providerConfig.Resources["tailscale_acl"]; !ok {
		kingpin.Fatalf("tailscale_acl not found in Resources map, available: %v", func() []string {
			keys := make([]string, 0, len(providerConfig.Resources))
			for k := range providerConfig.Resources {
				keys = append(keys, k)
			}
			return keys
		}())
	}
	log.Info("Provider initialized successfully", "resources", len(providerConfig.Resources))
	
	setupFn := clients.TerraformSetupBuilder(
		"1.5.5",
		clients.TerraformProviderSource,
		clients.TerraformProviderVersion,
	)

	// Setup controller options
	o := tjcontroller.Options{
		Options: xpcontroller.Options{
			Logger:                  log,
			MaxConcurrentReconciles: *maxReconcileRate,
			PollInterval:            *pollInterval,
			GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
			Features:                &feature.Flags{},
		},
		Provider:       providerConfig,
		SetupFn:        setupFn,
		WorkspaceStore: terraform.NewWorkspaceStore(log),
		PollJitter:     pollJitter,
	}

	if *enableManagementPolicies {
		o.Features.Enable(features.EnableBetaManagementPolicies)
		log.Info("Beta feature enabled", "flag", features.EnableBetaManagementPolicies)
	}

	// Setup all controllers (including ProviderConfig via generated zz_setup.go)
	kingpin.FatalIfError(controller.Setup(mgr, o), "Cannot setup controllers")

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
