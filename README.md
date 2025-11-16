# Crossplane Provider Tailscale

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub release](https://img.shields.io/github/release/millstonehq/provider-upjet-tailscale.svg)](https://github.com/millstonehq/provider-upjet-tailscale/releases)

A Crossplane provider for managing Tailscale infrastructure declaratively using Kubernetes-style APIs.

## Overview

This provider enables you to manage Tailscale resources through Crossplane, bringing GitOps-style infrastructure management to your Tailscale tailnet.

### Supported Resources

- **ACL** - Manage tailnet access control lists with HuJSON support
- **DNS Nameservers** - Configure custom DNS nameservers for your tailnet
- **Tailnet Keys** - Generate authentication keys with tags and policies
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
  package: ghcr.io/millstonehq/provider-tailscale:latest
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

This provider uses [Earthly](https://earthly.dev) for building and testing.

```bash
# Generate code
earthly +generate

# Build the provider
earthly +build

# Run tests
earthly +test

# Test with examples
earthly +test-examples

# Run all tests (unit + examples)
earthly +test-all

# Build and push images (requires authentication)
earthly --push +push
```

### Local Development

```bash
# Build provider package locally
earthly +package-local

# Install in your cluster
kubectl apply -f examples/providerconfig/
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

## Community & Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, or improving documentation, your help is appreciated.

### How to Contribute

1. **Fork the repository** on GitHub
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** and commit them (`git commit -m 'feat: add amazing feature'`)
4. **Push to your branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Development setup and building from source
- Code style and conventions
- Testing requirements
- PR submission process

### Getting Help

- 📚 [Documentation](https://github.com/millstonehq/provider-upjet-tailscale/tree/main/examples)
- 🐛 [Report a Bug](https://github.com/millstonehq/provider-upjet-tailscale/issues/new?labels=bug)
- 💡 [Request a Feature](https://github.com/millstonehq/provider-upjet-tailscale/issues/new?labels=enhancement)
- 💬 [Discussions](https://github.com/millstonehq/provider-upjet-tailscale/discussions)

### Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright 2025 Millstone Partners, LLC

## Support

For issues and questions:
- GitHub Issues: https://github.com/millstonehq/provider-upjet-tailscale/issues
- Documentation: https://github.com/millstonehq/provider-upjet-tailscale/tree/main/examples
- Discussions: https://github.com/millstonehq/provider-upjet-tailscale/discussions

## References

- [Tailscale API Documentation](https://tailscale.com/api)
- [Crossplane Documentation](https://docs.crossplane.io)
- [Upjet Documentation](https://github.com/crossplane/upjet)
