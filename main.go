
// Mimir Limit Optimizer v1.0.0
// Enterprise-grade guard rail system for Grafana Mimir
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/controller"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/metrics"
)

var (
	scheme    = runtime.NewScheme()
	setupLog  = ctrl.Log.WithName("setup")
	
	// Build-time variables (injected by ldflags during build)
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var configFile string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var logLevel string
	var showVersion bool
	var healthCheck bool

	flag.StringVar(&configFile, "config", "", "Path to the configuration file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit.")
	flag.BoolVar(&healthCheck, "health-check", false, "Perform health check and exit.")

	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Println(getBuildInfo())
		os.Exit(0)
	}

	// Handle health check flag
	if healthCheck {
		if err := performHealthCheck(probeAddr); err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Health check passed")
		os.Exit(0)
	}

	// Set log level
	switch logLevel {
	case "debug":
		opts.Development = true
	case "info":
		// Default level, no changes needed
	case "warn":
		// In controller-runtime zap, use default with lower verbosity
	case "error":
		// In controller-runtime zap, use default with lower verbosity
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Load configuration
	cfg, err := config.LoadConfigFromFile(configFile)
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		setupLog.Error(err, "invalid configuration")
		os.Exit(1)
	}

	setupLog.Info("Starting mimir-limit-optimizer", 
		"version", getBuildInfo(), 
		"mode", cfg.Mode,
		"updateInterval", cfg.UpdateInterval)

	// Check if we can run in standalone mode (without Kubernetes)
	standaloneMode := canRunStandalone(cfg)
	if standaloneMode {
		setupLog.Info("Running in standalone mode - no Kubernetes connectivity required")
		if err := runStandalone(cfg); err != nil {
			setupLog.Error(err, "failed to run standalone")
			os.Exit(1)
		}
		return
	}

	// Normal Kubernetes mode
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "mimir-limit-optimizer.akshaydubey29.github.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize metrics
	if err := metrics.RegisterMetrics(); err != nil {
		setupLog.Error(err, "unable to register metrics")
		os.Exit(1)
	}

	// Setup the controller
	if err = (&controller.MimirLimitController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
		Log:    ctrl.Log.WithName("controllers").WithName("MimirLimit"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MimirLimit")
		os.Exit(1)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getBuildInfo() string {
	return fmt.Sprintf("mimir-limit-optimizer version %s (commit: %s, built: %s)", Version, Commit, BuildDate)
}

// performHealthCheck performs a health check against the health probe endpoint
func performHealthCheck(probeAddr string) error {
	// If probeAddr starts with ":", prepend localhost
	if probeAddr[0] == ':' {
		probeAddr = "localhost" + probeAddr
	}
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(fmt.Sprintf("http://%s/healthz", probeAddr))
	if err != nil {
		return fmt.Errorf("failed to connect to health endpoint: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return nil
}

// canRunStandalone determines if the system can run without Kubernetes connectivity
func canRunStandalone(cfg *config.Config) bool {
	// Can run standalone if:
	// 1. We have fallback tenants configured OR synthetic mode is enabled
	// 2. AND we're in dry-run mode (not applying changes to Kubernetes)
	// 3. AND we're not using Kubernetes service discovery
	
	hasFallbackTenants := len(cfg.MetricsDiscovery.TenantDiscovery.FallbackTenants) > 0
	hasSyntheticMode := cfg.MetricsDiscovery.TenantDiscovery.EnableSynthetic || cfg.Synthetic.Enabled
	isDryRun := cfg.Mode == "dry-run"
	noKubernetesDiscovery := !cfg.MetricsDiscovery.Enabled
	
	// We can run standalone if we have tenant sources that don't require Kubernetes
	// and we're not applying changes to Kubernetes
	// Note: Having a metrics endpoint is fine - we can still make HTTP calls without Kubernetes
	canRunWithoutK8s := (hasFallbackTenants || hasSyntheticMode) && isDryRun && noKubernetesDiscovery
	
	return canRunWithoutK8s
}

// runStandalone runs the optimizer in standalone mode without Kubernetes
func runStandalone(cfg *config.Config) error {
	setupLog.Info("Initializing standalone mode",
		"hasFallbackTenants", len(cfg.MetricsDiscovery.TenantDiscovery.FallbackTenants),
		"syntheticEnabled", cfg.MetricsDiscovery.TenantDiscovery.EnableSynthetic,
		"metricsTenantID", cfg.MetricsDiscovery.TenantDiscovery.MetricsTenantID,
		"metricsEndpoint", cfg.MetricsEndpoint,
		"mode", cfg.Mode)
	
	// Initialize metrics
	if err := metrics.RegisterMetrics(); err != nil {
		return fmt.Errorf("unable to register metrics: %w", err)
	}
	
	// Test metrics connectivity if endpoint is configured
	if cfg.MetricsEndpoint != "" {
		setupLog.Info("Testing metrics connectivity with tenant headers")
		if err := testMetricsConnectivity(cfg); err != nil {
			setupLog.Info("Metrics connectivity test failed, falling back to tenant discovery", "error", err)
		} else {
			setupLog.Info("Metrics connectivity test successful")
		}
	}
	
	// Create a synthetic collector for tenant discovery
	collector := createStandaloneCollector(cfg)
	
	// Discover tenants
	tenants, err := collector.GetTenantList(nil)
	if err != nil {
		return fmt.Errorf("failed to discover tenants: %w", err)
	}
	
	setupLog.Info("Discovered tenants in standalone mode", 
		"count", len(tenants), 
		"tenants", tenants)
	
	// In standalone mode, we just demonstrate tenant discovery and exit
	// This shows that the fallback system works
	setupLog.Info("Standalone mode demonstration complete - tenant discovery successful")
	
	// In a real scenario, you might want to:
	// 1. Collect synthetic metrics
	// 2. Calculate recommended limits
	// 3. Output recommendations to a file
	// 4. Set up a web server to serve the recommendations
	
	return nil
}

// createStandaloneCollector creates a collector that can work without Kubernetes
func createStandaloneCollector(cfg *config.Config) standaloneCollector {
	return standaloneCollector{
		cfg: cfg,
		log: setupLog.WithName("standalone-collector"),
	}
}

// standaloneCollector implements basic tenant discovery without Kubernetes
type standaloneCollector struct {
	cfg *config.Config
	log logr.Logger
}

// GetTenantList implements tenant discovery for standalone mode
func (s *standaloneCollector) GetTenantList(ctx context.Context) ([]string, error) {
	// Try fallback tenants first
	if len(s.cfg.MetricsDiscovery.TenantDiscovery.FallbackTenants) > 0 {
		s.log.Info("Using configured fallback tenants", 
			"count", len(s.cfg.MetricsDiscovery.TenantDiscovery.FallbackTenants))
		return s.cfg.MetricsDiscovery.TenantDiscovery.FallbackTenants, nil
	}
	
	// Try synthetic tenants
	if s.cfg.MetricsDiscovery.TenantDiscovery.EnableSynthetic || s.cfg.Synthetic.Enabled {
		count := s.cfg.MetricsDiscovery.TenantDiscovery.SyntheticCount
		if count <= 0 {
			if s.cfg.Synthetic.Enabled {
				count = s.cfg.Synthetic.TenantCount
			} else {
				count = 3
			}
		}
		
		tenants := make([]string, count)
		for i := 0; i < count; i++ {
			tenants[i] = fmt.Sprintf("synthetic-tenant-%d", i+1)
		}
		
		s.log.Info("Generated synthetic tenants", "count", len(tenants))
		return tenants, nil
	}
	
	return nil, fmt.Errorf("no tenant discovery method available in standalone mode")
}

// testMetricsConnectivity tests HTTP connectivity to the metrics endpoint with tenant headers
func testMetricsConnectivity(cfg *config.Config) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Test with a simple query
	testURL := fmt.Sprintf("%s?query=up", cfg.MetricsEndpoint)
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}
	
	// Add tenant headers
	if cfg.MetricsDiscovery.TenantDiscovery.MetricsTenantID != "" {
		req.Header.Set("X-Scope-OrgID", cfg.MetricsDiscovery.TenantDiscovery.MetricsTenantID)
		setupLog.Info("Added tenant header for test", "tenant", cfg.MetricsDiscovery.TenantDiscovery.MetricsTenantID)
	}
	
	// Add any additional custom headers
	for key, value := range cfg.MetricsDiscovery.TenantDiscovery.TenantHeaders {
		req.Header.Set(key, value)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	setupLog.Info("Metrics endpoint test response", 
		"status", resp.Status, 
		"statusCode", resp.StatusCode)
	
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	
	return fmt.Errorf("metrics endpoint returned status: %d %s", resp.StatusCode, resp.Status)
} 