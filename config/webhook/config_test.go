// Package webhook contains configuration for Tailscale resources.
package webhook

import (
	"testing"

	"github.com/crossplane/upjet/v2/pkg/config"
)

type fakeAdder struct {
	names []string
	fns   []config.ResourceConfiguratorFn
}

func (f *fakeAdder) AddResourceConfigurator(name string, rc config.ResourceConfiguratorFn) {
	f.names = append(f.names, name)
	f.fns = append(f.fns, rc)
}

func TestConfigureRegistersConfigurators(t *testing.T) {
	f := &fakeAdder{}
	configureWithAdder(f)

	if len(f.names) == 0 {
		t.Fatal("no configurators registered")
	}

	// Execute each captured configurator and verify it doesn't panic
	for i, name := range f.names {
		t.Run(name, func(t *testing.T) {
			r := &config.Resource{}
			f.fns[i](r)

			// Basic validation that something was configured
			if r.ShortGroup == "" {
				t.Error("ShortGroup not set")
			}
			if r.Kind == "" {
				t.Error("Kind not set")
			}
			if r.ExternalName.GetExternalNameFn == nil {
				t.Error("ExternalName not configured")
			}
		})
	}
}
