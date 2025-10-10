# Crossplane Provider Tailscale - Deployment Guide

This guide walks through deploying the Crossplane Tailscale provider to the Mill infrastructure.

## Prerequisites

- Mill repository cloned locally
- `mill-context` configured for `mgmt-prod`
- Tailscale API key with appropriate permissions
- SOPS age key for `SOPS_AGE_KEY_MGMT_PROD`

## Step 1: Build the Provider

### Generate Code and Build

```bash
# Set context for mgmt-prod
mill-context use mgmt-prod

# Generate provider code from Terraform provider
earthly +crossplane-provider-tailscale-generate

# Build the provider binary
earthly +crossplane-provider-tailscale-build

# Build and push the provider image
earthly +crossplane-provider-tailscale-image --VERSION=v0.1.0
```

This will:
1. Run upjet code generation to create CRDs and controllers
2. Build the provider binary for linux/amd64
3. Package it in a distroless container
4. Push to `xpkg.upbound.io/millstonehq/provider-tailscale-controller:v0.1.0`

## Step 2: Create Tailscale API Key

1. Visit [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
2. Click "Generate API Key"
3. Select appropriate permissions:
   - ✅ Devices (read/write) - for device management
   - ✅ DNS (read/write) - for DNS configuration
   - ✅ ACLs (read/write) - for ACL management
   - ✅ Auth keys (read/write) - for key generation
4. Copy the generated key (starts with `tskey-api-`)

## Step 3: Encrypt Secrets with SOPS

```bash
# Navigate to the secrets directory
cd deploy/mgmt/prod/applications/crossplane-provider-tailscale/oci/us-phoenix-1/mgmt-prod-usp1-1

# Create the secrets file (all values must be base64 encoded)
cat > tailscale-creds.enc << EOF
{
  "TAILSCALE_API_KEY": "$(echo -n 'tskey-api-YOUR_KEY_HERE' | base64)",
  "TAILSCALE_TAILNET": "$(echo -n 'your-tailnet.ts.net' | base64)"
}
EOF

# Encrypt with SOPS (must be in target directory for path_regex matching)
sops -e -i tailscale-creds.enc

# Verify encryption
sops -d tailscale-creds.enc
```

## Step 4: Create ArgoCD Secret Manifest

The secret will be decrypted by ArgoCD Vault Plugin and created in the cluster:

```bash
# Create a secret manifest that references the encrypted JSON
cat > manifests/tailscale-secret.yaml << 'EOF'
apiVersion: v1
kind: Secret
metadata:
  name: tailscale-creds
  namespace: crossplane-system
type: Opaque
data:
  api_key: <path:tailscale-creds.enc#TAILSCALE_API_KEY | base64decode>
  tailnet: <path:tailscale-creds.enc#TAILSCALE_TAILNET | base64decode>
EOF
```

## Step 5: Deploy via ArgoCD

The provider is deployed via the ApplicationSet at:
`deploy/mgmt/prod/applicationsets/crossplane-provider-tailscale.yaml`

### Verify ApplicationSet

```bash
# Check if ApplicationSet exists
kubectl get applicationset -n argocd crossplane-provider-tailscale

# View the generated Application
kubectl get application -n argocd crossplane-provider-tailscale -o yaml
```

### Manual Sync (if needed)

```bash
# Sync the application
argocd app sync crossplane-provider-tailscale

# Watch the sync progress
argocd app wait crossplane-provider-tailscale --health
```

## Step 6: Verify Installation

### Check Provider Status

```bash
# Verify provider is installed
kubectl get providers

# Expected output:
# NAME                   INSTALLED   HEALTHY   PACKAGE                                                     AGE
# provider-tailscale     True        True      xpkg.upbound.io/millstonehq/provider-tailscale:v0.1.0      1m

# Check provider pod
kubectl get pods -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale

# View provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale --tail=50
```

### Verify CRDs are Installed

```bash
# List all Tailscale CRDs
kubectl get crds | grep tailscale

# Expected CRDs:
# acls.acl.tailscale.upbound.io
# authorizations.device.tailscale.upbound.io
# keys.tailnetkey.tailscale.upbound.io
# nameservers.dns.tailscale.upbound.io
# providerconfigs.tailscale.upbound.io
# tags.device.tailscale.upbound.io
```

### Verify ProviderConfig

```bash
# Check ProviderConfig status
kubectl get providerconfig default -o yaml

# Expected status.conditions:
# - type: Ready
#   status: "True"
```

## Step 7: Test with Example Resources

### Create a Test Auth Key

```bash
kubectl apply -f - <<EOF
apiVersion: tailnetkey.tailscale.upbound.io/v1alpha1
kind: Key
metadata:
  name: test-key
spec:
  forProvider:
    reusable: true
    preauthorized: true
    tags:
      - "tag:test"
    description: "Test key from Crossplane"
  writeConnectionSecretToRef:
    name: test-tailscale-key
    namespace: default
  providerConfigRef:
    name: default
EOF

# Check the resource status
kubectl get keys.tailnetkey.tailscale.upbound.io test-key

# View the generated secret
kubectl get secret test-tailscale-key -n default -o jsonpath='{.data.key}' | base64 -d
```

### Verify in Tailscale Admin Console

1. Visit [Tailscale Keys](https://login.tailscale.com/admin/settings/keys)
2. Verify the "Test key from Crossplane" appears in the list
3. Check that tags are applied correctly

## Step 8: Deploy Production ACLs

```bash
# Apply ACL configuration
kubectl apply -f - <<EOF
apiVersion: acl.tailscale.upbound.io/v1alpha1
kind: ACL
metadata:
  name: production-acl
spec:
  forProvider:
    acl: |
      {
        "groups": {
          "group:k8s-nodes": ["tag:k8s"],
          "group:admin": ["admin@example.com"],
        },
        "acls": [
          {
            "action": "accept",
            "src": ["group:admin"],
            "dst": ["*:*"],
          },
          {
            "action": "accept",
            "src": ["group:k8s-nodes"],
            "dst": ["group:k8s-nodes:*"],
          },
        ],
      }
  providerConfigRef:
    name: default
EOF

# Verify ACL is synced
kubectl get acls.acl.tailscale.upbound.io production-acl
```

## Troubleshooting

### Provider Pod Not Starting

```bash
# Check pod events
kubectl describe pod -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale

# Check provider installation
kubectl describe provider provider-tailscale

# View detailed logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale --all-containers=true
```

### Authentication Errors

```bash
# Verify secret exists and is decrypted
kubectl get secret tailscale-creds -n crossplane-system

# Check secret contents (decode base64)
kubectl get secret tailscale-creds -n crossplane-system -o jsonpath='{.data.api_key}' | base64 -d

# Verify ProviderConfig references correct secret
kubectl get providerconfig default -o yaml
```

### Resource Not Syncing

```bash
# Check resource status and conditions
kubectl describe <resource-kind> <resource-name>

# View provider controller logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale --tail=100

# Check for rate limiting or API errors
kubectl get events --sort-by='.lastTimestamp' | grep tailscale
```

### SOPS Decryption Issues

```bash
# Verify SOPS age key is set
echo $SOPS_AGE_KEY_MGMT_PROD | head -c 20

# Test decryption manually
cd deploy/mgmt/prod/applications/crossplane-provider-tailscale/oci/us-phoenix-1/mgmt-prod-usp1-1
export SOPS_AGE_KEY="$SOPS_AGE_KEY_MGMT_PROD"
sops -d tailscale-creds.enc.json

# Check ArgoCD vault plugin status
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-repo-server | grep sops
```

## Updating the Provider

### Update Provider Version

```bash
# Build new version
earthly +crossplane-provider-tailscale-image --VERSION=v0.2.0

# Update manifests
vim deploy/mgmt/prod/applications/crossplane-provider-tailscale/oci/us-phoenix-1/mgmt-prod-usp1-1/manifests/provider.yaml
# Change: package: xpkg.upbound.io/millstonehq/provider-tailscale:v0.2.0

# Commit and push
git add .
git commit -m "feat(crossplane): update tailscale provider to v0.2.0"
git push

# ArgoCD will auto-sync and update the provider
```

### Rolling Back

```bash
# Revert to previous version
git revert HEAD
git push

# Or manually update
kubectl patch provider provider-tailscale --type='merge' -p '{"spec":{"package":"xpkg.upbound.io/millstonehq/provider-tailscale:v0.1.0"}}'
```

## Monitoring

### Key Metrics to Monitor

- Provider pod health and restarts
- Resource sync success rate
- Tailscale API rate limits
- Authentication failures

### Useful Commands

```bash
# Watch all Tailscale resources
watch kubectl get acls,keys,nameservers,tags,authorizations

# Monitor provider health
kubectl get providers -w

# View recent events
kubectl get events -n crossplane-system --sort-by='.lastTimestamp' | tail -20
```

## Security Considerations

1. **API Key Rotation**: Rotate Tailscale API keys regularly
2. **SOPS Keys**: Never commit unencrypted SOPS age keys
3. **Least Privilege**: Grant minimal required permissions to API keys
4. **Audit Logs**: Monitor Tailscale audit logs for provider actions
5. **Secret Access**: Restrict access to crossplane-system namespace

## Next Steps

After successful deployment:

1. Create additional ProviderConfigs for different tailnets (if needed)
2. Deploy production ACL configurations
3. Set up device auto-tagging policies
4. Configure DNS nameservers for the tailnet
5. Generate auth keys for Kubernetes nodes

## References

- [Mill Infrastructure Repo](https://github.com/millstonehq/mill)
- [Provider README](./README.md)
- [Crossplane Documentation](https://docs.crossplane.io)
- [Tailscale API Docs](https://tailscale.com/api)
