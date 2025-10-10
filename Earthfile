VERSION 0.8
PROJECT millstonehq/mill

# Crossplane Provider Tailscale Build Pipeline
# Self-contained build pipeline for the Tailscale Crossplane provider

builder-base:
    FROM ../../lib/build-config/base/+base-builder

    USER root
    # Install Go and tools for building the provider
    RUN apk add go git make

    # Install goimports, controller-gen, and angryjet to /usr/local/bin (in PATH)
    RUN go install golang.org/x/tools/cmd/goimports@latest && \
        go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest && \
        go install github.com/crossplane/crossplane-tools/cmd/angryjet@latest && \
        mv /root/go/bin/goimports /usr/local/bin/ && \
        mv /root/go/bin/controller-gen /usr/local/bin/ && \
        mv /root/go/bin/angryjet /usr/local/bin/

    # Install crossplane CLI for building xpkg packages
    RUN curl -sL "https://releases.crossplane.io/stable/current/bin/linux_amd64/crank" -o /usr/local/bin/crossplane && \
        chmod +x /usr/local/bin/crossplane

    USER nonroot
    WORKDIR /app

deps:
    FROM +builder-base

    COPY go.mod go.sum ./
    RUN go mod download

schema:
    FROM ../../lib/build-config/terraform/+terraform-builder

    # Extract provider schema using OpenTofu
    WORKDIR /tmp/terraform
    RUN echo '{"terraform":[{"required_providers":[{"tailscale":{"source":"tailscale/tailscale","version":"0.19.0"}}]}]}' > main.tf.json && \
        tofu init && \
        tofu providers schema -json=true > /app/schema.json

    SAVE ARTIFACT /app/schema.json AS LOCAL config/schema.json

generate-raw:
    FROM +deps

    # Copy only source files, exclude ALL generated directories
    COPY --dir cmd config examples hack /app/providers/provider-upjet-tailscale/
    COPY --dir internal/clients internal/features /app/providers/provider-upjet-tailscale/internal/
    COPY --dir internal/controller/providerconfig /app/providers/provider-upjet-tailscale/internal/controller/
    COPY --dir apis/v1alpha1 apis/v1beta1 /app/providers/provider-upjet-tailscale/apis/
    COPY package/crossplane.yaml /app/providers/provider-upjet-tailscale/package/crossplane.yaml
    COPY go.mod go.sum /app/providers/provider-upjet-tailscale/
    COPY +schema/schema.json /app/providers/provider-upjet-tailscale/config/schema.json
    WORKDIR /app/providers/provider-upjet-tailscale

    # Download dependencies for the copied code (new config packages need resolution)
    RUN go mod download

    # Run Upjet code generation (generates fresh: apis/zz_register.go, apis/*/v1alpha1, internal/controller/*, package/crds/)
    RUN go run cmd/generator/main.go "$(pwd)"
    
    # Save generated code before controller-gen
    SAVE ARTIFACT apis AS LOCAL apis-generated
    SAVE ARTIFACT internal AS LOCAL internal-generated

generate:
    FROM +generate-raw

    # Generate DeepCopy methods for all API types
    RUN controller-gen object:headerFile=hack/boilerplate.go.txt paths="./apis/..."

    # Generate GetItems() and other resource methods using angryjet
    RUN angryjet generate-methodsets --header-file=hack/boilerplate.go.txt ./apis/...

    # Generate CRDs
    RUN controller-gen crd:allowDangerousTypes=true paths="./apis/..." output:crd:artifacts:config=package/crds

    # Debug: check what was generated
    RUN ls -la && ls -la apis/ && ls -la package/ && ls -la internal/controller/ || true

    SAVE ARTIFACT apis AS LOCAL apis
    SAVE ARTIFACT package AS LOCAL package
    SAVE ARTIFACT internal AS LOCAL internal

build:
    FROM +generate

    # Build the provider binary with optimizations
    # -ldflags="-s -w" strips debug info and symbol table (saves ~15MB)
    # -trimpath removes file system paths from binary
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w" \
        -trimpath \
        -o bin/provider \
        cmd/provider/main.go

    SAVE ARTIFACT bin/provider AS LOCAL bin/provider

image:
    # Production runtime: terraform-runtime (base-runtime + tofu only, no debug tools)
    FROM ../../lib/build-config/terraform/+terraform-runtime

    COPY +build/provider /usr/local/bin/provider

    ENTRYPOINT ["/usr/local/bin/provider"]

    ARG VERSION=v0.1.0
    # Local build: save both tags but don't push
    SAVE IMAGE ghcr.io/millstonehq/provider-tailscale:${VERSION}
    SAVE IMAGE ghcr.io/millstonehq/provider-tailscale:latest

push:
    # Push target: only push :latest tag to minimize GHCR storage
    # Run with: earthly --push +push
    FROM +image

    ARG VERSION=v0.1.0
    # Only push latest tag to save GHCR storage (175MB per image)
    # For versioned releases, manually tag and push specific versions
    SAVE IMAGE --push ghcr.io/millstonehq/provider-tailscale:latest

controller-tarball:
    LOCALLY
    # Ensure controller image is built
    BUILD +image
    RUN docker save ghcr.io/millstonehq/provider-tailscale:latest -o /tmp/provider-controller.tar
    SAVE ARTIFACT /tmp/provider-controller.tar controller.tar

package-build:
    FROM +generate

    # Get controller image tarball
    COPY +controller-tarball/controller.tar /tmp/controller.tar

    # Build xpkg package with embedded controller runtime tarball
    RUN crossplane xpkg build \
        --package-root=package \
        --embed-runtime-image-tarball=/tmp/controller.tar \
        -o package.xpkg

    SAVE ARTIFACT package.xpkg

package-local:
    ARG IMAGE_NAME=provider-tailscale:latest

    # Load the xpkg tarball into docker
    LOCALLY
    COPY +package-build/package.xpkg /tmp/provider-tailscale-package.xpkg
    RUN LOADED_ID=$(docker load -i /tmp/provider-tailscale-package.xpkg 2>&1 | grep -oP 'Loaded image ID: \K.*') && \
        docker tag $LOADED_ID $IMAGE_NAME
