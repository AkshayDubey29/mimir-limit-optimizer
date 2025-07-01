#!/bin/bash

# Fix ALB Health Check Issues for Mimir Limit Optimizer
# This script diagnoses and fixes common ALB ingress health check problems

set -e

NAMESPACE="mimir-limit-optimizer"
RELEASE_NAME="mimir-limit-optimizer"

echo "🏥 ALB Health Check Diagnostic and Fix Tool"
echo "============================================"

# Function to check current ingress status
check_ingress_status() {
    echo "🔍 Checking current ingress configuration..."
    
    if ! kubectl get ingress -n $NAMESPACE &>/dev/null; then
        echo "❌ No ingress found in namespace $NAMESPACE"
        exit 1
    fi
    
    echo "📋 Current ingress details:"
    kubectl get ingress -n $NAMESPACE
    echo ""
    
    echo "🔧 Ingress annotations:"
    kubectl get ingress -n $NAMESPACE -o yaml | grep -A20 "annotations:" || echo "No annotations found"
    echo ""
}

# Function to test service health
test_service_health() {
    echo "🧪 Testing service health..."
    
    # Check if services exist
    if ! kubectl get svc -n $NAMESPACE &>/dev/null; then
        echo "❌ No services found in namespace $NAMESPACE"
        return 1
    fi
    
    echo "📋 Available services:"
    kubectl get svc -n $NAMESPACE
    echo ""
    
    # Test UI service if it exists
    if kubectl get svc -n $NAMESPACE | grep -q "ui"; then
        echo "🔗 Testing UI service health..."
        kubectl port-forward -n $NAMESPACE svc/${RELEASE_NAME}-ui 8082:8082 &
        PF_PID=$!
        sleep 3
        
        echo "Testing root path:"
        curl -I http://localhost:8082/ 2>/dev/null || echo "❌ Root path failed"
        
        echo "Testing /api/health:"
        curl -I http://localhost:8082/api/health 2>/dev/null || echo "❌ /api/health failed"
        
        echo "Testing /healthz:"
        curl -I http://localhost:8082/healthz 2>/dev/null || echo "❌ /healthz failed"
        
        kill $PF_PID 2>/dev/null || true
        wait $PF_PID 2>/dev/null || true
    fi
    
    # Test main service
    echo "🔗 Testing main service health..."
    kubectl port-forward -n $NAMESPACE svc/${RELEASE_NAME} 8082:8082 &
    PF_PID=$!
    sleep 3
    
    echo "Testing root path on main service:"
    curl -I http://localhost:8082/ 2>/dev/null || echo "❌ Root path failed"
    
    echo "Testing /api/health on main service:"
    curl -I http://localhost:8082/api/health 2>/dev/null || echo "❌ /api/health failed"
    
    echo "Testing metrics endpoint:"
    curl -I http://localhost:8080/metrics 2>/dev/null || echo "❌ Metrics endpoint failed"
    
    kill $PF_PID 2>/dev/null || true
    wait $PF_PID 2>/dev/null || true
    echo ""
}

# Function to check pod status
check_pod_status() {
    echo "📦 Checking pod status..."
    kubectl get pods -n $NAMESPACE
    echo ""
    
    echo "📜 Recent logs:"
    kubectl logs -n $NAMESPACE deployment/${RELEASE_NAME} --tail=20 || echo "❌ Could not get logs"
    echo ""
}

# Function to apply health check fix
apply_health_check_fix() {
    local option=$1
    
    case $option in
        1)
            echo "🔧 Applying Fix 1: Using /api/health endpoint..."
            helm upgrade ${RELEASE_NAME} ./helm/mimir-limit-optimizer \
                --set ui.ingress.enabled=true \
                --set ui.ingress.className=alb \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-path"=/api/health \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-port"=8082 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/success-codes"=200 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-interval-seconds"=30 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/unhealthy-threshold-count"=3 \
                --namespace $NAMESPACE \
                --reuse-values
            ;;
        2)
            echo "🔧 Applying Fix 2: Accept 404 for root path..."
            helm upgrade ${RELEASE_NAME} ./helm/mimir-limit-optimizer \
                --set ui.ingress.enabled=true \
                --set ui.ingress.className=alb \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-path"=/ \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/success-codes"="200,404" \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-interval-seconds"=15 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/unhealthy-threshold-count"=5 \
                --namespace $NAMESPACE \
                --reuse-values
            ;;
        3)
            echo "🔧 Applying Fix 3: Use metrics endpoint for health check..."
            helm upgrade ${RELEASE_NAME} ./helm/mimir-limit-optimizer \
                --set ui.ingress.enabled=true \
                --set ui.ingress.className=alb \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-path"=/metrics \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-port"=8080 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/success-codes"=200 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-interval-seconds"=30 \
                --namespace $NAMESPACE \
                --reuse-values
            ;;
        4)
            echo "🔧 Applying Fix 4: Comprehensive ALB configuration..."
            helm upgrade ${RELEASE_NAME} ./helm/mimir-limit-optimizer \
                --set ui.ingress.enabled=true \
                --set ui.ingress.className=alb \
                --set ui.ingress.annotations."kubernetes\.io/ingress\.class"=alb \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/scheme"=internet-facing \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/target-type"=ip \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-path"=/api/health \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-port"=8082 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-protocol"=HTTP \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-interval-seconds"=30 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthcheck-timeout-seconds"=5 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/healthy-threshold-count"=2 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/unhealthy-threshold-count"=3 \
                --set ui.ingress.annotations."alb\.ingress\.kubernetes\.io/success-codes"=200 \
                --namespace $NAMESPACE \
                --reuse-values
            ;;
        *)
            echo "❌ Invalid option"
            return 1
            ;;
    esac
}

# Function to wait and verify fix
verify_fix() {
    echo "⏳ Waiting for ALB to update..."
    sleep 30
    
    echo "🔍 Checking updated ingress status..."
    kubectl describe ingress -n $NAMESPACE
    
    echo ""
    echo "📋 Updated annotations:"
    kubectl get ingress -n $NAMESPACE -o yaml | grep -A20 "annotations:"
    
    echo ""
    echo "⚡ To monitor ALB health check status:"
    echo "1. Check AWS console: EC2 > Load Balancers > Target Groups"
    echo "2. Or use AWS CLI: aws elbv2 describe-target-health --target-group-arn YOUR_TG_ARN"
}

# Main execution
echo "🏁 Starting ALB health check diagnosis..."
echo ""

check_ingress_status
test_service_health
check_pod_status

echo "🔧 Available health check fixes:"
echo "1. Use /api/health endpoint (recommended if API is working)"
echo "2. Accept 404 for root path (quick fix, less strict)"
echo "3. Use /metrics endpoint (fallback to metrics port)"
echo "4. Apply comprehensive ALB configuration (complete setup)"
echo ""

read -p "Which fix would you like to apply? (1-4, or 'q' to quit): " choice

case $choice in
    [1-4])
        apply_health_check_fix $choice
        verify_fix
        echo ""
        echo "✅ Fix applied! Monitor your ALB target group health in AWS console."
        echo "🕐 It may take 2-3 minutes for health checks to pass."
        ;;
    q|Q)
        echo "👋 Exiting without applying fixes."
        exit 0
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "🎯 Next steps:"
echo "1. Monitor ALB target group health in AWS console"
echo "2. Check ingress events: kubectl describe ingress -n $NAMESPACE"
echo "3. Verify DNS resolution to ALB endpoint"
echo "4. Test application access via ALB URL"
echo ""
echo "🆘 If issues persist:"
echo "1. Check AWS ALB Controller logs"
echo "2. Verify VPC/subnet configuration"
echo "3. Check security group rules"
echo "4. Ensure pods are in Ready state" 