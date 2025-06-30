package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the mimir-limit-optimizer
type Config struct {
	// Operating mode: "dry-run" or "prod"
	Mode string `yaml:"mode" json:"mode"`
	
	// Buffer percentage to add to calculated limits (e.g., 20 for 20%)
	BufferPercentage float64 `yaml:"bufferPercentage" json:"bufferPercentage"`
	
	// How often to update limits
	UpdateInterval time.Duration `yaml:"updateInterval" json:"updateInterval"`
	
	// Mimir configuration
	Mimir MimirConfig `yaml:"mimir" json:"mimir"`
	
	// Tenant scoping configuration
	TenantScoping TenantScopingConfig `yaml:"tenantScoping" json:"tenantScoping"`
	
	// Metrics discovery configuration
	MetricsDiscovery MetricsDiscoveryConfig `yaml:"metricsDiscovery" json:"metricsDiscovery"`
	
	// Fallback metrics endpoint (e.g., Prometheus URL)
	MetricsEndpoint string `yaml:"metricsEndpoint" json:"metricsEndpoint"`
	
	// Event-driven spike detection
	EventSpike EventSpikeConfig `yaml:"eventSpike" json:"eventSpike"`
	
	// Trend analysis configuration
	TrendAnalysis TrendAnalysisConfig `yaml:"trendAnalysis" json:"trendAnalysis"`
	
	// Limits configuration
	Limits LimitsConfig `yaml:"limits" json:"limits"`
	
	// Audit logging configuration
	AuditLog AuditLogConfig `yaml:"auditLog" json:"auditLog"`
	
	// Synthetic mode for testing
	Synthetic SyntheticConfig `yaml:"synthetic" json:"synthetic"`
	
	// Cost control and budget management
	CostControl CostControlConfig `yaml:"costControl" json:"costControl"`
	
	// Circuit breaker for blast protection
	CircuitBreaker CircuitBreakerConfig `yaml:"circuitBreaker" json:"circuitBreaker"`
	
	// Emergency controls
	Emergency EmergencyConfig `yaml:"emergency" json:"emergency"`
	
	// Advanced alerting
	Alerting AlertingConfig `yaml:"alerting" json:"alerting"`
	
	// Performance optimization
	Performance PerformanceConfig `yaml:"performance" json:"performance"`
}

type MimirConfig struct {
	// Namespace where Mimir is deployed
	Namespace string `yaml:"namespace" json:"namespace"`
	
	// Name of the runtime overrides ConfigMap
	ConfigMapName string `yaml:"configMapName" json:"configMapName"`
	
	// Whether to trigger rollout after ConfigMap changes
	// Default: false - Mimir runtime overrides are applied automatically without restarts
	// Set to true only if you need forced component restarts for other configuration changes
	TriggerRollout bool `yaml:"triggerRollout" json:"triggerRollout"`
	
	// Components to rollout (if TriggerRollout is true)
	RolloutComponents []string `yaml:"rolloutComponents" json:"rolloutComponents"`
}

type TenantScopingConfig struct {
	// List of tenant patterns to skip (glob or regex)
	SkipList []string `yaml:"skipList" json:"skipList"`
	
	// List of tenant patterns to include (empty means all, glob or regex)
	IncludeList []string `yaml:"includeList" json:"includeList"`
	
	// Whether to use regex instead of glob patterns
	UseRegex bool `yaml:"useRegex" json:"useRegex"`
}

type MetricsDiscoveryConfig struct {
	// Enable auto-discovery of metrics endpoints
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Namespace to discover services in
	Namespace string `yaml:"namespace" json:"namespace"`
	
	// Label selector for discovering services
	ServiceLabelSelector string `yaml:"serviceLabelSelector" json:"serviceLabelSelector"`
	
	// List of known service names to discover
	ServiceNames []string `yaml:"serviceNames" json:"serviceNames"`
	
	// Metrics path on services
	MetricsPath string `yaml:"metricsPath" json:"metricsPath"`
	
	// Port name for metrics
	PortName string `yaml:"portName" json:"portName"`
	
	// Port number (if PortName is not used)
	Port int `yaml:"port" json:"port"`
}

type EventSpikeConfig struct {
	// Enable event-driven spike detection
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Threshold multiplier for spike detection (e.g., 2.0 for 2x)
	Threshold float64 `yaml:"threshold" json:"threshold"`
	
	// Time window for spike detection
	DetectionWindow time.Duration `yaml:"detectionWindow" json:"detectionWindow"`
	
	// How long to keep increased limits after spike
	CooldownPeriod time.Duration `yaml:"cooldownPeriod" json:"cooldownPeriod"`
	
	// Maximum spike multiplier to apply
	MaxSpikeMultiplier float64 `yaml:"maxSpikeMultiplier" json:"maxSpikeMultiplier"`
}

type TrendAnalysisConfig struct {
	// Time window for trend analysis
	AnalysisWindow time.Duration `yaml:"analysisWindow" json:"analysisWindow"`
	
	// Percentile to use for trend analysis (e.g., 95 for 95th percentile)
	Percentile float64 `yaml:"percentile" json:"percentile"`
	
	// Use moving average in trend calculation
	UseMovingAverage bool `yaml:"useMovingAverage" json:"useMovingAverage"`
	
	// Include peak usage in calculations
	IncludePeaks bool `yaml:"includePeaks" json:"includePeaks"`
	
	// Time-of-day specific buffers
	TimeOfDayBuffers map[string]float64 `yaml:"timeOfDayBuffers" json:"timeOfDayBuffers"`
}

type LimitsConfig struct {
	// Minimum limits per tenant
	MinLimits map[string]interface{} `yaml:"minLimits" json:"minLimits"`
	
	// Maximum limits per tenant
	MaxLimits map[string]interface{} `yaml:"maxLimits" json:"maxLimits"`
	
	// Default limits for new tenants
	DefaultLimits map[string]interface{} `yaml:"defaultLimits" json:"defaultLimits"`
	
	// TTL for removing limits of inactive tenants
	InactiveTenantTTL time.Duration `yaml:"inactiveTenantTTL" json:"inactiveTenantTTL"`
	
	// Tenant tiers configuration
	TenantTiers map[string]TenantTierConfig `yaml:"tenantTiers" json:"tenantTiers"`
}

type TenantTierConfig struct {
	// Buffer percentage for this tier
	BufferPercentage float64 `yaml:"bufferPercentage" json:"bufferPercentage"`
	
	// Specific limits for this tier
	Limits map[string]interface{} `yaml:"limits" json:"limits"`
}

type AuditLogConfig struct {
	// Enable audit logging
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Storage type: "memory", "configmap", or "external"
	StorageType string `yaml:"storageType" json:"storageType"`
	
	// Maximum entries to keep in memory
	MaxEntries int `yaml:"maxEntries" json:"maxEntries"`
	
	// ConfigMap name for audit storage
	ConfigMapName string `yaml:"configMapName" json:"configMapName"`
	
	// External storage configuration
	ExternalStorage map[string]interface{} `yaml:"externalStorage" json:"externalStorage"`
}

type SyntheticConfig struct {
	// Enable synthetic mode for testing
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Number of synthetic tenants to simulate
	TenantCount int `yaml:"tenantCount" json:"tenantCount"`
	
	// Synthetic metrics configuration
	MetricsConfig map[string]interface{} `yaml:"metricsConfig" json:"metricsConfig"`
}

// CostControlConfig defines cost management and budget controls
type CostControlConfig struct {
	// Enable cost control features
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Cost calculation method: "samples", "series", "queries", "composite"
	CostMethod string `yaml:"costMethod" json:"costMethod"`
	
	// Cost per unit (e.g., per million samples)
	CostPerUnit float64 `yaml:"costPerUnit" json:"costPerUnit"`
	
	// Budget limits per tenant
	TenantBudgets map[string]BudgetConfig `yaml:"tenantBudgets" json:"tenantBudgets"`
	
	// Global budget limit
	GlobalBudget BudgetConfig `yaml:"globalBudget" json:"globalBudget"`
	
	// Cost alerting thresholds (percentage of budget)
	AlertThresholds []float64 `yaml:"alertThresholds" json:"alertThresholds"`
	
	// Automatic limit reduction when over budget
	AutoLimitReduction bool `yaml:"autoLimitReduction" json:"autoLimitReduction"`
	
	// Cost estimation window
	EstimationWindow time.Duration `yaml:"estimationWindow" json:"estimationWindow"`
}

type BudgetConfig struct {
	// Daily budget limit
	Daily float64 `yaml:"daily" json:"daily"`
	
	// Monthly budget limit
	Monthly float64 `yaml:"monthly" json:"monthly"`
	
	// Annual budget limit
	Annual float64 `yaml:"annual" json:"annual"`
	
	// Currency code
	Currency string `yaml:"currency" json:"currency"`
	
	// Enforce budget (block operations when exceeded)
	EnforceBudget bool `yaml:"enforceBudget" json:"enforceBudget"`
}

// CircuitBreakerConfig defines blast protection mechanisms
type CircuitBreakerConfig struct {
	// Enable circuit breaker
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Runtime enable/disable control (can be changed without restart)
	RuntimeEnabled bool `yaml:"runtimeEnabled" json:"runtimeEnabled"`
	
	// Auto-configuration mode: "manual", "auto", "hybrid"
	Mode string `yaml:"mode" json:"mode"`
	
	// Failure threshold (percentage) - used in manual mode
	FailureThreshold float64 `yaml:"failureThreshold" json:"failureThreshold"`
	
	// Request volume threshold - used in manual mode
	RequestVolumeThreshold int `yaml:"requestVolumeThreshold" json:"requestVolumeThreshold"`
	
	// Sleep window for half-open state
	SleepWindow time.Duration `yaml:"sleepWindow" json:"sleepWindow"`
	
	// Maximum requests in half-open state
	MaxRequestsInHalfOpen int `yaml:"maxRequestsInHalfOpen" json:"maxRequestsInHalfOpen"`
	
	// Automatic configuration based on limits and metrics
	AutoConfig AutoCircuitBreakerConfig `yaml:"autoConfig" json:"autoConfig"`
	
	// Rate limiting configuration
	RateLimit RateLimitConfig `yaml:"rateLimit" json:"rateLimit"`
	
	// Blast protection thresholds
	BlastProtection BlastProtectionConfig `yaml:"blastProtection" json:"blastProtection"`
}

// AutoCircuitBreakerConfig defines automatic threshold calculation
type AutoCircuitBreakerConfig struct {
	// Enable automatic threshold calculation
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Threshold multipliers based on current limits
	LimitMultipliers LimitMultiplierConfig `yaml:"limitMultipliers" json:"limitMultipliers"`
	
	// Real-time adaptation settings
	RealtimeAdaptation RealtimeAdaptationConfig `yaml:"realtimeAdaptation" json:"realtimeAdaptation"`
	
	// Baseline calculation window
	BaselineWindow time.Duration `yaml:"baselineWindow" json:"baselineWindow"`
	
	// Minimum observation period before auto-tuning
	MinObservationPeriod time.Duration `yaml:"minObservationPeriod" json:"minObservationPeriod"`
	
	// Safety margins for threshold calculation
	SafetyMargins SafetyMarginConfig `yaml:"safetyMargins" json:"safetyMargins"`
}

// LimitMultiplierConfig defines how to calculate thresholds from limits
type LimitMultiplierConfig struct {
	// Ingestion rate threshold as multiplier of current limit
	IngestionRateMultiplier float64 `yaml:"ingestionRateMultiplier" json:"ingestionRateMultiplier"`
	
	// Query rate threshold as multiplier of current limit  
	QueryRateMultiplier float64 `yaml:"queryRateMultiplier" json:"queryRateMultiplier"`
	
	// Series threshold as multiplier of current limit
	SeriesMultiplier float64 `yaml:"seriesMultiplier" json:"seriesMultiplier"`
	
	// Burst threshold as multiplier of current burst limit
	BurstMultiplier float64 `yaml:"burstMultiplier" json:"burstMultiplier"`
}

// RealtimeAdaptationConfig defines real-time threshold adaptation
type RealtimeAdaptationConfig struct {
	// Enable real-time adaptation
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Adaptation interval
	Interval time.Duration `yaml:"interval" json:"interval"`
	
	// Learning rate for threshold adjustment (0.0-1.0)
	LearningRate float64 `yaml:"learningRate" json:"learningRate"`
	
	// Maximum threshold change per adaptation cycle (percentage)
	MaxChangePercent float64 `yaml:"maxChangePercent" json:"maxChangePercent"`
	
	// Percentile to use for threshold calculation
	Percentile float64 `yaml:"percentile" json:"percentile"`
	
	// Enable seasonal pattern detection
	SeasonalPatterns bool `yaml:"seasonalPatterns" json:"seasonalPatterns"`
}

// SafetyMarginConfig defines safety margins for auto-calculated thresholds
type SafetyMarginConfig struct {
	// Minimum safety margin (percentage above calculated threshold)
	MinMargin float64 `yaml:"minMargin" json:"minMargin"`
	
	// Maximum safety margin
	MaxMargin float64 `yaml:"maxMargin" json:"maxMargin"`
	
	// Default safety margin
	DefaultMargin float64 `yaml:"defaultMargin" json:"defaultMargin"`
	
	// Tenant-specific margins
	TenantMargins map[string]float64 `yaml:"tenantMargins" json:"tenantMargins"`
}

type RateLimitConfig struct {
	// Enable rate limiting
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Requests per second per tenant
	RequestsPerSecond float64 `yaml:"requestsPerSecond" json:"requestsPerSecond"`
	
	// Burst capacity
	BurstCapacity int `yaml:"burstCapacity" json:"burstCapacity"`
	
	// Rate limit window
	Window time.Duration `yaml:"window" json:"window"`
}

type BlastProtectionConfig struct {
	// Manual thresholds (used when auto-config is disabled)
	ManualThresholds ManualThresholdConfig `yaml:"manualThresholds" json:"manualThresholds"`
	
	// Use automatic threshold calculation based on current limits
	UseAutoThresholds bool `yaml:"useAutoThresholds" json:"useAutoThresholds"`
	
	// Automatic emergency shutdown
	AutoEmergencyShutdown bool `yaml:"autoEmergencyShutdown" json:"autoEmergencyShutdown"`
	
	// Recovery time after blast
	RecoveryTime time.Duration `yaml:"recoveryTime" json:"recoveryTime"`
	
	// Per-tenant threshold overrides
	TenantOverrides map[string]ManualThresholdConfig `yaml:"tenantOverrides" json:"tenantOverrides"`
}

// ManualThresholdConfig defines manual threshold values
type ManualThresholdConfig struct {
	// Ingestion rate spike threshold (samples/sec)
	IngestionSpikeThreshold float64 `yaml:"ingestionSpikeThreshold" json:"ingestionSpikeThreshold"`
	
	// Query rate spike threshold (queries/sec)
	QuerySpikeThreshold float64 `yaml:"querySpikeThreshold" json:"querySpikeThreshold"`
	
	// Series creation spike threshold (series/sec)
	SeriesSpikeThreshold float64 `yaml:"seriesSpikeThreshold" json:"seriesSpikeThreshold"`
}

// EmergencyConfig defines emergency controls
type EmergencyConfig struct {
	// Enable emergency controls
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Emergency contact webhook URL
	WebhookURL string `yaml:"webhookURL" json:"webhookURL"`
	
	// Emergency contact emails
	Contacts []string `yaml:"contacts" json:"contacts"`
	
	// Panic mode configuration
	PanicMode PanicModeConfig `yaml:"panicMode" json:"panicMode"`
	
	// Emergency shutdown triggers
	ShutdownTriggers []EmergencyTrigger `yaml:"shutdownTriggers" json:"shutdownTriggers"`
	
	// Recovery procedures
	RecoveryProcedures RecoveryConfig `yaml:"recoveryProcedures" json:"recoveryProcedures"`
}

type PanicModeConfig struct {
	// Enable panic mode
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// CPU threshold for panic mode (percentage)
	CPUThreshold float64 `yaml:"cpuThreshold" json:"cpuThreshold"`
	
	// Memory threshold for panic mode (percentage)
	MemoryThreshold float64 `yaml:"memoryThreshold" json:"memoryThreshold"`
	
	// Error rate threshold for panic mode (errors/sec)
	ErrorRateThreshold float64 `yaml:"errorRateThreshold" json:"errorRateThreshold"`
	
	// Actions to take in panic mode
	Actions []string `yaml:"actions" json:"actions"`
}

type EmergencyTrigger struct {
	// Trigger name
	Name string `yaml:"name" json:"name"`
	
	// Metric to monitor
	Metric string `yaml:"metric" json:"metric"`
	
	// Threshold value
	Threshold float64 `yaml:"threshold" json:"threshold"`
	
	// Duration threshold must be exceeded
	Duration time.Duration `yaml:"duration" json:"duration"`
	
	// Action to take
	Action string `yaml:"action" json:"action"`
}

type RecoveryConfig struct {
	// Automatic recovery enabled
	AutoRecovery bool `yaml:"autoRecovery" json:"autoRecovery"`
	
	// Recovery check interval
	CheckInterval time.Duration `yaml:"checkInterval" json:"checkInterval"`
	
	// Health check timeout
	HealthCheckTimeout time.Duration `yaml:"healthCheckTimeout" json:"healthCheckTimeout"`
	
	// Maximum recovery attempts
	MaxAttempts int `yaml:"maxAttempts" json:"maxAttempts"`
}

// AlertingConfig defines advanced alerting configuration
type AlertingConfig struct {
	// Enable alerting
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Slack integration
	Slack SlackConfig `yaml:"slack" json:"slack"`
	
	// PagerDuty integration
	PagerDuty PagerDutyConfig `yaml:"pagerDuty" json:"pagerDuty"`
	
	// Email configuration
	Email EmailConfig `yaml:"email" json:"email"`
	
	// Webhook endpoints
	Webhooks []WebhookConfig `yaml:"webhooks" json:"webhooks"`
	
	// Alert routing rules
	RoutingRules []AlertRoutingRule `yaml:"routingRules" json:"routingRules"`
	
	// Escalation policies
	EscalationPolicies []EscalationPolicy `yaml:"escalationPolicies" json:"escalationPolicies"`
}

type SlackConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	WebhookURL string        `yaml:"webhookURL" json:"webhookURL"`
	Channel    string        `yaml:"channel" json:"channel"`
	Username   string        `yaml:"username" json:"username"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
}

type PagerDutyConfig struct {
	Enabled        bool          `yaml:"enabled" json:"enabled"`
	IntegrationKey string        `yaml:"integrationKey" json:"integrationKey"`
	Severity       string        `yaml:"severity" json:"severity"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
}

type EmailConfig struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	SMTPHost string   `yaml:"smtpHost" json:"smtpHost"`
	SMTPPort int      `yaml:"smtpPort" json:"smtpPort"`
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	From     string   `yaml:"from" json:"from"`
	To       []string `yaml:"to" json:"to"`
	UseTLS   bool     `yaml:"useTLS" json:"useTLS"`
}

type WebhookConfig struct {
	Name    string            `yaml:"name" json:"name"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Timeout time.Duration     `yaml:"timeout" json:"timeout"`
	Enabled bool              `yaml:"enabled" json:"enabled"`
	Method  string            `yaml:"method" json:"method"`
}

type AlertRoutingRule struct {
	Name      string            `yaml:"name" json:"name"`
	Condition string            `yaml:"condition" json:"condition"`
	Channels  []string          `yaml:"channels" json:"channels"`
	Metadata  map[string]string `yaml:"metadata" json:"metadata"`
}

type EscalationPolicy struct {
	Name     string             `yaml:"name" json:"name"`
	Levels   []EscalationLevel  `yaml:"levels" json:"levels"`
	Timeout  time.Duration      `yaml:"timeout" json:"timeout"`
}

type EscalationLevel struct {
	Level    int      `yaml:"level" json:"level"`
	Channels []string `yaml:"channels" json:"channels"`
	Delay    time.Duration `yaml:"delay" json:"delay"`
}

// PerformanceConfig defines performance optimization settings
type PerformanceConfig struct {
	// Enable performance optimizations
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Caching configuration
	Cache CacheConfig `yaml:"cache" json:"cache"`
	
	// Batch processing settings
	BatchProcessing BatchConfig `yaml:"batchProcessing" json:"batchProcessing"`
	
	// Resource optimization
	ResourceOptimization ResourceOptimizationConfig `yaml:"resourceOptimization" json:"resourceOptimization"`
	
	// Compression settings
	Compression CompressionConfig `yaml:"compression" json:"compression"`
}

type CacheConfig struct {
	// Enable caching
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Cache TTL
	TTL time.Duration `yaml:"ttl" json:"ttl"`
	
	// Cache size (MB)
	SizeMB int `yaml:"sizeMB" json:"sizeMB"`
	
	// Cache type: "memory", "redis", "memcached"
	Type string `yaml:"type" json:"type"`
	
	// Redis configuration (if type is redis)
	Redis RedisConfig `yaml:"redis" json:"redis"`
}

type RedisConfig struct {
	Address  string `yaml:"address" json:"address"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type BatchConfig struct {
	// Enable batch processing
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Batch size
	Size int `yaml:"size" json:"size"`
	
	// Batch timeout
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	
	// Max concurrent batches
	MaxConcurrent int `yaml:"maxConcurrent" json:"maxConcurrent"`
}

type ResourceOptimizationConfig struct {
	// CPU optimization
	CPUOptimization bool `yaml:"cpuOptimization" json:"cpuOptimization"`
	
	// Memory optimization
	MemoryOptimization bool `yaml:"memoryOptimization" json:"memoryOptimization"`
	
	// Garbage collection tuning
	GCTuning GCTuningConfig `yaml:"gcTuning" json:"gcTuning"`
}

type GCTuningConfig struct {
	// Target GC percentage
	TargetPercent int `yaml:"targetPercent" json:"targetPercent"`
	
	// Memory limit
	MemoryLimit string `yaml:"memoryLimit" json:"memoryLimit"`
}

type CompressionConfig struct {
	// Enable compression
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Compression algorithm: "gzip", "lz4", "snappy"
	Algorithm string `yaml:"algorithm" json:"algorithm"`
	
	// Compression level
	Level int `yaml:"level" json:"level"`
}

// LoadConfig loads configuration from file or environment
func LoadConfig(configFile string) (*Config, error) {
	mode := getEnvOrDefault("MODE", "dry-run")
	
	// Circuit breaker should be disabled by default in dry-run mode for observation purposes
	circuitBreakerEnabled := mode == "prod"
	
	config := &Config{
		Mode:             mode,
		BufferPercentage: 20.0,
		UpdateInterval:   5 * time.Minute,
		Mimir: MimirConfig{
			Namespace:         getEnvOrDefault("MIMIR_NAMESPACE", "mimir"),
			ConfigMapName:     getEnvOrDefault("MIMIR_CONFIGMAP_NAME", "mimir-runtime-overrides"),
			TriggerRollout:    false,
			RolloutComponents: []string{"ingester", "querier", "query-frontend"},
		},
		TenantScoping: TenantScopingConfig{
			SkipList:    []string{},
			IncludeList: []string{},
			UseRegex:    false,
		},
		MetricsDiscovery: MetricsDiscoveryConfig{
			Enabled:              false,
			Namespace:            getEnvOrDefault("MIMIR_NAMESPACE", "mimir"),
			ServiceLabelSelector: "mimir-metrics=true",
			MetricsPath:          "/metrics",
			PortName:             "http-metrics",
			Port:                 8080,
		},
		EventSpike: EventSpikeConfig{
			Enabled:            true,
			Threshold:          2.0,
			DetectionWindow:    5 * time.Minute,
			CooldownPeriod:     30 * time.Minute,
			MaxSpikeMultiplier: 5.0,
		},
		TrendAnalysis: TrendAnalysisConfig{
			AnalysisWindow:      48 * time.Hour,
			Percentile:          95.0,
			UseMovingAverage:    true,
			IncludePeaks:        true,
			TimeOfDayBuffers:    make(map[string]float64),
		},
		Limits: LimitsConfig{
			MinLimits:         make(map[string]interface{}),
			MaxLimits:         make(map[string]interface{}),
			DefaultLimits:     make(map[string]interface{}),
			InactiveTenantTTL: 7 * 24 * time.Hour,
			TenantTiers:       make(map[string]TenantTierConfig),
		},
		AuditLog: AuditLogConfig{
			Enabled:     true,
			StorageType: "memory",
			MaxEntries:  1000,
		},
		Synthetic: SyntheticConfig{
			Enabled:     false,
			TenantCount: 10,
		},
		CostControl: CostControlConfig{
			Enabled:            true,
			CostMethod:         "composite",
			CostPerUnit:        0.001, // $0.001 per million samples
			TenantBudgets:      make(map[string]BudgetConfig),
			GlobalBudget:       BudgetConfig{Daily: 1000, Monthly: 30000, Annual: 365000, Currency: "USD", EnforceBudget: false},
			AlertThresholds:    []float64{50, 75, 90, 95},
			AutoLimitReduction: false, // Optional: Set to true to enable automatic limit reduction on budget violations
			EstimationWindow:   24 * time.Hour,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:                circuitBreakerEnabled, // Disabled in dry-run, enabled in prod
			RuntimeEnabled:         circuitBreakerEnabled, // Disabled in dry-run, enabled in prod
			Mode:                  "auto", // "manual", "auto", "hybrid"
			FailureThreshold:       50.0,
			RequestVolumeThreshold: 20,
			SleepWindow:           30 * time.Second,
			MaxRequestsInHalfOpen:  5,
			AutoConfig: AutoCircuitBreakerConfig{
				Enabled:              true,
				BaselineWindow:       24 * time.Hour,
				MinObservationPeriod: 1 * time.Hour,
				LimitMultipliers: LimitMultiplierConfig{
					IngestionRateMultiplier: 1.5, // Trip at 150% of current limit
					QueryRateMultiplier:     2.0, // Trip at 200% of current limit
					SeriesMultiplier:        1.8, // Trip at 180% of current limit
					BurstMultiplier:         1.2, // Trip at 120% of burst limit
				},
				RealtimeAdaptation: RealtimeAdaptationConfig{
					Enabled:          true,
					Interval:         5 * time.Minute,
					LearningRate:     0.1,
					MaxChangePercent: 20.0,
					Percentile:       95.0,
					SeasonalPatterns: false,
				},
				SafetyMargins: SafetyMarginConfig{
					MinMargin:       10.0,
					MaxMargin:       50.0,
					DefaultMargin:   25.0,
					TenantMargins:   make(map[string]float64),
				},
			},
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 100,
				BurstCapacity:     200,
				Window:           time.Minute,
			},
			BlastProtection: BlastProtectionConfig{
				UseAutoThresholds: true,
				ManualThresholds: ManualThresholdConfig{
					IngestionSpikeThreshold: 1000000, // Fallback values
					QuerySpikeThreshold:     10000,
					SeriesSpikeThreshold:    100000,
				},
				AutoEmergencyShutdown: true,
				RecoveryTime:          5 * time.Minute,
				TenantOverrides:       make(map[string]ManualThresholdConfig),
			},
		},
		Emergency: EmergencyConfig{
			Enabled: true,
			PanicMode: PanicModeConfig{
				Enabled:            true,
				CPUThreshold:       90.0,
				MemoryThreshold:    90.0,
				ErrorRateThreshold: 100,
				Actions:           []string{"reduce_limits", "throttle_ingestion", "alert"},
			},
			RecoveryProcedures: RecoveryConfig{
				AutoRecovery:       true,
				CheckInterval:      30 * time.Second,
				HealthCheckTimeout: 10 * time.Second,
				MaxAttempts:        3,
			},
		},
		Alerting: AlertingConfig{
			Enabled:           true,
			RoutingRules:      []AlertRoutingRule{},
			EscalationPolicies: []EscalationPolicy{},
		},
		Performance: PerformanceConfig{
			Enabled: true,
			Cache: CacheConfig{
				Enabled: true,
				TTL:     5 * time.Minute,
				SizeMB:  256,
				Type:    "memory",
			},
			BatchProcessing: BatchConfig{
				Enabled:       true,
				Size:          100,
				Timeout:       30 * time.Second,
				MaxConcurrent: 10,
			},
			ResourceOptimization: ResourceOptimizationConfig{
				CPUOptimization:    true,
				MemoryOptimization: true,
				GCTuning: GCTuningConfig{
					TargetPercent: 100,
					MemoryLimit:   "512Mi",
				},
			},
			Compression: CompressionConfig{
				Enabled:   true,
				Algorithm: "gzip",
				Level:     6,
			},
		},
	}

	if configFile != "" {
		return loadConfigFromFile(configFile, config)
	}

	return config, nil
}

func loadConfigFromFile(configFile string, config *Config) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Mode != "dry-run" && c.Mode != "prod" {
		return fmt.Errorf("mode must be 'dry-run' or 'prod', got '%s'", c.Mode)
	}

	if c.BufferPercentage < 0 || c.BufferPercentage > 1000 {
		return fmt.Errorf("bufferPercentage must be between 0 and 1000, got %f", c.BufferPercentage)
	}

	if c.UpdateInterval <= 0 {
		return fmt.Errorf("updateInterval must be positive, got %v", c.UpdateInterval)
	}

	if c.Mimir.Namespace == "" {
		return fmt.Errorf("mimir.namespace cannot be empty")
	}

	if c.Mimir.ConfigMapName == "" {
		return fmt.Errorf("mimir.configMapName cannot be empty")
	}

	if c.EventSpike.Enabled {
		if c.EventSpike.Threshold <= 1.0 {
			return fmt.Errorf("eventSpike.threshold must be greater than 1.0, got %f", c.EventSpike.Threshold)
		}
		if c.EventSpike.DetectionWindow <= 0 {
			return fmt.Errorf("eventSpike.detectionWindow must be positive, got %v", c.EventSpike.DetectionWindow)
		}
		if c.EventSpike.CooldownPeriod <= 0 {
			return fmt.Errorf("eventSpike.cooldownPeriod must be positive, got %v", c.EventSpike.CooldownPeriod)
		}
	}

	if c.TrendAnalysis.AnalysisWindow <= 0 {
		return fmt.Errorf("trendAnalysis.analysisWindow must be positive, got %v", c.TrendAnalysis.AnalysisWindow)
	}

	if c.TrendAnalysis.Percentile < 0 || c.TrendAnalysis.Percentile > 100 {
		return fmt.Errorf("trendAnalysis.percentile must be between 0 and 100, got %f", c.TrendAnalysis.Percentile)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 