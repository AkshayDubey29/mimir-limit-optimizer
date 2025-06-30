package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeCostViolation     AlertType = "cost_violation"
	AlertTypeCircuitBreaker    AlertType = "circuit_breaker"
	AlertTypePanicMode         AlertType = "panic_mode"
	AlertTypeEmergency         AlertType = "emergency"
	AlertTypeRecovery          AlertType = "recovery"
	AlertTypeRecommendation    AlertType = "recommendation"
	AlertTypeHealthCheck       AlertType = "health_check"
	AlertTypeConfigurationError AlertType = "configuration_error"
)

// Priority levels for alerts
type Priority string

const (
	PriorityP0 Priority = "P0" // Critical - Panic mode, system failure
	PriorityP1 Priority = "P1" // High - Emergency limits, circuit breaker
	PriorityP2 Priority = "P2" // Medium - Budget violations, high usage
	PriorityP3 Priority = "P3" // Low - Recommendations, info
)

// Alert represents an alert to be sent
type Alert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Priority    Priority               `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Tenant      string                 `json:"tenant,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Channels    []string               `json:"channels"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAttempt time.Time              `json:"last_attempt"`
}

// Channel represents an alerting channel
type Channel interface {
	Name() string
	Send(ctx context.Context, alert *Alert) error
	IsHealthy() bool
	GetConfiguration() interface{}
	ValidateConfiguration() error
}

// CircuitBreakerState represents the state of a channel's circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// ChannelCircuitBreaker manages circuit breaker for individual channels
type ChannelCircuitBreaker struct {
	mu                sync.RWMutex
	state             CircuitBreakerState
	failureCount      int
	lastFailureTime   time.Time
	failureThreshold  int
	recoveryTimeout   time.Duration
	halfOpenMaxCalls  int
	halfOpenCalls     int
	halfOpenSuccesses int
}

// NewChannelCircuitBreaker creates a new circuit breaker for a channel
func NewChannelCircuitBreaker(failureThreshold int, recoveryTimeout time.Duration) *ChannelCircuitBreaker {
	return &ChannelCircuitBreaker{
		state:             CircuitBreakerClosed,
		failureThreshold:  failureThreshold,
		recoveryTimeout:   recoveryTimeout,
		halfOpenMaxCalls:  3,
	}
}

// CanSend checks if the circuit breaker allows sending
func (cb *ChannelCircuitBreaker) CanSend() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return time.Since(cb.lastFailureTime) >= cb.recoveryTimeout
	case CircuitBreakerHalfOpen:
		return cb.halfOpenCalls < cb.halfOpenMaxCalls
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *ChannelCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount = 0
	case CircuitBreakerHalfOpen:
		cb.halfOpenSuccesses++
		cb.halfOpenCalls++
		if cb.halfOpenSuccesses >= cb.halfOpenMaxCalls/2 {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
			cb.halfOpenCalls = 0
			cb.halfOpenSuccesses = 0
		}
	}
}

// RecordFailure records a failed operation
func (cb *ChannelCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitBreakerOpen
		}
	case CircuitBreakerHalfOpen:
		cb.state = CircuitBreakerOpen
		cb.halfOpenCalls = 0
		cb.halfOpenSuccesses = 0
	}
}

// GetState returns the current state
func (cb *ChannelCircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// TryHalfOpen attempts to transition to half-open state
func (cb *ChannelCircuitBreaker) TryHalfOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
		cb.state = CircuitBreakerHalfOpen
		cb.halfOpenCalls = 0
		cb.halfOpenSuccesses = 0
		return true
	}
	return false
}

// Manager manages all alerting operations with fault tolerance
type Manager struct {
	config         *config.AlertingConfig
	channels       map[string]Channel
	circuitBreakers map[string]*ChannelCircuitBreaker
	alertQueue     chan *Alert
	retryQueue     chan *Alert
	logger         logr.Logger
	metrics        *metrics.AlertingMetrics
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewManager creates a new alerting manager
func NewManager(config *config.AlertingConfig, logger logr.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Manager{
		config:          config,
		channels:        make(map[string]Channel),
		circuitBreakers: make(map[string]*ChannelCircuitBreaker),
		alertQueue:      make(chan *Alert, 1000), // Buffered queue
		retryQueue:      make(chan *Alert, 500),  // Smaller retry queue
		logger:          logger,
		metrics:         metrics.AlertingMetricsInstance,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start starts the alerting manager
func (m *Manager) Start() error {
	m.logger.Info("Starting alerting manager")
	
	// Initialize channels
	if err := m.initializeChannels(); err != nil {
		m.logger.Error(err, "Failed to initialize channels")
		// Don't return error - continue with available channels
	}
	
	// Start workers
	m.wg.Add(3)
	go m.alertWorker()
	go m.retryWorker()
	go m.healthCheckWorker()
	
	m.logger.Info("Alerting manager started successfully")
	return nil
}

// Stop stops the alerting manager
func (m *Manager) Stop() {
	m.logger.Info("Stopping alerting manager")
	m.cancel()
	
	// Close queues
	close(m.alertQueue)
	close(m.retryQueue)
	
	// Wait for workers to finish
	m.wg.Wait()
	
	m.logger.Info("Alerting manager stopped")
}

// SendAlert sends an alert through all configured channels
func (m *Manager) SendAlert(alert *Alert) {
	// Never block the caller - this is critical for system resilience
	select {
	case m.alertQueue <- alert:
		m.logger.V(1).Info("Alert queued successfully", 
			"alert_id", alert.ID,
			"type", alert.Type,
			"priority", alert.Priority)
	default:
		// Queue is full - log error but don't block
		m.logger.Error(fmt.Errorf("alert queue full"), 
			"Alert queue full, dropping alert",
			"alert_id", alert.ID,
			"type", alert.Type,
			"priority", alert.Priority)
		
		// Increment error metrics
		m.metrics.IncAlertChannelErrors("queue", "queue_full")
	}
}

// SendAlertSync sends an alert synchronously with timeout
func (m *Manager) SendAlertSync(alert *Alert, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()
	
	var lastErr error
	successCount := 0
	
	for _, channelName := range alert.Channels {
		if err := m.sendToChannel(ctx, alert, channelName); err != nil {
			lastErr = err
			m.logger.Error(err, "Failed to send alert to channel",
				"alert_id", alert.ID,
				"channel", channelName,
				"error", err.Error())
		} else {
			successCount++
		}
	}
	
	// Consider successful if at least one channel succeeded
	if successCount > 0 {
		return nil
	}
	
	return lastErr
}

// initializeChannels initializes all configured channels
func (m *Manager) initializeChannels() error {
	var errors []error
	
	// Initialize Slack channel
	if m.config.Slack.Enabled {
		slack := NewSlackChannel(m.config.Slack, m.logger.WithName("slack"))
		if err := slack.ValidateConfiguration(); err != nil {
			m.logger.Error(err, "Invalid Slack configuration")
			m.metrics.IncAlertConfigurationErrors("slack", "invalid_config")
			errors = append(errors, err)
		} else {
			m.channels["slack"] = slack
			m.circuitBreakers["slack"] = NewChannelCircuitBreaker(5, 5*time.Minute)
			m.logger.Info("Slack channel initialized successfully")
		}
	}
	
	// Initialize PagerDuty channel
	if m.config.PagerDuty.Enabled {
		pagerduty := NewPagerDutyChannel(m.config.PagerDuty, m.logger.WithName("pagerduty"))
		if err := pagerduty.ValidateConfiguration(); err != nil {
			m.logger.Error(err, "Invalid PagerDuty configuration")
			m.metrics.IncAlertConfigurationErrors("pagerduty", "invalid_config")
			errors = append(errors, err)
		} else {
			m.channels["pagerduty"] = pagerduty
			m.circuitBreakers["pagerduty"] = NewChannelCircuitBreaker(3, 5*time.Minute)
			m.logger.Info("PagerDuty channel initialized successfully")
		}
	}
	
	// Initialize Email channel
	if m.config.Email.Enabled {
		email := NewEmailChannel(m.config.Email, m.logger.WithName("email"))
		if err := email.ValidateConfiguration(); err != nil {
			m.logger.Error(err, "Invalid Email configuration")
			m.metrics.IncAlertConfigurationErrors("email", "invalid_config")
			errors = append(errors, err)
		} else {
			m.channels["email"] = email
			m.circuitBreakers["email"] = NewChannelCircuitBreaker(5, 10*time.Minute)
			m.logger.Info("Email channel initialized successfully")
		}
	}
	
	// Initialize Webhook channels
	for i, webhookConfig := range m.config.Webhooks {
		if webhookConfig.Enabled {
			webhook := NewWebhookChannel(webhookConfig.Name, webhookConfig, m.logger.WithName("webhook").WithValues("name", webhookConfig.Name))
			if err := webhook.ValidateConfiguration(); err != nil {
				m.logger.Error(err, "Invalid Webhook configuration", "webhook", webhookConfig.Name)
				m.metrics.IncAlertConfigurationErrors("webhook_"+webhookConfig.Name, "invalid_config")
				errors = append(errors, err)
			} else {
				m.channels["webhook_"+webhookConfig.Name] = webhook
				m.circuitBreakers["webhook_"+webhookConfig.Name] = NewChannelCircuitBreaker(5, 5*time.Minute)
				m.logger.Info("Webhook channel initialized successfully", "webhook", webhookConfig.Name, "index", i)
			}
		}
	}
	
	// Log summary
	m.logger.Info("Channel initialization complete",
		"total_channels", len(m.channels),
		"errors", len(errors))
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to initialize %d channels", len(errors))
	}
	
	return nil
}

// alertWorker processes alerts from the main queue
func (m *Manager) alertWorker() {
	defer m.wg.Done()
	
	m.logger.Info("Alert worker started")
	
	for {
		select {
		case alert, ok := <-m.alertQueue:
			if !ok {
				m.logger.Info("Alert queue closed, stopping worker")
				return
			}
			
			m.processAlert(alert)
			
		case <-m.ctx.Done():
			m.logger.Info("Context cancelled, stopping alert worker")
			return
		}
	}
}

// retryWorker processes alerts from the retry queue
func (m *Manager) retryWorker() {
	defer m.wg.Done()
	
	m.logger.Info("Retry worker started")
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case alert, ok := <-m.retryQueue:
			if !ok {
				m.logger.Info("Retry queue closed, stopping worker")
				return
			}
			
			m.processRetry(alert)
			
		case <-ticker.C:
			// Periodic retry processing
			m.processRetryQueue()
			
		case <-m.ctx.Done():
			m.logger.Info("Context cancelled, stopping retry worker")
			return
		}
	}
}

// healthCheckWorker periodically checks channel health
func (m *Manager) healthCheckWorker() {
	defer m.wg.Done()
	
	m.logger.Info("Health check worker started")
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.performHealthChecks()
			
		case <-m.ctx.Done():
			m.logger.Info("Context cancelled, stopping health check worker")
			return
		}
	}
}

// processAlert processes a single alert
func (m *Manager) processAlert(alert *Alert) {
	alert.LastAttempt = time.Now()
	
	m.logger.Info("Processing alert",
		"alert_id", alert.ID,
		"type", alert.Type,
		"priority", alert.Priority,
		"channels", alert.Channels)
	
	// Send to all configured channels
	var failedChannels []string
	for _, channelName := range alert.Channels {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		err := m.sendToChannel(ctx, alert, channelName)
		cancel()
		
		if err != nil {
			failedChannels = append(failedChannels, channelName)
			m.logger.Error(err, "Failed to send alert to channel",
				"alert_id", alert.ID,
				"channel", channelName)
		}
	}
	
	// Handle failed channels
	if len(failedChannels) > 0 && alert.RetryCount < alert.MaxRetries {
		alert.RetryCount++
		alert.Channels = failedChannels
		
		// Add to retry queue with exponential backoff
		go func() {
			backoffDuration := time.Duration(alert.RetryCount*alert.RetryCount) * time.Second
			time.Sleep(backoffDuration)
			
			select {
			case m.retryQueue <- alert:
				m.logger.V(1).Info("Alert added to retry queue",
					"alert_id", alert.ID,
					"retry_count", alert.RetryCount,
					"backoff", backoffDuration)
			default:
				m.logger.Error(fmt.Errorf("retry queue full"), 
					"Failed to add alert to retry queue",
					"alert_id", alert.ID)
			}
		}()
	}
}

// sendToChannel sends an alert to a specific channel
func (m *Manager) sendToChannel(ctx context.Context, alert *Alert, channelName string) error {
	startTime := time.Now()
	
	// Update queue size metric
	m.metrics.SetAlertQueueSize(channelName, float64(len(m.alertQueue)))
	
	// Check if channel exists
	channel, exists := m.channels[channelName]
	if !exists {
		m.metrics.IncAlertChannelErrors(channelName, "channel_not_found")
		return fmt.Errorf("channel %s not found", channelName)
	}
	
	// Check circuit breaker
	cb := m.circuitBreakers[channelName]
	if cb != nil && !cb.CanSend() {
		m.metrics.IncAlertChannelErrors(channelName, "circuit_breaker_open")
		m.metrics.SetAlertChannelCircuitBreakerState(channelName, float64(cb.GetState()))
		return fmt.Errorf("circuit breaker open for channel %s", channelName)
	}
	
	// Transition to half-open if possible
	if cb != nil && cb.GetState() == CircuitBreakerOpen {
		cb.TryHalfOpen()
	}
	
	// Send the alert
	err := channel.Send(ctx, alert)
	duration := time.Since(startTime).Seconds()
	
	// Record metrics
	if err != nil {
		m.metrics.IncAlertDeliveryTotal(channelName, string(alert.Type), "failure")
		m.metrics.IncAlertChannelErrors(channelName, "send_failed")
		m.metrics.SetAlertChannelHealth(channelName, 0)
		if cb != nil {
			cb.RecordFailure()
		}
	} else {
		m.metrics.IncAlertDeliveryTotal(channelName, string(alert.Type), "success")
		m.metrics.SetAlertChannelHealth(channelName, 1)
		m.metrics.SetLastSuccessfulAlertTime(channelName, float64(time.Now().Unix()))
		if cb != nil {
			cb.RecordSuccess()
		}
	}
	
	m.metrics.ObserveAlertDeliveryDuration(channelName, duration)
	if cb != nil {
		m.metrics.SetAlertChannelCircuitBreakerState(channelName, float64(cb.GetState()))
	}
	
	return err
}

// processRetry processes a retry alert
func (m *Manager) processRetry(alert *Alert) {
	m.logger.Info("Processing retry alert",
		"alert_id", alert.ID,
		"retry_count", alert.RetryCount,
		"channels", alert.Channels)
	
	m.metrics.IncAlertRetryAttempts("all", string(alert.Type))
	m.processAlert(alert)
}

// processRetryQueue processes periodic retry queue
func (m *Manager) processRetryQueue() {
	// This is a placeholder for more sophisticated retry logic
	// In a real implementation, you might want to:
	// 1. Persist failed alerts to disk/database
	// 2. Implement priority-based retry
	// 3. Implement dead letter queue for permanently failed alerts
}

// performHealthChecks performs health checks on all channels
func (m *Manager) performHealthChecks() {
	m.logger.V(1).Info("Performing channel health checks")
	
	for name, channel := range m.channels {
		healthy := channel.IsHealthy()
		m.metrics.SetAlertChannelHealth(name, map[bool]float64{true: 1, false: 0}[healthy])
		
		if !healthy {
			m.logger.Info("Channel health check failed", "channel", name)
		}
	}
}

// GetChannelStatus returns the status of all channels
func (m *Manager) GetChannelStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	status := make(map[string]interface{})
	
	for name, channel := range m.channels {
		cb := m.circuitBreakers[name]
		channelStatus := map[string]interface{}{
			"healthy":                channel.IsHealthy(),
			"circuit_breaker_state":  cb.GetState(),
			"configuration":          channel.GetConfiguration(),
		}
		status[name] = channelStatus
	}
	
	return status
}

// CreateAlert creates a new alert with default values
func CreateAlert(alertType AlertType, priority Priority, title, message string) *Alert {
	return &Alert{
		ID:          fmt.Sprintf("%s-%d", alertType, time.Now().UnixNano()),
		Type:        alertType,
		Priority:    priority,
		Title:       title,
		Message:     message,
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
		MaxRetries:  3,
		Details:     make(map[string]interface{}),
	}
}

// Helper functions for creating specific alert types

// CreateCostViolationAlert creates a cost violation alert
func CreateCostViolationAlert(tenant string, currentCost, budgetLimit float64, violationLevel string) *Alert {
	alert := CreateAlert(AlertTypeCostViolation, PriorityP2, 
		fmt.Sprintf("Budget violation detected for tenant %s", tenant),
		fmt.Sprintf("Tenant %s has exceeded budget limit. Current cost: $%.2f, Budget: $%.2f", 
			tenant, currentCost, budgetLimit))
	
	alert.Tenant = tenant
	alert.Details = map[string]interface{}{
		"current_cost":     currentCost,
		"budget_limit":     budgetLimit,
		"violation_level":  violationLevel,
		"percentage":       (currentCost / budgetLimit) * 100,
	}
	
	return alert
}

// CreatePanicModeAlert creates a panic mode alert
func CreatePanicModeAlert(reason string, details map[string]interface{}) *Alert {
	alert := CreateAlert(AlertTypePanicMode, PriorityP0,
		"PANIC MODE ACTIVATED",
		fmt.Sprintf("Panic mode has been activated due to: %s", reason))
	
	alert.Details = details
	
	return alert
}

// CreateCircuitBreakerAlert creates a circuit breaker alert
func CreateCircuitBreakerAlert(tenant string, blastType string, details map[string]interface{}) *Alert {
	alert := CreateAlert(AlertTypeCircuitBreaker, PriorityP1,
		fmt.Sprintf("Circuit breaker activated for tenant %s", tenant),
		fmt.Sprintf("Circuit breaker has been activated due to %s blast detection", blastType))
	
	alert.Tenant = tenant
	alert.Details = details
	
	return alert
} 