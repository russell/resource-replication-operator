# Resource Replication Operator

A Kubernetes operator that replicates resources across namespaces. Simplifies managing shared resources like TLS certificates, configuration data, and other secrets that need to be available in multiple namespaces.

## Features

- **Cross-namespace resource replication** - Copy secrets and other resources between namespaces
- **Automatic updates** - Resources are updated when the source changes
- **Owner references** - Replicated resources are cleaned up when the ReplicatedResource is deleted
- **Status tracking** - Monitor replication status with Kubernetes-native conditions

## Supported Resource Types

- **Secrets** - TLS certificates, authentication tokens, API keys
- ConfigMaps (planned)
- Custom resources (planned)

## Quick Start

### Installation

Deploy the operator to your cluster:

```bash
kubectl apply -f https://github.com/russell/resource-replication-operator/releases/latest/download/install.yaml
```

### Usage

Create a source secret in one namespace:

```bash
kubectl create secret generic my-secret \
  --from-literal=username=admin \
  --from-literal=password=secretpassword \
  -n certificates
```

Replicate it to another namespace:

```yaml
apiVersion: utils.simopolis.xyz/v1alpha1
kind: ReplicatedResource
metadata:
  name: replicated-credentials
  namespace: app-namespace
spec:
  source:
    namespace: certificates
    kind: Secret
    name: my-secret
```

The operator will create a secret named `replicated-credentials` in the `app-namespace` with the same data as the source secret.

## Configuration

### ReplicatedResource Spec

```yaml
spec:
  source:
    namespace: string    # Source namespace
    name: string         # Source resource name
    kind: string         # Resource type (Secret, ConfigMap)
```

### Status Conditions

The operator provides status information about replication:

```yaml
status:
  phase: "Completed" | "Failed"
  conditions:
  - type: Complete
    status: "True"
    reason: Replicated
    message: Successfully Replicated
```

## Development

### Prerequisites

- Go 1.24+
- Docker
- kubectl
- Access to a Kubernetes cluster (1.30+)

### Local Development

```bash
# Install dependencies
make manifests generate fmt vet

# Run tests
make test

# Run locally (requires cluster access)
make install run

# Build and deploy
make docker-build docker-push deploy
```

### Building

```bash
# Build manager binary
make build

# Build container image
make docker-build IMG=myregistry/resource-replication-operator:latest
```

## Architecture

The operator consists of:

- **ReplicatedResource Controller** - Watches for ReplicatedResource CRDs and orchestrates replication
- **Resource Replicators** - Implement replication logic for specific resource types (Secrets, ConfigMaps)
- **Field Indexing** - Enables efficient lookups for source resource changes
- **Network Policies** - Optional security policies to restrict traffic to metrics and webhook endpoints

## Technical Details

Built with:
- **Operator SDK v1.41.0** - Latest Kubernetes operator framework
- **Go 1.24** - Modern Go runtime with enhanced performance
- **Kubernetes API v0.33.0** - Support for latest Kubernetes features
- **Controller Runtime v0.21.0** - Advanced controller patterns and observability
- **Kubebuilder v4 Layout** - Modern project structure and scaffolding

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
