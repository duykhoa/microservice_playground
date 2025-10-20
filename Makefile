# Makefile for microservice-playground

# ====================================================================================
#  CLUSTER MANAGEMENT
# ====================================================================================

.PHONY: start-cluster
start-cluster: ## Start the local Kubernetes cluster
	@echo "Starting Minikube cluster..."
	@-minikube start

.PHONY: shutdown
shutdown: ## Shutdown the local Kubernetes cluster
	@echo "Shutting down the local Kubernetes cluster..."
	@-minikube stop

.PHONY: dashboard
dashboard:
	@echo "Opening the Minikube dashboard..."
	@-minikube minikube enable metrics-server
	@-minikube dashboard

# ====================================================================================
#  DEVELOPMENT
# ====================================================================================
.PHONY: dev
dev: start-cluster ## Run the application using skaffold
	@echo "Starting the application with skaffold for development..."
	@skaffold dev

.PHONY: skaffold-dev
skaffold-dev:
	@echo "Starting the application in development mode without calling start-cluster..."
	@skaffold dev

.PHONY: run
run: start-cluster ## Run the application using skaffold
	@echo "Starting the application with skaffold..."
	@skaffold run

.PHONY: build
build: build-backend build-frontend ## Build all applications

.PHONY: build-frontend
build-frontend: ## Build the frontend application
	@echo "Building the frontend application..."
	@cd web && npm install && npm run build

.PHONY: build-backend
build-backend: ## Build the backend services
	@echo "Building the backend services..."
	@cd services && go build -o order-service/order-service order-service/main.go
	@cd services && go build -o products-service/products-service products-service/main.go
	@cd services && go build -o warehouse/warehouse warehouse/main.go

.PHONY: restart-app
restart-app: ## Restart a specific application in the Kubernetes cluster
	@read -p "Enter the name of the app to restart (e.g., order-service): " app_name; \
	kubectl rollout restart deployment/$$app_name

# ====================================================================================
#  TESTING
# ====================================================================================

.PHONY: test
test: test-backend test-frontend ## Run all tests

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	@cd web && npm test

.PHONY: test-backend
test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@cd services && go test ./...

# ====================================================================================
#  HELP
# ====================================================================================

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help