package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-logr/logr"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"k8s.io/client-go/kubernetes"

	"github.com/tapasyadubey/mimir-limit-optimizer/internal/config"
	"github.com/tapasyadubey/mimir-limit-optimizer/internal/discovery"
	"github.com/tapasyadubey/mimir-limit-optimizer/internal/metrics"
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
		discoveredSources, err := c.discovery.DiscoverMetricsEndpoints(ctx)
		if err != nil {
			c.log.Error(err, "failed to discover metrics endpoints")
			metrics.DiscoveryMetricsInstance.IncDiscoveryErrors()
		} else {
			sources = append(sources, discoveredSources...)
			metrics.DiscoveryMetricsInstance.SetServicesDiscovered(float64(len(discoveredSources)))
		}
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
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	// Parse metrics
	parser := expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricfamilies(resp.Body)
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
	tenantMetrics, err := c.CollectMetrics(ctx)
	if err != nil {
		return nil, err
	}
	
	tenants := make([]string, 0, len(tenantMetrics))
	for tenant := range tenantMetrics {
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

// isRelevantMetric checks if a metric is relevant for limit optimization
func (c *MimirCollector) isRelevantMetric(metricName string) bool {
	relevantMetrics := []string{
		"cortex_ingester_ingested_samples_total",
		"cortex_ingester_ingested_samples_failures_total",
		"cortex_ingester_memory_series",
		"cortex_ingester_memory_users",
		"cortex_distributor_received_samples_total",
		"cortex_distributor_samples_in_total",
		"cortex_query_frontend_queries_total",
		"cortex_query_frontend_query_duration_seconds",
		"cortex_querier_queries_total",
		"cortex_querier_query_duration_seconds",
		"prometheus_remote_storage_samples_total",
		"prometheus_remote_storage_highest_timestamp_in_seconds",
		"prometheus_tsdb_head_series",
		"prometheus_tsdb_head_samples_appended_total",
		// Add more relevant metrics as needed
	}
	
	for _, relevant := range relevantMetrics {
		if strings.Contains(metricName, relevant) {
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