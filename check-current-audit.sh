#!/bin/bash

# Quick Check Script for Current Audit Log Contents
# This script examines existing audit logs to understand the blank changes issue

echo "🔍 CHECKING CURRENT AUDIT LOG CONTENTS"
echo "======================================"
echo ""

# Function to check audit logs in memory or ConfigMap
check_audit_logs() {
    echo "📋 SEARCHING FOR AUDIT LOGS:"
    echo "============================="
    
    # Check if running in Kubernetes with ConfigMap storage
    if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null; then
        echo "✅ Kubernetes is accessible - checking for audit ConfigMaps"
        echo ""
        
        # Check common namespaces for audit ConfigMaps
        for namespace in mimir-limit-optimizer mimir default; do
            echo "🔍 Checking namespace: $namespace"
            if kubectl get namespace "$namespace" &> /dev/null; then
                if kubectl get configmap mimir-limit-optimizer-audit -n "$namespace" &> /dev/null; then
                    echo "✅ Found audit ConfigMap in namespace: $namespace"
                    echo ""
                    echo "📊 Raw audit data:"
                    kubectl get configmap mimir-limit-optimizer-audit -n "$namespace" -o yaml
                    echo ""
                    
                    # Try to parse and show specific entries
                    echo "🔍 Parsed audit entries:"
                    kubectl get configmap mimir-limit-optimizer-audit -n "$namespace" -o jsonpath='{.data.audit\.json}' 2>/dev/null | jq -r '.[] | "ID: \(.id), Tenant: \(.tenant // "N/A"), Action: \(.action), Changes: \(.changes), Reason: \(.reason)"' 2>/dev/null || {
                        echo "Unable to parse JSON - showing raw data:"
                        kubectl get configmap mimir-limit-optimizer-audit -n "$namespace" -o jsonpath='{.data.audit\.json}' 2>/dev/null | head -5
                    }
                    return 0
                else
                    echo "❌ No audit ConfigMap found in $namespace"
                fi
            else
                echo "❌ Namespace $namespace does not exist"
            fi
            echo ""
        done
        
        echo "❌ No audit ConfigMaps found in any namespace"
    else
        echo "❌ Kubernetes not accessible - cannot check ConfigMap audit logs"
    fi
    
    return 1
}

# Function to explain why changes might be blank
explain_blank_changes() {
    echo ""
    echo "❓ WHY ARE AUDIT LOG CHANGES BLANK?"
    echo "=================================="
    echo ""
    echo "The 'changes' field in audit logs will be blank when:"
    echo ""
    echo "1️⃣  🚫 NO TENANTS DISCOVERED"
    echo "   - No tenant names found in Mimir metrics"
    echo "   - No fallback tenants configured"
    echo "   - ConfigMap parsing failed"
    echo ""
    echo "2️⃣  📊 NO METRICS COLLECTED"
    echo "   - Metrics endpoint unreachable"
    echo "   - Authentication/authorization issues"
    echo "   - Wrong tenant headers (X-Scope-OrgID)"
    echo ""
    echo "3️⃣  🧮 NO LIMITS CALCULATED"
    echo "   - Dynamic limits disabled"
    echo "   - Insufficient metrics data"
    echo "   - All calculated limits same as current limits"
    echo ""
    echo "4️⃣  ⚙️  CONFIGURATION ISSUES"
    echo "   - Limit definitions not enabled"
    echo "   - Wrong analysis time windows"
    echo "   - Buffer percentages set to 0"
    echo ""
}

# Function to provide actionable steps
provide_action_steps() {
    echo ""
    echo "🔧 HOW TO FIX BLANK CHANGES:"
    echo "============================"
    echo ""
    echo "STEP 1: Check your configuration"
    echo "   🔍 Find your Mimir/Prometheus endpoint:"
    kubectl get services -A | grep -E "(prometheus|mimir|query)" 2>/dev/null || echo "      Run: kubectl get services -A | grep -E '(prometheus|mimir|query)'"
    echo ""
    
    echo "STEP 2: Verify tenant names"
    echo "   🔍 Check existing tenants in runtime overrides:"
    kubectl get configmap -A | grep "mimir-runtime-overrides" 2>/dev/null || echo "      Run: kubectl get configmap -A | grep 'mimir-runtime-overrides'"
    echo ""
    
    echo "STEP 3: Test metrics connectivity"
    echo "   🔍 Try a sample query:"
    echo "      curl -G 'http://your-mimir-endpoint:8080/api/v1/query' \\"
    echo "           --data-urlencode 'query=up' \\"
    echo "           -H 'X-Scope-OrgID: couwatch'"
    echo ""
    
    echo "STEP 4: Run the comprehensive debug script"
    echo "   🚀 Execute: ./debug-blank-changes.sh"
    echo "   📝 This will walk you through complete diagnosis"
    echo ""
}

# Function to show expected vs actual audit log format
show_audit_format() {
    echo ""
    echo "📋 EXPECTED AUDIT LOG FORMAT:"
    echo "============================="
    echo ""
    echo "✅ HEALTHY AUDIT ENTRY (with changes):"
    cat << 'EOF'
{
  "id": "audit_1735559123456",
  "timestamp": "2025-06-30T12:00:00Z",
  "tenant": "couwatch",
  "action": "update-limits",
  "reason": "trend-analysis",
  "changes": {
    "ingestion_rate": 15000,
    "max_global_series_per_user": 120000,
    "max_samples_per_query": 5000000
  },
  "success": true
}
EOF
    echo ""
    echo "❌ PROBLEMATIC AUDIT ENTRY (blank changes):"
    cat << 'EOF'
{
  "id": "audit_1735559123457",
  "timestamp": "2025-06-30T12:00:30Z",
  "tenant": "couwatch",
  "action": "update-limits", 
  "reason": "trend-analysis",
  "changes": {},
  "success": true
}
EOF
    echo ""
}

# Main execution
main() {
    check_audit_logs
    explain_blank_changes
    show_audit_format
    provide_action_steps
    
    echo ""
    echo "🎯 NEXT STEPS:"
    echo "=============="
    echo "1. Review the findings above"
    echo "2. Update debug-audit-config.yaml with your environment details"
    echo "3. Run: ./debug-blank-changes.sh for comprehensive analysis"
    echo "4. Share debug-output.log if you need further assistance"
    echo ""
}

# Run main function
main "$@" 