# Grafana Dashboard Queries for Mimir Limit Optimizer

## 📊 **Tenant Optimization Overview**

### **1. Tenants Needing Optimization (Top 10)**
```promql
topk(10, increase(mimir_limit_optimizer_tenant_analysis_total{result="optimization_recommended"}[5m]))
```

### **2. Dry-Run Operations Rate**
```promql
rate(mimir_limit_optimizer_dry_run_total[5m])
```

### **3. Cost Analysis per Tenant**
```promql
mimir_limit_optimizer_tenant_cost_analysis{tenant!=""}
```

## 🎯 **Optimization Recommendations**

### **4. Limit Increase Recommendations**
```promql
increase(mimir_limit_optimizer_optimization_recommendations_total{recommendation="increase_limits"}[1h])
```

### **5. Limit Decrease Recommendations**
```promql
increase(mimir_limit_optimizer_optimization_recommendations_total{recommendation="decrease_limits"}[1h])
```

### **6. Optimization Success Rate**
```promql
rate(mimir_limit_optimizer_optimization_recommendations_total{result="success"}[5m]) / 
rate(mimir_limit_optimizer_optimization_recommendations_total[5m]) * 100
```

## 💰 **Cost Control Metrics**

### **7. Budget Utilization by Tenant**
```promql
(mimir_limit_optimizer_tenant_cost_current / mimir_limit_optimizer_tenant_budget_limit) * 100
```

### **8. Tenants Exceeding Budget**
```promql
count(mimir_limit_optimizer_tenant_cost_current > mimir_limit_optimizer_tenant_budget_limit)
```

### **9. Total Cost Savings**
```promql
sum(mimir_limit_optimizer_cost_savings_total)
```

## 📈 **Spike Detection**

### **10. Traffic Spikes Detected**
```promql
increase(mimir_limit_optimizer_spike_detection_total[5m])
```

### **11. Circuit Breaker Activations**
```promql
increase(mimir_limit_optimizer_circuit_breaker_state_changes_total{state="open"}[5m])
```

### **12. Top Tenants by Spike Frequency**
```promql
topk(5, increase(mimir_limit_optimizer_spike_detection_total[1h]))
```

## 🔍 **Analysis Metrics**

### **13. Tenant Analysis Rate**
```promql
rate(mimir_limit_optimizer_tenant_analysis_total[5m])
```

### **14. Active Tenants Count**
```promql
mimir_limit_optimizer_active_tenants_count
```

### **15. Audit Log Entries Rate**
```promql
rate(mimir_limit_optimizer_audit_entries_total[5m])
```

## 🚨 **Alerting Queries**

### **16. High Error Rate**
```promql
rate(mimir_limit_optimizer_errors_total[5m]) > 0.1
```

### **17. Stale Tenant Data**
```promql
time() - mimir_limit_optimizer_last_analysis_timestamp > 3600
```

### **18. Configuration Errors**
```promql
increase(mimir_limit_optimizer_config_errors_total[5m]) > 0
```

## 📋 **Dashboard Panel Examples**

### **Single Stat Panels:**
- Total Active Tenants: `mimir_limit_optimizer_active_tenants_count`
- Dry-Run Operations Today: `increase(mimir_limit_optimizer_dry_run_total[24h])`
- Cost Savings This Month: `increase(mimir_limit_optimizer_cost_savings_total[30d])`

### **Time Series Panels:**
- Optimization Rate: `rate(mimir_limit_optimizer_optimization_recommendations_total[5m])`
- Cost Analysis Rate: `rate(mimir_limit_optimizer_cost_analysis_total[5m])`
- Spike Detection Rate: `rate(mimir_limit_optimizer_spike_detection_total[5m])`

### **Table Panels:**
```promql
# Top Tenants by Optimization Needs
sort_desc(
  sum by (tenant) (
    increase(mimir_limit_optimizer_tenant_analysis_total{result="optimization_recommended"}[24h])
  )
)
```

### **Heatmap Panels:**
```promql
# Optimization Activity by Hour
sum by (hour) (
  increase(mimir_limit_optimizer_optimization_recommendations_total[1h])
)
```

## 🔧 **Variables for Dashboard**

Create these variables in Grafana for dynamic filtering:

### **Tenant Variable:**
```promql
label_values(mimir_limit_optimizer_tenant_analysis_total, tenant)
```

### **Namespace Variable:**
```promql
label_values(mimir_limit_optimizer_tenant_analysis_total, namespace)
```

### **Recommendation Type Variable:**
```promql
label_values(mimir_limit_optimizer_optimization_recommendations_total, recommendation)
```

## 📊 **Alert Rules**

### **High Optimization Volume:**
```yaml
- alert: HighOptimizationVolume
  expr: rate(mimir_limit_optimizer_optimization_recommendations_total[5m]) > 10
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High volume of optimization recommendations"
```

### **Dry-Run Mode Stuck:**
```yaml
- alert: DryRunModeStuck
  expr: time() - mimir_limit_optimizer_last_dry_run_timestamp > 1800
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Dry-run mode appears to be stuck"
```

### **Budget Threshold Exceeded:**
```yaml
- alert: BudgetThresholdExceeded
  expr: mimir_limit_optimizer_tenant_cost_current / mimir_limit_optimizer_tenant_budget_limit > 0.9
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "Tenant {{ $labels.tenant }} approaching budget limit"
``` 