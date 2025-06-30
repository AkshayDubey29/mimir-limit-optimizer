package config

// GetDefaultLimitDefinitions returns predefined configurations for all major Mimir limits
func GetDefaultLimitDefinitions() map[string]LimitDefinition {
	return map[string]LimitDefinition{
		// Ingestion Limits
		"ingestion_rate": {
			Name:         "ingestion_rate",
			Type:         "count", // Changed to count (integer) - samples per second should be whole numbers
			MetricSource: "prometheus_remote_storage_samples_in_total",
			DefaultValue: int64(25000), // Integer value
			MinValue:     int64(1000),
			MaxValue:     int64(10000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Rate limit for sample ingestion (samples/sec)",
		},
		"ingestion_burst_size": {
			Name:         "ingestion_burst_size",
			Type:         "count",
			MetricSource: "prometheus_remote_storage_samples_in_total",
			DefaultValue: int64(50000), // Integer value
			MinValue:     int64(2000),
			MaxValue:     int64(20000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Burst size for sample ingestion",
		},
		
		// Series Limits
		"max_global_series_per_user": {
			Name:         "max_global_series_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: int64(150000), // Integer value
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum series per tenant",
		},
		"max_global_series_per_metric": {
			Name:         "max_global_series_per_metric",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: int64(0), // 0 = unlimited
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum series per metric name",
		},
		
		// Query Limits
		"max_samples_per_query": {
			Name:         "max_samples_per_query",
			Type:         "count",
			MetricSource: "prometheus_engine_query_samples_total",
			DefaultValue: int64(50000000), // Integer value
			MinValue:     int64(1000),
			MaxValue:     int64(1000000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum samples a single query can load",
		},
		"max_series_per_query": {
			Name:         "max_series_per_query",
			Type:         "count",
			MetricSource: "prometheus_engine_query_series_total",
			DefaultValue: int64(100000), // Integer value
			MinValue:     int64(100),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum series a single query can return",
		},
		"max_query_lookback": {
			Name:         "max_query_lookback",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "24h",
			MinValue:     "1h",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum lookback period for queries",
		},
		"max_query_length": {
			Name:         "max_query_length",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "24h",
			MinValue:     "1h",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum query range length",
		},
		
		// Chunk Limits
		"max_fetched_chunks_per_query": {
			Name:         "max_fetched_chunks_per_query",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_chunks",
			DefaultValue: int64(2000000), // Integer value
			MinValue:     int64(1000),
			MaxValue:     int64(100000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunks a single query can fetch",
		},
		"max_fetched_series_per_query": {
			Name:         "max_fetched_series_per_query", 
			Type:         "count",
			MetricSource: "prometheus_engine_query_series_total",
			DefaultValue: int64(100000), // Integer value
			MinValue:     int64(100),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum series a single query can fetch",
		},
		"max_fetched_chunk_bytes_per_query": {
			Name:         "max_fetched_chunk_bytes_per_query",
			Type:         "size",
			MetricSource: "prometheus_tsdb_compaction_chunk_size_bytes",
			DefaultValue: int64(50000000), // 50MB - Integer value
			MinValue:     int64(1000000),  // 1MB
			MaxValue:     int64(1000000000), // 1GB
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunk bytes a single query can fetch",
		},
		
		// Metadata Limits
		"max_global_metadata_per_user": {
			Name:         "max_global_metadata_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: int64(8000), // Integer value
			MinValue:     int64(100),
			MaxValue:     int64(1000000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per tenant",
		},
		"max_global_metadata_per_metric": {
			Name:         "max_global_metadata_per_metric",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: int64(10), // Integer value
			MinValue:     int64(1),
			MaxValue:     int64(100),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per metric",
		},
		
		// Exemplar Limits
		"max_global_exemplars_per_user": {
			Name:         "max_global_exemplars_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_exemplar_exemplars_total",
			DefaultValue: int64(100000), // Integer value
			MinValue:     int64(1000),
			MaxValue:     int64(10000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum exemplars per tenant",
		},
		
		// Request Rate Limits
		"request_rate": {
			Name:         "request_rate",
			Type:         "rate",
			MetricSource: "http_requests_total",
			DefaultValue: 0.0, // 0 = unlimited - Keep as float for rate
			MinValue:     0.0,
			MaxValue:     10000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Request rate limit (requests/sec)",
		},
		"request_burst_size": {
			Name:         "request_burst_size",
			Type:         "count",
			MetricSource: "http_requests_total",
			DefaultValue: int64(0), // 0 = unlimited - Integer value
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Request burst size",
		},
		
		// Ruler Limits
		"ruler_max_rules_per_rule_group": {
			Name:         "ruler_max_rules_per_rule_group",
			Type:         "count",
			MetricSource: "prometheus_rule_group_rules",
			DefaultValue: int64(20), // Integer value
			MinValue:     int64(1),
			MaxValue:     int64(1000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum rules per rule group",
		},
		"ruler_max_rule_groups_per_tenant": {
			Name:         "ruler_max_rule_groups_per_tenant",
			Type:         "count",
			MetricSource: "prometheus_rule_group_rules",
			DefaultValue: int64(100), // Integer value
			MinValue:     int64(1),
			MaxValue:     int64(10000),
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum rule groups per tenant",
		},
		
		// Alertmanager Limits
		"alertmanager_notification_rate_limit": {
			Name:         "alertmanager_notification_rate_limit",
			Type:         "rate",
			MetricSource: "alertmanager_notifications_total",
			DefaultValue: 0.0, // 0 = unlimited - Keep as float for rate
			MinValue:     0.0,
			MaxValue:     1000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Alertmanager notification rate limit",
		},
		"alertmanager_max_dispatcher_aggregation_groups": {
			Name:         "alertmanager_max_dispatcher_aggregation_groups",
			Type:         "count",
			MetricSource: "alertmanager_dispatcher_aggregation_groups",
			DefaultValue: int64(0), // 0 = unlimited - Integer value
			MinValue:     int64(0),
			MaxValue:     int64(10000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum dispatcher aggregation groups",
		},
		"alertmanager_max_alerts_count": {
			Name:         "alertmanager_max_alerts_count",
			Type:         "count",
			MetricSource: "alertmanager_alerts",
			DefaultValue: int64(0), // 0 = unlimited - Integer value
			MinValue:     int64(0),
			MaxValue:     int64(1000000),
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum number of alerts",
		},
		"alertmanager_max_alerts_size_bytes": {
			Name:         "alertmanager_max_alerts_size_bytes",
			Type:         "size",
			MetricSource: "alertmanager_alerts_size_bytes",
			DefaultValue: int64(0), // 0 = unlimited - Integer value
			MinValue:     int64(0),
			MaxValue:     int64(100000000), // 100MB
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum size of all alerts in bytes",
		},
		
		// Compactor Limits
		"compactor_blocks_retention_period": {
			Name:         "compactor_blocks_retention_period",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s", // 0 = unlimited - Keep as string for duration
			MinValue:     "0s",
			MaxValue:     "8760h", // 1 year
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Retention period for compacted blocks",
		},
		
		// Store Gateway Limits
		"store_gateway_tenant_shard_size": {
			Name:         "store_gateway_tenant_shard_size",
			Type:         "count",
			MetricSource: "",
			DefaultValue: int64(0), // 0 = no sharding - Integer value
			MinValue:     int64(0),
			MaxValue:     int64(1000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for store gateway",
		},
		
		// Label Limits
		"max_label_names_per_series": {
			Name:         "max_label_names_per_series",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: int64(30), // Integer value
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
			DefaultValue: int64(1024), // Integer value
			MinValue:     int64(10),
			MaxValue:     int64(10000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label names",
		},
		"max_label_value_length": {
			Name:         "max_label_value_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: int64(2048), // Integer value
			MinValue:     int64(10),
			MaxValue:     int64(100000),
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label values",
		},
		
		// Out-of-order samples
		"out_of_order_time_window": {
			Name:         "out_of_order_time_window",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s", // 0 = disabled - Keep as string for duration
			MinValue:     "0s",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Time window to accept out-of-order samples",
		},
	}
} 