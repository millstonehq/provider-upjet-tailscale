VERSION 0.8
PROJECT millstonehq/mill

# Crossplane Provider Tailscale Build Pipeline
# Self-contained build pipeline for the Tailscale Crossplane provider

builder-base:
    ARG BUILDPLATFORM
    # Use pre-built crossplane:builder with Go, OpenTofu, make, and pre-compiled tools
    # (goimports, controller-gen, angryjet, crossplane CLI)
    FROM --platform=$BUILDPLATFORM ghcr.io/millstonehq/crossplane:builder

    WORKDIR /app

deps:
    FROM +builder-base

    COPY go.mod go.sum ./
    RUN go mod download

schema:
    FROM ghcr.io/millstonehq/tofu:builder

    # Copy source to extract version (single source of truth)
    COPY internal/clients/tailscale.go /tmp/tailscale.go

    # Extract provider schema using OpenTofu with version from Go source
    WORKDIR /tmp/terraform
    RUN PROVIDER_VERSION=$(grep 'TerraformProviderVersion = ' /tmp/tailscale.go | sed 's/.*"\(.*\)".*/\1/') && \
        echo "Using Terraform provider version: $PROVIDER_VERSION" && \
        echo "{\"terraform\":[{\"required_providers\":[{\"tailscale\":{\"source\":\"tailscale/tailscale\",\"version\":\"$PROVIDER_VERSION\"}}]}]}" > main.tf.json && \
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

test:
    FROM +generate

    # Copy test file (other source files already generated)
    COPY go.mod go.sum examples_test.go /app/providers/provider-upjet-tailscale/
    WORKDIR /app/providers/provider-upjet-tailscale

    # Run unit tests with coverage (CGO disabled for pure Go testing)
    RUN CGO_ENABLED=0 go test -v -cover -coverprofile=coverage.out \
        ./internal/clients/... ./config/...

    # Display coverage summary
    RUN go tool cover -func=coverage.out | tee coverage.txt

    # Calculate total coverage and ensure it meets minimum threshold
    RUN COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//') && \
        echo "Total Coverage: $COVERAGE%" && \
        if [ "$(echo "$COVERAGE < 40" | bc -l)" -eq 1 ]; then \
            echo "❌ Coverage $COVERAGE% is below minimum 40%"; \
            exit 1; \
        fi

    SAVE ARTIFACT coverage.out AS LOCAL coverage.out
    SAVE ARTIFACT coverage.txt AS LOCAL coverage.txt

test-examples:
    FROM +generate

    # Copy necessary files for example validation
    COPY --dir examples /app/providers/provider-upjet-tailscale/
    COPY examples_test.go /app/providers/provider-upjet-tailscale/
    WORKDIR /app/providers/provider-upjet-tailscale

    # Run example validation tests
    RUN CGO_ENABLED=0 go test -v ./examples_test.go

test-all:
    BUILD +test
    BUILD +test-examples

build:
    # Build on native platform (no QEMU) with cross-compilation
    ARG BUILDPLATFORM
    ARG GOOS=linux
    ARG GOARCH
    FROM --platform=$BUILDPLATFORM +generate

    # Build the provider binary with optimizations
    # -ldflags="-s -w" strips debug info and symbol table (saves ~15MB)
    # -trimpath removes file system paths from binary
    RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="-s -w" \
        -trimpath \
        -o bin/provider \
        cmd/provider/main.go

    SAVE ARTIFACT bin/provider AS LOCAL bin/provider

image:
    # Production runtime: tofu-runtime (base-runtime + tofu only, no debug tools)
    # Multi-platform build - TARGETPLATFORM/TARGETARCH are built-in and set by Earthly
    ARG TARGETPLATFORM
    ARG TARGETOS
    ARG TARGETARCH
    FROM --platform=$TARGETPLATFORM ghcr.io/millstonehq/tofu:runtime

    # Build the right binary for this platform, but compile on native arch (no QEMU)
    COPY (+build/provider --GOOS=$TARGETOS --GOARCH=$TARGETARCH) /usr/local/bin/provider

    ENTRYPOINT ["/usr/local/bin/provider"]

    ARG VERSION=v0.1.0
    # Save image for each platform
    SAVE IMAGE ghcr.io/millstonehq/provider-tailscale:${VERSION}
    SAVE IMAGE ghcr.io/millstonehq/provider-tailscale:latest

controller-tarball:
    # Build controller tarball for ARM64 (current cluster architecture)
    FROM alpine:latest
    RUN apk add docker-cli

    # Build ARM64 image first
    BUILD --platform=linux/arm64 +image

    # Load the ARM64 image and save it as tarball
    WITH DOCKER --load=ghcr.io/millstonehq/provider-tailscale:latest=+image --platform=linux/arm64
        RUN docker save ghcr.io/millstonehq/provider-tailscale:latest -o /tmp/controller.tar
    END

    SAVE ARTIFACT /tmp/controller.tar controller.tar

push-images:
    # Push multi-arch controller images to GHCR
    # Run with: earthly --push +push-images
    # Note: Requires docker login to ghcr.io (workflow does this)
    ARG VERSION=v0.1.0
    FROM alpine:latest

    RUN apk add docker-cli

    # Build and push both amd64 and arm64 images
    BUILD --platform=linux/amd64 --platform=linux/arm64 +image --VERSION=$VERSION

    # Create and push multi-arch manifest
    RUN docker buildx imagetools create -t ghcr.io/millstonehq/provider-tailscale:${VERSION} \
        -t ghcr.io/millstonehq/provider-tailscale:latest \
        ghcr.io/millstonehq/provider-tailscale:${VERSION}

push:
    # Push xpkg package with embedded ARM64 controller runtime to GHCR
    # Uses crossplane CLI to properly push OCI artifacts with embedded images
    # Run with: earthly --push +push --GITHUB_TOKEN=<token>
    FROM +builder-base

    ARG VERSION=v0.1.0
    ARG IMAGE_NAME=ghcr.io/millstonehq/provider-tailscale:latest
    ARG GITHUB_USER=millstonehq

    COPY +package-build/package.xpkg /tmp/provider-tailscale-package.xpkg

    # Use crossplane CLI to push xpkg with embedded runtime artifacts
    USER root
    RUN apk add docker-cli

    # Authenticate to GHCR using GitHub token passed as secret
    RUN --secret GITHUB_TOKEN \
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_USER" --password-stdin

    USER nonroot
    RUN crossplane xpkg push -f /tmp/provider-tailscale-package.xpkg $IMAGE_NAME

package-build:
    FROM +generate

    # Get ARM64 controller image tarball
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
