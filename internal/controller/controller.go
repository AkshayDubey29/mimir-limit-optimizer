package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/analyzer"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/auditlog"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/circuitbreaker"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/collector"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/costcontrol"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/patcher"
)

// MimirLimitController orchestrates the complete limit optimization workflow
type MimirLimitController struct {
	client.Client
	Scheme          *runtime.Scheme
	Config          *config.Config
	Log             logr.Logger
	KubeClient      kubernetes.Interface
	
	// Core components
	Collector       collector.Collector
	Analyzer        analyzer.Analyzer
	Patcher         patcher.Patcher
	AuditLogger     auditlog.AuditLogger
	
	// Enterprise components
	CostController  *costcontrol.CostController
	BlastProtector  *circuitbreaker.BlastProtector
	
	// Internal state
	lastReconcile   time.Time
	reconcileCount  int64
	tenantFilter    *TenantFilter
}

// TenantFilter handles tenant filtering logic
type TenantFilter struct {
	config *config.Config
	log    logr.Logger
}

// NewTenantFilter creates a new tenant filter
func NewTenantFilter(cfg *config.Config, log logr.Logger) *TenantFilter {
	return &TenantFilter{
		config: cfg,
		log:    log,
	}
}

// ShouldProcessTenant determines if a tenant should be processed
func (tf *TenantFilter) ShouldProcessTenant(tenant string) bool {
	// Check skip list
	for _, pattern := range tf.config.TenantScoping.SkipList {
		if tf.matchPattern(tenant, pattern) {
			tf.log.V(1).Info("skipping tenant due to skip list", "tenant", tenant, "pattern", pattern)
			return false
		}
	}
	
	// Check include list (if specified, only include matching tenants)
	if len(tf.config.TenantScoping.IncludeList) > 0 {
		for _, pattern := range tf.config.TenantScoping.IncludeList {
			if tf.matchPattern(tenant, pattern) {
				return true
			}
		}
		tf.log.V(1).Info("skipping tenant not in include list", "tenant", tenant)
		return false
	}
	
	return true
}

// matchPattern performs pattern matching (glob or regex)
func (tf *TenantFilter) matchPattern(tenant, pattern string) bool {
	// Simple glob matching for now
	if pattern == "*" {
		return true
	}
	
	// Prefix matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(tenant) >= len(prefix) && tenant[:len(prefix)] == prefix
	}
	
	// Suffix matching
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(tenant) >= len(suffix) && tenant[len(tenant)-len(suffix):] == suffix
	}
	
	// Exact matching
	return tenant == pattern
}

// FilterTenants filters a list of tenants based on configuration
func (tf *TenantFilter) FilterTenants(tenants []string) (monitored, skipped []string) {
	for _, tenant := range tenants {
		if tf.ShouldProcessTenant(tenant) {
			monitored = append(monitored, tenant)
		} else {
			skipped = append(skipped, tenant)
		}
	}
	return
}

// SetupWithManager sets up the controller with the Manager
func (r *MimirLimitController) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize Kubernetes client
	config := mgr.GetConfig()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	r.KubeClient = kubeClient
	
	// Initialize components
	r.AuditLogger = auditlog.NewAuditLogger(r.Config, r.Client, r.Log.WithName("audit"))
	r.Collector = collector.NewCollector(r.Config, kubeClient, r.Log.WithName("collector"))
	r.Analyzer = analyzer.NewAnalyzer(r.Config, r.Log.WithName("analyzer"))
	r.Patcher = patcher.NewPatcher(r.Client, kubeClient, r.Config, r.AuditLogger, r.Log.WithName("patcher"))
	r.tenantFilter = NewTenantFilter(r.Config, r.Log.WithName("filter"))
	
	// Initialize enterprise components
	r.CostController = costcontrol.NewCostController(r.Config, r.Log.WithName("cost"))
	r.BlastProtector = circuitbreaker.NewBlastProtector(r.Config, r.Log.WithName("protection"))
	
	// Set up periodic reconciliation instead of watching resources
	return mgr.Add(&PeriodicReconciler{
		Controller: r,
		Interval:   r.Config.UpdateInterval,
		Log:        r.Log,
	})
}

// PeriodicReconciler runs the reconciliation loop periodically
type PeriodicReconciler struct {
	Controller *MimirLimitController
	Interval   time.Duration
	Log        logr.Logger
	stopCh     chan struct{}
}

// Start starts the periodic reconciliation
func (pr *PeriodicReconciler) Start(ctx context.Context) error {
	pr.stopCh = make(chan struct{})
	
	pr.Log.Info("starting periodic reconciler", "interval", pr.Interval)
	
	ticker := time.NewTicker(pr.Interval)
	defer ticker.Stop()
	
	// Initial reconciliation
	if err := pr.Controller.reconcile(ctx); err != nil {
		pr.Log.Error(err, "initial reconciliation failed")
	}
	
	for {
		select {
		case <-ctx.Done():
			pr.Log.Info("context cancelled, stopping reconciler")
			return nil
		case <-pr.stopCh:
			pr.Log.Info("stop signal received, stopping reconciler")
			return nil
		case <-ticker.C:
			if err := pr.Controller.reconcile(ctx); err != nil {
				pr.Log.Error(err, "reconciliation failed")
				metrics.ReconcileMetricsInstance.IncReconcileTotal("error")
			}
		}
	}
}

// Stop stops the periodic reconciliation
func (pr *PeriodicReconciler) Stop() {
	if pr.stopCh != nil {
		close(pr.stopCh)
	}
}

// reconcile performs the main reconciliation logic
func (r *MimirLimitController) reconcile(ctx context.Context) error {
	startTime := time.Now()
	r.reconcileCount++
	
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ReconcileMetricsInstance.ObserveReconcileDuration("success", duration)
		metrics.ReconcileMetricsInstance.SetLastReconcileTime(float64(time.Now().Unix()))
		r.lastReconcile = time.Now()
	}()
	
	r.Log.Info("starting reconciliation", "count", r.reconcileCount)
	
	// Update health status
	metrics.HealthMetricsInstance.SetHealthStatus("controller", 1)
	
	// Step 1: Collect metrics from all sources
	tenantMetrics, err := r.Collector.CollectMetrics(ctx)
	if err != nil {
		metrics.HealthMetricsInstance.SetHealthStatus("collector", 0)
		metrics.HealthMetricsInstance.IncErrorTotal("collector", "metrics-collection")
		return fmt.Errorf("failed to collect metrics: %w", err)
	}
	metrics.HealthMetricsInstance.SetHealthStatus("collector", 1)
	
	r.Log.Info("collected metrics", "tenants", len(tenantMetrics))
	
	// Step 2: Filter tenants based on configuration
	allTenants := make([]string, 0, len(tenantMetrics))
	for tenant := range tenantMetrics {
		allTenants = append(allTenants, tenant)
	}
	
	monitoredTenants, skippedTenants := r.tenantFilter.FilterTenants(allTenants)
	
	// Update metrics
	metrics.TenantMetricsInstance.SetTenantsMonitored(float64(len(monitoredTenants)))
	metrics.TenantMetricsInstance.SetTenantsSkipped(float64(len(skippedTenants)))
	
	r.Log.Info("filtered tenants", 
		"monitored", len(monitoredTenants), 
		"skipped", len(skippedTenants))
	
	// Filter tenant metrics to only include monitored tenants
	filteredMetrics := make(map[string]*collector.TenantMetrics)
	for _, tenant := range monitoredTenants {
		if tm, exists := tenantMetrics[tenant]; exists {
			filteredMetrics[tenant] = tm
		}
	}
	
	// Step 2.5: Apply blast protection and circuit breaker
	protectedMetrics, err := r.BlastProtector.ProcessMetrics(ctx, filteredMetrics)
	if err != nil {
		r.Log.Error(err, "failed to apply blast protection")
		protectedMetrics = filteredMetrics // Continue with original metrics
	}
	
	// Step 3: Calculate costs (enterprise feature)
	var tenantCosts map[string]*costcontrol.TenantCostData
	if r.Config.CostControl.Enabled {
		tenantCosts, err = r.CostController.CalculateCosts(ctx, protectedMetrics)
		if err != nil {
			r.Log.Error(err, "failed to calculate costs")
		} else {
			r.Log.Info("calculated tenant costs", "tenants", len(tenantCosts))
		}
	}
	
	// Step 4: Detect spikes (if enabled)
	if r.Config.EventSpike.Enabled {
		spikes, err := r.Analyzer.DetectSpikes(ctx, protectedMetrics)
		if err != nil {
			r.Log.Error(err, "failed to detect spikes")
			metrics.HealthMetricsInstance.IncErrorTotal("analyzer", "spike-detection")
		} else {
			r.handleSpikes(ctx, spikes)
		}
	}
	
	// Step 5: Analyze trends and calculate recommended limits
	analysisResults, err := r.Analyzer.AnalyzeTrends(ctx, protectedMetrics)
	if err != nil {
		metrics.HealthMetricsInstance.SetHealthStatus("analyzer", 0)
		metrics.HealthMetricsInstance.IncErrorTotal("analyzer", "trend-analysis")
		return fmt.Errorf("failed to analyze trends: %w", err)
	}
	metrics.HealthMetricsInstance.SetHealthStatus("analyzer", 1)
	
	r.Log.Info("analyzed trends", "tenants", len(analysisResults))
	
	// Step 6: Calculate optimized limits
	optimizedLimits, err := r.Analyzer.CalculateLimits(ctx, analysisResults)
	if err != nil {
		metrics.HealthMetricsInstance.IncErrorTotal("analyzer", "limit-calculation")
		return fmt.Errorf("failed to calculate limits: %w", err)
	}
	
	r.Log.Info("calculated optimized limits", "tenants", len(optimizedLimits))
	
	// Step 7: Apply cost control and budget enforcement
	finalLimits := optimizedLimits
	if r.Config.CostControl.Enabled && tenantCosts != nil {
		finalLimits, err = r.CostController.EnforceBudgets(ctx, tenantCosts, optimizedLimits)
		if err != nil {
			r.Log.Error(err, "failed to enforce budgets")
			finalLimits = optimizedLimits // Continue with original limits
		} else {
			r.Log.Info("applied cost control", "tenants", len(finalLimits))
		}
	}
	
	// Step 8: Apply blast protection to final limits
	protectedLimits, err := r.BlastProtector.ApplyProtection(ctx, finalLimits)
	if err != nil {
		r.Log.Error(err, "failed to apply blast protection to limits")
		protectedLimits = finalLimits // Continue with unprotected limits
	}
	
	// Step 9: Apply limits (or preview in dry-run mode)
	if r.Config.Mode == "dry-run" {
		r.Log.Info("DRY-RUN mode: limits preview with enterprise protections")
		preview, err := r.Patcher.PreviewLimits(ctx, protectedLimits)
		if err != nil {
			metrics.HealthMetricsInstance.IncErrorTotal("patcher", "preview")
			return fmt.Errorf("failed to preview limits: %w", err)
		}
		
		r.logPreview(preview)
		r.logEnterpriseStatus(ctx, tenantCosts, r.BlastProtector.GetProtectionStatus())
	} else {
		// Production mode: apply limits
		if err := r.Patcher.ApplyLimits(ctx, protectedLimits); err != nil {
			metrics.HealthMetricsInstance.SetHealthStatus("patcher", 0)
			metrics.HealthMetricsInstance.IncErrorTotal("patcher", "apply-limits")
			return fmt.Errorf("failed to apply limits: %w", err)
		}
		metrics.HealthMetricsInstance.SetHealthStatus("patcher", 1)
	}
	
	// Step 10: Update current limits metrics
	r.updateCurrentLimitsMetrics(ctx, protectedLimits)
	
	// Step 11: Cleanup old audit entries (if enabled)
	if r.Config.AuditLog.Enabled {
		cutoff := time.Now().Add(-7 * 24 * time.Hour) // Keep 7 days of audit logs
		if err := r.AuditLogger.PurgeOldEntries(ctx, cutoff); err != nil {
			r.Log.Error(err, "failed to purge old audit entries")
		}
	}
	
	metrics.ReconcileMetricsInstance.IncReconcileTotal("success")
	r.Log.Info("reconciliation completed successfully with enterprise protection", 
		"duration", time.Since(startTime),
		"tenants_processed", len(protectedLimits),
		"cost_control_enabled", r.Config.CostControl.Enabled,
		"blast_protection_enabled", r.Config.CircuitBreaker.Enabled)
	
	return nil
}

// handleSpikes processes detected spikes
func (r *MimirLimitController) handleSpikes(ctx context.Context, spikes map[string]map[string]bool) {
	for tenant, tenantSpikes := range spikes {
		for metricName := range tenantSpikes {
			r.Log.Info("spike detected", "tenant", tenant, "metric", metricName)
			
			// Log spike detection to audit trail
			entry := auditlog.NewSpikeDetectionEntry(tenant, metricName, 0, 0) // Values would be filled by analyzer
			if err := r.AuditLogger.LogEntry(entry); err != nil {
				r.Log.Error(err, "failed to log spike detection", "tenant", tenant)
			}
		}
	}
}

// logPreview logs the preview results in dry-run mode
func (r *MimirLimitController) logPreview(preview *patcher.PreviewResult) {
	r.Log.Info("DRY-RUN Preview Results",
		"configmap", preview.ConfigMapName,
		"namespace", preview.Namespace,
		"affected_tenants", len(preview.AffectedTenants),
		"estimated_changes", preview.EstimatedChanges)
	
	for _, tenant := range preview.AffectedTenants {
		r.Log.Info("DRY-RUN: would update tenant", "tenant", tenant)
	}
}

// logEnterpriseStatus logs enterprise feature status in dry-run mode
func (r *MimirLimitController) logEnterpriseStatus(ctx context.Context, costs map[string]*costcontrol.TenantCostData, protectionStatus map[string]interface{}) {
	if r.Config.CostControl.Enabled && costs != nil {
		totalCost := 0.0
		for _, cost := range costs {
			totalCost += cost.DailyCost
		}
		r.Log.Info("DRY-RUN Cost Control Status",
			"total_daily_cost", totalCost,
			"tenants_with_costs", len(costs),
			"currency", r.Config.CostControl.GlobalBudget.Currency)
	}
	
	if r.Config.CircuitBreaker.Enabled {
		r.Log.Info("DRY-RUN Blast Protection Status",
			"circuit_breaker_state", protectionStatus["circuit_breaker_state"],
			"emergency_mode", protectionStatus["emergency_mode"],
			"panic_mode", protectionStatus["panic_mode"])
	}
}

// updateCurrentLimitsMetrics updates metrics with current limit values
func (r *MimirLimitController) updateCurrentLimitsMetrics(ctx context.Context, limits map[string]*analyzer.TenantLimits) {
	for tenant, limit := range limits {
		if limit.IngestionRate > 0 {
			metrics.TenantMetricsInstance.SetTenantCurrentLimits(tenant, "ingestion_rate", limit.IngestionRate)
		}
		if limit.MaxSeries > 0 {
			metrics.TenantMetricsInstance.SetTenantCurrentLimits(tenant, "max_series", limit.MaxSeries)
		}
		if limit.MaxSamplesPerQuery > 0 {
			metrics.TenantMetricsInstance.SetTenantCurrentLimits(tenant, "max_samples_per_query", limit.MaxSamplesPerQuery)
		}
	}
}

// GetStatus returns the current status of the controller
func (r *MimirLimitController) GetStatus() *ControllerStatus {
	return &ControllerStatus{
		LastReconcile:    r.lastReconcile,
		ReconcileCount:   r.reconcileCount,
		Mode:             r.Config.Mode,
		UpdateInterval:   r.Config.UpdateInterval,
		ComponentsHealth: r.getComponentsHealth(),
	}
}

// ControllerStatus represents the current status of the controller
type ControllerStatus struct {
	LastReconcile    time.Time         `json:"last_reconcile"`
	ReconcileCount   int64             `json:"reconcile_count"`
	Mode             string            `json:"mode"`
	UpdateInterval   time.Duration     `json:"update_interval"`
	ComponentsHealth map[string]bool   `json:"components_health"`
}

// getComponentsHealth checks the health of all components
func (r *MimirLimitController) getComponentsHealth() map[string]bool {
	return map[string]bool{
		"collector":    true, // TODO: Implement actual health checks
		"analyzer":     true,
		"patcher":      true,
		"audit_logger": r.Config.AuditLog.Enabled,
	}
}

// TriggerReconciliation manually triggers a reconciliation (for testing/debugging)
func (r *MimirLimitController) TriggerReconciliation(ctx context.Context) error {
	r.Log.Info("manually triggered reconciliation")
	return r.reconcile(ctx)
}

// GetAuditEntries retrieves audit entries with optional filtering
func (r *MimirLimitController) GetAuditEntries(ctx context.Context, filter *auditlog.AuditFilter) ([]*auditlog.AuditEntry, error) {
	if r.AuditLogger == nil {
		return nil, fmt.Errorf("audit logging not enabled")
	}
	return r.AuditLogger.GetEntries(ctx, filter)
}

// RollbackLastChange rolls back the last configuration change
func (r *MimirLimitController) RollbackLastChange(ctx context.Context) error {
	r.Log.Info("rolling back last configuration change")
	
	if err := r.Patcher.RollbackChanges(ctx); err != nil {
		return fmt.Errorf("failed to rollback changes: %w", err)
	}
	
	// Log rollback to audit trail
	entry := &auditlog.AuditEntry{
		Action:  "rollback",
		Reason:  "manual-rollback",
		Success: true,
		Changes: map[string]interface{}{"rollback": "configuration restored"},
	}
	
	if err := r.AuditLogger.LogEntry(entry); err != nil {
		r.Log.Error(err, "failed to log rollback")
	}
	
	return nil
}

// Shutdown gracefully shuts down the controller
func (r *MimirLimitController) Shutdown(ctx context.Context) error {
	r.Log.Info("shutting down controller")
	
	if r.AuditLogger != nil {
		if err := r.AuditLogger.Close(); err != nil {
			r.Log.Error(err, "failed to close audit logger")
		}
	}
	
	return nil
} 