# 🚀 Mimir Limit Optimizer: Enterprise Guard Rail System with Dynamic Limits

<div align="center">

[![CI/CD Pipeline](https://github.com/AkshayDubey29/mimir-limit-optimizer/actions/workflows/ci.yml/badge.svg)](https://github.com/AkshayDubey29/mimir-limit-optimizer/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/AkshayDubey29/mimir-limit-optimizer/branch/main/graph/badge.svg)](https://codecov.io/gh/AkshayDubey29/mimir-limit-optimizer)
[![Go Report Card](https://goreportcard.com/badge/github.com/AkshayDubey29/mimir-limit-optimizer)](https://goreportcard.com/report/github.com/AkshayDubey29/mimir-limit-optimizer)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=AkshayDubey29_mimir-limit-optimizer&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=AkshayDubey29_mimir-limit-optimizer)

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/AkshayDubey29/mimir-limit-optimizer?sort=semver&color=green)](https://github.com/AkshayDubey29/mimir-limit-optimizer/releases)
[![Docker Image Size](https://img.shields.io/docker/image-size/ghcr.io/akshaydubey29/mimir-limit-optimizer/latest?label=Docker%20Image)](https://github.com/AkshayDubey29/mimir-limit-optimizer/pkgs/container/mimir-limit-optimizer)
[![GitHub stars](https://img.shields.io/github/stars/AkshayDubey29/mimir-limit-optimizer?style=social)](https://github.com/AkshayDubey29/mimir-limit-optimizer/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/AkshayDubey29/mimir-limit-optimizer?style=social)](https://github.com/AkshayDubey29/mimir-limit-optimizer/network/members)

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.24+-326ce5?style=for-the-badge&logo=kubernetes)](https://kubernetes.io/)
[![Helm](https://img.shields.io/badge/Helm-3.0+-0F1689?style=for-the-badge&logo=helm)](https://helm.sh/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge)](CONTRIBUTING.md)
[![GitHub issues](https://img.shields.io/github/issues/AkshayDubey29/mimir-limit-optimizer?style=for-the-badge)](https://github.com/AkshayDubey29/mimir-limit-optimizer/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/AkshayDubey29/mimir-limit-optimizer?style=for-the-badge)](https://github.com/AkshayDubey29/mimir-limit-optimizer/pulls)

</div>

**Mimir Limit Optimizer** is an **enterprise-grade Kubernetes-native controller** that transforms Grafana Mimir into a **self-protecting, cost-aware, and intelligently optimized** observability platform. It acts as a comprehensive **Guard Rail for Mimir**, providing real-time protection against metric blasts, automated cost control, and intelligent performance optimization.

> **🎯 NEW: Revolutionary Dynamic Limits System**  
> **Supporting 30+ Mimir configuration parameters** with intelligent type-safe optimization, replacing hardcoded approaches with enterprise-grade flexibility. **Zero code changes** required to add new limits!  
> 📚 **Complete Guide:** [Dynamic Limits Documentation](docs/DYNAMIC-LIMITS.md)

## 🎯 **Dynamic Limits System - NEW Enterprise Feature**

**Revolutionary dynamic limit management supporting 30+ Mimir configuration parameters, replacing hardcoded approaches with intelligent, type-safe optimization.**

### **🚀 From Hardcoded to Enterprise-Grade Dynamic**

| **Previous Limitation** | **New Dynamic System** |
|------------------------|------------------------|
| ❌ **4 hardcoded limits only** | ✅ **30+ Mimir limits supported** |
| ❌ **Manual configuration required** | ✅ **Automated limit discovery** |
| ❌ **Code changes to add limits** | ✅ **Configuration-driven approach** |
| ❌ **Type-unsafe handling** | ✅ **Type-safe with validation** |
| ❌ **No per-limit customization** | ✅ **Per-limit min/max/buffers** |

### **📊 Supported Mimir Limits (30+)**

The dynamic system intelligently manages all major Mimir limit categories:

#### **🔄 Ingestion Limits** 
- `ingestion_rate` - Sample ingestion rate (samples/sec)
- `ingestion_burst_size` - Burst capacity for ingestion spikes
- `max_global_series_per_user` - Maximum series per tenant
- `max_global_series_per_metric` - Maximum series per metric name

#### **⚡ Query Performance Limits**
- `max_samples_per_query` - Maximum samples per query execution
- `max_series_per_query` - Maximum series a query can return
- `max_fetched_chunks_per_query` - Chunk fetch limits for performance
- `max_fetched_series_per_query` - Series fetch optimization
- `max_fetched_chunk_bytes_per_query` - Memory-aware chunk limits
- `max_query_lookback` - Historical query time window limits
- `max_query_length` - Query range duration limits

#### **🏷️ Metadata & Label Limits**
- `max_global_metadata_per_user` - Metadata entries per tenant
- `max_global_metadata_per_metric` - Metadata per metric optimization
- `max_label_names_per_series` - Label cardinality control
- `max_label_name_length` - Label name size limits
- `max_label_value_length` - Label value size optimization

#### **📈 Exemplar & Specialized Limits**
- `max_global_exemplars_per_user` - Exemplar storage limits
- `out_of_order_time_window` - Time-series ordering tolerance

#### **📏 Ruler & Alerting Limits**
- `ruler_max_rules_per_rule_group` - Recording/alerting rule limits
- `ruler_max_rule_groups_per_tenant` - Rule group organization
- `alertmanager_notification_rate_limit` - Alert delivery control
- `alertmanager_max_dispatcher_aggregation_groups` - Alert grouping
- `alertmanager_max_alerts_count` - Active alert limits
- `alertmanager_max_alerts_size_bytes` - Alert payload size limits

#### **🌐 Request & Network Limits**
- `request_rate` - API request rate limiting
- `request_burst_size` - Request burst capacity

#### **💾 Storage & Compactor Limits**
- `compactor_blocks_retention_period` - Data retention optimization
- `store_gateway_tenant_shard_size` - Query sharding efficiency

### **🛠️ Dynamic Configuration**

**Basic Dynamic Limits Setup:**
```yaml
dynamicLimits:
  enabled: true
  defaultBuffer: 20.0
  autoDetect: true
  
  # Enable specific limits for optimization
  enabledLimits:
    - ingestion_rate
    - ingestion_burst_size
    - max_global_series_per_user
    - max_samples_per_query
    - max_fetched_chunks_per_query
    # Add any limits you need
```

**Advanced Per-Limit Configuration:**
```yaml
dynamicLimits:
  limitOverrides:
    ingestion_rate:
      defaultValue: 50000.0
      minValue: 5000.0
      maxValue: 5000000.0
      bufferFactor: 30.0
      enabled: true
    max_global_series_per_user:
      defaultValue: 200000.0
      minValue: 10000.0
      maxValue: 50000000.0
      bufferFactor: 25.0
      enabled: true
```

**Environment-Specific Configurations:**

🧪 **Development Environment:**
```yaml
dynamicLimits:
  enabledLimits:
    - ingestion_rate
    - max_global_series_per_user
  limitOverrides:
    ingestion_rate:
      defaultValue: 10000.0
      bufferFactor: 50.0  # Higher buffer for testing
```

🏭 **Production Environment:**
```yaml
dynamicLimits:
  enabledLimits:
    - ingestion_rate
    - ingestion_burst_size
    - max_global_series_per_user
    - max_samples_per_query
    - max_fetched_chunks_per_query
    - max_fetched_series_per_query
    - max_fetched_chunk_bytes_per_query
    - max_global_exemplars_per_user
    - request_rate
    - request_burst_size
  limitOverrides:
    ingestion_rate:
      bufferFactor: 20.0  # Conservative for production
```

### **🎯 Key Benefits**

✅ **Enterprise Flexibility** - Support for all Mimir limits, not just 4  
✅ **Zero Code Changes** - Add new limits via configuration only  
✅ **Type Safety** - Intelligent type conversion and validation  
✅ **Granular Control** - Per-limit enable/disable and customization  
✅ **Environment Adaptive** - Different configs for dev/staging/prod  
✅ **Backward Compatible** - Existing deployments continue working  
✅ **Full Observability** - Metrics and audit logs for all limit types  

### **🔄 Migration from Hardcoded System**

**Automatic Migration:** Existing hardcoded configurations automatically work with the new dynamic system.

**Enhanced Migration:**
```yaml
# Old hardcoded approach (still works)
limits:
  ingestion_rate: 25000
  max_series: 150000

# New dynamic approach (recommended)
dynamicLimits:
  enabled: true
  enabledLimits:
    - ingestion_rate
    - ingestion_burst_size  
    - max_global_series_per_user
    - max_samples_per_query
    - max_fetched_chunks_per_query
    # Enable additional limits as needed
```

**📚 Complete Documentation:** [Dynamic Limits Guide](docs/DYNAMIC-LIMITS.md)

## 🎛️ **Web UI Dashboard - NEW Enterprise Feature**

**Modern React-based web interface providing real-time monitoring, configuration management, and comprehensive analytics for your Mimir Limit Optimizer deployment.**

### **🚀 Dashboard Overview**

The Web UI transforms the command-line experience into an intuitive, enterprise-grade dashboard with **7 comprehensive pages** for complete system visibility and control.

| **Page** | **Purpose** | **Key Features** |
|----------|-------------|------------------|
| **📊 Dashboard** | System overview & metrics | Real-time stats, tenant performance, limit usage |
| **👥 Tenants** | Tenant management | Per-tenant limits, usage analytics, cost tracking |
| **📈 Metrics** | Metrics monitoring | Live charts, trend analysis, performance indicators |
| **⚙️ Config** | Configuration management | Dynamic limit configuration, YAML editor |
| **🔍 Audit Log** | Audit & compliance | Complete activity history, limit changes, events |
| **📋 Diff Viewer** | Change visualization | Before/after comparisons, configuration diffs |
| **🛠️ Test Tools** | Testing & validation | Synthetic data generation, configuration testing |

### **✨ Key UI Features**

#### **🎨 Modern Design & UX**
- **Dark/Light Theme**: Automatic theme switching with user preference
- **Responsive Design**: Optimized for desktop, tablet, and mobile
- **Real-time Updates**: Live data refresh with WebSocket connections
- **Interactive Charts**: Recharts integration for beautiful data visualization

#### **📊 Comprehensive Analytics**
- **Tenant Performance**: Per-tenant ingestion rates, series counts, query performance
- **Cost Analytics**: Real-time cost tracking, budget utilization, forecasting
- **Limit Utilization**: Visual representation of limit usage vs. capacity
- **Historical Trends**: 24h, 7d, 30d trend analysis with percentile calculations

#### **⚙️ Configuration Management**
- **Dynamic Limits Editor**: Visual editor for all 30+ Mimir limits
- **YAML Configuration**: Syntax-highlighted YAML editor with validation
- **Live Preview**: Real-time configuration validation and preview
- **Change Management**: Track configuration changes with diff visualization

#### **🔍 Advanced Monitoring**
- **Real-time Metrics**: Live dashboard with auto-refresh
- **Alert Integration**: Visual alert status and notification center
- **Circuit Breaker Status**: Real-time circuit breaker state visualization
- **System Health**: Component status, resource utilization, performance metrics

### **🎛️ UI Configuration**

#### **Basic UI Setup** (Default: Enabled)
```yaml
# Helm values.yaml
ui:
  enabled: true       # Enable/disable the web dashboard
  port: 8082         # UI server port (avoid conflict with health probe on 8081)
  
  service:
    type: ClusterIP   # Service type for UI access
    port: 8082        # Service port
    annotations: {}   # Service annotations
  
  ingress:
    enabled: false    # Enable for external access
    className: ""     # Ingress class
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
```

#### **External Access Configuration**
```yaml
# Enable external access via Ingress
ui:
  enabled: true
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
    hosts:
      - host: mimir-optimizer.yourdomain.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.yourdomain.com
```

#### **Production Security Configuration**
```yaml
# Enhanced security for production
ui:
  enabled: true
  service:
    annotations:
      # Network policies
      networking.gke.io/allow-tcp-8082: "true"
      # Service mesh integration
      linkerd.io/inject: enabled
  
  ingress:
    enabled: true
    annotations:
      # Security headers
      nginx.ingress.kubernetes.io/configuration-snippet: |
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
      # Rate limiting
      nginx.ingress.kubernetes.io/rate-limit: "100"
      nginx.ingress.kubernetes.io/rate-limit-window: "1m"
```

### **🚀 Quick UI Access**

#### **Local Access (Port Forward)**
```bash
# Access UI via port forwarding (development/testing)
kubectl port-forward -n mimir-limit-optimizer \
  svc/mimir-limit-optimizer-ui 8082:8082

# Open browser: http://localhost:8082
```

#### **Internal Cluster Access**
```bash
# Access via cluster DNS (from within cluster)
curl http://mimir-limit-optimizer-ui.mimir-limit-optimizer.svc.cluster.local:8082

# Service discovery
kubectl get svc -n mimir-limit-optimizer mimir-limit-optimizer-ui
```

#### **External Access (Production)**
```bash
# Deploy with external access
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.ingress.enabled=true \
  --set ui.ingress.hosts[0].host=mimir-optimizer.yourdomain.com \
  --set ui.ingress.className=nginx \
  --namespace mimir-limit-optimizer

# Access via: https://mimir-optimizer.yourdomain.com
```

### **🎯 UI Deployment Examples**

#### **Enable UI in Existing Deployment**
```bash
# Add UI to existing deployment
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.port=8082 \
  --reuse-values
```

#### **Disable UI for Minimal Deployment**
```bash
# Disable UI for resource-constrained environments
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=false \
  --reuse-values
```

#### **UI with LoadBalancer (Cloud Environments)**
```bash
# Direct LoadBalancer access (AWS/GCP/Azure)
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.service.type=LoadBalancer \
  --set ui.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-type"=nlb \
  --reuse-values
```

### **📊 UI Pages Deep Dive**

#### **1. Dashboard Page**
- **System Overview**: Tenant count, active limits, system health
- **Real-time Metrics**: Ingestion rates, query performance, error rates
- **Resource Utilization**: CPU, memory, storage usage across components
- **Quick Actions**: Emergency controls, circuit breaker toggle

#### **2. Tenants Page**
- **Tenant List**: All active tenants with status indicators
- **Per-tenant Analytics**: Individual limit usage, cost tracking
- **Tenant Configuration**: View and modify tenant-specific settings
- **Usage Patterns**: Historical usage trends and predictions

#### **3. Metrics Page**
- **Live Charts**: Real-time metrics visualization with Recharts
- **Time Series Analysis**: Historical trends with configurable time ranges
- **Performance Indicators**: SLA metrics, availability, latency
- **Custom Queries**: PromQL query interface for advanced users

#### **4. Configuration Page**
- **Dynamic Limits**: Visual editor for all 30+ Mimir limits
- **YAML Editor**: Syntax-highlighted configuration editor
- **Validation**: Real-time configuration validation and error checking  
- **Templates**: Pre-configured templates for common scenarios

#### **5. Audit Log Page**
- **Activity History**: Complete log of all system activities
- **Change Tracking**: Detailed history of limit changes and updates
- **Search & Filter**: Advanced filtering by tenant, action, time range
- **Export**: CSV/JSON export for compliance and reporting

#### **6. Diff Viewer Page**
- **Configuration Changes**: Visual diff of configuration changes
- **Before/After**: Side-by-side comparison of settings
- **Impact Analysis**: Predicted impact of configuration changes
- **Rollback**: Easy rollback to previous configurations

#### **7. Test Tools Page**
- **Synthetic Data**: Generate test data for validation
- **Configuration Testing**: Test configurations before applying
- **Scenario Simulation**: Simulate various load scenarios
- **Validation Tools**: Verify system behavior under different conditions

### **🔧 Troubleshooting UI**

#### **Common Issues & Solutions**

**UI Not Accessible:**
```bash
# Check UI pod status
kubectl get pods -n mimir-limit-optimizer -l app.kubernetes.io/component=ui

# Check UI service
kubectl get svc -n mimir-limit-optimizer mimir-limit-optimizer-ui

# Check UI logs
kubectl logs -n mimir-limit-optimizer deployment/mimir-limit-optimizer -c ui
```

**Build Issues:**
```bash
# Verify UI build assets exist
kubectl exec -n mimir-limit-optimizer deployment/mimir-limit-optimizer -- ls -la /app/ui/build

# Check Go embed directive
kubectl exec -n mimir-limit-optimizer deployment/mimir-limit-optimizer -- find /app -name "*.go" -exec grep -l "go:embed" {} \;
```

**Port Conflicts:**
```bash
# Verify no port conflicts (UI: 8082, Health: 8081, Metrics: 8080)
kubectl exec -n mimir-limit-optimizer deployment/mimir-limit-optimizer -- netstat -tlnp
```

### **📈 UI Performance & Scaling**

#### **Resource Requirements**
```yaml
# Recommended resources with UI enabled
resources:
  requests:
    cpu: 200m      # Increased for UI server
    memory: 256Mi  # Increased for UI assets
  limits:
    cpu: 1000m     # Higher limit for UI + controller
    memory: 1Gi    # UI assets + React serving
```

#### **Production Optimizations**
- **CDN Integration**: Serve static assets via CDN for better performance
- **Caching**: Redis caching for API responses and static content
- **Load Balancing**: Multiple UI replicas behind load balancer
- **Monitoring**: UI-specific metrics and alerting

**Complete UI Documentation:** [Web UI Implementation Guide](WEB_UI_IMPLEMENTATION.md)

## 🎛️ **Operating Modes: Dry-Run vs Production**

The system operates in two distinct modes, each optimized for different phases of deployment and operational requirements:

### 🔍 **Dry-Run Mode** (Default)
**Purpose**: Safe observation and analysis without any system impact

| Component | Behavior | Purpose |
|-----------|----------|---------|
| **Limit Calculator** | ✅ Calculates optimal limits | Learn traffic patterns |
| **Limit Application** | ❌ **NO changes applied** | Zero risk observation |
| **Circuit Breaker** | ❌ **DISABLED** | Uninterrupted traffic flow |
| **Cost Control** | 🔍 **Monitoring only** | Budget tracking & alerts |
| **Emergency Systems** | ❌ **DISABLED** | No protective actions |
| **Audit Logging** | ✅ **Full logging** | Complete activity tracking |

**Use Cases:**
- 🆕 **Initial deployment** - Understand current traffic patterns
- 📊 **Capacity planning** - Analyze resource requirements
- 🧪 **Testing configurations** - Validate settings before production
- 📈 **Baseline establishment** - Build historical data for ML models

### 🚀 **Production Mode**
**Purpose**: Active protection and optimization with full feature set

| Component | Behavior | Purpose |
|-----------|----------|---------|
| **Limit Calculator** | ✅ Calculates optimal limits | Continuous optimization |
| **Limit Application** | ✅ **Changes applied to Mimir** | Active protection |
| **Circuit Breaker** | ✅ **ENABLED & Auto-configured** | Blast protection |
| **Cost Control** | 💰 **Full enforcement** | Budget protection |
| **Emergency Systems** | 🚨 **ACTIVE** | Panic mode & recovery |
| **Audit Logging** | ✅ **Full logging** | Compliance & debugging |

**Use Cases:**
- 🏭 **Production workloads** - Full protection and optimization
- 💰 **Cost optimization** - Automatic budget enforcement
- 🛡️ **Blast protection** - Real-time traffic spike protection
- 📈 **Auto-scaling** - Dynamic limit adjustments

## 🏢 **Enterprise Guard Rail Features**

Transform your Mimir deployment into a **cost-controlled, blast-protected enterprise system** with these advanced features:

### 💰 **Cost Control & Budget Management**

**Real-time cost monitoring with optional budget enforcement to prevent cost overruns while maintaining observability.**

#### **Key Features:**
- **🔍 Multiple Cost Models**: Sample-based, Series-based, Query-based, or Composite metrics
- **🎛️ Optional Enforcement**: Choose between monitoring-only or automatic limit reduction
- **📊 Multi-level Budgets**: Global, per-tenant, and per-team budget controls
- **⚠️ Smart Alerting**: Progressive alerts at 50%, 75%, 90%, 95%, and 100% thresholds
- **📈 Predictive Analytics**: Cost forecasting based on usage trends
- **💡 Optimization Recommendations**: AI-powered cost reduction suggestions

#### **Operating Modes:**

**🔍 Monitoring-Only Mode** (Default - Safe)
```yaml
costControl:
  autoLimitReduction: false  # No automatic limit changes
  budgetAlerts: true         # Alert on budget violations
  costOptimization: true     # Generate recommendations
```

**💰 Enforcement Mode** (Optional - Protective)
```yaml
costControl:
  autoLimitReduction: true   # Automatic limit reduction
  globalBudget:
    enforceBudget: true      # Global budget enforcement
  perTenantBudgets:
    enforceBudget: true      # Per-tenant enforcement
```

#### **Cost Calculation Methods:**
1. **Sample-based**: `cost = samples_per_second × sample_rate × time_period`
2. **Series-based**: `cost = active_series × series_rate × time_period`
3. **Query-based**: `cost = queries_per_second × query_rate × time_period`
4. **Composite**: `cost = (samples × 0.4) + (series × 0.4) + (queries × 0.2)`

#### **Dry-Run vs Production Behavior:**

| Feature | Dry-Run Mode | Production Mode |
|---------|--------------|-----------------|
| **Cost Calculation** | ✅ Full calculation | ✅ Full calculation |
| **Budget Tracking** | ✅ Complete tracking | ✅ Complete tracking |
| **Alerts** | ✅ All alerts sent | ✅ All alerts sent |
| **Limit Reduction** | ❌ **Never applied** | ✅ **Applied if enabled** |
| **Recommendations** | ✅ Generated | ✅ Generated |
| **Audit Logging** | ✅ Full logging | ✅ Full logging |

### 🛡️ **Blast Protection & Circuit Breaker**

**Intelligent protection system that automatically shields Mimir from traffic spikes and metric blasts using circuit breaker patterns and rate limiting.**

#### **How Circuit Breaker Works:**

The Circuit Breaker operates in **three states** to provide progressive protection:

1. **🟢 CLOSED** (Normal Operation)
   - All requests pass through normally
   - Continuous monitoring of failure rates and traffic patterns
   - Automatic threshold calculation based on tenant limits

2. **🔴 OPEN** (Protection Active)
   - Automatic rate limiting and request throttling
   - Blast protection algorithms active
   - Periodic testing for recovery conditions

3. **🟡 HALF-OPEN** (Testing Recovery)
   - Limited requests allowed through for testing
   - Success → return to CLOSED state
   - Failure → back to OPEN state

#### **Auto-Configuration System:**

**Smart Threshold Calculation:**
```yaml
# Automatic calculation based on current limits
ingestionThreshold = currentLimit × 1.5 + safetyMargin
queryThreshold = currentLimit × 2.0 + safetyMargin
seriesThreshold = currentLimit × 1.2 + safetyMargin
```

**Configuration Modes:**
- **🤖 Auto Mode**: Thresholds calculated from current tenant limits
- **🔧 Manual Mode**: Fixed thresholds set by administrators
- **🔄 Hybrid Mode**: Mix of auto and manual per tenant

**Real-time Adaptation:**
- **Learning Interval**: 5 minutes (configurable)
- **Adaptation Rate**: 10% of traffic change (configurable)
- **Safety Margins**: 10-20% buffer above calculated thresholds

#### **Blast Detection Algorithms:**

| Type | Trigger Condition | Response |
|------|------------------|----------|
| **Ingestion Spike** | > 1M samples/sec | Rate limiting + throttling |
| **Query Flood** | > 10K queries/sec | Query rate limiting |
| **Series Explosion** | > 100K new series/sec | Series creation throttling |
| **Memory Pressure** | > 80% memory usage | Emergency limits |
| **CPU Overload** | > 90% CPU usage | Panic mode |

#### **Dry-Run vs Production Behavior:**

| Feature | Dry-Run Mode | Production Mode |
|---------|--------------|-----------------|
| **Blast Detection** | ✅ Full detection | ✅ Full detection |
| **Threshold Calculation** | ✅ Calculated | ✅ Calculated |
| **Circuit Breaker** | ❌ **DISABLED** | ✅ **ENABLED** |
| **Rate Limiting** | ❌ **No limiting** | ✅ **Active limiting** |
| **Auto-Configuration** | ✅ Configuration only | ✅ **Applied to traffic** |
| **Metrics Collection** | ✅ Full metrics | ✅ Full metrics |

**Why Circuit Breaker is Disabled in Dry-Run:**
- 🔍 **Uninterrupted Observation**: Study actual traffic patterns without interference
- 📊 **Baseline Establishment**: Collect data for proper threshold calculation
- 🧪 **Configuration Testing**: Validate settings before enabling protection
- 📈 **Pattern Learning**: Build historical data for ML-based adaptation

#### **Runtime Control:**

**Enable/Disable without Restart:**
```bash
# Via ConfigMap patch
kubectl patch configmap mimir-limit-optimizer-config \
  --patch '{"data":{"circuitBreaker.runtimeEnabled":"true"}}'

# Via API (future enhancement)
curl -X POST /api/v1/circuit-breaker/enable
```

**Per-Tenant Control:**
```yaml
circuitBreaker:
  perTenantConfig:
    tenant-a:
      enabled: true
      mode: "auto"
    tenant-b:
      enabled: false
    tenant-c:
      mode: "manual"
      thresholds:
        ingestion: 500000
```

### 🚨 **Panic Mode & Emergency Controls**

**Last-resort protection system that activates during critical system overload to prevent complete system failure.**

#### **Emergency Escalation Levels:**

1. **🟡 Warning Level** (75% threshold)
   - Enhanced monitoring and alerting
   - Prepare for potential escalation
   - Notify operations team

2. **🟠 Emergency Level** (90% threshold)
   - Activate circuit breakers
   - Implement rate limiting
   - Reduce non-critical operations

3. **🔴 PANIC MODE** (95% threshold)
   - **Immediate Actions**: Reduce all limits by 50%
   - **Traffic Control**: Throttle all ingestion
   - **Resource Protection**: Disable non-critical queries
   - **System Alerts**: Notify all emergency contacts

#### **Panic Mode Triggers:**

| Resource | Trigger Condition | Action |
|----------|------------------|---------|
| **CPU Usage** | > 90% for 2 minutes | Reduce limits by 50% |
| **Memory Usage** | > 85% for 1 minute | Emergency throttling |
| **Disk Usage** | > 95% | Stop non-critical operations |
| **Network Overload** | > 80% bandwidth | Rate limit all traffic |
| **Mimir Failure** | Component crash | Force circuit breaker open |
| **Ingestion Overload** | > 5M samples/sec | Emergency ingestion limits |

#### **Emergency Actions:**

**Immediate Response:**
```yaml
panicMode:
  actions:
    - reduce_limits: 50%      # Reduce all limits by 50%
    - throttle_ingestion: 80% # Throttle to 20% capacity
    - disable_queries: true   # Disable non-critical queries
    - force_gc: true          # Force garbage collection
    - circuit_breaker: open   # Force circuit breaker open
```

**Progressive Actions:**
1. **Phase 1**: Warning alerts and enhanced monitoring
2. **Phase 2**: Circuit breaker activation and rate limiting
3. **Phase 3**: Emergency limits and traffic throttling
4. **Phase 4**: Panic mode with extreme protection measures

#### **Auto-Recovery System:**

**Recovery Process:**
1. **🔍 Health Checks**: Continuous monitoring of system health
2. **📈 Gradual Restoration**: Slowly increase limits (10% every 5 minutes)
3. **🎯 Validation**: Verify system stability at each step
4. **✅ Confirmation**: Return to normal operation when stable

**Recovery Conditions:**
- CPU usage < 70% for 10 minutes
- Memory usage < 75% for 10 minutes
- All Mimir components healthy
- No error rate spikes

#### **Emergency Contacts & Notifications:**

**Multi-Channel Alerts:**
- **📞 PagerDuty**: Critical incident escalation
- **📱 Slack**: Immediate team notification
- **📧 Email**: Management and ops team alerts
- **🔗 Webhooks**: Custom integrations (OpsGenie, ServiceNow)

**Alert Prioritization:**
- **P0**: Panic mode activation
- **P1**: Emergency level reached
- **P2**: Warning level threshold
- **P3**: Recovery completion

#### **Dry-Run vs Production Behavior:**

| Feature | Dry-Run Mode | Production Mode |
|---------|--------------|-----------------|
| **Resource Monitoring** | ✅ Full monitoring | ✅ Full monitoring |
| **Panic Detection** | ✅ Detects conditions | ✅ Detects conditions |
| **Emergency Actions** | ❌ **No actions taken** | ✅ **Full actions** |
| **Alerts** | ✅ All alerts sent | ✅ All alerts sent |
| **Recovery Process** | ❌ **Simulation only** | ✅ **Active recovery** |
| **Limit Changes** | ❌ **No changes** | ✅ **Applied immediately** |

**Dry-Run Panic Mode Benefits:**
- 🧪 **Test Alert Systems**: Validate notification channels
- 📊 **Threshold Tuning**: Adjust panic thresholds before production
- 🔍 **Scenario Planning**: Understand system behavior during stress
- 📈 **Baseline Establishment**: Learn normal vs. emergency patterns

### 📱 **Advanced Multi-Channel Alerting**

**Comprehensive alerting system with multiple channels, escalation policies, and intelligent routing.**

#### **Supported Channels:**
- **📱 Slack**: Real-time alerts with rich formatting and threading
- **📞 PagerDuty**: Critical incident management with on-call rotation
- **📧 Email**: SMTP-based notifications with HTML/text formats
- **🔗 Webhooks**: Custom integrations (OpsGenie, ServiceNow, Teams)
- **📊 Prometheus**: Alert Manager integration for advanced routing

#### **Alert Types & Prioritization:**
| Priority | Trigger | Channels | Escalation |
|----------|---------|----------|------------|
| **P0** | Panic mode, system failure | All channels | Immediate |
| **P1** | Emergency limits, circuit breaker | PagerDuty, Slack | 5 minutes |
| **P2** | Budget violations, high usage | Slack, Email | 15 minutes |
| **P3** | Recommendations, info | Email, Webhook | 1 hour |

#### **Intelligent Alert Routing:**
```yaml
alerting:
  routing:
    - match:
        severity: "critical"
      channels: ["pagerduty", "slack"]
      escalation: "immediate"
    - match:
        component: "cost-control"
      channels: ["email", "slack"]
      escalation: "15m"
```

### ⚡ **Performance Optimization**

**Advanced performance tuning for high-throughput environments with intelligent resource management.**

#### **Optimization Features:**
- **🧠 Intelligent Caching**: Memory/Redis with adaptive TTL
- **⚡ Batch Processing**: Concurrent operations with backpressure
- **🎯 Resource Optimization**: CPU/memory tuning with GC optimization
- **🗜️ Compression**: Multi-algorithm compression (gzip, lz4, snappy)
- **🔗 Connection Pooling**: Optimized database and API connections
- **📊 Metrics Optimization**: Efficient metrics collection and storage

## 🔄 **Complete System Integration Flow**

**This diagram shows how all enterprise features work together in the main controller loop:**

## 📊 **Monitoring & Observability**

**Comprehensive monitoring and observability features for complete system visibility and operational insights.**

#### **Core Features:**
- **📈 50+ Prometheus Metrics**: Detailed system and business metrics
- **📝 Comprehensive Audit Logging**: Complete action tracking with compliance support
- **🎯 Health Monitoring**: System health checks and component status
- **📊 Performance Dashboards**: Pre-built Grafana dashboards
- **🔍 Deep Debugging**: Detailed logging with configurable levels
- **🚨 Fault-Tolerant Alerting**: Resilient alerting that never blocks the system

#### **Available Metrics:**

**System Metrics:**
- `mimir_limit_optimizer_reconcile_total`
- `mimir_limit_optimizer_reconcile_duration_seconds`
- `mimir_limit_optimizer_errors_total`
- `mimir_limit_optimizer_limits_applied_total`

**Cost Control Metrics:**
- `mimir_limit_optimizer_cost_current`
- `mimir_limit_optimizer_budget_usage_ratio`
- `mimir_limit_optimizer_cost_recommendations_total`
- `mimir_limit_optimizer_budget_violations_total`

**Circuit Breaker Metrics:**
- `mimir_limit_optimizer_circuit_breaker_state`
- `mimir_limit_optimizer_rate_limit_requests_total`
- `mimir_limit_optimizer_blast_detections_total`
- `mimir_limit_optimizer_throttled_requests_total`

**Emergency System Metrics:**
- `mimir_limit_optimizer_panic_mode_activations_total`
- `mimir_limit_optimizer_emergency_actions_total`
- `mimir_limit_optimizer_recovery_attempts_total`
- `mimir_limit_optimizer_resource_usage_percent`

**Alerting Resilience Metrics:**
- `mimir_limit_optimizer_alert_delivery_total`
- `mimir_limit_optimizer_alert_delivery_duration_seconds`
- `mimir_limit_optimizer_alert_channel_health`
- `mimir_limit_optimizer_alert_channel_errors_total`
- `mimir_limit_optimizer_alert_retry_attempts_total`
- `mimir_limit_optimizer_alert_channel_circuit_breaker_state`
- `mimir_limit_optimizer_alert_queue_size`
- `mimir_limit_optimizer_alert_configuration_errors_total`
- `mimir_limit_optimizer_last_successful_alert_timestamp`
- `mimir_limit_optimizer_alert_channel_response_time_seconds`

#### **Audit Logging:**

**Log Categories:**
- **👤 User Actions**: Manual configuration changes
- **🤖 System Actions**: Automated limit adjustments
- **💰 Cost Events**: Budget violations and cost optimizations
- **🛡️ Security Events**: Circuit breaker activations and emergency actions
- **📊 Performance Events**: System health and optimization events

**Log Format:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "component": "cost-controller",
  "action": "budget_violation",
  "tenant": "tenant-a",
  "details": {
    "current_cost": 1500.00,
    "budget_limit": 1000.00,
    "percentage": 150,
    "action_taken": "reduce_limits",
    "new_limits": {
      "ingestion_rate": 50000,
      "query_rate": 1000
    }
  }
}
```

#### **Health Monitoring:**

**Health Check Endpoints:**
- `/health` - Overall system health
- `/ready` - Readiness probe
- `/metrics` - Prometheus metrics
- `/debug/pprof` - Performance profiling

**System Health Indicators:**
- **Controller Status**: Active/Inactive/Error
- **Mimir Connectivity**: API reachability and response times  
- **Cost Calculation**: Cost calculation accuracy and performance
- **Circuit Breaker**: Protection system status
- **Resource Utilization**: CPU, memory, and disk usage

## 🚨 **Fault-Tolerant Alerting System**

**Enterprise-grade alerting that guarantees system resilience - external alerting failures never impact core functionality.**

### **🛡️ Core Resilience Principles**

#### **1. Never Block the Main System**
```go
// ✅ Non-blocking alert sending
func (m *Manager) SendAlert(alert *Alert) {
    select {
    case m.alertQueue <- alert:
        // Alert queued successfully
    default:
        // Queue full - log error but continue
        log.Error("Alert queue full, dropping alert")
        metrics.IncAlertChannelErrors("queue", "queue_full")
    }
}
```

#### **2. Comprehensive Error Handling**
- **Configuration Errors**: Invalid Slack/PagerDuty/Email configs logged and metrics tracked
- **Network Failures**: Timeout and connection errors handled gracefully
- **Service Unavailability**: Circuit breakers prevent cascade failures
- **Rate Limiting**: Queue management and backpressure handling

#### **3. Circuit Breaker per Channel**
Each alerting channel has its own circuit breaker:
- **🟢 Closed**: Normal operation (failures < threshold)
- **🔴 Open**: Channel protection active (failures ≥ threshold)
- **🟡 Half-Open**: Testing recovery (limited requests allowed)

### **📊 Alerting Resilience Metrics**

**Every alerting operation is instrumented with detailed metrics:**

| Metric | Purpose | Labels |
|--------|---------|---------|
| `alert_delivery_total` | Track delivery success/failure | `channel`, `alert_type`, `result` |
| `alert_delivery_duration_seconds` | Monitor response times | `channel` |
| `alert_channel_health` | Channel health status (1=healthy, 0=unhealthy) | `channel` |
| `alert_channel_errors_total` | Error tracking by type | `channel`, `error_type` |
| `alert_retry_attempts_total` | Retry behavior monitoring | `channel`, `alert_type` |
| `alert_channel_circuit_breaker_state` | Circuit breaker status | `channel` |
| `alert_queue_size` | Queue depth monitoring | `channel` |
| `alert_configuration_errors_total` | Config validation errors | `channel`, `config_error_type` |
| `last_successful_alert_timestamp` | Last successful delivery | `channel` |
| `alert_channel_response_time_seconds` | End-to-end response time | `channel` |

### **🔄 Multi-Channel Architecture**

**The alerting system routes alerts to multiple channels with independent circuit breakers and health monitoring:**

### **📱 Supported Channels**

#### **1. Slack Integration**
- **Rich Formatting**: Color-coded alerts based on priority
- **Structured Data**: Alert details in organized fields
- **Channel Routing**: Configurable destination channels
- **Health Checks**: API endpoint validation
- **Error Handling**: Webhook failures tracked and retried

```yaml
alerting:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#mimir-alerts"
    username: "Mimir Limit Optimizer"
```

#### **2. PagerDuty Integration**
- **Incident Management**: Automatic incident creation
- **Priority Filtering**: Only P0 and P1 alerts sent
- **Deduplication**: Prevents alert spam
- **Escalation**: Follows your PagerDuty escalation policies
- **Resolution**: Auto-resolve when issues are fixed

```yaml
alerting:
  pagerduty:
    enabled: true
    integration_key: "your-integration-key"
    timeout: 30s
```

#### **3. Email Notifications**
- **SMTP Support**: Works with any SMTP server
- **Priority Headers**: Email priority based on alert level
- **Rich Content**: Detailed alert information
- **Multiple Recipients**: Support for distribution lists
- **TLS Support**: Secure email transmission

```yaml
alerting:
  email:
    enabled: true
    smtp_host: "smtp.company.com"
    smtp_port: 587
    username: "alerts@company.com"
    password: "secure-password"
    from: "mimir-alerts@company.com"
    to: ["ops@company.com", "oncall@company.com"]
    use_tls: true
```

#### **4. Custom Webhooks**
- **Generic Integration**: Works with any webhook endpoint
- **Custom Headers**: Support for authentication and routing
- **Configurable Methods**: POST, PUT, PATCH support
- **JSON Payload**: Structured alert data
- **Multiple Webhooks**: Support for different destinations

```yaml
alerting:
  webhooks:
    opsgenie:
      enabled: true
      url: "https://api.opsgenie.com/v1/json/alert"
      method: "POST"
      headers:
        Authorization: "GenieKey your-api-key"
    servicenow:
      enabled: true
      url: "https://company.service-now.com/api/now/table/incident"
      method: "POST"
      headers:
        Authorization: "Basic base64-encoded-credentials"
```

### **🔄 Fault-Tolerance Features**

#### **Queue Management**
- **Main Queue**: 1000 alert buffer (non-blocking)
- **Retry Queue**: 500 alert buffer (failed alerts)
- **Exponential Backoff**: Intelligent retry delays
- **Dead Letter Handling**: Permanent failure tracking

#### **Circuit Breaker Behavior**

| Channel | Failure Threshold | Recovery Timeout | Max Half-Open Calls |
|---------|------------------|------------------|-------------------|
| **Slack** | 5 failures | 5 minutes | 3 |
| **PagerDuty** | 3 failures | 5 minutes | 3 |
| **Email** | 5 failures | 10 minutes | 3 |
| **Webhooks** | 5 failures | 5 minutes | 3 |

#### **Error Types Tracked**
- `channel_not_found` - Channel configuration missing
- `circuit_breaker_open` - Circuit breaker protecting channel
- `send_failed` - Network/API failure
- `invalid_config` - Configuration validation error
- `timeout` - Request timeout exceeded
- `queue_full` - Alert queue overflow

### **📊 Operational Insights**

#### **Comprehensive Logging**
Every alerting operation is logged with structured data:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "component": "alerting-manager",
  "action": "alert_sent",
  "channel": "slack",
  "alert_id": "cost_violation-1642248600123",
  "alert_type": "cost_violation",
  "priority": "P2",
  "tenant": "tenant-a",
  "duration": "1.234s",
  "result": "success"
}
```

#### **Error Logging Examples**
```json
{
  "timestamp": "2024-01-15T10:35:00Z",
  "level": "ERROR",
  "component": "slack-channel",
  "action": "send_failed",
  "alert_id": "panic_mode-1642248900456",
  "error": "webhook returned status 500",
  "duration": "5.678s",
  "retry_count": 1,
  "circuit_breaker_state": "closed"
}
```

#### **Configuration Error Tracking**
```json
{
  "timestamp": "2024-01-15T10:40:00Z",
  "level": "WARN",
  "component": "alerting-manager", 
  "action": "channel_init_failed",
  "channel": "email",
  "error": "SMTP host configuration missing",
  "impact": "email alerts disabled until configuration fixed"
}
```

### **✅ System Resilience Guarantees**

#### **Core System Protection**
1. **Non-Blocking Operations**: Alert failures never block limit calculations
2. **Independent Operation**: Core functionality works even if all alerting fails
3. **Graceful Degradation**: System continues with reduced alerting capability
4. **Self-Healing**: Automatic recovery when channels become available

#### **Operational Continuity**
1. **Partial Channel Failures**: Working channels continue to operate
2. **Configuration Hot-Reload**: Fix configs without system restart
3. **Queue Persistence**: Alerts survive temporary system restarts
4. **Monitoring Integration**: Full observability of alerting health

#### **Enterprise Compliance**
1. **Audit Trail**: Complete record of all alerting attempts
2. **SLA Monitoring**: Response time and success rate tracking
3. **Capacity Planning**: Queue utilization and performance metrics
4. **Incident Response**: Detailed failure analysis and remediation

### **📊 Monitoring Dashboard Examples**

#### **Alerting Health Dashboard**

**Key Metrics to Monitor:**

```promql
# Channel Health Overview
mimir_limit_optimizer_alert_channel_health

# Alert Delivery Success Rate (last 24h)
rate(mimir_limit_optimizer_alert_delivery_total{result="success"}[24h]) / 
rate(mimir_limit_optimizer_alert_delivery_total[24h]) * 100

# Circuit Breaker States
mimir_limit_optimizer_alert_channel_circuit_breaker_state

# Queue Utilization
mimir_limit_optimizer_alert_queue_size

# Error Rate by Channel
rate(mimir_limit_optimizer_alert_channel_errors_total[5m])

# Average Response Time
rate(mimir_limit_optimizer_alert_channel_response_time_seconds_sum[5m]) / 
rate(mimir_limit_optimizer_alert_channel_response_time_seconds_count[5m])
```

#### **Sample Grafana Alert Rules**

```yaml
# Alert when any channel is unhealthy for > 5 minutes
- alert: AlertChannelUnhealthy
  expr: mimir_limit_optimizer_alert_channel_health == 0
  for: 5m
  annotations:
    summary: "Alert channel {{ $labels.channel }} is unhealthy"
    description: "Channel has failed health checks for > 5 minutes"

# Alert when circuit breaker is open
- alert: AlertChannelCircuitBreakerOpen
  expr: mimir_limit_optimizer_alert_channel_circuit_breaker_state == 1
  for: 1m
  annotations:
    summary: "Circuit breaker open for {{ $labels.channel }}"
    description: "Alert channel circuit breaker has opened due to failures"

# Alert when queue is getting full
- alert: AlertQueueNearFull
  expr: mimir_limit_optimizer_alert_queue_size > 800
  for: 2m
  annotations:
    summary: "Alert queue for {{ $labels.channel }} is nearly full"
    description: "Queue size: {{ $value }}/1000 - may start dropping alerts"
```

### **🔧 Troubleshooting Guide**

#### **Common Scenarios**

**Scenario 1: Slack Not Working**
```bash
# Check Slack channel health
kubectl logs -f deployment/mimir-limit-optimizer | grep "slack.*error"

# Expected behavior:
# ✅ System continues operating normally
# ✅ Other channels (email, pagerduty) still work
# ✅ Metrics show: alert_channel_health{channel="slack"} = 0
# ✅ Circuit breaker protects system from further failures
# ✅ Automatic retry attempts logged
```

**Scenario 2: PagerDuty Configuration Error**
```bash
# Check PagerDuty configuration errors
kubectl logs -f deployment/mimir-limit-optimizer | grep "pagerduty.*config"

# Expected behavior:
# ✅ System starts successfully despite bad config
# ✅ PagerDuty channel marked as unhealthy
# ✅ Metrics show: alert_configuration_errors_total{channel="pagerduty"} > 0
# ✅ P0/P1 alerts still go to other channels
# ✅ Configuration can be fixed without restart
```

**Scenario 3: All Alerting Channels Down**
```bash
# Check overall alerting status
kubectl logs -f deployment/mimir-limit-optimizer | grep "alerting.*error"

# Expected behavior:
# ✅ Core system continues limit optimization
# ✅ All actions logged in audit trail
# ✅ Metrics collection continues normally
# ✅ System doesn't crash or hang
# ✅ Alerts queued for retry when channels recover
```

#### **Recovery Procedures**

**1. Fix Channel Configuration**
```bash
# Update configuration without restart
kubectl patch configmap mimir-limit-optimizer-config \
  --patch '{"data":{"alerting.slack.webhook_url":"https://correct-webhook-url"}}'

# Monitor recovery
kubectl logs -f deployment/mimir-limit-optimizer | grep "channel.*initialized"
```

**2. Manual Circuit Breaker Reset**
```bash
# Check circuit breaker state
curl -s http://mimir-limit-optimizer:8080/debug/alerting/status | jq '.channels'

# Circuit breakers self-heal, but you can monitor recovery:
kubectl logs -f deployment/mimir-limit-optimizer | grep "circuit.*half-open"
```

**3. Queue Management**
```bash
# Monitor queue sizes
curl -s http://mimir-limit-optimizer:8080/metrics | grep alert_queue_size

# If queues are backing up, check channel health:
curl -s http://mimir-limit-optimizer:8080/debug/alerting/status
```

## 📋 Prerequisites
- **Helm**: 3.0+ for deployment
- **Grafana Mimir**: Deployed and accessible
- **Prometheus**: For metrics collection (or Mimir's built-in metrics)
- **Docker**: For building container images

## 🏃‍♂️ **Quick Start Guide**

### **Phase 1: Initial Setup (Safe Observation)**

#### 1. **Build & Prepare the Container**

```bash
# Clone the repository
git clone https://github.com/AkshayDubey29/mimir-limit-optimizer.git
cd mimir-limit-optimizer

# Build the Docker image
docker build -t mimir-limit-optimizer:latest .

# Optional: Tag and push to your registry
docker tag mimir-limit-optimizer:latest your-registry.com/mimir-limit-optimizer:latest
docker push your-registry.com/mimir-limit-optimizer:latest
```

#### 2. **Deploy in Dry-Run Mode** (Recommended First Step)

```bash
# Install in dry-run mode for safe observation
# ✅ Collects metrics and calculates optimal limits
# ✅ NEW: Dynamic limits system supports 30+ Mimir limits
# ✅ NEW: Web UI Dashboard enabled for monitoring
# ❌ NO changes applied to Mimir
# ❌ Circuit breaker DISABLED for uninterrupted traffic study
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set image.repository=your-registry.com/mimir-limit-optimizer \
  --set image.tag=latest \
  --set controller.mode=dry-run \
  --set mimir.namespace=mimir-system \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=false \
  --set dynamicLimits.enabled=true \
  --set ui.enabled=true \
  --set ui.port=8082 \
  --set-json='dynamicLimits.enabledLimits=["ingestion_rate","ingestion_burst_size","max_global_series_per_user","max_samples_per_query","max_fetched_chunks_per_query"]' \
  --namespace mimir-limit-optimizer \
  --create-namespace

# Monitor the deployment and observe logs
kubectl logs -f deployment/mimir-limit-optimizer -n mimir-limit-optimizer

# Access the Web UI Dashboard
kubectl port-forward -n mimir-limit-optimizer svc/mimir-limit-optimizer-ui 8082:8082
# Open browser: http://localhost:8082
```

#### 3. **Verify Dry-Run Results** (Recommended: 24-48 hours)

```bash
# Check audit logs for recommendations
kubectl logs deployment/mimir-limit-optimizer -n mimir-limit-optimizer | grep "recommendation"

# View calculated limits (no changes applied)
kubectl logs deployment/mimir-limit-optimizer -n mimir-limit-optimizer | grep "calculated_limits"

# Monitor cost tracking
kubectl logs deployment/mimir-limit-optimizer -n mimir-limit-optimizer | grep "cost_analysis"
```

### **Phase 2: Production Deployment (Active Protection)**

#### 4. **Switch to Production Mode**

```bash
# After verifying dry-run results, activate production mode
# ✅ Limit changes will be applied to Mimir
# ✅ Circuit breaker automatically ENABLED
# ✅ Full protection suite activated
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --reuse-values
```

### **Phase 3: Cost Control Configuration**

#### 5. **Choose Your Cost Control Strategy**

**Option A: Monitoring-Only Mode** (Default - Safe)
```bash
# Track costs and budgets with alerts only
# ✅ Complete cost visibility
# ✅ Budget violation alerts
# ❌ NO automatic limit reduction
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=false \
  --set costControl.budgetAlerts=true \
  --reuse-values
```

**Option B: Enforcement Mode** (Automatic Protection)
```bash
# Automatic cost protection with budget enforcement
# ✅ Automatic limit reduction on budget violations
# ✅ Global and per-tenant budget enforcement
# ⚠️ May impact high-volume tenants
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=true \
  --set costControl.globalBudget.enforceBudget=true \
  --reuse-values
```

### **Phase 4: Dynamic Limits Configuration**

#### 6. **Configure Dynamic Limits (NEW)**

**Enable All Major Limits:**
```bash
# Enable comprehensive limit optimization (30+ limits)
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.enabledLimits=[
    "ingestion_rate",
    "ingestion_burst_size",
    "max_global_series_per_user",
    "max_samples_per_query",
    "max_series_per_query",
    "max_fetched_chunks_per_query",
    "max_fetched_series_per_query",
    "max_fetched_chunk_bytes_per_query",
    "max_global_exemplars_per_user",
    "request_rate",
    "request_burst_size"
  ]' \
  --reuse-values
```

**High-Cardinality Environment:**
```bash
# Optimized for high-cardinality workloads
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.limitOverrides={
    "max_global_series_per_user": {
      "defaultValue": 1000000.0,
      "maxValue": 50000000.0,
      "bufferFactor": 15.0
    },
    "max_label_names_per_series": {
      "defaultValue": 50.0,
      "maxValue": 100.0
    }
  }' \
  --reuse-values
```

**Cost-Optimized Setup:**
```bash
# Tighter limits for cost optimization
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.limitOverrides={
    "ingestion_rate": {
      "bufferFactor": 10.0
    },
    "max_samples_per_query": {
      "defaultValue": 10000000.0,
      "bufferFactor": 25.0
    }
  }' \
  --reuse-values
```

### **Phase 5: Advanced Configuration**

#### 7. **Circuit Breaker Configuration**

**Runtime Control** (Enable/Disable without restart):
```bash
# Enable circuit breaker at runtime
kubectl patch configmap mimir-limit-optimizer-config -n mimir-limit-optimizer \
  --patch '{"data":{"circuitBreaker.runtimeEnabled":"true"}}'

# Disable circuit breaker at runtime
kubectl patch configmap mimir-limit-optimizer-config -n mimir-limit-optimizer \
  --patch '{"data":{"circuitBreaker.runtimeEnabled":"false"}}'
```

**Auto-Configuration Mode**:
```bash
# Enable auto-configuration based on tenant limits
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.autoConfiguration.enabled=true \
  --set circuitBreaker.autoConfiguration.mode=auto \
  --reuse-values
```

#### 7. **Multi-Channel Alerting Setup**

**Slack Integration**:
```bash
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set alerting.slack.enabled=true \
  --set alerting.slack.webhook="https://hooks.slack.com/your-webhook" \
  --set alerting.slack.channel="#mimir-alerts" \
  --reuse-values
```

**PagerDuty Integration**:
```bash
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set alerting.pagerduty.enabled=true \
  --set alerting.pagerduty.integrationKey="your-integration-key" \
  --reuse-values
```

## 🎯 **Deployment Scenarios & Best Practices**

### **Scenario 1: Development Environment**
**Goal**: Aggressive optimization with fast feedback loops and dynamic limits

```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --set controller.reconcileInterval=30s \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=true \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.enabledLimits=["ingestion_rate","max_global_series_per_user","max_samples_per_query"]' \
  --set-json='dynamicLimits.limitOverrides={"ingestion_rate":{"bufferFactor":50.0}}' \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.autoConfiguration.enabled=true \
  --set circuitBreaker.thresholds.ingestionMultiplier=1.2 \
  --set circuitBreaker.thresholds.queryMultiplier=1.5 \
  --set alerting.slack.enabled=true \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **Scenario 2: Staging Environment**
**Goal**: Production-like testing with safety margins and comprehensive dynamic limits

```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --set controller.reconcileInterval=2m \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=false \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.enabledLimits=[
    "ingestion_rate","ingestion_burst_size","max_global_series_per_user",
    "max_samples_per_query","max_fetched_chunks_per_query","max_fetched_series_per_query"
  ]' \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.autoConfiguration.enabled=true \
  --set circuitBreaker.thresholds.safetyMargin=0.2 \
  --set emergency.enabled=true \
  --set emergency.panicMode.enabled=false \
  --set alerting.slack.enabled=true \
  --set alerting.email.enabled=true \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **Scenario 3: Production Environment**
**Goal**: Conservative protection with high safety margins and full dynamic limits

```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --set controller.reconcileInterval=5m \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=false \
  --set dynamicLimits.enabled=true \
  --set-json='dynamicLimits.enabledLimits=[
    "ingestion_rate","ingestion_burst_size","max_global_series_per_user",
    "max_samples_per_query","max_series_per_query","max_fetched_chunks_per_query",
    "max_fetched_series_per_query","max_fetched_chunk_bytes_per_query",
    "max_global_exemplars_per_user","request_rate","request_burst_size"
  ]' \
  --set-json='dynamicLimits.limitOverrides={"ingestion_rate":{"bufferFactor":20.0}}' \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.autoConfiguration.enabled=true \
  --set circuitBreaker.thresholds.safetyMargin=0.3 \
  --set emergency.enabled=true \
  --set emergency.panicMode.enabled=true \
  --set alerting.pagerduty.enabled=true \
  --set alerting.slack.enabled=true \
  --set alerting.email.enabled=true \
  --set performance.caching.enabled=true \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **Scenario 4: High-Volume Production**
**Goal**: Maximum performance with enterprise features

```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --set controller.reconcileInterval=1m \
  --set costControl.enabled=true \
  --set costControl.autoLimitReduction=true \
  --set costControl.globalBudget.enabled=true \
  --set costControl.globalBudget.amount=50000 \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.autoConfiguration.enabled=true \
  --set circuitBreaker.autoConfiguration.mode=auto \
  --set emergency.enabled=true \
  --set emergency.panicMode.enabled=true \
  --set alerting.pagerduty.enabled=true \
  --set alerting.slack.enabled=true \
  --set performance.caching.enabled=true \
  --set performance.caching.type=redis \
  --set performance.batchProcessing.enabled=true \
  --set performance.compression.enabled=true \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **Scenario 5: Multi-Tenant SaaS**
**Goal**: Per-tenant isolation with granular controls

```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --set controller.reconcileInterval=2m \
  --set costControl.enabled=true \
  --set costControl.perTenantBudgets.enabled=true \
  --set circuitBreaker.enabled=true \
  --set circuitBreaker.perTenantConfig.enabled=true \
  --set tenantScoping.enabled=true \
  --set tenantScoping.tierBased.enabled=true \
  --set emergency.enabled=true \
  --set alerting.webhook.enabled=true \
  --set performance.caching.enabled=true \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **Best Practices Summary**

#### **🔍 Initial Deployment**
1. **Always start with dry-run mode** for 24-48 hours
2. **Monitor audit logs** for recommendations
3. **Validate cost calculations** before enforcement
4. **Test alert channels** before production

#### **💰 Cost Control**
1. **Start with monitoring-only mode** in production
2. **Set realistic budgets** based on historical data
3. **Use progressive enforcement** (50% → 75% → 100%)
4. **Monitor tenant impact** closely

#### **🛡️ Circuit Breaker**
1. **Use auto-configuration** for dynamic environments
2. **Set appropriate safety margins** (20-30% for production)
3. **Monitor false positive rates** and adjust thresholds
4. **Test recovery procedures** regularly

#### **🚨 Emergency Systems**
1. **Configure multiple alert channels** for redundancy
2. **Test panic mode** in staging environments
3. **Document recovery procedures** for operations team
4. **Regular disaster recovery drills**

## 🔧 Building the Docker Container

### Development Build

```bash
# Build locally for development
docker build -t mimir-limit-optimizer:dev .

# Run locally (requires kubeconfig)
docker run --rm -it \
  -v ~/.kube:/home/nonroot/.kube:ro \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  mimir-limit-optimizer:dev \
  --config=/config/config.yaml \
  --log-level=debug
```

### Production Build

```bash
# Multi-arch build for production
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 \
  -t your-registry.com/mimir-limit-optimizer:v0.1.0 \
  --push .
```

### Build Arguments

The Dockerfile supports the following build arguments:

```bash
docker build \
  --build-arg VERSION=v0.1.0 \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -t mimir-limit-optimizer:v0.1.0 .
```

## 📦 Helm Chart Deployment

### Installation

```bash
# Add your custom values
cat << EOF > custom-values.yaml
image:
  repository: your-registry.com/mimir-limit-optimizer
  tag: v3.0.0

controller:
  mode: dry-run
  bufferPercentage: 20
  updateInterval: "5m"

# Web UI Dashboard configuration
ui:
  enabled: true
  port: 8082
  service:
    type: ClusterIP
    port: 8082
  # Optional: Enable external access
  ingress:
    enabled: false
    className: ""
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix

mimir:
  namespace: mimir-system
  configMapName: mimir-runtime-overrides

tenantScoping:
  skipList:
    - "internal-*"
    - "test-*"
  includeList: []

# Dynamic Limits (NEW)
dynamicLimits:
  enabled: true
  enabledLimits:
    - ingestion_rate
    - max_global_series_per_user
    - max_samples_per_query

eventSpike:
  enabled: true
  threshold: 2.0
  cooldownPeriod: "30m"

# Updated resource requirements for UI
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
EOF

# Install with custom values
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  -f custom-values.yaml \
  --namespace mimir-limit-optimizer \
  --create-namespace

# Access the Web UI
kubectl port-forward -n mimir-limit-optimizer svc/mimir-limit-optimizer-ui 8082:8082
# Open browser: http://localhost:8082
```

### Upgrade

```bash
# Upgrade with new image
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set image.tag=v0.1.1 \
  --reuse-values
```

### Uninstall

```bash
helm uninstall mimir-limit-optimizer -n mimir-limit-optimizer
kubectl delete namespace mimir-limit-optimizer
```

## ⚙️ Configuration

### Environment Variables

Key environment variables for quick configuration:

```bash
# Operating mode
MODE=dry-run  # or "prod"

# Mimir settings
MIMIR_NAMESPACE=mimir-system
MIMIR_CONFIGMAP_NAME=mimir-runtime-overrides

# Update frequency
UPDATE_INTERVAL=5m

# Buffer percentage
BUFFER_PERCENTAGE=20
```

### Configuration File

Create a comprehensive configuration file:

```yaml
# config.yaml
mode: "dry-run"
bufferPercentage: 20
updateInterval: "5m"

# Web UI Dashboard configuration
ui:
  enabled: true
  port: 8082

mimir:
  namespace: "mimir-system"
  configMapName: "mimir-runtime-overrides"

# Enterprise Features Configuration
costControl:
  enabled: true
  costMethod: "composite"
  costPerUnit: 0.001  # $0.001 per million samples
  globalBudget:
    daily: 1000
    monthly: 30000
    annual: 365000
    currency: "USD"
    enforceBudget: false  # Optional: Set to true for budget enforcement
  alertThresholds: [50, 75, 90, 95]
  
  # IMPORTANT: Budget enforcement is OPTIONAL
  # false = Monitoring-only mode (alerts without limit changes)
  # true = Enforcement mode (automatic limit reduction when over budget)
  autoLimitReduction: false

circuitBreaker:
  enabled: true
  failureThreshold: 50.0
  requestVolumeThreshold: 20
  rateLimit:
    enabled: true
    requestsPerSecond: 100
    burstCapacity: 200
  blastProtection:
    ingestionSpikeThreshold: 1000000  # 1M samples/sec
    querySpikeThreshold: 10000        # 10K queries/sec
    autoEmergencyShutdown: true

emergency:
  enabled: true
  panicMode:
    enabled: true
    cpuThreshold: 90.0
    memoryThreshold: 90.0
    actions: ["reduce_limits", "throttle_ingestion", "alert"]

alerting:
  enabled: true
  slack:
    enabled: true
    webhookURL: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
    channel: "#mimir-alerts"

performance:
  enabled: true
  cache:
    enabled: true
    sizeMB: 256
    ttl: "5m"
  batchProcessing:
    enabled: true
    size: 100
    maxConcurrent: 10

# Dynamic Limits Configuration (NEW)
dynamicLimits:
  enabled: true
  defaultBuffer: 20.0
  autoDetect: true
  enabledLimits:
    - ingestion_rate
    - ingestion_burst_size
    - max_global_series_per_user
    - max_samples_per_query
    - max_fetched_chunks_per_query
    - max_series_per_query
    - max_fetched_series_per_query
    - max_global_exemplars_per_user
  
  # Per-limit customization (optional)
  limitOverrides:
    ingestion_rate:
      defaultValue: 50000.0
      minValue: 5000.0
      maxValue: 5000000.0
      bufferFactor: 25.0
    max_global_series_per_user:
      defaultValue: 200000.0
      minValue: 10000.0
      maxValue: 50000000.0
      bufferFactor: 20.0

# Standard Configuration
tenantScoping:
  skipList: ["internal-*", "staging-*"]
  includeList: []
  useRegex: false

metricsDiscovery:
  enabled: true
  namespace: "mimir-system"
  serviceLabelSelector: "app.kubernetes.io/name=mimir"
  serviceNames: ["mimir-distributor", "mimir-ingester", "mimir-querier"]
  metricsPath: "/metrics"
  portName: "http-metrics"

eventSpike:
  enabled: true
  threshold: 2.0
  detectionWindow: "5m"
  cooldownPeriod: "30m"
  maxSpikeMultiplier: 5.0

trendAnalysis:
  analysisWindow: "48h"
  percentile: 95.0
  useMovingAverage: true
  includePeaks: true
  timeOfDayBuffers:
    "9-17": 1.5   # Business hours
    "0-8": 0.8    # Off hours

limits:
  minLimits:
    ingestion_rate: 1000
    max_series: 10000
    max_samples_per_query: 1000000
  maxLimits:
    ingestion_rate: 1000000
    max_series: 10000000
    max_samples_per_query: 100000000
  tenantTiers:
    enterprise:
      bufferPercentage: 30
      limits:
        ingestion_rate: 500000
        max_series: 5000000
    standard:
      bufferPercentage: 20
      limits:
        ingestion_rate: 100000
        max_series: 1000000

auditLog:
  enabled: true
  storageType: "memory"
  maxEntries: 1000

# Dynamic Limits Configuration (NEW)
dynamicLimits:
  enabled: true
  defaultBuffer: 20.0
  autoDetect: true
  
  # Enable/disable specific limits for optimization
  # Choose which limits to optimize based on your needs
  enabledLimits:
    # Essential limits (recommended for all environments)
    - ingestion_rate
    - ingestion_burst_size
    - max_global_series_per_user
    - max_samples_per_query
    
    # Performance limits (recommended for high-load environments)
    - max_fetched_chunks_per_query
    - max_fetched_series_per_query
    - max_fetched_chunk_bytes_per_query
    
    # Additional limits (enable as needed)
    # - max_global_exemplars_per_user
    # - max_query_lookback
    # - max_query_length
    # - request_rate
    # - request_burst_size
    # - ruler_max_rules_per_rule_group
    # - alertmanager_notification_rate_limit
    # - max_label_names_per_series
    # - out_of_order_time_window
  
  # Override default values for specific limits
  limitOverrides:
    ingestion_rate:
      defaultValue: 50000.0
      minValue: 5000.0
      maxValue: 5000000.0
      bufferFactor: 30.0
      enabled: true
    max_global_series_per_user:
      defaultValue: 200000.0
      minValue: 10000.0
      maxValue: 50000000.0
      bufferFactor: 25.0
      enabled: true
    max_samples_per_query:
      defaultValue: 50000000.0
      minValue: 1000000.0
      maxValue: 500000000.0
      bufferFactor: 20.0
      enabled: true
```

## 💰 Cost Control Operating Modes

The cost control system supports **flexible operating modes** to suit different organizational needs:

### 🔍 **Monitoring-Only Mode** (Default)
Perfect for organizations that want cost visibility without automatic enforcement:

```yaml
costControl:
  enabled: true
  autoLimitReduction: false  # No automatic limit changes
  globalBudget:
    enforceBudget: false     # Alerts only, no enforcement
```

**Benefits:**
- ✅ Real-time cost tracking and alerts
- ✅ Budget violation notifications
- ✅ Cost optimization recommendations
- ❌ No automatic limit reductions
- ❌ No service disruption

### 🛡️ **Enforcement Mode** (Optional)
For organizations requiring automatic cost protection:

```yaml
costControl:
  enabled: true
  autoLimitReduction: true   # Enable automatic limit reduction
  globalBudget:
    enforceBudget: true      # Enforce budget limits
```

**Benefits:**
- ✅ All monitoring-only features
- ✅ Automatic limit reduction when over budget
- ✅ Prevents cost overruns
- ⚠️ May temporarily reduce service capacity

### 🎯 **Hybrid Mode** (Per-Tenant Control)
Mix monitoring and enforcement per tenant:

```yaml
costControl:
  enabled: true
  autoLimitReduction: true
  globalBudget:
    enforceBudget: false     # Global default: monitoring only
  tenantBudgets:
    critical-service:
      enforceBudget: false   # Monitoring only for critical services
         development:
       enforceBudget: true    # Enforcement for development tenants
```

## 🛡️ Automated Circuit Breaker System

The Circuit Breaker provides **intelligent blast protection** with automatic configuration based on your actual tenant limits and real-time traffic patterns.

> **📝 Important**: Circuit breaker is **automatically disabled in dry-run mode** to ensure uninterrupted traffic observation. This allows you to study actual traffic patterns and calculate appropriate limits before enabling protection in production mode.

### 🎯 **Operation Modes**

| Operating Mode | Circuit Breaker | Purpose | When to Use |
|----------------|----------------|---------|-------------|
| **🔍 Dry-Run** | **Disabled** | Observe traffic patterns | Initial deployment, analysis, testing |
| **🚀 Production** | **Enabled** | Active protection | Live environments with known traffic patterns |

### 🔧 **Configuration Modes** (When Enabled)

| Mode | Description | Use Case | Thresholds |
|------|-------------|----------|------------|
| **🔧 Manual** | Fixed thresholds set by operators | Static environments | User-defined values |
| **🤖 Auto** | Thresholds calculated from limits | Dynamic environments | 150% of ingestion limit, 200% of query limit |
| **🔀 Hybrid** | Mix of manual and auto per tenant | Mixed environments | Per-tenant configuration |

### 🚀 **Auto-Configuration Features**

#### **🔍 Intelligent Threshold Calculation**
```yaml
# Thresholds automatically calculated from current tenant limits
autoConfig:
  limitMultipliers:
    ingestionRateMultiplier: 1.5  # Trip at 150% of current limit
    queryRateMultiplier: 2.0      # Trip at 200% of current limit
    seriesMultiplier: 1.8         # Trip at 180% of current limit
    burstMultiplier: 1.2          # Trip at 120% of burst limit
```

**Example**: If a tenant has an ingestion limit of 100K samples/sec, the circuit breaker will trip at 150K samples/sec (plus safety margin).

#### **📈 Real-time Adaptation**
```yaml
# Self-tuning based on observed traffic patterns
realtimeAdaptation:
  enabled: true
  interval: "5m"                # Adapt every 5 minutes
  learningRate: 0.1             # How fast to adapt (10%)
  maxChangePercent: 20.0        # Maximum change per cycle
  percentile: 95.0              # Use 95th percentile for calculations
```

#### **🛡️ Safety Margins**
```yaml
# Configurable safety margins prevent false positives
safetyMargins:
  defaultMargin: 25.0          # 25% safety buffer
  tenantMargins:
    critical-service: 40.0     # Higher safety for critical services
    test-environment: 15.0     # Lower safety for testing
```

### 🎛️ **Runtime Control**

#### **Enable/Disable Without Restart**
```bash
# Enable circuit breaker
kubectl patch configmap mimir-limit-optimizer-config \
  --patch '{"data":{"circuitBreaker.runtimeEnabled":"true"}}'

# Disable circuit breaker
kubectl patch configmap mimir-limit-optimizer-config \
  --patch '{"data":{"circuitBreaker.runtimeEnabled":"false"}}'
```

#### **API Control** (Future Enhancement)
```bash
# Enable via API
curl -X POST http://localhost:8080/api/v1/circuit-breaker/enable

# Disable via API  
curl -X POST http://localhost:8080/api/v1/circuit-breaker/disable

# Get status
curl http://localhost:8080/api/v1/circuit-breaker/status
```

### 📊 **How It Works: Step by Step**

#### **Phase 1: Initialization**
1. **📋 Collect Current Limits**: System reads tenant limits from Mimir config
2. **🧮 Calculate Thresholds**: Apply multipliers and safety margins
3. **⏰ Start Observation**: Begin monitoring traffic patterns
4. **📈 Build Baseline**: Calculate normal operating rates over 24 hours

#### **Phase 2: Real-time Monitoring**
1. **🔍 Monitor Traffic**: Track ingestion, queries, and series rates per tenant
2. **⚖️ Compare Thresholds**: Check if current rates exceed calculated thresholds
3. **🚨 Detect Blasts**: Identify sudden spikes or sustained high traffic
4. **🔄 Adapt Thresholds**: Continuously tune based on observed patterns

#### **Phase 3: Protection Actions**
1. **🟡 Half-Open State**: Reduce limits by 50% and test recovery
2. **🔴 Open State**: Reduce limits by 90% for full protection
3. **🟠 Emergency Mode**: System-wide protection (80% reduction)
4. **🚨 Panic Mode**: Extreme protection (95% reduction)

### 🔧 **Configuration Examples**

> **Note**: All configuration examples below apply only when circuit breaker is enabled (i.e., in production mode or manually enabled in dry-run mode).

#### **Conservative Auto-Configuration (Production)**
```yaml
circuitBreaker:
  enabled: true
  runtimeEnabled: true
  mode: "auto"
  
  autoConfig:
    enabled: true
    limitMultipliers:
      ingestionRateMultiplier: 2.0   # Higher threshold for safety
      queryRateMultiplier: 3.0       # Very conservative query limits
      seriesMultiplier: 2.5          # Higher series threshold
    
    safetyMargins:
      defaultMargin: 40.0            # Large safety buffer
    
    realtimeAdaptation:
      enabled: false                 # Disable adaptation for stability
```

#### **Aggressive Auto-Configuration (Development)**
```yaml
circuitBreaker:
  enabled: true
  runtimeEnabled: true
  mode: "auto"
  
  autoConfig:
    enabled: true
    limitMultipliers:
      ingestionRateMultiplier: 1.2   # Tight thresholds
      queryRateMultiplier: 1.5       # Quick protection
      seriesMultiplier: 1.3          # Fast response
    
    safetyMargins:
      defaultMargin: 15.0            # Small safety buffer
    
    realtimeAdaptation:
      enabled: true                  # Enable learning
      interval: "2m"                 # Fast adaptation
      learningRate: 0.2              # Quick learning
```

#### **Hybrid Configuration (Mixed Environment)**
```yaml
circuitBreaker:
  enabled: true
  runtimeEnabled: true
  mode: "hybrid"
  
  blastProtection:
    useAutoThresholds: true
    tenantOverrides:
      # Critical services: manual high thresholds
      production-api:
        ingestionSpikeThreshold: 10000000  # 10M samples/sec
        querySpikeThreshold: 100000        # 100K queries/sec
      
      # Everything else: auto-configured
```

### 🏁 **Quick Start Guide**

1. **Start in Dry-Run Mode** (circuit breaker automatically disabled):
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=dry-run \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

2. **Monitor for 24-48 hours** to observe actual traffic patterns without interference

3. **Switch to Production Mode** (circuit breaker automatically enabled):
```bash
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=prod \
  --reuse-values
```

4. **Fine-tune safety margins** based on your risk tolerance:
```bash
# Increase safety for critical environments
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set circuitBreaker.autoConfig.safetyMargins.defaultMargin=35.0 \
  --reuse-values
```

5. **Enable real-time adaptation** after confidence is built:
```bash
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set circuitBreaker.autoConfig.realtimeAdaptation.enabled=true \
  --reuse-values
```

### 📈 **Monitoring Auto-Configuration**

#### **Dry-Run Mode Monitoring**
```bash
# Verify circuit breaker is disabled
kubectl logs deployment/mimir-limit-optimizer | grep "circuit breaker.*disabled"

# Monitor traffic patterns (no protection applied)
kubectl logs deployment/mimir-limit-optimizer | grep "tenant metrics collected"

# Check limit calculations (dry-run only)
kubectl logs deployment/mimir-limit-optimizer | grep "would update limits"
```

#### **Production Mode Monitoring**
```bash
# View auto-calculated thresholds
kubectl logs deployment/mimir-limit-optimizer | grep "recalculated thresholds"

# Monitor adaptation changes  
kubectl logs deployment/mimir-limit-optimizer | grep "adapted thresholds"

# Check circuit breaker status and state changes
kubectl logs deployment/mimir-limit-optimizer | grep "circuit breaker"

# Monitor protection actions
kubectl logs deployment/mimir-limit-optimizer | grep "blast.*detected"
```

The automated circuit breaker makes your Mimir deployment **self-protecting** and **self-tuning**, eliminating the guesswork of manual threshold configuration! 🚀

## 📊 Monitoring and Observability

### Prometheus Metrics

The controller exposes comprehensive metrics on `:8080/metrics`:

```bash
# Controller metrics
mimir_limit_optimizer_reconcile_total
mimir_limit_optimizer_reconcile_duration_seconds
mimir_limit_optimizer_last_reconcile_timestamp

# Tenant metrics
mimir_limit_optimizer_tenants_monitored_total
mimir_limit_optimizer_tenant_current_limits
mimir_limit_optimizer_tenant_recommended_limits

# Spike detection
mimir_limit_optimizer_spikes_detected_total
mimir_limit_optimizer_spike_multiplier

# Health metrics
mimir_limit_optimizer_health_status
mimir_limit_optimizer_errors_total
```

### Grafana Dashboard

Import the provided Grafana dashboard for comprehensive monitoring:

```bash
# Dashboard JSON available in ./monitoring/grafana-dashboard.json
kubectl create configmap mimir-limit-optimizer-dashboard \
  --from-file=dashboard.json=monitoring/grafana-dashboard.json \
  -n monitoring
```

### Health Checks

```bash
# Health check endpoint
curl http://localhost:8081/healthz

# Readiness check
curl http://localhost:8081/readyz

# Metrics endpoint
curl http://localhost:8080/metrics
```

## 🔍 Usage Examples

### Dry-Run Mode

Monitor what changes would be made without applying them:

```bash
# Deploy in dry-run mode
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.mode=dry-run

# Check logs for recommendations
kubectl logs deployment/mimir-limit-optimizer | grep "would update"
```

### Production Deployment

```bash
# Production configuration
cat << EOF > production-values.yaml
controller:
  mode: prod
  bufferPercentage: 25
  updateInterval: "10m"

mimir:
  namespace: mimir
  configMapName: mimir-runtime-overrides
  triggerRollout: false  # Runtime overrides are applied without restarts

tenantScoping:
  skipList:
    - "internal-*"
    - "test-*"
    - "dev-*"

eventSpike:
  enabled: true
  threshold: 1.8
  cooldownPeriod: "45m"

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

replicaCount: 2
EOF

helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  -f production-values.yaml \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### Tenant-Specific Configuration

```yaml
# Configure different tiers for different tenants
limits:
  tenantTiers:
    enterprise:
      bufferPercentage: 40
      limits:
        ingestion_rate: 1000000
        max_series: 10000000
    standard:
      bufferPercentage: 25
      limits:
        ingestion_rate: 100000
        max_series: 1000000
    basic:
      bufferPercentage: 15
      limits:
        ingestion_rate: 10000
        max_series: 100000

# Apply tenant to tier mapping
tenantTierMapping:
  "enterprise-*": "enterprise"
  "premium-*": "enterprise"
  "standard-*": "standard"
  "*": "basic"
```

## 🛠️ Development

### Local Development

```bash
# Prerequisites
go version  # 1.21+
kubectl version
helm version

# Run locally
export KUBECONFIG=~/.kube/config
go run main.go --config=config.yaml --log-level=debug

# Run tests
go test ./...

# Build
go build -o bin/mimir-limit-optimizer main.go
```

### Testing

```bash
# Unit tests
go test -v ./internal/...

# Integration tests (requires kind cluster)
make test-integration

# Load tests
make test-load
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Controller Not Starting

```bash
# Check logs
kubectl logs deployment/mimir-limit-optimizer -n mimir-limit-optimizer

# Common causes:
# - RBAC permissions
# - ConfigMap not found
# - Invalid configuration
```

#### 2. Metrics Collection Failing

```bash
# Check service discovery
kubectl get services -l app.kubernetes.io/name=mimir -n mimir-system

# Verify endpoints
kubectl get endpoints -n mimir-system

# Test connectivity
kubectl exec deployment/mimir-limit-optimizer -- \
  wget -qO- http://mimir-distributor.mimir-system:8080/metrics
```

#### 3. ConfigMap Updates Not Applied

```bash
# Check RBAC permissions
kubectl auth can-i update configmaps --as=system:serviceaccount:mimir-limit-optimizer:mimir-limit-optimizer

# Verify ConfigMap exists
kubectl get configmap mimir-runtime-overrides -n mimir-system

# Check controller logs
kubectl logs deployment/mimir-limit-optimizer | grep configmap
```

### Debug Mode

```bash
# Enable debug logging
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set controller.logLevel=debug \
  --reuse-values

# Enable synthetic mode for testing
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set synthetic.enabled=true \
  --set synthetic.tenantCount=10 \
  --reuse-values
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Prometheus    │    │      Mimir       │    │   Controller    │
│   /metrics      │◄───┤   Components     │◄───┤   Manager       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                       ┌─────────────────────────────────┼─────────────────────────────────┐
                       │                                 │                                 │
              ┌─────────▼──────────┐         ┌──────────▼──────────┐         ┌──────────▼──────────┐
              │     Collector      │         │      Analyzer       │         │      Patcher        │
              │  - Service Discovery│         │  - Trend Analysis   │         │  - ConfigMap Updates│
              │  - Metrics Collection│        │  - Spike Detection  │         │  - Backup/Rollback  │
              │  - Auto-discovery   │         │  - Limit Calculation│         │  - Validation       │
              └─────────────────────┘         └─────────────────────┘         └─────────────────────┘
                       │                                 │                                 │
                       └─────────────────────────────────┼─────────────────────────────────┘
                                                         │
                       ┌─────────────────────────────────▼─────────────────────────────────┐
                       │                        Audit Logger                                │
                       │                  - Change Tracking                                 │
                       │                  - Rollback Support                               │
                       │                  - Multiple Storage Backends                      │
                       └─────────────────────────────────────────────────────────────────────┘
```

## 🎉 **Why Choose Mimir Limit Optimizer?**

### **🚀 Enterprise-Grade Dynamic Limits**
- **30+ Mimir Limits Supported** - Far beyond the 4 limits of traditional approaches
- **Zero Code Changes** - Add new limits through configuration only
- **Type-Safe Operations** - Intelligent handling of rates, counts, sizes, and durations
- **Per-Limit Customization** - Individual min/max/buffer settings for each limit
- **Environment Adaptive** - Different configurations for dev/staging/production

### **💰 Intelligent Cost Control**
- **Real-time Cost Monitoring** - Track costs across samples, series, and queries
- **Flexible Budget Enforcement** - Choose between monitoring-only or automatic enforcement
- **Predictive Analytics** - Forecast costs and prevent overruns
- **Multi-level Budgets** - Global, per-tenant, and per-team budget controls

### **🛡️ Advanced Blast Protection**
- **Intelligent Circuit Breaker** - Auto-configured based on actual tenant limits
- **Multi-layer Protection** - Rate limiting, throttling, and emergency controls
- **Panic Mode Recovery** - Automatic system recovery from overload conditions
- **Real-time Adaptation** - Thresholds adapt to traffic patterns

### **📊 Enterprise Observability**
- **Comprehensive Metrics** - 50+ Prometheus metrics for deep insights
- **Audit Logging** - Complete change tracking for compliance
- **Multi-channel Alerting** - Slack, PagerDuty, email, webhooks
- **Health Monitoring** - Component status and performance tracking

### **⚡ Production-Ready Performance**
- **High Throughput** - Optimized for large-scale Mimir deployments
- **Resource Efficient** - Intelligent caching and batch processing
- **Fault Tolerant** - Resilient design with graceful degradation
- **Hot Configuration** - Runtime configuration changes without restarts

### **🔄 Seamless Integration**
- **Kubernetes Native** - Built for cloud-native environments
- **Helm Chart Included** - One-command deployment with best practices
- **Backward Compatible** - Existing configurations continue working
- **Service Discovery** - Automatic Mimir service detection and metrics collection

**Transform your Mimir deployment from a basic metrics system into an enterprise-grade, self-protecting, cost-aware observability platform.** 🚀

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/tapasyadubey/mimir-limit-optimizer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/tapasyadubey/mimir-limit-optimizer/discussions)
- **Documentation**: [Wiki](https://github.com/tapasyadubey/mimir-limit-optimizer/wiki)

## 🙏 Acknowledgments

- [Grafana Mimir](https://grafana.com/oss/mimir/) team for the excellent TSDB
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) for the Kubernetes controller framework
- [Prometheus](https://prometheus.io/) for metrics and monitoring # Rebuild test at Wed Jul  2 19:50:19 IST 2025
