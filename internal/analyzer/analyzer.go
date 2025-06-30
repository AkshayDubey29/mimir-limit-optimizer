package analyzer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-logr/logr"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/collector"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

// TenantLimits represents calculated limits for a tenant
type TenantLimits struct {
	Tenant            string
	IngestionRate     float64
	IngestionBurst    float64
	MaxSeries         float64
	MaxSamplesPerQuery float64
	MaxQueryLookback  time.Duration
	LastUpdated       time.Time
	Reason            string
	Source            string
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
		"cortex_distributor_received_samples_total",
		"cortex_ingester_memory_series",
		"cortex_querier_queries_total",
		"cortex_query_frontend_queries_total",
	}

	for _, analyzable := range analyzableMetrics {
		if metricName == analyzable {
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

func (a *TrendAnalyzer) applyMetricToLimits(limits *TenantLimits, result AnalysisResult) {
	switch result.MetricName {
	case "cortex_distributor_received_samples_total":
		limits.IngestionRate = math.Max(limits.IngestionRate, result.RecommendedLimit)
		limits.IngestionBurst = math.Max(limits.IngestionBurst, result.RecommendedLimit*2)
	case "cortex_ingester_memory_series":
		limits.MaxSeries = math.Max(limits.MaxSeries, result.RecommendedLimit)
	case "cortex_querier_queries_total", "cortex_query_frontend_queries_total":
		limits.MaxSamplesPerQuery = math.Max(limits.MaxSamplesPerQuery, result.RecommendedLimit*1000)
	}
}

func (a *TrendAnalyzer) applyBufferPercentage(limits *TenantLimits, tenant string) {
	bufferPercentage := a.config.BufferPercentage

	// Check for tenant-specific buffer
	if tierConfig, exists := a.config.Limits.TenantTiers[tenant]; exists {
		bufferPercentage = tierConfig.BufferPercentage
	}

	multiplier := 1.0 + (bufferPercentage / 100.0)

	limits.IngestionRate *= multiplier
	limits.IngestionBurst *= multiplier
	limits.MaxSeries *= multiplier
	limits.MaxSamplesPerQuery *= multiplier
}

func (a *TrendAnalyzer) applyConstraints(limits *TenantLimits, tenant string) {
	// Apply global min/max constraints
	if minLimits := a.config.Limits.MinLimits; minLimits != nil {
		if val, ok := minLimits["ingestion_rate"].(float64); ok {
			limits.IngestionRate = math.Max(limits.IngestionRate, val)
		}
		if val, ok := minLimits["max_series"].(float64); ok {
			limits.MaxSeries = math.Max(limits.MaxSeries, val)
		}
	}

	if maxLimits := a.config.Limits.MaxLimits; maxLimits != nil {
		if val, ok := maxLimits["ingestion_rate"].(float64); ok {
			limits.IngestionRate = math.Min(limits.IngestionRate, val)
		}
		if val, ok := maxLimits["max_series"].(float64); ok {
			limits.MaxSeries = math.Min(limits.MaxSeries, val)
		}
	}
}

// NewAnalyzer creates the appropriate analyzer based on configuration
func NewAnalyzer(cfg *config.Config, log logr.Logger) Analyzer {
	return NewTrendAnalyzer(cfg, log)
} 