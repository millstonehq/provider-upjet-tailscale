# Crossplane Provider Tailscale Examples

This directory contains example manifests for all supported Tailscale resources.

## Getting Started

### 1. Install the Provider

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-tailscale
spec:
  package: ghcr.io/millstonehq/provider-upjet-tailscale:latest
EOF
```

### 2. Configure Authentication

Create a Tailscale API key secret:

```bash
kubectl apply -f providerconfig/secret.yaml
kubectl apply -f providerconfig/providerconfig.yaml
```

**Note:** Replace `tskey-api-xxxxx` in `secret.yaml` with your actual Tailscale API key.

## Available Examples

### Provider Configuration

- **[providerconfig/secret.yaml](providerconfig/secret.yaml)** - Kubernetes Secret containing Tailscale credentials
- **[providerconfig/providerconfig.yaml](providerconfig/providerconfig.yaml)** - ProviderConfig referencing the credentials secret

### ACL Management

- **[acl/acl.yaml](acl/acl.yaml)** - Complete ACL configuration with groups, hosts, and rules using HuJSON format

### DNS Configuration

- **[dns/nameservers.yaml](dns/nameservers.yaml)** - Configure custom DNS nameservers for your tailnet

### Authentication Keys

- **[tailnetkey/key.yaml](tailnetkey/key.yaml)** - Generate reusable authentication keys with tags and policies

### Device Management

- **[device/authorization.yaml](device/authorization.yaml)** - Approve or manage device authorizations
- **[device/tags.yaml](device/tags.yaml)** - Assign tags to devices in your tailnet

## Usage Pattern

Most resources follow this pattern:

```yaml
apiVersion: <group>.tailscale.upbound.io/v1alpha1
kind: <ResourceKind>
metadata:
  name: example-resource
spec:
  forProvider:
    # Resource-specific configuration
  providerConfigRef:
    name: default  # References the ProviderConfig
```

## Testing Examples

To test an example:

```bash
# Apply the manifest
kubectl apply -f <path-to-example>.yaml

# Check resource status
kubectl get <resource-kind> <name> -o yaml

# View conditions
kubectl describe <resource-kind> <name>
```

## Writing Connection Secrets

Some resources (like `tailnetkey.Key`) support writing sensitive data to Kubernetes Secrets:

```yaml
spec:
  writeConnectionSecretToRef:
    name: my-secret-name
    namespace: default
```

## Common Issues

### Authentication Errors

If you see authentication errors:
1. Verify your API key is correct in the secret
2. Check that the ProviderConfig references the correct secret
3. Ensure the API key has appropriate permissions

### Resource Not Syncing

Check the provider pod logs:
```bash
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale
```

## Additional Resources

- [Tailscale API Documentation](https://tailscale.com/api)
- [Provider Documentation](../README.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [GitHub Issues](https://github.com/millstonehq/provider-upjet-tailscale/issues)
