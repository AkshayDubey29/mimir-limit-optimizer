package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// HealthScanner provides comprehensive health monitoring for Mimir infrastructure
type HealthScanner struct {
	client client.Client
	config *config.Config
	log    logr.Logger
}

// ResourceHealth represents the health status of a Kubernetes resource
type ResourceHealth struct {
	Name          string              `json:"name"`
	Namespace     string              `json:"namespace"`
	Kind          string              `json:"kind"`
	Status        string              `json:"status"`       // Healthy, Warning, Critical, Unknown
	HealthScore   float64             `json:"health_score"` // 0-100
	Replicas      ResourceReplicas    `json:"replicas,omitempty"`
	Conditions    []ResourceCondition `json:"conditions"`
	ResourceUsage ResourceUsage       `json:"resource_usage"`
	LastUpdated   time.Time           `json:"last_updated"`
	Issues        []HealthIssue       `json:"issues,omitempty"`
	Metrics       map[string]float64  `json:"metrics"`
	Labels        map[string]string   `json:"labels"`
	Age           time.Duration       `json:"age"`
}

// ResourceReplicas contains replica information for workloads
type ResourceReplicas struct {
	Desired     int32 `json:"desired"`
	Ready       int32 `json:"ready"`
	Available   int32 `json:"available"`
	Unavailable int32 `json:"unavailable"`
}

// ResourceCondition represents a condition of a resource
type ResourceCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// ResourceUsage contains resource utilization information
type ResourceUsage struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	CPULimit     string  `json:"cpu_limit,omitempty"`
	MemoryLimit  string  `json:"memory_limit,omitempty"`
	StorageUsage float64 `json:"storage_usage,omitempty"`
}

// HealthIssue represents a health issue detected
type HealthIssue struct {
	Severity    string `json:"severity"` // Critical, Warning, Info
	Category    string `json:"category"` // Performance, Availability, Configuration, Security
	Title       string `json:"title"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// MimirInfrastructureHealth represents the overall health of Mimir infrastructure
type MimirInfrastructureHealth struct {
	OverallHealth   string                `json:"overall_health"`
	OverallScore    float64               `json:"overall_score"`
	ComponentsCount ResourceTypeCount     `json:"components_count"`
	HealthSummary   HealthSummary         `json:"health_summary"`
	Resources       []ResourceHealth      `json:"resources"`
	LastScanTime    time.Time             `json:"last_scan_time"`
	ScanDuration    time.Duration         `json:"scan_duration"`
	Alerts          []InfrastructureAlert `json:"alerts"`
	Recommendations []AIRecommendation    `json:"recommendations"`
}

// ResourceTypeCount contains counts by resource type
type ResourceTypeCount struct {
	Deployments  int `json:"deployments"`
	StatefulSets int `json:"statefulsets"`
	DaemonSets   int `json:"daemonsets"`
	Services     int `json:"services"`
	ConfigMaps   int `json:"configmaps"`
	Secrets      int `json:"secrets"`
	Pods         int `json:"pods"`
	PVCs         int `json:"pvcs"`
}

// HealthSummary provides a summary of health status
type HealthSummary struct {
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
	Unknown  int `json:"unknown"`
}

// InfrastructureAlert represents an infrastructure-level alert
type InfrastructureAlert struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Component   string    `json:"component"`
	CreatedAt   time.Time `json:"created_at"`
}

// AIRecommendation represents an AI-generated recommendation
type AIRecommendation struct {
	ID          string    `json:"id"`
	Priority    string    `json:"priority"` // High, Medium, Low
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	Impact      string    `json:"impact"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewHealthScanner creates a new HealthScanner instance
func NewHealthScanner(client client.Client, cfg *config.Config, log logr.Logger) *HealthScanner {
	return &HealthScanner{
		client: client,
		config: cfg,
		log:    log.WithName("health-scanner"),
	}
}

// ScanMimirInfrastructure performs comprehensive health scan of Mimir infrastructure
func (h *HealthScanner) ScanMimirInfrastructure(ctx context.Context) (*MimirInfrastructureHealth, error) {
	startTime := time.Now()
	h.log.Info("starting comprehensive Mimir infrastructure health scan", "namespace", h.config.Mimir.Namespace)

	var allResources []ResourceHealth
	var componentCount ResourceTypeCount

	// Scan Deployments
	deployments, err := h.scanDeployments(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan deployments")
	} else {
		allResources = append(allResources, deployments...)
		componentCount.Deployments = len(deployments)
	}

	// Scan StatefulSets
	statefulSets, err := h.scanStatefulSets(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan statefulsets")
	} else {
		allResources = append(allResources, statefulSets...)
		componentCount.StatefulSets = len(statefulSets)
	}

	// Scan DaemonSets
	daemonSets, err := h.scanDaemonSets(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan daemonsets")
	} else {
		allResources = append(allResources, daemonSets...)
		componentCount.DaemonSets = len(daemonSets)
	}

	// Scan Services
	services, err := h.scanServices(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan services")
	} else {
		allResources = append(allResources, services...)
		componentCount.Services = len(services)
	}

	// Scan ConfigMaps
	configMaps, err := h.scanConfigMaps(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan configmaps")
	} else {
		allResources = append(allResources, configMaps...)
		componentCount.ConfigMaps = len(configMaps)
	}

	// Scan Secrets
	secrets, err := h.scanSecrets(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan secrets")
	} else {
		allResources = append(allResources, secrets...)
		componentCount.Secrets = len(secrets)
	}

	// Scan Pods
	pods, err := h.scanPods(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan pods")
	} else {
		allResources = append(allResources, pods...)
		componentCount.Pods = len(pods)
	}

	// Scan PVCs
	pvcs, err := h.scanPVCs(ctx)
	if err != nil {
		h.log.Error(err, "failed to scan pvcs")
	} else {
		allResources = append(allResources, pvcs...)
		componentCount.PVCs = len(pvcs)
	}

	// Calculate overall health
	healthSummary := h.calculateHealthSummary(allResources)
	overallHealth, overallScore := h.calculateOverallHealth(allResources)

	// Generate alerts and recommendations
	alerts := h.generateInfrastructureAlerts(allResources)
	recommendations := h.generateAIRecommendations(allResources)

	scanDuration := time.Since(startTime)

	result := &MimirInfrastructureHealth{
		OverallHealth:   overallHealth,
		OverallScore:    overallScore,
		ComponentsCount: componentCount,
		HealthSummary:   healthSummary,
		Resources:       allResources,
		LastScanTime:    startTime,
		ScanDuration:    scanDuration,
		Alerts:          alerts,
		Recommendations: recommendations,
	}

	h.log.Info("completed Mimir infrastructure health scan",
		"duration", scanDuration,
		"total_resources", len(allResources),
		"overall_health", overallHealth,
		"overall_score", overallScore)

	return result, nil
}

// scanDeployments scans all deployments in the Mimir namespace
func (h *HealthScanner) scanDeployments(ctx context.Context) ([]ResourceHealth, error) {
	deploymentList := &appsv1.DeploymentList{}
	err := h.client.List(ctx, deploymentList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var resources []ResourceHealth
	for _, dep := range deploymentList.Items {
		resource := h.analyzeDeploymentHealth(&dep)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeDeploymentHealth analyzes the health of a deployment
func (h *HealthScanner) analyzeDeploymentHealth(dep *appsv1.Deployment) ResourceHealth {
	var conditions []ResourceCondition
	for _, cond := range dep.Status.Conditions {
		conditions = append(conditions, ResourceCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Time,
		})
	}

	replicas := ResourceReplicas{
		Desired:     *dep.Spec.Replicas,
		Ready:       dep.Status.ReadyReplicas,
		Available:   dep.Status.AvailableReplicas,
		Unavailable: dep.Status.UnavailableReplicas,
	}

	// Calculate health score and status
	healthScore, status := h.calculateDeploymentHealth(dep)

	// Detect issues
	issues := h.detectDeploymentIssues(dep)

	return ResourceHealth{
		Name:          dep.Name,
		Namespace:     dep.Namespace,
		Kind:          "Deployment",
		Status:        status,
		HealthScore:   healthScore,
		Replicas:      replicas,
		Conditions:    conditions,
		ResourceUsage: h.getResourceUsage(dep.Spec.Template.Spec.Containers),
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getDeploymentMetrics(dep),
		Labels:        dep.Labels,
		Age:           time.Since(dep.CreationTimestamp.Time),
	}
}

// calculateDeploymentHealth calculates health score and status for a deployment
func (h *HealthScanner) calculateDeploymentHealth(dep *appsv1.Deployment) (float64, string) {
	if dep.Spec.Replicas == nil {
		return 0, "Unknown"
	}

	desired := *dep.Spec.Replicas
	ready := dep.Status.ReadyReplicas
	unavailable := dep.Status.UnavailableReplicas

	if desired == 0 {
		return 100, "Healthy"
	}

	// Calculate availability percentage
	availability := float64(ready) / float64(desired) * 100

	// Check for deployment conditions
	progressing := false
	available := false
	for _, cond := range dep.Status.Conditions {
		switch cond.Type {
		case appsv1.DeploymentProgressing:
			progressing = (cond.Status == corev1.ConditionTrue)
		case appsv1.DeploymentAvailable:
			available = (cond.Status == corev1.ConditionTrue)
		}
	}

	// Determine status and score
	if availability == 100 && available && progressing {
		return 100, "Healthy"
	} else if availability >= 80 && available {
		return availability, "Warning"
	} else if unavailable > 0 || !available {
		return availability, "Critical"
	}

	return availability, "Unknown"
}

// detectDeploymentIssues detects issues in a deployment
func (h *HealthScanner) detectDeploymentIssues(dep *appsv1.Deployment) []HealthIssue {
	var issues []HealthIssue

	// Check replica availability
	if dep.Spec.Replicas != nil && dep.Status.UnavailableReplicas > 0 {
		issues = append(issues, HealthIssue{
			Severity:    "Critical",
			Category:    "Availability",
			Title:       "Unavailable Replicas",
			Description: fmt.Sprintf("%d out of %d replicas are unavailable", dep.Status.UnavailableReplicas, *dep.Spec.Replicas),
			Suggestion:  "Check pod logs and events for deployment issues",
		})
	}

	// Check for resource limits
	for _, container := range dep.Spec.Template.Spec.Containers {
		if container.Resources.Limits == nil || container.Resources.Requests == nil {
			issues = append(issues, HealthIssue{
				Severity:    "Warning",
				Category:    "Configuration",
				Title:       "Missing Resource Limits",
				Description: fmt.Sprintf("Container %s has no resource limits or requests defined", container.Name),
				Suggestion:  "Define resource limits and requests for better resource management",
			})
		}
	}

	// Check for old deployment
	age := time.Since(dep.CreationTimestamp.Time)
	if age > 30*24*time.Hour { // 30 days
		issues = append(issues, HealthIssue{
			Severity:    "Info",
			Category:    "Configuration",
			Title:       "Long-running Deployment",
			Description: fmt.Sprintf("Deployment is %v old", age.Truncate(time.Hour)),
			Suggestion:  "Consider reviewing and updating deployment configuration",
		})
	}

	return issues
}

// getResourceUsage extracts resource usage information from containers
func (h *HealthScanner) getResourceUsage(containers []corev1.Container) ResourceUsage {
	// This is a placeholder - in production, you'd collect actual metrics
	return ResourceUsage{
		CPUUsage:    0.0, // Would be populated from metrics
		MemoryUsage: 0.0, // Would be populated from metrics
	}
}

// getDeploymentMetrics gets deployment-specific metrics
func (h *HealthScanner) getDeploymentMetrics(dep *appsv1.Deployment) map[string]float64 {
	metrics := make(map[string]float64)

	if dep.Spec.Replicas != nil {
		metrics["desired_replicas"] = float64(*dep.Spec.Replicas)
	}
	metrics["ready_replicas"] = float64(dep.Status.ReadyReplicas)
	metrics["available_replicas"] = float64(dep.Status.AvailableReplicas)
	metrics["unavailable_replicas"] = float64(dep.Status.UnavailableReplicas)

	return metrics
}

// Additional scan functions for other resource types would follow similar patterns...

// scanStatefulSets scans all statefulsets in the Mimir namespace
func (h *HealthScanner) scanStatefulSets(ctx context.Context) ([]ResourceHealth, error) {
	statefulSetList := &appsv1.StatefulSetList{}
	err := h.client.List(ctx, statefulSetList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	var resources []ResourceHealth
	for _, sts := range statefulSetList.Items {
		resource := h.analyzeStatefulSetHealth(&sts)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeStatefulSetHealth analyzes the health of a statefulset
func (h *HealthScanner) analyzeStatefulSetHealth(sts *appsv1.StatefulSet) ResourceHealth {
	var conditions []ResourceCondition
	for _, cond := range sts.Status.Conditions {
		conditions = append(conditions, ResourceCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Time,
		})
	}

	replicas := ResourceReplicas{
		Desired:   *sts.Spec.Replicas,
		Ready:     sts.Status.ReadyReplicas,
		Available: sts.Status.ReadyReplicas, // StatefulSets don't have AvailableReplicas
	}

	healthScore, status := h.calculateStatefulSetHealth(sts)
	issues := h.detectStatefulSetIssues(sts)

	return ResourceHealth{
		Name:          sts.Name,
		Namespace:     sts.Namespace,
		Kind:          "StatefulSet",
		Status:        status,
		HealthScore:   healthScore,
		Replicas:      replicas,
		Conditions:    conditions,
		ResourceUsage: h.getResourceUsage(sts.Spec.Template.Spec.Containers),
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getStatefulSetMetrics(sts),
		Labels:        sts.Labels,
		Age:           time.Since(sts.CreationTimestamp.Time),
	}
}

// calculateStatefulSetHealth calculates health score and status for a statefulset
func (h *HealthScanner) calculateStatefulSetHealth(sts *appsv1.StatefulSet) (float64, string) {
	if sts.Spec.Replicas == nil {
		return 0, "Unknown"
	}

	desired := *sts.Spec.Replicas
	ready := sts.Status.ReadyReplicas

	if desired == 0 {
		return 100, "Healthy"
	}

	availability := float64(ready) / float64(desired) * 100

	if availability == 100 {
		return 100, "Healthy"
	} else if availability >= 80 {
		return availability, "Warning"
	} else {
		return availability, "Critical"
	}
}

// detectStatefulSetIssues detects issues in a statefulset
func (h *HealthScanner) detectStatefulSetIssues(sts *appsv1.StatefulSet) []HealthIssue {
	var issues []HealthIssue

	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		issues = append(issues, HealthIssue{
			Severity:    "Critical",
			Category:    "Availability",
			Title:       "Not All Replicas Ready",
			Description: fmt.Sprintf("%d out of %d replicas are ready", sts.Status.ReadyReplicas, *sts.Spec.Replicas),
			Suggestion:  "Check pod logs and persistent volume status",
		})
	}

	return issues
}

// getStatefulSetMetrics gets statefulset-specific metrics
func (h *HealthScanner) getStatefulSetMetrics(sts *appsv1.StatefulSet) map[string]float64 {
	metrics := make(map[string]float64)

	if sts.Spec.Replicas != nil {
		metrics["desired_replicas"] = float64(*sts.Spec.Replicas)
	}
	metrics["ready_replicas"] = float64(sts.Status.ReadyReplicas)
	metrics["current_replicas"] = float64(sts.Status.CurrentReplicas)
	metrics["updated_replicas"] = float64(sts.Status.UpdatedReplicas)

	return metrics
}

// scanDaemonSets scans all daemonsets in the Mimir namespace
func (h *HealthScanner) scanDaemonSets(ctx context.Context) ([]ResourceHealth, error) {
	daemonSetList := &appsv1.DaemonSetList{}
	err := h.client.List(ctx, daemonSetList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	var resources []ResourceHealth
	for _, ds := range daemonSetList.Items {
		resource := h.analyzeDaemonSetHealth(&ds)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeDaemonSetHealth analyzes the health of a daemonset
func (h *HealthScanner) analyzeDaemonSetHealth(ds *appsv1.DaemonSet) ResourceHealth {
	var conditions []ResourceCondition
	for _, cond := range ds.Status.Conditions {
		conditions = append(conditions, ResourceCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Time,
		})
	}

	replicas := ResourceReplicas{
		Desired:     ds.Status.DesiredNumberScheduled,
		Ready:       ds.Status.NumberReady,
		Available:   ds.Status.NumberAvailable,
		Unavailable: ds.Status.NumberUnavailable,
	}

	healthScore, status := h.calculateDaemonSetHealth(ds)
	issues := h.detectDaemonSetIssues(ds)

	return ResourceHealth{
		Name:          ds.Name,
		Namespace:     ds.Namespace,
		Kind:          "DaemonSet",
		Status:        status,
		HealthScore:   healthScore,
		Replicas:      replicas,
		Conditions:    conditions,
		ResourceUsage: h.getResourceUsage(ds.Spec.Template.Spec.Containers),
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getDaemonSetMetrics(ds),
		Labels:        ds.Labels,
		Age:           time.Since(ds.CreationTimestamp.Time),
	}
}

// calculateDaemonSetHealth calculates health score and status for a daemonset
func (h *HealthScanner) calculateDaemonSetHealth(ds *appsv1.DaemonSet) (float64, string) {
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	if desired == 0 {
		return 100, "Healthy"
	}

	availability := float64(ready) / float64(desired) * 100

	if availability == 100 {
		return 100, "Healthy"
	} else if availability >= 80 {
		return availability, "Warning"
	} else {
		return availability, "Critical"
	}
}

// detectDaemonSetIssues detects issues in a daemonset
func (h *HealthScanner) detectDaemonSetIssues(ds *appsv1.DaemonSet) []HealthIssue {
	var issues []HealthIssue

	if ds.Status.NumberUnavailable > 0 {
		issues = append(issues, HealthIssue{
			Severity:    "Critical",
			Category:    "Availability",
			Title:       "Unavailable DaemonSet Pods",
			Description: fmt.Sprintf("%d pods are unavailable", ds.Status.NumberUnavailable),
			Suggestion:  "Check node status and pod scheduling constraints",
		})
	}

	return issues
}

// getDaemonSetMetrics gets daemonset-specific metrics
func (h *HealthScanner) getDaemonSetMetrics(ds *appsv1.DaemonSet) map[string]float64 {
	metrics := make(map[string]float64)

	metrics["desired_scheduled"] = float64(ds.Status.DesiredNumberScheduled)
	metrics["current_scheduled"] = float64(ds.Status.CurrentNumberScheduled)
	metrics["number_ready"] = float64(ds.Status.NumberReady)
	metrics["number_available"] = float64(ds.Status.NumberAvailable)
	metrics["number_unavailable"] = float64(ds.Status.NumberUnavailable)

	return metrics
}

// scanServices scans all services in the Mimir namespace
func (h *HealthScanner) scanServices(ctx context.Context) ([]ResourceHealth, error) {
	serviceList := &corev1.ServiceList{}
	err := h.client.List(ctx, serviceList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var resources []ResourceHealth
	for _, svc := range serviceList.Items {
		resource := h.analyzeServiceHealth(&svc)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeServiceHealth analyzes the health of a service
func (h *HealthScanner) analyzeServiceHealth(svc *corev1.Service) ResourceHealth {
	healthScore, status := h.calculateServiceHealth(svc)
	issues := h.detectServiceIssues(svc)

	return ResourceHealth{
		Name:          svc.Name,
		Namespace:     svc.Namespace,
		Kind:          "Service",
		Status:        status,
		HealthScore:   healthScore,
		Conditions:    []ResourceCondition{}, // Services don't have conditions
		ResourceUsage: ResourceUsage{},       // Services don't have resource usage
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getServiceMetrics(svc),
		Labels:        svc.Labels,
		Age:           time.Since(svc.CreationTimestamp.Time),
	}
}

// calculateServiceHealth calculates health score and status for a service
func (h *HealthScanner) calculateServiceHealth(svc *corev1.Service) (float64, string) {
	// Basic health check for services
	if len(svc.Spec.Ports) == 0 {
		return 50, "Warning"
	}

	// Check if service has endpoints
	// This is a simplified check - in production, you'd check actual endpoints
	return 100, "Healthy"
}

// detectServiceIssues detects issues in a service
func (h *HealthScanner) detectServiceIssues(svc *corev1.Service) []HealthIssue {
	var issues []HealthIssue

	if len(svc.Spec.Ports) == 0 {
		issues = append(issues, HealthIssue{
			Severity:    "Warning",
			Category:    "Configuration",
			Title:       "No Ports Defined",
			Description: "Service has no ports defined",
			Suggestion:  "Define appropriate ports for the service",
		})
	}

	return issues
}

// getServiceMetrics gets service-specific metrics
func (h *HealthScanner) getServiceMetrics(svc *corev1.Service) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["port_count"] = float64(len(svc.Spec.Ports))
	return metrics
}

// scanConfigMaps scans all configmaps in the Mimir namespace
func (h *HealthScanner) scanConfigMaps(ctx context.Context) ([]ResourceHealth, error) {
	configMapList := &corev1.ConfigMapList{}
	err := h.client.List(ctx, configMapList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	var resources []ResourceHealth
	for _, cm := range configMapList.Items {
		resource := h.analyzeConfigMapHealth(&cm)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeConfigMapHealth analyzes the health of a configmap
func (h *HealthScanner) analyzeConfigMapHealth(cm *corev1.ConfigMap) ResourceHealth {
	healthScore, status := h.calculateConfigMapHealth(cm)
	issues := h.detectConfigMapIssues(cm)

	return ResourceHealth{
		Name:          cm.Name,
		Namespace:     cm.Namespace,
		Kind:          "ConfigMap",
		Status:        status,
		HealthScore:   healthScore,
		Conditions:    []ResourceCondition{},
		ResourceUsage: ResourceUsage{},
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getConfigMapMetrics(cm),
		Labels:        cm.Labels,
		Age:           time.Since(cm.CreationTimestamp.Time),
	}
}

// calculateConfigMapHealth calculates health score and status for a configmap
func (h *HealthScanner) calculateConfigMapHealth(cm *corev1.ConfigMap) (float64, string) {
	// ConfigMaps are generally healthy if they exist
	if len(cm.Data) == 0 && len(cm.BinaryData) == 0 {
		return 80, "Warning"
	}
	return 100, "Healthy"
}

// detectConfigMapIssues detects issues in a configmap
func (h *HealthScanner) detectConfigMapIssues(cm *corev1.ConfigMap) []HealthIssue {
	var issues []HealthIssue

	if len(cm.Data) == 0 && len(cm.BinaryData) == 0 {
		issues = append(issues, HealthIssue{
			Severity:    "Warning",
			Category:    "Configuration",
			Title:       "Empty ConfigMap",
			Description: "ConfigMap contains no data",
			Suggestion:  "Verify if this ConfigMap is still needed",
		})
	}

	return issues
}

// getConfigMapMetrics gets configmap-specific metrics
func (h *HealthScanner) getConfigMapMetrics(cm *corev1.ConfigMap) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["data_keys"] = float64(len(cm.Data))
	metrics["binary_data_keys"] = float64(len(cm.BinaryData))
	return metrics
}

// scanSecrets scans all secrets in the Mimir namespace
func (h *HealthScanner) scanSecrets(ctx context.Context) ([]ResourceHealth, error) {
	secretList := &corev1.SecretList{}
	err := h.client.List(ctx, secretList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var resources []ResourceHealth
	for _, secret := range secretList.Items {
		resource := h.analyzeSecretHealth(&secret)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzeSecretHealth analyzes the health of a secret
func (h *HealthScanner) analyzeSecretHealth(secret *corev1.Secret) ResourceHealth {
	healthScore, status := h.calculateSecretHealth(secret)
	issues := h.detectSecretIssues(secret)

	return ResourceHealth{
		Name:          secret.Name,
		Namespace:     secret.Namespace,
		Kind:          "Secret",
		Status:        status,
		HealthScore:   healthScore,
		Conditions:    []ResourceCondition{},
		ResourceUsage: ResourceUsage{},
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getSecretMetrics(secret),
		Labels:        secret.Labels,
		Age:           time.Since(secret.CreationTimestamp.Time),
	}
}

// calculateSecretHealth calculates health score and status for a secret
func (h *HealthScanner) calculateSecretHealth(secret *corev1.Secret) (float64, string) {
	// Secrets are generally healthy if they exist
	if len(secret.Data) == 0 {
		return 80, "Warning"
	}
	return 100, "Healthy"
}

// detectSecretIssues detects issues in a secret
func (h *HealthScanner) detectSecretIssues(secret *corev1.Secret) []HealthIssue {
	var issues []HealthIssue

	if len(secret.Data) == 0 {
		issues = append(issues, HealthIssue{
			Severity:    "Warning",
			Category:    "Security",
			Title:       "Empty Secret",
			Description: "Secret contains no data",
			Suggestion:  "Verify if this Secret is still needed",
		})
	}

	return issues
}

// getSecretMetrics gets secret-specific metrics
func (h *HealthScanner) getSecretMetrics(secret *corev1.Secret) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["data_keys"] = float64(len(secret.Data))
	return metrics
}

// scanPods scans all pods in the Mimir namespace
func (h *HealthScanner) scanPods(ctx context.Context) ([]ResourceHealth, error) {
	podList := &corev1.PodList{}
	err := h.client.List(ctx, podList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var resources []ResourceHealth
	for _, pod := range podList.Items {
		resource := h.analyzePodHealth(&pod)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzePodHealth analyzes the health of a pod
func (h *HealthScanner) analyzePodHealth(pod *corev1.Pod) ResourceHealth {
	var conditions []ResourceCondition
	for _, cond := range pod.Status.Conditions {
		conditions = append(conditions, ResourceCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Time,
		})
	}

	healthScore, status := h.calculatePodHealth(pod)
	issues := h.detectPodIssues(pod)

	return ResourceHealth{
		Name:          pod.Name,
		Namespace:     pod.Namespace,
		Kind:          "Pod",
		Status:        status,
		HealthScore:   healthScore,
		Conditions:    conditions,
		ResourceUsage: h.getResourceUsage(pod.Spec.Containers),
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getPodMetrics(pod),
		Labels:        pod.Labels,
		Age:           time.Since(pod.CreationTimestamp.Time),
	}
}

// calculatePodHealth calculates health score and status for a pod
func (h *HealthScanner) calculatePodHealth(pod *corev1.Pod) (float64, string) {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		// Check if all containers are ready
		ready := true
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				ready = false
				break
			}
		}
		if ready {
			return 100, "Healthy"
		}
		return 75, "Warning"
	case corev1.PodPending:
		return 50, "Warning"
	case corev1.PodSucceeded:
		return 100, "Healthy"
	case corev1.PodFailed:
		return 0, "Critical"
	default:
		return 0, "Unknown"
	}
}

// detectPodIssues detects issues in a pod
func (h *HealthScanner) detectPodIssues(pod *corev1.Pod) []HealthIssue {
	var issues []HealthIssue

	if pod.Status.Phase == corev1.PodFailed {
		issues = append(issues, HealthIssue{
			Severity:    "Critical",
			Category:    "Availability",
			Title:       "Pod Failed",
			Description: fmt.Sprintf("Pod is in failed state: %s", pod.Status.Reason),
			Suggestion:  "Check pod logs and events for failure reasons",
		})
	}

	// Check container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.RestartCount > 5 {
			issues = append(issues, HealthIssue{
				Severity:    "Warning",
				Category:    "Performance",
				Title:       "High Restart Count",
				Description: fmt.Sprintf("Container %s has restarted %d times", containerStatus.Name, containerStatus.RestartCount),
				Suggestion:  "Investigate container crashes and resource limits",
			})
		}
	}

	return issues
}

// getPodMetrics gets pod-specific metrics
func (h *HealthScanner) getPodMetrics(pod *corev1.Pod) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["container_count"] = float64(len(pod.Spec.Containers))

	totalRestarts := int32(0)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += containerStatus.RestartCount
	}
	metrics["total_restarts"] = float64(totalRestarts)

	return metrics
}

// scanPVCs scans all persistent volume claims in the Mimir namespace
func (h *HealthScanner) scanPVCs(ctx context.Context) ([]ResourceHealth, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	err := h.client.List(ctx, pvcList, client.InNamespace(h.config.Mimir.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list pvcs: %w", err)
	}

	var resources []ResourceHealth
	for _, pvc := range pvcList.Items {
		resource := h.analyzePVCHealth(&pvc)
		resources = append(resources, resource)
	}

	return resources, nil
}

// analyzePVCHealth analyzes the health of a PVC
func (h *HealthScanner) analyzePVCHealth(pvc *corev1.PersistentVolumeClaim) ResourceHealth {
	healthScore, status := h.calculatePVCHealth(pvc)
	issues := h.detectPVCIssues(pvc)

	return ResourceHealth{
		Name:          pvc.Name,
		Namespace:     pvc.Namespace,
		Kind:          "PersistentVolumeClaim",
		Status:        status,
		HealthScore:   healthScore,
		Conditions:    []ResourceCondition{},
		ResourceUsage: ResourceUsage{},
		LastUpdated:   time.Now(),
		Issues:        issues,
		Metrics:       h.getPVCMetrics(pvc),
		Labels:        pvc.Labels,
		Age:           time.Since(pvc.CreationTimestamp.Time),
	}
}

// calculatePVCHealth calculates health score and status for a PVC
func (h *HealthScanner) calculatePVCHealth(pvc *corev1.PersistentVolumeClaim) (float64, string) {
	switch pvc.Status.Phase {
	case corev1.ClaimBound:
		return 100, "Healthy"
	case corev1.ClaimPending:
		return 50, "Warning"
	case corev1.ClaimLost:
		return 0, "Critical"
	default:
		return 0, "Unknown"
	}
}

// detectPVCIssues detects issues in a PVC
func (h *HealthScanner) detectPVCIssues(pvc *corev1.PersistentVolumeClaim) []HealthIssue {
	var issues []HealthIssue

	if pvc.Status.Phase == corev1.ClaimPending {
		issues = append(issues, HealthIssue{
			Severity:    "Warning",
			Category:    "Performance",
			Title:       "PVC Pending",
			Description: "PersistentVolumeClaim is in pending state",
			Suggestion:  "Check storage class and persistent volume availability",
		})
	}

	if pvc.Status.Phase == corev1.ClaimLost {
		issues = append(issues, HealthIssue{
			Severity:    "Critical",
			Category:    "Availability",
			Title:       "PVC Lost",
			Description: "PersistentVolumeClaim is in lost state",
			Suggestion:  "Investigate persistent volume issues and data recovery",
		})
	}

	return issues
}

// getPVCMetrics gets PVC-specific metrics
func (h *HealthScanner) getPVCMetrics(pvc *corev1.PersistentVolumeClaim) map[string]float64 {
	metrics := make(map[string]float64)

	if capacity, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
		metrics["capacity_bytes"] = float64(capacity.Value())
	}

	return metrics
}

// calculateHealthSummary calculates health summary from all resources
func (h *HealthScanner) calculateHealthSummary(resources []ResourceHealth) HealthSummary {
	summary := HealthSummary{}

	for _, resource := range resources {
		switch resource.Status {
		case "Healthy":
			summary.Healthy++
		case "Warning":
			summary.Warning++
		case "Critical":
			summary.Critical++
		default:
			summary.Unknown++
		}
	}

	return summary
}

// calculateOverallHealth calculates overall health status and score
func (h *HealthScanner) calculateOverallHealth(resources []ResourceHealth) (string, float64) {
	if len(resources) == 0 {
		return "Unknown", 0
	}

	totalScore := 0.0
	criticalCount := 0
	warningCount := 0

	for _, resource := range resources {
		totalScore += resource.HealthScore
		if resource.Status == "Critical" {
			criticalCount++
		} else if resource.Status == "Warning" {
			warningCount++
		}
	}

	averageScore := totalScore / float64(len(resources))

	// Determine overall status
	if criticalCount > 0 {
		return "Critical", averageScore
	} else if warningCount > 0 {
		return "Warning", averageScore
	} else if averageScore >= 95 {
		return "Healthy", averageScore
	} else if averageScore >= 80 {
		return "Warning", averageScore
	} else {
		return "Critical", averageScore
	}
}

// generateInfrastructureAlerts generates infrastructure-level alerts
func (h *HealthScanner) generateInfrastructureAlerts(resources []ResourceHealth) []InfrastructureAlert {
	var alerts []InfrastructureAlert

	criticalCount := 0
	for _, resource := range resources {
		if resource.Status == "Critical" {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		alerts = append(alerts, InfrastructureAlert{
			ID:          fmt.Sprintf("infra-critical-%d", time.Now().Unix()),
			Severity:    "Critical",
			Title:       "Critical Infrastructure Issues",
			Description: fmt.Sprintf("%d resources are in critical state", criticalCount),
			Component:   "Infrastructure",
			CreatedAt:   time.Now(),
		})
	}

	return alerts
}

// generateAIRecommendations generates AI-powered recommendations
func (h *HealthScanner) generateAIRecommendations(resources []ResourceHealth) []AIRecommendation {
	var recommendations []AIRecommendation

	// Example AI-powered recommendations based on patterns
	missingLimitsCount := 0
	highRestartCount := 0

	for _, resource := range resources {
		for _, issue := range resource.Issues {
			if issue.Title == "Missing Resource Limits" {
				missingLimitsCount++
			}
			if issue.Title == "High Restart Count" {
				highRestartCount++
			}
		}
	}

	if missingLimitsCount > 0 {
		recommendations = append(recommendations, AIRecommendation{
			ID:          fmt.Sprintf("ai-limits-%d", time.Now().Unix()),
			Priority:    "High",
			Category:    "Resource Management",
			Title:       "Implement Resource Limits",
			Description: fmt.Sprintf("%d resources are missing resource limits", missingLimitsCount),
			Action:      "Define CPU and memory limits for all containers",
			Impact:      "Improved resource utilization and cluster stability",
			CreatedAt:   time.Now(),
		})
	}

	if highRestartCount > 0 {
		recommendations = append(recommendations, AIRecommendation{
			ID:          fmt.Sprintf("ai-restarts-%d", time.Now().Unix()),
			Priority:    "Medium",
			Category:    "Stability",
			Title:       "Investigate Container Restarts",
			Description: fmt.Sprintf("%d resources have high restart counts", highRestartCount),
			Action:      "Analyze container logs and adjust resource limits",
			Impact:      "Reduced downtime and improved application stability",
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}
