# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
GOBIN=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN=$(shell go env GOPATH)/bin
endif

.PHONY: all
all: build

.PHONY: generate
generate: controller-gen ## Generate code
	$(CONTROLLER_GEN) object paths=./...
	$(CONTROLLER_GEN) crd paths=./api/... output:crd:dir=./config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths=./internal/... output:rbac:dir=./config/rbac

.PHONY: build
build: generate ## Build the project
	go mod tidy
	go mod vendor
	go build -o bin/manager ./...

.PHONY: run
run: manifests ## Run operator locally
	go run ./cmd/main.go

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	kubectl apply -f config/crd/bases
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/rbac/role-binding.yaml
	kubectl apply -f config/rbac/service_account.yaml
	
.PHONY: controller-gen
controller-gen: ## Download controller-gen if necessary
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@latest)

CONTROLLER_GEN=$(GOBIN)/controller-gen

# go-get-tool will 'go install' any package with custom version if needed
define go-get-tool
@[ -f $(1) ] || { \
set -e; \
echo "Downloading $(2)"; \
GOBIN=$(GOBIN) go install $(2); \
}
endef

.PHONY: clean
clean: ## Remove generated files
	rm -f bin/manager
	rm -rf config/rbac/*.yaml