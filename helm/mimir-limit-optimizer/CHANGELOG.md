# Helm Chart Changelog

## [1.3.0] - 2025-06-30

### 🎯 **Major Feature: Comprehensive Audit Log Retention System**

This release introduces a complete audit log retention system to prevent indefinite growth of audit logs in both memory and ConfigMap storage.

### ✨ **Added**

#### **Enhanced Audit Log Configuration**
- **Comprehensive retention policies** with time, count, and size-based limits
- **Configurable cleanup scheduling** with background maintenance processes
- **Emergency protection** to prevent ConfigMap size limit violations (1MB Kubernetes limit)
- **Environment-specific value files** for production and development optimized settings

#### **New Configuration Options**
```yaml
auditLog:
  retention:
    retentionPeriod: "168h"        # Time-based retention
    maxEntries: 2000               # Count-based retention  
    maxSizeBytes: 819200           # Size-based retention (800KB)
    cleanupInterval: "1h"          # Cleanup scheduling
    cleanupBatchSize: 100          # Batch processing
    emergencyThresholdPercent: 90.0 # Emergency cleanup threshold
```

#### **Environment-Specific Values Files**
- **`values-production.yaml`**: High-volume audit logging with aggressive retention
  - 5000 max entries, 900KB size limit, 30-minute cleanup intervals
- **`values-development.yaml`**: Conservative settings for development
  - 500 max entries, 400KB size limit, 2-hour cleanup intervals

#### **Enhanced Chart Metadata**
- Updated chart version from `1.2.5` to `1.3.0`
- Updated app version from `1.2.5` to `v2.4.0`
- Comprehensive Helm chart documentation

### 🔧 **Changed**

#### **Updated Default Values**
- Increased default `auditLog.maxEntries` from `1000` to `2000`
- Changed default `auditLog.storageType` remains `"configmap"` with enhanced configuration
- Added comprehensive retention configuration with safe production defaults

#### **Enhanced ConfigMap Template**
- Added retention configuration templating in `configmap.yaml`
- Support for external storage configuration (future use)
- Proper value type handling for all retention parameters

### 🛡️ **Security & RBAC**
- Existing RBAC permissions already include necessary ConfigMap operations
- Audit log ConfigMap permissions properly scoped to specific ConfigMap name
- No additional security changes required

### 📊 **Configuration Migration**

#### **Automatic Migration**
Existing deployments will automatically receive the new retention configuration on upgrade:

**Before (v1.2.x):**
```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 1000
```

**After (v1.3.0):**
```yaml
auditLog:
  enabled: true
  storageType: "configmap"
  maxEntries: 2000  # Automatically increased
  retention:
    retentionPeriod: "168h"      # NEW: 7 days retention
    maxEntries: 2000             # NEW: Enhanced count limit
    maxSizeBytes: 819200         # NEW: Size protection
    cleanupInterval: "1h"        # NEW: Scheduled cleanup
    cleanupBatchSize: 100        # NEW: Batch processing
    emergencyThresholdPercent: 90.0  # NEW: Emergency protection
```

### 📋 **Deployment Examples**

#### **Default Installation**
```bash
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer
```

#### **Production Installation**
```bash
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-production.yaml \
  --namespace mimir-system \
  --create-namespace
```

#### **Development Installation**
```bash
helm install mimir-optimizer-dev mimir-limit-optimizer/mimir-limit-optimizer \
  -f values-development.yaml \
  --namespace mimir-dev \
  --create-namespace
```

#### **Custom Retention Configuration**
```bash
helm install mimir-optimizer mimir-limit-optimizer/mimir-limit-optimizer \
  --set auditLog.retention.retentionPeriod="72h" \
  --set auditLog.retention.maxEntries="3000" \
  --set auditLog.retention.cleanupInterval="30m"
```

### 🔄 **Upgrade Instructions**

#### **From v1.2.x to v1.3.0**

1. **Backup existing audit logs** (if any):
   ```bash
   kubectl get configmap mimir-limit-optimizer-audit -o yaml > audit-backup.yaml
   ```

2. **Perform upgrade**:
   ```bash
   helm upgrade mimir-limit-optimizer mimir-limit-optimizer/mimir-limit-optimizer
   ```

3. **Verify new configuration**:
   ```bash
   kubectl get configmap mimir-limit-optimizer-config -o yaml | grep -A 10 "retention:"
   ```

4. **Monitor initial cleanup**:
   ```bash
   kubectl logs -f -l app.kubernetes.io/name=mimir-limit-optimizer | grep retention
   ```

### ⚡ **Performance Impact**

#### **Memory Usage**
- **Before**: Unlimited growth → potential OOM kill
- **After**: Bounded by `maxEntries` × average entry size

#### **ConfigMap Performance**
- **Before**: Linear degradation, potential 1MB limit violation
- **After**: Consistent performance, guaranteed under size limits

#### **Cleanup Overhead**
- **Background cleanup**: Minimal impact with configurable intervals
- **Batch processing**: Efficient cleanup with configurable batch sizes
- **Emergency protection**: Automatic prevention of performance issues

### 🧪 **Testing & Validation**

#### **Template Rendering**
- ✅ Default values render correctly
- ✅ Production values render correctly  
- ✅ Development values render correctly
- ✅ All retention parameters properly templated

#### **Compatibility**
- ✅ Backward compatible with existing deployments
- ✅ Automatic migration of configuration
- ✅ No breaking changes for existing users

### 📚 **Documentation**

#### **New Documentation Files**
- `AUDIT_LOG_HELM_GUIDE.md`: Comprehensive Helm chart guide
- `values-production.yaml`: Production-optimized configuration
- `values-development.yaml`: Development-optimized configuration

#### **Updated Documentation**
- Enhanced `values.yaml` with detailed retention configuration comments
- Updated Chart.yaml with new version and metadata

### 🎯 **Benefits Delivered**

1. **✅ Prevents System Failures**: No more OOM kills or ConfigMap limit errors
2. **✅ Production Ready**: Conservative defaults for immediate deployment
3. **✅ Environment Specific**: Optimized configurations for different use cases
4. **✅ Zero Maintenance**: Fully automated retention management
5. **✅ Configurable**: Tune all retention settings based on requirements
6. **✅ Emergency Protection**: Handles high-volume scenarios gracefully
7. **✅ Performance**: Consistent performance regardless of audit volume

---

## [1.2.5] - Previous Release
Previous release notes...

---

**📚 Additional Resources:**
- [Audit Log Retention Documentation](../../docs/AUDIT_LOG_RETENTION.md)
- [Helm Chart Configuration Guide](./AUDIT_LOG_HELM_GUIDE.md)
- [Implementation Details](../../AUDIT_LOG_RETENTION_IMPLEMENTATION.md) 