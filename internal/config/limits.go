package config

// GetDefaultLimitDefinitions returns comprehensive configurations for all major Mimir runtime overrides
func GetDefaultLimitDefinitions() map[string]LimitDefinition {
	return map[string]LimitDefinition{
		
		// ===========================================
		// INGESTION LIMITS
		// ===========================================
		
		"ingestion_rate": {
			Name:         "ingestion_rate",
			Type:         "count",
			MetricSource: "cortex_distributor_received_samples_total",
			DefaultValue: int64(25000),
			MinValue:     int64(1000),
			MaxValue:     int64(10000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Rate limit for sample ingestion per tenant (samples/sec)",
		},
		"ingestion_burst_size": {
			Name:         "ingestion_burst_size",
			Type:         "count",
			MetricSource: "cortex_distributor_received_samples_total",
			DefaultValue: int64(50000),
			MinValue:     int64(2000),
			MaxValue:     int64(20000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Burst size for sample ingestion per tenant",
		},
		"ingestion_rate_strategy": {
			Name:         "ingestion_rate_strategy",
			Type:         "string",
			MetricSource: "",
			DefaultValue: "global",
			MinValue:     "",
			MaxValue:     "",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Strategy for ingestion rate limiting (local/global)",
		},
		"ingestion_tenant_shard_size": {
			Name:         "ingestion_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for ingestion (0 = no sharding)",
		},
		
		// ===========================================
		// SERIES LIMITS
		// ===========================================
		
		"max_global_series_per_user": {
			Name:         "max_global_series_per_user",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series",
			DefaultValue: int64(150000),
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum active series per tenant across all ingesters",
		},
		"max_global_series_per_metric": {
			Name:         "max_global_series_per_metric",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series_per_metric",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum series per metric name across all ingesters (0 = unlimited)",
		},
		
		// ===========================================
		// QUERY LIMITS
		// ===========================================
		
		"max_samples_per_query": {
			Name:         "max_samples_per_query",
			Type:         "count",
			MetricSource: "cortex_query_frontend_query_range_duration_seconds",
			DefaultValue: int64(50000000),
			MinValue:     int64(1000),
			MaxValue:     int64(1000000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum samples a single query can load",
		},
		"max_series_per_query": {
			Name:         "max_series_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_series_fetched",
			DefaultValue: int64(100000),
			MinValue:     int64(100),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum series a single query can return",
		},
		"max_concurrent_queries": {
			Name:         "max_concurrent_queries",
			Type:         "count",
			MetricSource: "cortex_query_frontend_queries_in_progress",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum concurrent queries per tenant",
		},
		"max_query_length": {
			Name:         "max_query_length",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "8760h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum query time range (0 = unlimited)",
		},
		"max_query_lookback": {
			Name:         "max_query_lookback",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "8760h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum lookback period for queries (0 = unlimited)",
		},
		"max_partial_query_length": {
			Name:         "max_partial_query_length",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "8760h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum partial query time range (0 = unlimited)",
		},
		"max_query_parallelism": {
			Name:         "max_query_parallelism",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(14),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum parallelism for query execution",
		},
		"max_cache_freshness": {
			Name:         "max_cache_freshness",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "1m",
			MinValue:     "0s",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum age for cached query results",
		},
		
		// ===========================================
		// CHUNK/STORAGE LIMITS
		// ===========================================
		
		"max_fetched_chunks_per_query": {
			Name:         "max_fetched_chunks_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_chunks_fetched",
			DefaultValue: int64(2000000),
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunks a single query can fetch",
		},
		"max_fetched_series_per_query": {
			Name:         "max_fetched_series_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_series_fetched",
			DefaultValue: int64(100000),
			MinValue:     int64(100),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum series a single query can fetch",
		},
		"max_fetched_chunk_bytes_per_query": {
			Name:         "max_fetched_chunk_bytes_per_query",
			Type:         "size",
			MetricSource: "cortex_querier_chunks_fetched_bytes",
			DefaultValue: int64(50000000),
			MinValue:     int64(1000000),
			MaxValue:     int64(1000000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunk bytes a single query can fetch",
		},
		"max_estimated_memory_consumption_per_query": {
			Name:         "max_estimated_memory_consumption_per_query",
			Type:         "size",
			MetricSource: "cortex_querier_estimated_memory_consumption_bytes",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum estimated memory consumption per query (0 = unlimited)",
		},
		"max_estimated_fetched_chunks_per_query": {
			Name:         "max_estimated_fetched_chunks_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_estimated_chunks_fetched",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(100000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum estimated chunks per query (0 = unlimited)",
		},
		
		// ===========================================
		// METADATA LIMITS
		// ===========================================
		
		"max_global_metadata_per_user": {
			Name:         "max_global_metadata_per_user",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_metadata",
			DefaultValue: int64(8000),
			MinValue:     int64(100),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per tenant",
		},
		"max_global_metadata_per_metric": {
			Name:         "max_global_metadata_per_metric",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_metadata_per_metric",
			DefaultValue: int64(10),
			MinValue:     int64(1),
			MaxValue:     int64(100),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per metric",
		},
		
		// ===========================================
		// EXEMPLAR LIMITS
		// ===========================================
		
		"max_global_exemplars_per_user": {
			Name:         "max_global_exemplars_per_user",
			Type:         "count",
			MetricSource: "cortex_ingester_tsdb_exemplar_series_with_exemplars_in_storage",
			DefaultValue: int64(100000),
			MinValue:     int64(1000),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum exemplars per tenant",
		},
		"max_exemplars_per_query": {
			Name:         "max_exemplars_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_exemplars_fetched",
			DefaultValue: int64(100000),
			MinValue:     int64(100),
			MaxValue:     int64(1000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum exemplars per query",
		},
		
		// ===========================================
		// REQUEST RATE LIMITS
		// ===========================================
		
		"request_rate": {
			Name:         "request_rate",
			Type:         "rate",
			MetricSource: "cortex_request_duration_seconds",
			DefaultValue: 0.0,
			MinValue:     0.0,
			MaxValue:     10000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Request rate limit per tenant (requests/sec, 0 = unlimited)",
		},
		"request_burst_size": {
			Name:         "request_burst_size",
			Type:         "count",
			MetricSource: "cortex_request_duration_seconds",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Request burst size per tenant (0 = unlimited)",
		},
		
		// ===========================================
		// RULER LIMITS
		// ===========================================
		
		"ruler_max_rules_per_rule_group": {
			Name:         "ruler_max_rules_per_rule_group",
			Type:         "count",
			MetricSource: "cortex_ruler_rule_group_rules",
			DefaultValue: int64(20),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum rules per rule group",
		},
		"ruler_max_rule_groups_per_tenant": {
			Name:         "ruler_max_rule_groups_per_tenant",
			Type:         "count",
			MetricSource: "cortex_ruler_rule_groups_per_user",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum rule groups per tenant",
		},
		"ruler_evaluation_delay_duration": {
			Name:         "ruler_evaluation_delay_duration",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "10m",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Delay between rule evaluation time and rule execution",
		},
		"ruler_tenant_shard_size": {
			Name:         "ruler_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for ruler (0 = no sharding)",
		},
		"ruler_max_rules_per_tenant": {
			Name:         "ruler_max_rules_per_tenant",
			Type:         "count",
			MetricSource: "cortex_ruler_rules_per_user",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(100000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum rules per tenant (0 = unlimited)",
		},
		
		// ===========================================
		// ALERTMANAGER LIMITS
		// ===========================================
		
		"alertmanager_notification_rate_limit": {
			Name:         "alertmanager_notification_rate_limit",
			Type:         "rate",
			MetricSource: "cortex_alertmanager_notifications_total",
			DefaultValue: 0.0,
			MinValue:     0.0,
			MaxValue:     1000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Alertmanager notification rate limit (0 = unlimited)",
		},
		"alertmanager_max_dispatcher_aggregation_groups": {
			Name:         "alertmanager_max_dispatcher_aggregation_groups",
			Type:         "count",
			MetricSource: "cortex_alertmanager_dispatcher_aggregation_groups",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(10000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum dispatcher aggregation groups (0 = unlimited)",
		},
		"alertmanager_max_alerts_count": {
			Name:         "alertmanager_max_alerts_count",
			Type:         "count",
			MetricSource: "cortex_alertmanager_alerts",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum number of alerts (0 = unlimited)",
		},
		"alertmanager_max_alerts_size_bytes": {
			Name:         "alertmanager_max_alerts_size_bytes",
			Type:         "size",
			MetricSource: "cortex_alertmanager_alerts_size_bytes",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(100000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum size of all alerts in bytes (0 = unlimited)",
		},
		"alertmanager_max_config_size_bytes": {
			Name:         "alertmanager_max_config_size_bytes",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(10000000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum Alertmanager configuration size in bytes (0 = unlimited)",
		},
		"alertmanager_max_templates_count": {
			Name:         "alertmanager_max_templates_count",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum number of templates (0 = unlimited)",
		},
		"alertmanager_max_template_size_bytes": {
			Name:         "alertmanager_max_template_size_bytes",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum template size in bytes (0 = unlimited)",
		},
		
		// ===========================================
		// COMPACTOR LIMITS
		// ===========================================
		
		"compactor_blocks_retention_period": {
			Name:         "compactor_blocks_retention_period",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "8760h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Retention period for compacted blocks (0 = unlimited)",
		},
		"compactor_split_and_merge_shards": {
			Name:         "compactor_split_and_merge_shards",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Number of shards for split-and-merge compaction (0 = disabled)",
		},
		"compactor_split_groups": {
			Name:         "compactor_split_groups",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(1),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Number of groups for split compaction",
		},
		"compactor_tenant_shard_size": {
			Name:         "compactor_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for compactor (0 = no sharding)",
		},
		
		// ===========================================
		// STORE GATEWAY LIMITS
		// ===========================================
		
		"store_gateway_tenant_shard_size": {
			Name:         "store_gateway_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for store gateway (0 = no sharding)",
		},
		
		// ===========================================
		// LABEL LIMITS
		// ===========================================
		
		"max_label_names_per_series": {
			Name:         "max_label_names_per_series",
			Type:         "count",
			MetricSource: "cortex_ingester_active_series",
			DefaultValue: int64(30),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum label names per series",
		},
		"max_label_name_length": {
			Name:         "max_label_name_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(1024),
			MinValue:     int64(10),
			MaxValue:     int64(10000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label names in bytes",
		},
		"max_label_value_length": {
			Name:         "max_label_value_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(2048),
			MinValue:     int64(10),
			MaxValue:     int64(100000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label values in bytes",
		},
		"max_metadata_length": {
			Name:         "max_metadata_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(1024),
			MinValue:     int64(10),
			MaxValue:     int64(10000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of metric metadata in bytes",
		},
		
		// ===========================================
		// CARDINALITY ANALYSIS LIMITS
		// ===========================================
		
		"cardinality_analysis_enabled": {
			Name:         "cardinality_analysis_enabled",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Enable cardinality analysis endpoints",
		},
		"label_names_and_values_results_max_size_bytes": {
			Name:         "label_names_and_values_results_max_size_bytes",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(4194304),
			MinValue:     int64(1024),
			MaxValue:     int64(104857600),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum size of label names and values query results",
		},
		"label_values_max_cardinality_label_names_per_request": {
			Name:         "label_values_max_cardinality_label_names_per_request",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum label names per cardinality request",
		},
		
		// ===========================================
		// QUERY FRONTEND LIMITS
		// ===========================================
		
		"max_outstanding_per_tenant": {
			Name:         "max_outstanding_per_tenant",
			Type:         "count",
			MetricSource: "cortex_query_frontend_queue_length",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum outstanding queries per tenant in queue",
		},
		"max_queriers_per_tenant": {
			Name:         "max_queriers_per_tenant",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum queriers per tenant (0 = unlimited)",
		},
		"query_ingesters_within": {
			Name:         "query_ingesters_within",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "13h",
			MinValue:     "0s",
			MaxValue:     "168h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum lookback to query ingesters",
		},
		"split_queries_by_interval": {
			Name:         "split_queries_by_interval",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "24h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Split queries by time interval (0 = disabled)",
		},
		
		// ===========================================
		// ADVANCED INGESTION LIMITS
		// ===========================================
		
		"out_of_order_time_window": {
			Name:         "out_of_order_time_window",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s",
			MinValue:     "0s",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Time window to accept out-of-order samples (0 = disabled)",
		},
		"out_of_order_blocks_external_label_enabled": {
			Name:         "out_of_order_blocks_external_label_enabled",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Enable external labels on out-of-order blocks",
		},
		"separate_metrics_group_label": {
			Name:         "separate_metrics_group_label",
			Type:         "string",
			MetricSource: "",
			DefaultValue: "",
			MinValue:     "",
			MaxValue:     "",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Label to separate metrics into groups",
		},
		"max_chunks_per_query": {
			Name:         "max_chunks_per_query",
			Type:         "count",
			MetricSource: "cortex_querier_chunks_fetched",
			DefaultValue: int64(2000000),
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum chunks per query (deprecated, use max_fetched_chunks_per_query)",
		},
		
		// ===========================================
		// NATIVE HISTOGRAMS
		// ===========================================
		
		"native_histograms_ingestion_enabled": {
			Name:         "native_histograms_ingestion_enabled",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Enable native histogram ingestion",
		},
		"active_series_metrics_enabled": {
			Name:         "active_series_metrics_enabled",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Enable active series metrics",
		},
		"active_series_metrics_idle_timeout": {
			Name:         "active_series_metrics_idle_timeout",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "10m",
			MinValue:     "1m",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Idle timeout for active series metrics",
		},
		
		// ===========================================
		// VALIDATION LIMITS
		// ===========================================
		
		"create_grace_period": {
			Name:         "create_grace_period",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "10m",
			MinValue:     "0s",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Grace period for timestamp validation on sample creation",
		},
		"enforce_metadata_metric_name": {
			Name:         "enforce_metadata_metric_name",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: true,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Enforce metadata metric name validation",
		},
		"ingestion_partition_tenant_shard_size": {
			Name:         "ingestion_partition_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for ingestion partitions (0 = no sharding)",
		},
		
		// ===========================================
		// BYTE-BASED INGESTION LIMITS
		// ===========================================
		
		"max_ingestion_rate_bytes": {
			Name:         "max_ingestion_rate_bytes",
			Type:         "size",
			MetricSource: "cortex_distributor_received_samples_bytes_total",
			DefaultValue: int64(25000000), // 25MB/sec
			MinValue:     int64(1000000),  // 1MB/sec
			MaxValue:     int64(10000000000), // 10GB/sec
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Rate limit for ingestion in bytes per second per tenant",
		},
		"max_ingestion_burst_size_bytes": {
			Name:         "max_ingestion_burst_size_bytes", 
			Type:         "size",
			MetricSource: "cortex_distributor_received_samples_bytes_total",
			DefaultValue: int64(50000000), // 50MB burst
			MinValue:     int64(2000000),  // 2MB
			MaxValue:     int64(20000000000), // 20GB
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Burst size for ingestion in bytes per tenant",
		},
		
		// ===========================================
		// SAMPLE/DATA VALIDATION LIMITS
		// ===========================================
		
		"max_sample_age": {
			Name:         "max_sample_age",
			Type:         "duration",
			MetricSource: "cortex_distributor_latest_seen_sample_timestamp_seconds",
			DefaultValue: "336h", // 14 days
			MinValue:     "1h",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "Maximum age of samples that can be ingested",
		},
		"enforce_metric_name_validation": {
			Name:         "enforce_metric_name_validation",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: true,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether to enforce metric name validation",
		},
		
		// ===========================================
		// CHUNK STORAGE LIMITS
		// ===========================================
		
		"max_chunk_age": {
			Name:         "max_chunk_age",
			Type:         "duration",
			MetricSource: "cortex_ingester_oldest_unshipped_block_timestamp_seconds",
			DefaultValue: "12h",
			MinValue:     "1h", 
			MaxValue:     "72h",
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "Maximum age of chunks before they must be shipped",
		},
		"max_chunk_size_bytes": {
			Name:         "max_chunk_size_bytes",
			Type:         "size",
			MetricSource: "cortex_ingester_chunk_size_bytes",
			DefaultValue: int64(1048576), // 1MB
			MinValue:     int64(1024),    // 1KB
			MaxValue:     int64(104857600), // 100MB
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum size of individual chunks in bytes",
		},
		
		// ===========================================
		// TENANT MANAGEMENT LIMITS
		// ===========================================
		
		"max_tenants": {
			Name:         "max_tenants",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_users",
			DefaultValue: int64(1000),
			MinValue:     int64(1),
			MaxValue:     int64(100000),
			BufferFactor: 10.0,
			Enabled:      true,
			Description:  "Maximum number of tenants per ingester",
		},
		"enforce_tenant_id_header": {
			Name:         "enforce_tenant_id_header",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: true,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether to enforce X-Scope-OrgID header presence",
		},
		"per_tenant_override": {
			Name:         "per_tenant_override",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: true,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether per-tenant overrides are enabled",
		},
		"subtenant_limits": {
			Name:         "subtenant_limits",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether hierarchical sub-tenant limits are enabled",
		},
		
		// ===========================================
		// REMOTE WRITE LIMITS
		// ===========================================
		
		"remote_write_deadline": {
			Name:         "remote_write_deadline",
			Type:         "duration",
			MetricSource: "cortex_distributor_push_duration_seconds",
			DefaultValue: "30s",
			MinValue:     "1s",
			MaxValue:     "300s",
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "Deadline for remote write requests",
		},
		"remote_write_max_samples_per_send": {
			Name:         "remote_write_max_samples_per_send",
			Type:         "count",
			MetricSource: "cortex_distributor_samples_in_total",
			DefaultValue: int64(10000),
			MinValue:     int64(100),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum samples per remote write request",
		},
		
		// ===========================================
		// OBSERVABILITY LIMITS
		// ===========================================
		
		"trace_sampling_rate": {
			Name:         "trace_sampling_rate",
			Type:         "percentage",
			MetricSource: "",
			DefaultValue: 1.0, // 1%
			MinValue:     0.0,
			MaxValue:     100.0,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Sampling rate for distributed tracing (0-100%)",
		},
		"log_level": {
			Name:         "log_level",
			Type:         "string",
			MetricSource: "",
			DefaultValue: "info",
			MinValue:     "",
			MaxValue:     "",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Log level for tenant operations (debug, info, warn, error)",
		},
		
		// ===========================================
		// QUERY TIMEOUT & SCHEDULING LIMITS
		// ===========================================
		
		"query_timeout": {
			Name:         "query_timeout",
			Type:         "duration",
			MetricSource: "cortex_query_frontend_query_duration_seconds",
			DefaultValue: "300s", // 5 minutes
			MinValue:     "1s",
			MaxValue:     "3600s", // 1 hour
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "Maximum query execution timeout",
		},
		"query_scheduler_max_outstanding_requests_per_tenant": {
			Name:         "query_scheduler_max_outstanding_requests_per_tenant",
			Type:         "count",
			MetricSource: "cortex_query_scheduler_queue_length",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum outstanding requests per tenant in query scheduler",
		},
		"query_scheduler_max_queriers_per_tenant": {
			Name:         "query_scheduler_max_queriers_per_tenant",
			Type:         "count",
			MetricSource: "cortex_query_scheduler_queriers_connected",
			DefaultValue: int64(10),
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum queriers per tenant in query scheduler",
		},
		"query_scheduler_max_outstanding_requests": {
			Name:         "query_scheduler_max_outstanding_requests",
			Type:         "count",
			MetricSource: "cortex_query_scheduler_queue_length",
			DefaultValue: int64(1000),
			MinValue:     int64(10),
			MaxValue:     int64(100000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Global maximum outstanding requests in query scheduler",
		},
		"query_scheduler_max_active_requests": {
			Name:         "query_scheduler_max_active_requests",
			Type:         "count",
			MetricSource: "cortex_query_scheduler_queries_in_progress",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum active requests in query scheduler",
		},
		"enable_query_scheduling": {
			Name:         "enable_query_scheduling",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether query scheduling is enabled for tenant",
		},
		
		// ===========================================
		// STORAGE GATEWAY LIMITS
		// ===========================================
		
		"store_gateway_max_queries_in_flight": {
			Name:         "store_gateway_max_queries_in_flight",
			Type:         "count",
			MetricSource: "cortex_bucket_store_queries_in_flight",
			DefaultValue: int64(100),
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum concurrent queries per store gateway",
		},
		"blocks_storage_tenant_shard_size": {
			Name:         "blocks_storage_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0), // 0 = no sharding
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for blocks storage (0 = no sharding)",
		},
		"blocks_storage_per_tenant_override": {
			Name:         "blocks_storage_per_tenant_override",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: false,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether per-tenant blocks storage overrides are enabled",
		},
		
		// ===========================================
		// TSDB SPECIFIC LIMITS
		// ===========================================
		
		"tsdb_retention_period": {
			Name:         "tsdb_retention_period",
			Type:         "duration",
			MetricSource: "prometheus_tsdb_blocks_loaded",
			DefaultValue: "336h", // 14 days
			MinValue:     "24h",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "TSDB block retention period per tenant",
		},
		
		// ===========================================
		// API SPECIFIC LIMITS
		// ===========================================
		
		"api_limit_max_series_per_metric_name": {
			Name:         "api_limit_max_series_per_metric_name",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series_per_metric",
			DefaultValue: int64(50000),
			MinValue:     int64(100),
			MaxValue:     int64(10000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "API limit for maximum series per metric name",
		},
		"api_limit_max_label_value_length": {
			Name:         "api_limit_max_label_value_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(4096), // 4KB
			MinValue:     int64(256),
			MaxValue:     int64(65536), // 64KB
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "API limit for maximum label value length in API responses",
		},
		
		// ===========================================
		// CONCURRENT REQUEST LIMITS
		// ===========================================
		
		"max_concurrent_requests": {
			Name:         "max_concurrent_requests",
			Type:         "count",
			MetricSource: "cortex_request_duration_seconds",
			DefaultValue: int64(1000),
			MinValue:     int64(1),
			MaxValue:     int64(100000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum concurrent requests per tenant across all components",
		},
		
		// ===========================================
		// BYTES-BASED QUERY LIMITS
		// ===========================================
		
		"max_bytes_per_query": {
			Name:         "max_bytes_per_query",
			Type:         "size",
			MetricSource: "cortex_querier_chunks_fetched_bytes",
			DefaultValue: int64(1073741824), // 1GB
			MinValue:     int64(1048576),    // 1MB
			MaxValue:     int64(107374182400), // 100GB
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum bytes a single query can process",
		},
		
		// ===========================================
		// RETENTION & TSDB LIMITS
		// ===========================================
		
		"retention_period": {
			Name:         "retention_period",
			Type:         "duration",
			MetricSource: "prometheus_tsdb_blocks_loaded",
			DefaultValue: "336h", // 14 days
			MinValue:     "24h",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      true,
			Description:  "General retention period per tenant",
		},
		
		// ===========================================
		// CARDINALITY MANAGEMENT LIMITS  
		// ===========================================
		
		"cardinality_limit": {
			Name:         "cardinality_limit",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series",
			DefaultValue: int64(100000),
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Overall cardinality limit per tenant",
		},
		"enforce_metadata_validation": {
			Name:         "enforce_metadata_validation",
			Type:         "bool",
			MetricSource: "",
			DefaultValue: true,
			MinValue:     false,
			MaxValue:     true,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Whether to enforce metadata validation",
		},
		
		// ===========================================
		// DEPRECATED COMPATIBILITY LIMITS
		// ===========================================
		
		"max_metadata_per_user": {
			Name:         "max_metadata_per_user",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_metadata",
			DefaultValue: int64(8000),
			MinValue:     int64(100),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "DEPRECATED: Use max_global_metadata_per_user instead",
		},
		
		// ===========================================
		// DEPRECATED/COMPATIBILITY LIMITS
		// ===========================================
		
		"max_series_per_metric": {
			Name:         "max_series_per_metric",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series_per_metric",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Deprecated: use max_global_series_per_metric instead",
		},
		"max_series_per_user": {
			Name:         "max_series_per_user",
			Type:         "count",
			MetricSource: "cortex_ingester_memory_series",
			DefaultValue: int64(0),
			MinValue:     int64(0),
			MaxValue:     int64(100000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Deprecated: use max_global_series_per_user instead",
		},
	}
} 