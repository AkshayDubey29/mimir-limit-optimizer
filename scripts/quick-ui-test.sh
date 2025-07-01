#!/bin/bash
# quick-ui-test.sh - Quick UI testing script with port forwarding

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="mimir-limit-optimizer"
RELEASE_NAME="mimir-limit-optimizer"
UI_PORT="8082"
METRICS_PORT="8080"

echo -e "${BLUE}🚀 Mimir Limit Optimizer UI Testing Script${NC}"
echo "============================================="

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo -e "${RED}❌ helm is not installed or not in PATH${NC}"
    exit 1
fi

# Check if connected to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Not connected to a Kubernetes cluster${NC}"
    echo "Please configure kubectl to connect to your cluster"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites checked${NC}"

# Deploy with UI enabled
echo -e "${YELLOW}🚀 Deploying Mimir Limit Optimizer with UI enabled...${NC}"

helm upgrade --install $RELEASE_NAME ./helm/mimir-limit-optimizer \
  --set image.repository=artifactory.coupang.net/ghcr-remote/akshaydubey29/mimir-limit-optimizer \
  --set image.tag=v3.0.0 \
  --set controller.mode=dry-run \
  --set mimir.namespace=mimir \
  --set metricsDiscovery.enabled=false \
  --set ui.enabled=true \
  --set ui.port=$UI_PORT \
  --set ui.service.type=ClusterIP \
  --namespace $NAMESPACE \
  --create-namespace

echo -e "${YELLOW}⏳ Waiting for pods to be ready...${NC}"
if kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mimir-limit-optimizer -n $NAMESPACE --timeout=120s; then
    echo -e "${GREEN}✅ Deployment ready${NC}"
else
    echo -e "${RED}❌ Deployment failed to become ready within timeout${NC}"
    echo "Checking pod status..."
    kubectl get pods -n $NAMESPACE
    echo "Recent logs:"
    kubectl logs -n $NAMESPACE deployment/mimir-limit-optimizer --tail=20
    exit 1
fi

# Show deployment status
echo -e "${BLUE}📊 Deployment Status:${NC}"
kubectl get pods,svc -n $NAMESPACE

# Use main service which now includes both metrics and UI ports
UI_SERVICE="mimir-limit-optimizer"

echo -e "${BLUE}🔍 Service Details:${NC}"
kubectl describe svc -n $NAMESPACE $UI_SERVICE

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up port forwards...${NC}"
    # Kill any port-forward processes
    pkill -f "kubectl port-forward" || true
}

# Set trap to cleanup on script exit
trap cleanup EXIT

# Start port forwarding
echo -e "${GREEN}🌐 Starting port forward to UI...${NC}"
echo -e "${BLUE}UI will be available at: http://localhost:$UI_PORT${NC}"
echo -e "${BLUE}Metrics available at: http://localhost:$METRICS_PORT/metrics${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop port forwarding and exit${NC}"
echo ""

# Start port forwarding in background and capture PID
kubectl port-forward -n $NAMESPACE service/$UI_SERVICE $UI_PORT:$UI_PORT $METRICS_PORT:$METRICS_PORT &
PORT_FORWARD_PID=$!

# Wait a moment for port forward to establish
sleep 3

# Test if UI is responding
echo -e "${BLUE}🔍 Testing UI connectivity...${NC}"
if curl -s --max-time 5 http://localhost:$UI_PORT/ > /dev/null; then
    echo -e "${GREEN}✅ UI is responding at http://localhost:$UI_PORT${NC}"
else
    echo -e "${YELLOW}⚠️  UI might not be ready yet, but port forward is active${NC}"
fi

# Test metrics endpoint
if curl -s --max-time 5 http://localhost:$METRICS_PORT/metrics > /dev/null; then
    echo -e "${GREEN}✅ Metrics endpoint responding at http://localhost:$METRICS_PORT/metrics${NC}"
else
    echo -e "${YELLOW}⚠️  Metrics endpoint might not be ready yet${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Setup complete! You can now test the UI.${NC}"
echo ""
echo "🔗 Quick Links:"
echo "  • UI Dashboard: http://localhost:$UI_PORT"
echo "  • Health Check: http://localhost:$UI_PORT/health"
echo "  • API Tenants:  http://localhost:$UI_PORT/api/tenants" 
echo "  • Metrics:      http://localhost:$METRICS_PORT/metrics"
echo ""
echo "📋 Useful Commands (in another terminal):"
echo "  • View logs:    kubectl logs -n $NAMESPACE deployment/mimir-limit-optimizer -f"
echo "  • Check pods:   kubectl get pods -n $NAMESPACE"
echo "  • Describe svc: kubectl describe svc -n $NAMESPACE $UI_SERVICE"
echo ""

# Wait for port forward process
wait $PORT_FORWARD_PID 