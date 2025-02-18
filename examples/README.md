# NamespaceClass Examples

This directory contains example manifests for the NamespaceClass operator.

## Available Examples

### 1. Internal Network NamespaceClass
File: `restricted-namespaceclass.yaml`

This example creates an internal network environment with:
- Network Policy that only allows DNS traffic to/from kube-system
- ServiceAccount with automountServiceAccountToken disabled

### 2. Public Network NamespaceClass
File: `open-namespaceclass.yaml`

This example creates a public network environment with:
- Network Policy that allows all traffic
- ServiceAccount with automountServiceAccountToken enabled

### 3. Example Namespaces
File: `example-namespace.yaml`

Contains example namespaces that use the above NamespaceClasses.

## Usage

1. Apply the NamespaceClasses:
```bash
kubectl apply -f restricted-namespaceclass.yaml
kubectl apply -f open-namespaceclass.yaml
```

2. Create namespaces that use these classes:
```bash
kubectl apply -f example-namespace.yaml
```

3. Verify the resources:
```bash
# Check NetworkPolicies
kubectl get networkpolicies -n example-internal
kubectl get networkpolicies -n example-public

# Check ServiceAccounts
kubectl get serviceaccounts -n example-internal
kubectl get serviceaccounts -n example-public
```
