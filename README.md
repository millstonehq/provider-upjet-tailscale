# Crossplane Provider Tailscale

A Crossplane provider for managing Tailscale infrastructure declaratively using Kubernetes-style APIs.

## Overview

This provider enables you to manage Tailscale resources through Crossplane, bringing GitOps-style infrastructure management to your Tailscale tailnet.

### Supported Resources (Phase 1)

- **ACL** - Manage tailnet access control lists with HuJSON support
- **DNS Nameservers** - Configure custom DNS nameservers for your tailnet
- **Auth Keys** - Generate authentication keys with tags and policies
- **Device Tags** - Assign tags to devices in your tailnet
- **Device Authorization** - Approve or manage device authorizations

## Installation

### Prerequisites

- Kubernetes cluster with Crossplane installed (v1.14.0+)
- Tailscale account with API access
- Crossplane 2.0+ installed in your cluster

### Install the Provider

```bash
# Create the provider
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-tailscale
spec:
  package: xpkg.upbound.io/millstonehq/provider-tailscale:v0.1.0
EOF

# Verify installation
kubectl get providers
```

### Configure Authentication

1. **Create a Tailscale API Key**

   Visit the [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys) and create a new API key with appropriate permissions.

2. **Create a Kubernetes Secret**

   ```bash
   kubectl create secret generic tailscale-creds \
     --namespace crossplane-system \
     --from-literal=api_key='tskey-api-xxxxx'
   ```

3. **Create a ProviderConfig**

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: tailscale.upbound.io/v1beta1
   kind: ProviderConfig
   metadata:
     name: default
   spec:
     credentials:
       source: Secret
       secretRef:
         name: tailscale-creds
         namespace: crossplane-system
         key: api_key
   EOF
   ```

## Usage Examples

### ACL Management

```yaml
apiVersion: acl.tailscale.upbound.io/v1alpha1
kind: ACL
metadata:
  name: production-acl
spec:
  forProvider:
    acl: |
      {
        "groups": {
          "group:admin": ["admin@example.com"],
        },
        "acls": [
          {
            "action": "accept",
            "src": ["group:admin"],
            "dst": ["*:*"],
          },
        ],
      }
  providerConfigRef:
    name: default
```

### Generate Auth Keys

```yaml
apiVersion: tailnetkey.tailscale.upbound.io/v1alpha1
kind: Key
metadata:
  name: k8s-node-key
spec:
  forProvider:
    reusable: true
    preauthorized: true
    tags:
      - "tag:k8s"
      - "tag:production"
  writeConnectionSecretToRef:
    name: tailscale-k8s-key
    namespace: default
  providerConfigRef:
    name: default
```

### Configure DNS Nameservers

```yaml
apiVersion: dns.tailscale.upbound.io/v1alpha1
kind: Nameservers
metadata:
  name: custom-dns
spec:
  forProvider:
    nameservers:
      - "1.1.1.1"
      - "8.8.8.8"
  providerConfigRef:
    name: default
```

### Device Tag Management

```yaml
apiVersion: device.tailscale.upbound.io/v1alpha1
kind: Tags
metadata:
  name: web-server-tags
spec:
  forProvider:
    deviceId: "12345678901234567"
    tags:
      - "tag:production"
      - "tag:webserver"
  providerConfigRef:
    name: default
```

### Device Authorization

```yaml
apiVersion: device.tailscale.upbound.io/v1alpha1
kind: Authorization
metadata:
  name: approve-device
spec:
  forProvider:
    deviceId: "12345678901234567"
    authorized: true
  providerConfigRef:
    name: default
```

## Development

### Building from Source

```bash
# Clone the repository
cd providers/crossplane-provider-tailscale

# Initialize submodules (if using upjet build system)
make submodules

# Generate code
make generate

# Build the provider
make build

# Run locally (out-of-cluster)
make run
```

### Testing

```bash
# Run unit tests
go test -v ./...

# Run with a local Kubernetes cluster
kind create cluster
make run
```

## Architecture

This provider is built using:
- **Upjet v2.0.0** - Code generation framework for Terraform-based Crossplane providers
- **Crossplane Runtime v2.0.0** - Core Crossplane functionality
- **Terraform Provider Tailscale v0.18.0** - Underlying Terraform provider

### Authentication Methods

The provider supports two authentication methods:

1. **API Key** (Recommended for simplicity)
   ```yaml
   stringData:
     api_key: "tskey-api-xxxxx"
   ```

2. **OAuth Credentials** (For programmatic access)
   ```yaml
   stringData:
     oauth_client_id: "xxxxx"
     oauth_client_secret: "tskey-client-xxxxx"
   ```

## Contributing

Contributions are welcome! Please see the [Mill repository](https://github.com/millstonehq/mill) for contribution guidelines.

## License

Apache License 2.0 - Copyright 2024 Millstone Services LLC

## Support

For issues and questions:
- GitHub Issues: https://github.com/millstonehq/mill/issues
- Documentation: https://github.com/millstonehq/mill/tree/main/providers/crossplane-provider-tailscale

## Roadmap

### Phase 2 (Future)
- Tailnet settings management
- Device subnet routes
- Webhook configurations
- Policy files
- User management

## References

- [Tailscale API Documentation](https://tailscale.com/api)
- [Crossplane Documentation](https://docs.crossplane.io)
- [Upjet Documentation](https://github.com/crossplane/upjet)
