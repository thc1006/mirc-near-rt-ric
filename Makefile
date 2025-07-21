# Makefile for O-RAN Near-RT RIC Platform

# Go variables
GO_CMD := go
GO_FLAGS := -ldflags="-w -s"
GO_PACKAGES := ./...
GO_MAIN_FILES := $(wildcard ./src/app/backend/*.go)
GO_BIN_DIR := ./bin

# Node.js variables
NPM_CMD := npm
ANGULAR_MAIN_DIR := ./dashboard-master/dashboard-master
ANGULAR_XAPP_DIR := ./xAPP_dashboard-master

# Docker variables
DOCKER_CMD := docker
REGISTRY := ghcr.io
IMAGE_NAMESPACE := $(shell echo ${{ github.repository_owner }} | tr '[:upper:]' '[:lower:]')
MAIN_DASHBOARD_IMAGE := $(REGISTRY)/$(IMAGE_NAMESPACE)/main-dashboard
XAPP_DASHBOARD_IMAGE := $(REGISTRY)/$(IMAGE_NAMESPACE)/xapp-dashboard
FL_COORDINATOR_IMAGE := $(REGISTRY)/$(IMAGE_NAMESPACE)/fl-coordinator

# Helm variables
HELM_CMD := helm
HELM_CHART_DIR := ./helm/oran-nearrt-ric
RELEASE_NAME := oran-nearrt-ric
ORAN_NAMESPACE := oran-nearrt-ric

.PHONY: all clean lint test build docker-build docker-push helm-lint helm-package deploy

all: lint test build

# ==============================================================================
# Cleaning
# ==============================================================================
clean:
	@echo ">> cleaning up..."
	@rm -rf $(GO_BIN_DIR)
	@rm -rf $(ANGULAR_MAIN_DIR)/dist
	@rm -rf $(ANGULAR_XAPP_DIR)/dist
	@$(GO_CMD) clean -modcache

# ==============================================================================
# Linting
# ==============================================================================
lint: lint-go lint-ts

lint-go:
	@echo ">> linting Go code..."
	@docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run $(GO_PACKAGES)

lint-ts:
	@echo ">> linting TypeScript code..."
	@$(NPM_CMD) --prefix $(ANGULAR_MAIN_DIR) run lint
	@$(NPM_CMD) --prefix $(ANGULAR_XAPP_DIR) run lint

# ==============================================================================
# Testing
# ==============================================================================
test: test-go test-ts

test-go:
	@echo ">> running Go tests..."
	@$(GO_CMD) test -v -race -coverprofile=coverage.out $(GO_PACKAGES)

test-ts:
	@echo ">> running TypeScript tests..."
	@$(NPM_CMD) --prefix $(ANGULAR_MAIN_DIR) run test -- --watch=false --browsers=ChromeHeadless
	@$(NPM_CMD) --prefix $(ANGULAR_XAPP_DIR) run test -- --watch=false --browsers=ChromeHeadless

# ==============================================================================
# Building
# ==============================================================================
build: build-go build-ts

build-go:
	@echo ">> building Go binaries..."
	@CGO_ENABLED=0 $(GO_CMD) build $(GO_FLAGS) -o $(GO_BIN_DIR)/main-dashboard $(GO_MAIN_FILES)

build-ts:
	@echo ">> building Angular applications..."
	@$(NPM_CMD) --prefix $(ANGULAR_MAIN_DIR) run build -- --configuration=production
	@$(NPM_CMD) --prefix $(ANGULAR_XAPP_DIR) run build -- --configuration=production

# ==============================================================================
# Docker
# ==============================================================================
docker-build:
	@echo ">> building Docker images..."
	@$(DOCKER_CMD) buildx build --platform linux/amd64,linux/arm64 -t $(MAIN_DASHBOARD_IMAGE):latest -f $(ANGULAR_MAIN_DIR)/Dockerfile .
	@$(DOCKER_CMD) buildx build --platform linux/amd64,linux/arm64 -t $(XAPP_DASHBOARD_IMAGE):latest -f $(ANGULAR_XAPP_DIR)/Dockerfile .
	@$(DOCKER_CMD) buildx build --platform linux/amd64,linux/arm64 -t $(FL_COORDINATOR_IMAGE):latest -f $(ANGULAR_MAIN_DIR)/Dockerfile.fl .

docker-push:
	@echo ">> pushing Docker images..."
	@$(DOCKER_CMD) push $(MAIN_DASHBOARD_IMAGE):latest
	@$(DOCKER_CMD) push $(XAPP_DASHBOARD_IMAGE):latest
	@$(DOCKER_CMD) push $(FL_COORDINATOR_IMAGE):latest

# ==============================================================================
# Helm
# ==============================================================================
helm-lint:
	@echo ">> linting Helm chart..."
	@$(HELM_CMD) lint $(HELM_CHART_DIR)

helm-package:
	@echo ">> packaging Helm chart..."
	@$(HELM_CMD) package $(HELM_CHART_DIR)

# ==============================================================================
# Deployment
# ==============================================================================
deploy: helm-package
	@echo ">> deploying to Kubernetes..."
	@$(HELM_CMD) upgrade --install $(RELEASE_NAME) ./$(RELEASE_NAME)-*.tgz \
		--namespace $(ORAN_NAMESPACE) \
		--create-namespace \
		--wait
