package patcher

import (
	"context"
	"fmt"
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

// ApplyLimits applies the calculated limits to the Mimir runtime overrides ConfigMap
func (p *ConfigMapPatcher) ApplyLimits(ctx context.Context, limits map[string]*analyzer.TenantLimits) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ConfigMapMetricsInstance.ObserveConfigMapUpdateDuration("success", duration)
	}()

	// Get current ConfigMap
	currentConfigMap, err := p.getCurrentConfigMap(ctx)
	if err != nil {
		metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
		return fmt.Errorf("failed to get current ConfigMap: %w", err)
	}

	// Create backup
	p.createBackup(currentConfigMap)

	// Parse current overrides
	currentOverrides, err := p.parseOverrides(currentConfigMap)
	if err != nil {
		metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
		return fmt.Errorf("failed to parse current overrides: %w", err)
	}

	// Apply new limits
	updatedOverrides := p.applyLimitsToOverrides(currentOverrides, limits)

	// Update ConfigMap
	if err := p.updateConfigMap(ctx, currentConfigMap, updatedOverrides); err != nil {
		metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("error")
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	// Log changes to audit trail
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

// RollbackChanges rolls back to the previous configuration
func (p *ConfigMapPatcher) RollbackChanges(ctx context.Context) error {
	if p.lastBackup == nil {
		return fmt.Errorf("no backup available for rollback")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ConfigMapMetricsInstance.ObserveConfigMapUpdateDuration("rollback", duration)
	}()

	// Get current ConfigMap
	currentConfigMap, err := p.getCurrentConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current ConfigMap for rollback: %w", err)
	}

	// Restore data from backup
	currentConfigMap.Data = p.lastBackup.Data
	
	if err := p.client.Update(ctx, currentConfigMap); err != nil {
		metrics.ConfigMapMetricsInstance.IncConfigMapUpdates("rollback-error")
		return fmt.Errorf("failed to rollback ConfigMap: %w", err)
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
			p.log.Error(err, "failed to log audit entry for rollback")
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
	for tenant, limits := range limits {
		// Filter tenants based on configuration
		if p.shouldSkipTenant(tenant) {
			continue
		}

		tenantLimits := make(map[string]interface{})

		if limits.IngestionRate > 0 {
			tenantLimits["ingestion_rate"] = limits.IngestionRate
		}
		if limits.IngestionBurst > 0 {
			tenantLimits["ingestion_burst_size"] = limits.IngestionBurst
		}
		if limits.MaxSeries > 0 {
			tenantLimits["max_global_series_per_user"] = limits.MaxSeries
		}
		if limits.MaxSamplesPerQuery > 0 {
			tenantLimits["max_samples_per_query"] = limits.MaxSamplesPerQuery
		}
		if limits.MaxQueryLookback > 0 {
			tenantLimits["max_query_lookback"] = limits.MaxQueryLookback.String()
		}

		// Only update if there are actual limits to set
		if len(tenantLimits) > 0 {
			tenantOverrides[tenant] = tenantLimits
		}
	}

	return overrides
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
	configMap.Labels["mimir-limit-optimizer/last-update"] = time.Now().Format(time.RFC3339)

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
	// Get the deployment
	deployment, err := p.kubeClient.AppsV1().Deployments(p.config.Mimir.Namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", deploymentName, err)
	}

	// Add restart annotation
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["mimir-limit-optimizer/restarted-at"] = time.Now().Format(time.RFC3339)

	// Update deployment
	_, err = p.kubeClient.AppsV1().Deployments(p.config.Mimir.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment %s: %w", deploymentName, err)
	}

	p.log.Info("triggered rollout for component", "component", deploymentName)
	return nil
}



func (p *ConfigMapPatcher) logChanges(oldOverrides, newOverrides map[string]interface{}, limits map[string]*analyzer.TenantLimits) {
	if p.auditLog == nil {
		return
	}

	for tenant, limit := range limits {
		entry := &auditlog.AuditEntry{
			Timestamp: time.Now(),
			Tenant:    tenant,
			Action:    "update-limits",
			Reason:    limit.Reason,
			Changes: map[string]interface{}{
				"ingestion_rate":          limit.IngestionRate,
				"ingestion_burst":         limit.IngestionBurst,
				"max_series":              limit.MaxSeries,
				"max_samples_per_query":   limit.MaxSamplesPerQuery,
			},
		}
		
		if err := p.auditLog.LogEntry(entry); err != nil {
			p.log.Error(err, "failed to log audit entry for tenant limits", "tenant", tenant)
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
			LastUpdated: time.Now(),
			Source:      "current-configmap",
		}

		if val, ok := tenantLimits["ingestion_rate"].(float64); ok {
			limit.IngestionRate = val
		}
		if val, ok := tenantLimits["ingestion_burst_size"].(float64); ok {
			limit.IngestionBurst = val
		}
		if val, ok := tenantLimits["max_global_series_per_user"].(float64); ok {
			limit.MaxSeries = val
		}
		if val, ok := tenantLimits["max_samples_per_query"].(float64); ok {
			limit.MaxSamplesPerQuery = val
		}

		limits[tenant] = limit
	}

	return limits
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