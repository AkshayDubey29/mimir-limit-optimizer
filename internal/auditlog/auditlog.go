package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Tenant      string                 `json:"tenant,omitempty"`
	Action      string                 `json:"action"`
	Reason      string                 `json:"reason"`
	Changes     map[string]interface{} `json:"changes"`
	OldValues   map[string]interface{} `json:"old_values,omitempty"`
	NewValues   map[string]interface{} `json:"new_values,omitempty"`
	Source      string                 `json:"source"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Component   string                 `json:"component"`
	User        string                 `json:"user,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// AuditLogger interface defines methods for audit logging
type AuditLogger interface {
	LogEntry(entry *AuditEntry) error
	GetEntries(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error)
	GetEntry(ctx context.Context, id string) (*AuditEntry, error)
	PurgeOldEntries(ctx context.Context, olderThan time.Time) error
	Close() error
}

// AuditFilter defines filtering criteria for audit entries
type AuditFilter struct {
	Tenant    string
	Action    string
	StartTime *time.Time
	EndTime   *time.Time
	Success   *bool
	Limit     int
	Offset    int
}

// MemoryAuditLogger implements audit logging in memory
type MemoryAuditLogger struct {
	entries    []*AuditEntry
	maxEntries int
	mu         sync.RWMutex
	log        logr.Logger
}

// NewMemoryAuditLogger creates a new in-memory audit logger
func NewMemoryAuditLogger(maxEntries int, log logr.Logger) *MemoryAuditLogger {
	return &MemoryAuditLogger{
		entries:    make([]*AuditEntry, 0, maxEntries),
		maxEntries: maxEntries,
		log:        log,
	}
}

// LogEntry logs an audit entry
func (m *MemoryAuditLogger) LogEntry(entry *AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}

	// Set defaults
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Component == "" {
		entry.Component = "mimir-limit-optimizer"
	}
	if entry.Source == "" {
		entry.Source = "controller"
	}

	// Add entry
	m.entries = append(m.entries, entry)

	// Trim if necessary
	if len(m.entries) > m.maxEntries {
		// Remove oldest entries
		removeCount := len(m.entries) - m.maxEntries
		m.entries = m.entries[removeCount:]
	}

	m.log.Info("audit entry logged",
		"id", entry.ID,
		"tenant", entry.Tenant,
		"action", entry.Action,
		"reason", entry.Reason,
		"success", entry.Success)

	return nil
}

// GetEntries retrieves audit entries with optional filtering
func (m *MemoryAuditLogger) GetEntries(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []*AuditEntry

	for _, entry := range m.entries {
		if m.matchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}

	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(filtered) {
			filtered = filtered[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(filtered) {
			filtered = filtered[:filter.Limit]
		}
	}

	return filtered, nil
}

// GetEntry retrieves a specific audit entry by ID
func (m *MemoryAuditLogger) GetEntry(ctx context.Context, id string) (*AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("audit entry not found: %s", id)
}

// PurgeOldEntries removes entries older than the specified time
func (m *MemoryAuditLogger) PurgeOldEntries(ctx context.Context, olderThan time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var filtered []*AuditEntry
	purgedCount := 0

	for _, entry := range m.entries {
		if entry.Timestamp.After(olderThan) {
			filtered = append(filtered, entry)
		} else {
			purgedCount++
		}
	}

	m.entries = filtered
	m.log.Info("purged old audit entries", "count", purgedCount, "older_than", olderThan)

	return nil
}

// Close closes the audit logger
func (m *MemoryAuditLogger) Close() error {
	return nil
}

// matchesFilter checks if an entry matches the filter criteria
func (m *MemoryAuditLogger) matchesFilter(entry *AuditEntry, filter *AuditFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Tenant != "" && entry.Tenant != filter.Tenant {
		return false
	}

	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}

	if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
		return false
	}

	if filter.Success != nil && entry.Success != *filter.Success {
		return false
	}

	return true
}

// ConfigMapAuditLogger implements audit logging using ConfigMaps
type ConfigMapAuditLogger struct {
	client        client.Client
	configMapName string
	namespace     string
	maxEntries    int
	log           logr.Logger
}

// NewConfigMapAuditLogger creates a new ConfigMap-based audit logger
func NewConfigMapAuditLogger(c client.Client, configMapName, namespace string, maxEntries int, log logr.Logger) *ConfigMapAuditLogger {
	return &ConfigMapAuditLogger{
		client:        c,
		configMapName: configMapName,
		namespace:     namespace,
		maxEntries:    maxEntries,
		log:           log,
	}
}

// LogEntry logs an audit entry to a ConfigMap with retry logic for conflict resolution
func (c *ConfigMapAuditLogger) LogEntry(entry *AuditEntry) error {
	ctx := context.Background()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}

	// Set defaults
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Component == "" {
		entry.Component = "mimir-limit-optimizer"
	}

	// Retry logic with exponential backoff for conflict resolution
	maxRetries := 5
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		configMap, err := c.getOrCreateConfigMap(ctx)
		if err != nil {
			return fmt.Errorf("failed to get or create audit ConfigMap: %w", err)
		}

		entries, err := c.parseEntries(configMap)
		if err != nil {
			return fmt.Errorf("failed to parse existing audit entries: %w", err)
		}

		// Add new entry
		entries = append(entries, entry)

		// Apply retention policies before saving
		entries, shouldEmergencyCleanup := c.applyRetentionPolicies(entries)

		// Log emergency cleanup if it occurred
		if shouldEmergencyCleanup {
			c.log.Info("emergency cleanup triggered during audit entry addition",
				"reason", "retention_policies_exceeded",
				"remaining_entries", len(entries))
		}

		// Try to update ConfigMap with retained entries
		if err := c.updateConfigMap(ctx, configMap, entries); err != nil {
			// Check if it's a conflict error
			if apierrors.IsConflict(err) {
				if attempt < maxRetries-1 {
					// Wait with exponential backoff before retrying
					delay := time.Duration(1<<attempt) * baseDelay
					c.log.V(1).Info("audit ConfigMap conflict during LogEntry, retrying",
						"attempt", attempt+1,
						"delay", delay,
						"entry_id", entry.ID)
					time.Sleep(delay)
					continue
				}
				// Max retries exceeded
				c.log.Error(err, "failed to log audit entry after retries",
					"attempts", maxRetries,
					"entry_id", entry.ID)
				return fmt.Errorf("failed to log audit entry after %d retries: %w", maxRetries, err)
			}
			// Non-conflict error, return immediately
			return fmt.Errorf("failed to update audit ConfigMap: %w", err)
		}

		// Success
		c.log.V(1).Info("audit entry logged to ConfigMap",
			"id", entry.ID,
			"tenant", entry.Tenant,
			"action", entry.Action,
			"success", entry.Success,
			"attempt", attempt+1,
			"total_entries", len(entries))
		return nil
	}

	// This should never be reached due to the logic above, but just in case
	return fmt.Errorf("unexpected end of retry loop for audit LogEntry")
}

// applyRetentionPolicies applies all retention policies and returns cleaned entries
func (c *ConfigMapAuditLogger) applyRetentionPolicies(entries []*AuditEntry) ([]*AuditEntry, bool) {
	originalCount := len(entries)
	emergencyCleanup := false

	// Get retention config (use defaults if not available)
	retentionPeriod := 7 * 24 * time.Hour  // Default: 7 days
	maxEntries := c.maxEntries             // Use configured maxEntries
	maxSizeBytes := int64(800 * 1024)      // Default: 800KB
	emergencyThreshold := 90.0             // Default: 90%

	// TODO: Get these from config once retention config is passed to the logger
	// For now, using sensible defaults

	// 1. Apply time-based retention
	cutoff := time.Now().Add(-retentionPeriod)
	var timeFiltered []*AuditEntry
	for _, entry := range entries {
		if entry.Timestamp.After(cutoff) {
			timeFiltered = append(timeFiltered, entry)
		}
	}

	// 2. Apply count-based retention (keep most recent entries)
	if len(timeFiltered) > maxEntries {
		// Sort by timestamp (newest first) to keep most recent
		sort.Slice(timeFiltered, func(i, j int) bool {
			return timeFiltered[i].Timestamp.After(timeFiltered[j].Timestamp)
		})
		timeFiltered = timeFiltered[:maxEntries]
		emergencyCleanup = true
	}

	// 3. Apply size-based retention
	sizeFiltered := c.applySizeBasedRetention(timeFiltered, maxSizeBytes, emergencyThreshold)
	if len(sizeFiltered) < len(timeFiltered) {
		emergencyCleanup = true
	}

	cleanedCount := originalCount - len(sizeFiltered)
	if cleanedCount > 0 {
		c.log.V(1).Info("retention policies applied",
			"original_entries", originalCount,
			"cleaned_entries", cleanedCount,
			"remaining_entries", len(sizeFiltered),
			"emergency_cleanup", emergencyCleanup)
	}

	return sizeFiltered, emergencyCleanup
}

// applySizeBasedRetention removes entries to stay under size limit
func (c *ConfigMapAuditLogger) applySizeBasedRetention(entries []*AuditEntry, maxSizeBytes int64, emergencyThreshold float64) []*AuditEntry {
	if len(entries) == 0 {
		return entries
	}

	// Calculate current size
	currentSize := c.calculateEntriesSize(entries)
	if currentSize <= maxSizeBytes {
		return entries // Within limits
	}

	// Sort by timestamp (newest first) to keep most recent entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Remove oldest entries until we're under the limit
	var retained []*AuditEntry
	runningSize := int64(0)
	targetSize := int64(float64(maxSizeBytes) * (emergencyThreshold / 100.0))

	for _, entry := range entries {
		entrySize := c.calculateEntrySize(entry)
		if runningSize+entrySize <= targetSize {
			retained = append(retained, entry)
			runningSize += entrySize
		} else {
			break // Would exceed target size
		}
	}

	c.log.Info("size-based retention applied",
		"original_size_bytes", currentSize,
		"target_size_bytes", targetSize,
		"final_size_bytes", runningSize,
		"entries_removed", len(entries)-len(retained))

	return retained
}

// calculateEntriesSize estimates the size of all entries when marshaled to JSON
func (c *ConfigMapAuditLogger) calculateEntriesSize(entries []*AuditEntry) int64 {
	data, err := json.Marshal(entries)
	if err != nil {
		// Fallback estimation: 500 bytes per entry average
		return int64(len(entries) * 500)
	}
	return int64(len(data))
}

// calculateEntrySize estimates the size of a single entry when marshaled to JSON
func (c *ConfigMapAuditLogger) calculateEntrySize(entry *AuditEntry) int64 {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback estimation: 500 bytes per entry
		return 500
	}
	return int64(len(data))
}

// GetEntries retrieves audit entries from ConfigMap
func (c *ConfigMapAuditLogger) GetEntries(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error) {
	configMap, err := c.getConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit ConfigMap: %w", err)
	}

	entries, err := c.parseEntries(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit entries: %w", err)
	}

	// Apply filtering
	var filtered []*AuditEntry
	for _, entry := range entries {
		if c.matchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}

	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(filtered) {
			filtered = filtered[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(filtered) {
			filtered = filtered[:filter.Limit]
		}
	}

	return filtered, nil
}

// GetEntry retrieves a specific audit entry by ID
func (c *ConfigMapAuditLogger) GetEntry(ctx context.Context, id string) (*AuditEntry, error) {
	entries, err := c.GetEntries(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("audit entry not found: %s", id)
}

// PurgeOldEntries removes old entries from ConfigMap with retry logic for conflict resolution
func (c *ConfigMapAuditLogger) PurgeOldEntries(ctx context.Context, olderThan time.Time) error {
	// Retry logic with exponential backoff for conflict resolution
	maxRetries := 5
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		configMap, err := c.getConfigMap(ctx)
		if err != nil {
			return fmt.Errorf("failed to get audit ConfigMap: %w", err)
		}

		entries, err := c.parseEntries(configMap)
		if err != nil {
			return fmt.Errorf("failed to parse audit entries: %w", err)
		}

		var filtered []*AuditEntry
		purgedCount := 0

		for _, entry := range entries {
			if entry.Timestamp.After(olderThan) {
				filtered = append(filtered, entry)
			} else {
				purgedCount++
			}
		}

		// Try to update ConfigMap
		if err := c.updateConfigMap(ctx, configMap, filtered); err != nil {
			// Check if it's a conflict error
			if apierrors.IsConflict(err) {
				if attempt < maxRetries-1 {
					// Wait with exponential backoff before retrying
					delay := time.Duration(1<<attempt) * baseDelay
					c.log.V(1).Info("audit ConfigMap conflict during purge, retrying",
						"attempt", attempt+1,
						"delay", delay,
						"purge_count", purgedCount)
					time.Sleep(delay)
					continue
				}
				// Max retries exceeded - log but don't fail the operation
				c.log.Error(err, "failed to purge audit ConfigMap after retries",
					"attempts", maxRetries,
					"purge_count", purgedCount)
				return nil
			}
			// Non-conflict error, return immediately
			return fmt.Errorf("failed to update audit ConfigMap: %w", err)
		}

		// Success
		c.log.Info("purged old audit entries from ConfigMap", 
			"count", purgedCount, 
			"older_than", olderThan,
			"attempt", attempt+1)
		return nil
	}

	// This should never be reached due to the logic above, but just in case
	c.log.Error(nil, "unexpected end of retry loop for audit purge")
	return nil
}

// Close closes the audit logger
func (c *ConfigMapAuditLogger) Close() error {
	return nil
}

// Helper methods for ConfigMapAuditLogger

func (c *ConfigMapAuditLogger) getOrCreateConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap, err := c.getConfigMap(ctx)
	if apierrors.IsNotFound(err) {
		return c.createConfigMap(ctx)
	}
	return configMap, err
}

func (c *ConfigMapAuditLogger) getConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := c.client.Get(ctx, types.NamespacedName{
		Name:      c.configMapName,
		Namespace: c.namespace,
	}, configMap)
	return configMap, err
}

func (c *ConfigMapAuditLogger) createConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.configMapName,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "mimir-limit-optimizer",
				"app.kubernetes.io/component":  "audit-log",
				"app.kubernetes.io/managed-by": "mimir-limit-optimizer",
			},
		},
		Data: map[string]string{
			"audit.json": "[]",
		},
	}

	if err := c.client.Create(ctx, configMap); err != nil {
		return nil, fmt.Errorf("failed to create audit ConfigMap: %w", err)
	}

	return configMap, nil
}

func (c *ConfigMapAuditLogger) parseEntries(configMap *corev1.ConfigMap) ([]*AuditEntry, error) {
	auditJSON, exists := configMap.Data["audit.json"]
	if !exists {
		return []*AuditEntry{}, nil
	}

	var entries []*AuditEntry
	if err := json.Unmarshal([]byte(auditJSON), &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit entries: %w", err)
	}

	return entries, nil
}

func (c *ConfigMapAuditLogger) updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap, entries []*AuditEntry) error {
	auditJSON, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entries: %w", err)
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	configMap.Data["audit.json"] = string(auditJSON)

	return c.client.Update(ctx, configMap)
}

func (c *ConfigMapAuditLogger) matchesFilter(entry *AuditEntry, filter *AuditFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Tenant != "" && entry.Tenant != filter.Tenant {
		return false
	}

	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}

	if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
		return false
	}

	if filter.Success != nil && entry.Success != *filter.Success {
		return false
	}

	return true
}

// NewAuditLogger creates the appropriate audit logger based on configuration
func NewAuditLogger(cfg *config.Config, client client.Client, log logr.Logger) AuditLogger {
	if !cfg.AuditLog.Enabled {
		return &NoOpAuditLogger{}
	}

	switch cfg.AuditLog.StorageType {
	case "configmap":
		return NewConfigMapAuditLogger(
			client,
			cfg.AuditLog.ConfigMapName,
			cfg.Mimir.Namespace,
			cfg.AuditLog.MaxEntries,
			log,
		)
	case "memory":
		return NewMemoryAuditLogger(cfg.AuditLog.MaxEntries, log)
	default:
		log.Info("unknown audit log storage type, using memory", "type", cfg.AuditLog.StorageType)
		return NewMemoryAuditLogger(cfg.AuditLog.MaxEntries, log)
	}
}

// NoOpAuditLogger is a no-op implementation of AuditLogger
type NoOpAuditLogger struct{}

func (n *NoOpAuditLogger) LogEntry(entry *AuditEntry) error                              { return nil }
func (n *NoOpAuditLogger) GetEntries(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error) { return []*AuditEntry{}, nil }
func (n *NoOpAuditLogger) GetEntry(ctx context.Context, id string) (*AuditEntry, error) { return nil, fmt.Errorf("not found") }
func (n *NoOpAuditLogger) PurgeOldEntries(ctx context.Context, olderThan time.Time) error { return nil }
func (n *NoOpAuditLogger) Close() error                                                 { return nil }

// Helper functions for creating common audit entries

// NewLimitUpdateEntry creates an audit entry for limit updates
func NewLimitUpdateEntry(tenant, reason string, oldLimits, newLimits map[string]interface{}) *AuditEntry {
	// Create proper old vs new changes map
	changes := make(map[string]interface{})
	
	// Process all new limits
	for key, newValue := range newLimits {
		oldValue, hadOldValue := oldLimits[key]
		
		// Always include the change, even if it's a new limit (old = nil)
		change := map[string]interface{}{
			"new": newValue,
		}
		
		if hadOldValue {
			change["old"] = oldValue
		} else {
			change["old"] = nil // Explicitly show this is a new limit
		}
		
		changes[key] = change
	}
	
	// Also check for removed limits (existed in old but not in new)
	for key, oldValue := range oldLimits {
		if _, existsInNew := newLimits[key]; !existsInNew {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": nil, // Explicitly show this limit was removed
			}
		}
	}
	
	return &AuditEntry{
		Tenant:    tenant,
		Action:    "update-limits",
		Reason:    reason,
		Changes:   changes,     // Now contains proper old vs new comparison
		OldValues: oldLimits,   // Keep for backward compatibility
		NewValues: newLimits,   // Keep for backward compatibility
		Success:   true,
	}
}

// NewErrorEntry creates an audit entry for errors
func NewErrorEntry(tenant, action, reason string, err error) *AuditEntry {
	return &AuditEntry{
		Tenant:  tenant,
		Action:  action,
		Reason:  reason,
		Success: false,
		Error:   err.Error(),
	}
}

// NewSpikeDetectionEntry creates an audit entry for spike detection
func NewSpikeDetectionEntry(tenant, metricName string, oldValue, newValue float64) *AuditEntry {
	return &AuditEntry{
		Tenant: tenant,
		Action: "spike-detected",
		Reason: "automatic-spike-scaling",
		Changes: map[string]interface{}{
			"metric": metricName,
			"value": map[string]interface{}{
				"old": oldValue,
				"new": newValue,
			},
		},
		Success: true,
	}
} 