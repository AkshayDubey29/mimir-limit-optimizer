package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/auditlog"
)

// SystemStatus represents the overall system status
type SystemStatus struct {
	Mode                string            `json:"mode"`
	LastReconcile       time.Time        `json:"last_reconcile"`
	ReconcileCount      int64            `json:"reconcile_count"`
	UpdateInterval      time.Duration    `json:"update_interval"`
	ComponentsHealth    map[string]bool  `json:"components_health"`
	CircuitBreakerState string           `json:"circuit_breaker_state"`
	SpikeDetectionState string           `json:"spike_detection_state"`
	PanicModeActive     bool             `json:"panic_mode_active"`
	TotalTenants        int              `json:"total_tenants"`
	MonitoredTenants    int              `json:"monitored_tenants"`
	SkippedTenants      int              `json:"skipped_tenants"`
	ConfigMapName       string           `json:"config_map_name"`
	Version             string           `json:"version"`
	BuildInfo           BuildInfo        `json:"build_info"`
}

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

type TenantInfo struct {
	ID                    string                 `json:"id"`
	IngestionRate         float64               `json:"ingestion_rate"`
	ActiveSeries          int64                 `json:"active_series"`
	AppliedLimits         map[string]interface{} `json:"applied_limits"`
	SuggestedLimits       map[string]interface{} `json:"suggested_limits"`
	SpikeDetected         bool                  `json:"spike_detected"`
	LastConfigChange      time.Time             `json:"last_config_change"`
	BufferUsagePercent    float64               `json:"buffer_usage_percent"`
	UsageSparkline        []float64             `json:"usage_sparkline"`
	Status                string                `json:"status"`
}

type ConfigUpdateRequest struct {
	Mode                string        `json:"mode"`
	BufferPercentage    float64       `json:"buffer_percentage"`
	SpikeThreshold      float64       `json:"spike_threshold"`
	UpdateInterval      time.Duration `json:"update_interval"`
	CircuitBreakerEnabled bool        `json:"circuit_breaker_enabled"`
	AutoDiscoveryEnabled bool         `json:"auto_discovery_enabled"`
	SkipList            []string      `json:"skip_list"`
	IncludeList         []string      `json:"include_list"`
	EnabledLimits       []string      `json:"enabled_limits"`
}

type DiffItem struct {
	LimitName      string      `json:"limit_name"`
	DryRunValue    interface{} `json:"dry_run_value"`
	AppliedValue   interface{} `json:"applied_value"`
	Delta          interface{} `json:"delta"`
	Status         string      `json:"status"` // "identical", "mismatched", "dry_run_only"
	TenantID       string      `json:"tenant_id"`
}

// handleStatus returns the current system status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	controllerStatus := s.controller.GetStatus()
	
	status := SystemStatus{
		Mode:               s.config.Mode,
		LastReconcile:      controllerStatus.LastReconcile,
		ReconcileCount:     controllerStatus.ReconcileCount,
		UpdateInterval:     controllerStatus.UpdateInterval,
		ComponentsHealth:   controllerStatus.ComponentsHealth,
		ConfigMapName:      s.config.Mimir.ConfigMapName,
		CircuitBreakerState: "CLOSED", // TODO: Get actual state from controller
		SpikeDetectionState: "ACTIVE", // TODO: Get actual state from controller
		PanicModeActive:    false,     // TODO: Get actual state from controller
		BuildInfo: BuildInfo{
			Version:   "dev", // TODO: Get from build info
			Commit:    "unknown",
			BuildDate: "unknown",
		},
	}
	
	s.writeJSON(w, status)
}

// handleConfig handles configuration get/update requests
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.writeJSON(w, s.config)
	case "POST":
		var updateReq ConfigUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		
		// Update configuration
		s.updateConfig(&updateReq)
		
		// In production mode, also update the ConfigMap
		if s.config.Mode == "prod" {
			// TODO: Update ConfigMap via controller
			s.log.Info("updating config map", "config", updateReq)
		}
		
		s.writeJSON(w, map[string]string{"status": "updated"})
	}
}

// handleTenants returns a list of all tenants with their basic info
func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get tenant list from collector
	tenants, err := s.controller.Collector.GetTenantList(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get tenant list")
		return
	}
	
	// Filter tenants
	tenantFilter := s.controller.GetTenantFilter()
	monitored, skipped := tenantFilter.FilterTenants(tenants)
	
	var tenantInfos []TenantInfo
	for _, tenant := range monitored {
		info := s.getTenantInfo(ctx, tenant)
		tenantInfos = append(tenantInfos, info)
	}
	
	response := map[string]interface{}{
		"tenants":          tenantInfos,
		"total_tenants":    len(tenants),
		"monitored_count":  len(monitored),
		"skipped_count":    len(skipped),
		"skipped_tenants":  skipped,
	}
	
	s.writeJSON(w, response)
}

// handleTenantDetail returns detailed information for a specific tenant
func (s *Server) handleTenantDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]
	
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "Tenant ID is required")
		return
	}
	
	ctx := r.Context()
	tenantInfo := s.getTenantInfo(ctx, tenantID)
	
	// Get additional detailed metrics
	detailed := map[string]interface{}{
		"tenant_info":      tenantInfo,
		"usage_trends":     s.getTenantUsageTrends(ctx, tenantID),
		"recent_changes":   s.getTenantRecentChanges(ctx, tenantID),
		"limit_comparison": s.getTenantLimitComparison(ctx, tenantID),
	}
	
	s.writeJSON(w, detailed)
}

// handleDiff returns the diff between dry-run and applied limits
func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get current applied limits
	appliedLimits, err := s.getAppliedLimits(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get applied limits")
		return
	}
	
	// Get dry-run suggestions
	dryRunLimits, err := s.getDryRunLimits(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get dry-run limits")
		return
	}
	
	// Compare and create diff
	diffs := s.compareLimits(appliedLimits, dryRunLimits)
	
	response := map[string]interface{}{
		"differences":      diffs,
		"total_diffs":      len(diffs),
		"identical_count":  s.countByStatus(diffs, "identical"),
		"mismatched_count": s.countByStatus(diffs, "mismatched"),
		"dry_run_only":     s.countByStatus(diffs, "dry_run_only"),
		"timestamp":        time.Now(),
	}
	
	s.writeJSON(w, response)
}

// handleAudit returns audit log entries
func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 100 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	
	filter := &auditlog.AuditFilter{
		Limit:    limit,
		Offset:   offset,
	}
	
	entries, err := s.controller.GetAuditEntries(ctx, filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get audit entries")
		return
	}
	
	s.writeJSON(w, map[string]interface{}{
		"entries": entries,
		"total":   len(entries),
		"filter":  filter,
	})
}

// handleMetrics returns Prometheus-formatted metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// This is handled by the Prometheus handler in the router
	http.Redirect(w, r, "/metrics", http.StatusTemporaryRedirect)
}

// handleTestSpike triggers a synthetic ingestion spike
func (s *Server) handleTestSpike(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID   string  `json:"tenant_id"`
		Multiplier float64 `json:"multiplier"`
		Duration   string  `json:"duration"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}
	
	// TODO: Implement synthetic spike generation
	s.log.Info("synthetic spike triggered", "tenant", req.TenantID, "multiplier", req.Multiplier, "duration", duration)
	
	s.writeJSON(w, map[string]string{"status": "spike_triggered"})
}

// handleTestAlert triggers a test alert
func (s *Server) handleTestAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	
	// TODO: Implement test alert via alerting system
	s.log.Info("test alert triggered", "channel", req.Channel, "message", req.Message)
	
	s.writeJSON(w, map[string]string{"status": "alert_sent"})
}

// handleTestReconcile triggers a manual reconciliation
func (s *Server) handleTestReconcile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := s.controller.TriggerReconciliation(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to trigger reconciliation")
		return
	}
	
	s.writeJSON(w, map[string]string{"status": "reconciliation_triggered"})
}

// handleHealthCheck performs a basic health check
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()), // TODO: Track actual uptime
	}
	
	s.writeJSON(w, health)
}

// Helper methods

func (s *Server) updateConfig(req *ConfigUpdateRequest) {
	if req.Mode != "" {
		s.config.Mode = req.Mode
	}
	if req.BufferPercentage > 0 {
		s.config.BufferPercentage = req.BufferPercentage
	}
	if req.SpikeThreshold > 0 {
		s.config.EventSpike.Threshold = req.SpikeThreshold
	}
	if req.UpdateInterval > 0 {
		s.config.UpdateInterval = req.UpdateInterval
	}
	if len(req.SkipList) > 0 {
		s.config.TenantScoping.SkipList = req.SkipList
	}
	if len(req.IncludeList) > 0 {
		s.config.TenantScoping.IncludeList = req.IncludeList
	}
	// TODO: Update other configuration fields
}

func (s *Server) getTenantInfo(ctx context.Context, tenantID string) TenantInfo {
	// TODO: Get actual tenant metrics from collector/analyzer
	return TenantInfo{
		ID:                 tenantID,
		IngestionRate:      1000.0, // placeholder
		ActiveSeries:       10000,  // placeholder
		AppliedLimits:      map[string]interface{}{"ingestion_rate": 1200.0},
		SuggestedLimits:    map[string]interface{}{"ingestion_rate": 1100.0},
		SpikeDetected:      false,
		LastConfigChange:   time.Now().Add(-1 * time.Hour),
		BufferUsagePercent: 85.5,
		UsageSparkline:     []float64{100, 120, 110, 150, 130, 140, 135},
		Status:             "active",
	}
}

func (s *Server) getTenantUsageTrends(ctx context.Context, tenantID string) map[string]interface{} {
	// TODO: Implement actual trend analysis
	return map[string]interface{}{
		"ingestion_trend": []float64{100, 105, 110, 115, 120},
		"series_trend":    []int64{1000, 1050, 1100, 1150, 1200},
		"query_trend":     []float64{50, 55, 60, 58, 62},
	}
}

func (s *Server) getTenantRecentChanges(ctx context.Context, tenantID string) []map[string]interface{} {
	// TODO: Get actual recent changes from audit log
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-2 * time.Hour),
			"action":    "limit_increase",
			"field":     "ingestion_rate",
			"old_value": 1000.0,
			"new_value": 1200.0,
		},
	}
}

func (s *Server) getTenantLimitComparison(ctx context.Context, tenantID string) map[string]interface{} {
	// TODO: Get actual limit comparison
	return map[string]interface{}{
		"current_limits":    map[string]interface{}{"ingestion_rate": 1200.0},
		"suggested_limits":  map[string]interface{}{"ingestion_rate": 1100.0},
		"baseline_usage":    map[string]interface{}{"ingestion_rate": 950.0},
	}
}

func (s *Server) getAppliedLimits(ctx context.Context) (map[string]map[string]interface{}, error) {
	// TODO: Get actual applied limits from ConfigMap
	return map[string]map[string]interface{}{
		"tenant1": {"ingestion_rate": 1200.0},
		"tenant2": {"ingestion_rate": 800.0},
	}, nil
}

func (s *Server) getDryRunLimits(ctx context.Context) (map[string]map[string]interface{}, error) {
	// TODO: Get actual dry-run limits from analyzer
	return map[string]map[string]interface{}{
		"tenant1": {"ingestion_rate": 1100.0},
		"tenant2": {"ingestion_rate": 850.0},
	}, nil
}

func (s *Server) compareLimits(applied, dryRun map[string]map[string]interface{}) []DiffItem {
	var diffs []DiffItem
	
	// Compare limits for each tenant
	for tenant, dryLimits := range dryRun {
		appliedLimits, exists := applied[tenant]
		
		for limitName, dryValue := range dryLimits {
			diff := DiffItem{
				LimitName: limitName,
				TenantID:  tenant,
				DryRunValue: dryValue,
			}
			
			if !exists {
				diff.Status = "dry_run_only"
				diff.AppliedValue = nil
			} else if appliedValue, limitExists := appliedLimits[limitName]; !limitExists {
				diff.Status = "dry_run_only"
				diff.AppliedValue = nil
			} else {
				diff.AppliedValue = appliedValue
				if s.valuesEqual(dryValue, appliedValue) {
					diff.Status = "identical"
					diff.Delta = 0
				} else {
					diff.Status = "mismatched"
					diff.Delta = s.calculateDelta(dryValue, appliedValue)
				}
			}
			
			diffs = append(diffs, diff)
		}
	}
	
	return diffs
}

func (s *Server) countByStatus(diffs []DiffItem, status string) int {
	count := 0
	for _, diff := range diffs {
		if diff.Status == status {
			count++
		}
	}
	return count
}

func (s *Server) valuesEqual(a, b interface{}) bool {
	// Simple equality check - could be enhanced for float comparison
	return a == b
}

func (s *Server) calculateDelta(a, b interface{}) interface{} {
	// Calculate delta between two values
	if aFloat, aOk := a.(float64); aOk {
		if bFloat, bOk := b.(float64); bOk {
			return aFloat - bFloat
		}
	}
	return "N/A"
} 