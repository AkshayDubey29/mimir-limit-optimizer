# 📦 Helm Chart: Audit Log Retention Configuration

## Overview

The Mimir Limit Optimizer Helm chart now includes comprehensive audit log retention configuration to prevent indefinite growth of audit logs in Kubernetes environments. This guide covers all the new Helm chart features related to audit log management.

## 🔄 **What's New in v1.3.0**

### **Enhanced Audit Log Configuration**
- ✅ **Comprehensive retention policies** - Time, count, and size-based retention
- ✅ **Environment-specific values files** - Production and development optimized settings
- ✅ **Automatic cleanup scheduling** - Background maintenance processes
- ✅ **Emergency protection** - Prevents ConfigMap size limit violations
- ✅ **Production-ready defaults** - Safe settings for immediate deployment

### **Updated Chart Version**
- **Chart Version**: `1.3.0` (updated from `1.2.5`)
- **App Version**: `v2.4.0` (updated from `1.2.5`)

## 📋 **Installation & Upgrade**

### **New Installation**
```bash
# Add the repository (if not already added)
helm repo add mimir-limit-optimizer https://your-chart-repo.com/

# Install with default settings
helm install mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer

# Install with production settings
helm install mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-production.yaml

# Install with development settings
helm install mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-development.yaml
```

### **Upgrade from Previous Versions**
```bash
# Upgrade to latest version (will automatically apply new retention settings)
helm upgrade mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer

# Upgrade with custom retention configuration
helm upgrade mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.retention.retentionPeriod="72h" \
  --set auditLog.retention.maxEntries="3000"
```

## ⚙️ **Configuration Reference**

### **Complete Audit Log Configuration**

```yaml
auditLog:
  # Enable audit logging
  enabled: true
  
  # Storage type: "memory", "configmap", or "external"
  storageType: "configmap"
  
  # Maximum entries (fallback limit)
  maxEntries: 2000
  
  # ConfigMap name for audit storage
  configMapName: "mimir-limit-optimizer-audit"
  
  # External storage configuration (for future use)
  externalStorage: {}
  
  # Comprehensive retention policies
  retention:
    # Time-based retention
    retentionPeriod: "168h"  # 7 days
    
    # Count-based retention
    maxEntries: 2000
    
    # Size-based retention (ConfigMap limit protection)
    maxSizeBytes: 819200  # 800KB
    
    # Cleanup scheduling
    cleanupInterval: "1h"
    
    # Batch processing
    cleanupBatchSize: 100
    
    # Emergency cleanup threshold
    emergencyThresholdPercent: 90.0
```

### **Configuration Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `auditLog.enabled` | bool | `true` | Enable audit logging |
| `auditLog.storageType` | string | `"configmap"` | Storage backend |
| `auditLog.maxEntries` | int | `2000` | Fallback entry limit |
| `auditLog.configMapName` | string | `"mimir-limit-optimizer-audit"` | ConfigMap name |
| `auditLog.retention.retentionPeriod` | duration | `"168h"` | How long to keep entries |
| `auditLog.retention.maxEntries` | int | `2000` | Maximum number of entries |
| `auditLog.retention.maxSizeBytes` | int | `819200` | Maximum ConfigMap size |
| `auditLog.retention.cleanupInterval` | duration | `"1h"` | Cleanup frequency |
| `auditLog.retention.cleanupBatchSize` | int | `100` | Cleanup batch size |
| `auditLog.retention.emergencyThresholdPercent` | float | `90.0` | Emergency cleanup trigger |

## 🏗️ **Environment-Specific Configurations**

### **Production Environment (`values-production.yaml`)**

**Optimized for high-volume audit logging:**

```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 5000  # Higher capacity for production
  retention:
    retentionPeriod: "168h"  # 7 days (compliance)
    maxEntries: 5000
    maxSizeBytes: 900000  # 900KB (near 1MB limit)
    cleanupInterval: "30m"  # Frequent cleanup
    cleanupBatchSize: 200  # Large batches
    emergencyThresholdPercent: 85.0  # Earlier emergency
```

**Deploy Production:**
```bash
helm install mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-production.yaml \
  --namespace mimir-system \
  --create-namespace
```

### **Development Environment (`values-development.yaml`)**

**Optimized for low-volume development:**

```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 500  # Lower capacity for development
  configMapName: "mimir-limit-optimizer-audit-dev"
  retention:
    retentionPeriod: "24h"  # 1 day only
    maxEntries: 500
    maxSizeBytes: 409600  # 400KB
    cleanupInterval: "2h"  # Less frequent cleanup
    cleanupBatchSize: 50  # Small batches
    emergencyThresholdPercent: 90.0
```

**Deploy Development:**
```bash
helm install mimir-limit-optimizer-dev mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-development.yaml \
  --namespace mimir-dev \
  --create-namespace
```

## 🔧 **Custom Configuration Examples**

### **High-Frequency Environment**
For environments with very high audit volume:

```yaml
auditLog:
  retention:
    retentionPeriod: "48h"  # Shorter retention
    maxEntries: 10000  # Very high capacity
    maxSizeBytes: 950000  # Close to limit
    cleanupInterval: "15m"  # Very frequent cleanup
    cleanupBatchSize: 500  # Large batches
    emergencyThresholdPercent: 80.0  # Early emergency
```

### **Compliance Environment**
For environments with long audit retention requirements:

```yaml
auditLog:
  retention:
    retentionPeriod: "720h"  # 30 days
    maxEntries: 3000  # Moderate capacity
    maxSizeBytes: 800000  # Safe size
    cleanupInterval: "6h"  # Less frequent cleanup
    cleanupBatchSize: 100  # Standard batches
    emergencyThresholdPercent: 95.0  # Late emergency
```

### **Memory-Only Environment**
For environments where ConfigMap persistence isn't needed:

```yaml
auditLog:
  storageType: "memory"
  maxEntries: 1000
  retention:
    retentionPeriod: "24h"
    maxEntries: 1000
    cleanupInterval: "30m"
    # maxSizeBytes not applicable for memory storage
```

## 🚀 **Deployment Commands**

### **Quick Start Commands**

```bash
# Default installation
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer

# Production with high audit volume
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.retention.maxEntries=5000 \
  --set auditLog.retention.cleanupInterval="30m" \
  --set auditLog.retention.emergencyThresholdPercent=85.0

# Development with minimal audit retention
helm install mimir-optimizer-dev mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.retention.retentionPeriod="24h" \
  --set auditLog.retention.maxEntries=500 \
  --set auditLog.retention.maxSizeBytes=409600

# Custom ConfigMap name and namespace
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.configMapName="custom-audit-logs" \
  --namespace custom-namespace \
  --create-namespace
```

### **Upgrade Commands**

```bash
# Upgrade with new retention settings
helm upgrade mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.retention.retentionPeriod="72h"

# Upgrade to production values
helm upgrade mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-production.yaml

# Upgrade with memory storage
helm upgrade mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.storageType="memory" \
  --set auditLog.maxEntries=2000
```

## 🔍 **Monitoring & Troubleshooting**

### **Check Audit Log ConfigMap**
```bash
# Check ConfigMap size
kubectl get configmap mimir-limit-optimizer-audit -o yaml | wc -c

# View audit log entries
kubectl get configmap mimir-limit-optimizer-audit -o yaml | yq '.data["audit.json"]'

# Check ConfigMap labels
kubectl get configmap mimir-limit-optimizer-audit --show-labels
```

### **Monitor Cleanup Logs**
```bash
# View cleanup logs
kubectl logs -l app.kubernetes.io/name=mimir-limit-optimizer | grep "retention"

# Watch real-time cleanup
kubectl logs -f -l app.kubernetes.io/name=mimir-limit-optimizer | grep "cleanup"
```

### **Emergency Cleanup Monitoring**
```bash
# Check for emergency cleanup events
kubectl logs -l app.kubernetes.io/name=mimir-limit-optimizer | grep "emergency"

# Monitor ConfigMap size in real-time
watch 'kubectl get configmap mimir-limit-optimizer-audit -o yaml | wc -c'
```

## 🛠️ **RBAC Configuration**

The Helm chart automatically creates the necessary RBAC permissions:

```yaml
# ClusterRole permissions for ConfigMaps
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]

# Role permissions for audit ConfigMap
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  resourceNames: ["mimir-limit-optimizer-audit"]
```

## 📊 **Migration Guide**

### **From v1.2.x to v1.3.0**

1. **Backup Current Audit Logs** (if any):
   ```bash
   kubectl get configmap mimir-limit-optimizer-audit -o yaml > audit-backup.yaml
   ```

2. **Upgrade the Chart**:
   ```bash
   helm upgrade mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer
   ```

3. **Verify New Configuration**:
   ```bash
   kubectl get configmap mimir-limit-optimizer-config -o yaml | grep -A 10 "retention:"
   ```

4. **Monitor Initial Cleanup**:
   ```bash
   kubectl logs -f -l app.kubernetes.io/name=mimir-limit-optimizer
   ```

### **Configuration Migration**

**Old Configuration (v1.2.x):**
```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 1000
```

**New Configuration (v1.3.0):**
```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 2000  # Increased default
  retention:
    retentionPeriod: "168h"
    maxEntries: 2000
    maxSizeBytes: 819200
    cleanupInterval: "1h"
    cleanupBatchSize: 100
    emergencyThresholdPercent: 90.0
```

## ✅ **Best Practices**

1. **Production Deployments**:
   - Use `values-production.yaml` as starting point
   - Monitor ConfigMap size during initial deployment
   - Set appropriate retention based on compliance requirements

2. **Development Deployments**:
   - Use `values-development.yaml` for testing
   - Enable debug logging for troubleshooting
   - Use shorter retention periods to save resources

3. **High-Volume Environments**:
   - Decrease `cleanupInterval` to "15m" or "30m"
   - Increase `cleanupBatchSize` to 200-500
   - Set `emergencyThresholdPercent` to 80-85%

4. **Resource Management**:
   - Monitor memory usage with high `maxEntries`
   - Adjust `maxSizeBytes` based on Kubernetes version
   - Use memory storage for ephemeral environments

---

**📚 Additional Resources:**
- [Audit Log Retention System Documentation](../../docs/AUDIT_LOG_RETENTION.md)
- [Implementation Details](../../AUDIT_LOG_RETENTION_IMPLEMENTATION.md)
- [Chart Values Reference](./values.yaml) 