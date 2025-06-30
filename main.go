
// Mimir Limit Optimizer v1.0.0
// Enterprise-grade guard rail system for Grafana Mimir
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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
	cfg, err := config.LoadConfig(configFile)
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
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return nil
} 