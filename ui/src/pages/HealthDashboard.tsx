import React, { useState, useEffect, useCallback } from 'react';
import { useApi } from '../context/ApiContext';
import { useTheme } from '../context/ThemeContext';

// Types for health monitoring data
interface ResourceHealth {
  name: string;
  namespace: string;
  kind: string;
  status: 'Healthy' | 'Warning' | 'Critical' | 'Unknown';
  health_score: number;
  replicas?: {
    desired: number;
    ready: number;
    available: number;
    unavailable: number;
  };
  conditions: Array<{
    type: string;
    status: string;
    reason?: string;
    message?: string;
    last_transition_time: string;
  }>;
  resource_usage: {
    cpu_usage: number;
    memory_usage: number;
    cpu_limit?: string;
    memory_limit?: string;
    storage_usage?: number;
  };
  last_updated: string;
  issues: Array<{
    severity: 'Critical' | 'Warning' | 'Info';
    category: string;
    title: string;
    description: string;
    suggestion?: string;
  }>;
  metrics: Record<string, number>;
  labels: Record<string, string>;
  age: string;
}

interface HealthSummary {
  healthy: number;
  warning: number;
  critical: number;
  unknown: number;
}

interface ComponentsCount {
  deployments: number;
  statefulsets: number;
  daemonsets: number;
  services: number;
  configmaps: number;
  secrets: number;
  pods: number;
  pvcs: number;
}

interface InfrastructureAlert {
  id: string;
  severity: 'Critical' | 'Warning' | 'Info';
  title: string;
  description: string;
  component: string;
  created_at: string;
}

interface AIRecommendation {
  id: string;
  priority: 'High' | 'Medium' | 'Low';
  category: string;
  title: string;
  description: string;
  action: string;
  impact: string;
  created_at: string;
}

interface IngestionCapacity {
  current_ingestion_rate: number; // samples/sec
  max_ingestion_capacity: number; // samples/sec
  capacity_utilization: number; // percentage (0-100)
  available_capacity: number; // samples/sec
  sustainable_hours: number; // hours we can sustain current rate
  burst_capacity: number; // samples/sec for burst scenarios
  ingestion_efficiency: number; // percentage (0-100)
  data_source?: string; // "real_metrics" or "synthetic"
  tenant_count?: number; // number of tenants
  active_series?: number; // active series count
  calculations?: {
    current_ingestion_rate?: string;
    max_ingestion_capacity?: string;
    capacity_utilization?: string;
    available_capacity?: string;
    sustainable_hours?: string;
    burst_capacity?: string;
    ingestion_efficiency?: string;
  };
  metadata?: {
    calculation_timestamp?: string;
    metrics_source?: string;
    estimation_method?: string;
  };
}

interface HealthMetrics {
  overall_health: string;
  overall_score: number;
  health_summary: HealthSummary;
  components_count: ComponentsCount;
  ingestion_capacity: IngestionCapacity;
  last_scan_time: string;
  scan_duration_ms: number;
  alert_count: number;
  recommendation_count: number;
  resource_breakdown: any;
  trend_data: any;
}

// Tooltip Component
interface TooltipProps {
  content: string;
  children: React.ReactNode;
  className?: string;
}

const Tooltip: React.FC<TooltipProps> = ({ content, children, className = "" }) => {
  const [visible, setVisible] = useState(false);
  const { darkMode } = useTheme();

  return (
    <div className={`relative inline-block ${className}`}>
      <div
        onMouseEnter={() => setVisible(true)}
        onMouseLeave={() => setVisible(false)}
        className="cursor-help"
      >
        {children}
      </div>
      {visible && (
        <div className={`absolute z-50 px-3 py-2 text-sm rounded-lg shadow-lg max-w-xs whitespace-normal break-words
          ${darkMode ? 'bg-gray-700 text-white border border-gray-600' : 'bg-gray-900 text-white'} 
          bottom-full left-1/2 transform -translate-x-1/2 mb-2`}
        >
          {content}
          <div className={`absolute top-full left-1/2 transform -translate-x-1/2 w-2 h-2 rotate-45 
            ${darkMode ? 'bg-gray-700' : 'bg-gray-900'}`}></div>
        </div>
      )}
    </div>
  );
};

const HealthDashboard: React.FC = () => {
  const { apiRequest, loading, error } = useApi();
  const { darkMode } = useTheme();
  
  const [healthMetrics, setHealthMetrics] = useState<HealthMetrics | null>(null);
  const [resources, setResources] = useState<ResourceHealth[]>([]);
  const [alerts, setAlerts] = useState<InfrastructureAlert[]>([]);
  const [recommendations, setRecommendations] = useState<AIRecommendation[]>([]);
  const [selectedResource, setSelectedResource] = useState<ResourceHealth | null>(null);
  const [filterKind, setFilterKind] = useState<string>('');
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Safe accessors with default values
  const safeAlerts = alerts || [];
  const safeRecommendations = recommendations || [];
  const safeResources = resources || [];

  // Fetch health data
  const fetchHealthData = useCallback(async () => {
    try {
      // Try to fetch health data, but provide fallbacks if endpoints don't exist
      let healthData: HealthMetrics | null = null;
      let resourcesData: ResourceHealth[] = [];
      let alertsData: InfrastructureAlert[] = [];
      let recommendationsData: AIRecommendation[] = [];

      try {
        // Try health endpoints first
        const [metricsData, alertsResponse, recommendationsResponse, resourcesResponse] = await Promise.all([
          apiRequest('/health/metrics').catch(() => null),
          apiRequest('/health/alerts').catch(() => []),
          apiRequest('/health/recommendations').catch(() => []),
          apiRequest('/health/resources').catch(() => ({ resources: [] }))
        ]);

        healthData = metricsData;
        alertsData = alertsResponse || [];
        recommendationsData = recommendationsResponse || [];
        resourcesData = resourcesResponse?.resources || [];
      } catch (err) {
        console.log('Health endpoints not available, using fallback data');
      }

      // If health endpoints don't exist, create fallback data using available endpoints
      if (!healthData) {
        try {
          const [tenantsResponse] = await Promise.all([
            apiRequest('/tenants').catch(() => ({ tenants: [] })),
            apiRequest('/status').catch(() => ({ mode: 'unknown' }))
          ]);

          const tenants = tenantsResponse?.tenants || [];
          // const status = statusResponse || {}; // Unused variable

          // Create synthetic health data
                     healthData = {
             overall_health: 'Healthy',
             overall_score: 85.5,
             health_summary: {
               healthy: 8,
               warning: 2,
               critical: 0,
               unknown: 1
             },
             components_count: {
               deployments: 6,
               statefulsets: 3,
               daemonsets: 1,
               services: 8,
               configmaps: 4,
               secrets: 2,
               pods: 15,
               pvcs: 3
             },
             ingestion_capacity: {
               current_ingestion_rate: 125000, // 125K samples/sec
               max_ingestion_capacity: 200000, // 200K samples/sec
               capacity_utilization: 62.5, // 62.5% utilization
               available_capacity: 75000, // 75K samples/sec available
               sustainable_hours: 168, // 1 week sustainable at current rate
               burst_capacity: 350000, // 350K samples/sec burst capacity
               ingestion_efficiency: 92.3 // 92.3% efficiency
             },
             last_scan_time: new Date().toISOString(),
             scan_duration_ms: 1250,
             alert_count: (alertsData || []).length,
             recommendation_count: (recommendationsData || []).length,
             resource_breakdown: {},
             trend_data: {}
           };

                     // Create some sample alerts if none exist
           if ((alertsData || []).length === 0) {
            alertsData = [
              {
                id: 'alert-1',
                severity: 'Warning' as const,
                title: 'Health endpoints not implemented',
                description: 'Some health monitoring endpoints are not yet available. Using fallback data.',
                component: 'health-monitor',
                created_at: new Date().toISOString()
              }
            ];
          }

                     // Create sample recommendations
           if ((recommendationsData || []).length === 0) {
            recommendationsData = [
              {
                id: 'rec-1',
                priority: 'Medium' as const,
                category: 'monitoring',
                title: 'Implement Health Endpoints',
                description: 'Health monitoring endpoints should be implemented for better visibility.',
                action: 'Add /api/health/* endpoints to the backend',
                impact: 'Improved monitoring and alerting capabilities',
                created_at: new Date().toISOString()
              },
              {
                id: 'rec-2',
                priority: 'Low' as const,
                category: 'optimization',
                title: 'Tenant Configuration Review',
                description: `Found ${tenants.length} tenants. Review configurations for optimization opportunities.`,
                action: 'Review tenant limits and usage patterns',
                impact: 'Better resource utilization and cost optimization',
                created_at: new Date().toISOString()
              }
            ];
          }

        } catch (fallbackError) {
          console.error('Failed to create fallback data:', fallbackError);
        }
      }

      setHealthMetrics(healthData);
      setAlerts(alertsData);
      setRecommendations(recommendationsData);
      setResources(resourcesData);
    } catch (err) {
      console.error('Error fetching health data:', err);
    }
  }, [apiRequest]);

  // Auto-refresh effect
  useEffect(() => {
    fetchHealthData();
    
    if (autoRefresh) {
      const interval = setInterval(fetchHealthData, 30000); // Refresh every 30 seconds
      return () => clearInterval(interval);
    }
  }, [autoRefresh, fetchHealthData]);

  // Get status color
  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'Healthy': return 'text-green-600 bg-green-100';
      case 'Warning': return 'text-yellow-600 bg-yellow-100';
      case 'Critical': return 'text-red-600 bg-red-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  // Get severity color
  const getSeverityColor = (severity: string): string => {
    switch (severity) {
      case 'Critical': return 'text-red-600 bg-red-100';
      case 'Warning': return 'text-yellow-600 bg-yellow-100';
      case 'Info': return 'text-blue-600 bg-blue-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  // Get priority color
  const getPriorityColor = (priority: string): string => {
    switch (priority) {
      case 'High': return 'text-red-600 bg-red-100';
      case 'Medium': return 'text-yellow-600 bg-yellow-100';
      case 'Low': return 'text-green-600 bg-green-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  // Format duration
  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  // Format ingestion rate
  const formatIngestionRate = (samplesPerSec: number): string => {
    if (samplesPerSec >= 1000000) return `${(samplesPerSec / 1000000).toFixed(1)}M/s`;
    if (samplesPerSec >= 1000) return `${(samplesPerSec / 1000).toFixed(1)}K/s`;
    return `${samplesPerSec}/s`;
  };

  // Get capacity utilization color
  const getCapacityColor = (utilization: number): string => {
    if (utilization >= 90) return 'text-red-600 bg-red-100';
    if (utilization >= 75) return 'text-yellow-600 bg-yellow-100';
    return 'text-green-600 bg-green-100';
  };

  // Filter resources
  const filteredResources = safeResources.filter(resource => {
    if (filterKind && resource.kind !== filterKind) return false;
    if (filterStatus && resource.status !== filterStatus) return false;
    return true;
  });

  if (loading && !healthMetrics) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p>{error}</p>
        <button 
          onClick={fetchHealthData}
          className="mt-2 bg-red-100 hover:bg-red-200 px-3 py-1 rounded text-sm"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className={`p-6 max-w-7xl mx-auto ${darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'}`}>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className={`text-3xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>Mimir Infrastructure Health</h1>
            <p className={`mt-1 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>
              AI-powered monitoring and optimization for Grafana Mimir
            </p>
          </div>
          <div className="flex items-center space-x-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="mr-2"
              />
              Auto-refresh
            </label>
            <button
              onClick={fetchHealthData}
              disabled={loading}
              className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg disabled:opacity-50"
            >
              {loading ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>
        </div>
      </div>

      {healthMetrics && (
        <>
          {/* Overall Health Status */}
          <div className={`rounded-xl shadow-sm border p-6 mb-8 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
            <h2 className={`text-xl font-semibold mb-4 ${darkMode ? 'text-white' : 'text-gray-900'}`}>Overall Health Status</h2>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div className="text-center">
                <div className={`inline-flex items-center px-4 py-2 rounded-full text-lg font-semibold ${getStatusColor(healthMetrics.overall_health)}`}>
                  {healthMetrics.overall_health}
                </div>
                <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Overall Status</p>
              </div>
              <div className="text-center">
                <div className={`text-3xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
                  {healthMetrics.overall_score.toFixed(1)}%
                </div>
                <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Health Score</p>
              </div>
              <div className="text-center">
                <div className={`text-3xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
                  {healthMetrics.alert_count}
                </div>
                <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Active Alerts</p>
              </div>
              <div className="text-center">
                <div className={`text-3xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
                  {formatDuration(healthMetrics.scan_duration_ms)}
                </div>
                <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Last Scan Duration</p>
              </div>
            </div>
          </div>

          {/* Ingestion Capacity */}
          {healthMetrics.ingestion_capacity && (
            <div className={`rounded-xl shadow-sm border p-6 mb-8 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
              <div className="flex items-center justify-between mb-4">
                <h2 className={`text-xl font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>📊 Ingestion Capacity & Performance</h2>
                <div className="flex items-center space-x-2">
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    healthMetrics.ingestion_capacity.data_source === 'real_metrics' 
                      ? 'bg-green-100 text-green-700 border border-green-200' 
                      : 'bg-yellow-100 text-yellow-700 border border-yellow-200'
                  }`}>
                    {healthMetrics.ingestion_capacity.data_source === 'real_metrics' ? '🔗 Real Data' : '⚡ Synthetic Data'}
                  </span>
                  {healthMetrics.ingestion_capacity.tenant_count && (
                    <span className={`px-2 py-1 rounded-full text-xs ${darkMode ? 'bg-gray-700 text-gray-300' : 'bg-gray-100 text-gray-600'}`}>
                      {healthMetrics.ingestion_capacity.tenant_count} tenants
                    </span>
                  )}
                </div>
              </div>
              
              {/* Capacity Overview */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
                <Tooltip content={healthMetrics.ingestion_capacity.calculations?.current_ingestion_rate || "Current ingestion rate across all tenants"}>
                  <div className="text-center p-4 rounded-lg border border-dashed border-blue-300 hover:border-blue-500 transition-colors">
                    <div className={`text-3xl font-bold ${darkMode ? 'text-blue-400' : 'text-blue-600'}`}>
                      {formatIngestionRate(healthMetrics.ingestion_capacity.current_ingestion_rate)}
                    </div>
                    <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Current Rate</p>
                    <span className="text-xs opacity-50">ⓘ Hover for details</span>
                  </div>
                </Tooltip>

                <Tooltip content={healthMetrics.ingestion_capacity.calculations?.max_ingestion_capacity || "Maximum theoretical ingestion capacity of the cluster"}>
                  <div className="text-center p-4 rounded-lg border border-dashed border-green-300 hover:border-green-500 transition-colors">
                    <div className={`text-3xl font-bold ${darkMode ? 'text-green-400' : 'text-green-600'}`}>
                      {formatIngestionRate(healthMetrics.ingestion_capacity.max_ingestion_capacity)}
                    </div>
                    <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Max Capacity</p>
                    <span className="text-xs opacity-50">ⓘ Hover for details</span>
                  </div>
                </Tooltip>

                <Tooltip content={healthMetrics.ingestion_capacity.calculations?.available_capacity || "Remaining ingestion capacity available"}>
                  <div className="text-center p-4 rounded-lg border border-dashed border-yellow-300 hover:border-yellow-500 transition-colors">
                    <div className={`text-3xl font-bold ${darkMode ? 'text-yellow-400' : 'text-yellow-600'}`}>
                      {formatIngestionRate(healthMetrics.ingestion_capacity.available_capacity)}
                    </div>
                    <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Available</p>
                    <span className="text-xs opacity-50">ⓘ Hover for details</span>
                  </div>
                </Tooltip>

                <Tooltip content={healthMetrics.ingestion_capacity.calculations?.burst_capacity || "Maximum temporary burst capacity for traffic spikes"}>
                  <div className="text-center p-4 rounded-lg border border-dashed border-purple-300 hover:border-purple-500 transition-colors">
                    <div className={`text-3xl font-bold ${darkMode ? 'text-purple-400' : 'text-purple-600'}`}>
                      {formatIngestionRate(healthMetrics.ingestion_capacity.burst_capacity)}
                    </div>
                    <p className={`mt-2 ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>Burst Capacity</p>
                    <span className="text-xs opacity-50">ⓘ Hover for details</span>
                  </div>
                </Tooltip>
              </div>

            {/* Performance Indicators */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Tooltip content={healthMetrics.ingestion_capacity.calculations?.capacity_utilization || "Percentage of maximum capacity currently being used"}>
                <div className={`p-4 rounded-lg border border-dashed border-gray-300 hover:border-gray-500 transition-colors ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className={`font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>Capacity Utilization</h3>
                    <span className={`px-2 py-1 rounded-full text-sm font-medium ${getCapacityColor(healthMetrics.ingestion_capacity.capacity_utilization)}`}>
                      {healthMetrics.ingestion_capacity.capacity_utilization.toFixed(1)}%
                    </span>
                  </div>
                  <div className={`w-full rounded-full h-3 ${darkMode ? 'bg-gray-600' : 'bg-gray-200'}`}>
                    <div 
                      className={`h-3 rounded-full transition-all duration-300 ${
                        healthMetrics.ingestion_capacity.capacity_utilization >= 90 ? 'bg-red-500' :
                        healthMetrics.ingestion_capacity.capacity_utilization >= 75 ? 'bg-yellow-500' : 'bg-green-500'
                      }`}
                      style={{ width: `${healthMetrics.ingestion_capacity.capacity_utilization}%` }}
                    ></div>
                  </div>
                  <span className="text-xs opacity-50 mt-1 block">ⓘ Hover for calculation</span>
                </div>
              </Tooltip>

              <Tooltip content={healthMetrics.ingestion_capacity.calculations?.ingestion_efficiency || "Percentage of successfully processed samples vs attempted"}>
                <div className={`p-4 rounded-lg border border-dashed border-gray-300 hover:border-gray-500 transition-colors ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className={`font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>Ingestion Efficiency</h3>
                    <span className={`px-2 py-1 rounded-full text-sm font-medium ${
                      healthMetrics.ingestion_capacity.ingestion_efficiency >= 95 ? 'text-green-600 bg-green-100' :
                      healthMetrics.ingestion_capacity.ingestion_efficiency >= 85 ? 'text-yellow-600 bg-yellow-100' :
                      'text-red-600 bg-red-100'
                    }`}>
                      {healthMetrics.ingestion_capacity.ingestion_efficiency.toFixed(1)}%
                    </span>
                  </div>
                  <div className={`w-full rounded-full h-3 ${darkMode ? 'bg-gray-600' : 'bg-gray-200'}`}>
                    <div 
                      className={`h-3 rounded-full transition-all duration-300 ${
                        healthMetrics.ingestion_capacity.ingestion_efficiency >= 95 ? 'bg-green-500' :
                        healthMetrics.ingestion_capacity.ingestion_efficiency >= 85 ? 'bg-yellow-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${healthMetrics.ingestion_capacity.ingestion_efficiency}%` }}
                    ></div>
                  </div>
                  <span className="text-xs opacity-50 mt-1 block">ⓘ Hover for calculation</span>
                </div>
              </Tooltip>

              <Tooltip content={healthMetrics.ingestion_capacity.calculations?.sustainable_hours || "How long the current ingestion rate can be sustained"}>
                <div className={`p-4 rounded-lg border border-dashed border-gray-300 hover:border-gray-500 transition-colors ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                  <h3 className={`font-semibold mb-2 ${darkMode ? 'text-white' : 'text-gray-900'}`}>Sustainability</h3>
                  <div className="flex items-center justify-between">
                    <span className={`text-sm ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>At Current Rate:</span>
                    <span className={`font-bold ${
                      healthMetrics.ingestion_capacity.sustainable_hours >= 168 ? 'text-green-600' :
                      healthMetrics.ingestion_capacity.sustainable_hours >= 72 ? 'text-yellow-600' : 'text-red-600'
                    }`}>
                      {healthMetrics.ingestion_capacity.sustainable_hours >= 168 ? 
                        `${Math.floor(healthMetrics.ingestion_capacity.sustainable_hours / 168)} week${Math.floor(healthMetrics.ingestion_capacity.sustainable_hours / 168) > 1 ? 's' : ''}` :
                        healthMetrics.ingestion_capacity.sustainable_hours >= 24 ?
                        `${Math.floor(healthMetrics.ingestion_capacity.sustainable_hours / 24)} day${Math.floor(healthMetrics.ingestion_capacity.sustainable_hours / 24) > 1 ? 's' : ''}` :
                        `${healthMetrics.ingestion_capacity.sustainable_hours} hours`
                      }
                    </span>
                  </div>
                  <span className="text-xs opacity-50 mt-1 block">ⓘ Hover for calculation</span>
                </div>
              </Tooltip>
            </div>

            {/* Data Source Information */}
            {healthMetrics.ingestion_capacity.metadata && (
              <div className={`mt-6 p-4 rounded-lg ${darkMode ? 'bg-gray-700 border-gray-600' : 'bg-blue-50 border-blue-200'} border`}>
                <h4 className={`font-semibold mb-2 ${darkMode ? 'text-white' : 'text-gray-900'}`}>📈 Calculation Methodology</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <strong className={darkMode ? 'text-gray-300' : 'text-gray-700'}>Data Source:</strong>
                    <span className={`ml-2 ${darkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                      {healthMetrics.ingestion_capacity.metadata.metrics_source || 'Mimir/Cortex metrics'}
                    </span>
                  </div>
                  <div>
                    <strong className={darkMode ? 'text-gray-300' : 'text-gray-700'}>Method:</strong>
                    <span className={`ml-2 ${darkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                      {healthMetrics.ingestion_capacity.metadata.estimation_method || 'Real-time calculation'}
                    </span>
                  </div>
                  {healthMetrics.ingestion_capacity.active_series && (
                    <div>
                      <strong className={darkMode ? 'text-gray-300' : 'text-gray-700'}>Active Series:</strong>
                      <span className={`ml-2 ${darkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                        {healthMetrics.ingestion_capacity.active_series.toLocaleString()}
                      </span>
                    </div>
                  )}
                  <div>
                    <strong className={darkMode ? 'text-gray-300' : 'text-gray-700'}>Last Updated:</strong>
                    <span className={`ml-2 ${darkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                      {healthMetrics.ingestion_capacity.metadata.calculation_timestamp 
                        ? new Date(healthMetrics.ingestion_capacity.metadata.calculation_timestamp).toLocaleTimeString()
                        : 'N/A'
                      }
                    </span>
                  </div>
                </div>
              </div>
            )}
          </div>
          )}

          {/* Health Summary Cards */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            {Object.entries(healthMetrics.health_summary || {}).map(([status, count]) => (
              <div
                key={status}
                className={`rounded-lg shadow-sm border p-4 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className={`text-2xl font-bold ${getStatusColor(status).split(' ')[0]}`}>
                      {count}
                    </p>
                    <p className={`text-sm capitalize ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>{status}</p>
                  </div>
                  <div className={`w-3 h-3 rounded-full ${getStatusColor(status).split(' ')[1]}`}></div>
                </div>
              </div>
            ))}
          </div>

          {/* Component Types Overview */}
          <div className={`rounded-xl shadow-sm border p-6 mb-8 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
            <h2 className={`text-xl font-semibold mb-4 ${darkMode ? 'text-white' : 'text-gray-900'}`}>Infrastructure Components</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
              {Object.entries(healthMetrics.components_count || {}).map(([type, count]) => (
                <div key={type} className="text-center">
                  <div className={`text-2xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>{count}</div>
                  <p className={`text-sm capitalize ${darkMode ? 'text-gray-300' : 'text-gray-600'}`}>{type}</p>
                </div>
              ))}
            </div>
          </div>
        </>
      )}

      {/* Alerts Section */}
      {safeAlerts.length > 0 && (
        <div className={`rounded-xl shadow-sm border p-6 mb-8 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
          <h2 className={`text-xl font-semibold mb-4 ${darkMode ? 'text-white' : 'text-gray-900'}`}>Active Alerts</h2>
          <div className="space-y-3">
            {safeAlerts.map((alert) => (
              <div key={alert.id} className="border border-gray-200 rounded-lg p-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getSeverityColor(alert.severity)}`}>
                        {alert.severity}
                      </span>
                      <span className="text-gray-500 text-sm">{alert.component}</span>
                    </div>
                    <h3 className="font-semibold text-gray-900">{alert.title}</h3>
                    <p className="text-gray-600 text-sm mt-1">{alert.description}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-gray-500 text-xs">
                      {new Date(alert.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* AI Recommendations */}
      {safeRecommendations.length > 0 && (
        <div className={`rounded-xl shadow-sm border p-6 mb-8 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
          <h2 className={`text-xl font-semibold mb-4 ${darkMode ? 'text-white' : 'text-gray-900'}`}>🤖 AI Recommendations</h2>
          <div className="space-y-4">
            {safeRecommendations.map((rec) => (
              <div key={rec.id} className="border border-gray-200 rounded-lg p-4">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getPriorityColor(rec.priority)}`}>
                      {rec.priority} Priority
                    </span>
                    <span className="text-gray-500 text-sm">{rec.category}</span>
                  </div>
                </div>
                <h3 className="font-semibold text-gray-900 mb-2">{rec.title}</h3>
                <p className="text-gray-600 text-sm mb-3">{rec.description}</p>
                <div className="bg-gray-50 rounded-lg p-3">
                  <p className="text-sm"><span className="font-medium">Action:</span> {rec.action}</p>
                  <p className="text-sm mt-1"><span className="font-medium">Impact:</span> {rec.impact}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Resources Table */}
      <div className={`rounded-xl shadow-sm border p-6 ${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
        <div className="flex items-center justify-between mb-4">
          <h2 className={`text-xl font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>Infrastructure Resources</h2>
          <div className="flex space-x-4">
            <select
              value={filterKind}
              onChange={(e) => setFilterKind(e.target.value)}
              className="border border-gray-300 rounded-lg px-3 py-2 text-sm"
            >
              <option value="">All Types</option>
              <option value="Deployment">Deployments</option>
              <option value="StatefulSet">StatefulSets</option>
              <option value="DaemonSet">DaemonSets</option>
              <option value="Service">Services</option>
              <option value="Pod">Pods</option>
              <option value="ConfigMap">ConfigMaps</option>
              <option value="Secret">Secrets</option>
              <option value="PersistentVolumeClaim">PVCs</option>
            </select>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="border border-gray-300 rounded-lg px-3 py-2 text-sm"
            >
              <option value="">All Status</option>
              <option value="Healthy">Healthy</option>
              <option value="Warning">Warning</option>
              <option value="Critical">Critical</option>
              <option value="Unknown">Unknown</option>
            </select>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Resource
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Health Score
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Replicas
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Issues
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Age
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredResources.map((resource, index) => (
                <tr 
                  key={`${resource.kind}-${resource.name}`}
                  className="hover:bg-gray-50 cursor-pointer"
                  onClick={() => setSelectedResource(resource)}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <div className="text-sm font-medium text-gray-900">{resource.name}</div>
                      <div className="text-sm text-gray-500">{resource.kind}</div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(resource.status)}`}>
                      {resource.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="text-sm text-gray-900">{resource.health_score.toFixed(1)}%</div>
                      <div className="ml-2 w-16 bg-gray-200 rounded-full h-2">
                        <div 
                          className={`h-2 rounded-full ${
                            resource.health_score >= 95 ? 'bg-green-500' :
                            resource.health_score >= 80 ? 'bg-yellow-500' : 'bg-red-500'
                          }`}
                          style={{ width: `${resource.health_score}%` }}
                        ></div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {resource.replicas ? (
                      <span title={`Desired: ${resource.replicas.desired}, Ready: ${resource.replicas.ready}, Available: ${resource.replicas.available}, Unavailable: ${resource.replicas.unavailable}`}>
                        {resource.replicas.ready}/{resource.replicas.desired}
                      </span>
                    ) : (
                      '-'
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {(resource.issues || []).length > 0 ? (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-600">
                        {(resource.issues || []).length} issue{(resource.issues || []).length > 1 ? 's' : ''}
                      </span>
                    ) : (
                      <span className="text-green-600">No issues</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {resource.age}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredResources.length === 0 && (
          <div className="text-center py-8">
            <p className="text-gray-500">No resources found matching the selected filters.</p>
          </div>
        )}
      </div>

      {/* Resource Detail Modal */}
      {selectedResource && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-2xl font-bold text-gray-900">{selectedResource.name}</h2>
                  <p className="text-gray-600">{selectedResource.kind} in {selectedResource.namespace}</p>
                </div>
                <button
                  onClick={() => setSelectedResource(null)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Status and Metrics */}
                <div>
                  <h3 className="text-lg font-semibold mb-3">Status & Metrics</h3>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span>Status:</span>
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(selectedResource.status)}`}>
                        {selectedResource.status}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>Health Score:</span>
                      <span>{selectedResource.health_score.toFixed(1)}%</span>
                    </div>
                    {selectedResource.replicas && (
                      <>
                        <div className="flex justify-between">
                          <span>Desired Replicas:</span>
                          <span>{selectedResource.replicas.desired}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Ready Replicas:</span>
                          <span>{selectedResource.replicas.ready}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Available Replicas:</span>
                          <span>{selectedResource.replicas.available}</span>
                        </div>
                      </>
                    )}
                  </div>
                </div>

                {/* Issues */}
                <div>
                  <h3 className="text-lg font-semibold mb-3">Issues</h3>
                  {(selectedResource.issues || []).length > 0 ? (
                    <div className="space-y-2">
                      {(selectedResource.issues || []).map((issue, index) => (
                        <div key={index} className="border border-gray-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2 mb-1">
                            <span className={`px-2 py-1 rounded-full text-xs font-medium ${getSeverityColor(issue.severity)}`}>
                              {issue.severity}
                            </span>
                            <span className="text-sm text-gray-500">{issue.category}</span>
                          </div>
                          <h4 className="font-medium text-gray-900">{issue.title}</h4>
                          <p className="text-sm text-gray-600 mt-1">{issue.description}</p>
                          {issue.suggestion && (
                            <p className="text-sm text-blue-600 mt-2">💡 {issue.suggestion}</p>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-green-600">No issues detected</p>
                  )}
                </div>
              </div>

              {/* Conditions */}
              {(selectedResource.conditions || []).length > 0 && (
                <div className="mt-6">
                  <h3 className="text-lg font-semibold mb-3">Conditions</h3>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Reason</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Message</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {(selectedResource.conditions || []).map((condition, index) => (
                          <tr key={index}>
                            <td className="px-4 py-2 text-sm font-medium text-gray-900">{condition.type}</td>
                            <td className="px-4 py-2 text-sm text-gray-900">{condition.status}</td>
                            <td className="px-4 py-2 text-sm text-gray-900">{condition.reason || '-'}</td>
                            <td className="px-4 py-2 text-sm text-gray-900">{condition.message || '-'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* Labels */}
              {Object.keys(selectedResource.labels || {}).length > 0 && (
                <div className="mt-6">
                  <h3 className="text-lg font-semibold mb-3">Labels</h3>
                  <div className="flex flex-wrap gap-2">
                    {Object.entries(selectedResource.labels || {}).map(([key, value]) => (
                      <span key={key} className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-100 text-gray-800">
                        {key}: {value}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default HealthDashboard; 