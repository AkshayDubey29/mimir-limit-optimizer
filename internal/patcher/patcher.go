package patcher

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/analyzer"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/auditlog"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

// Patcher interface defines methods for patching configurations
type Patcher interface {
	ApplyLimits(ctx context.Context, limits map[string]*analyzer.TenantLimits) error
	PreviewLimits(ctx context.Context, limits map[string]*analyzer.TenantLimits) (*PreviewResult, error)
	RollbackChanges(ctx context.Context) error
	GetCurrentLimits(ctx context.Context) (map[string]*analyzer.TenantLimits, error)
}

// PreviewResult contains the preview of changes to be made
type PreviewResult struct {
	ConfigMapName     string
	Namespace         string
	CurrentData       map[string]interface{}
	ProposedData      map[string]interface{}
	AffectedTenants   []string
	EstimatedChanges  int
	PreviewTime       time.Time
}

// ConfigMapPatcher implements the Patcher interface for ConfigMap-based runtime overrides
type ConfigMapPatcher struct {
	client        client.Client
	kubeClient    kubernetes.Interface
	config        *config.Config
	auditLog      auditlog.AuditLogger
	log           logr.Logger
	lastBackup    *corev1.ConfigMap
}

// NewConfigMapPatcher creates a new ConfigMapPatcher
func NewConfigMapPatcher(c client.Client, kubeClient kubernetes.Interface, cfg *config.Config, auditLogger auditlog.AuditLogger, log logr.Logger) *ConfigMapPatcher {
	return &ConfigMapPatcher{
		client:     c,
		kubeClient: kubeClient,
		config:     cfg,
		auditLog:   auditLogger,
		log:        log,
	}
}

// ApplyLimits applies the calculated limits to the Mimir runtime overrides ConfigMap with retry logic for conflict resolution
func (p *ConfigMapPatcher) ApplyLimits(ctx context.Context, limits map[string]*analyzer.TenantLimits) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ConfigMapMetricsInstance.ObserveConfigMapUpdateDuration("success", duration)
	}()

	// Retry logic with exponential backoff for conflict resolution
	maxRetries := 5
	baseDelay := 150 * time.Millisecond
	var currentOverrides, updatedOverrides map[string]interface{}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get current ConfigMap (fresh read each time)
		currentConfigMap, err := p.getCurrentConfigMap(ctx)
		if err != nil {
			metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
			return fmt.Errorf("failed to get current ConfigMap: %w", err)
		}

		// Create backup on first attempt only
		if attempt == 0 {
			p.createBackup(currentConfigMap)
		}

		// Parse current overrides
		currentOverrides, err = p.parseOverrides(currentConfigMap)
		if err != nil {
			metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
			return fmt.Errorf("failed to parse current overrides: %w", err)
		}

		// Apply new limits
		updatedOverrides = p.applyLimitsToOverrides(currentOverrides, limits)

		// Try to update ConfigMap
		if err := p.updateConfigMap(ctx, currentConfigMap, updatedOverrides); err != nil {
			// Check if it's a conflict error
			if apierrors.IsConflict(err) {
				if attempt < maxRetries-1 {
					// Wait with exponential backoff before retrying
					delay := time.Duration(1<<attempt) * baseDelay
					p.log.V(1).Info("runtime overrides ConfigMap conflict, retrying",
						"attempt", attempt+1,
						"delay", delay,
						"configmap", p.config.Mimir.ConfigMapName,
						"tenants", len(limits))
					time.Sleep(delay)
					continue
				}
				// Max retries exceeded
				metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
				return fmt.Errorf("failed to update runtime overrides ConfigMap after %d retries due to conflicts: %w", maxRetries, err)
			}
			// Non-conflict error, return immediately
			metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}

		// Success - break out of retry loop
		break
	}

	// Log changes to audit trail (using the final successful values)
	p.logChanges(currentOverrides, updatedOverrides, limits)

	// Trigger rollout if configured (optional - runtime overrides work without restarts)
	if p.config.Mimir.TriggerRollout {
		if err := p.triggerRollout(ctx); err != nil {
			p.log.Error(err, "failed to trigger optional rollout (continuing anyway - runtime overrides still work)")
		} else {
			p.log.Info("triggered optional component rollouts", "note", "runtime overrides work without restarts")
		}
	}

	metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("success")
	metrics.ConfigMapMetricsInstance.SetLastConfigMapUpdate(float64(time.Now().Unix()))

	if p.config.Mode == "dry-run" {
		p.log.Info("successfully wrote optimized limits to ConfigMap for verification", 
			"tenants", len(limits),
			"mode", "dry-run",
			"configmap", p.config.Mimir.ConfigMapName,
			"namespace", p.config.Mimir.Namespace)
	} else {
		p.log.Info("successfully applied limits for production use", 
			"tenants", len(limits),
			"mode", "production",
			"configmap", p.config.Mimir.ConfigMapName,
			"namespace", p.config.Mimir.Namespace)
	}

	return nil
}

// PreviewLimits previews the changes that would be made without applying them
func (p *ConfigMapPatcher) PreviewLimits(ctx context.Context, limits map[string]*analyzer.TenantLimits) (*PreviewResult, error) {
	// Get current ConfigMap
	currentConfigMap, err := p.getCurrentConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current ConfigMap: %w", err)
	}

	// Parse current overrides
	currentOverrides, err := p.parseOverrides(currentConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current overrides: %w", err)
	}

	// Apply new limits to a copy
	proposedOverrides := p.applyLimitsToOverrides(copyOverrides(currentOverrides), limits)

	// Calculate affected tenants
	affectedTenants := make([]string, 0, len(limits))
	for tenant := range limits {
		affectedTenants = append(affectedTenants, tenant)
	}

	// Count estimated changes
	changes := p.countChanges(currentOverrides, proposedOverrides)

	return &PreviewResult{
		ConfigMapName:    p.config.Mimir.ConfigMapName,
		Namespace:        p.config.Mimir.Namespace,
		CurrentData:      currentOverrides,
		ProposedData:     proposedOverrides,
		AffectedTenants:  affectedTenants,
		EstimatedChanges: changes,
		PreviewTime:      time.Now(),
	}, nil
}

// RollbackChanges rolls back to the previous configuration with retry logic for conflict resolution
func (p *ConfigMapPatcher) RollbackChanges(ctx context.Context) error {
	if p.lastBackup == nil {
		return fmt.Errorf("no backup available for rollback")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ConfigMapMetricsInstance.ObserveConfigMapUpdateDuration("rollback", duration)
	}()

	// Retry logic with exponential backoff for conflict resolution
	maxRetries := 5
	baseDelay := 150 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get current ConfigMap (fresh read each time)
		currentConfigMap, err := p.getCurrentConfigMap(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current ConfigMap for rollback: %w", err)
		}

		// Restore data from backup
		currentConfigMap.Data = p.lastBackup.Data
		
		if err := p.client.Update(ctx, currentConfigMap); err != nil {
			// Check if it's a conflict error
			if apierrors.IsConflict(err) {
				if attempt < maxRetries-1 {
					// Wait with exponential backoff before retrying
					delay := time.Duration(1<<attempt) * baseDelay
					p.log.V(1).Info("rollback ConfigMap conflict, retrying",
						"attempt", attempt+1,
						"delay", delay,
						"configmap", p.config.Mimir.ConfigMapName)
					time.Sleep(delay)
					continue
				}
				// Max retries exceeded
				metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("rollback-error")
				return fmt.Errorf("failed to rollback ConfigMap after %d retries due to conflicts: %w", maxRetries, err)
			}
			// Non-conflict error, return immediately
			metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("rollback-error")
			return fmt.Errorf("failed to rollback ConfigMap: %w", err)
		}

		// Success - break out of retry loop
		break
	}

	// Log rollback to audit trail
	if p.auditLog != nil {
		entry := &auditlog.AuditEntry{
			Timestamp: time.Now(),
			Action:    "rollback",
			Reason:    "manual-rollback",
			Changes:   map[string]interface{}{"rollback": "restored from backup"},
		}
		if err := p.auditLog.LogEntry(entry); err != nil {
			// Audit logging failures should not interrupt the rollback operation
			p.log.Error(err, "failed to log audit entry for rollback (audit failure is non-critical)",
				"action", entry.Action,
				"reason", entry.Reason)
		}
	}

	metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("rollback-success")
	p.log.Info("successfully rolled back ConfigMap changes")

	return nil
}

// GetCurrentLimits retrieves the current limits from the ConfigMap
func (p *ConfigMapPatcher) GetCurrentLimits(ctx context.Context) (map[string]*analyzer.TenantLimits, error) {
	currentConfigMap, err := p.getCurrentConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current ConfigMap: %w", err)
	}

	overrides, err := p.parseOverrides(currentConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overrides: %w", err)
	}

	return p.parseCurrentLimits(overrides), nil
}

// Helper methods

func (p *ConfigMapPatcher) getCurrentConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      p.config.Mimir.ConfigMapName,
		Namespace: p.config.Mimir.Namespace,
	}, configMap)

	if apierrors.IsNotFound(err) {
		// Create empty ConfigMap if it doesn't exist
		return p.createInitialConfigMap(ctx)
	}

	return configMap, err
}

func (p *ConfigMapPatcher) createInitialConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.config.Mimir.ConfigMapName,
			Namespace: p.config.Mimir.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "mimir",
				"app.kubernetes.io/component":  "runtime-overrides",
				"app.kubernetes.io/managed-by": "mimir-limit-optimizer",
			},
		},
		Data: map[string]string{
			"overrides.yaml": "overrides: {}\n",
		},
	}

	if err := p.client.Create(ctx, configMap); err != nil {
		return nil, fmt.Errorf("failed to create initial ConfigMap: %w", err)
	}

	p.log.Info("created initial ConfigMap", "name", configMap.Name, "namespace", configMap.Namespace)
	return configMap, nil
}

func (p *ConfigMapPatcher) parseOverrides(configMap *corev1.ConfigMap) (map[string]interface{}, error) {
	overridesYAML, exists := configMap.Data["overrides.yaml"]
	if !exists {
		return map[string]interface{}{
			"overrides": make(map[string]interface{}),
		}, nil
	}

	var overrides map[string]interface{}
	if err := yaml.Unmarshal([]byte(overridesYAML), &overrides); err != nil {
		return nil, fmt.Errorf("failed to unmarshal overrides YAML: %w", err)
	}

	return overrides, nil
}

func (p *ConfigMapPatcher) applyLimitsToOverrides(overrides map[string]interface{}, limits map[string]*analyzer.TenantLimits) map[string]interface{} {
	// Ensure overrides structure exists
	if overrides["overrides"] == nil {
		overrides["overrides"] = make(map[string]interface{})
	}

	tenantOverrides, ok := overrides["overrides"].(map[string]interface{})
	if !ok {
		tenantOverrides = make(map[string]interface{})
		overrides["overrides"] = tenantOverrides
	}

	// Apply limits for each tenant
	for tenant, tenantLimits := range limits {
		// Filter tenants based on configuration
		if p.shouldSkipTenant(tenant) {
			continue
		}

		// PRESERVE EXISTING TENANT CONFIGURATION
		// Get existing tenant config or create new one
		var existingTenantConfig map[string]interface{}
		if existing, exists := tenantOverrides[tenant]; exists {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				existingTenantConfig = existingMap
			} else {
				existingTenantConfig = make(map[string]interface{})
			}
		} else {
			existingTenantConfig = make(map[string]interface{})
		}

		// MERGE NEW LIMITS WITH EXISTING LIMITS (don't replace!)
		updatedLimits := make([]string, 0)
		hasUpdates := false

		// Apply all configured dynamic limits
		for limitName, limitValue := range tenantLimits.Limits {
			// Check if this limit is enabled in configuration
			if limitDef, exists := p.config.DynamicLimits.LimitDefinitions[limitName]; exists && limitDef.Enabled {
				// Apply the limit value based on type
				if limitValue != nil && !p.isZeroValue(limitValue) {
					// Check if this is actually a change
					if existingValue, hadExisting := existingTenantConfig[limitName]; !hadExisting || existingValue != limitValue {
						existingTenantConfig[limitName] = limitValue
						updatedLimits = append(updatedLimits, limitName)
						hasUpdates = true
					}
				}
			}
		}

		// ADD METADATA COMMENTS if there were updates
		if hasUpdates {
			// Add metadata about the optimization
			timestamp := time.Now().Format("2006-01-02T15:04:05Z07:00")
			existingTenantConfig["# mimir-limit-optimizer"] = map[string]interface{}{
				"last_updated":    timestamp,
				"updated_limits":  updatedLimits,
				"reason":          tenantLimits.Reason,
				"source":          tenantLimits.Source,
			}

			// Update the tenant configuration (preserving all existing limits)
			tenantOverrides[tenant] = existingTenantConfig
			
			p.log.V(1).Info("updated tenant limits while preserving existing configuration",
				"tenant", tenant,
				"updated_limits", updatedLimits,
				"existing_limits_preserved", len(existingTenantConfig)-1, // -1 for metadata
				"timestamp", timestamp)
		}
	}

	return overrides
}

// isZeroValue checks if a value is considered zero/empty for its type
func (p *ConfigMapPatcher) isZeroValue(value interface{}) bool {
	switch v := value.(type) {
	case float64:
		return v <= 0
	case int64:
		return v <= 0
	case string:
		return v == "" || v == "0s"
	default:
		return false
	}
}

func (p *ConfigMapPatcher) updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap, overrides map[string]interface{}) error {
	// Convert overrides back to YAML
	overridesYAML, err := yaml.Marshal(overrides)
	if err != nil {
		return fmt.Errorf("failed to marshal overrides to YAML: %w", err)
	}

	// Update ConfigMap data
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	configMap.Data["overrides.yaml"] = string(overridesYAML)

	// Add labels for tracking
	if configMap.Labels == nil {
		configMap.Labels = make(map[string]string)
	}
	// Use Unix timestamp as it contains only digits and is Kubernetes label-safe
	configMap.Labels["mimir-limit-optimizer/last-update"] = strconv.FormatInt(time.Now().Unix(), 10)

	return p.client.Update(ctx, configMap)
}

func (p *ConfigMapPatcher) createBackup(configMap *corev1.ConfigMap) {
	p.lastBackup = configMap.DeepCopy()
}

func (p *ConfigMapPatcher) shouldSkipTenant(tenant string) bool {
	// Check skip list
	for _, pattern := range p.config.TenantScoping.SkipList {
		if p.matchPattern(tenant, pattern) {
			return true
		}
	}

	// Check include list (if specified, only include matching tenants)
	if len(p.config.TenantScoping.IncludeList) > 0 {
		for _, pattern := range p.config.TenantScoping.IncludeList {
			if p.matchPattern(tenant, pattern) {
				return false
			}
		}
		return true // Not in include list
	}

	return false
}

func (p *ConfigMapPatcher) matchPattern(tenant, pattern string) bool {
	if p.config.TenantScoping.UseRegex {
		// TODO: Implement regex matching
		return strings.Contains(tenant, strings.Trim(pattern, "*"))
	}

	// Simple glob matching
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(tenant, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(tenant, suffix)
	}
	return tenant == pattern
}

func (p *ConfigMapPatcher) triggerRollout(ctx context.Context) error {
	for _, component := range p.config.Mimir.RolloutComponents {
		if err := p.restartDeployment(ctx, component); err != nil {
			p.log.Error(err, "failed to restart component", "component", component)
		}
	}
	return nil
}

func (p *ConfigMapPatcher) restartDeployment(ctx context.Context, deploymentName string) error {
	// Retry logic with exponential backoff for conflict resolution
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get the deployment (fresh read each time)
		deployment, err := p.kubeClient.AppsV1().Deployments(p.config.Mimir.Namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment %s: %w", deploymentName, err)
		}

		// Add restart annotation
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		// Use Unix timestamp for annotations as well to ensure consistency
		deployment.Spec.Template.Annotations["mimir-limit-optimizer/restarted-at"] = strconv.FormatInt(time.Now().Unix(), 10)

		// Update deployment
		_, err = p.kubeClient.AppsV1().Deployments(p.config.Mimir.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			// Check if it's a conflict error (using apierrors works for all K8s API errors)
			if apierrors.IsConflict(err) {
				if attempt < maxRetries-1 {
					// Wait with exponential backoff before retrying
					delay := time.Duration(1<<attempt) * baseDelay
					p.log.V(1).Info("deployment update conflict, retrying",
						"attempt", attempt+1,
						"delay", delay,
						"deployment", deploymentName)
					time.Sleep(delay)
					continue
				}
				// Max retries exceeded
				return fmt.Errorf("failed to update deployment %s after %d retries due to conflicts: %w", deploymentName, maxRetries, err)
			}
			// Non-conflict error, return immediately
			return fmt.Errorf("failed to update deployment %s: %w", deploymentName, err)
		}

		// Success - break out of retry loop
		break
	}

	p.log.Info("triggered rollout for component", "component", deploymentName)
	return nil
}

func (p *ConfigMapPatcher) logChanges(oldOverrides, newOverrides map[string]interface{}, limits map[string]*analyzer.TenantLimits) {
	if p.auditLog == nil {
		return
	}

	// Extract old tenant overrides (new values are in the limits parameter)
	oldTenantOverrides := make(map[string]interface{})
	
	if oldOverrides["overrides"] != nil {
		if tenantOverrides, ok := oldOverrides["overrides"].(map[string]interface{}); ok {
			oldTenantOverrides = tenantOverrides
		}
	}

	for tenant, limit := range limits {
		// Extract old limits for this tenant (new limits are in the limit.Limits map)
		oldTenantLimits := make(map[string]interface{})
		
		if oldTenantConfig, exists := oldTenantOverrides[tenant]; exists {
			if oldLimits, ok := oldTenantConfig.(map[string]interface{}); ok {
				oldTenantLimits = oldLimits
			}
		}
		
		// Build proper old vs new audit log entries
		oldValues := make(map[string]interface{})
		newValues := make(map[string]interface{})
		
		// Check each limit that was updated
		for limitName, limitValue := range limit.Limits {
			// Only log enabled limits
			if limitDef, exists := p.config.DynamicLimits.LimitDefinitions[limitName]; exists && limitDef.Enabled {
				// Get old value
				if oldValue, hadOld := oldTenantLimits[limitName]; hadOld {
					oldValues[limitName] = oldValue
				} else {
					oldValues[limitName] = nil // New limit
				}
				
				// Get new value
				newValues[limitName] = limitValue
			}
		}
		
		// Create audit entry using the enhanced format
		entry := auditlog.NewLimitUpdateEntry(tenant, limit.Reason, oldValues, newValues)
		entry.Source = limit.Source
		entry.Component = "mimir-limit-optimizer"
		
		if err := p.auditLog.LogEntry(entry); err != nil {
			// Audit logging failures should not interrupt the main operation
			// Log the error but continue processing other tenants
			p.log.Error(err, "failed to log audit entry for tenant limits (audit failure is non-critical)", 
				"tenant", tenant,
				"action", entry.Action,
				"reason", entry.Reason)
		}
		metrics.TenantMetricsInstance.IncTenantLimitsUpdated(tenant, limit.Reason)
	}
}

func (p *ConfigMapPatcher) parseCurrentLimits(overrides map[string]interface{}) map[string]*analyzer.TenantLimits {
	limits := make(map[string]*analyzer.TenantLimits)

	tenantOverrides, ok := overrides["overrides"].(map[string]interface{})
	if !ok {
		return limits
	}

	for tenant, tenantConfig := range tenantOverrides {
		tenantLimits, ok := tenantConfig.(map[string]interface{})
		if !ok {
			continue
		}

		limit := &analyzer.TenantLimits{
			Tenant:      tenant,
			Limits:      make(map[string]interface{}),
			LastUpdated: time.Now(),
			Source:      "current-configmap",
		}

		// Parse all configured dynamic limits
		for limitName, limitDef := range p.config.DynamicLimits.LimitDefinitions {
			if limitDef.Enabled {
				if val, exists := tenantLimits[limitName]; exists {
					// Convert value based on limit type
					convertedVal, err := p.convertLimitValue(val, limitDef.Type)
					if err != nil {
						p.log.Error(err, "failed to convert limit value", 
							"tenant", tenant, "limit", limitName, "value", val, "type", limitDef.Type)
						continue
					}
					limit.Limits[limitName] = convertedVal
				}
			}
		}

		if len(limit.Limits) > 0 {
			limits[tenant] = limit
		}
	}

	return limits
}

// convertLimitValue converts a value to the appropriate type based on limit definition
func (p *ConfigMapPatcher) convertLimitValue(value interface{}, limitType string) (interface{}, error) {
	switch limitType {
	case "rate", "count", "size", "percentage":
		switch v := value.(type) {
		case float64:
			return v, nil
		case int64:
			return float64(v), nil
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				return parsed, nil
			}
			return nil, fmt.Errorf("cannot convert string %q to numeric value", v)
		default:
			return nil, fmt.Errorf("unsupported value type %T for numeric limit", value)
		}
	case "duration":
		switch v := value.(type) {
		case string:
			return v, nil // Keep as string for duration values
		default:
			return fmt.Sprintf("%v", v), nil
		}
	default:
		return value, nil // Return as-is for unknown types
	}
}

func (p *ConfigMapPatcher) countChanges(old, new map[string]interface{}) int {
	// Simple implementation - count number of tenants being updated
	newOverrides, ok := new["overrides"].(map[string]interface{})
	if !ok {
		return 0
	}
	return len(newOverrides)
}

func copyOverrides(overrides map[string]interface{}) map[string]interface{} {
	// Deep copy implementation
	result := make(map[string]interface{})
	for k, v := range overrides {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = copyOverrides(val)
		default:
			result[k] = val
		}
	}
	return result
}

// NewPatcher creates the appropriate patcher based on configuration
func NewPatcher(c client.Client, kubeClient kubernetes.Interface, cfg *config.Config, auditLogger auditlog.AuditLogger, log logr.Logger) Patcher {
	return NewConfigMapPatcher(c, kubeClient, cfg, auditLogger, log)
} 