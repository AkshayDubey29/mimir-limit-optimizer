package alerting

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// SlackChannel implements the Channel interface for Slack
type SlackChannel struct {
	config *config.SlackConfig
	logger logr.Logger
	client *http.Client
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel(config config.SlackConfig, logger logr.Logger) *SlackChannel {
	return &SlackChannel{
		config: &config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SlackChannel) Name() string {
	return "slack"
}

func (s *SlackChannel) ValidateConfiguration() error {
	if !s.config.Enabled {
		return fmt.Errorf("slack channel is disabled")
	}
	if s.config.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL is required")
	}
	if s.config.Channel == "" {
		return fmt.Errorf("slack channel is required")
	}
	return nil
}

func (s *SlackChannel) GetConfiguration() interface{} {
	// Return safe configuration (without sensitive data)
	return map[string]interface{}{
		"enabled":  s.config.Enabled,
		"channel":  s.config.Channel,
		"username": s.config.Username,
		"timeout":  s.config.Timeout,
	}
}

func (s *SlackChannel) IsHealthy() bool {
	// Perform a simple health check without sending an actual alert
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/api.test", nil)
	if err != nil {
		s.logger.V(1).Info("Slack health check failed - request creation", "error", err)
		return false
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.V(1).Info("Slack health check failed - request execution", "error", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	
	return resp.StatusCode == http.StatusOK
}

func (s *SlackChannel) Send(ctx context.Context, alert *Alert) error {
	if !s.config.Enabled {
		return fmt.Errorf("slack channel is disabled")
	}
	
	payload := s.buildSlackPayload(alert)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error(err, "Failed to marshal Slack payload", "alert_id", alert.ID)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Error(err, "Failed to create Slack request", "alert_id", alert.ID)
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	startTime := time.Now()
	resp, err := s.client.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		s.logger.Error(err, "Failed to send Slack alert", 
			"alert_id", alert.ID,
			"duration", duration)
				return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error(fmt.Errorf("slack returned non-200 status"), 
			"Failed to send Slack alert",
			"alert_id", alert.ID,
			"status_code", resp.StatusCode,
			"duration", duration)
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	
	s.logger.Info("Slack alert sent successfully", 
		"alert_id", alert.ID,
		"channel", s.config.Channel,
		"duration", duration)
	
	return nil
}

func (s *SlackChannel) buildSlackPayload(alert *Alert) map[string]interface{} {
	color := s.getColorForPriority(alert.Priority)
	
	attachment := map[string]interface{}{
		"color":     color,
		"title":     alert.Title,
		"text":      alert.Message,
		"timestamp": alert.Timestamp.Unix(),
		"fields": []map[string]interface{}{
			{
				"title": "Priority",
				"value": string(alert.Priority),
				"short": true,
			},
			{
				"title": "Type",
				"value": string(alert.Type),
				"short": true,
			},
		},
	}
	
	if alert.Tenant != "" {
		fields := attachment["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"title": "Tenant",
			"value": alert.Tenant,
			"short": true,
		})
		attachment["fields"] = fields
	}
	
	// Add details as fields
	if len(alert.Details) > 0 {
		fields := attachment["fields"].([]map[string]interface{})
		for key, value := range alert.Details {
			// Use cases.Title instead of deprecated strings.Title
			caser := cases.Title(language.English)
			fields = append(fields, map[string]interface{}{
				"title": caser.String(strings.ReplaceAll(key, "_", " ")),
				"value": fmt.Sprintf("%v", value),
				"short": true,
			})
		}
		attachment["fields"] = fields
	}
	
	payload := map[string]interface{}{
		"channel":     s.config.Channel,
		"username":    s.config.Username,
		"icon_emoji":  ":warning:",
		"attachments": []map[string]interface{}{attachment},
	}
	
	return payload
}

func (s *SlackChannel) getColorForPriority(priority Priority) string {
	switch priority {
	case PriorityP0:
		return "danger"
	case PriorityP1:
		return "warning"
	case PriorityP2:
		return "good"
	case PriorityP3:
		return "#439FE0"
	default:
		return "good"
	}
}

// PagerDutyChannel implements the Channel interface for PagerDuty
type PagerDutyChannel struct {
	config *config.PagerDutyConfig
	logger logr.Logger
	client *http.Client
}

// NewPagerDutyChannel creates a new PagerDuty channel
func NewPagerDutyChannel(config config.PagerDutyConfig, logger logr.Logger) *PagerDutyChannel {
	return &PagerDutyChannel{
		config: &config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *PagerDutyChannel) Name() string {
	return "pagerduty"
}

func (p *PagerDutyChannel) ValidateConfiguration() error {
	if !p.config.Enabled {
		return fmt.Errorf("pagerduty channel is disabled")
	}
	if p.config.IntegrationKey == "" {
		return fmt.Errorf("pagerduty integration key is required")
	}
	return nil
}

func (p *PagerDutyChannel) GetConfiguration() interface{} {
	return map[string]interface{}{
		"enabled": p.config.Enabled,
		"timeout": p.config.Timeout,
	}
}

func (p *PagerDutyChannel) IsHealthy() bool {
	// PagerDuty doesn't have a simple health check endpoint
	// We'll consider it healthy if the configuration is valid
	return p.ValidateConfiguration() == nil
}

func (p *PagerDutyChannel) Send(ctx context.Context, alert *Alert) error {
	if !p.config.Enabled {
		return fmt.Errorf("pagerduty channel is disabled")
	}
	
	// Only send P0 and P1 alerts to PagerDuty
	if alert.Priority != PriorityP0 && alert.Priority != PriorityP1 {
		p.logger.V(1).Info("Skipping non-critical alert for PagerDuty", 
			"alert_id", alert.ID,
			"priority", alert.Priority)
		return nil
	}
	
	payload := p.buildPagerDutyPayload(alert)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error(err, "Failed to marshal PagerDuty payload", "alert_id", alert.ID)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(payloadBytes))
	if err != nil {
		p.logger.Error(err, "Failed to create PagerDuty request", "alert_id", alert.ID)
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	startTime := time.Now()
	resp, err := p.client.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		p.logger.Error(err, "Failed to send PagerDuty alert", 
			"alert_id", alert.ID,
			"duration", duration)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	if resp.StatusCode != http.StatusAccepted {
		p.logger.Error(fmt.Errorf("pagerduty returned non-202 status"), 
			"Failed to send PagerDuty alert",
			"alert_id", alert.ID,
			"status_code", resp.StatusCode,
			"duration", duration)
		return fmt.Errorf("pagerduty returned status %d", resp.StatusCode)
	}
	
	p.logger.Info("PagerDuty alert sent successfully", 
		"alert_id", alert.ID,
		"duration", duration)
	
	return nil
}

func (p *PagerDutyChannel) buildPagerDutyPayload(alert *Alert) map[string]interface{} {
	severity := p.getSeverityForPriority(alert.Priority)
	
	payload := map[string]interface{}{
		"routing_key":  p.config.IntegrationKey,
		"event_action": "trigger",
		"dedup_key":    fmt.Sprintf("mimir-limit-optimizer-%s", alert.ID),
		"payload": map[string]interface{}{
			"summary":        alert.Title,
			"source":         "mimir-limit-optimizer",
			"severity":       severity,
			"component":      "mimir",
			"group":          "limit-optimizer",
			"class":          string(alert.Type),
			"custom_details": alert.Details,
		},
	}
	
	if alert.Tenant != "" {
		payload["payload"].(map[string]interface{})["custom_details"].(map[string]interface{})["tenant"] = alert.Tenant
	}
	
	return payload
}

func (p *PagerDutyChannel) getSeverityForPriority(priority Priority) string {
	switch priority {
	case PriorityP0:
		return "critical"
	case PriorityP1:
		return "error"
	case PriorityP2:
		return "warning"
	case PriorityP3:
		return "info"
	default:
		return "info"
	}
}

// EmailChannel implements the Channel interface for Email
type EmailChannel struct {
	config *config.EmailConfig
	logger logr.Logger
}

// NewEmailChannel creates a new Email channel
func NewEmailChannel(config config.EmailConfig, logger logr.Logger) *EmailChannel {
	return &EmailChannel{
		config: &config,
		logger: logger,
	}
}

func (e *EmailChannel) Name() string {
	return "email"
}

func (e *EmailChannel) ValidateConfiguration() error {
	if !e.config.Enabled {
		return fmt.Errorf("email channel is disabled")
	}
	if e.config.SMTPHost == "" {
		return fmt.Errorf("email SMTP host is required")
	}
	if e.config.SMTPPort == 0 {
		return fmt.Errorf("email SMTP port is required")
	}
	if e.config.From == "" {
		return fmt.Errorf("email from address is required")
	}
	if len(e.config.To) == 0 {
		return fmt.Errorf("email to addresses are required")
	}
	return nil
}

func (e *EmailChannel) GetConfiguration() interface{} {
	return map[string]interface{}{
		"enabled":   e.config.Enabled,
		"smtp_host": e.config.SMTPHost,
		"smtp_port": e.config.SMTPPort,
		"from":      e.config.From,
		"to_count":  len(e.config.To),
		"use_tls":   e.config.UseTLS,
	}
}

func (e *EmailChannel) IsHealthy() bool {
	// Try to connect to SMTP server
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)
	
	var conn interface {
		Close() error
	}
	var err error
	
	if e.config.UseTLS {
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			ServerName: e.config.SMTPHost,
		})
	} else {
		conn, err = smtp.Dial(addr)
	}
	
	if err != nil {
		e.logger.V(1).Info("Email health check failed", "error", err)
		return false
	}
	
	defer func() { _ = conn.Close() }()
	return true
}

func (e *EmailChannel) Send(ctx context.Context, alert *Alert) error {
	if !e.config.Enabled {
		return fmt.Errorf("email channel is disabled")
	}
	
	startTime := time.Now()
	
	// Build email content
	subject := fmt.Sprintf("[%s] %s", alert.Priority, alert.Title)
	body := e.buildEmailBody(alert)
	
	// Create message
	msg := e.buildEmailMessage(subject, body, alert)
	
	// Send email
	err := e.sendEmail(msg)
	duration := time.Since(startTime)
	
	if err != nil {
		e.logger.Error(err, "Failed to send email alert", 
			"alert_id", alert.ID,
			"duration", duration)
		return err
	}
	
	e.logger.Info("Email alert sent successfully", 
		"alert_id", alert.ID,
		"recipients", len(e.config.To),
		"duration", duration)
	
	return nil
}

func (e *EmailChannel) buildEmailBody(alert *Alert) string {
	var body strings.Builder
	
	body.WriteString(fmt.Sprintf("Alert: %s\n", alert.Title))
	body.WriteString(fmt.Sprintf("Priority: %s\n", alert.Priority))
	body.WriteString(fmt.Sprintf("Type: %s\n", alert.Type))
	body.WriteString(fmt.Sprintf("Timestamp: %s\n", alert.Timestamp.Format(time.RFC3339)))
	
	if alert.Tenant != "" {
		body.WriteString(fmt.Sprintf("Tenant: %s\n", alert.Tenant))
	}
	
	body.WriteString(fmt.Sprintf("\nMessage:\n%s\n", alert.Message))
	
	if len(alert.Details) > 0 {
		body.WriteString("\nDetails:\n")
		for key, value := range alert.Details {
			body.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
	}
	
	body.WriteString(fmt.Sprintf("\nAlert ID: %s\n", alert.ID))
	body.WriteString("Generated by: Mimir Limit Optimizer\n")
	body.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	
	return body.String()
}

func (e *EmailChannel) buildEmailMessage(subject, body string, alert *Alert) []byte {
	var msg strings.Builder
	
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.config.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString(fmt.Sprintf("X-Priority: %s\r\n", e.getEmailPriority(alert.Priority)))
	msg.WriteString("\r\n")
	msg.WriteString(body)
	
	return []byte(msg.String())
}

func (e *EmailChannel) getEmailPriority(priority Priority) string {
	switch priority {
	case PriorityP0:
		return "1 (Highest)"
	case PriorityP1:
		return "2 (High)"
	case PriorityP2:
		return "3 (Normal)"
	case PriorityP3:
		return "4 (Low)"
	default:
		return "3 (Normal)"
	}
}

func (e *EmailChannel) sendEmail(msg []byte) error {
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)
	
	var auth smtp.Auth
	if e.config.Username != "" && e.config.Password != "" {
		auth = smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	}
	
	return smtp.SendMail(addr, auth, e.config.From, e.config.To, msg)
}

// WebhookChannel implements the Channel interface for generic webhooks
type WebhookChannel struct {
	name   string
	config *config.WebhookConfig
	logger logr.Logger
	client *http.Client
}

// NewWebhookChannel creates a new Webhook channel
func NewWebhookChannel(name string, config config.WebhookConfig, logger logr.Logger) *WebhookChannel {
	return &WebhookChannel{
		name:   name,
		config: &config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (w *WebhookChannel) Name() string {
	return fmt.Sprintf("webhook_%s", w.name)
}

func (w *WebhookChannel) ValidateConfiguration() error {
	if !w.config.Enabled {
		return fmt.Errorf("webhook %s is disabled", w.name)
	}
	if w.config.URL == "" {
		return fmt.Errorf("webhook %s URL is required", w.name)
	}
	return nil
}

func (w *WebhookChannel) GetConfiguration() interface{} {
	return map[string]interface{}{
		"name":    w.name,
		"enabled": w.config.Enabled,
		"method":  w.config.Method,
		"timeout": w.config.Timeout,
		"headers": len(w.config.Headers),
	}
}

func (w *WebhookChannel) IsHealthy() bool {
	// Perform a simple health check (HEAD request if supported, otherwise GET)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", w.config.URL, nil)
	if err != nil {
		w.logger.V(1).Info("Webhook health check failed - request creation", 
			"webhook", w.name, "error", err)
		return false
	}
	
	// Add custom headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := w.client.Do(req)
	if err != nil {
		w.logger.V(1).Info("Webhook health check failed - request execution", 
			"webhook", w.name, "error", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	
	// Consider 2xx and 404 (method not allowed) as healthy
	return resp.StatusCode < 300 || resp.StatusCode == http.StatusNotFound
}

func (w *WebhookChannel) Send(ctx context.Context, alert *Alert) error {
	if !w.config.Enabled {
		return fmt.Errorf("webhook %s is disabled", w.name)
	}
	
	payload := w.buildWebhookPayload(alert)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		w.logger.Error(err, "Failed to marshal webhook payload", 
			"webhook", w.name, "alert_id", alert.ID)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	method := w.config.Method
	if method == "" {
		method = "POST"
	}
	
	req, err := http.NewRequestWithContext(ctx, method, w.config.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		w.logger.Error(err, "Failed to create webhook request", 
			"webhook", w.name, "alert_id", alert.ID)
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mimir-limit-optimizer/1.0")
	
	// Add custom headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}
	
	startTime := time.Now()
	resp, err := w.client.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		w.logger.Error(err, "Failed to send webhook alert", 
			"webhook", w.name,
			"alert_id", alert.ID,
			"duration", duration)
				return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		w.logger.Error(fmt.Errorf("webhook returned error status"), 
			"Failed to send webhook alert",
			"webhook", w.name,
			"alert_id", alert.ID,
			"status_code", resp.StatusCode,
			"duration", duration)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	w.logger.Info("Webhook alert sent successfully", 
		"webhook", w.name,
		"alert_id", alert.ID,
		"status_code", resp.StatusCode,
		"duration", duration)
	
	return nil
}

func (w *WebhookChannel) buildWebhookPayload(alert *Alert) map[string]interface{} {
	payload := map[string]interface{}{
		"id":          alert.ID,
		"type":        alert.Type,
		"priority":    alert.Priority,
		"title":       alert.Title,
		"message":     alert.Message,
		"timestamp":   alert.Timestamp.Format(time.RFC3339),
		"created_at":  alert.CreatedAt.Format(time.RFC3339),
		"source":      "mimir-limit-optimizer",
		"version":     "1.0",
	}
	
	if alert.Tenant != "" {
		payload["tenant"] = alert.Tenant
	}
	
	if len(alert.Details) > 0 {
		payload["details"] = alert.Details
	}
	
	// Add webhook-specific metadata
	payload["webhook"] = map[string]interface{}{
		"name":      w.name,
		"sent_at":   time.Now().Format(time.RFC3339),
		"retry_count": alert.RetryCount,
	}
	
	return payload
} 