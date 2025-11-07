// Package main contains tests for validating example manifests.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// KubernetesObject represents a minimal Kubernetes object for validation
type KubernetesObject struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
}

func TestExampleManifests(t *testing.T) {
	// Find all YAML files in examples directory
	exampleFiles, err := filepath.Glob("examples/**/*.yaml")
	if err != nil {
		t.Fatalf("Failed to glob example files: %v", err)
	}

	if len(exampleFiles) == 0 {
		t.Fatal("No example files found")
	}

	t.Logf("Found %d example files to validate", len(exampleFiles))

	for _, file := range exampleFiles {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}

			// Parse YAML
			var obj KubernetesObject
			if err := yaml.Unmarshal(data, &obj); err != nil {
				t.Fatalf("Failed to parse YAML in %s: %v", file, err)
			}

			// Validate required fields
			if obj.APIVersion == "" {
				t.Errorf("Missing apiVersion in %s", file)
			}

			if obj.Kind == "" {
				t.Errorf("Missing kind in %s", file)
			}

			if obj.Metadata == nil {
				t.Errorf("Missing metadata in %s", file)
			} else {
				// Check for name
				if name, ok := obj.Metadata["name"]; !ok || name == "" {
					t.Errorf("Missing metadata.name in %s", file)
				}
			}

			// Validate API version format (should contain a group and version)
			if obj.APIVersion != "" && obj.APIVersion != "v1" {
				parts := strings.Split(obj.APIVersion, "/")
				if len(parts) != 2 {
					t.Errorf("Invalid apiVersion format in %s: expected 'group/version', got '%s'", file, obj.APIVersion)
				}
			}

			t.Logf("✓ Validated %s: %s/%s", file, obj.APIVersion, obj.Kind)
		})
	}
}

func TestProviderConfigExamples(t *testing.T) {
	// Validate ProviderConfig examples specifically
	providerConfigFile := "examples/providerconfig/providerconfig.yaml"
	secretFile := "examples/providerconfig/secret.yaml"

	t.Run("ProviderConfig", func(t *testing.T) {
		data, err := os.ReadFile(providerConfigFile)
		if err != nil {
			t.Fatalf("Failed to read providerconfig: %v", err)
		}

		var obj KubernetesObject
		if err := yaml.Unmarshal(data, &obj); err != nil {
			t.Fatalf("Failed to parse providerconfig: %v", err)
		}

		// Validate specific fields for ProviderConfig
		if obj.APIVersion != "tailscale.upbound.io/v1beta1" {
			t.Errorf("Expected apiVersion 'tailscale.upbound.io/v1beta1', got '%s'", obj.APIVersion)
		}

		if obj.Kind != "ProviderConfig" {
			t.Errorf("Expected kind 'ProviderConfig', got '%s'", obj.Kind)
		}

		// Check spec.credentials exists
		if obj.Spec == nil {
			t.Fatal("Missing spec in ProviderConfig")
		}

		if _, ok := obj.Spec["credentials"]; !ok {
			t.Error("Missing spec.credentials in ProviderConfig")
		}

		t.Logf("✓ ProviderConfig validated successfully")
	})

	t.Run("Secret", func(t *testing.T) {
		data, err := os.ReadFile(secretFile)
		if err != nil {
			t.Fatalf("Failed to read secret: %v", err)
		}

		var obj KubernetesObject
		if err := yaml.Unmarshal(data, &obj); err != nil {
			t.Fatalf("Failed to parse secret: %v", err)
		}

		// Validate specific fields for Secret
		if obj.APIVersion != "v1" {
			t.Errorf("Expected apiVersion 'v1', got '%s'", obj.APIVersion)
		}

		if obj.Kind != "Secret" {
			t.Errorf("Expected kind 'Secret', got '%s'", obj.Kind)
		}

		t.Logf("✓ Secret validated successfully")
	})
}

func TestResourceExamples(t *testing.T) {
	// Test that all resource examples have required fields
	resourceExamples := map[string]struct {
		apiGroup string
		kind     string
	}{
		"examples/acl/acl.yaml": {
			apiGroup: "acl.tailscale.upbound.io/v1alpha1",
			kind:     "ACL",
		},
		"examples/dns/nameservers.yaml": {
			apiGroup: "dns.tailscale.upbound.io/v1alpha1",
			kind:     "Nameservers",
		},
		"examples/tailnetkey/key.yaml": {
			apiGroup: "tailnetkey.tailscale.upbound.io/v1alpha1",
			kind:     "Key",
		},
		"examples/device/authorization.yaml": {
			apiGroup: "device.tailscale.upbound.io/v1alpha1",
			kind:     "Authorization",
		},
		"examples/device/tags.yaml": {
			apiGroup: "device.tailscale.upbound.io/v1alpha1",
			kind:     "Tags",
		},
	}

	for file, expected := range resourceExamples {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			var obj KubernetesObject
			if err := yaml.Unmarshal(data, &obj); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			// Validate API version
			if obj.APIVersion != expected.apiGroup {
				t.Errorf("Expected apiVersion '%s', got '%s'", expected.apiGroup, obj.APIVersion)
			}

			// Validate kind
			if obj.Kind != expected.kind {
				t.Errorf("Expected kind '%s', got '%s'", expected.kind, obj.Kind)
			}

			// Validate spec exists
			if obj.Spec == nil {
				t.Error("Missing spec field")
			}

			// Validate spec.forProvider exists
			if obj.Spec != nil {
				if _, ok := obj.Spec["forProvider"]; !ok {
					t.Error("Missing spec.forProvider field")
				}

				// Validate providerConfigRef exists
				if _, ok := obj.Spec["providerConfigRef"]; !ok {
					t.Error("Missing spec.providerConfigRef field")
				}
			}

			t.Logf("✓ Resource validated: %s/%s", obj.APIVersion, obj.Kind)
		})
	}
}

func TestExampleCombinations(t *testing.T) {
	// Test that provider config + resource examples can be parsed together
	t.Run("ACL with ProviderConfig", func(t *testing.T) {
		files := []string{
			"examples/providerconfig/providerconfig.yaml",
			"examples/acl/acl.yaml",
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", file, err)
			}

			var obj KubernetesObject
			if err := yaml.Unmarshal(data, &obj); err != nil {
				t.Fatalf("Failed to parse %s: %v", file, err)
			}
		}

		t.Log("✓ ACL + ProviderConfig combination validated")
	})

	t.Run("Device examples with ProviderConfig", func(t *testing.T) {
		files := []string{
			"examples/providerconfig/providerconfig.yaml",
			"examples/device/authorization.yaml",
			"examples/device/tags.yaml",
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", file, err)
			}

			var obj KubernetesObject
			if err := yaml.Unmarshal(data, &obj); err != nil {
				t.Fatalf("Failed to parse %s: %v", file, err)
			}
		}

		t.Log("✓ Device examples + ProviderConfig combination validated")
	})
}

func TestREADMEExists(t *testing.T) {
	// Ensure examples README exists and is not empty
	readmePath := "examples/README.md"
	
	info, err := os.Stat(readmePath)
	if err != nil {
		t.Fatalf("Examples README not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Examples README is empty")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README: %v", err)
	}

	// Check for key sections
	requiredSections := []string{
		"Getting Started",
		"Installation",
		"Available Examples",
		"Usage Pattern",
	}

	for _, section := range requiredSections {
		if !strings.Contains(string(content), section) {
			t.Errorf("README missing section: %s", section)
		}
	}

	t.Log("✓ Examples README validated")
}

func TestExampleDocumentation(t *testing.T) {
	// Ensure each example directory has proper documentation
	exampleDirs := []string{
		"examples/acl",
		"examples/device",
		"examples/dns",
		"examples/providerconfig",
		"examples/tailnetkey",
	}

	for _, dir := range exampleDirs {
		t.Run(dir, func(t *testing.T) {
			// Check that directory exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Fatalf("Example directory does not exist: %s", dir)
			}

			// Check that directory contains at least one YAML file
			files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
			if err != nil {
				t.Fatalf("Failed to glob YAML files in %s: %v", dir, err)
			}

			if len(files) == 0 {
				t.Errorf("No YAML files found in %s", dir)
			}

			t.Logf("✓ Found %d example(s) in %s", len(files), dir)
		})
	}
}
