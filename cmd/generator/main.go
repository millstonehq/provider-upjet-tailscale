// Package main generates Crossplane resources from Terraform provider schema.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crossplane/upjet/v2/pkg/pipeline"

	"github.com/millstonehq/provider-upjet-tailscale/config"
)

func main() {
	if len(os.Args) < 2 {
		panic("root directory is required as argument")
	}

	rootDir := os.Args[1]
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot get absolute path for root directory: %v", err))
	}

	pc := config.GetProvider()

	// Tailscale resources are cluster-scoped only for v1
	// Passing nil for namespaced provider generates resources in apis/ directly
	pipeline.Run(pc, nil, absRootDir)
}
