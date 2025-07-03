package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"

	// Kubernetes imports
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// AutonomousScanner provides comprehensive AI-enabled Mimir infrastructure scanning
type AutonomousScanner struct {
	client    kubernetes.Interface
	config    *config.Config
	log       logr.Logger
	namespace string
}

// MimirInfrastructure represents the complete Mimir infrastructure discovery
type MimirInfrastructure struct {
	Namespace       string                          `json:"namespace"`
	Components      map[string]*MimirComponent      `json:"components"`
	Tenants         map[string]*TenantConfiguration `json:"tenants"`
	Metrics         *MetricsDiscovery               `json:"metrics"`
	Resources       *ResourceInventory              `json:"resources"`
	Configuration   *ConfigurationScan              `json:"configuration"`
	Health          *InfrastructureHealth           `json:"health"`
	Recommendations []InfrastructureRecommendation  `json:"recommendations"`
	LastScan        time.Time                       `json:"lastScan"`
}

// MimirComponent represents a discovered Mimir component
type MimirComponent struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Role          string                 `json:"role"`
	Status        string                 `json:"status"`
	Replicas      int32                  `json:"replicas"`
	ReadyReplicas int32                  `json:"readyReplicas"`
	Services      []ServiceInfo          `json:"services"`
	Endpoints     []EndpointInfo         `json:"endpoints"`
	MetricsURLs   []string               `json:"metricsUrls"`
	Configuration map[string]interface{} `json:"configuration"`
	Resources     ResourceUsage          `json:"resources"`
	Health        ComponentHealth        `json:"health"`
}

// TenantConfiguration represents a discovered tenant and its limits
type TenantConfiguration struct {
	TenantID          string                 `json:"tenantId"`
	Source            string                 `json:"source"`
	Limits            map[string]interface{} `json:"limits"`
	CurrentUsage      map[string]float64     `json:"currentUsage"`
	RecommendedLimits map[string]interface{} `json:"recommendedLimits"`
	Status            string                 `json:"status"`
	LastSeen          time.Time              `json:"lastSeen"`
}

// MetricsDiscovery represents all discovered metrics endpoints
type MetricsDiscovery struct {
	Endpoints     []MetricsEndpoint   `json:"endpoints"`
	Categories    map[string][]string `json:"categories"`
	TenantMetrics map[string][]string `json:"tenantMetrics"`
	HealthMetrics []string            `json:"healthMetrics"`
}

// ResourceInventory represents all Kubernetes resources in the namespace
type ResourceInventory struct {
	Deployments  []ResourceSummary `json:"deployments"`
	StatefulSets []ResourceSummary `json:"statefulSets"`
	Services     []ResourceSummary `json:"services"`
	ConfigMaps   []ResourceSummary `json:"configMaps"`
	Secrets      []ResourceSummary `json:"secrets"`
	Ingresses    []ResourceSummary `json:"ingresses"`
	PVCs         []ResourceSummary `json:"pvcs"`
	Pods         []ResourceSummary `json:"pods"`
}

// ConfigurationScan represents all Mimir configuration discovered
type ConfigurationScan struct {
	RuntimeOverrides   map[string]interface{} `json:"runtimeOverrides"`
	MainConfig         map[string]interface{} `json:"mainConfig"`
	TenantConfigs      map[string]interface{} `json:"tenantConfigs"`
	CompactorConfig    map[string]interface{} `json:"compactorConfig"`
	AlertmanagerConfig map[string]interface{} `json:"alertmanagerConfig"`
	ConfigSources      []ConfigSource         `json:"configSources"`
}

// InfrastructureRecommendation represents infrastructure-specific AI recommendations
type InfrastructureRecommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Action      string    `json:"action"`
	Component   string    `json:"component"`
	Tenant      string    `json:"tenant,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Supporting types
type ServiceInfo struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIp"`
	Ports     map[string]int32  `json:"ports"`
	Labels    map[string]string `json:"labels"`
}

type EndpointInfo struct {
	Service   string   `json:"service"`
	Addresses []string `json:"addresses"`
	Ports     []int32  `json:"ports"`
	Ready     bool     `json:"ready"`
}

type ComponentHealth struct {
	Status    string             `json:"status"`
	Issues    []string           `json:"issues"`
	Metrics   map[string]float64 `json:"metrics"`
	LastCheck time.Time          `json:"lastCheck"`
}

type InfrastructureHealth struct {
	OverallStatus   string                     `json:"overallStatus"`
	ComponentHealth map[string]ComponentHealth `json:"componentHealth"`
	Issues          []string                   `json:"issues"`
	Score           float64                    `json:"score"`
}

type MetricsEndpoint struct {
	URL        string            `json:"url"`
	Component  string            `json:"component"`
	Port       int32             `json:"port"`
	Path       string            `json:"path"`
	Labels     map[string]string `json:"labels"`
	Accessible bool              `json:"accessible"`
}

type ResourceSummary struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Age         time.Duration          `json:"age"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Details     map[string]interface{} `json:"details"`
}

// Note: ResourceUsage is already defined in health_scanner.go

type ConfigSource struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Source string   `json:"source"`
	Keys   []string `json:"keys"`
}

// NewAutonomousScanner creates a new comprehensive Mimir scanner
func NewAutonomousScanner(client kubernetes.Interface, cfg *config.Config, log logr.Logger) *AutonomousScanner {
	return &AutonomousScanner{
		client:    client,
		config:    cfg,
		log:       log.WithName("autonomous-scanner"),
		namespace: cfg.Mimir.Namespace,
	}
}

// ScanMimirInfrastructure performs a comprehensive scan of the entire Mimir infrastructure
func (s *AutonomousScanner) ScanMimirInfrastructure(ctx context.Context) (*MimirInfrastructure, error) {
	s.log.Info("Starting comprehensive Mimir infrastructure scan", "namespace", s.namespace)

	infrastructure := &MimirInfrastructure{
		Namespace:  s.namespace,
		Components: make(map[string]*MimirComponent),
		Tenants:    make(map[string]*TenantConfiguration),
		LastScan:   time.Now(),
	}

	// Step 1: Discover all Kubernetes resources
	s.log.Info("Step 1: Scanning Kubernetes resources")
	resources, err := s.scanKubernetesResources(ctx)
	if err != nil {
		s.log.Error(err, "Failed to scan Kubernetes resources")
		return nil, fmt.Errorf("failed to scan Kubernetes resources: %w", err)
	}
	infrastructure.Resources = resources

	// Step 2: Identify and analyze Mimir components
	s.log.Info("Step 2: Identifying Mimir components")
	components, err := s.identifyMimirComponents(ctx, resources)
	if err != nil {
		s.log.Error(err, "Failed to identify Mimir components")
		return nil, fmt.Errorf("failed to identify Mimir components: %w", err)
	}
	infrastructure.Components = components

	// Step 3: Discover metrics endpoints
	s.log.Info("Step 3: Discovering metrics endpoints")
	metrics, err := s.discoverMetricsEndpoints(ctx, components)
	if err != nil {
		s.log.Error(err, "Failed to discover metrics endpoints")
		return nil, fmt.Errorf("failed to discover metrics endpoints: %w", err)
	}
	infrastructure.Metrics = metrics

	// Step 4: Extract configuration and tenant information
	s.log.Info("Step 4: Extracting configuration and tenant information")
	configuration, tenants, err := s.extractConfiguration(ctx)
	if err != nil {
		s.log.Error(err, "Failed to extract configuration")
		return nil, fmt.Errorf("failed to extract configuration: %w", err)
	}
	infrastructure.Configuration = configuration
	infrastructure.Tenants = tenants

	// Step 5: Assess infrastructure health
	s.log.Info("Step 5: Assessing infrastructure health")
	health, err := s.assessInfrastructureHealth(ctx, infrastructure)
	if err != nil {
		s.log.Error(err, "Failed to assess infrastructure health")
		return nil, fmt.Errorf("failed to assess infrastructure health: %w", err)
	}
	infrastructure.Health = health

	// Step 6: Generate AI recommendations
	s.log.Info("Step 6: Generating AI recommendations")
	recommendations := s.generateAIRecommendations(infrastructure)
	infrastructure.Recommendations = recommendations

	s.log.Info("Comprehensive Mimir infrastructure scan completed",
		"components", len(infrastructure.Components),
		"tenants", len(infrastructure.Tenants),
		"metrics_endpoints", len(infrastructure.Metrics.Endpoints),
		"recommendations", len(infrastructure.Recommendations))

	return infrastructure, nil
}

// scanKubernetesResources scans all Kubernetes resources in the namespace
func (s *AutonomousScanner) scanKubernetesResources(ctx context.Context) (*ResourceInventory, error) {
	inventory := &ResourceInventory{}

	// Scan Deployments
	deployments, err := s.client.AppsV1().Deployments(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	for _, d := range deployments.Items {
		inventory.Deployments = append(inventory.Deployments, ResourceSummary{
			Name:        d.Name,
			Type:        "Deployment",
			Status:      s.getDeploymentStatus(&d),
			Age:         time.Since(d.CreationTimestamp.Time),
			Labels:      d.Labels,
			Annotations: d.Annotations,
			Details: map[string]interface{}{
				"replicas":      d.Status.Replicas,
				"readyReplicas": d.Status.ReadyReplicas,
				"strategy":      d.Spec.Strategy.Type,
			},
		})
	}

	// Scan StatefulSets
	statefulSets, err := s.client.AppsV1().StatefulSets(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}
	for _, ss := range statefulSets.Items {
		inventory.StatefulSets = append(inventory.StatefulSets, ResourceSummary{
			Name:        ss.Name,
			Type:        "StatefulSet",
			Status:      s.getStatefulSetStatus(&ss),
			Age:         time.Since(ss.CreationTimestamp.Time),
			Labels:      ss.Labels,
			Annotations: ss.Annotations,
			Details: map[string]interface{}{
				"replicas":      ss.Status.Replicas,
				"readyReplicas": ss.Status.ReadyReplicas,
				"serviceName":   ss.Spec.ServiceName,
			},
		})
	}

	// Scan Services
	services, err := s.client.CoreV1().Services(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	for _, svc := range services.Items {
		inventory.Services = append(inventory.Services, ResourceSummary{
			Name:        svc.Name,
			Type:        "Service",
			Status:      "Active",
			Age:         time.Since(svc.CreationTimestamp.Time),
			Labels:      svc.Labels,
			Annotations: svc.Annotations,
			Details: map[string]interface{}{
				"type":      svc.Spec.Type,
				"clusterIP": svc.Spec.ClusterIP,
				"ports":     len(svc.Spec.Ports),
			},
		})
	}

	// Scan ConfigMaps
	configMaps, err := s.client.CoreV1().ConfigMaps(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}
	for _, cm := range configMaps.Items {
		inventory.ConfigMaps = append(inventory.ConfigMaps, ResourceSummary{
			Name:        cm.Name,
			Type:        "ConfigMap",
			Status:      "Active",
			Age:         time.Since(cm.CreationTimestamp.Time),
			Labels:      cm.Labels,
			Annotations: cm.Annotations,
			Details: map[string]interface{}{
				"dataKeys": len(cm.Data),
			},
		})
	}

	// Scan Secrets
	secrets, err := s.client.CoreV1().Secrets(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, secret := range secrets.Items {
		inventory.Secrets = append(inventory.Secrets, ResourceSummary{
			Name:        secret.Name,
			Type:        "Secret",
			Status:      "Active",
			Age:         time.Since(secret.CreationTimestamp.Time),
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
			Details: map[string]interface{}{
				"type":     secret.Type,
				"dataKeys": len(secret.Data),
			},
		})
	}

	// Scan Ingresses
	ingresses, err := s.client.NetworkingV1().Ingresses(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}
	for _, ing := range ingresses.Items {
		inventory.Ingresses = append(inventory.Ingresses, ResourceSummary{
			Name:        ing.Name,
			Type:        "Ingress",
			Status:      s.getIngressStatus(&ing),
			Age:         time.Since(ing.CreationTimestamp.Time),
			Labels:      ing.Labels,
			Annotations: ing.Annotations,
			Details: map[string]interface{}{
				"rules": len(ing.Spec.Rules),
			},
		})
	}

	// Scan PVCs
	pvcs, err := s.client.CoreV1().PersistentVolumeClaims(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pvcs: %w", err)
	}
	for _, pvc := range pvcs.Items {
		inventory.PVCs = append(inventory.PVCs, ResourceSummary{
			Name:        pvc.Name,
			Type:        "PVC",
			Status:      string(pvc.Status.Phase),
			Age:         time.Since(pvc.CreationTimestamp.Time),
			Labels:      pvc.Labels,
			Annotations: pvc.Annotations,
			Details: map[string]interface{}{
				"storageClass": pvc.Spec.StorageClassName,
				"capacity":     pvc.Status.Capacity,
			},
		})
	}

	// Scan Pods
	pods, err := s.client.CoreV1().Pods(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	for _, pod := range pods.Items {
		inventory.Pods = append(inventory.Pods, ResourceSummary{
			Name:        pod.Name,
			Type:        "Pod",
			Status:      string(pod.Status.Phase),
			Age:         time.Since(pod.CreationTimestamp.Time),
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
			Details: map[string]interface{}{
				"node":       pod.Spec.NodeName,
				"containers": len(pod.Spec.Containers),
				"restarts":   s.getPodRestartCount(&pod),
			},
		})
	}

	return inventory, nil
}

// identifyMimirComponents identifies and categorizes Mimir components
func (s *AutonomousScanner) identifyMimirComponents(ctx context.Context, resources *ResourceInventory) (map[string]*MimirComponent, error) {
	components := make(map[string]*MimirComponent)

	// Known Mimir component patterns
	mimirComponents := map[string]string{
		"distributor":        "ingestion",
		"ingester":           "ingestion",
		"querier":            "query",
		"query-frontend":     "query",
		"query-scheduler":    "query",
		"compactor":          "storage",
		"store-gateway":      "storage",
		"ruler":              "alerting",
		"alertmanager":       "alerting",
		"nginx":              "gateway",
		"gateway":            "gateway",
		"overrides-exporter": "monitoring",
	}

	// Process deployments and statefulsets
	allWorkloads := []ResourceSummary{}
	allWorkloads = append(allWorkloads, resources.Deployments...)
	allWorkloads = append(allWorkloads, resources.StatefulSets...)

	for _, workload := range allWorkloads {
		componentType := s.identifyComponentType(workload.Name, workload.Labels)
		if componentType != "" {
			role := mimirComponents[componentType]
			if role == "" {
				role = "unknown"
			}

			var replicas, readyReplicas int32
			if val, ok := workload.Details["replicas"]; ok {
				replicas = val.(int32)
			}
			if val, ok := workload.Details["readyReplicas"]; ok {
				readyReplicas = val.(int32)
			}

			component := &MimirComponent{
				Name:          workload.Name,
				Type:          componentType,
				Role:          role,
				Status:        workload.Status,
				Replicas:      replicas,
				ReadyReplicas: readyReplicas,
				Services:      []ServiceInfo{},
				Endpoints:     []EndpointInfo{},
				MetricsURLs:   []string{},
				Configuration: make(map[string]interface{}),
				Health: ComponentHealth{
					Status:    s.getComponentHealthStatus(workload.Status, replicas, readyReplicas),
					Issues:    []string{},
					Metrics:   make(map[string]float64),
					LastCheck: time.Now(),
				},
			}

			// Find associated services
			for _, service := range resources.Services {
				if s.isServiceForComponent(service, workload) {
					serviceInfo := ServiceInfo{
						Name:      service.Name,
						Type:      service.Details["type"].(string),
						ClusterIP: service.Details["clusterIP"].(string),
						Ports:     make(map[string]int32),
						Labels:    service.Labels,
					}
					component.Services = append(component.Services, serviceInfo)
				}
			}

			components[workload.Name] = component
		}
	}

	return components, nil
}

// identifyComponentType identifies what type of Mimir component this is
func (s *AutonomousScanner) identifyComponentType(name string, labels map[string]string) string {
	name = strings.ToLower(name)

	// Check common label patterns first
	if labels != nil {
		if component, ok := labels["app.kubernetes.io/component"]; ok {
			return component
		}
		if component, ok := labels["component"]; ok {
			return component
		}
	}

	// Fallback to name-based detection
	componentPatterns := []string{
		"distributor", "ingester", "querier", "query-frontend", "query-scheduler",
		"compactor", "store-gateway", "ruler", "alertmanager", "nginx", "gateway",
		"overrides-exporter",
	}

	for _, pattern := range componentPatterns {
		if strings.Contains(name, pattern) {
			return pattern
		}
	}

	// If it contains "mimir" but no specific component, it's likely a Mimir component
	if strings.Contains(name, "mimir") {
		return "unknown-mimir-component"
	}

	return ""
}

// discoverMetricsEndpoints discovers all metrics endpoints from components
func (s *AutonomousScanner) discoverMetricsEndpoints(ctx context.Context, components map[string]*MimirComponent) (*MetricsDiscovery, error) {
	discovery := &MetricsDiscovery{
		Endpoints:     []MetricsEndpoint{},
		Categories:    make(map[string][]string),
		TenantMetrics: make(map[string][]string),
		HealthMetrics: []string{},
	}

	// Standard metrics paths to try
	metricsPaths := []string{"/metrics", "/prometheus/metrics", "/debug/pprof/metrics"}

	for _, component := range components {
		for _, service := range component.Services {
			for _, path := range metricsPaths {
				// Try common metrics ports
				ports := []int32{8080, 9090, 8081, 3000, 80}

				for _, port := range ports {
					endpoint := MetricsEndpoint{
						URL:        fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s", service.Name, s.namespace, port, path),
						Component:  component.Name,
						Port:       port,
						Path:       path,
						Labels:     service.Labels,
						Accessible: true, // We'll verify this later
					}
					discovery.Endpoints = append(discovery.Endpoints, endpoint)
					component.MetricsURLs = append(component.MetricsURLs, endpoint.URL)
				}
			}
		}
	}

	// Categorize metrics by component role
	for _, component := range components {
		role := component.Role
		if discovery.Categories[role] == nil {
			discovery.Categories[role] = []string{}
		}
		discovery.Categories[role] = append(discovery.Categories[role], component.MetricsURLs...)
	}

	return discovery, nil
}

// extractConfiguration extracts all Mimir configuration from ConfigMaps and Secrets
func (s *AutonomousScanner) extractConfiguration(ctx context.Context) (*ConfigurationScan, map[string]*TenantConfiguration, error) {
	configScan := &ConfigurationScan{
		RuntimeOverrides:   make(map[string]interface{}),
		MainConfig:         make(map[string]interface{}),
		TenantConfigs:      make(map[string]interface{}),
		CompactorConfig:    make(map[string]interface{}),
		AlertmanagerConfig: make(map[string]interface{}),
		ConfigSources:      []ConfigSource{},
	}

	tenantConfigs := make(map[string]*TenantConfiguration)

	// Get all ConfigMaps
	configMaps, err := s.client.CoreV1().ConfigMaps(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	// Process each ConfigMap
	for _, cm := range configMaps.Items {
		configSource := ConfigSource{
			Name:   cm.Name,
			Type:   "ConfigMap",
			Source: s.namespace,
			Keys:   []string{},
		}

		for key, data := range cm.Data {
			configSource.Keys = append(configSource.Keys, key)

			// Try to parse as YAML
			var parsedData interface{}
			if err := yaml.Unmarshal([]byte(data), &parsedData); err == nil {
				// Categorize configuration based on ConfigMap name and key
				if strings.Contains(cm.Name, "runtime") || strings.Contains(cm.Name, "override") {
					configScan.RuntimeOverrides[cm.Name+"."+key] = parsedData

					// Extract tenant configurations
					if tenants := s.extractTenantsFromConfig(parsedData); len(tenants) > 0 {
						for tenantID, tenantConfig := range tenants {
							tenantConfigs[tenantID] = &TenantConfiguration{
								TenantID:          tenantID,
								Source:            fmt.Sprintf("ConfigMap:%s/%s", cm.Name, key),
								Limits:            tenantConfig,
								CurrentUsage:      make(map[string]float64),
								RecommendedLimits: make(map[string]interface{}),
								Status:            "discovered",
								LastSeen:          time.Now(),
							}
						}
					}
				} else if strings.Contains(cm.Name, "config") || strings.Contains(cm.Name, "mimir") {
					configScan.MainConfig[cm.Name+"."+key] = parsedData
				} else if strings.Contains(cm.Name, "compactor") {
					configScan.CompactorConfig[cm.Name+"."+key] = parsedData
				} else if strings.Contains(cm.Name, "alertmanager") {
					configScan.AlertmanagerConfig[cm.Name+"."+key] = parsedData
				}
			}
		}

		configScan.ConfigSources = append(configScan.ConfigSources, configSource)
	}

	// If no tenants found in config, generate synthetic ones for testing
	if len(tenantConfigs) == 0 {
		s.log.Info("No tenants found in configuration, generating synthetic tenants for testing")
		syntheticTenants := []string{"tenant-a", "tenant-b", "tenant-c", "enterprise-customer", "small-team"}

		for _, tenantID := range syntheticTenants {
			tenantConfigs[tenantID] = &TenantConfiguration{
				TenantID: tenantID,
				Source:   "synthetic",
				Limits: map[string]interface{}{
					"ingestion_rate":             25000,
					"ingestion_burst_size":       250000,
					"max_global_series_per_user": 150000,
					"max_samples_per_query":      1000000,
				},
				CurrentUsage: map[string]float64{
					"ingestion_rate":             15000 + float64(len(tenantID)*1000),
					"max_global_series_per_user": 75000 + float64(len(tenantID)*5000),
				},
				RecommendedLimits: make(map[string]interface{}),
				Status:            "synthetic",
				LastSeen:          time.Now(),
			}
		}
	}

	return configScan, tenantConfigs, nil
}

// extractTenantsFromConfig extracts tenant configurations from parsed YAML
func (s *AutonomousScanner) extractTenantsFromConfig(data interface{}) map[string]map[string]interface{} {
	tenants := make(map[string]map[string]interface{})

	// Handle map[string]interface{} (parsed YAML)
	if configMap, ok := data.(map[string]interface{}); ok {
		for key, value := range configMap {
			// Skip common non-tenant keys
			if s.isSystemKey(key) {
				continue
			}

			// If the value is a map, it's likely a tenant configuration
			if tenantConfig, ok := value.(map[string]interface{}); ok {
				tenants[key] = tenantConfig
			}
		}
	}

	return tenants
}

// isSystemKey checks if a key is a system configuration key (not a tenant)
func (s *AutonomousScanner) isSystemKey(key string) bool {
	systemKeys := []string{
		"overrides", "defaults", "global", "common", "shared",
		"ingester", "distributor", "querier", "compactor", "ruler",
		"alertmanager", "store_gateway", "query_frontend", "query_scheduler",
		"limits_config", "server", "memberlist", "storage", "blocks_storage",
		"ruler_storage", "alertmanager_storage", "compactor_config",
	}

	keyLower := strings.ToLower(key)
	for _, sysKey := range systemKeys {
		if keyLower == sysKey || strings.Contains(keyLower, sysKey) {
			return true
		}
	}

	return false
}

// assessInfrastructureHealth assesses the overall health of the Mimir infrastructure
func (s *AutonomousScanner) assessInfrastructureHealth(ctx context.Context, infra *MimirInfrastructure) (*InfrastructureHealth, error) {
	health := &InfrastructureHealth{
		ComponentHealth: make(map[string]ComponentHealth),
		Issues:          []string{},
		Score:           100.0,
	}

	totalComponents := 0
	healthyComponents := 0

	// Assess each component health
	for name, component := range infra.Components {
		componentHealth := component.Health
		health.ComponentHealth[name] = componentHealth
		totalComponents++

		if componentHealth.Status == "healthy" {
			healthyComponents++
		} else {
			health.Issues = append(health.Issues, fmt.Sprintf("Component %s is %s", name, componentHealth.Status))
		}
	}

	// Calculate overall health score
	if totalComponents > 0 {
		health.Score = (float64(healthyComponents) / float64(totalComponents)) * 100.0
	}

	// Determine overall status
	if health.Score >= 90 {
		health.OverallStatus = "healthy"
	} else if health.Score >= 70 {
		health.OverallStatus = "degraded"
	} else {
		health.OverallStatus = "unhealthy"
	}

	return health, nil
}

// generateAIRecommendations generates intelligent recommendations based on the infrastructure scan
func (s *AutonomousScanner) generateAIRecommendations(infra *MimirInfrastructure) []InfrastructureRecommendation {
	var recommendations []InfrastructureRecommendation

	// Recommendation 1: Component Health Issues
	for name, component := range infra.Components {
		if component.Health.Status != "healthy" {
			rec := InfrastructureRecommendation{
				ID:          fmt.Sprintf("health-%s", name),
				Type:        "health",
				Priority:    "high",
				Title:       fmt.Sprintf("Component %s Health Issue", name),
				Description: fmt.Sprintf("Component %s is reporting status: %s", name, component.Health.Status),
				Impact:      "May affect query/ingestion performance",
				Action:      "Check component logs and resource allocation",
				Component:   name,
				CreatedAt:   time.Now(),
			}
			recommendations = append(recommendations, rec)
		}
	}

	// Recommendation 2: Tenant Limit Optimization
	for _, tenant := range infra.Tenants {
		if s.needsLimitOptimization(tenant) {
			rec := InfrastructureRecommendation{
				ID:          fmt.Sprintf("limits-%s", tenant.TenantID),
				Type:        "optimization",
				Priority:    "medium",
				Title:       fmt.Sprintf("Optimize Limits for Tenant %s", tenant.TenantID),
				Description: "Current usage patterns suggest limits can be optimized",
				Impact:      "Better resource utilization and cost optimization",
				Action:      "Review and adjust tenant limits based on usage patterns",
				Component:   "limits",
				Tenant:      tenant.TenantID,
				CreatedAt:   time.Now(),
			}
			recommendations = append(recommendations, rec)
		}
	}

	// Recommendation 3: Missing Metrics Endpoints
	if len(infra.Metrics.Endpoints) < len(infra.Components) {
		rec := InfrastructureRecommendation{
			ID:          "metrics-coverage",
			Type:        "monitoring",
			Priority:    "medium",
			Title:       "Incomplete Metrics Coverage",
			Description: "Some components may not have accessible metrics endpoints",
			Impact:      "Limited observability and optimization capability",
			Action:      "Verify metrics endpoints are accessible and properly configured",
			Component:   "monitoring",
			CreatedAt:   time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Recommendation 4: Configuration Optimization
	if len(infra.Configuration.ConfigSources) == 0 {
		rec := InfrastructureRecommendation{
			ID:          "config-missing",
			Type:        "configuration",
			Priority:    "high",
			Title:       "Missing Mimir Configuration",
			Description: "No Mimir configuration found in the namespace",
			Impact:      "Cannot perform limit optimization without configuration access",
			Action:      "Ensure Mimir configuration ConfigMaps are accessible",
			Component:   "configuration",
			CreatedAt:   time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// Helper methods
func (s *AutonomousScanner) getDeploymentStatus(d *appsv1.Deployment) string {
	if d.Status.ReadyReplicas == d.Status.Replicas && d.Status.Replicas > 0 {
		return "ready"
	} else if d.Status.ReadyReplicas > 0 {
		return "partial"
	}
	return "not-ready"
}

func (s *AutonomousScanner) getStatefulSetStatus(ss *appsv1.StatefulSet) string {
	if ss.Status.ReadyReplicas == ss.Status.Replicas && ss.Status.Replicas > 0 {
		return "ready"
	} else if ss.Status.ReadyReplicas > 0 {
		return "partial"
	}
	return "not-ready"
}

func (s *AutonomousScanner) getIngressStatus(ing *networkingv1.Ingress) string {
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		return "ready"
	}
	return "pending"
}

func (s *AutonomousScanner) getPodRestartCount(pod *corev1.Pod) int32 {
	var totalRestarts int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += containerStatus.RestartCount
	}
	return totalRestarts
}

func (s *AutonomousScanner) getComponentHealthStatus(status string, replicas, readyReplicas int32) string {
	if status == "ready" && replicas == readyReplicas && replicas > 0 {
		return "healthy"
	} else if readyReplicas > 0 {
		return "degraded"
	}
	return "unhealthy"
}

func (s *AutonomousScanner) isServiceForComponent(service ResourceSummary, workload ResourceSummary) bool {
	if strings.Contains(service.Name, workload.Name) {
		return true
	}

	for key, value := range workload.Labels {
		if service.Labels[key] == value {
			return true
		}
	}

	return false
}

func (s *AutonomousScanner) needsLimitOptimization(tenant *TenantConfiguration) bool {
	// Simple heuristic: if current usage is significantly different from limits
	for limitName, limitValue := range tenant.Limits {
		if usage, exists := tenant.CurrentUsage[limitName]; exists {
			if limitFloat, ok := limitValue.(float64); ok {
				usageRatio := usage / limitFloat
				// Recommend optimization if usage is very low (<30%) or high (>80%)
				if usageRatio < 0.3 || usageRatio > 0.8 {
					return true
				}
			}
		}
	}
	return false
}
