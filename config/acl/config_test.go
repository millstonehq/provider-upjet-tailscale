// Package acl contains configuration for Tailscale ACL resources.
package acl

import (
	"testing"

	"github.com/crossplane/upjet/v2/pkg/config"
)

type fakeAdder struct {
	name string
	fn   config.ResourceConfiguratorFn
}

func (f *fakeAdder) AddResourceConfigurator(name string, rc config.ResourceConfiguratorFn) {
	f.name = name
	f.fn = rc
}

func TestConfigureRegistersConfigurator(t *testing.T) {
	f := &fakeAdder{}
	configureWithAdder(f)

	if f.name != "tailscale_acl" {
		t.Fatalf("registered for %q, want tailscale_acl", f.name)
	}

	// Execute the captured configurator on a fresh Resource and assert fields.
	r := &config.Resource{}
	f.fn(r)

	if r.ExternalName.GetExternalNameFn == nil {
		t.Error("ExternalName not configured (expected IdentifierFromProvider)")
	}
	if r.ShortGroup != "acl" {
		t.Errorf("ShortGroup = %q, want %q", r.ShortGroup, "acl")
	}
	if r.Kind != "ACL" {
		t.Errorf("Kind = %q, want %q", r.Kind, "ACL")
	}
	if r.UseAsync {
		t.Error("UseAsync = true, want false")
	}
}
