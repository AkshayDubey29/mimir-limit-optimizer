package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/discovery"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

// MetricData represents collected metric data for a tenant
type MetricData struct {
	Tenant        string
	MetricName    string
	Value         float64
	Timestamp     time.Time
	Labels        map[string]string
	Source        string
}

// TenantMetrics holds all metrics for a specific tenant
type TenantMetrics struct {
	Tenant     string
	Metrics    map[string][]MetricData
	LastUpdate time.Time
}

// Collector interface defines methods for collecting metrics
type Collector interface {
	CollectMetrics(ctx context.Context) (map[string]*TenantMetrics, error)
	GetTenantList(ctx context.Context) ([]string, error)
}

// MimirCollector implements the Collector interface for Mimir/Prometheus
type MimirCollector struct {
	config      *config.Config
	client      kubernetes.Interface
	discovery   *discovery.ServiceDiscovery
	httpClient  *http.Client
	log         logr.Logger
}

// NewMimirCollector creates a new MimirCollector
func NewMimirCollector(cfg *config.Config, client kubernetes.Interface, log logr.Logger) *MimirCollector {
	return &MimirCollector{
		config: cfg,
		client: client,
		discovery: discovery.NewServiceDiscovery(client, cfg, log.WithName("discovery")),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log,
	}
}

// CollectMetrics collects metrics from all configured sources
func (c *MimirCollector) CollectMetrics(ctx context.Context) (map[string]*TenantMetrics, error) {
	startTime := time.Now()
	tenantMetrics := make(map[string]*TenantMetrics)
	
	var sources []string
	
	// Add primary metrics endpoint if configured
	if c.config.MetricsEndpoint != "" {
		sources = append(sources, c.config.MetricsEndpoint)
	}
	
	// Add discovered services if auto-discovery is enabled
	if c.config.MetricsDiscovery.Enabled {
		c.log.V(1).Info("auto-discovery enabled, discovering metrics endpoints")
		discoveredSources, err := c.discovery.DiscoverMetricsEndpoints(ctx)
		if err != nil {
			c.log.Error(err, "failed to discover metrics endpoints")
			metrics.DiscoveryMetricsInstance.IncDiscoveryErrors()
		} else {
			c.log.V(1).Info("discovered metrics endpoints", "count", len(discoveredSources), "endpoints", discoveredSources)
			sources = append(sources, discoveredSources...)
			metrics.DiscoveryMetricsInstance.SetServicesDiscovered(float64(len(discoveredSources)))
		}
	} else {
		c.log.V(1).Info("auto-discovery disabled")
	}
	
	if len(sources) == 0 {
		return nil, fmt.Errorf("no metrics sources configured")
	}
	
	// Collect from all sources
	for _, source := range sources {
		sourceMetrics, err := c.collectFromSource(ctx, source)
		if err != nil {
			c.log.Error(err, "failed to collect from source", "source", source)
			metrics.CollectionMetricsInstance.IncMetricsCollectionTotal(source, "error")
			continue
		}
		
		// Merge metrics
		for tenant, tm := range sourceMetrics {
			if existing, exists := tenantMetrics[tenant]; exists {
				c.mergeMetrics(existing, tm)
			} else {
				tenantMetrics[tenant] = tm
			}
		}
		
		metrics.CollectionMetricsInstance.IncMetricsCollectionTotal(source, "success")
		metrics.CollectionMetricsInstance.SetLastMetricsCollectionTime(source, float64(time.Now().Unix()))
	}
	
	duration := time.Since(startTime).Seconds()
	for _, source := range sources {
		metrics.CollectionMetricsInstance.ObserveMetricsCollectionDuration(source, duration)
	}
	
	c.log.Info("collected metrics", "tenants", len(tenantMetrics), "sources", len(sources), "duration", duration)
	
	return tenantMetrics, nil
}

// collectFromSource collects metrics from a single source
func (c *MimirCollector) collectFromSource(ctx context.Context, source string) (map[string]*TenantMetrics, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add tenant headers for multi-tenant Mimir
	c.addTenantHeaders(req)
	
	// Debug logging: Show the complete request details
	c.log.V(1).Info("making HTTP request for metrics",
		"url", source,
		"method", req.Method,
		"headers", req.Header)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(err, "HTTP request failed",
			"url", source,
			"method", req.Method,
			"headers", req.Header)
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	// Debug logging: Show response details
	c.log.V(1).Info("received HTTP response",
		"url", source,
		"statusCode", resp.StatusCode,
		"status", resp.Status,
		"responseHeaders", resp.Header,
		"contentLength", resp.ContentLength)
	
	if resp.StatusCode != http.StatusOK {
		// Read response body for debugging 404 errors
		bodyBytes, readErr := io.ReadAll(resp.Body)
		bodyContent := "unable to read body"
		if readErr == nil {
			bodyContent = string(bodyBytes)
		}
		
		c.log.Error(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "HTTP request failed with non-200 status",
			"url", source,
			"statusCode", resp.StatusCode,
			"status", resp.Status,
			"responseHeaders", resp.Header,
			"responseBody", bodyContent,
			"requestHeaders", req.Header)
		
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyContent)
	}
	
	// Parse metrics
	parser := expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}
	
	tenantMetrics := make(map[string]*TenantMetrics)
	
	// Process relevant metrics
	for name, mf := range metricFamilies {
		if !c.isRelevantMetric(name) {
			continue
		}
		
		for _, metric := range mf.Metric {
			tenant := c.extractTenant(metric.Label)
			if tenant == "" {
				continue
			}
			
			if _, exists := tenantMetrics[tenant]; !exists {
				tenantMetrics[tenant] = &TenantMetrics{
					Tenant:  tenant,
					Metrics: make(map[string][]MetricData),
				}
			}
			
			value := c.extractValue(metric)
			labels := c.extractLabels(metric.Label)
			
			metricData := MetricData{
				Tenant:     tenant,
				MetricName: name,
				Value:      value,
				Timestamp:  time.Now(),
				Labels:     labels,
				Source:     source,
			}
			
			tenantMetrics[tenant].Metrics[name] = append(tenantMetrics[tenant].Metrics[name], metricData)
			tenantMetrics[tenant].LastUpdate = time.Now()
		}
	}
	
	return tenantMetrics, nil
}

// GetTenantList returns list of all known tenants
func (c *MimirCollector) GetTenantList(ctx context.Context) ([]string, error) {
	// Try metrics-based discovery first
	tenantMetrics, err := c.CollectMetrics(ctx)
	if err == nil {
		tenants := make([]string, 0, len(tenantMetrics))
		for tenant := range tenantMetrics {
			tenants = append(tenants, tenant)
		}
		return tenants, nil
	}
	
	c.log.Info("metrics collection failed, falling back to ConfigMap tenant discovery", "error", err)
	
	// Fallback: discover tenants from Mimir ConfigMap
	return c.getTenantListFromConfigMap(ctx)
}

// getTenantListFromConfigMap discovers tenants from Mimir runtime overrides ConfigMap
func (c *MimirCollector) getTenantListFromConfigMap(ctx context.Context) ([]string, error) {
	// First try predefined fallback tenants from configuration
	if len(c.config.MetricsDiscovery.TenantDiscovery.FallbackTenants) > 0 {
		c.log.Info("using configured fallback tenants", "count", len(c.config.MetricsDiscovery.TenantDiscovery.FallbackTenants))
		return c.config.MetricsDiscovery.TenantDiscovery.FallbackTenants, nil
	}
	
	// Try configured ConfigMap names
	configMapNames := c.config.MetricsDiscovery.TenantDiscovery.ConfigMapNames
	if len(configMapNames) == 0 {
		// Default ConfigMap names to try
		configMapNames = []string{"overrides", "mimir-runtime-overrides", "runtime-config"}
	}
	
	var configMap *corev1.ConfigMap
	var err error
	
	for _, cmName := range configMapNames {
		configMap, err = c.client.CoreV1().ConfigMaps(c.config.Mimir.Namespace).Get(ctx, cmName, metav1.GetOptions{})
		if err == nil {
			c.log.Info("found tenant configuration in ConfigMap", "configMap", cmName)
			break
		}
		c.log.V(1).Info("ConfigMap not found", "name", cmName, "error", err)
	}
	
	if err != nil {
		// Try synthetic tenants if enabled
		if c.config.MetricsDiscovery.TenantDiscovery.EnableSynthetic {
			return c.generateSyntheticTenants(), nil
		}
		return nil, fmt.Errorf("failed to discover tenants: metrics collection failed and ConfigMap fallback failed: %w", err)
	}
	
	// Parse YAML data to extract tenant IDs
	var tenants []string
	
	// Look for overrides.yaml or runtime-config.yaml key
	yamlData := ""
	if data, ok := configMap.Data["overrides.yaml"]; ok {
		yamlData = data
	} else if data, ok := configMap.Data["runtime-config.yaml"]; ok {
		yamlData = data
	} else if data, ok := configMap.Data["config.yaml"]; ok {
		yamlData = data
	}
	
	if yamlData == "" {
		c.log.Info("no YAML configuration found in ConfigMap, using synthetic tenant for testing")
		return []string{"synthetic-tenant-1"}, nil
	}
	
	// Simple parsing to extract tenant IDs
	// Look for patterns like "tenant-1:" or "user-123:" 
	lines := strings.Split(yamlData, "\n")
	tenantSet := make(map[string]bool)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			// This looks like a tenant key
			tenantID := strings.TrimSuffix(line, ":")
			if tenantID != "" && !strings.Contains(tenantID, "overrides") && !strings.Contains(tenantID, "default") {
				tenantSet[tenantID] = true
			}
		}
	}
	
	// Convert to slice
	for tenant := range tenantSet {
		tenants = append(tenants, tenant)
	}
	
	if len(tenants) == 0 {
		c.log.Info("no tenants found in ConfigMap, using synthetic tenant for testing")
		return []string{"synthetic-tenant-1"}, nil
	}
	
	c.log.Info("discovered tenants from ConfigMap", "count", len(tenants), "tenants", tenants)
	return tenants, nil
}

// generateSyntheticTenants creates synthetic tenant IDs for testing
func (c *MimirCollector) generateSyntheticTenants() []string {
	count := c.config.MetricsDiscovery.TenantDiscovery.SyntheticCount
	if count <= 0 {
		count = 3 // Default to 3 synthetic tenants
	}
	
	tenants := make([]string, count)
	for i := 0; i < count; i++ {
		tenants[i] = fmt.Sprintf("synthetic-tenant-%d", i+1)
	}
	
	c.log.Info("generated synthetic tenants", "count", len(tenants), "tenants", tenants)
	return tenants
}

// addTenantHeaders adds tenant-specific headers for multi-tenant Mimir access
func (c *MimirCollector) addTenantHeaders(req *http.Request) {
	// Add primary tenant ID if configured
	if c.config.MetricsDiscovery.TenantDiscovery.MetricsTenantID != "" {
		req.Header.Set("X-Scope-OrgID", c.config.MetricsDiscovery.TenantDiscovery.MetricsTenantID)
		c.log.V(1).Info("added tenant header", "tenant", c.config.MetricsDiscovery.TenantDiscovery.MetricsTenantID)
	} else {
		c.log.V(1).Info("no tenant ID configured - will query without tenant scoping")
	}
	
	// Add any additional custom headers
	for key, value := range c.config.MetricsDiscovery.TenantDiscovery.TenantHeaders {
		req.Header.Set(key, value)
		c.log.V(1).Info("added custom tenant header", "header", key, "value", value)
	}
}

// isRelevantMetric checks if a metric is relevant for limit optimization
func (c *MimirCollector) isRelevantMetric(metricName string) bool {
	relevantMetrics := []string{
		// PRIMARY MIMIR METRICS - Core metrics for limit optimization
		"cortex_distributor_received_samples_total",
		"cortex_distributor_samples_in_total",
		"cortex_ingester_ingested_samples_total",
		"cortex_ingester_ingested_samples_failures_total",
		"cortex_ingester_memory_series",
		"cortex_ingester_memory_users",
		"cortex_query_frontend_queries_total",
		"cortex_query_frontend_query_duration_seconds",
		"cortex_querier_queries_total",
		"cortex_querier_query_duration_seconds",
		
		// EXTENDED MIMIR METRICS - Additional metrics for comprehensive analysis
		"cortex_distributor_deduped_samples_total",
		"cortex_distributor_non_ha_samples_received_total",
		"cortex_distributor_latest_seen_sample_timestamp_seconds",
		"cortex_ingester_chunks_created_total",
		"cortex_ingester_series_removed_total",
		"cortex_querier_series_fetched_total",
		"cortex_querier_chunks_fetched_total",
		"cortex_querier_estimated_series_count",
		"cortex_query_scheduler_queue_duration_seconds",
		"cortex_compactor_runs_total",
		"cortex_ruler_queries_total",
		
		// PROMETHEUS FALLBACK METRICS
		"prometheus_remote_storage_samples_total",
		"prometheus_remote_storage_samples_in_total",
		"prometheus_remote_storage_highest_timestamp_in_seconds",
		"prometheus_tsdb_head_series",
		"prometheus_tsdb_head_samples_appended_total",
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
	
	for _, relevant := range relevantMetrics {
		if strings.Contains(metricName, relevant) {
			return true
		}
	}
	
	// Also allow metrics that contain these key patterns (for dynamic discovery)
	relevantPatterns := []string{
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
	
	for _, pattern := range relevantPatterns {
		if strings.Contains(metricName, pattern) {
			return true
		}
	}
	
	return false
}

// extractTenant extracts tenant ID from metric labels
func (c *MimirCollector) extractTenant(labels []*dto.LabelPair) string {
	for _, label := range labels {
		if label.GetName() == "user" || label.GetName() == "tenant" || label.GetName() == "tenant_id" {
			return label.GetValue()
		}
	}
	return ""
}

// extractValue extracts the numeric value from a metric
func (c *MimirCollector) extractValue(metric *dto.Metric) float64 {
	if metric.Counter != nil {
		return metric.Counter.GetValue()
	}
	if metric.Gauge != nil {
		return metric.Gauge.GetValue()
	}
	if metric.Histogram != nil {
		return float64(metric.Histogram.GetSampleCount())
	}
	if metric.Summary != nil {
		return float64(metric.Summary.GetSampleCount())
	}
	return 0
}

// extractLabels extracts all labels from a metric
func (c *MimirCollector) extractLabels(labels []*dto.LabelPair) map[string]string {
	result := make(map[string]string)
	for _, label := range labels {
		result[label.GetName()] = label.GetValue()
	}
	return result
}

// mergeMetrics merges two TenantMetrics objects
func (c *MimirCollector) mergeMetrics(existing, new *TenantMetrics) {
	for metricName, data := range new.Metrics {
		existing.Metrics[metricName] = append(existing.Metrics[metricName], data...)
	}
	if new.LastUpdate.After(existing.LastUpdate) {
		existing.LastUpdate = new.LastUpdate
	}
}

// SyntheticCollector implements synthetic metrics for testing
type SyntheticCollector struct {
	config *config.Config
	log    logr.Logger
}

// NewSyntheticCollector creates a new synthetic collector
func NewSyntheticCollector(cfg *config.Config, log logr.Logger) *SyntheticCollector {
	return &SyntheticCollector{
		config: cfg,
		log:    log,
	}
}

// CollectMetrics generates synthetic metrics for testing
func (s *SyntheticCollector) CollectMetrics(ctx context.Context) (map[string]*TenantMetrics, error) {
	tenantMetrics := make(map[string]*TenantMetrics)
	
	for i := 0; i < s.config.Synthetic.TenantCount; i++ {
		tenant := fmt.Sprintf("synthetic-tenant-%d", i)
		
		tm := &TenantMetrics{
			Tenant:     tenant,
			Metrics:    make(map[string][]MetricData),
			LastUpdate: time.Now(),
		}
		
		// Generate synthetic ingestion metrics
		tm.Metrics["cortex_distributor_received_samples_total"] = []MetricData{
			{
				Tenant:     tenant,
				MetricName: "cortex_distributor_received_samples_total",
				Value:      float64(1000 + i*500), // Varying sample rates
				Timestamp:  time.Now(),
				Labels:     map[string]string{"user": tenant},
				Source:     "synthetic",
			},
		}
		
		// Generate synthetic series metrics
		tm.Metrics["cortex_ingester_memory_series"] = []MetricData{
			{
				Tenant:     tenant,
				MetricName: "cortex_ingester_memory_series",
				Value:      float64(10000 + i*2000), // Varying series counts
				Timestamp:  time.Now(),
				Labels:     map[string]string{"user": tenant},
				Source:     "synthetic",
			},
		}
		
		tenantMetrics[tenant] = tm
	}
	
	s.log.Info("generated synthetic metrics", "tenants", len(tenantMetrics))
	
	return tenantMetrics, nil
}

// GetTenantList returns synthetic tenant list
func (s *SyntheticCollector) GetTenantList(ctx context.Context) ([]string, error) {
	tenants := make([]string, s.config.Synthetic.TenantCount)
	for i := 0; i < s.config.Synthetic.TenantCount; i++ {
		tenants[i] = fmt.Sprintf("synthetic-tenant-%d", i)
	}
	return tenants, nil
}

// NewCollector creates the appropriate collector based on configuration
func NewCollector(cfg *config.Config, client kubernetes.Interface, log logr.Logger) Collector {
	if cfg.Synthetic.Enabled {
		return NewSyntheticCollector(cfg, log.WithName("synthetic"))
	}
	return NewMimirCollector(cfg, client, log.WithName("mimir"))
} 