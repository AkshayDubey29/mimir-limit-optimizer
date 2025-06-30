package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// Controller metrics
	reconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_reconcile_total",
			Help: "Total number of reconciliations performed",
		},
		[]string{"result"},
	)

	reconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_reconcile_duration_seconds",
			Help:    "Time spent on reconciliations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	lastReconcileTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_last_reconcile_timestamp",
			Help: "Timestamp of the last successful reconciliation",
		},
	)

	// Tenant metrics
	tenantsMonitored = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_tenants_monitored_total",
			Help: "Number of tenants currently being monitored",
		},
	)

	tenantsSkipped = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_tenants_skipped_total",
			Help: "Number of tenants skipped due to filtering",
		},
	)

	tenantLimitsUpdated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_tenant_limits_updated_total",
			Help: "Total number of tenant limit updates",
		},
		[]string{"tenant", "reason"},
	)

	tenantCurrentLimits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_tenant_current_limits",
			Help: "Current limits for each tenant",
		},
		[]string{"tenant", "limit_type"},
	)

	tenantRecommendedLimits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_tenant_recommended_limits",
			Help: "Recommended limits for each tenant (dry-run mode)",
		},
		[]string{"tenant", "limit_type"},
	)

	// Metrics collection
	metricsCollectionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_metrics_collection_total",
			Help: "Total number of metrics collection attempts",
		},
		[]string{"source", "result"},
	)

	metricsCollectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_metrics_collection_duration_seconds",
			Help:    "Time spent collecting metrics",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"source"},
	)

	lastMetricsCollectionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_last_metrics_collection_timestamp",
			Help: "Timestamp of the last successful metrics collection",
		},
		[]string{"source"},
	)

	// Spike detection metrics
	spikesDetected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_spikes_detected_total",
			Help: "Total number of usage spikes detected",
		},
		[]string{"tenant", "metric_type"},
	)

	spikeMultiplier = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_spike_multiplier",
			Help: "Current spike multiplier applied to tenant limits",
		},
		[]string{"tenant"},
	)

	// ConfigMap operations
	configMapUpdates = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_configmap_updates_total",
			Help: "Total number of ConfigMap update attempts",
		},
		[]string{"result"},
	)

	configMapUpdateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_configmap_update_duration_seconds",
			Help:    "Time spent updating ConfigMaps",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	lastConfigMapUpdate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_last_configmap_update_timestamp",
			Help: "Timestamp of the last ConfigMap update",
		},
	)

	// Health and error metrics
	healthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_health_status",
			Help: "Health status of the controller (1=healthy, 0=unhealthy)",
		},
		[]string{"component"},
	)

	errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_errors_total",
			Help: "Total number of errors encountered",
		},
		[]string{"component", "error_type"},
	)

	// Trend analysis metrics
	trendAnalysisDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_trend_analysis_duration_seconds",
			Help:    "Time spent analyzing trends",
			Buckets: prometheus.DefBuckets,
		},
	)

	tenantUsagePercentile = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_tenant_usage_percentile",
			Help: "Usage percentile for each tenant",
		},
		[]string{"tenant", "metric_type", "percentile"},
	)

	// Discovery metrics
	servicesDiscovered = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_services_discovered_total",
			Help: "Number of services discovered for metrics collection",
		},
	)

	discoveryErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_discovery_errors_total",
			Help: "Total number of service discovery errors",
		},
	)

	// Cost Control metrics
	costCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_cost_current",
			Help: "Current cost for each tenant",
		},
		[]string{"tenant", "cost_type"},
	)

	budgetUsageRatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_budget_usage_ratio",
			Help: "Budget usage ratio (0.0-1.0) for each tenant",
		},
		[]string{"tenant"},
	)

	costRecommendationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_cost_recommendations_total",
			Help: "Total number of cost optimization recommendations generated",
		},
		[]string{"tenant", "recommendation_type"},
	)

	budgetViolationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_budget_violations_total",
			Help: "Total number of budget violations detected",
		},
		[]string{"tenant", "violation_level"},
	)

	// Circuit Breaker metrics
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"tenant"},
	)

	rateLimitRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_rate_limit_requests_total",
			Help: "Total number of rate limit requests",
		},
		[]string{"tenant", "result"},
	)

	blastDetectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_blast_detections_total",
			Help: "Total number of blast detections",
		},
		[]string{"tenant", "blast_type"},
	)

	throttledRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_throttled_requests_total",
			Help: "Total number of throttled requests",
		},
		[]string{"tenant", "throttle_reason"},
	)

	// Emergency System metrics
	panicModeActivationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_panic_mode_activations_total",
			Help: "Total number of panic mode activations",
		},
		[]string{"trigger_reason"},
	)

	emergencyActionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_emergency_actions_total",
			Help: "Total number of emergency actions taken",
		},
		[]string{"action_type", "result"},
	)

	recoveryAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_recovery_attempts_total",
			Help: "Total number of recovery attempts",
		},
		[]string{"recovery_type", "result"},
	)

	resourceUsagePercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_resource_usage_percent",
			Help: "Current resource usage percentage",
		},
		[]string{"resource_type"},
	)

	// Alerting Resilience metrics
	alertDeliveryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_alert_delivery_total",
			Help: "Total number of alert delivery attempts",
		},
		[]string{"channel", "alert_type", "result"},
	)

	alertDeliveryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_alert_delivery_duration_seconds",
			Help:    "Time spent delivering alerts to each channel",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"channel"},
	)

	alertChannelHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_alert_channel_health",
			Help: "Health status of alert channels (1=healthy, 0=unhealthy)",
		},
		[]string{"channel"},
	)

	alertChannelErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_alert_channel_errors_total",
			Help: "Total number of alert channel errors",
		},
		[]string{"channel", "error_type"},
	)

	alertRetryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_alert_retry_attempts_total",
			Help: "Total number of alert retry attempts",
		},
		[]string{"channel", "alert_type"},
	)

	alertChannelCircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_alert_channel_circuit_breaker_state",
			Help: "Circuit breaker state for alert channels (0=closed, 1=open, 2=half-open)",
		},
		[]string{"channel"},
	)

	alertQueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_alert_queue_size",
			Help: "Current size of alert queue for each channel",
		},
		[]string{"channel"},
	)

	alertConfigurationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mimir_limit_optimizer_alert_configuration_errors_total",
			Help: "Total number of alert configuration errors",
		},
		[]string{"channel", "config_error_type"},
	)

	lastSuccessfulAlertTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mimir_limit_optimizer_last_successful_alert_timestamp",
			Help: "Timestamp of the last successful alert delivery",
		},
		[]string{"channel"},
	)

	alertChannelResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mimir_limit_optimizer_alert_channel_response_time_seconds",
			Help:    "Response time from alert channels",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"channel"},
	)
)

// RegisterMetrics registers all metrics with the controller-runtime metrics registry
func RegisterMetrics() error {
	metrics.Registry.MustRegister(
		// Controller metrics
		reconcileTotal,
		reconcileDuration,
		lastReconcileTime,
		
		// Tenant metrics
		tenantsMonitored,
		tenantsSkipped,
		tenantLimitsUpdated,
		tenantCurrentLimits,
		tenantRecommendedLimits,
		tenantUsagePercentile,
		
		// Collection metrics
		metricsCollectionTotal,
		metricsCollectionDuration,
		lastMetricsCollectionTime,
		
		// Spike detection metrics
		spikesDetected,
		spikeMultiplier,
		
		// ConfigMap metrics
		configMapUpdates,
		configMapUpdateDuration,
		lastConfigMapUpdate,
		
		// Health metrics
		healthStatus,
		errorTotal,
		
		// Trend analysis metrics
		trendAnalysisDuration,
		
		// Discovery metrics
		servicesDiscovered,
		discoveryErrors,
		
		// Cost Control metrics
		costCurrent,
		budgetUsageRatio,
		costRecommendationsTotal,
		budgetViolationsTotal,
		
		// Circuit Breaker metrics
		circuitBreakerState,
		rateLimitRequestsTotal,
		blastDetectionsTotal,
		throttledRequestsTotal,
		
		// Emergency System metrics
		panicModeActivationsTotal,
		emergencyActionsTotal,
		recoveryAttemptsTotal,
		resourceUsagePercent,
		
		// Alerting Resilience metrics
		alertDeliveryTotal,
		alertDeliveryDuration,
		alertChannelHealth,
		alertChannelErrors,
		alertRetryAttempts,
		alertChannelCircuitBreakerState,
		alertQueueSize,
		alertConfigurationErrors,
		lastSuccessfulAlertTime,
		alertChannelResponseTime,
	)
	return nil
}

// ReconcileMetrics provides access to reconciliation metrics
type ReconcileMetrics struct{}

func (r *ReconcileMetrics) IncReconcileTotal(result string) {
	reconcileTotal.WithLabelValues(result).Inc()
}

func (r *ReconcileMetrics) ObserveReconcileDuration(result string, duration float64) {
	reconcileDuration.WithLabelValues(result).Observe(duration)
}

func (r *ReconcileMetrics) SetLastReconcileTime(timestamp float64) {
	lastReconcileTime.Set(timestamp)
}

// TenantMetrics provides access to tenant-related metrics
type TenantMetrics struct{}

func (t *TenantMetrics) SetTenantsMonitored(count float64) {
	tenantsMonitored.Set(count)
}

func (t *TenantMetrics) SetTenantsSkipped(count float64) {
	tenantsSkipped.Set(count)
}

func (t *TenantMetrics) IncTenantLimitsUpdated(tenant, reason string) {
	tenantLimitsUpdated.WithLabelValues(tenant, reason).Inc()
}

func (t *TenantMetrics) SetTenantCurrentLimits(tenant, limitType string, value float64) {
	tenantCurrentLimits.WithLabelValues(tenant, limitType).Set(value)
}

func (t *TenantMetrics) SetTenantRecommendedLimits(tenant, limitType string, value float64) {
	tenantRecommendedLimits.WithLabelValues(tenant, limitType).Set(value)
}

func (t *TenantMetrics) SetTenantUsagePercentile(tenant, metricType, percentile string, value float64) {
	tenantUsagePercentile.WithLabelValues(tenant, metricType, percentile).Set(value)
}

// CollectionMetrics provides access to metrics collection metrics
type CollectionMetrics struct{}

func (c *CollectionMetrics) IncMetricsCollectionTotal(source, result string) {
	metricsCollectionTotal.WithLabelValues(source, result).Inc()
}

func (c *CollectionMetrics) ObserveMetricsCollectionDuration(source string, duration float64) {
	metricsCollectionDuration.WithLabelValues(source).Observe(duration)
}

func (c *CollectionMetrics) SetLastMetricsCollectionTime(source string, timestamp float64) {
	lastMetricsCollectionTime.WithLabelValues(source).Set(timestamp)
}

// SpikeMetrics provides access to spike detection metrics
type SpikeMetrics struct{}

func (s *SpikeMetrics) IncSpikesDetected(tenant, metricType string) {
	spikesDetected.WithLabelValues(tenant, metricType).Inc()
}

func (s *SpikeMetrics) SetSpikeMultiplier(tenant string, multiplier float64) {
	spikeMultiplier.WithLabelValues(tenant).Set(multiplier)
}

// ConfigMapMetrics provides access to ConfigMap operation metrics
type ConfigMapMetrics struct{}

func (c *ConfigMapMetrics) IncConfigMapUpdates(result string) {
	configMapUpdates.WithLabelValues(result).Inc()
}

func (c *ConfigMapMetrics) ObserveConfigMapUpdateDuration(result string, duration float64) {
	configMapUpdateDuration.WithLabelValues(result).Observe(duration)
}

func (c *ConfigMapMetrics) SetLastConfigMapUpdate(timestamp float64) {
	lastConfigMapUpdate.Set(timestamp)
}

// HealthMetrics provides access to health and error metrics
type HealthMetrics struct{}

func (h *HealthMetrics) SetHealthStatus(component string, status float64) {
	healthStatus.WithLabelValues(component).Set(status)
}

func (h *HealthMetrics) IncErrorTotal(component, errorType string) {
	errorTotal.WithLabelValues(component, errorType).Inc()
}

// TrendMetrics provides access to trend analysis metrics
type TrendMetrics struct{}

func (t *TrendMetrics) ObserveTrendAnalysisDuration(duration float64) {
	trendAnalysisDuration.Observe(duration)
}

// DiscoveryMetrics provides access to service discovery metrics
type DiscoveryMetrics struct{}

func (d *DiscoveryMetrics) SetServicesDiscovered(count float64) {
	servicesDiscovered.Set(count)
}

func (d *DiscoveryMetrics) IncDiscoveryErrors() {
	discoveryErrors.Inc()
}

// CostControlMetrics provides access to cost control metrics
type CostControlMetrics struct{}

func (c *CostControlMetrics) SetCostCurrent(tenant, costType string, cost float64) {
	costCurrent.WithLabelValues(tenant, costType).Set(cost)
}

func (c *CostControlMetrics) SetBudgetUsageRatio(tenant string, ratio float64) {
	budgetUsageRatio.WithLabelValues(tenant).Set(ratio)
}

func (c *CostControlMetrics) IncCostRecommendations(tenant, recommendationType string) {
	costRecommendationsTotal.WithLabelValues(tenant, recommendationType).Inc()
}

func (c *CostControlMetrics) IncBudgetViolations(tenant, violationLevel string) {
	budgetViolationsTotal.WithLabelValues(tenant, violationLevel).Inc()
}

// CircuitBreakerMetrics provides access to circuit breaker metrics
type CircuitBreakerMetrics struct{}

func (c *CircuitBreakerMetrics) SetCircuitBreakerState(tenant string, state float64) {
	circuitBreakerState.WithLabelValues(tenant).Set(state)
}

func (c *CircuitBreakerMetrics) IncRateLimitRequests(tenant, result string) {
	rateLimitRequestsTotal.WithLabelValues(tenant, result).Inc()
}

func (c *CircuitBreakerMetrics) IncBlastDetections(tenant, blastType string) {
	blastDetectionsTotal.WithLabelValues(tenant, blastType).Inc()
}

func (c *CircuitBreakerMetrics) IncThrottledRequests(tenant, reason string) {
	throttledRequestsTotal.WithLabelValues(tenant, reason).Inc()
}

// EmergencyMetrics provides access to emergency system metrics
type EmergencyMetrics struct{}

func (e *EmergencyMetrics) IncPanicModeActivations(reason string) {
	panicModeActivationsTotal.WithLabelValues(reason).Inc()
}

func (e *EmergencyMetrics) IncEmergencyActions(actionType, result string) {
	emergencyActionsTotal.WithLabelValues(actionType, result).Inc()
}

func (e *EmergencyMetrics) IncRecoveryAttempts(recoveryType, result string) {
	recoveryAttemptsTotal.WithLabelValues(recoveryType, result).Inc()
}

func (e *EmergencyMetrics) SetResourceUsage(resourceType string, percentage float64) {
	resourceUsagePercent.WithLabelValues(resourceType).Set(percentage)
}

// AlertingMetrics provides access to alerting resilience metrics
type AlertingMetrics struct{}

func (a *AlertingMetrics) IncAlertDeliveryTotal(channel, alertType, result string) {
	alertDeliveryTotal.WithLabelValues(channel, alertType, result).Inc()
}

func (a *AlertingMetrics) ObserveAlertDeliveryDuration(channel string, duration float64) {
	alertDeliveryDuration.WithLabelValues(channel).Observe(duration)
}

func (a *AlertingMetrics) SetAlertChannelHealth(channel string, health float64) {
	alertChannelHealth.WithLabelValues(channel).Set(health)
}

func (a *AlertingMetrics) IncAlertChannelErrors(channel, errorType string) {
	alertChannelErrors.WithLabelValues(channel, errorType).Inc()
}

func (a *AlertingMetrics) IncAlertRetryAttempts(channel, alertType string) {
	alertRetryAttempts.WithLabelValues(channel, alertType).Inc()
}

func (a *AlertingMetrics) SetAlertChannelCircuitBreakerState(channel string, state float64) {
	alertChannelCircuitBreakerState.WithLabelValues(channel).Set(state)
}

func (a *AlertingMetrics) SetAlertQueueSize(channel string, size float64) {
	alertQueueSize.WithLabelValues(channel).Set(size)
}

func (a *AlertingMetrics) IncAlertConfigurationErrors(channel, errorType string) {
	alertConfigurationErrors.WithLabelValues(channel, errorType).Inc()
}

func (a *AlertingMetrics) SetLastSuccessfulAlertTime(channel string, timestamp float64) {
	lastSuccessfulAlertTime.WithLabelValues(channel).Set(timestamp)
}

func (a *AlertingMetrics) ObserveAlertChannelResponseTime(channel string, duration float64) {
	alertChannelResponseTime.WithLabelValues(channel).Observe(duration)
}

// Global metric instances
var (
	ReconcileMetricsInstance     = &ReconcileMetrics{}
	TenantMetricsInstance        = &TenantMetrics{}
	CollectionMetricsInstance    = &CollectionMetrics{}
	SpikeMetricsInstance         = &SpikeMetrics{}
	ConfigMapMetricsInstance     = &ConfigMapMetrics{}
	HealthMetricsInstance        = &HealthMetrics{}
	TrendMetricsInstance         = &TrendMetrics{}
	DiscoveryMetricsInstance     = &DiscoveryMetrics{}
	CostControlMetricsInstance   = &CostControlMetrics{}
	CircuitBreakerMetricsInstance = &CircuitBreakerMetrics{}
	EmergencyMetricsInstance     = &EmergencyMetrics{}
	AlertingMetricsInstance      = &AlertingMetrics{}
) 