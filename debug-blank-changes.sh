#!/bin/bash

# Comprehensive Debug Script for Blank Audit Log Changes
# This script diagnoses why the "changes" field in audit logs is empty

set -e

echo "🔍 DEBUGGING BLANK AUDIT LOG CHANGES"
echo "====================================="
echo ""

# Function to check if kubectl is available and working
check_kubernetes() {
    echo "🔧 CHECKING KUBERNETES ACCESS:"
    echo "=============================="
    if command -v kubectl &> /dev/null; then
        echo "✅ kubectl is available"
        if kubectl cluster-info &> /dev/null; then
            echo "✅ Kubernetes cluster is accessible"
            kubectl get nodes --no-headers | wc -l | xargs echo "✅ Cluster has nodes:"
            return 0
        else
            echo "❌ Kubernetes cluster is not accessible"
            return 1
        fi
    else
        echo "❌ kubectl is not available"
        return 1
    fi
}

# Function to check if Mimir services exist
check_mimir_services() {
    echo ""
    echo "🏗️  CHECKING MIMIR SERVICES:"
    echo "============================"
    
    MIMIR_NAMESPACE=${1:-mimir}
    echo "Checking namespace: $MIMIR_NAMESPACE"
    
    if kubectl get namespace "$MIMIR_NAMESPACE" &> /dev/null; then
        echo "✅ Namespace $MIMIR_NAMESPACE exists"
        
        echo ""
        echo "Services in $MIMIR_NAMESPACE:"
        kubectl get services -n "$MIMIR_NAMESPACE" --no-headers 2>/dev/null | while read line; do
            echo "  📦 $line"
        done
        
        echo ""
        echo "ConfigMaps in $MIMIR_NAMESPACE:"
        kubectl get configmaps -n "$MIMIR_NAMESPACE" --no-headers 2>/dev/null | while read line; do
            echo "  📄 $line"
        done
        
        # Check for runtime overrides ConfigMap specifically
        if kubectl get configmap mimir-runtime-overrides -n "$MIMIR_NAMESPACE" &> /dev/null; then
            echo "✅ mimir-runtime-overrides ConfigMap exists"
            echo ""
            echo "Current overrides content:"
            kubectl get configmap mimir-runtime-overrides -n "$MIMIR_NAMESPACE" -o yaml | grep -A 20 "data:" || echo "  (empty or not found)"
        else
            echo "❌ mimir-runtime-overrides ConfigMap not found"
        fi
        
    else
        echo "❌ Namespace $MIMIR_NAMESPACE does not exist"
        echo ""
        echo "Available namespaces:"
        kubectl get namespaces | grep -E "(mimir|prometheus|monitoring)" || echo "  No Mimir-related namespaces found"
    fi
}

# Function to build and run the optimizer with debug output
run_debug_optimizer() {
    echo ""
    echo "🚀 RUNNING OPTIMIZER WITH DEBUG LOGGING:"
    echo "========================================"
    
    # Build if needed
    if [ ! -f "./mimir-limit-optimizer" ]; then
        echo "🔨 Building mimir-limit-optimizer..."
        go build -o mimir-limit-optimizer .
    fi
    
    echo "🏃 Running optimizer for 90 seconds with debug logging..."
    echo "📝 Output will be saved to: debug-output.log"
    echo ""
    
    # Run with timeout and capture all output
    timeout 90s ./mimir-limit-optimizer -config debug-audit-config.yaml -v=2 > debug-output.log 2>&1 &
    OPTIMIZER_PID=$!
    
    echo "⏳ Running... (PID: $OPTIMIZER_PID)"
    echo "💡 Tip: In another terminal, run 'tail -f debug-output.log' to watch real-time logs"
    echo ""
    
    # Show progress for 30 seconds, then wait
    for i in {1..30}; do
        if ! kill -0 $OPTIMIZER_PID 2>/dev/null; then
            echo "❌ Process exited early!"
            break
        fi
        printf "."
        sleep 1
    done
    
    echo ""
    echo "⏳ Continuing to run for remaining time..."
    
    # Wait for completion or timeout
    wait $OPTIMIZER_PID 2>/dev/null || true
    
    echo "✅ Debug run completed!"
}

# Function to analyze the debug output
analyze_debug_output() {
    echo ""
    echo "📊 ANALYZING DEBUG OUTPUT:"
    echo "=========================="
    
    if [ ! -f "debug-output.log" ]; then
        echo "❌ Debug output file not found!"
        return 1
    fi
    
    echo ""
    echo "1️⃣  STARTUP AND CONNECTIVITY:"
    echo "----------------------------"
    if grep -q "Starting mimir-limit-optimizer" debug-output.log; then
        echo "✅ Application started successfully"
        grep "Starting mimir-limit-optimizer" debug-output.log | head -1
    else
        echo "❌ Application failed to start"
    fi
    
    if grep -q "unable to load in-cluster config\|unable to get kubeconfig" debug-output.log; then
        echo "❌ Kubernetes connectivity issues found:"
        grep "unable to load in-cluster config\|unable to get kubeconfig" debug-output.log | head -2
    else
        echo "✅ No Kubernetes connectivity issues"
    fi
    
    echo ""
    echo "2️⃣  TENANT DISCOVERY:"
    echo "--------------------"
    if grep -q "tenant.*discovered\|found.*tenant\|GetTenantList" debug-output.log; then
        echo "✅ Tenant discovery attempted:"
        grep -i "tenant.*discovered\|found.*tenant\|GetTenantList" debug-output.log | head -5
    else
        echo "❌ No tenant discovery logs found"
    fi
    
    if grep -q "fallback.*tenant\|synthetic.*tenant" debug-output.log; then
        echo "ℹ️  Fallback tenant mechanisms used:"
        grep -i "fallback.*tenant\|synthetic.*tenant" debug-output.log | head -3
    fi
    
    echo ""
    echo "3️⃣  METRICS COLLECTION:"
    echo "----------------------"
    if grep -q "collected metrics\|metrics collection" debug-output.log; then
        echo "✅ Metrics collection attempted:"
        grep -i "collected metrics\|metrics collection" debug-output.log | head -5
    else
        echo "❌ No metrics collection logs found"
    fi
    
    if grep -q "no metrics sources configured\|failed to collect" debug-output.log; then
        echo "❌ Metrics collection issues:"
        grep -i "no metrics sources configured\|failed to collect" debug-output.log | head -3
    fi
    
    echo ""
    echo "4️⃣  LIMIT CALCULATIONS:"
    echo "----------------------"
    if grep -q "AnalyzeTrends\|CalculateLimits\|analyzed trends" debug-output.log; then
        echo "✅ Limit analysis attempted:"
        grep -i "AnalyzeTrends\|CalculateLimits\|analyzed trends" debug-output.log | head -5
    else
        echo "❌ No limit calculation logs found"
    fi
    
    if grep -q "successfully applied limits\|successfully wrote optimized limits" debug-output.log; then
        echo "✅ Limits were applied:"
        grep "successfully applied limits\|successfully wrote optimized limits" debug-output.log
    else
        echo "❌ No successful limit applications found"
    fi
    
    echo ""
    echo "5️⃣  AUDIT LOG ENTRIES:"
    echo "---------------------"
    if grep -q "audit entry logged" debug-output.log; then
        echo "✅ Audit entries were created:"
        grep "audit entry logged" debug-output.log | head -10
        
        # Extract details about the audit entries
        echo ""
        echo "🔍 Audit entry details:"
        grep -o '"tenant":"[^"]*"' debug-output.log | sort | uniq | head -5
        grep -o '"action":"[^"]*"' debug-output.log | sort | uniq | head -5
        grep -o '"reason":"[^"]*"' debug-output.log | sort | uniq | head -5
    else
        echo "❌ No audit entries created"
    fi
    
    echo ""
    echo "6️⃣  ERRORS AND WARNINGS:"
    echo "------------------------"
    if grep -q '"level":"error"' debug-output.log; then
        echo "❌ Errors found:"
        grep '"level":"error"' debug-output.log | head -5
    else
        echo "✅ No errors found"
    fi
    
    if grep -q '"level":"warn"' debug-output.log; then
        echo "⚠️  Warnings found:"
        grep '"level":"warn"' debug-output.log | head -3
    fi
}

# Function to check actual audit log contents
check_audit_configmap() {
    echo ""
    echo "📋 CHECKING AUDIT CONFIGMAP:"
    echo "============================"
    
    if kubectl get configmap mimir-limit-optimizer-audit -n mimir-limit-optimizer &> /dev/null; then
        echo "✅ Audit ConfigMap exists"
        echo ""
        echo "📊 Audit log entries:"
        kubectl get configmap mimir-limit-optimizer-audit -n mimir-limit-optimizer -o jsonpath='{.data.audit\.json}' | jq -r '.[] | select(.action == "update-limits") | "Tenant: \(.tenant), Action: \(.action), Changes: \(.changes)"' 2>/dev/null || {
            echo "Raw audit data:"
            kubectl get configmap mimir-limit-optimizer-audit -n mimir-limit-optimizer -o yaml | grep -A 50 "data:" | head -20
        }
    else
        echo "❌ Audit ConfigMap not found in mimir-limit-optimizer namespace"
        
        # Check other possible locations
        echo ""
        echo "🔍 Searching for audit ConfigMaps in other namespaces:"
        kubectl get configmaps --all-namespaces | grep "mimir-limit-optimizer-audit" || echo "  None found"
    fi
}

# Function to provide specific recommendations
provide_recommendations() {
    echo ""
    echo "💡 RECOMMENDATIONS TO FIX BLANK CHANGES:"
    echo "========================================"
    echo ""
    
    if ! grep -q "collected metrics" debug-output.log 2>/dev/null; then
        echo "🎯 ISSUE: No metrics are being collected"
        echo "   SOLUTIONS:"
        echo "   1. Update metricsEndpoint in debug-audit-config.yaml to your actual Prometheus/Mimir endpoint"
        echo "   2. Verify network connectivity from the pod to the metrics endpoint"
        echo "   3. Check if multi-tenant headers are correct (X-Scope-OrgID: couwatch)"
        echo ""
    fi
    
    if ! grep -q "tenant.*discovered\|found.*tenant" debug-output.log 2>/dev/null; then
        echo "🎯 ISSUE: No tenants are being discovered"
        echo "   SOLUTIONS:"
        echo "   1. Verify fallbackTenants list includes your actual tenant names"
        echo "   2. Check if mimir-runtime-overrides ConfigMap exists and contains tenant data"
        echo "   3. Enable synthetic tenants as a fallback (enableSynthetic: true)"
        echo ""
    fi
    
    if ! grep -q "CalculateLimits" debug-output.log 2>/dev/null; then
        echo "🎯 ISSUE: Limit calculations are not running"
        echo "   SOLUTIONS:"
        echo "   1. Ensure dynamicLimits.enabled: true"
        echo "   2. Verify limit definitions are properly configured and enabled"
        echo "   3. Check that analysis windows and buffer percentages are set"
        echo ""
    fi
    
    echo "🔧 IMMEDIATE NEXT STEPS:"
    echo "1. Update the metrics endpoint in debug-audit-config.yaml"
    echo "2. Run: kubectl get services -n mimir  # to find correct service names"
    echo "3. Run: kubectl get configmap mimir-runtime-overrides -n mimir -o yaml  # to see existing tenants"
    echo "4. Re-run this debug script after making changes"
}

# Main execution
main() {
    echo "Starting comprehensive debug analysis..."
    echo ""
    
    # Check Kubernetes access
    if check_kubernetes; then
        check_mimir_services "mimir"
    else
        echo "⚠️  Kubernetes not accessible - will check application logs only"
    fi
    
    # Update config file with user's environment
    echo ""
    echo "🔧 CONFIGURING DEBUG ENVIRONMENT:"
    echo "================================"
    echo "📝 Edit debug-audit-config.yaml and update:"
    echo "   - metricsEndpoint: to your actual Prometheus/Mimir endpoint"
    echo "   - mimir.namespace: to your Mimir namespace"
    echo "   - tenantHeaders.X-Scope-OrgID: to your tenant ID"
    echo ""
    read -p "Press Enter after updating the configuration, or Ctrl+C to exit..."
    
    # Run debug version
    run_debug_optimizer
    
    # Analyze results
    analyze_debug_output
    
    # Check audit ConfigMap if available
    if check_kubernetes; then
        check_audit_configmap
    fi
    
    # Provide recommendations
    provide_recommendations
    
    echo ""
    echo "🎯 DEBUG COMPLETE!"
    echo "=================="
    echo "📁 Full debug log available in: debug-output.log"
    echo "📋 Configuration used: debug-audit-config.yaml"
    echo ""
    echo "For further assistance, share the debug-output.log contents."
}

# Run the main function
main "$@" 