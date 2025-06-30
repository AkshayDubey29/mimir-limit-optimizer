package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/analyzer"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/collector"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// BlastProtector implements circuit breaker pattern with blast protection
type BlastProtector struct {
	config         *config.Config
	log            logr.Logger
	mu             sync.RWMutex
	state          CircuitBreakerState
	failures       int
	requests       int
	lastStateChange time.Time
	halfOpenRequests int
	
	// Rate limiting
	rateLimiters map[string]*TenantRateLimiter
	
	// Blast detection
	blastDetector  *BlastDetector
	emergencyMode  bool
	panicMode      bool
	
	// Auto-configuration
	autoConfig     *AutoConfig
	lastAdaptation time.Time
	initialized    bool
}

// AutoConfig holds dynamic configuration based on real-time metrics
type AutoConfig struct {
	mu                   sync.RWMutex
	tenantThresholds     map[string]*TenantThresholds
	currentLimits        map[string]*analyzer.TenantLimits
	lastUpdate           time.Time
	observationStartTime time.Time
}

// TenantThresholds holds calculated thresholds for a tenant
type TenantThresholds struct {
	IngestionThreshold float64
	QueryThreshold     float64
	SeriesThreshold    float64
	BurstThreshold     float64
	LastCalculated     time.Time
	BasedOnLimits      *analyzer.TenantLimits
	SafetyMargin       float64
}

// TenantRateLimiter implements per-tenant rate limiting
type TenantRateLimiter struct {
	tenant        string
	tokens        float64
	lastUpdate    time.Time
	requestsPerSec float64
	burstCapacity  int
	mu            sync.Mutex
}

// BlastDetector monitors for sudden traffic spikes
type BlastDetector struct {
	config          *config.Config
	log             logr.Logger
	mu              sync.RWMutex
	metrics         map[string]*BlastMetrics
	alertSent       map[string]time.Time
}

// BlastMetrics tracks metrics for blast detection
type BlastMetrics struct {
	IngestionRate  float64
	QueryRate      float64
	SeriesRate     float64
	ErrorRate      float64
	LastUpdate     time.Time
	BaselineRates  BaselineRates
}

// BaselineRates stores normal operating rates for comparison
type BaselineRates struct {
	IngestionRate float64
	QueryRate     float64
	SeriesRate    float64
	ErrorRate     float64
	LastCalculated time.Time
}

// ProtectionAction represents actions to take during protection
type ProtectionAction struct {
	Action    string
	Tenant    string
	Reason    string
	Severity  string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// NewBlastProtector creates a new circuit breaker with blast protection
func NewBlastProtector(cfg *config.Config, log logr.Logger) *BlastProtector {
	bp := &BlastProtector{
		config:          cfg,
		log:             log,
		state:           StateClosed,
		rateLimiters:    make(map[string]*TenantRateLimiter),
		lastStateChange: time.Now(),
		blastDetector: &BlastDetector{
			config:    cfg,
			log:       log,
			metrics:   make(map[string]*BlastMetrics),
			alertSent: make(map[string]time.Time),
		},
		autoConfig: &AutoConfig{
			tenantThresholds:     make(map[string]*TenantThresholds),
			currentLimits:        make(map[string]*analyzer.TenantLimits),
			observationStartTime: time.Now(),
		},
		initialized: false,
	}
	
	return bp
}

// ProcessMetrics processes incoming metrics and applies protection
func (bp *BlastProtector) ProcessMetrics(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string]*collector.TenantMetrics, error) {
	if !bp.config.CircuitBreaker.Enabled || !bp.config.CircuitBreaker.RuntimeEnabled {
		return tenantMetrics, nil
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Initialize if not done yet
	if !bp.initialized {
		bp.initializeAutoConfig()
	}

	// Update auto-configuration if enabled
	if bp.config.CircuitBreaker.AutoConfig.Enabled {
		bp.updateAutoConfiguration(tenantMetrics)
	}

	// Update blast detection metrics
	bp.blastDetector.updateMetrics(tenantMetrics)

	// Check for blast conditions using auto-calculated or manual thresholds
	blastDetected := bp.blastDetector.detectBlast(tenantMetrics)
	if blastDetected {
		bp.handleBlastDetection(ctx)
	}

	// Apply rate limiting
	filteredMetrics := bp.applyRateLimiting(tenantMetrics)

	// Update circuit breaker state
	bp.updateCircuitBreakerState()

	return filteredMetrics, nil
}

// ApplyProtection applies protection measures to tenant limits
func (bp *BlastProtector) ApplyProtection(ctx context.Context, limits map[string]*analyzer.TenantLimits) (map[string]*analyzer.TenantLimits, error) {
	if !bp.config.CircuitBreaker.Enabled {
		return limits, nil
	}

	bp.mu.RLock()
	defer bp.mu.RUnlock()

	protectedLimits := make(map[string]*analyzer.TenantLimits)

	for tenant, limit := range limits {
		adjustedLimit := bp.adjustLimitsBasedOnState(tenant, limit)
		protectedLimits[tenant] = adjustedLimit
	}

	return protectedLimits, nil
}

// EnterEmergencyMode puts the system into emergency protection mode
func (bp *BlastProtector) EnterEmergencyMode(reason string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.emergencyMode {
		return
	}

	bp.emergencyMode = true
	bp.state = StateOpen
	bp.lastStateChange = time.Now()

	bp.log.Error(fmt.Errorf("emergency mode activated: %s", reason), "EMERGENCY MODE ACTIVATED", "reason", reason)

	// Send emergency alerts
	bp.sendEmergencyAlert("emergency_mode_activated", reason, "critical")

	// Execute emergency actions
	bp.executeEmergencyActions()
}

// EnterPanicMode puts the system into panic mode for extreme situations
func (bp *BlastProtector) EnterPanicMode(reason string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.panicMode {
		return
	}

	bp.panicMode = true
	bp.emergencyMode = true
	bp.state = StateOpen
	bp.lastStateChange = time.Now()

	bp.log.Error(fmt.Errorf("panic mode activated: %s", reason), "PANIC MODE ACTIVATED", "reason", reason)

	// Send panic alerts
	bp.sendEmergencyAlert("panic_mode_activated", reason, "critical")

	// Execute panic actions
	bp.executePanicActions()
}

// ExitEmergencyMode exits emergency mode with recovery procedures
func (bp *BlastProtector) ExitEmergencyMode() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if !bp.emergencyMode {
		return nil
	}

	// Check if conditions allow recovery
	if !bp.isRecoveryPossible() {
		return fmt.Errorf("recovery conditions not met")
	}

	bp.emergencyMode = false
	bp.panicMode = false
	bp.state = StateHalfOpen
	bp.lastStateChange = time.Now()
	bp.halfOpenRequests = 0

	bp.log.Info("exiting emergency mode, entering recovery phase")

	return nil
}

// GetProtectionStatus returns current protection status
func (bp *BlastProtector) GetProtectionStatus() map[string]interface{} {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	return map[string]interface{}{
		"circuit_breaker_state": bp.state.String(),
		"emergency_mode":        bp.emergencyMode,
		"panic_mode":            bp.panicMode,
		"failures":              bp.failures,
		"requests":              bp.requests,
		"last_state_change":     bp.lastStateChange,
		"half_open_requests":    bp.halfOpenRequests,
		"active_rate_limiters":  len(bp.rateLimiters),
	}
}

// Private methods

func (bp *BlastProtector) handleBlastDetection(ctx context.Context) {
	bp.log.Info("blast detected, applying protection measures")

	if bp.config.CircuitBreaker.BlastProtection.AutoEmergencyShutdown {
		bp.EnterEmergencyMode("blast_detected")
	} else {
		// Apply gradual protection
		bp.failures += 10 // Increase failure count to trigger circuit breaker
		bp.log.Info("applied gradual protection due to blast detection")
	}
}

func (bp *BlastProtector) applyRateLimiting(tenantMetrics map[string]*collector.TenantMetrics) map[string]*collector.TenantMetrics {
	if !bp.config.CircuitBreaker.RateLimit.Enabled {
		return tenantMetrics
	}

	filteredMetrics := make(map[string]*collector.TenantMetrics)

	for tenant, metrics := range tenantMetrics {
		rateLimiter := bp.getRateLimiter(tenant)
		
		if rateLimiter.allowRequest() {
			filteredMetrics[tenant] = metrics
		} else {
			// Rate limited - reduce metrics or block
			bp.log.V(1).Info("rate limited tenant", "tenant", tenant)
			// Could implement partial metrics reduction here
		}
	}

	return filteredMetrics
}

func (bp *BlastProtector) getRateLimiter(tenant string) *TenantRateLimiter {
	if limiter, exists := bp.rateLimiters[tenant]; exists {
		return limiter
	}

	// Create new rate limiter for tenant
	limiter := &TenantRateLimiter{
		tenant:         tenant,
		tokens:         float64(bp.config.CircuitBreaker.RateLimit.BurstCapacity),
		lastUpdate:     time.Now(),
		requestsPerSec: bp.config.CircuitBreaker.RateLimit.RequestsPerSecond,
		burstCapacity:  bp.config.CircuitBreaker.RateLimit.BurstCapacity,
	}

	bp.rateLimiters[tenant] = limiter
	return limiter
}

func (trl *TenantRateLimiter) allowRequest() bool {
	trl.mu.Lock()
	defer trl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(trl.lastUpdate).Seconds()

	// Add tokens based on rate
	trl.tokens += elapsed * trl.requestsPerSec
	if trl.tokens > float64(trl.burstCapacity) {
		trl.tokens = float64(trl.burstCapacity)
	}

	trl.lastUpdate = now

	if trl.tokens >= 1.0 {
		trl.tokens -= 1.0
		return true
	}

	return false
}

func (bp *BlastProtector) updateCircuitBreakerState() {
	now := time.Now()

	switch bp.state {
	case StateClosed:
		if bp.shouldOpenCircuit() {
			bp.state = StateOpen
			bp.lastStateChange = now
			bp.log.Info("circuit breaker opened", "failures", bp.failures, "requests", bp.requests)
		}

	case StateOpen:
		if now.Sub(bp.lastStateChange) >= bp.config.CircuitBreaker.SleepWindow {
			bp.state = StateHalfOpen
			bp.halfOpenRequests = 0
			bp.lastStateChange = now
			bp.log.Info("circuit breaker entered half-open state")
		}

	case StateHalfOpen:
		if bp.halfOpenRequests >= bp.config.CircuitBreaker.MaxRequestsInHalfOpen {
			if bp.failures == 0 {
				bp.state = StateClosed
				bp.failures = 0
				bp.requests = 0
				bp.log.Info("circuit breaker closed - recovery successful")
			} else {
				bp.state = StateOpen
				bp.log.Info("circuit breaker reopened - recovery failed")
			}
			bp.lastStateChange = now
		}
	}
}

func (bp *BlastProtector) shouldOpenCircuit() bool {
	if bp.requests < bp.config.CircuitBreaker.RequestVolumeThreshold {
		return false
	}

	failureRate := float64(bp.failures) / float64(bp.requests) * 100
	return failureRate >= bp.config.CircuitBreaker.FailureThreshold
}

func (bp *BlastProtector) adjustLimitsBasedOnState(tenant string, limit *analyzer.TenantLimits) *analyzer.TenantLimits {
	adjustedLimit := &analyzer.TenantLimits{
		Tenant:      limit.Tenant,
		Limits:      make(map[string]interface{}),
		LastUpdated: time.Now(),
		Reason:      limit.Reason,
		Source:      "circuit-breaker",
	}

	// Copy all limits and apply reduction factors based on state
	reductionFactor := 1.0
	var reason string

	switch bp.state {
	case StateOpen:
		// Drastically reduce limits during open state
		reductionFactor = 0.1
		reason = "circuit_breaker_open"

	case StateHalfOpen:
		// Moderately reduce limits during half-open state
		reductionFactor = 0.5
		reason = "circuit_breaker_half_open"

	case StateClosed:
		// Normal operation
		if bp.emergencyMode {
			// Still in emergency mode but circuit is closed
			reductionFactor = 0.8
			reason = "emergency_mode"
		}
	}

	if bp.panicMode {
		// Extreme reduction in panic mode
		reductionFactor = 0.05
		reason = "panic_mode"
	}

	// Apply reduction to all dynamic limits
	for limitName, limitValue := range limit.Limits {
		adjustedValue := limitValue
		
		if reductionFactor < 1.0 {
			switch v := limitValue.(type) {
			case float64:
				adjustedValue = v * reductionFactor
			case int64:
				adjustedValue = int64(float64(v) * reductionFactor)
			}
		}
		
		adjustedLimit.Limits[limitName] = adjustedValue
	}

	if reason != "" {
		adjustedLimit.Reason = reason
	}

	return adjustedLimit
}

func (bp *BlastProtector) executeEmergencyActions() {
	actions := bp.config.Emergency.PanicMode.Actions

	for _, action := range actions {
		switch action {
		case "reduce_limits":
			bp.log.Info("executing emergency action: reduce_limits")
			// This is handled in adjustLimitsBasedOnState
		case "throttle_ingestion":
			bp.log.Info("executing emergency action: throttle_ingestion")
			// Additional throttling logic could go here
		case "alert":
			bp.log.Info("executing emergency action: alert")
			bp.sendEmergencyAlert("emergency_action_executed", action, "high")
		}
	}
}

func (bp *BlastProtector) executePanicActions() {
	// Execute even more aggressive actions in panic mode
	bp.log.Error(fmt.Errorf("executing panic mode actions"), "executing panic mode actions")
	
	// Could implement additional panic actions like:
	// - Temporary service degradation
	// - Emergency contact notification
	// - Automated rollbacks
}

func (bp *BlastProtector) isRecoveryPossible() bool {
	// Check if system metrics are back to normal
	// This would integrate with system monitoring
	return true // Simplified for now
}

func (bp *BlastProtector) sendEmergencyAlert(alertType, reason, severity string) {
	// TODO: Integrate with alerting system
	bp.log.Error(fmt.Errorf("emergency alert: %s - %s", alertType, reason), "EMERGENCY ALERT", 
		"type", alertType,
		"reason", reason,
		"severity", severity,
		"timestamp", time.Now())
}

// BlastDetector methods

func (bd *BlastDetector) updateMetrics(tenantMetrics map[string]*collector.TenantMetrics) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	for tenant, metrics := range tenantMetrics {
		blastMetrics := bd.getOrCreateBlastMetrics(tenant)
		
		// Update current rates
		blastMetrics.IngestionRate = bd.calculateRate(metrics, "cortex_distributor_received_samples_total")
		blastMetrics.QueryRate = bd.calculateRate(metrics, "cortex_querier_queries_total")
		blastMetrics.SeriesRate = bd.calculateRate(metrics, "cortex_ingester_memory_series")
		blastMetrics.LastUpdate = time.Now()

		// Update baseline if needed
		if time.Since(blastMetrics.BaselineRates.LastCalculated) > 24*time.Hour {
			bd.updateBaseline(blastMetrics)
		}
	}
}

func (bd *BlastDetector) detectBlast(tenantMetrics map[string]*collector.TenantMetrics) bool {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	for tenant, blastMetrics := range bd.metrics {
		if bd.isBlastCondition(tenant, blastMetrics) {
			bd.log.Info("blast condition detected", "tenant", tenant)
			return true
		}
	}

	return false
}

func (bd *BlastDetector) isBlastCondition(tenant string, metrics *BlastMetrics) bool {
	config := bd.config.CircuitBreaker.BlastProtection

	// Use auto-calculated thresholds if available and enabled
	if config.UseAutoThresholds {
		return bd.checkAutoThresholds(tenant, metrics)
	}

	// Fall back to manual thresholds
	return bd.checkManualThresholds(tenant, metrics)
}

// checkAutoThresholds checks blast conditions using auto-calculated thresholds
func (bd *BlastDetector) checkAutoThresholds(tenant string, metrics *BlastMetrics) bool {
	// This method needs access to the BlastProtector's autoConfig
	// For now, we'll implement a simplified version
	// In a full implementation, we'd need a reference to the BlastProtector
	
	// Check if we have tenant-specific overrides
	config := bd.config.CircuitBreaker.BlastProtection
	if override, exists := config.TenantOverrides[tenant]; exists {
		return bd.checkThresholds(metrics, override.IngestionSpikeThreshold, 
			override.QuerySpikeThreshold, override.SeriesSpikeThreshold)
	}

	// Fall back to manual thresholds
	return bd.checkManualThresholds(tenant, metrics)
}

// checkManualThresholds checks blast conditions using manual thresholds
func (bd *BlastDetector) checkManualThresholds(tenant string, metrics *BlastMetrics) bool {
	config := bd.config.CircuitBreaker.BlastProtection.ManualThresholds

	// Check if we have tenant-specific overrides
	if override, exists := bd.config.CircuitBreaker.BlastProtection.TenantOverrides[tenant]; exists {
		return bd.checkThresholds(metrics, override.IngestionSpikeThreshold, 
			override.QuerySpikeThreshold, override.SeriesSpikeThreshold)
	}

	// Use global manual thresholds
	return bd.checkThresholds(metrics, config.IngestionSpikeThreshold, 
		config.QuerySpikeThreshold, config.SeriesSpikeThreshold)
}

// checkThresholds performs the actual threshold comparison
func (bd *BlastDetector) checkThresholds(metrics *BlastMetrics, ingestionThreshold, queryThreshold, seriesThreshold float64) bool {
	// Check absolute thresholds
	if metrics.IngestionRate > ingestionThreshold {
		bd.log.V(1).Info("ingestion rate blast detected", 
			"rate", metrics.IngestionRate, "threshold", ingestionThreshold)
		return true
	}

	if metrics.QueryRate > queryThreshold {
		bd.log.V(1).Info("query rate blast detected", 
			"rate", metrics.QueryRate, "threshold", queryThreshold)
		return true
	}

	if metrics.SeriesRate > seriesThreshold {
		bd.log.V(1).Info("series rate blast detected", 
			"rate", metrics.SeriesRate, "threshold", seriesThreshold)
		return true
	}

	// Check against baseline (if available)
	if !metrics.BaselineRates.LastCalculated.IsZero() {
		blastMultiplier := 5.0
		if metrics.IngestionRate > metrics.BaselineRates.IngestionRate*blastMultiplier ||
			metrics.QueryRate > metrics.BaselineRates.QueryRate*blastMultiplier ||
			metrics.SeriesRate > metrics.BaselineRates.SeriesRate*blastMultiplier {
			bd.log.V(1).Info("baseline multiplier blast detected", 
				"multiplier", blastMultiplier)
			return true
		}
	}

	return false
}

// EnableCircuitBreaker enables the circuit breaker at runtime
func (bp *BlastProtector) EnableCircuitBreaker() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.config.CircuitBreaker.RuntimeEnabled = true
	bp.log.Info("circuit breaker enabled at runtime")
}

// DisableCircuitBreaker disables the circuit breaker at runtime
func (bp *BlastProtector) DisableCircuitBreaker() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.config.CircuitBreaker.RuntimeEnabled = false
	bp.log.Info("circuit breaker disabled at runtime")
}

// IsEnabled returns the current runtime enabled state
func (bp *BlastProtector) IsEnabled() bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.config.CircuitBreaker.Enabled && bp.config.CircuitBreaker.RuntimeEnabled
}

// UpdateCurrentLimits updates the current tenant limits for auto-configuration
func (bp *BlastProtector) UpdateCurrentLimits(limits map[string]*analyzer.TenantLimits) {
	if !bp.config.CircuitBreaker.AutoConfig.Enabled {
		return
	}

	bp.autoConfig.mu.Lock()
	defer bp.autoConfig.mu.Unlock()

	bp.autoConfig.currentLimits = limits
	bp.autoConfig.lastUpdate = time.Now()

	// Recalculate thresholds based on new limits
	bp.recalculateThresholds()
}

// initializeAutoConfig initializes the auto-configuration system
func (bp *BlastProtector) initializeAutoConfig() {
	if !bp.config.CircuitBreaker.AutoConfig.Enabled {
		bp.initialized = true
		return
	}

	bp.log.Info("initializing circuit breaker auto-configuration")
	bp.autoConfig.observationStartTime = time.Now()
	bp.initialized = true
}

// updateAutoConfiguration updates thresholds based on real-time metrics
func (bp *BlastProtector) updateAutoConfiguration(tenantMetrics map[string]*collector.TenantMetrics) {
	config := bp.config.CircuitBreaker.AutoConfig
	
	// Check if we should adapt thresholds
	if !config.RealtimeAdaptation.Enabled {
		return
	}

	now := time.Now()
	if now.Sub(bp.lastAdaptation) < config.RealtimeAdaptation.Interval {
		return
	}

	// Check minimum observation period
	if now.Sub(bp.autoConfig.observationStartTime) < config.MinObservationPeriod {
		return
	}

	bp.adaptThresholds(tenantMetrics)
	bp.lastAdaptation = now
}

// recalculateThresholds recalculates thresholds based on current limits
func (bp *BlastProtector) recalculateThresholds() {
	multipliers := bp.config.CircuitBreaker.AutoConfig.LimitMultipliers
	safetyConfig := bp.config.CircuitBreaker.AutoConfig.SafetyMargins

	for tenant, limits := range bp.autoConfig.currentLimits {
		safetyMargin := safetyConfig.DefaultMargin
		if tenantMargin, exists := safetyConfig.TenantMargins[tenant]; exists {
			safetyMargin = tenantMargin
		}

		// Extract values from dynamic limits map
		var ingestionRate, querySamples, maxSeries, ingestionBurst float64

		if val, exists := limits.Limits["ingestion_rate"]; exists {
			if v, ok := val.(float64); ok {
				ingestionRate = v
			}
		}

		if val, exists := limits.Limits["max_samples_per_query"]; exists {
			if v, ok := val.(float64); ok {
				querySamples = v
			}
		}

		if val, exists := limits.Limits["max_global_series_per_user"]; exists {
			if v, ok := val.(float64); ok {
				maxSeries = v
			}
		}

		if val, exists := limits.Limits["ingestion_burst_size"]; exists {
			if v, ok := val.(float64); ok {
				ingestionBurst = v
			}
		}

		threshold := &TenantThresholds{
			IngestionThreshold: ingestionRate * multipliers.IngestionRateMultiplier * (1 + safetyMargin/100),
			QueryThreshold:     querySamples * multipliers.QueryRateMultiplier * (1 + safetyMargin/100),
			SeriesThreshold:    maxSeries * multipliers.SeriesMultiplier * (1 + safetyMargin/100),
			BurstThreshold:     ingestionBurst * multipliers.BurstMultiplier * (1 + safetyMargin/100),
			LastCalculated:     time.Now(),
			BasedOnLimits:      limits,
			SafetyMargin:       safetyMargin,
		}

		bp.autoConfig.tenantThresholds[tenant] = threshold
		
		bp.log.V(1).Info("recalculated thresholds for tenant",
			"tenant", tenant,
			"ingestionThreshold", threshold.IngestionThreshold,
			"queryThreshold", threshold.QueryThreshold,
			"seriesThreshold", threshold.SeriesThreshold,
			"safetyMargin", safetyMargin)
	}
}

// adaptThresholds adapts thresholds based on observed metrics
func (bp *BlastProtector) adaptThresholds(tenantMetrics map[string]*collector.TenantMetrics) {
	config := bp.config.CircuitBreaker.AutoConfig.RealtimeAdaptation
	
	for tenant := range tenantMetrics {
		threshold, exists := bp.autoConfig.tenantThresholds[tenant]
		if !exists {
			continue
		}

		// Get baseline metrics for comparison
		blastMetrics := bp.blastDetector.getOrCreateBlastMetrics(tenant)
		
		// Calculate percentile values from recent metrics
		ingestionPercentile := bp.calculatePercentile(blastMetrics.IngestionRate, config.Percentile)
		queryPercentile := bp.calculatePercentile(blastMetrics.QueryRate, config.Percentile)
		seriesPercentile := bp.calculatePercentile(blastMetrics.SeriesRate, config.Percentile)

		// Adapt thresholds with learning rate
		learningRate := config.LearningRate
		maxChange := config.MaxChangePercent / 100

		// Adapt ingestion threshold
		targetIngestion := ingestionPercentile * 1.2 // 20% above observed percentile
		change := (targetIngestion - threshold.IngestionThreshold) * learningRate
		if abs(change/threshold.IngestionThreshold) > maxChange {
			if change > 0 {
				change = threshold.IngestionThreshold * maxChange
			} else {
				change = -threshold.IngestionThreshold * maxChange
			}
		}
		threshold.IngestionThreshold += change

		// Adapt query threshold
		targetQuery := queryPercentile * 1.3 // 30% above observed percentile
		change = (targetQuery - threshold.QueryThreshold) * learningRate
		if abs(change/threshold.QueryThreshold) > maxChange {
			if change > 0 {
				change = threshold.QueryThreshold * maxChange
			} else {
				change = -threshold.QueryThreshold * maxChange
			}
		}
		threshold.QueryThreshold += change

		// Adapt series threshold
		targetSeries := seriesPercentile * 1.25 // 25% above observed percentile
		change = (targetSeries - threshold.SeriesThreshold) * learningRate
		if abs(change/threshold.SeriesThreshold) > maxChange {
			if change > 0 {
				change = threshold.SeriesThreshold * maxChange
			} else {
				change = -threshold.SeriesThreshold * maxChange
			}
		}
		threshold.SeriesThreshold += change

		threshold.LastCalculated = time.Now()
		
		bp.log.V(2).Info("adapted thresholds for tenant",
			"tenant", tenant,
			"newIngestionThreshold", threshold.IngestionThreshold,
			"newQueryThreshold", threshold.QueryThreshold,
			"newSeriesThreshold", threshold.SeriesThreshold)
	}
}

// calculatePercentile calculates the specified percentile (simplified implementation)
func (bp *BlastProtector) calculatePercentile(value float64, percentile float64) float64 {
	// Simplified percentile calculation - in a real implementation, 
	// you would maintain a sliding window of values and calculate the actual percentile
	return value * (percentile / 100)
}

// abs returns the absolute value of x
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetAutoConfiguration returns the current auto-configuration state
func (bp *BlastProtector) GetAutoConfiguration() map[string]interface{} {
	bp.autoConfig.mu.RLock()
	defer bp.autoConfig.mu.RUnlock()

	result := map[string]interface{}{
		"enabled":             bp.config.CircuitBreaker.AutoConfig.Enabled,
		"mode":               bp.config.CircuitBreaker.Mode,
		"initialized":        bp.initialized,
		"observationStart":   bp.autoConfig.observationStartTime,
		"lastUpdate":         bp.autoConfig.lastUpdate,
		"lastAdaptation":     bp.lastAdaptation,
		"tenantThresholds":   make(map[string]interface{}),
	}

	for tenant, threshold := range bp.autoConfig.tenantThresholds {
		result["tenantThresholds"].(map[string]interface{})[tenant] = map[string]interface{}{
			"ingestionThreshold": threshold.IngestionThreshold,
			"queryThreshold":     threshold.QueryThreshold,
			"seriesThreshold":    threshold.SeriesThreshold,
			"burstThreshold":     threshold.BurstThreshold,
			"lastCalculated":     threshold.LastCalculated,
			"safetyMargin":       threshold.SafetyMargin,
		}
	}

	return result
}

func (bd *BlastDetector) getOrCreateBlastMetrics(tenant string) *BlastMetrics {
	if metrics, exists := bd.metrics[tenant]; exists {
		return metrics
	}

	metrics := &BlastMetrics{
		LastUpdate: time.Now(),
	}
	bd.metrics[tenant] = metrics
	return metrics
}

func (bd *BlastDetector) calculateRate(metrics *collector.TenantMetrics, metricName string) float64 {
	if data, exists := metrics.Metrics[metricName]; exists && len(data) > 0 {
		return data[len(data)-1].Value
	}
	return 0
}

func (bd *BlastDetector) updateBaseline(metrics *BlastMetrics) {
	// Calculate baseline as average of recent stable periods
	metrics.BaselineRates.IngestionRate = metrics.IngestionRate * 0.8 // Conservative baseline
	metrics.BaselineRates.QueryRate = metrics.QueryRate * 0.8
	metrics.BaselineRates.SeriesRate = metrics.SeriesRate * 0.8
	metrics.BaselineRates.LastCalculated = time.Now()
} 