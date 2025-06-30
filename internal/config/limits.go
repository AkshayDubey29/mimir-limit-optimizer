package config

// GetDefaultLimitDefinitions returns predefined configurations for all major Mimir limits
func GetDefaultLimitDefinitions() map[string]LimitDefinition {
	return map[string]LimitDefinition{
		// Ingestion Limits
		"ingestion_rate": {
			Name:         "ingestion_rate",
			Type:         "rate",
			MetricSource: "prometheus_remote_storage_samples_in_total",
			DefaultValue: 25000.0,
			MinValue:     1000.0,
			MaxValue:     10000000.0,
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Rate limit for sample ingestion (samples/sec)",
		},
		"ingestion_burst_size": {
			Name:         "ingestion_burst_size",
			Type:         "count",
			MetricSource: "prometheus_remote_storage_samples_in_total",
			DefaultValue: 50000.0,
			MinValue:     2000.0,
			MaxValue:     20000000.0,
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Burst size for sample ingestion",
		},
		
		// Series Limits
		"max_global_series_per_user": {
			Name:         "max_global_series_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: 150000.0,
			MinValue:     1000.0,
			MaxValue:     100000000.0,
			BufferFactor: 20.0,
			Enabled:      true,
			Description:  "Maximum series per tenant",
		},
		"max_global_series_per_metric": {
			Name:         "max_global_series_per_metric",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: 0.0, // 0 = unlimited
			MinValue:     0.0,
			MaxValue:     1000000.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum series per metric name",
		},
		
		// Query Limits
		"max_samples_per_query": {
			Name:         "max_samples_per_query",
			Type:         "count",
			MetricSource: "prometheus_engine_query_samples_total",
			DefaultValue: 50000000.0,
			MinValue:     1000.0,
			MaxValue:     1000000000.0,
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum samples a single query can load",
		},
		"max_series_per_query": {
			Name:         "max_series_per_query",
			Type:         "count",
			MetricSource: "prometheus_engine_query_series_total",
			DefaultValue: 100000.0,
			MinValue:     100.0,
			MaxValue:     10000000.0,
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
			DefaultValue: 2000000.0,
			MinValue:     1000.0,
			MaxValue:     100000000.0,
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunks a single query can fetch",
		},
		"max_fetched_series_per_query": {
			Name:         "max_fetched_series_per_query", 
			Type:         "count",
			MetricSource: "prometheus_engine_query_series_total",
			DefaultValue: 100000.0,
			MinValue:     100.0,
			MaxValue:     10000000.0,
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum series a single query can fetch",
		},
		"max_fetched_chunk_bytes_per_query": {
			Name:         "max_fetched_chunk_bytes_per_query",
			Type:         "size",
			MetricSource: "prometheus_tsdb_compaction_chunk_size_bytes",
			DefaultValue: 50000000.0, // 50MB
			MinValue:     1000000.0,  // 1MB
			MaxValue:     1000000000.0, // 1GB
			BufferFactor: 50.0,
			Enabled:      true,
			Description:  "Maximum chunk bytes a single query can fetch",
		},
		
		// Metadata Limits
		"max_global_metadata_per_user": {
			Name:         "max_global_metadata_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: 8000.0,
			MinValue:     100.0,
			MaxValue:     1000000.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per tenant",
		},
		"max_global_metadata_per_metric": {
			Name:         "max_global_metadata_per_metric",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: 10.0,
			MinValue:     1.0,
			MaxValue:     100.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum metadata entries per metric",
		},
		
		// Exemplar Limits
		"max_global_exemplars_per_user": {
			Name:         "max_global_exemplars_per_user",
			Type:         "count",
			MetricSource: "prometheus_tsdb_exemplar_exemplars_total",
			DefaultValue: 100000.0,
			MinValue:     1000.0,
			MaxValue:     10000000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum exemplars per tenant",
		},
		
		// Request Rate Limits
		"request_rate": {
			Name:         "request_rate",
			Type:         "rate",
			MetricSource: "http_requests_total",
			DefaultValue: 0.0, // 0 = unlimited
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
			DefaultValue: 0.0, // 0 = unlimited
			MinValue:     0.0,
			MaxValue:     1000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Request burst size",
		},
		
		// Ruler Limits
		"ruler_max_rules_per_rule_group": {
			Name:         "ruler_max_rules_per_rule_group",
			Type:         "count",
			MetricSource: "prometheus_rule_group_rules",
			DefaultValue: 100.0,
			MinValue:     1.0,
			MaxValue:     10000.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum rules per rule group",
		},
		"ruler_max_rule_groups_per_tenant": {
			Name:         "ruler_max_rule_groups_per_tenant",
			Type:         "count",
			MetricSource: "prometheus_rule_group_rules",
			DefaultValue: 100.0,
			MinValue:     1.0,
			MaxValue:     10000.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum rule groups per tenant",
		},
		
		// Alertmanager Limits
		"alertmanager_notification_rate_limit": {
			Name:         "alertmanager_notification_rate_limit",
			Type:         "rate",
			MetricSource: "alertmanager_notifications_total",
			DefaultValue: 0.0, // 0 = unlimited
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
			DefaultValue: 0.0, // 0 = unlimited
			MinValue:     0.0,
			MaxValue:     10000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum dispatcher aggregation groups",
		},
		"alertmanager_max_alerts_count": {
			Name:         "alertmanager_max_alerts_count",
			Type:         "count",
			MetricSource: "alertmanager_alerts",
			DefaultValue: 0.0, // 0 = unlimited
			MinValue:     0.0,
			MaxValue:     1000000.0,
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum number of alerts",
		},
		"alertmanager_max_alerts_size_bytes": {
			Name:         "alertmanager_max_alerts_size_bytes",
			Type:         "size",
			MetricSource: "alertmanager_alerts_size_bytes",
			DefaultValue: 0.0, // 0 = unlimited
			MinValue:     0.0,
			MaxValue:     100000000.0, // 100MB
			BufferFactor: 50.0,
			Enabled:      false,
			Description:  "Maximum size of all alerts in bytes",
		},
		
		// Compactor Limits
		"compactor_blocks_retention_period": {
			Name:         "compactor_blocks_retention_period",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s", // 0 = unlimited
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
			DefaultValue: 0.0, // 0 = no sharding
			MinValue:     0.0,
			MaxValue:     1000.0,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Tenant shard size for store gateway",
		},
		
		// Label Limits
		"max_label_names_per_series": {
			Name:         "max_label_names_per_series",
			Type:         "count",
			MetricSource: "prometheus_tsdb_head_series",
			DefaultValue: 30.0,
			MinValue:     1.0,
			MaxValue:     1000.0,
			BufferFactor: 20.0,
			Enabled:      false,
			Description:  "Maximum label names per series",
		},
		"max_label_name_length": {
			Name:         "max_label_name_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: 1024.0,
			MinValue:     10.0,
			MaxValue:     10000.0,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label names",
		},
		"max_label_value_length": {
			Name:         "max_label_value_length",
			Type:         "size",
			MetricSource: "",
			DefaultValue: 2048.0,
			MinValue:     10.0,
			MaxValue:     100000.0,
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Maximum length of label values",
		},
		
		// Out-of-order samples
		"out_of_order_time_window": {
			Name:         "out_of_order_time_window",
			Type:         "duration",
			MetricSource: "",
			DefaultValue: "0s", // 0 = disabled
			MinValue:     "0s",
			MaxValue:     "1h",
			BufferFactor: 0.0,
			Enabled:      false,
			Description:  "Time window to accept out-of-order samples",
		},
	}
} 