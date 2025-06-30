# Dynamic Limits System

The Mimir Limit Optimizer now supports **dynamic limit management** for all Mimir configuration parameters, replacing the previous hardcoded approach that only supported 4 limits.

## Overview

### Previous Limitation
The old system only supported **4 hardcoded limits**:
- `ingestion_rate` 
- `ingestion_burst_size`
- `max_global_series_per_user`
- `max_samples_per_query`

### New Dynamic System
The new system supports **30+ Mimir limits** dynamically with full configurability:

#### Ingestion Limits
- `ingestion_rate` - Sample ingestion rate limit (samples/sec)
- `ingestion_burst_size` - Burst size for sample ingestion
- `max_global_series_per_user` - Maximum series per tenant
- `max_global_series_per_metric` - Maximum series per metric name

#### Query Limits  
- `max_samples_per_query` - Maximum samples a query can load
- `max_series_per_query` - Maximum series a query can return
- `max_query_lookback` - Maximum lookback period for queries
- `max_query_length` - Maximum query range length

#### Chunk & Fetch Limits
- `max_fetched_chunks_per_query` - Maximum chunks a query can fetch
- `max_fetched_series_per_query` - Maximum series a query can fetch  
- `max_fetched_chunk_bytes_per_query` - Maximum chunk bytes per query

#### Metadata & Exemplar Limits
- `max_global_metadata_per_user` - Maximum metadata entries per tenant
- `max_global_metadata_per_metric` - Maximum metadata entries per metric
- `max_global_exemplars_per_user` - Maximum exemplars per tenant

#### Request Rate Limits
- `request_rate` - Request rate limit (requests/sec)
- `request_burst_size` - Request burst size

#### Ruler Limits
- `ruler_max_rules_per_rule_group` - Maximum rules per rule group
- `ruler_max_rule_groups_per_tenant` - Maximum rule groups per tenant

#### Alertmanager Limits
- `alertmanager_notification_rate_limit` - Notification rate limit
- `alertmanager_max_dispatcher_aggregation_groups` - Max aggregation groups
- `alertmanager_max_alerts_count` - Maximum number of alerts
- `alertmanager_max_alerts_size_bytes` - Maximum alerts size in bytes

#### Storage & Compactor Limits
- `compactor_blocks_retention_period` - Retention period for blocks
- `store_gateway_tenant_shard_size` - Tenant shard size for store gateway

#### Label Limits
- `max_label_names_per_series` - Maximum label names per series
- `max_label_name_length` - Maximum length of label names
- `max_label_value_length` - Maximum length of label values

#### Specialized Limits
- `out_of_order_time_window` - Time window for out-of-order samples

## Configuration

### Enable Dynamic Limits

In your `values.yaml`:

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
    # Add more limits as needed
```

### Limit Types

Each limit has a **type** that determines how it's processed:

- **`rate`** - Rates like samples/sec, requests/sec
- **`count`** - Counts like number of series, rules
- **`size`** - Sizes in bytes 
- **`duration`** - Time durations like "1h", "24h"
- **`percentage`** - Percentage values 0-100

### Custom Limit Configuration

Override default values for specific limits:

```yaml
dynamicLimits:
  limitOverrides:
    ingestion_rate:
      defaultValue: 50000.0
      minValue: 5000.0
      maxValue: 5000000.0
      bufferFactor: 30.0
    max_global_series_per_user:
      defaultValue: 200000.0
      minValue: 10000.0
      maxValue: 50000000.0
      bufferFactor: 25.0
```

## Metric-to-Limit Mapping

The system automatically maps Prometheus metrics to Mimir limits:

| Metric | Mimir Limit |
|--------|-------------|
| `prometheus_remote_storage_samples_in_total` | `ingestion_rate`, `ingestion_burst_size` |
| `prometheus_tsdb_head_series` | `max_global_series_per_user` |
| `prometheus_engine_query_samples_total` | `max_samples_per_query` |
| `prometheus_engine_query_series_total` | `max_series_per_query` |
| `prometheus_tsdb_head_chunks` | `max_fetched_chunks_per_query` |
| `prometheus_tsdb_compaction_chunk_size_bytes` | `max_fetched_chunk_bytes_per_query` |
| `prometheus_tsdb_exemplar_exemplars_total` | `max_global_exemplars_per_user` |
| `prometheus_rule_group_rules` | `ruler_max_rules_per_rule_group` |
| `alertmanager_notifications_total` | `alertmanager_notification_rate_limit` |
| `alertmanager_alerts` | `alertmanager_max_alerts_count` |
| `http_requests_total` | `request_rate` |

## Migration Guide

### From Hardcoded to Dynamic

**Before** (Hardcoded):
```yaml
# Only 4 limits supported
limits:
  ingestion_rate: 25000
  max_series: 150000
```

**After** (Dynamic):
```yaml
dynamicLimits:
  enabled: true
  enabledLimits:
    - ingestion_rate
    - ingestion_burst_size  
    - max_global_series_per_user
    - max_samples_per_query
    - max_fetched_chunks_per_query
    # Enable any additional limits as needed
```

### Backward Compatibility

The system is **backward compatible**. Existing configurations will:
1. Continue to work with hardcoded limits
2. Automatically migrate to dynamic system when enabled
3. Use sensible defaults for all unconfigured limits

## Advanced Usage

### Environment-Specific Configurations

**Development**:
```yaml
dynamicLimits:
  enabledLimits:
    - ingestion_rate
    - max_global_series_per_user
  limitOverrides:
    ingestion_rate:
      defaultValue: 10000.0
      bufferFactor: 50.0  # Higher buffer for dev
```

**Production**:
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
  limitOverrides:
    ingestion_rate:
      defaultValue: 100000.0
      bufferFactor: 20.0  # Tighter buffer for prod
```

### Custom Limit Definitions

Add completely custom limits:

```yaml
dynamicLimits:
  customLimits:
    custom_tenant_specific_limit:
      type: "count"
      metricSource: "custom_metric_total"
      defaultValue: 1000.0
      minValue: 100.0
      maxValue: 10000.0
      bufferFactor: 25.0
      enabled: true
      description: "Custom tenant-specific limit"
```

## Monitoring & Observability

### Metrics

The dynamic limits system exposes metrics for all managed limits:

```promql
# Track limit utilization per tenant per limit type
mimir_optimizer_limit_utilization{tenant="tenant1", limit="ingestion_rate"}

# Monitor limit updates
mimir_optimizer_limits_updated_total{tenant="tenant1", limit="max_series", reason="trend-analysis"}

# Track enabled limits count
mimir_optimizer_enabled_limits_total{limit_type="rate"}
```

### Audit Logging

All limit changes are logged with full context:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "tenant": "tenant1", 
  "action": "update-limits",
  "reason": "trend-analysis",
  "changes": {
    "ingestion_rate": 45000.0,
    "max_global_series_per_user": 180000.0,
    "max_samples_per_query": 60000000.0,
    "max_fetched_chunks_per_query": 2400000.0
  }
}
```

## Benefits

### Flexibility
- **Support for 30+ limits** vs. previous 4
- **Type-safe** limit handling
- **Custom limit definitions** support

### Maintainability  
- **No more hardcoded limits**
- **Configuration-driven** approach
- **Easy to add new limits** without code changes

### Operational Excellence
- **Granular control** over which limits to optimize
- **Per-limit configuration** (min/max/buffer)
- **Environment-specific** configurations
- **Comprehensive monitoring** and auditing

### Enterprise-Ready
- **Production-grade** with full observability
- **Backward compatible** migration path
- **Extensive documentation** and examples
- **Type safety** with validation

## Examples

### Enable All Major Limits
```yaml
dynamicLimits:
  enabled: true
  enabledLimits:
    # Ingestion
    - ingestion_rate
    - ingestion_burst_size
    - max_global_series_per_user
    
    # Query Performance  
    - max_samples_per_query
    - max_series_per_query
    - max_fetched_chunks_per_query
    - max_fetched_series_per_query
    - max_fetched_chunk_bytes_per_query
    
    # Metadata & Exemplars
    - max_global_exemplars_per_user
    
    # Request Rate
    - request_rate
    - request_burst_size
```

### High-Cardinality Environment
```yaml
dynamicLimits:
  limitOverrides:
    max_global_series_per_user:
      defaultValue: 1000000.0  # 1M series
      maxValue: 50000000.0     # 50M max
      bufferFactor: 15.0       # Lower buffer
    max_label_names_per_series:
      defaultValue: 50.0       # More labels allowed
      maxValue: 100.0
```

### Cost-Optimized Setup
```yaml
dynamicLimits:
  limitOverrides:
    ingestion_rate:
      bufferFactor: 10.0       # Tighter limits
    max_samples_per_query:
      defaultValue: 10000000.0 # Lower query limits
      bufferFactor: 25.0
```

This dynamic system provides **enterprise-grade flexibility** while maintaining **simplicity** for basic use cases. 