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
    RUN echo '{"terraform":[{"required_providers":[{"tailscale":{"source":"tailscale/tailscale","version":"0.22.0"}}]}]}' > main.tf.json && \
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
    # TARGETARCH is built-in and set automatically by Earthly based on --platform
    ARG TARGETARCH
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
        -ldflags="-s -w" \
        -trimpath \
        -o bin/provider \
        cmd/provider/main.go

    SAVE ARTIFACT bin/provider AS LOCAL bin/provider

image:
    # Production runtime: terraform-runtime (base-runtime + tofu only, no debug tools)
    # Multi-platform build - TARGETPLATFORM/TARGETARCH are built-in and set by Earthly
    ARG TARGETPLATFORM
    ARG TARGETARCH
    FROM --platform=$TARGETPLATFORM ../../lib/build-config/terraform/+terraform-runtime

    COPY +build/provider /usr/local/bin/provider

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
    # Run with: earthly --push +push-images --GITHUB_TOKEN=$GITHUB_TOKEN
    ARG VERSION=v0.1.0
    ARG GITHUB_TOKEN
    FROM alpine:latest

    RUN apk add docker-cli

    # Login to GHCR
    RUN echo "$GITHUB_TOKEN" | docker login ghcr.io -u millstone-bot --password-stdin

    # Build and push both amd64 and arm64 images
    BUILD --platform=linux/amd64 --platform=linux/arm64 +image --VERSION=$VERSION

    # Create and push multi-arch manifest
    RUN docker buildx imagetools create -t ghcr.io/millstonehq/provider-tailscale:${VERSION} \
        -t ghcr.io/millstonehq/provider-tailscale:latest \
        ghcr.io/millstonehq/provider-tailscale:${VERSION}

push:
    # Push xpkg package with embedded ARM64 controller runtime to GHCR
    # Uses crossplane CLI to properly push OCI artifacts with embedded images
    # Run with: earthly --push +push --GITHUB_TOKEN=$GITHUB_TOKEN
    #
    # ⚠️  SECURITY NOTE: This target uses ARG GITHUB_TOKEN which bakes the token into
    # this ephemeral alpine image's layers. This image is NEVER pushed (no SAVE IMAGE).
    # Only the pre-built xpkg package (which doesn't contain the token) is pushed.
    # DO NOT add SAVE IMAGE to this target.
    FROM +builder-base

    ARG VERSION=v0.1.0
    ARG IMAGE_NAME=ghcr.io/millstonehq/provider-tailscale:latest
    ARG GITHUB_TOKEN

    COPY +package-build/package.xpkg /tmp/provider-tailscale-package.xpkg

    # Use crossplane CLI to push xpkg with embedded runtime artifacts
    # crossplane CLI uses docker credentials, so login with docker first
    USER root
    RUN apk add docker-cli
    USER nonroot
    RUN echo "$GITHUB_TOKEN" | docker login ghcr.io -u millstone-bot --password-stdin && \
        crossplane xpkg push -f /tmp/provider-tailscale-package.xpkg $IMAGE_NAME

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
