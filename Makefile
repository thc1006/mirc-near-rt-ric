# =================================================================================================
# Makefile for O-RAN Near-RT RIC
#
# This Makefile provides a structured way to build, test, and deploy the RIC components.
# =================================================================================================

.PHONY: all help build test lint fmt security-scan docker-build docker-push clean deploy-local undeploy-local

# Go parameters
GO_VERSION   := 1.21
GO_CMD       := go
GO_FLAGS     := -ldflags='-w -s'
CGO_ENABLED  := 0
GOOS         := linux
GOARCH       := amd64

# Docker parameters
DOCKER_CMD      ?= docker
DOCKER_REGISTRY ?= ghcr.io/your-org
IMAGE_TAG       ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "latest")

# Service definitions
# Binaries to be built
BINARIES := ric-a1 ric-e2 ric-o1 ric-control

# Docker images to be built
IMAGES := ric-a1 ric-e2 ric-o1 ric-control fl-coordinator xapp-dashboard

# Default target
all: build

# =================================================================================================
# Help Target
# =================================================================================================

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all                  - Build all binaries (default)."
	@echo "  help                 - Show this help message."
	@echo ""
	@echo "Development:"
	@echo "  build                - Build all service binaries."
	@echo "  test                 - Run all unit tests."
	@echo "  lint                 - Run static analysis checks."
	@echo "  fmt                  - Format Go source code."
	@echo "  clean                - Clean up build artifacts."
	@echo ""
	@echo "Security:"
	@echo "  security-scan        - Run security vulnerability scans (requires trivy)."
	@echo ""
	@echo "Docker:"
	@echo "  docker-build         - Build all Docker images."
	@echo "  docker-push          - Push all Docker images to the registry."
	@echo ""
	@echo "Deployment:"
	@echo "  deploy-local         - Deploy all components to the local Kubernetes cluster using Helm."
	@echo "  undeploy-local       - Remove all deployed components from the local cluster."
	@echo ""

# =================================================================================================
# Development Targets
# =================================================================================================

build: $(BINARIES)

$(BINARIES):
	@echo "--> Building binary: $@"
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_CMD) build $(GO_FLAGS) -o ./bin/$@ ./cmd/$@

test:
	@echo "--> Running unit tests..."
	@$(GO_CMD) test -v -cover -race ./...

lint:
	@echo "--> Running linter..."
	@golangci-lint run

fmt:
	@echo "--> Formatting Go code..."
	@$(GO_CMD) fmt ./...

clean:
	@echo "--> Cleaning up build artifacts..."
	@rm -rf ./bin
	@$(GO_CMD) clean -cache

# =================================================================================================
# Security Targets
# =================================================================================================

security-scan:
	@echo "--> Running security scan on Docker images..."
	@if ! command -v trivy &> /dev/null; then \
		echo "Trivy not found. Please install trivy to run security scans."; \
		exit 1; \
	fi
	@for img in $(IMAGES); do \
		echo "Scanning image: $(DOCKER_REGISTRY)/$$img:$(IMAGE_TAG)"; \
		trivy image --exit-code 1 --severity HIGH,CRITICAL $(DOCKER_REGISTRY)/$$img:$(IMAGE_TAG); \
	done

# =================================================================================================
# Docker Targets
# =================================================================================================

docker-build: $(addprefix docker-build-, $(IMAGES))

docker-build-%:
	@echo "--> Building Docker image for: $*"
	@if [ -f ./cmd/$*/Dockerfile ]; then \
		$(DOCKER_CMD) build -t $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG) -f ./cmd/$*/Dockerfile . ; \
	else \
		$(DOCKER_CMD) build -t $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG) -f ./docker/Dockerfile.$* . ; \
	fi

docker-push: $(addprefix docker-push-, $(IMAGES))

docker-push-%:
	@echo "--> Pushing Docker image: $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG)"
	@$(DOCKER_CMD) push $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG)

# =================================================================================================
# Deployment Targets
# =================================================================================================

deploy-local:
	@echo "--> Deploying Near-RT RIC to local Kubernetes cluster..."
	@helm upgrade --install oran-nearrt-ric ./helm/oran-nearrt-ric \
		--namespace oran --create-namespace \
		--set image.tag=$(IMAGE_TAG) \
		--values ./helm/oran-nearrt-ric/values.yaml

undeploy-local:
	@echo "--> Undeploying Near-RT RIC from local Kubernetes cluster..."
	@helm uninstall oran-nearrt-ric --namespace oran