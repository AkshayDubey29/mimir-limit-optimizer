#!/bin/bash

# Production deployment script for Mimir Limit Optimizer
# Usage: ./deploy-production.sh [ENVIRONMENT] [DOMAIN]
# Example: ./deploy-production.sh prod mimir-optimizer.example.com

set -e

# Configuration
ENVIRONMENT=${1:-prod}
DOMAIN=${2:-mimir-optimizer.yourdomain.com}
NAMESPACE="mimir-optimizer-${ENVIRONMENT}"
RELEASE_NAME="mimir-optimizer-${ENVIRONMENT}"
CHART_PATH="./helm/mimir-limit-optimizer"
VALUES_FILE="prod-values.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Pre-flight checks
log "Starting Mimir Limit Optimizer production deployment..."
log "Environment: ${ENVIRONMENT}"
log "Domain: ${DOMAIN}"
log "Namespace: ${NAMESPACE}"
log "Release: ${RELEASE_NAME}"

# Check if kubectl is available and connected
if ! kubectl cluster-info &> /dev/null; then
    error "kubectl not available or not connected to cluster"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    error "helm not found. Please install Helm."
    exit 1
fi

# Check if chart exists
if [[ ! -d "$CHART_PATH" ]]; then
    error "Helm chart not found at $CHART_PATH"
    exit 1
fi

# Check if values file exists
if [[ ! -f "$VALUES_FILE" ]]; then
    error "Values file not found at $VALUES_FILE"
    exit 1
fi

# Update domain in values file
log "Updating domain in values file..."
if [[ "$DOMAIN" != "mimir-optimizer.yourdomain.com" ]]; then
    sed -i.bak "s/mimir-optimizer\.yourdomain\.com/$DOMAIN/g" "$VALUES_FILE"
    log "Updated domain to $DOMAIN"
fi

# Create namespace if it doesn't exist
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    log "Creating namespace $NAMESPACE..."
    kubectl create namespace "$NAMESPACE"
    success "Namespace $NAMESPACE created"
else
    log "Namespace $NAMESPACE already exists"
fi

# Check if release already exists
if helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
    warn "Release $RELEASE_NAME already exists. This will upgrade the existing release."
    read -p "Continue with upgrade? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "Deployment cancelled"
        exit 0
    fi
    HELM_COMMAND="upgrade"
else
    HELM_COMMAND="install"
fi

# Deploy with Helm
log "Deploying Mimir Limit Optimizer..."
helm $HELM_COMMAND "$RELEASE_NAME" "$CHART_PATH" \
    --namespace "$NAMESPACE" \
    --values "$VALUES_FILE" \
    --wait \
    --timeout 10m

if [[ $? -eq 0 ]]; then
    success "Deployment successful!"
else
    error "Deployment failed!"
    exit 1
fi

# Post-deployment checks
log "Performing post-deployment checks..."

# Wait for pods to be ready
log "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mimir-limit-optimizer -n "$NAMESPACE" --timeout=300s

# Check deployment status
PODS_READY=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=mimir-limit-optimizer --no-headers | grep Running | wc -l)
PODS_TOTAL=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=mimir-limit-optimizer --no-headers | wc -l)

if [[ $PODS_READY -eq $PODS_TOTAL && $PODS_READY -gt 0 ]]; then
    success "All pods are running ($PODS_READY/$PODS_TOTAL)"
else
    warn "Not all pods are ready ($PODS_READY/$PODS_TOTAL)"
fi

# Show service information
log "Service information:"
kubectl get svc -n "$NAMESPACE" -l app.kubernetes.io/name=mimir-limit-optimizer

# Show ingress information
if kubectl get ingress -n "$NAMESPACE" &> /dev/null; then
    log "Ingress information:"
    kubectl get ingress -n "$NAMESPACE"
fi

# Port forwarding instructions
log ""
success "Deployment completed successfully!"
log ""
log "=== Access Information ==="
log "Web UI: https://$DOMAIN"
log ""
log "For local access via port forwarding:"
log "kubectl port-forward -n $NAMESPACE svc/${RELEASE_NAME}-ui 8080:8082"
log "Then access: http://localhost:8080"
log ""
log "=== Monitoring ==="
log "Check pod logs:"
log "kubectl logs -f -n $NAMESPACE -l app.kubernetes.io/name=mimir-limit-optimizer"
log ""
log "Check pod status:"
log "kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=mimir-limit-optimizer"
log ""
log "=== Troubleshooting ==="
log "Describe deployment:"
log "kubectl describe deployment -n $NAMESPACE ${RELEASE_NAME}"
log ""
log "Check events:"
log "kubectl get events -n $NAMESPACE --sort-by='.firstTimestamp'"

# Restore original values file if we modified it
if [[ -f "${VALUES_FILE}.bak" ]]; then
    mv "${VALUES_FILE}.bak" "$VALUES_FILE"
    log "Restored original values file"
fi

log ""
success "🎉 Mimir Limit Optimizer deployed successfully!"
log "Happy optimizing! 🚀" 