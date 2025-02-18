# Kubernetes Resource Wrapper Operator

A basic Kubernetes operator written in Go without using operator frameworks.

## Building and Running

1. Build the operator:
```bash
go build -o operator .
```

2. Run locally (requires kubeconfig):
```bash
make generate
make run
```

3. Build Docker image:
```bash
docker build -t resource-wrapper:latest .
```

## Requirements

- Go 1.21 or higher
- Access to a Kubernetes cluster
- Docker (for building container images)