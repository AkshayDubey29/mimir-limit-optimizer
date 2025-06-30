# Comprehensive Mimir Limits & Metric Sources Guide

## Overview

The Mimir Limit Optimizer now supports **65+ comprehensive runtime overrides** covering all major categories of Mimir configuration limits. This represents a complete coverage of Mimir's tenant-configurable limits with proper metric sources for dynamic optimization.

## 🎯 Major Improvements

### ✅ **Complete Coverage Categories:**
- **Ingestion Limits** (4 limits) - Core ingestion control
- **Series Limits** (2 limits) - Cardinality management  
- **Query Limits** (8 limits) - Query performance & concurrency
- **Chunk/Storage Limits** (5 limits) - Storage optimization
- **Metadata Limits** (2 limits) - Metadata management
- **Exemplar Limits** (2 limits) - Exemplar management
- **Request Rate Limits** (2 limits) - Request throttling
- **Ruler Limits** (5 limits) - Alerting rules management
- **Alertmanager Limits** (7 limits) - Alert management
- **Compactor Limits** (4 limits) - Block compaction
- **Store Gateway Limits** (1 limit) - Store sharding
- **Label Limits** (4 limits) - Label validation
- **Cardinality Analysis** (3 limits) - Cardinality monitoring
- **Query Frontend** (4 limits) - Query queue management
- **Advanced Ingestion** (4 limits) - Out-of-order & advanced features
- **Native Histograms** (3 limits) - Modern Prometheus features
- **Validation Limits** (3 limits) - Input validation
- **Deprecated/Compatibility** (2 limits) - Backwards compatibility

### 🔧 **Corrected Metric Sources:**
- **Before:** Prometheus-style generic metrics (`prometheus_tsdb_head_series`)
- **After:** Mimir-specific tenant-aware metrics (`cortex_ingester_memory_series`)
- **Multi-tenant Support:** All metrics now use `cortex_*`/`mimir_*` prefixes with tenant labels

### 📊 **Enhanced Type Support:**
- **count** - Integer limits (series, samples, rules)
- **rate** - Float limits (samples/sec, requests/sec)  
- **size** - Byte limits (memory, storage, configs)
- **duration** - Time limits ("5m", "1h", "0s")
- **bool** - Feature toggles (enable/disable)
- **string** - Configuration values (strategies, labels)
- **percentage** - Percentage limits (0.0-100.0)

## 📋 Detailed Limit Categories

### 🔄 **Ingestion Limits**
Controls how data enters Mimir:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `ingestion_rate` | count | `cortex_distributor_received_samples_total` | Samples/sec per tenant |
| `ingestion_burst_size` | count | `cortex_distributor_received_samples_total` | Burst allowance |
| `ingestion_rate_strategy` | string | - | Local vs global strategy |
| `ingestion_tenant_shard_size` | count | - | Tenant sharding |

### 📈 **Series Limits**
Controls cardinality and memory usage:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_global_series_per_user` | count | `cortex_ingester_memory_series` | Total series per tenant |
| `max_global_series_per_metric` | count | `cortex_ingester_memory_series_per_metric` | Series per metric name |

### 🔍 **Query Limits** 
Controls query performance and resource usage:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_samples_per_query` | count | `cortex_query_frontend_query_range_duration_seconds` | Sample limit per query |
| `max_series_per_query` | count | `cortex_querier_series_fetched` | Series limit per query |
| `max_concurrent_queries` | count | `cortex_query_frontend_queries_in_progress` | Concurrent query limit |
| `max_query_length` | duration | - | Time range limit |
| `max_query_lookback` | duration | - | Lookback limit |
| `max_partial_query_length` | duration | - | Partial query limit |
| `max_query_parallelism` | count | - | Query parallelism |
| `max_cache_freshness` | duration | - | Cache TTL |

### 💾 **Chunk/Storage Limits**
Controls storage access patterns:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_fetched_chunks_per_query` | count | `cortex_querier_chunks_fetched` | Chunk fetch limit |
| `max_fetched_series_per_query` | count | `cortex_querier_series_fetched` | Series fetch limit |
| `max_fetched_chunk_bytes_per_query` | size | `cortex_querier_chunks_fetched_bytes` | Byte fetch limit |
| `max_estimated_memory_consumption_per_query` | size | `cortex_querier_estimated_memory_consumption_bytes` | Memory estimate limit |
| `max_estimated_fetched_chunks_per_query` | count | `cortex_querier_estimated_chunks_fetched` | Chunk estimate limit |

### 📊 **Metadata Limits**
Controls metric metadata:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_global_metadata_per_user` | count | `cortex_ingester_memory_metadata` | Total metadata per tenant |
| `max_global_metadata_per_metric` | count | `cortex_ingester_memory_metadata_per_metric` | Metadata per metric |

### 🔍 **Exemplar Limits**
Controls exemplar storage:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_global_exemplars_per_user` | count | `cortex_ingester_tsdb_exemplar_series_with_exemplars_in_storage` | Total exemplars per tenant |
| `max_exemplars_per_query` | count | `cortex_querier_exemplars_fetched` | Exemplars per query |

### 🚦 **Request Rate Limits**
Controls request throttling:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `request_rate` | rate | `cortex_request_duration_seconds` | Requests/sec per tenant |
| `request_burst_size` | count | `cortex_request_duration_seconds` | Request burst allowance |

### 📏 **Ruler Limits**
Controls alerting rules:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `ruler_max_rules_per_rule_group` | count | `cortex_ruler_rule_group_rules` | Rules per group |
| `ruler_max_rule_groups_per_tenant` | count | `cortex_ruler_rule_groups_per_user` | Groups per tenant |
| `ruler_evaluation_delay_duration` | duration | - | Evaluation delay |
| `ruler_tenant_shard_size` | count | - | Ruler sharding |
| `ruler_max_rules_per_tenant` | count | `cortex_ruler_rules_per_user` | Total rules per tenant |

### 🚨 **Alertmanager Limits**
Controls alert management:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `alertmanager_notification_rate_limit` | rate | `cortex_alertmanager_notifications_total` | Notification rate |
| `alertmanager_max_dispatcher_aggregation_groups` | count | `cortex_alertmanager_dispatcher_aggregation_groups` | Aggregation groups |
| `alertmanager_max_alerts_count` | count | `cortex_alertmanager_alerts` | Total alerts |
| `alertmanager_max_alerts_size_bytes` | size | `cortex_alertmanager_alerts_size_bytes` | Alert size |
| `alertmanager_max_config_size_bytes` | size | - | Config size |
| `alertmanager_max_templates_count` | count | - | Template count |
| `alertmanager_max_template_size_bytes` | size | - | Template size |

### 🗜️ **Compactor Limits**
Controls block compaction:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `compactor_blocks_retention_period` | duration | - | Block retention |
| `compactor_split_and_merge_shards` | count | - | Split-merge shards |
| `compactor_split_groups` | count | - | Split groups |
| `compactor_tenant_shard_size` | count | - | Compactor sharding |

### 🏪 **Store Gateway Limits**
Controls store gateway behavior:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `store_gateway_tenant_shard_size` | count | - | Store sharding |

### 🏷️ **Label Limits**
Controls label validation:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_label_names_per_series` | count | `cortex_ingester_active_series` | Labels per series |
| `max_label_name_length` | size | - | Label name length |
| `max_label_value_length` | size | - | Label value length |
| `max_metadata_length` | size | - | Metadata length |

### 📊 **Cardinality Analysis Limits**
Controls cardinality monitoring:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `cardinality_analysis_enabled` | bool | - | Enable cardinality endpoints |
| `label_names_and_values_results_max_size_bytes` | size | - | Result size limit |
| `label_values_max_cardinality_label_names_per_request` | count | - | Labels per request |

### 🎛️ **Query Frontend Limits**
Controls query queue behavior:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_outstanding_per_tenant` | count | `cortex_query_frontend_queue_length` | Queue length |
| `max_queriers_per_tenant` | count | - | Queriers per tenant |
| `query_ingesters_within` | duration | - | Ingester lookback |
| `split_queries_by_interval` | duration | - | Query splitting |

### 🔄 **Advanced Ingestion Limits**
Controls advanced ingestion features:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `out_of_order_time_window` | duration | - | Out-of-order window |
| `out_of_order_blocks_external_label_enabled` | bool | - | External labels |
| `separate_metrics_group_label` | string | - | Metrics grouping |
| `max_chunks_per_query` | count | `cortex_querier_chunks_fetched` | Chunks per query (deprecated) |

### 📊 **Native Histograms**
Controls modern Prometheus features:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `native_histograms_ingestion_enabled` | bool | - | Native histogram support |
| `active_series_metrics_enabled` | bool | - | Active series tracking |
| `active_series_metrics_idle_timeout` | duration | - | Active series timeout |

### ✅ **Validation Limits**
Controls input validation:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `create_grace_period` | duration | - | Timestamp validation |
| `enforce_metadata_metric_name` | bool | - | Metadata validation |
| `ingestion_partition_tenant_shard_size` | count | - | Partition sharding |

### 🔄 **Deprecated/Compatibility Limits**
For backwards compatibility:

| Limit | Type | Metric Source | Purpose |
|-------|------|---------------|---------|
| `max_series_per_metric` | count | `cortex_ingester_memory_series_per_metric` | Deprecated alias |
| `max_series_per_user` | count | `cortex_ingester_memory_series` | Deprecated alias |

## 🎯 Key Advantages

### 1. **Complete Coverage**
- **65+ limits** covering all Mimir runtime overrides
- **No gaps** in critical limit categories
- **Future-proof** with native histogram and modern features

### 2. **Accurate Metric Sources**
- **Tenant-aware metrics** with proper `user` labels
- **Mimir-specific metrics** (`cortex_*` prefixes)
- **Multi-tenant compatible** metric sources

### 3. **Enhanced Type System**
- **7 data types** supported (count, rate, size, duration, bool, string, percentage)
- **Automatic type conversion** in patcher
- **Proper zero-value handling** for all types

### 4. **Smart Defaults**
- **Production-ready defaults** based on Grafana Labs recommendations
- **Conservative limits** to prevent system overload
- **Sensible min/max boundaries** for all limits

### 5. **Operational Excellence**
- **Categorized organization** for easy management
- **Clear descriptions** for each limit
- **Deprecation handling** for backwards compatibility

## 🚀 Usage Examples

### **Enable Core Limits** (Recommended Start)
```yaml
dynamicLimits:
  enabled: true
  limitDefinitions:
    ingestion_rate:
      enabled: true
    ingestion_burst_size:
      enabled: true
    max_global_series_per_user:
      enabled: true
    max_samples_per_query:
      enabled: true
    max_concurrent_queries:
      enabled: true
```

### **Enable Advanced Query Optimization**
```yaml
dynamicLimits:
  enabled: true
  limitDefinitions:
    max_fetched_chunks_per_query:
      enabled: true
    max_fetched_series_per_query:
      enabled: true
    max_fetched_chunk_bytes_per_query:
      enabled: true
    query_ingesters_within:
      enabled: true
    split_queries_by_interval:
      enabled: true
```

### **Enable Cardinality Management**
```yaml
dynamicLimits:
  enabled: true
  limitDefinitions:
    max_global_series_per_user:
      enabled: true
    max_global_series_per_metric:
      enabled: true
    cardinality_analysis_enabled:
      enabled: true
    max_label_names_per_series:
      enabled: true
```

### **Enable Modern Features**
```yaml
dynamicLimits:
  enabled: true
  limitDefinitions:
    native_histograms_ingestion_enabled:
      enabled: true
    active_series_metrics_enabled:
      enabled: true
    out_of_order_time_window:
      enabled: true
```

## 🔧 Migration Guide

### **From v2.0.x to v2.1.0:**

1. **Update Configuration:**
   - Review new limit categories
   - Enable relevant limits for your environment
   - Check metric source compatibility

2. **Validate Metric Sources:**
   - Ensure your Mimir deployment exposes `cortex_*` metrics
   - Verify tenant labels are properly set
   - Test metric queries against your Mimir instance

3. **Enable Gradually:**
   - Start with core limits (ingestion, series, queries)
   - Add advanced limits after testing
   - Monitor for any unexpected behavior

4. **Check Deprecated Limits:**
   - `max_series_per_metric` → `max_global_series_per_metric`
   - `max_series_per_user` → `max_global_series_per_user`

This comprehensive revision ensures that the Mimir Limit Optimizer now covers virtually all tenant-configurable limits in Mimir with proper metric sources and intelligent dynamic optimization capabilities! 