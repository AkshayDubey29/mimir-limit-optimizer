# 🗂️ Audit Log Retention System

## Overview

The Mimir Limit Optimizer now includes a comprehensive audit log retention system that prevents indefinite growth of audit logs in both memory and ConfigMap storage. This addresses the critical issue where audit logs would grow without bounds, potentially causing memory exhaustion or exceeding Kubernetes ConfigMap size limits.

## 🚨 **Why Retention is Critical**

Without proper retention, audit logs will:
- **Memory Mode**: Consume unlimited memory until OOM kill
- **ConfigMap Mode**: Exceed 1MB Kubernetes ConfigMap limit
- **Performance**: Degrade over time due to large datasets
- **Storage**: Increase storage costs indefinitely

## 🔧 **Retention Mechanisms**

### **1. Time-Based Retention**
- **Purpose**: Remove entries older than specified duration
- **Default**: 7 days (`168h`)
- **Configurable**: `retention.retentionPeriod`
- **Applied**: During both scheduled cleanup and entry addition

### **2. Count-Based Retention**
- **Purpose**: Limit total number of entries
- **Default**: 1000 entries (memory), 2000 entries (ConfigMap)
- **Configurable**: `retention.maxEntries`
- **Applied**: Keep most recent entries when limit exceeded

### **3. Size-Based Retention** (ConfigMap Only)
- **Purpose**: Stay under Kubernetes ConfigMap 1MB limit
- **Default**: 800KB (safe margin)
- **Configurable**: `retention.maxSizeBytes`
- **Applied**: Remove oldest entries to stay under limit

### **4. Emergency Cleanup**
- **Purpose**: Prevent system failures during high-volume periods
- **Default**: Trigger at 90% capacity
- **Configurable**: `retention.emergencyThresholdPercent`
- **Applied**: Immediate cleanup when thresholds exceeded

## ⚙️ **Configuration**

### **Complete Configuration Example**

```yaml
auditLog:
  enabled: true
  storageType: "configmap"  # "memory", "configmap", or "external"
  maxEntries: 2000         # Fallback limit
  configMapName: "mimir-limit-optimizer-audit"
  
  # Comprehensive retention policies
  retention:
    # Time-based retention - how long to keep entries
    retentionPeriod: "168h"        # 7 days
    
    # Count-based retention - maximum number of entries
    maxEntries: 2000               # Override root maxEntries
    
    # Size-based retention - maximum ConfigMap size
    maxSizeBytes: 819200           # 800KB (safe under 1MB limit)
    
    # Cleanup scheduling
    cleanupInterval: "1h"          # Run cleanup every hour
    cleanupBatchSize: 100          # Process 100 entries per batch
    
    # Emergency thresholds
    emergencyThresholdPercent: 90.0 # Emergency cleanup at 90%
```

### **Configuration Options**

| Setting | Type | Default | Purpose |
|---------|------|---------|---------|
| `retentionPeriod` | Duration | `168h` | How long to keep entries |
| `maxEntries` | Integer | `1000`/`2000` | Maximum number of entries |
| `maxSizeBytes` | Integer | `819200` | Maximum size in bytes |
| `cleanupInterval` | Duration | `1h` | How often to run cleanup |
| `cleanupBatchSize` | Integer | `100` | Batch size for cleanup |
| `emergencyThresholdPercent` | Float | `90.0` | Emergency cleanup trigger |

## 🔄 **Cleanup Scheduling**

### **Dual Cleanup Strategy**

1. **Scheduled Cleanup**
   - **Frequency**: Every `cleanupInterval` (default: 1 hour)
   - **Background**: Runs independently of reconciliation
   - **Purpose**: Regular maintenance

2. **Reconciliation Cleanup**
   - **Frequency**: During each reconciliation cycle
   - **Synchronous**: Part of main loop
   - **Purpose**: Ensure cleanup even if scheduled fails

### **Cleanup Process**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Time-Based     │───▶│  Count-Based    │───▶│  Size-Based     │
│  Retention      │    │  Retention      │    │  Retention      │
│  (7 days)       │    │  (2000 entries) │    │  (800KB)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Process Order:**
1. Remove entries older than `retentionPeriod`
2. If still over `maxEntries`, keep most recent entries
3. If still over `maxSizeBytes`, remove oldest until under limit
4. Log cleanup statistics

## 📊 **Storage-Specific Behavior**

### **Memory Storage**
- ✅ Time-based retention
- ✅ Count-based retention  
- ✅ Scheduled cleanup
- ❌ Size-based retention (not needed)

### **ConfigMap Storage**
- ✅ Time-based retention
- ✅ Count-based retention
- ✅ Size-based retention
- ✅ Scheduled cleanup
- ✅ Conflict resolution with retry logic

## 🔍 **Monitoring & Health**

### **Cleanup Metrics**
The system logs detailed cleanup information:

```json
{
  "level": "info",
  "msg": "retention policies applied",
  "original_entries": 2500,
  "cleaned_entries": 500,
  "remaining_entries": 2000,
  "emergency_cleanup": true
}
```

### **Size Monitoring**
```json
{
  "level": "info", 
  "msg": "size-based retention applied",
  "original_size_bytes": 950000,
  "target_size_bytes": 737280,
  "final_size_bytes": 720000,
  "entries_removed": 150
}
```

## 🚀 **Performance Characteristics**

### **Memory Usage**
- **Before**: Unlimited growth → OOM kill
- **After**: Bounded by `maxEntries` and `retentionPeriod`

### **ConfigMap Performance**
- **Before**: Linear degradation, eventual 1MB limit hit
- **After**: Consistent performance, stays under size limits

### **Cleanup Performance**
- **Batch Processing**: `cleanupBatchSize` entries at a time
- **Efficient Sorting**: Timestamp-based for chronological order
- **Minimal JSON Marshaling**: Only when necessary

## 🔧 **Troubleshooting**

### **Common Issues**

1. **High Memory Usage**
   ```yaml
   # Reduce retention settings
   retention:
     retentionPeriod: "24h"  # Reduce from 7 days
     maxEntries: 500         # Reduce from 2000
   ```

2. **ConfigMap Size Errors**
   ```yaml
   # Lower size limits
   retention:
     maxSizeBytes: 500000    # 500KB instead of 800KB
     emergencyThresholdPercent: 80.0  # Trigger earlier
   ```

3. **Frequent Emergency Cleanup**
   ```yaml
   # More aggressive regular cleanup
   retention:
     cleanupInterval: "30m"  # Every 30 minutes
     cleanupBatchSize: 200   # Larger batches
   ```

### **Debug Logging**
Enable verbose logging to see retention details:
```yaml
# In your deployment
env:
- name: LOG_LEVEL
  value: "1"  # Enable V(1) logs for retention details
```

## 🔒 **Security Considerations**

### **Data Loss Prevention**
- **Graduated Retention**: Time → Count → Size order prevents excessive deletion
- **Emergency Buffers**: 90% threshold leaves safety margin
- **Most Recent Priority**: Always keeps newest entries

### **Resource Protection**
- **Memory Bounds**: Prevents OOM conditions
- **Storage Limits**: Respects Kubernetes ConfigMap constraints
- **Performance Isolation**: Background cleanup doesn't block main operations

## 📈 **Recommended Settings**

### **Production Environment**
```yaml
retention:
  retentionPeriod: "168h"     # 7 days
  maxEntries: 2000            # 2000 entries  
  maxSizeBytes: 819200        # 800KB
  cleanupInterval: "1h"       # Every hour
  emergencyThresholdPercent: 90.0
```

### **Development Environment**
```yaml
retention:
  retentionPeriod: "24h"      # 1 day
  maxEntries: 500             # 500 entries
  maxSizeBytes: 409600        # 400KB
  cleanupInterval: "30m"      # Every 30 minutes
  emergencyThresholdPercent: 85.0
```

### **High-Volume Environment**
```yaml
retention:
  retentionPeriod: "72h"      # 3 days
  maxEntries: 5000            # 5000 entries
  maxSizeBytes: 900000        # 900KB
  cleanupInterval: "15m"      # Every 15 minutes
  cleanupBatchSize: 500       # Large batches
  emergencyThresholdPercent: 85.0
```

## 🎯 **Migration Guide**

### **From Previous Versions**

1. **Update Configuration**
   ```yaml
   # Old (v2.3.0 and earlier)
   auditLog:
     enabled: true
     maxEntries: 1000
   
   # New (v2.4.0+)
   auditLog:
     enabled: true
     maxEntries: 1000
     retention:
       retentionPeriod: "168h"
       maxEntries: 1000
       cleanupInterval: "1h"
   ```

2. **Monitor Initial Cleanup**
   - First startup may trigger large cleanup
   - Check logs for retention statistics
   - Adjust settings based on observed patterns

3. **Validate Settings**
   ```bash
   # Check ConfigMap size
   kubectl get configmap mimir-limit-optimizer-audit -o yaml | wc -c
   
   # Should be under 1MB (1048576 bytes)
   ```

## ✅ **Benefits**

1. **Prevents System Failures**: No more OOM kills or ConfigMap limit errors
2. **Predictable Performance**: Consistent memory and storage usage
3. **Configurable Retention**: Tune based on your compliance requirements
4. **Automatic Cleanup**: Zero-maintenance operation
5. **Emergency Protection**: Handles high-volume scenarios gracefully
6. **Production Ready**: Battle-tested retention algorithms

---

**Note**: This retention system is automatically enabled for all new installations. Existing installations should update their configuration to take advantage of the new features. 