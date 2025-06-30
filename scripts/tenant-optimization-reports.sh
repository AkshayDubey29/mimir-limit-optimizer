#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

NAMESPACE="${1:-default}"
DEPLOYMENT="mimir-limit-optimizer"

echo -e "${BLUE}📊 Mimir Limit Optimizer - Tenant Optimization Reports${NC}"
echo -e "${BLUE}======================================================${NC}"

function show_usage() {
    echo "Usage: $0 [namespace]"
    echo ""
    echo "Examples:"
    echo "  $0                    # Use 'default' namespace"
    echo "  $0 mimir-optimizer   # Use 'mimir-optimizer' namespace"
    echo ""
}

function check_deployment() {
    if ! kubectl get deployment $DEPLOYMENT -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "${RED}❌ Deployment $DEPLOYMENT not found in namespace $NAMESPACE${NC}"
        echo -e "${YELLOW}💡 Available deployments:${NC}"
        kubectl get deployments -n $NAMESPACE
        exit 1
    fi
}

function live_optimization_feed() {
    echo -e "\n${GREEN}🔄 Live Tenant Optimization Feed${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop...${NC}"
    echo "----------------------------------------"
    
    kubectl logs -n $NAMESPACE deployment/$DEPLOYMENT -f | \
    grep -E --line-buffered "(tenant|optimization|recommendation|dry-run|analysis_result)" | \
    while read line; do
        timestamp=$(echo $line | grep -o '"ts":"[^"]*"' | cut -d'"' -f4)
        tenant=$(echo $line | grep -o '"tenant":"[^"]*"' | cut -d'"' -f4)
        
        if echo $line | grep -q "optimization_recommended"; then
            echo -e "${GREEN}✅ [$timestamp] OPTIMIZATION: $tenant${NC}"
        elif echo $line | grep -q "dry-run"; then
            echo -e "${CYAN}🧪 [$timestamp] DRY-RUN: $tenant${NC}"
        elif echo $line | grep -q "cost_exceeded"; then
            echo -e "${RED}💰 [$timestamp] COST ALERT: $tenant${NC}"
        elif echo $line | grep -q "spike_detected"; then
            echo -e "${YELLOW}📈 [$timestamp] SPIKE: $tenant${NC}"
        else
            echo -e "${PURPLE}📋 [$timestamp] ANALYSIS: $tenant${NC}"
        fi
        echo "   $line"
        echo ""
    done
}

function audit_log_report() {
    echo -e "\n${GREEN}📋 Audit Log Report${NC}"
    echo "----------------------------------------"
    
    if kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Found audit ConfigMap${NC}"
        
        echo -e "\n${CYAN}📊 Recent Tenant Activities:${NC}"
        kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE -o jsonpath='{.data.audit\.json}' | \
        jq -r '.[] | select(.action == "optimization_analysis") | "Tenant: \(.tenant) | Action: \(.action) | Timestamp: \(.timestamp) | Recommendation: \(.metadata.recommendation)"' | \
        tail -20
        
        echo -e "\n${CYAN}🎯 Optimization Recommendations:${NC}"
        kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE -o jsonpath='{.data.audit\.json}' | \
        jq -r '.[] | select(.metadata.recommendation == "increase_limits" or .metadata.recommendation == "decrease_limits") | "Tenant: \(.tenant) | Action: \(.metadata.recommendation) | Reason: \(.metadata.reason)"' | \
        tail -10
        
        echo -e "\n${CYAN}💰 Cost Analysis Results:${NC}"
        kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE -o jsonpath='{.data.audit\.json}' | \
        jq -r '.[] | select(.action == "cost_analysis") | "Tenant: \(.tenant) | Current Cost: \(.metadata.current_cost) | Projected Cost: \(.metadata.projected_cost)"' | \
        tail -10
        
    else
        echo -e "${YELLOW}⚠️  Audit ConfigMap not found. Enable audit logging with storageType: 'configmap'${NC}"
        echo -e "${BLUE}💡 To enable: Set auditLog.storageType: 'configmap' in values.yaml${NC}"
    fi
}

function metrics_report() {
    echo -e "\n${GREEN}📈 Metrics-Based Report${NC}"
    echo "----------------------------------------"
    
    # Check if metrics are accessible
    POD=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=mimir-limit-optimizer -o jsonpath='{.items[0].metadata.name}')
    
    if [ -n "$POD" ]; then
        echo -e "${GREEN}✅ Found pod: $POD${NC}"
        
        echo -e "\n${CYAN}🔢 Key Metrics:${NC}"
        kubectl exec -n $NAMESPACE $POD -- curl -s http://localhost:8080/metrics | \
        grep -E "(tenant_analysis_total|optimization_recommendations_total|cost_analysis_total|dry_run_total)" | \
        head -20
        
        echo -e "\n${CYAN}👥 Tenant Count Metrics:${NC}"
        kubectl exec -n $NAMESPACE $POD -- curl -s http://localhost:8080/metrics | \
        grep -E "tenant_count" | head -10
        
    else
        echo -e "${YELLOW}⚠️  No pods found for metrics collection${NC}"
    fi
}

function tenant_summary() {
    echo -e "\n${GREEN}📊 Tenant Optimization Summary${NC}"
    echo "----------------------------------------"
    
    echo -e "${CYAN}Recent Log Analysis (last 100 lines):${NC}"
    
    # Get recent logs and analyze
    LOGS=$(kubectl logs -n $NAMESPACE deployment/$DEPLOYMENT --tail=100 2>/dev/null)
    
    if [ -n "$LOGS" ]; then
        echo -e "\n${YELLOW}🎯 Tenants Needing Optimization:${NC}"
        echo "$LOGS" | grep -E "optimization_recommended|increase_limits|decrease_limits" | \
        sed 's/.*tenant[":]*\([^"]*\).*/\1/' | sort | uniq -c | sort -nr
        
        echo -e "\n${YELLOW}💰 Cost-Related Alerts:${NC}"
        echo "$LOGS" | grep -E "cost_exceeded|budget_warning" | \
        sed 's/.*tenant[":]*\([^"]*\).*/\1/' | sort | uniq -c | sort -nr
        
        echo -e "\n${YELLOW}📈 Spike Detections:${NC}"
        echo "$LOGS" | grep -E "spike_detected|event_spike" | \
        sed 's/.*tenant[":]*\([^"]*\).*/\1/' | sort | uniq -c | sort -nr
        
        echo -e "\n${YELLOW}🧪 Dry-Run Activities:${NC}"
        echo "$LOGS" | grep "dry-run" | wc -l | xargs echo "Total dry-run operations:"
        
    else
        echo -e "${YELLOW}⚠️  No recent logs found${NC}"
    fi
}

function detailed_tenant_report() {
    if [ -n "$1" ]; then
        echo -e "\n${GREEN}🔍 Detailed Report for Tenant: $1${NC}"
        echo "----------------------------------------"
        
        kubectl logs -n $NAMESPACE deployment/$DEPLOYMENT | \
        grep "$1" | \
        tail -20 | \
        while read line; do
            if echo $line | grep -q "optimization"; then
                echo -e "${GREEN}✅ OPTIMIZATION: $line${NC}"
            elif echo $line | grep -q "cost"; then
                echo -e "${YELLOW}💰 COST: $line${NC}"
            elif echo $line | grep -q "spike"; then
                echo -e "${RED}📈 SPIKE: $line${NC}"
            else
                echo -e "${CYAN}📋 INFO: $line${NC}"
            fi
        done
    fi
}

# Main menu
case "${2:-menu}" in
    "live")
        check_deployment
        live_optimization_feed
        ;;
    "audit")
        check_deployment
        audit_log_report
        ;;
    "metrics")
        check_deployment
        metrics_report
        ;;
    "summary")
        check_deployment
        tenant_summary
        ;;
    "tenant")
        check_deployment
        detailed_tenant_report "$3"
        ;;
    "all")
        check_deployment
        tenant_summary
        audit_log_report
        metrics_report
        ;;
    "menu"|*)
        show_usage
        echo -e "${BLUE}📋 Available Report Types:${NC}"
        echo "  $0 $NAMESPACE live     # Live feed of optimization decisions"
        echo "  $0 $NAMESPACE audit    # Audit log report from ConfigMap"
        echo "  $0 $NAMESPACE metrics  # Metrics-based tenant analysis"
        echo "  $0 $NAMESPACE summary  # Quick tenant optimization summary"
        echo "  $0 $NAMESPACE tenant <name>  # Detailed report for specific tenant"
        echo "  $0 $NAMESPACE all      # All reports"
        echo ""
        echo -e "${YELLOW}💡 Examples:${NC}"
        echo "  $0 mimir-optimizer live"
        echo "  $0 mimir-optimizer tenant enterprise-customer"
        echo "  $0 mimir-optimizer all"
        ;;
esac 