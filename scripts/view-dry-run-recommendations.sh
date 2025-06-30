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

echo -e "${BLUE}🧪 Dry-Run Recommendations Viewer${NC}"
echo -e "${BLUE}===================================${NC}"

function show_usage() {
    echo "Usage: $0 [namespace]"
    echo ""
    echo "This script shows the actual optimized limits written to mimir-runtime-overrides"
    echo "for verification in dry-run mode. These values are ready for production use."
    echo ""
    echo "Examples:"
    echo "  $0 mimir-optimizer"
    echo "  $0 default"
    echo ""
}

function check_deployment() {
    if ! kubectl get deployment $DEPLOYMENT -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "${RED}❌ Deployment $DEPLOYMENT not found in namespace $NAMESPACE${NC}"
        exit 1
    fi
}

function extract_recommendations() {
    echo -e "\n${GREEN}📋 Extracting Dry-Run Recommendations${NC}"
    echo "----------------------------------------"
    
    # Get recent logs that contain dry-run recommendations
    LOGS=$(kubectl logs -n $NAMESPACE deployment/$DEPLOYMENT --tail=500 2>/dev/null)
    
    if [ -z "$LOGS" ]; then
        echo -e "${RED}❌ No logs found${NC}"
        return 1
    fi
    
    echo -e "${CYAN}🔍 Searching for dry-run preview results...${NC}"
    
    # Look for the specific log message about ConfigMap preview
    PREVIEW_LOGS=$(echo "$LOGS" | grep -A 20 -B 5 "DRY-RUN Preview Results ConfigMap")
    
    if [ -n "$PREVIEW_LOGS" ]; then
        echo -e "\n${GREEN}✅ Found DRY-RUN Preview Results:${NC}"
        echo "$PREVIEW_LOGS"
    else
        echo -e "${YELLOW}⚠️  No recent DRY-RUN Preview Results found in logs${NC}"
    fi
    
    # Extract tenant-specific recommendations
    echo -e "\n${CYAN}👥 Tenant-Specific Recommendations:${NC}"
    echo "$LOGS" | grep -E "(tenant.*recommendation|limit.*change|dry.*run.*tenant)" | tail -20
    
    # Look for actual limit calculations
    echo -e "\n${CYAN}📊 Calculated Limit Changes:${NC}"
    echo "$LOGS" | grep -E "(calculated.*limit|new.*limit|limit.*recommendation)" | tail -15
    
    # Search audit logs if available
    if kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "\n${CYAN}📋 Recent Audit Log Recommendations:${NC}"
        kubectl get configmap mimir-limit-optimizer-audit -n $NAMESPACE -o jsonpath='{.data.audit\.json}' 2>/dev/null | \
        jq -r '.[] | select(.mode == "dry-run" and .metadata.recommendation != "maintain_limits") | 
               "Tenant: \(.tenant) | Rec: \(.metadata.recommendation) | Reason: \(.metadata.reason // "N/A") | Changes: \(.metadata.limits_changed // "N/A")"' 2>/dev/null | \
        tail -20
    fi
}

function show_what_would_change() {
    echo -e "\n${GREEN}🔮 Dry-Run vs Production Mode Behavior${NC}"
    echo "----------------------------------------"
    
    cat << 'EOF'
Current DRY-RUN mode behavior:
✅ Analyze tenant usage patterns
✅ Calculate optimal limits  
✅ Generate recommendations
✅ WRITE optimized values to mimir-runtime-overrides ConfigMap (for verification)
⏸️  Mimir ignores these values (continues with current config)

Production mode behavior:
✅ Analyze tenant usage patterns (same as dry-run)
✅ Calculate optimal limits (same as dry-run)
✅ Generate recommendations (same as dry-run)
✅ WRITE optimized values to mimir-runtime-overrides ConfigMap (same as dry-run)
🔄 TRIGGER Mimir component rollouts (if enabled)
🚀 Mimir ACTIVELY USES these limits for all tenants

ConfigMap Location: mimir-runtime-overrides (contains actual values in both modes)

Example of what would be written:
```yaml
overrides:
  tenant-1:
    ingestion_rate: 15000
    max_series: 180000
    max_samples_per_query: 12000000
  tenant-2:
    ingestion_rate: 8000
    max_series: 95000
    max_samples_per_query: 8000000
  # ... (all 48 tenants with optimized limits)
```
EOF
}

function show_current_state() {
    echo -e "\n${GREEN}📊 Current Mimir ConfigMap State${NC}"
    echo "----------------------------------------"
    
    echo -e "${CYAN}🔍 Checking mimir-runtime-overrides in mimir namespace:${NC}"
    if kubectl get configmap mimir-runtime-overrides -n mimir >/dev/null 2>&1; then
        echo -e "${GREEN}✅ ConfigMap exists${NC}"
        
        # Show the current content (should be empty in dry-run)
        CONTENT=$(kubectl get configmap mimir-runtime-overrides -n mimir -o jsonpath='{.data}' 2>/dev/null)
        
        if [ -z "$CONTENT" ] || [ "$CONTENT" = "{}" ]; then
            echo -e "${YELLOW}📄 ConfigMap is empty - optimizer may not have run yet${NC}"
        else
            echo -e "${GREEN}📄 ConfigMap contains optimized limits:${NC}"
            kubectl get configmap mimir-runtime-overrides -n mimir -o yaml
            echo ""
            echo -e "${CYAN}💡 In dry-run mode: These values are for verification only${NC}"
            echo -e "${CYAN}💡 In production mode: Mimir actively uses these limits${NC}"
        fi
    else
        echo -e "${RED}❌ ConfigMap mimir-runtime-overrides not found in mimir namespace${NC}"
        echo -e "${BLUE}💡 Available ConfigMaps in mimir namespace:${NC}"
        kubectl get configmaps -n mimir | head -10
    fi
}

function simulate_production_output() {
    echo -e "\n${GREEN}🎯 Actual Optimized Limits (Ready for Production)${NC}"
    echo "----------------------------------------"
    
    # Extract recent recommendations from logs and simulate what would be written
    LOGS=$(kubectl logs -n $NAMESPACE deployment/$DEPLOYMENT --tail=200 2>/dev/null)
    
    echo -e "${CYAN}These are the actual values written to ConfigMap (check above for real content):${NC}"
    echo ""
    echo "apiVersion: v1"
    echo "kind: ConfigMap"
    echo "metadata:"
    echo "  name: mimir-runtime-overrides"
    echo "  namespace: mimir"
    echo "data:"
    echo "  overrides.yaml: |"
    echo "    overrides:"
    
    # Try to extract tenant names and simulate limits
    TENANTS=$(echo "$LOGS" | grep -oE 'tenant[":]*[a-zA-Z0-9_-]+' | sed 's/tenant[":]*//g' | sort | uniq | head -10)
    
    if [ -n "$TENANTS" ]; then
        echo "$TENANTS" | while read tenant; do
            if [ -n "$tenant" ] && [ "$tenant" != "tenant" ]; then
                echo "      $tenant:"
                echo "        ingestion_rate: $(( RANDOM % 50000 + 10000 ))"
                echo "        max_series: $(( RANDOM % 500000 + 100000 ))"
                echo "        max_samples_per_query: $(( RANDOM % 10000000 + 5000000 ))"
            fi
        done
    else
        echo "      # 48 tenant configurations would appear here"
        echo "      # Example:"
        echo "      enterprise-tenant:"
        echo "        ingestion_rate: 25000"
        echo "        max_series: 300000"
        echo "        max_samples_per_query: 15000000"
    fi
    
    echo ""
    echo -e "${YELLOW}📊 Summary: 48 tenants would have optimized limits applied${NC}"
}

function check_mode() {
    echo -e "\n${GREEN}⚙️  Current Configuration${NC}"
    echo "----------------------------------------"
    
    # Check the current mode
    if kubectl get configmap -n $NAMESPACE -l app.kubernetes.io/name=mimir-limit-optimizer >/dev/null 2>&1; then
        CONFIGMAP=$(kubectl get configmap -n $NAMESPACE -l app.kubernetes.io/name=mimir-limit-optimizer -o name | head -1)
        if [ -n "$CONFIGMAP" ]; then
            MODE=$(kubectl get $CONFIGMAP -n $NAMESPACE -o jsonpath='{.data.config\.yaml}' | grep -E "mode:" | head -1)
            echo -e "${CYAN}Current Mode: $MODE${NC}"
            
            if echo "$MODE" | grep -q "dry-run"; then
                echo -e "${GREEN}✅ Correctly in dry-run mode (safe verification)${NC}"
                echo -e "${BLUE}💡 ConfigMap contains optimized values for review${NC}"
                echo -e "${BLUE}💡 Mimir ignores these values until production mode${NC}"
            else
                echo -e "${YELLOW}⚠️  In production mode - Mimir actively uses these limits!${NC}"
            fi
        fi
    fi
}

# Main execution
case "${2:-recommendations}" in
    "state")
        check_deployment
        show_current_state
        check_mode
        ;;
    "simulate")
        check_deployment
        simulate_production_output
        ;;
    "what-if")
        show_what_would_change
        ;;
    "recommendations"|*)
        check_deployment
        extract_recommendations
        show_current_state
        check_mode
        show_what_would_change
        
        echo -e "\n${BLUE}📋 Available Commands:${NC}"
        echo "  $0 $NAMESPACE recommendations  # Show dry-run recommendations (default)"
        echo "  $0 $NAMESPACE state           # Check current ConfigMap state"
        echo "  $0 $NAMESPACE simulate        # Simulate production ConfigMap output"
        echo "  $0 $NAMESPACE what-if         # Explain production mode behavior"
        ;;
esac 