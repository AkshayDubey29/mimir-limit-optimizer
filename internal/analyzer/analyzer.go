package analyzer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/collector"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

// TenantLimits represents calculated limits for a tenant with dynamic limit support
type TenantLimits struct {
	Tenant      string
	Limits      map[string]interface{} // Dynamic limits map supporting all Mimir limit types
	LastUpdated time.Time
	Reason      string
	Source      string
}

// LimitDefinition defines how to handle a specific limit type
type LimitDefinition struct {
	Name          string      // Mimir configuration name (e.g., "ingestion_rate")
	Type          string      // "rate", "count", "size", "duration", "percentage"
	MetricSource  string      // Which metric to use for calculations
	DefaultValue  interface{} // Default value if not specified
	MinValue      interface{} // Minimum allowed value
	MaxValue      interface{} // Maximum allowed value
	BufferFactor  float64     // Buffer percentage to apply
	Enabled       bool        // Whether this limit should be optimized
}

// AnalysisResult contains the results of trend analysis
type AnalysisResult struct {
	Tenant           string
	MetricName       string
	CurrentValue     float64
	MovingAverage    float64
	Percentile       float64
	Peak             float64
	Trend            float64
	SpikeDetected    bool
	SpikeMultiplier  float64
	RecommendedLimit float64
	AnalysisTime     time.Time
}

// Analyzer interface defines methods for analyzing metrics and calculating limits
type Analyzer interface {
	AnalyzeTrends(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string][]AnalysisResult, error)
	CalculateLimits(ctx context.Context, analysisResults map[string][]AnalysisResult) (map[string]*TenantLimits, error)
	DetectSpikes(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string]map[string]bool, error)
}

// TrendAnalyzer implements the Analyzer interface
type TrendAnalyzer struct {
	config          *config.Config
	log             logr.Logger
	historicalData  map[string]map[string][]collector.MetricData
	spikeState      map[string]map[string]*SpikeInfo
}

// SpikeInfo tracks spike detection state
type SpikeInfo struct {
	Detected     bool
	StartTime    time.Time
	Multiplier   float64
	BaseValue    float64
	CooldownUntil time.Time
}

// NewTrendAnalyzer creates a new TrendAnalyzer
func NewTrendAnalyzer(cfg *config.Config, log logr.Logger) *TrendAnalyzer {
	return &TrendAnalyzer{
		config:         cfg,
		log:            log,
		historicalData: make(map[string]map[string][]collector.MetricData),
		spikeState:     make(map[string]map[string]*SpikeInfo),
	}
}

// AnalyzeTrends analyzes trends in tenant metrics
func (a *TrendAnalyzer) AnalyzeTrends(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string][]AnalysisResult, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.TrendMetricsInstance.ObserveTrendAnalysisDuration(duration)
	}()

	results := make(map[string][]AnalysisResult)

	// Update historical data
	a.updateHistoricalData(tenantMetrics)

	for tenant, tm := range tenantMetrics {
		var tenantResults []AnalysisResult

		for metricName, metricData := range tm.Metrics {
			if !a.isAnalyzableMetric(metricName) {
				continue
			}

			analysis, err := a.analyzeMetric(tenant, metricName, metricData)
			if err != nil {
				a.log.Error(err, "failed to analyze metric", "tenant", tenant, "metric", metricName)
				continue
			}

			tenantResults = append(tenantResults, *analysis)

			// Update metrics
			metrics.TenantMetricsInstance.SetTenantUsagePercentile(
				tenant, metricName, fmt.Sprintf("%.0f", a.config.TrendAnalysis.Percentile), analysis.Percentile)
		}

		if len(tenantResults) > 0 {
			results[tenant] = tenantResults
		}
	}

	a.log.Info("analyzed trends", "tenants", len(results))

	return results, nil
}

// CalculateLimits calculates optimal limits based on analysis results
func (a *TrendAnalyzer) CalculateLimits(ctx context.Context, analysisResults map[string][]AnalysisResult) (map[string]*TenantLimits, error) {
	limits := make(map[string]*TenantLimits)

	for tenant, results := range analysisResults {
		tenantLimits := &TenantLimits{
			Tenant:      tenant,
			Limits:      make(map[string]interface{}),
			LastUpdated: time.Now(),
			Reason:      "trend-analysis",
			Source:      "analyzer",
		}

		// Calculate limits based on different metrics
		for _, result := range results {
			a.applyMetricToLimits(tenantLimits, result)
		}

		// Apply buffer percentage
		a.applyBufferPercentage(tenantLimits, tenant)

		// Apply min/max constraints
		a.applyConstraints(tenantLimits, tenant)

		limits[tenant] = tenantLimits
	}

	return limits, nil
}

// DetectSpikes detects usage spikes in real-time
func (a *TrendAnalyzer) DetectSpikes(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string]map[string]bool, error) {
	if !a.config.EventSpike.Enabled {
		return nil, nil
	}

	spikes := make(map[string]map[string]bool)

	for tenant, tm := range tenantMetrics {
		tenantSpikes := make(map[string]bool)

		for metricName, metricData := range tm.Metrics {
			if !a.isAnalyzableMetric(metricName) {
				continue
			}

			spikeDetected := a.detectSpikeForMetric(tenant, metricName, metricData)
			if spikeDetected {
				tenantSpikes[metricName] = true
				metrics.SpikeMetricsInstance.IncSpikesDetected(tenant, metricName)
			}
		}

		if len(tenantSpikes) > 0 {
			spikes[tenant] = tenantSpikes
		}
	}

	return spikes, nil
}

// analyzeMetric analyzes a single metric for trend patterns
func (a *TrendAnalyzer) analyzeMetric(tenant, metricName string, data []collector.MetricData) (*AnalysisResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data available for metric %s", metricName)
	}

	// Get historical data for better analysis
	historical := a.getHistoricalData(tenant, metricName)
	allData := append(historical, data...)

	// Sort by timestamp
	sort.Slice(allData, func(i, j int) bool {
		return allData[i].Timestamp.Before(allData[j].Timestamp)
	})

	// Filter data within analysis window
	cutoff := time.Now().Add(-a.config.TrendAnalysis.AnalysisWindow)
	var windowData []collector.MetricData
	for _, d := range allData {
		if d.Timestamp.After(cutoff) {
			windowData = append(windowData, d)
		}
	}

	if len(windowData) == 0 {
		return nil, fmt.Errorf("no data within analysis window for metric %s", metricName)
	}

	// Calculate statistics
	values := make([]float64, len(windowData))
	for i, d := range windowData {
		values[i] = d.Value
	}

	result := &AnalysisResult{
		Tenant:       tenant,
		MetricName:   metricName,
		CurrentValue: values[len(values)-1],
		AnalysisTime: time.Now(),
	}

	// Calculate moving average
	if a.config.TrendAnalysis.UseMovingAverage {
		result.MovingAverage = a.calculateMovingAverage(values)
	}

	// Calculate percentile
	result.Percentile = a.calculatePercentile(values, a.config.TrendAnalysis.Percentile)

	// Calculate peak
	if a.config.TrendAnalysis.IncludePeaks {
		result.Peak = a.calculatePeak(values)
	}

	// Calculate trend direction
	result.Trend = a.calculateTrend(values)

	// Check for spikes
	if a.config.EventSpike.Enabled {
		spikeInfo := a.getSpikeInfo(tenant, metricName)
		if spikeInfo != nil && spikeInfo.Detected {
			result.SpikeDetected = true
			result.SpikeMultiplier = spikeInfo.Multiplier
		}
	}

	// Calculate recommended limit
	result.RecommendedLimit = a.calculateRecommendedLimit(result)

	return result, nil
}

// updateHistoricalData updates the historical data cache
func (a *TrendAnalyzer) updateHistoricalData(tenantMetrics map[string]*collector.TenantMetrics) {
	for tenant, tm := range tenantMetrics {
		if a.historicalData[tenant] == nil {
			a.historicalData[tenant] = make(map[string][]collector.MetricData)
		}

		for metricName, data := range tm.Metrics {
			// Append new data
			a.historicalData[tenant][metricName] = append(a.historicalData[tenant][metricName], data...)

			// Cleanup old data (keep only data within analysis window + buffer)
			cutoff := time.Now().Add(-a.config.TrendAnalysis.AnalysisWindow * 2)
			var filtered []collector.MetricData
			for _, d := range a.historicalData[tenant][metricName] {
				if d.Timestamp.After(cutoff) {
					filtered = append(filtered, d)
				}
			}
			a.historicalData[tenant][metricName] = filtered
		}
	}
}

// getHistoricalData retrieves historical data for a metric
func (a *TrendAnalyzer) getHistoricalData(tenant, metricName string) []collector.MetricData {
	if a.historicalData[tenant] == nil {
		return nil
	}
	return a.historicalData[tenant][metricName]
}

// detectSpikeForMetric detects spikes for a specific metric
func (a *TrendAnalyzer) detectSpikeForMetric(tenant, metricName string, data []collector.MetricData) bool {
	if len(data) == 0 {
		return false
	}

	// Get spike state
	spikeInfo := a.getSpikeInfo(tenant, metricName)
	if spikeInfo == nil {
		spikeInfo = &SpikeInfo{}
		a.setSpikeInfo(tenant, metricName, spikeInfo)
	}

	// Check if we're in cooldown
	if time.Now().Before(spikeInfo.CooldownUntil) {
		return spikeInfo.Detected
	}

	// Get recent baseline
	historical := a.getHistoricalData(tenant, metricName)
	if len(historical) < 10 { // Need enough data for baseline
		return false
	}

	// Calculate baseline (average of older data)
	baselineCutoff := time.Now().Add(-a.config.EventSpike.DetectionWindow * 2)
	var baselineValues []float64
	for _, d := range historical {
		if d.Timestamp.Before(baselineCutoff) {
			baselineValues = append(baselineValues, d.Value)
		}
	}

	if len(baselineValues) < 5 {
		return false
	}

	baseline := a.calculateMovingAverage(baselineValues)
	currentValue := data[len(data)-1].Value

	// Check for spike
	if currentValue > baseline*a.config.EventSpike.Threshold {
		if !spikeInfo.Detected {
			// New spike detected
			spikeInfo.Detected = true
			spikeInfo.StartTime = time.Now()
			spikeInfo.BaseValue = baseline
			spikeInfo.Multiplier = math.Min(currentValue/baseline, a.config.EventSpike.MaxSpikeMultiplier)
			spikeInfo.CooldownUntil = time.Now().Add(a.config.EventSpike.CooldownPeriod)

			metrics.SpikeMetricsInstance.SetSpikeMultiplier(tenant, spikeInfo.Multiplier)
		}
		return true
	}

	// No spike detected, reset if previously detected
	if spikeInfo.Detected {
		spikeInfo.Detected = false
		spikeInfo.Multiplier = 1.0
		metrics.SpikeMetricsInstance.SetSpikeMultiplier(tenant, 1.0)
	}

	return false
}

// Helper methods for calculations
func (a *TrendAnalyzer) calculateMovingAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (a *TrendAnalyzer) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := (percentile / 100.0) * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (a *TrendAnalyzer) calculatePeak(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func (a *TrendAnalyzer) calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Simple linear trend calculation
	n := float64(len(values))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, v := range values {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

func (a *TrendAnalyzer) calculateRecommendedLimit(result *AnalysisResult) float64 {
	base := result.Percentile

	if a.config.TrendAnalysis.UseMovingAverage && result.MovingAverage > 0 {
		base = math.Max(base, result.MovingAverage)
	}

	if a.config.TrendAnalysis.IncludePeaks && result.Peak > 0 {
		base = math.Max(base, result.Peak*0.8) // Use 80% of peak
	}

	// Apply spike multiplier if detected
	if result.SpikeDetected {
		base *= result.SpikeMultiplier
	}

	return base
}

// Utility methods
func (a *TrendAnalyzer) isAnalyzableMetric(metricName string) bool {
	analyzableMetrics := []string{
		// PRIMARY MIMIR METRICS - Core metrics for limit optimization
		"cortex_distributor_received_samples_total",
		"cortex_distributor_samples_in_total",
		"cortex_ingester_ingested_samples_total",
		"cortex_ingester_memory_series",
		"cortex_ingester_memory_users",
		"cortex_query_frontend_queries_total",
		"cortex_querier_queries_total",
		"cortex_query_frontend_query_duration_seconds",
		"cortex_querier_query_duration_seconds",
		"cortex_ingester_ingested_samples_failures_total",
		
		// BYTE-BASED INGESTION METRICS - NEW COMPREHENSIVE COVERAGE
		"cortex_distributor_received_samples_bytes_total",
		"cortex_distributor_push_duration_seconds",
		"cortex_distributor_latest_seen_sample_timestamp_seconds",
		"cortex_ingester_oldest_unshipped_block_timestamp_seconds",
		"cortex_ingester_chunk_size_bytes",
		
		// QUERY SCHEDULER METRICS - NEW COMPREHENSIVE COVERAGE
		"cortex_query_scheduler_queue_length",
		"cortex_query_scheduler_queriers_connected",
		"cortex_query_scheduler_queries_in_progress",
		"cortex_query_frontend_queue_length",
		"cortex_query_frontend_queries_in_progress",
		
		// STORAGE GATEWAY METRICS - NEW COMPREHENSIVE COVERAGE
		"cortex_bucket_store_queries_in_flight",
		
		// REQUEST & CONCURRENCY METRICS - NEW COMPREHENSIVE COVERAGE
		"cortex_request_duration_seconds",
		"cortex_querier_chunks_fetched_bytes",
		
		// CARDINALITY & SERIES METRICS - ENHANCED COVERAGE
		"cortex_ingester_memory_series_per_metric",
		"cortex_ingester_memory_metadata",
		"cortex_ingester_memory_metadata_per_metric",
		"cortex_querier_series_fetched",
		"cortex_querier_chunks_fetched",
		"cortex_querier_estimated_memory_consumption_bytes",
		"cortex_querier_estimated_chunks_fetched",
		"cortex_querier_exemplars_fetched",
		
		// RULER METRICS - ENHANCED COVERAGE
		"cortex_ruler_rule_group_rules",
		"cortex_ruler_rule_groups_per_user",
		"cortex_ruler_rules_per_user",
		
		// ALERTMANAGER METRICS - COMPREHENSIVE COVERAGE
		"cortex_alertmanager_notifications_total",
		"cortex_alertmanager_dispatcher_aggregation_groups",
		"cortex_alertmanager_alerts",
		"cortex_alertmanager_alerts_size_bytes",
		
		// INGESTER TSDB METRICS - ENHANCED COVERAGE
		"cortex_ingester_tsdb_exemplar_series_with_exemplars_in_storage",
		"cortex_ingester_active_series",
		
		// EXTENDED MIMIR METRICS - Additional metrics for comprehensive analysis
		"cortex_distributor_deduped_samples_total",
		"cortex_distributor_non_ha_samples_received_total",
		"cortex_ingester_chunks_created_total",
		"cortex_ingester_series_removed_total",
		"cortex_querier_series_fetched_total",
		"cortex_querier_chunks_fetched_total",
		"cortex_querier_estimated_series_count",
		"cortex_compactor_runs_total",
		"cortex_ruler_queries_total",
		
		// PROMETHEUS FALLBACK METRICS
		"prometheus_remote_storage_samples_in_total",
		"prometheus_tsdb_head_series",
		"prometheus_engine_query_samples_total",
		"prometheus_engine_query_series_total",
		"prometheus_tsdb_head_chunks",
		"prometheus_tsdb_compaction_chunk_size_bytes",
		"prometheus_tsdb_exemplar_exemplars_total",
		"prometheus_rule_group_rules",
		"alertmanager_notifications_total",
		"alertmanager_alerts",
		"http_requests_total",
	}

	for _, analyzable := range analyzableMetrics {
		if metricName == analyzable {
			return true
		}
	}
	
	// Also allow metrics that contain these patterns (for dynamic discovery)
	analyzablePatterns := []string{
		"cortex_distributor",
		"cortex_ingester", 
		"cortex_querier",
		"cortex_query_frontend",
		"cortex_ruler",
		"cortex_compactor",
		"prometheus_tsdb",
		"prometheus_engine",
		"prometheus_remote_storage",
	}
	
	for _, pattern := range analyzablePatterns {
		if strings.Contains(metricName, pattern) {
			return true
		}
	}
	
	return false
}

func (a *TrendAnalyzer) getSpikeInfo(tenant, metricName string) *SpikeInfo {
	if a.spikeState[tenant] == nil {
		return nil
	}
	return a.spikeState[tenant][metricName]
}

func (a *TrendAnalyzer) setSpikeInfo(tenant, metricName string, info *SpikeInfo) {
	if a.spikeState[tenant] == nil {
		a.spikeState[tenant] = make(map[string]*SpikeInfo)
	}
	a.spikeState[tenant][metricName] = info
}

// applyMetricToLimits applies analysis results to the dynamic limits map
func (a *TrendAnalyzer) applyMetricToLimits(limits *TenantLimits, result AnalysisResult) {
	// Get the limit mapping from metric name to limit field
	limitMapping := a.getMetricToLimitMapping()
	
	if limitName, exists := limitMapping[result.MetricName]; exists {
		// Check if this limit is enabled in configuration
		if limitDef, found := a.config.DynamicLimits.LimitDefinitions[limitName]; found && limitDef.Enabled {
			limits.Limits[limitName] = result.RecommendedLimit
		}
	}
}

// getMetricToLimitMapping returns mapping from metric names to Mimir limit names
func (a *TrendAnalyzer) getMetricToLimitMapping() map[string]string {
	return map[string]string{
		// MIMIR/CORTEX METRICS - Primary mappings for Mimir deployments
		"cortex_distributor_received_samples_total":     "ingestion_rate",
		"cortex_distributor_samples_in_total":           "ingestion_rate",
		"cortex_ingester_ingested_samples_total":        "ingestion_rate",
		"cortex_ingester_memory_series":                 "max_global_series_per_user",
		"cortex_ingester_memory_users":                  "max_tenants",
		"cortex_query_frontend_queries_total":           "max_samples_per_query",
		"cortex_querier_queries_total":                  "max_samples_per_query",
		"cortex_query_frontend_query_duration_seconds":  "query_timeout",
		"cortex_querier_query_duration_seconds":         "query_timeout",
		"cortex_ingester_ingested_samples_failures_total": "ingestion_rate",
		
		// BYTE-BASED INGESTION METRICS - NEW COMPREHENSIVE MAPPINGS
		"cortex_distributor_received_samples_bytes_total": "max_ingestion_rate_bytes",
		"cortex_distributor_push_duration_seconds":       "remote_write_deadline",
		"cortex_distributor_latest_seen_sample_timestamp_seconds": "max_sample_age",
		"cortex_ingester_oldest_unshipped_block_timestamp_seconds": "max_chunk_age",
		"cortex_ingester_chunk_size_bytes":               "max_chunk_size_bytes",
		
		// QUERY SCHEDULER METRICS - NEW COMPREHENSIVE MAPPINGS
		"cortex_query_scheduler_queue_length":            "query_scheduler_max_outstanding_requests_per_tenant",
		"cortex_query_scheduler_queriers_connected":      "query_scheduler_max_queriers_per_tenant", 
		"cortex_query_scheduler_queries_in_progress":     "query_scheduler_max_active_requests",
		"cortex_query_frontend_queue_length":             "max_outstanding_per_tenant",
		"cortex_query_frontend_queries_in_progress":      "max_concurrent_queries",
		
		// STORAGE GATEWAY METRICS - NEW COMPREHENSIVE MAPPINGS
		"cortex_bucket_store_queries_in_flight":          "store_gateway_max_queries_in_flight",
		
		// REQUEST & CONCURRENCY METRICS - NEW COMPREHENSIVE MAPPINGS
		"cortex_request_duration_seconds":                "max_concurrent_requests",
		"cortex_querier_chunks_fetched_bytes":            "max_bytes_per_query",
		
		// CARDINALITY & SERIES METRICS - ENHANCED MAPPINGS
		"cortex_ingester_memory_series_per_metric":       "max_global_series_per_metric",
		"cortex_ingester_memory_metadata":                "max_global_metadata_per_user",
		"cortex_ingester_memory_metadata_per_metric":     "max_global_metadata_per_metric",
		
		// PROMETHEUS METRICS - Fallback mappings for Prometheus deployments
		"prometheus_remote_storage_samples_in_total":     "ingestion_rate",
		"prometheus_remote_storage_samples_burst":        "ingestion_burst_size",
		"prometheus_tsdb_head_series":                    "max_global_series_per_user",
		"prometheus_tsdb_blocks_loaded":                  "retention_period",
		"prometheus_engine_query_samples_total":          "max_samples_per_query",
		"prometheus_engine_query_series_total":           "max_series_per_query",
		"prometheus_tsdb_head_chunks":                    "max_fetched_chunks_per_query",
		"prometheus_tsdb_compaction_chunk_size_bytes":    "max_fetched_chunk_bytes_per_query",
		"prometheus_tsdb_exemplar_exemplars_total":       "max_global_exemplars_per_user",
		"prometheus_rule_group_rules":                    "ruler_max_rules_per_rule_group",
		"alertmanager_notifications_total":              "alertmanager_notification_rate_limit",
		"alertmanager_alerts":                            "alertmanager_max_alerts_count",
		"http_requests_total":                            "request_rate",
		
		// EXTENDED MIMIR METRICS - Comprehensive limit support
		"cortex_distributor_deduped_samples_total":       "ingestion_rate",
		"cortex_distributor_non_ha_samples_received_total": "ingestion_rate",
		"cortex_ingester_chunks_created_total":           "max_chunks_per_query",
		"cortex_ingester_series_removed_total":           "max_global_series_per_user",
		"cortex_querier_series_fetched_total":            "max_series_per_query",
		"cortex_querier_series_fetched":                  "max_fetched_series_per_query",
		"cortex_querier_chunks_fetched_total":            "max_fetched_chunks_per_query",
		"cortex_querier_chunks_fetched":                  "max_fetched_chunks_per_query",
		"cortex_querier_estimated_series_count":          "max_series_per_query",
		"cortex_querier_estimated_memory_consumption_bytes": "max_estimated_memory_consumption_per_query",
		"cortex_querier_estimated_chunks_fetched":        "max_estimated_fetched_chunks_per_query",
		"cortex_querier_exemplars_fetched":               "max_exemplars_per_query",
		"cortex_compactor_runs_total":                    "compactor_blocks_retention_period",
		"cortex_ruler_queries_total":                     "ruler_max_rule_groups_per_tenant",
		"cortex_ruler_rule_group_rules":                  "ruler_max_rules_per_rule_group",
		"cortex_ruler_rule_groups_per_user":              "ruler_max_rule_groups_per_tenant",
		"cortex_ruler_rules_per_user":                    "ruler_max_rules_per_tenant",
		
		// ALERTMANAGER METRICS - Comprehensive coverage
		"cortex_alertmanager_notifications_total":        "alertmanager_notification_rate_limit",
		"cortex_alertmanager_dispatcher_aggregation_groups": "alertmanager_max_dispatcher_aggregation_groups",
		"cortex_alertmanager_alerts":                     "alertmanager_max_alerts_count",
		"cortex_alertmanager_alerts_size_bytes":          "alertmanager_max_alerts_size_bytes",
		
		// INGESTER TSDB METRICS - Enhanced coverage
		"cortex_ingester_tsdb_exemplar_series_with_exemplars_in_storage": "max_global_exemplars_per_user",
		"cortex_ingester_active_series":                  "max_label_names_per_series",
	}
}

// applyBufferPercentage applies buffer to all dynamic limits
func (a *TrendAnalyzer) applyBufferPercentage(limits *TenantLimits, tenant string) {
	for limitName, limitValue := range limits.Limits {
		if limitDef, exists := a.config.DynamicLimits.LimitDefinitions[limitName]; exists {
			bufferFactor := limitDef.BufferFactor
			if bufferFactor == 0 {
				bufferFactor = a.config.DynamicLimits.DefaultBuffer
			}
			
			// Apply buffer based on limit type
			switch limitDef.Type {
			case "rate", "count", "size":
				if val, ok := limitValue.(float64); ok && bufferFactor > 0 {
					bufferedValue := val * (1 + bufferFactor/100)
					limits.Limits[limitName] = bufferedValue
				}
			case "percentage":
				// Percentage limits don't typically need buffers
			case "duration":
				// Duration limits are handled as strings, skip buffering
			}
		}
	}
}

// applyConstraints applies min/max constraints to all dynamic limits
func (a *TrendAnalyzer) applyConstraints(limits *TenantLimits, tenant string) {
	for limitName, limitValue := range limits.Limits {
		if limitDef, exists := a.config.DynamicLimits.LimitDefinitions[limitName]; exists {
			// Apply constraints based on limit type
			switch limitDef.Type {
			case "rate", "count", "size", "percentage":
				val, ok := limitValue.(float64)
				if !ok {
					continue
				}
				
				// Apply minimum constraint
				if limitDef.MinValue != nil {
					if minVal, ok := limitDef.MinValue.(float64); ok && val < minVal {
						limits.Limits[limitName] = minVal
						val = minVal
					}
				}
				
				// Apply maximum constraint
				if limitDef.MaxValue != nil {
					if maxVal, ok := limitDef.MaxValue.(float64); ok && val > maxVal {
						limits.Limits[limitName] = maxVal
					}
				}
			case "duration":
				// Duration constraints would need parsing - skip for now
			}
		}
	}
}

// NewAnalyzer creates the appropriate analyzer based on configuration
func NewAnalyzer(cfg *config.Config, log logr.Logger) Analyzer {
	return NewTrendAnalyzer(cfg, log)
} 