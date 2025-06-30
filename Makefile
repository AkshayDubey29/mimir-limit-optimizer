# Makefile for mimir-limit-optimizer

# Variables
PROJECT_NAME := mimir-limit-optimizer
REGISTRY ?= your-registry.com
IMAGE_NAME := $(REGISTRY)/$(PROJECT_NAME)
VERSION ?= latest
NAMESPACE ?= mimir-limit-optimizer
MIMIR_NAMESPACE ?= mimir-system

# Go variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

# Build info
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Mimir Limit Optimizer - Build and Deployment Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: build
build: ## Build the binary
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(GOBIN)
	@go build $(LDFLAGS) -o $(GOBIN)/$(PROJECT_NAME) ./main.go

.PHONY: run
run: ## Run the application locally
	@echo "Running $(PROJECT_NAME) locally..."
	@go run ./main.go --config=config.yaml --log-level=debug

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: mod-tidy
mod-tidy: ## Clean up go.mod
	@echo "Tidying go modules..."
	@go mod tidy

##@ Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image: $(IMAGE_NAME):$(VERSION)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(PROJECT_NAME):$(VERSION) \
		-t $(IMAGE_NAME):$(VERSION) \
		.

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "Pushing Docker image: $(IMAGE_NAME):$(VERSION)"
	@docker push $(IMAGE_NAME):$(VERSION)

.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --rm -it \
		-v ~/.kube:/home/nonroot/.kube:ro \
		-v $(PWD)/config.yaml:/config/config.yaml:ro \
		$(PROJECT_NAME):$(VERSION) \
		--config=/config/config.yaml \
		--log-level=debug

.PHONY: docker-build-push
docker-build-push: docker-build docker-push ## Build and push Docker image

##@ Helm
.PHONY: helm-template
helm-template: ## Render Helm templates
	@echo "Rendering Helm templates..."
	@helm template $(PROJECT_NAME) ./helm/$(PROJECT_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(VERSION) \
		--namespace $(NAMESPACE)

.PHONY: helm-lint
helm-lint: ## Lint Helm charts
	@echo "Linting Helm charts..."
	@helm lint ./helm/$(PROJECT_NAME)

.PHONY: helm-install-dry-run
helm-install-dry-run: ## Install Helm chart in dry-run mode
	@echo "Installing $(PROJECT_NAME) in dry-run mode..."
	@helm install $(PROJECT_NAME) ./helm/$(PROJECT_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(VERSION) \
		--set controller.mode=dry-run \
		--set mimir.namespace=$(MIMIR_NAMESPACE) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--dry-run

.PHONY: helm-install
helm-install: ## Install Helm chart
	@echo "Installing $(PROJECT_NAME)..."
	@helm install $(PROJECT_NAME) ./helm/$(PROJECT_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(VERSION) \
		--set controller.mode=dry-run \
		--set mimir.namespace=$(MIMIR_NAMESPACE) \
		--namespace $(NAMESPACE) \
		--create-namespace

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade Helm chart
	@echo "Upgrading $(PROJECT_NAME)..."
	@helm upgrade $(PROJECT_NAME) ./helm/$(PROJECT_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(VERSION) \
		--namespace $(NAMESPACE) \
		--reuse-values

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm chart
	@echo "Uninstalling $(PROJECT_NAME)..."
	@helm uninstall $(PROJECT_NAME) --namespace $(NAMESPACE)
	@kubectl delete namespace $(NAMESPACE) --ignore-not-found

.PHONY: helm-status
helm-status: ## Show Helm status
	@helm status $(PROJECT_NAME) --namespace $(NAMESPACE)

##@ Deployment
.PHONY: deploy-dry-run
deploy-dry-run: docker-build-push helm-install ## Build, push, and deploy in dry-run mode

.PHONY: deploy-prod
deploy-prod: ## Deploy in production mode
	@echo "Deploying $(PROJECT_NAME) in production mode..."
	@helm upgrade $(PROJECT_NAME) ./helm/$(PROJECT_NAME) \
		--set controller.mode=prod \
		--namespace $(NAMESPACE) \
		--reuse-values

.PHONY: logs
logs: ## Show application logs
	@kubectl logs -f deployment/$(PROJECT_NAME) -n $(NAMESPACE)

.PHONY: status
status: ## Show deployment status
	@echo "Deployment status:"
	@kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Service status:"
	@kubectl get services -n $(NAMESPACE)
	@echo ""
	@echo "ConfigMap status:"
	@kubectl get configmap -n $(NAMESPACE)

##@ Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@rm -f coverage.out coverage.html
	@docker image prune -f

.PHONY: clean-all
clean-all: clean helm-uninstall ## Clean everything including deployments
	@echo "Cleaning Docker images..."
	@docker rmi $(PROJECT_NAME):$(VERSION) $(IMAGE_NAME):$(VERSION) 2>/dev/null || true

##@ Utilities
.PHONY: check-deps
check-deps: ## Check if required dependencies are installed
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "kubectl is not installed"; exit 1; }
	@command -v helm >/dev/null 2>&1 || { echo "Helm is not installed"; exit 1; }
	@echo "All dependencies are installed!"

.PHONY: setup-dev
setup-dev: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod download
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready!"

# Example config file
.PHONY: example-config
example-config: ## Generate example configuration file
	@echo "Generating example configuration file..."
	@cat > config.yaml << 'EOF'
mode: "dry-run"
bufferPercentage: 20
updateInterval: "5m"

mimir:
  namespace: "$(MIMIR_NAMESPACE)"
  configMapName: "mimir-runtime-overrides"
  triggerRollout: false

tenantScoping:
  skipList:
    - "internal-*"
    - "test-*"
  includeList: []

metricsDiscovery:
  enabled: true
  namespace: "$(MIMIR_NAMESPACE)"
  serviceLabelSelector: "app.kubernetes.io/name=mimir"

eventSpike:
  enabled: true
  threshold: 2.0
  cooldownPeriod: "30m"

trendAnalysis:
  analysisWindow: "48h"
  percentile: 95.0
  useMovingAverage: true

auditLog:
  enabled: true
  storageType: "memory"
  maxEntries: 1000
EOF
	@echo "Example configuration saved to config.yaml" 