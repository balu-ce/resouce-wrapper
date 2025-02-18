# Kubernetes Resource Wrapper Operator

A basic Kubernetes operator written in Go without using operator frameworks.

## Building and Running

1. Build the operator:
```bash
go build -o operator .
```

2. Run locally (requires kubeconfig):
```bash
./operator --kubeconfig=$HOME/.kube/config
```

By using make command,
```bash
make generate
make run
```

3. Build Docker image:
```bash
docker build -t resource-wrapper:latest .
```

## Development

This is a basic Kubernetes operator template that you can extend by:

1. Adding Custom Resource Definitions (CRDs)
2. Implementing the reconciliation logic in `internal/controller/`
3. Adding informers and event handlers
4. Implementing your business logic

## Requirements

- Go 1.21 or higher
- Access to a Kubernetes cluster
- Docker (for building container images)