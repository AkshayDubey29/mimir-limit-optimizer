package discovery

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// TenantNamespaceInfo represents detailed information about a tenant namespace
type TenantNamespaceInfo struct {
	Name              string                `json:"name"`
	Namespace         string                `json:"namespace"`
	Status            string                `json:"status"`
	CreationTimestamp time.Time             `json:"creation_timestamp"`
	Labels            map[string]string     `json:"labels"`
	Annotations       map[string]string     `json:"annotations"`
	ResourceQuota     *ResourceQuotaInfo    `json:"resource_quota,omitempty"`
	MimirComponents   []MimirComponentInfo  `json:"mimir_components"`
	Services          []TenantServiceInfo   `json:"services"`
	Deployments       []DeploymentInfo      `json:"deployments"`
	Pods              []PodInfo             `json:"pods"`
	ConfigMaps        []ConfigMapInfo       `json:"config_maps"`
	Secrets           []SecretInfo          `json:"secrets"`
	IngressEndpoints  []IngressEndpointInfo `json:"ingress_endpoints"`
	MetricsEndpoints  []string              `json:"metrics_endpoints"`
	IngestionRate     float64               `json:"ingestion_rate"`
	ActiveSeries      int64                 `json:"active_series"`
	LastActivity      time.Time             `json:"last_activity"`
	HealthScore       int                   `json:"health_score"`
	ArchitectureFlow  *ArchitectureFlowInfo `json:"architecture_flow"`
}

// ResourceQuotaInfo represents resource quota information for a namespace
type ResourceQuotaInfo struct {
	Name   string            `json:"name"`
	Hard   map[string]string `json:"hard"`
	Used   map[string]string `json:"used"`
	Status string            `json:"status"`
}

// MimirComponentInfo represents a Mimir component in the namespace
type MimirComponentInfo struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"` // ingester, distributor, querier, etc.
	Status        string            `json:"status"`
	Replicas      int32             `json:"replicas"`
	ReadyReplicas int32             `json:"ready_replicas"`
	Image         string            `json:"image"`
	Labels        map[string]string `json:"labels"`
	Endpoints     []string          `json:"endpoints"`
}

// TenantServiceInfo represents a service in the namespace (renamed to avoid conflict)
type TenantServiceInfo struct {
	Name      string                  `json:"name"`
	Type      string                  `json:"type"`
	ClusterIP string                  `json:"cluster_ip"`
	Ports     []TenantServicePortInfo `json:"ports"`
	Selector  map[string]string       `json:"selector"`
	Labels    map[string]string       `json:"labels"`
}

// TenantServicePortInfo represents a service port (renamed to avoid conflict)
type TenantServicePortInfo struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// DeploymentInfo represents a deployment in the namespace
type DeploymentInfo struct {
	Name            string            `json:"name"`
	Replicas        int32             `json:"replicas"`
	ReadyReplicas   int32             `json:"ready_replicas"`
	UpdatedReplicas int32             `json:"updated_replicas"`
	Image           string            `json:"image"`
	Labels          map[string]string `json:"labels"`
	Status          string            `json:"status"`
}

// PodInfo represents a pod in the namespace
type PodInfo struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	Ready             bool              `json:"ready"`
	RestartCount      int32             `json:"restart_count"`
	NodeName          string            `json:"node_name"`
	PodIP             string            `json:"pod_ip"`
	ContainerStatuses []ContainerStatus `json:"container_statuses"`
	Labels            map[string]string `json:"labels"`
}

// ContainerStatus represents a container's status within a pod
type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
	Image        string `json:"image"`
}

// ConfigMapInfo represents a config map in the namespace
type ConfigMapInfo struct {
	Name        string            `json:"name"`
	DataKeys    []string          `json:"data_keys"`
	Labels      map[string]string `json:"labels"`
	IsMimirConf bool              `json:"is_mimir_conf"`
}

// SecretInfo represents a secret in the namespace
type SecretInfo struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

// IngressEndpointInfo represents an ingress endpoint
type IngressEndpointInfo struct {
	Host string `json:"host"`
	Path string `json:"path"`
	Port int32  `json:"port"`
	TLS  bool   `json:"tls"`
}

// ArchitectureFlowInfo represents the flow of metrics through Mimir components
type ArchitectureFlowInfo struct {
	Distributors   []ComponentFlowInfo `json:"distributors"`
	Ingesters      []ComponentFlowInfo `json:"ingesters"`
	Queriers       []ComponentFlowInfo `json:"queriers"`
	QueryFrontends []ComponentFlowInfo `json:"query_frontends"`
	Compactors     []ComponentFlowInfo `json:"compactors"`
	StoreGateways  []ComponentFlowInfo `json:"store_gateways"`
	Rulers         []ComponentFlowInfo `json:"rulers"`
	Alertmanagers  []ComponentFlowInfo `json:"alertmanagers"`
	Flow           []FlowStep          `json:"flow"`
}

// ComponentFlowInfo represents a component in the flow
type ComponentFlowInfo struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Connections int     `json:"connections"`
	Load        float64 `json:"load"`
	Endpoint    string  `json:"endpoint"`
}

// FlowStep represents a step in the metrics flow
type FlowStep struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Type       string  `json:"type"` // ingestion, query, alert, etc.
	Active     bool    `json:"active"`
	Throughput float64 `json:"throughput"`
	Latency    float64 `json:"latency"`
}

// NamespaceScanner handles scanning of tenant namespaces
type NamespaceScanner struct {
	client kubernetes.Interface
	config *config.Config
	log    logr.Logger
}

// NewNamespaceScanner creates a new NamespaceScanner
func NewNamespaceScanner(client kubernetes.Interface, cfg *config.Config, log logr.Logger) *NamespaceScanner {
	return &NamespaceScanner{
		client: client,
		config: cfg,
		log:    log,
	}
}

// ScanAllTenantNamespaces scans all tenant namespaces and returns detailed information
func (ns *NamespaceScanner) ScanAllTenantNamespaces(ctx context.Context) ([]TenantNamespaceInfo, error) {
	// Get all namespaces
	namespaces, err := ns.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var tenantNamespaces []TenantNamespaceInfo

	for _, namespace := range namespaces.Items {
		// Skip system namespaces
		if ns.isSystemNamespace(namespace.Name) {
			continue
		}

		// Check if this looks like a tenant namespace
		if ns.isTenantNamespace(&namespace) {
			tenantInfo, err := ns.scanTenantNamespace(ctx, &namespace)
			if err != nil {
				ns.log.Error(err, "failed to scan tenant namespace", "namespace", namespace.Name)
				continue
			}
			tenantNamespaces = append(tenantNamespaces, *tenantInfo)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(tenantNamespaces, func(i, j int) bool {
		return tenantNamespaces[i].CreationTimestamp.After(tenantNamespaces[j].CreationTimestamp)
	})

	ns.log.Info("scanned tenant namespaces", "count", len(tenantNamespaces))
	return tenantNamespaces, nil
}

// scanTenantNamespace scans a single tenant namespace for detailed information
func (ns *NamespaceScanner) scanTenantNamespace(ctx context.Context, namespace *corev1.Namespace) (*TenantNamespaceInfo, error) {
	tenantInfo := &TenantNamespaceInfo{
		Name:              namespace.Name,
		Namespace:         namespace.Name,
		Status:            string(namespace.Status.Phase),
		CreationTimestamp: namespace.CreationTimestamp.Time,
		Labels:            namespace.Labels,
		Annotations:       namespace.Annotations,
	}

	// Scan resource quota
	tenantInfo.ResourceQuota = ns.scanResourceQuota(ctx, namespace.Name)

	// Scan services
	tenantInfo.Services = ns.scanServices(ctx, namespace.Name)

	// Scan deployments
	tenantInfo.Deployments = ns.scanDeployments(ctx, namespace.Name)

	// Scan pods
	tenantInfo.Pods = ns.scanPods(ctx, namespace.Name)

	// Scan config maps
	tenantInfo.ConfigMaps = ns.scanConfigMaps(ctx, namespace.Name)

	// Scan secrets
	tenantInfo.Secrets = ns.scanSecrets(ctx, namespace.Name)

	// Identify Mimir components
	tenantInfo.MimirComponents = ns.identifyMimirComponents(tenantInfo.Deployments, tenantInfo.Services)

	// Build metrics endpoints
	tenantInfo.MetricsEndpoints = ns.buildMetricsEndpoints(tenantInfo.Services)

	// Build architecture flow
	tenantInfo.ArchitectureFlow = ns.buildArchitectureFlow(tenantInfo.MimirComponents)

	// Calculate health score
	tenantInfo.HealthScore = ns.calculateHealthScore(tenantInfo)

	// Set last activity
	tenantInfo.LastActivity = ns.getLastActivity(tenantInfo)

	return tenantInfo, nil
}

// isSystemNamespace checks if a namespace is a system namespace
func (ns *NamespaceScanner) isSystemNamespace(namespaceName string) bool {
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default",
		"kubernetes-dashboard",
		"cert-manager",
		"ingress-nginx",
		"istio-system",
		"monitoring",
		"prometheus",
		"grafana",
		"logging",
		"elastic-system",
	}

	for _, system := range systemNamespaces {
		if namespaceName == system {
			return true
		}
	}
	return false
}

// isTenantNamespace checks if a namespace is a tenant namespace
func (ns *NamespaceScanner) isTenantNamespace(namespace *corev1.Namespace) bool {
	name := namespace.Name
	labels := namespace.Labels

	// Check for tenant-specific labels
	if labels != nil {
		if tenant, exists := labels["tenant"]; exists && tenant != "" {
			return true
		}
		if tenant, exists := labels["mimir.io/tenant"]; exists && tenant != "" {
			return true
		}
		if tenant, exists := labels["app.kubernetes.io/name"]; exists && strings.Contains(tenant, "mimir") {
			return true
		}
	}

	// Check for tenant-like naming patterns
	tenantPatterns := []string{
		"tenant-",
		"mimir-",
		"-tenant",
		"-mimir",
	}

	for _, pattern := range tenantPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

// scanResourceQuota scans resource quota for a namespace
func (ns *NamespaceScanner) scanResourceQuota(ctx context.Context, namespaceName string) *ResourceQuotaInfo {
	quotas, err := ns.client.CoreV1().ResourceQuotas(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil || len(quotas.Items) == 0 {
		return nil
	}

	quota := quotas.Items[0]
	return &ResourceQuotaInfo{
		Name:   quota.Name,
		Hard:   resourceListToMap(quota.Status.Hard),
		Used:   resourceListToMap(quota.Status.Used),
		Status: "active",
	}
}

// scanServices scans services in a namespace
func (ns *NamespaceScanner) scanServices(ctx context.Context, namespaceName string) []TenantServiceInfo {
	services, err := ns.client.CoreV1().Services(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var serviceInfos []TenantServiceInfo
	for _, service := range services.Items {
		ports := make([]TenantServicePortInfo, len(service.Spec.Ports))
		for i, port := range service.Spec.Ports {
			ports[i] = TenantServicePortInfo{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort.String(),
				Protocol:   string(port.Protocol),
			}
		}

		serviceInfos = append(serviceInfos, TenantServiceInfo{
			Name:      service.Name,
			Type:      string(service.Spec.Type),
			ClusterIP: service.Spec.ClusterIP,
			Ports:     ports,
			Selector:  service.Spec.Selector,
			Labels:    service.Labels,
		})
	}

	return serviceInfos
}

// scanDeployments scans deployments in a namespace
func (ns *NamespaceScanner) scanDeployments(ctx context.Context, namespaceName string) []DeploymentInfo {
	deployments, err := ns.client.AppsV1().Deployments(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		image := ""
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			image = deployment.Spec.Template.Spec.Containers[0].Image
		}

		status := "Running"
		if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
			status = "Pending"
		}

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Name:            deployment.Name,
			Replicas:        *deployment.Spec.Replicas,
			ReadyReplicas:   deployment.Status.ReadyReplicas,
			UpdatedReplicas: deployment.Status.UpdatedReplicas,
			Image:           image,
			Labels:          deployment.Labels,
			Status:          status,
		})
	}

	return deploymentInfos
}

// scanPods scans pods in a namespace
func (ns *NamespaceScanner) scanPods(ctx context.Context, namespaceName string) []PodInfo {
	pods, err := ns.client.CoreV1().Pods(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		containerStatuses := make([]ContainerStatus, len(pod.Status.ContainerStatuses))
		for i, cs := range pod.Status.ContainerStatuses {
			state := "Running"
			if cs.State.Waiting != nil {
				state = "Waiting"
			} else if cs.State.Terminated != nil {
				state = "Terminated"
			}

			containerStatuses[i] = ContainerStatus{
				Name:         cs.Name,
				Ready:        cs.Ready,
				RestartCount: cs.RestartCount,
				State:        state,
				Image:        cs.Image,
			}
		}

		ready := true
		restartCount := int32(0)
		for _, cs := range containerStatuses {
			if !cs.Ready {
				ready = false
			}
			restartCount += cs.RestartCount
		}

		podInfos = append(podInfos, PodInfo{
			Name:              pod.Name,
			Status:            string(pod.Status.Phase),
			Ready:             ready,
			RestartCount:      restartCount,
			NodeName:          pod.Spec.NodeName,
			PodIP:             pod.Status.PodIP,
			ContainerStatuses: containerStatuses,
			Labels:            pod.Labels,
		})
	}

	return podInfos
}

// scanConfigMaps scans config maps in a namespace
func (ns *NamespaceScanner) scanConfigMaps(ctx context.Context, namespaceName string) []ConfigMapInfo {
	configMaps, err := ns.client.CoreV1().ConfigMaps(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var configMapInfos []ConfigMapInfo
	for _, cm := range configMaps.Items {
		dataKeys := make([]string, 0, len(cm.Data))
		for key := range cm.Data {
			dataKeys = append(dataKeys, key)
		}

		isMimirConf := strings.Contains(cm.Name, "mimir") || strings.Contains(cm.Name, "config")

		configMapInfos = append(configMapInfos, ConfigMapInfo{
			Name:        cm.Name,
			DataKeys:    dataKeys,
			Labels:      cm.Labels,
			IsMimirConf: isMimirConf,
		})
	}

	return configMapInfos
}

// scanSecrets scans secrets in a namespace
func (ns *NamespaceScanner) scanSecrets(ctx context.Context, namespaceName string) []SecretInfo {
	secrets, err := ns.client.CoreV1().Secrets(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var secretInfos []SecretInfo
	for _, secret := range secrets.Items {
		secretInfos = append(secretInfos, SecretInfo{
			Name:   secret.Name,
			Type:   string(secret.Type),
			Labels: secret.Labels,
		})
	}

	return secretInfos
}

// identifyMimirComponents identifies Mimir components from deployments and services
func (ns *NamespaceScanner) identifyMimirComponents(deployments []DeploymentInfo, services []TenantServiceInfo) []MimirComponentInfo {
	var components []MimirComponentInfo

	mimirComponentTypes := []string{
		"distributor",
		"ingester",
		"querier",
		"query-frontend",
		"compactor",
		"store-gateway",
		"ruler",
		"alertmanager",
	}

	for _, deployment := range deployments {
		for _, componentType := range mimirComponentTypes {
			if strings.Contains(strings.ToLower(deployment.Name), componentType) {
				// Find corresponding service
				var endpoints []string
				for _, service := range services {
					if strings.Contains(strings.ToLower(service.Name), componentType) {
						for _, port := range service.Ports {
							endpoint := fmt.Sprintf("%s:%d", service.ClusterIP, port.Port)
							endpoints = append(endpoints, endpoint)
						}
					}
				}

				components = append(components, MimirComponentInfo{
					Name:          deployment.Name,
					Type:          componentType,
					Status:        deployment.Status,
					Replicas:      deployment.Replicas,
					ReadyReplicas: deployment.ReadyReplicas,
					Image:         deployment.Image,
					Labels:        deployment.Labels,
					Endpoints:     endpoints,
				})
				break
			}
		}
	}

	return components
}

// buildMetricsEndpoints builds metrics endpoints from services
func (ns *NamespaceScanner) buildMetricsEndpoints(services []TenantServiceInfo) []string {
	var endpoints []string

	for _, service := range services {
		for _, port := range service.Ports {
			if port.Name == "metrics" || port.Name == "http-metrics" || port.Port == 8080 {
				endpoint := fmt.Sprintf("http://%s:%d/metrics", service.ClusterIP, port.Port)
				endpoints = append(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

// buildArchitectureFlow builds the architecture flow information
func (ns *NamespaceScanner) buildArchitectureFlow(components []MimirComponentInfo) *ArchitectureFlowInfo {
	flow := &ArchitectureFlowInfo{
		Flow: []FlowStep{},
	}

	// Organize components by type
	componentsByType := make(map[string][]ComponentFlowInfo)
	for _, comp := range components {
		flowComp := ComponentFlowInfo{
			Name:        comp.Name,
			Status:      comp.Status,
			Connections: len(comp.Endpoints),
			Load:        float64(comp.ReadyReplicas) / float64(comp.Replicas) * 100,
		}
		if len(comp.Endpoints) > 0 {
			flowComp.Endpoint = comp.Endpoints[0]
		}
		componentsByType[comp.Type] = append(componentsByType[comp.Type], flowComp)
	}

	// Set components
	flow.Distributors = componentsByType["distributor"]
	flow.Ingesters = componentsByType["ingester"]
	flow.Queriers = componentsByType["querier"]
	flow.QueryFrontends = componentsByType["query-frontend"]
	flow.Compactors = componentsByType["compactor"]
	flow.StoreGateways = componentsByType["store-gateway"]
	flow.Rulers = componentsByType["ruler"]
	flow.Alertmanagers = componentsByType["alertmanager"]

	// Build flow steps
	flow.Flow = []FlowStep{
		{From: "Prometheus", To: "Distributor", Type: "ingestion", Active: true, Throughput: 1000, Latency: 5},
		{From: "Distributor", To: "Ingester", Type: "ingestion", Active: true, Throughput: 1000, Latency: 10},
		{From: "Ingester", To: "Store", Type: "storage", Active: true, Throughput: 800, Latency: 20},
		{From: "Query Frontend", To: "Querier", Type: "query", Active: true, Throughput: 200, Latency: 15},
		{From: "Querier", To: "Store Gateway", Type: "query", Active: true, Throughput: 200, Latency: 25},
		{From: "Querier", To: "Ingester", Type: "query", Active: true, Throughput: 150, Latency: 12},
		{From: "Compactor", To: "Store", Type: "compaction", Active: true, Throughput: 50, Latency: 100},
		{From: "Ruler", To: "Alertmanager", Type: "alert", Active: true, Throughput: 10, Latency: 30},
	}

	return flow
}

// calculateHealthScore calculates a health score for the tenant namespace
func (ns *NamespaceScanner) calculateHealthScore(tenantInfo *TenantNamespaceInfo) int {
	score := 100

	// Check deployment health
	for _, deployment := range tenantInfo.Deployments {
		if deployment.ReadyReplicas < deployment.Replicas {
			score -= 10
		}
	}

	// Check pod health
	for _, pod := range tenantInfo.Pods {
		if !pod.Ready {
			score -= 5
		}
		if pod.RestartCount > 5 {
			score -= 3
		}
	}

	// Check Mimir components
	for _, comp := range tenantInfo.MimirComponents {
		if comp.Status != "Running" {
			score -= 15
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// getLastActivity gets the last activity time from various resources
func (ns *NamespaceScanner) getLastActivity(tenantInfo *TenantNamespaceInfo) time.Time {
	lastActivity := tenantInfo.CreationTimestamp

	// Check for more recent activity from pod restarts, etc.
	// This is a simplified implementation - in production, you'd check
	// events, logs, and other activity indicators

	return lastActivity
}

// resourceListToMap converts a resource list to a map
func resourceListToMap(resourceList corev1.ResourceList) map[string]string {
	result := make(map[string]string)
	for key, value := range resourceList {
		result[string(key)] = value.String()
	}
	return result
}
