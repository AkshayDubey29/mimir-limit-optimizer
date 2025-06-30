package discovery

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/tapasyadubey/mimir-limit-optimizer/internal/config"
)

// ServiceDiscovery handles automatic discovery of Mimir services
type ServiceDiscovery struct {
	client kubernetes.Interface
	config *config.Config
	log    logr.Logger
}

// DiscoveredEndpoint represents a discovered metrics endpoint
type DiscoveredEndpoint struct {
	ServiceName string
	Namespace   string
	URL         string
	Port        int32
	Labels      map[string]string
}

// NewServiceDiscovery creates a new ServiceDiscovery instance
func NewServiceDiscovery(client kubernetes.Interface, cfg *config.Config, log logr.Logger) *ServiceDiscovery {
	return &ServiceDiscovery{
		client: client,
		config: cfg,
		log:    log,
	}
}

// DiscoverMetricsEndpoints discovers all available Mimir metrics endpoints
func (d *ServiceDiscovery) DiscoverMetricsEndpoints(ctx context.Context) ([]string, error) {
	if !d.config.MetricsDiscovery.Enabled {
		return nil, nil
	}

	var endpoints []DiscoveredEndpoint
	var err error

	// Try label selector discovery first
	if d.config.MetricsDiscovery.ServiceLabelSelector != "" {
		labelEndpoints, err := d.discoverByLabels(ctx)
		if err != nil {
			d.log.Error(err, "failed to discover services by labels")
		} else {
			endpoints = append(endpoints, labelEndpoints...)
		}
	}

	// Try known service names discovery
	if len(d.config.MetricsDiscovery.ServiceNames) > 0 {
		nameEndpoints, err := d.discoverByNames(ctx)
		if err != nil {
			d.log.Error(err, "failed to discover services by names")
		} else {
			endpoints = append(endpoints, nameEndpoints...)
		}
	}

	if err != nil && len(endpoints) == 0 {
		return nil, fmt.Errorf("failed to discover any endpoints: %w", err)
	}

	// Convert to URLs
	urls := make([]string, 0, len(endpoints))
	seen := make(map[string]bool)

	for _, endpoint := range endpoints {
		url := d.buildMetricsURL(endpoint)
		if !seen[url] {
			urls = append(urls, url)
			seen[url] = true
		}
	}

	d.log.Info("discovered metrics endpoints", "count", len(urls), "endpoints", urls)

	return urls, nil
}

// discoverByLabels discovers services using label selectors
func (d *ServiceDiscovery) discoverByLabels(ctx context.Context) ([]DiscoveredEndpoint, error) {
	selector, err := labels.Parse(d.config.MetricsDiscovery.ServiceLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	services, err := d.client.CoreV1().Services(d.config.MetricsDiscovery.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var endpoints []DiscoveredEndpoint

	for _, service := range services.Items {
		endpoint := d.serviceToEndpoint(&service)
		if endpoint != nil {
			endpoints = append(endpoints, *endpoint)
		}
	}

	return endpoints, nil
}

// discoverByNames discovers services using known service names
func (d *ServiceDiscovery) discoverByNames(ctx context.Context) ([]DiscoveredEndpoint, error) {
	var endpoints []DiscoveredEndpoint

	for _, serviceName := range d.config.MetricsDiscovery.ServiceNames {
		service, err := d.client.CoreV1().Services(d.config.MetricsDiscovery.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			d.log.Error(err, "failed to get service", "service", serviceName)
			continue
		}

		endpoint := d.serviceToEndpoint(service)
		if endpoint != nil {
			endpoints = append(endpoints, *endpoint)
		}
	}

	return endpoints, nil
}

// serviceToEndpoint converts a Kubernetes service to a DiscoveredEndpoint
func (d *ServiceDiscovery) serviceToEndpoint(service *corev1.Service) *DiscoveredEndpoint {
	var port int32

	// Find the metrics port
	if d.config.MetricsDiscovery.PortName != "" {
		for _, p := range service.Spec.Ports {
			if p.Name == d.config.MetricsDiscovery.PortName {
				port = p.Port
				break
			}
		}
	}

	// Fallback to configured port number
	if port == 0 && d.config.MetricsDiscovery.Port > 0 {
		port = int32(d.config.MetricsDiscovery.Port)
	}

	// Default fallback
	if port == 0 {
		port = 8080
	}

	// Check if the service has the required port
	hasPort := false
	for _, p := range service.Spec.Ports {
		if p.Port == port {
			hasPort = true
			break
		}
	}

	if !hasPort {
		d.log.V(1).Info("service does not expose metrics port", 
			"service", service.Name, 
			"namespace", service.Namespace, 
			"port", port)
		return nil
	}

	return &DiscoveredEndpoint{
		ServiceName: service.Name,
		Namespace:   service.Namespace,
		Port:        port,
		Labels:      service.Labels,
	}
}

// buildMetricsURL builds the full metrics URL for an endpoint
func (d *ServiceDiscovery) buildMetricsURL(endpoint DiscoveredEndpoint) string {
	metricsPath := d.config.MetricsDiscovery.MetricsPath
	if metricsPath == "" {
		metricsPath = "/metrics"
	}

	return fmt.Sprintf("http://%s.%s.svc:%d%s",
		endpoint.ServiceName,
		endpoint.Namespace,
		endpoint.Port,
		metricsPath)
}

// DiscoverAllServices discovers all services in the configured namespace
func (d *ServiceDiscovery) DiscoverAllServices(ctx context.Context) ([]corev1.Service, error) {
	services, err := d.client.CoreV1().Services(d.config.MetricsDiscovery.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all services: %w", err)
	}

	return services.Items, nil
}

// GetServiceEndpoints gets endpoints for a specific service
func (d *ServiceDiscovery) GetServiceEndpoints(ctx context.Context, serviceName string) (*corev1.Endpoints, error) {
	endpoints, err := d.client.CoreV1().Endpoints(d.config.MetricsDiscovery.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints for service %s: %w", serviceName, err)
	}

	return endpoints, nil
}

// ValidateService checks if a service is suitable for metrics collection
func (d *ServiceDiscovery) ValidateService(service *corev1.Service) bool {
	// Check if service has required labels (if using label selector)
	if d.config.MetricsDiscovery.ServiceLabelSelector != "" {
		selector, err := labels.Parse(d.config.MetricsDiscovery.ServiceLabelSelector)
		if err != nil {
			d.log.Error(err, "invalid label selector")
			return false
		}

		if !selector.Matches(labels.Set(service.Labels)) {
			return false
		}
	}

	// Check if service has metrics port
	endpoint := d.serviceToEndpoint(service)
	return endpoint != nil
}

// GetMimirComponents returns a list of common Mimir component services
func (d *ServiceDiscovery) GetMimirComponents() []string {
	return []string{
		"mimir-distributor",
		"mimir-ingester",
		"mimir-querier",
		"mimir-query-frontend",
		"mimir-query-scheduler",
		"mimir-compactor",
		"mimir-store-gateway",
		"mimir-ruler",
		"mimir-alertmanager",
		"mimir-nginx",
	}
}

// DiscoverMimirServices discovers services that match common Mimir component patterns
func (d *ServiceDiscovery) DiscoverMimirServices(ctx context.Context) ([]DiscoveredEndpoint, error) {
	var endpoints []DiscoveredEndpoint

	// Get all services in the namespace
	services, err := d.DiscoverAllServices(ctx)
	if err != nil {
		return nil, err
	}

	mimirComponents := d.GetMimirComponents()

	for _, service := range services {
		// Check if service name matches Mimir component patterns
		isMimirService := false
		for _, component := range mimirComponents {
			if service.Name == component {
				isMimirService = true
				break
			}
		}

		// Also check for services with "mimir" in the name
		if !isMimirService && containsString(service.Name, "mimir") {
			isMimirService = true
		}

		if isMimirService && d.ValidateService(&service) {
			endpoint := d.serviceToEndpoint(&service)
			if endpoint != nil {
				endpoints = append(endpoints, *endpoint)
			}
		}
	}

	d.log.Info("discovered Mimir services", "count", len(endpoints))

	return endpoints, nil
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr ||
		  contains(s, substr)))
}

// contains is a simple substring search
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 