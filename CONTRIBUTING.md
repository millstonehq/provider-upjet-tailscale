# Contributing to Crossplane Provider Tailscale

Thank you for your interest in contributing to the Crossplane Provider Tailscale! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Adding New Resources](#adding-new-resources)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/provider-upjet-tailscale.git
   cd provider-upjet-tailscale
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/millstonehq/provider-upjet-tailscale.git
   ```

## Development Setup

### Prerequisites

- Go 1.24 or later
- Docker (for building container images)
- Kubernetes cluster (kind, minikube, or similar for testing)
- Crossplane v2.0.0+ installed in your cluster

### Install Dependencies

```bash
# Download Go dependencies
go mod download

# Install code generation tools
go install golang.org/x/tools/cmd/goimports@latest
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
go install github.com/crossplane/crossplane-tools/cmd/angryjet@latest
```

### Building Locally

```bash
# Generate code (APIs, controllers, CRDs)
go run cmd/generator/main.go "$(pwd)"

# Generate DeepCopy methods
controller-gen object:headerFile=hack/boilerplate.go.txt paths="./apis/..."

# Generate resource methods
angryjet generate-methodsets --header-file=hack/boilerplate.go.txt ./apis/...

# Generate CRDs
controller-gen crd:allowDangerousTypes=true paths="./apis/..." output:crd:artifacts:config=package/crds

# Build the provider binary
go build -o bin/provider cmd/provider/main.go

# Run tests
go test -v ./...
```

### Running Locally (Development Mode)

```bash
# Run the provider against your kubeconfig context
go run cmd/provider/main.go --debug
```

## Making Changes

### Branch Naming

Use descriptive branch names with prefixes:
- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications

Example: `feat/add-device-posture-resource`

### Commit Messages

Follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

**Example:**
```
feat(device): add device posture integration resource

Add support for managing device posture integrations with
providers like Intune, CrowdStrike, and Kolide.

Closes #42
```

## Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./internal/controller/device/...

# Run with coverage
go test -cover ./...
```

### Integration Tests

To test against a live Tailscale account:

1. Create a test tailnet (recommended)
2. Generate an API key with appropriate permissions
3. Create a secret in your test cluster:
   ```bash
   kubectl create secret generic tailscale-creds \
     --namespace crossplane-system \
     --from-literal=api_key='tskey-api-xxxxx'
   ```
4. Apply test manifests from `examples/` directory

### Manual Testing

```bash
# Install provider in local cluster
kubectl apply -f examples/providerconfig/

# Test a resource
kubectl apply -f examples/dns/nameservers.yaml

# Check resource status
kubectl get nameservers -o yaml

# View controller logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-tailscale
```

## Submitting Changes

### Pull Request Process

1. **Update your fork** with latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Open a Pull Request** on GitHub with:
   - Clear title and description
   - Reference to related issues (e.g., "Closes #123")
   - Description of changes and testing performed
   - Screenshots or examples if applicable

4. **Address review feedback** promptly

5. **Ensure CI passes** - all checks must pass before merge

### PR Requirements

- [ ] Code follows project style guidelines
- [ ] Tests added/updated for new functionality
- [ ] Documentation updated (README, examples, etc.)
- [ ] CRDs regenerated if API changes made
- [ ] Commit messages follow conventional format
- [ ] No breaking changes (or clearly documented if unavoidable)

## Code Style

### Go Code

- Follow standard Go conventions and [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Keep functions small and focused
- Add comments for exported types and functions

### Resource Configuration

When adding or modifying Upjet resource configurations in `config/`:

```go
// config/myresource/config.go
package myresource

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure adds configuration for myresource resources.
func Configure(p *config.Provider) {
    p.AddResourceConfigurator("tailscale_my_resource", func(r *config.Resource) {
        // Use appropriate ExternalName strategy
        r.ExternalName = config.IdentifierFromProvider
        
        // Set short group name for API group
        r.ShortGroup = "mygroup"
        
        // Add references to other resources if needed
        // r.References["other_resource_id"] = config.Reference{...}
    })
}
```

### Example Manifests

When adding examples in `examples/`:

- Use clear, descriptive metadata names
- Include inline comments explaining each field
- Use realistic but generic values
- Follow YAML formatting conventions

## Adding New Resources

To add support for a new Tailscale resource:

1. **Check the Terraform provider** for the resource definition:
   - https://registry.terraform.io/providers/tailscale/tailscale/latest/docs

2. **Add configuration** in `config/<group>/config.go`:
   ```go
   p.AddResourceConfigurator("tailscale_new_resource", func(r *config.Resource) {
       r.ShortGroup = "group"
       r.ExternalName = config.IdentifierFromProvider
   })
   ```

3. **Update provider.go** to include the new group:
   ```go
   import "github.com/millstonehq/provider-upjet-tailscale/config/newgroup"
   
   // In GetProvider():
   newgroup.Configure(pc)
   ```

4. **Generate code**:
   ```bash
   go run cmd/generator/main.go "$(pwd)"
   controller-gen object:headerFile=hack/boilerplate.go.txt paths="./apis/..."
   angryjet generate-methodsets --header-file=hack/boilerplate.go.txt ./apis/...
   controller-gen crd:allowDangerousTypes=true paths="./apis/..." output:crd:artifacts:config=package/crds
   ```

5. **Add example manifest** in `examples/<group>/`:
   ```yaml
   apiVersion: <group>.tailscale.upbound.io/v1alpha1
   kind: NewResource
   metadata:
     name: example-new-resource
   spec:
     forProvider:
       # Resource-specific fields
     providerConfigRef:
       name: default
   ```

6. **Test the resource** in a live environment

7. **Update documentation**:
   - Add resource to README.md
   - Document any special configuration in RESOURCES.md

## Questions?

- Open a [Discussion](https://github.com/millstonehq/provider-upjet-tailscale/discussions) for general questions
- Create an [Issue](https://github.com/millstonehq/provider-upjet-tailscale/issues) for bugs or feature requests

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
