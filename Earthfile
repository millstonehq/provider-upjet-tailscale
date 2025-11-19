VERSION 0.8
PROJECT millstonehq/mill

# Crossplane Provider Tailscale Build Pipeline
# Self-contained build pipeline for the Tailscale Crossplane provider

builder-base:
    ARG BUILDPLATFORM
    # Use pre-built crossplane:builder with Go, OpenTofu, make, and pre-compiled tools
    # (goimports, controller-gen, angryjet, crossplane CLI)
    FROM --platform=$BUILDPLATFORM ghcr.io/millstonehq/crossplane:builder

    # Install Upbound up CLI for multi-arch xpkg push support
    USER root
    RUN curl -sL "https://cli.upbound.io" | sh && \
        mv up /usr/local/bin/up && \
        up version
    USER nonroot

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
    # Allow manual override for cross-compilation (used by controller-tarball)
    ARG IMAGE_OS=${TARGETOS:-linux}
    ARG IMAGE_ARCH=${TARGETARCH:-amd64}
    ARG IMAGE_PLATFORM=${TARGETPLATFORM:-linux/${IMAGE_ARCH}}
    FROM --platform=$IMAGE_PLATFORM ghcr.io/millstonehq/tofu:runtime

    # Build the right binary for this platform, but compile on native arch (no QEMU)
    COPY (+build/provider --GOOS=$IMAGE_OS --GOARCH=$IMAGE_ARCH) /usr/local/bin/provider

    ENTRYPOINT ["/usr/local/bin/provider"]

    ARG VERSION=v0.1.0
    ARG IMAGE_SUFFIX=""
    # Save image for each platform with optional suffix (e.g. -runtime, -xpkg)
    SAVE IMAGE --push ghcr.io/millstonehq/provider-tailscale${IMAGE_SUFFIX}:${VERSION}
    SAVE IMAGE --push ghcr.io/millstonehq/provider-tailscale${IMAGE_SUFFIX}:latest


push-runtime:
    # Push multi-arch controller runtime images to GHCR (-runtime tag)
    # This will be merged with xpkg package later to create final multi-manifest artifact
    # Run with: earthly --push +push-runtime
    ARG VERSION=v0.1.0

    # Build and push both amd64 and arm64 images to -runtime tag
    BUILD --platform=linux/amd64 --platform=linux/arm64 +image --VERSION=$VERSION --IMAGE_SUFFIX=-runtime

push:
    # Push complete multi-manifest package to GHCR
    # Creates single OCI artifact with 3 manifests (amd64 runtime, arm64 runtime, xpkg package)
    # Run with: earthly --push +push
    BUILD +push-runtime
    BUILD +push-package
    BUILD +merge-manifests

push-package:
    # Push xpkg package to GHCR (-xpkg tag)
    # This will be merged with runtime images later to create final multi-manifest artifact
    # Run with: earthly --push +push-package (requires -P for WITH DOCKER)
    FROM +builder-base

    COPY +package-build/package.xpkg /tmp/provider-tailscale-package.xpkg

    # Install docker before WITH DOCKER (prevents auto-install attempt)
    USER root
    RUN apk add docker docker-cli-buildx

    # Use WITH DOCKER to access docker daemon for authentication
    # Note: WITH DOCKER only allows a single RUN command
    ARG VERSION=v0.1.0
    ARG IMAGE_NAME=ghcr.io/millstonehq/provider-tailscale-xpkg:latest
    ARG GITHUB_USER=millstonehq
    WITH DOCKER
        RUN --secret GITHUB_TOKEN \
            mkdir -p /root/.docker && \
            auth=$(printf '%s:%s' "$GITHUB_USER" "$GITHUB_TOKEN" | base64 | tr -d '\n') && \
            printf '{"auths":{"ghcr.io":{"auth":"%s"}}}' "$auth" > /root/.docker/config.json && \
            up xpkg push -f /tmp/provider-tailscale-package.xpkg $IMAGE_NAME
    END

merge-manifests:
    # Merge runtime images (amd64+arm64) and xpkg package into single multi-manifest artifact
    # Uses docker buildx imagetools to create combined manifest list
    # Run with: earthly --push +merge-manifests (after +push-runtime and +push-package)
    FROM ghcr.io/millstonehq/base:builder

    # Install docker before WITH DOCKER (prevents auto-install attempt)
    USER root
    RUN apk add docker docker-cli-buildx

    # Use WITH DOCKER to get access to Docker daemon for buildx imagetools
    # Note: WITH DOCKER only allows a single RUN command
    ARG VERSION=v0.1.0
    ARG GITHUB_USER=millstonehq
    WITH DOCKER
        RUN --secret GITHUB_TOKEN \
            mkdir -p /root/.docker && \
            auth=$(printf '%s:%s' "$GITHUB_USER" "$GITHUB_TOKEN" | base64 | tr -d '\n') && \
            printf '{"auths":{"ghcr.io":{"auth":"%s"}}}' "$auth" > /root/.docker/config.json && \
            docker buildx imagetools create \
                --tag ghcr.io/millstonehq/provider-tailscale:latest \
                --tag ghcr.io/millstonehq/provider-tailscale:$VERSION \
                ghcr.io/millstonehq/provider-tailscale-runtime:latest \
                ghcr.io/millstonehq/provider-tailscale-xpkg:latest
    END

package-build:
    FROM +generate

    # Build xpkg package (references external multi-arch runtime images)
    # Controller images are published separately via +push-runtime
    RUN crossplane xpkg build \
        --package-root=package \
        -o package.xpkg

    SAVE ARTIFACT package.xpkg

package-local:
    ARG IMAGE_NAME=provider-tailscale:latest

    # Load the xpkg tarball into docker
    LOCALLY
    COPY +package-build/package.xpkg /tmp/provider-tailscale-package.xpkg
    RUN LOADED_ID=$(docker load -i /tmp/provider-tailscale-package.xpkg 2>&1 | grep -oP 'Loaded image ID: \K.*') && \
        docker tag $LOADED_ID $IMAGE_NAME
