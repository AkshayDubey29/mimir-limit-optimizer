#!/bin/bash

# Test Ingress Configuration for Mimir Limit Optimizer
# This script tests both UI service and main service ingress configurations

set -e

NAMESPACE="mimir-limit-optimizer"
RELEASE_NAME="mimir-limit-optimizer"
TEST_HOST="mimir-optimizer.test.local"

echo "🌐 Testing Ingress Configuration for Mimir Limit Optimizer"
echo "============================================================"

# Function to wait for pod to be ready
wait_for_pod() {
    echo "⏳ Waiting for pod to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mimir-limit-optimizer \
        -n $NAMESPACE --timeout=120s
}

# Function to check ingress
check_ingress() {
    echo "🔍 Checking ingress status..."
    kubectl get ingress -n $NAMESPACE
    echo ""
    kubectl describe ingress -n $NAMESPACE
}

# Function to test ingress endpoint
test_endpoint() {
    local service_name=$1
    echo "🧪 Testing ingress endpoint for service: $service_name"
    
    # Get ingress IP
    INGRESS_IP=$(kubectl get ingress -n $NAMESPACE -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [ -z "$INGRESS_IP" ]; then
        echo "⚠️  No LoadBalancer IP found, trying to get from ingress controller..."
        INGRESS_IP=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "127.0.0.1")
    fi
    
    echo "📍 Using ingress IP: $INGRESS_IP"
    
    # Test with curl using Host header
    echo "🔗 Testing with curl (Host header method)..."
    curl -H "Host: $TEST_HOST" -I http://$INGRESS_IP/ 2>/dev/null | head -5 || echo "❌ Connection failed"
    
    echo ""
}

# Test 1: Default UI Service Ingress
echo "🎯 TEST 1: Testing with default UI service ingress"
echo "=================================================="

helm upgrade --install $RELEASE_NAME ./helm/mimir-limit-optimizer \
    --set ui.enabled=true \
    --set ui.ingress.enabled=true \
    --set ui.ingress.hosts[0].host=$TEST_HOST \
    --set ui.ingress.hosts[0].paths[0].path=/ \
    --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
    --namespace $NAMESPACE \
    --create-namespace

wait_for_pod
check_ingress
test_endpoint "ui-service"

echo ""
echo "✅ TEST 1 completed"
echo ""

# Test 2: Main Service Ingress
echo "🎯 TEST 2: Testing with main service ingress"
echo "============================================"

helm upgrade $RELEASE_NAME ./helm/mimir-limit-optimizer \
    --set ui.enabled=true \
    --set ui.ingress.enabled=true \
    --set ui.ingress.serviceName=$RELEASE_NAME \
    --set ui.ingress.servicePort=8082 \
    --set ui.ingress.hosts[0].host=$TEST_HOST \
    --set ui.ingress.hosts[0].paths[0].path=/ \
    --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
    --namespace $NAMESPACE

wait_for_pod
check_ingress
test_endpoint "main-service"

echo ""
echo "✅ TEST 2 completed"
echo ""

# Test 3: NGINX Ingress with Annotations
echo "🎯 TEST 3: Testing with NGINX ingress and annotations"
echo "===================================================="

helm upgrade $RELEASE_NAME ./helm/mimir-limit-optimizer \
    --set ui.enabled=true \
    --set ui.ingress.enabled=true \
    --set ui.ingress.className=nginx \
    --set ui.ingress.annotations."nginx\.ingress\.kubernetes\.io/rewrite-target"=/ \
    --set ui.ingress.annotations."nginx\.ingress\.kubernetes\.io/ssl-redirect"=false \
    --set ui.ingress.hosts[0].host=$TEST_HOST \
    --set ui.ingress.hosts[0].paths[0].path=/ \
    --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
    --namespace $NAMESPACE

wait_for_pod
check_ingress
test_endpoint "nginx-ingress"

echo ""
echo "✅ TEST 3 completed"
echo ""

# Summary
echo "📋 TEST SUMMARY"
echo "==============="
echo "✅ UI Service Ingress: Tested"
echo "✅ Main Service Ingress: Tested"
echo "✅ NGINX Ingress with Annotations: Tested"
echo ""
echo "🔍 To manually test ingress:"
echo "1. Add to /etc/hosts: [INGRESS_IP] $TEST_HOST"
echo "2. Open browser: http://$TEST_HOST"
echo "3. Or use curl: curl -H \"Host: $TEST_HOST\" http://[INGRESS_IP]/"
echo ""
echo "🧹 To cleanup:"
echo "helm uninstall $RELEASE_NAME -n $NAMESPACE"
echo "kubectl delete namespace $NAMESPACE" 