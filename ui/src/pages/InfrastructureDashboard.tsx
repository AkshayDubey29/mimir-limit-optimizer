import React, { useState, useEffect } from 'react';
import { useApi } from '../context/ApiContext';
import { useTheme } from '../context/ThemeContext';

// Type definitions for the infrastructure data
interface ComponentHealth {
  status: string;
  issues: string[];
  metrics: Record<string, number>;
  lastCheck: string;
}

interface MimirComponent {
  name: string;
  type: string;
  role: string;
  status: string;
  replicas: number;
  readyReplicas: number;
  services: ServiceInfo[];
  endpoints: EndpointInfo[];
  metricsUrls: string[];
  configuration: Record<string, any>;
  health: ComponentHealth;
}

interface ServiceInfo {
  name: string;
  type: string;
  clusterIp: string;
  ports: Record<string, number>;
  labels: Record<string, string>;
}

interface EndpointInfo {
  service: string;
  addresses: string[];
  ports: number[];
  ready: boolean;
}

interface TenantConfiguration {
  tenantId: string;
  source: string;
  limits: Record<string, any>;
  currentUsage: Record<string, number>;
  recommendedLimits: Record<string, any>;
  status: string;
  lastSeen: string;
}

interface InfrastructureRecommendation {
  id: string;
  type: string;
  priority: string;
  title: string;
  description: string;
  impact: string;
  action: string;
  component: string;
  tenant?: string;
  createdAt: string;
}

interface InfrastructureAnalytics {
  overview: {
    totalComponents: number;
    healthyComponents: number;
    unhealthyComponents: number;
    overallHealthScore: number;
    totalTenants: number;
    totalEndpoints: number;
    totalRecommendations: number;
  };
  componentHealth: Record<string, {
    status: string;
    replicas: number;
    ready: number;
    healthScore: number;
  }>;
  tenantDistribution: Record<string, number>;
  metricsEndpoints: {
    total: number;
    accessible: number;
    byRole: Record<string, number>;
  };
  recommendations: {
    byPriority: Record<string, number>;
    byType: Record<string, number>;
    critical: number;
  };
  resourceUtilization: {
    totalPods: number;
    totalServices: number;
    totalConfigMaps: number;
    totalSecrets: number;
  };
  lastScan: string;
}

interface ScanResult {
  namespace: string;
  components: Record<string, MimirComponent>;
  tenants: Record<string, TenantConfiguration>;
  recommendations: InfrastructureRecommendation[];
  lastScan: string;
  scanDuration: string;
  scanId: string;
}

export default function InfrastructureDashboard() {
  const { apiRequest } = useApi();
  const { darkMode } = useTheme();
  
  const [analytics, setAnalytics] = useState<InfrastructureAnalytics | null>(null);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [components, setComponents] = useState<Record<string, MimirComponent>>({});
  const [tenants, setTenants] = useState<Record<string, TenantConfiguration>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'components' | 'tenants' | 'recommendations'>('overview');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // Fetch comprehensive infrastructure data
  const fetchInfrastructureData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch all data in parallel for better performance
      const [analyticsResponse, scanResponse, componentsResponse, tenantsResponse] = await Promise.all([
        apiRequest('/api/infrastructure/analytics'),
        apiRequest('/api/infrastructure/scan'),
        apiRequest('/api/infrastructure/components'),
        apiRequest('/api/infrastructure/tenants')
      ]);

      setAnalytics(analyticsResponse);
      setScanResult(scanResponse);
      setComponents(componentsResponse.components || {});
      setTenants(tenantsResponse.tenants || {});
      setLastUpdate(new Date());
    } catch (err) {
      console.error('Failed to fetch infrastructure data:', err);
      setError('Failed to load infrastructure data. Please check your configuration.');
    } finally {
      setLoading(false);
    }
  };

  // Auto-refresh functionality
  useEffect(() => {
    fetchInfrastructureData();
    
    if (autoRefresh) {
      const interval = setInterval(fetchInfrastructureData, 30000); // 30 seconds
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  // Helper functions
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'healthy':
      case 'ready':
      case 'active':
        return 'text-green-600 bg-green-100';
      case 'degraded':
      case 'partial':
      case 'warning':
        return 'text-yellow-600 bg-yellow-100';
      case 'unhealthy':
      case 'critical':
      case 'not-ready':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority.toLowerCase()) {
      case 'high':
      case 'critical':
        return 'text-red-600 bg-red-100';
      case 'medium':
        return 'text-yellow-600 bg-yellow-100';
      case 'low':
        return 'text-green-600 bg-green-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  const formatDuration = (duration: string) => {
    // Convert duration string to more readable format
    return duration.replace(/(\d+)(\w+)/g, '$1 $2');
  };

  const calculateHealthPercentage = (healthy: number, total: number) => {
    return total > 0 ? Math.round((healthy / total) * 100) : 0;
  };

  if (loading && !analytics) {
    return (
      <div className={`min-h-screen p-6 ${darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'}`}>
        <div className="max-w-7xl mx-auto">
          <div className="flex items-center justify-center h-64">
            <div className="flex flex-col items-center space-y-4">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
              <div className="text-lg font-semibold">🔍 Scanning Mimir Infrastructure...</div>
              <div className="text-sm text-gray-500">This may take a few moments</div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`min-h-screen p-6 ${darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'}`}>
        <div className="max-w-7xl mx-auto">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <div className="text-red-600 text-lg font-semibold mb-2">⚠️ Infrastructure Scan Failed</div>
            <div className="text-red-700 mb-4">{error}</div>
            <button
              onClick={fetchInfrastructureData}
              className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors"
            >
              Retry Scan
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen p-6 ${darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'}`}>
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-blue-600">🤖 AI-Enabled Infrastructure Dashboard</h1>
            <p className="text-gray-600 mt-1">
              Comprehensive autonomous scanning of Mimir infrastructure in <code>{scanResult?.namespace}</code>
            </p>
          </div>
          <div className="flex items-center space-x-4">
            <div className="text-sm text-gray-500">
              Last update: {lastUpdate.toLocaleTimeString()}
            </div>
            <button
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`px-3 py-1 rounded-lg text-sm font-medium transition-colors ${
                autoRefresh 
                  ? 'bg-green-100 text-green-700 hover:bg-green-200' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {autoRefresh ? '🔄 Auto-refresh ON' : '⏸️ Auto-refresh OFF'}
            </button>
            <button
              onClick={fetchInfrastructureData}
              disabled={loading}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {loading ? '🔄 Scanning...' : '🔍 Rescan'}
            </button>
          </div>
        </div>

        {/* Key Metrics Overview */}
        {analytics && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className={`p-6 rounded-lg shadow ${darkMode ? 'bg-gray-800' : 'bg-white'}`}>
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-blue-600">{analytics.overview.totalComponents}</div>
                  <div className="text-sm text-gray-500">Total Components</div>
                </div>
                <div className="text-3xl">🏗️</div>
              </div>
              <div className="mt-4 flex items-center text-sm">
                <span className="text-green-600 font-medium">
                  {analytics.overview.healthyComponents} healthy
                </span>
                <span className="mx-2 text-gray-400">•</span>
                <span className="text-red-600 font-medium">
                  {analytics.overview.unhealthyComponents} issues
                </span>
              </div>
            </div>

            <div className={`p-6 rounded-lg shadow ${darkMode ? 'bg-gray-800' : 'bg-white'}`}>
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-green-600">
                    {Math.round(analytics.overview.overallHealthScore)}%
                  </div>
                  <div className="text-sm text-gray-500">Overall Health</div>
                </div>
                <div className="text-3xl">💚</div>
              </div>
              <div className="mt-4">
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-green-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${analytics.overview.overallHealthScore}%` }}
                  ></div>
                </div>
              </div>
            </div>

            <div className={`p-6 rounded-lg shadow ${darkMode ? 'bg-gray-800' : 'bg-white'}`}>
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-purple-600">{analytics.overview.totalTenants}</div>
                  <div className="text-sm text-gray-500">Discovered Tenants</div>
                </div>
                <div className="text-3xl">👥</div>
              </div>
              <div className="mt-4 text-sm text-gray-600">
                From {Object.keys(analytics.tenantDistribution).length} sources
              </div>
            </div>

            <div className={`p-6 rounded-lg shadow ${darkMode ? 'bg-gray-800' : 'bg-white'}`}>
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-orange-600">{analytics.recommendations.critical}</div>
                  <div className="text-sm text-gray-500">Critical Issues</div>
                </div>
                <div className="text-3xl">⚠️</div>
              </div>
              <div className="mt-4 text-sm text-gray-600">
                {analytics.overview.totalRecommendations} total recommendations
              </div>
            </div>
          </div>
        )}

        {/* Tab Navigation */}
        <div className={`${darkMode ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow`}>
          <div className="border-b border-gray-200">
            <nav className="flex space-x-8 px-6">
              {[
                { id: 'overview', label: '📊 Overview', icon: '📊' },
                { id: 'components', label: '🏗️ Components', icon: '🏗️' },
                { id: 'tenants', label: '👥 Tenants', icon: '👥' },
                { id: 'recommendations', label: '🤖 AI Recommendations', icon: '🤖' }
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          {/* Tab Content */}
          <div className="p-6">
            {activeTab === 'overview' && analytics && (
              <div className="space-y-6">
                {/* Scan Information */}
                <div className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                  <h3 className="text-lg font-semibold mb-3">🔍 Last Scan Information</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                    <div>
                      <span className="font-medium">Scan ID:</span> {scanResult?.scanId}
                    </div>
                    <div>
                      <span className="font-medium">Duration:</span> {scanResult?.scanDuration && formatDuration(scanResult.scanDuration)}
                    </div>
                    <div>
                      <span className="font-medium">Timestamp:</span> {new Date(analytics.lastScan).toLocaleString()}
                    </div>
                  </div>
                </div>

                {/* Resource Utilization */}
                <div>
                  <h3 className="text-lg font-semibold mb-4">📦 Resource Utilization</h3>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {[
                      { label: 'Pods', value: analytics.resourceUtilization.totalPods, icon: '🔵' },
                      { label: 'Services', value: analytics.resourceUtilization.totalServices, icon: '🌐' },
                      { label: 'ConfigMaps', value: analytics.resourceUtilization.totalConfigMaps, icon: '⚙️' },
                      { label: 'Secrets', value: analytics.resourceUtilization.totalSecrets, icon: '🔐' }
                    ].map((resource) => (
                      <div key={resource.label} className={`p-4 rounded-lg text-center ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                        <div className="text-2xl mb-1">{resource.icon}</div>
                        <div className="text-xl font-bold">{resource.value}</div>
                        <div className="text-sm text-gray-500">{resource.label}</div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Metrics Endpoints */}
                <div>
                  <h3 className="text-lg font-semibold mb-4">📡 Metrics Coverage</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                      <div className="text-2xl font-bold text-blue-600">{analytics.metricsEndpoints.total}</div>
                      <div className="text-sm text-gray-500">Total Endpoints</div>
                    </div>
                    <div className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                      <div className="text-2xl font-bold text-green-600">{analytics.metricsEndpoints.accessible}</div>
                      <div className="text-sm text-gray-500">Accessible</div>
                    </div>
                    <div className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                      <div className="text-2xl font-bold text-purple-600">
                        {Math.round((analytics.metricsEndpoints.accessible / analytics.metricsEndpoints.total) * 100)}%
                      </div>
                      <div className="text-sm text-gray-500">Coverage</div>
                    </div>
                  </div>
                </div>

                {/* Tenant Distribution */}
                <div>
                  <h3 className="text-lg font-semibold mb-4">👥 Tenant Distribution</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {Object.entries(analytics.tenantDistribution).map(([source, count]) => (
                      <div key={source} className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-50'}`}>
                        <div className="flex justify-between items-center">
                          <div className="text-sm font-medium capitalize">{source}</div>
                          <div className="text-lg font-bold text-blue-600">{count}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'components' && (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold">🏗️ Discovered Mimir Components</h3>
                  <div className="text-sm text-gray-500">{Object.keys(components).length} components found</div>
                </div>
                
                <div className="grid gap-4">
                  {Object.entries(components).map(([name, component]) => (
                    <div key={name} className={`p-4 border rounded-lg ${darkMode ? 'bg-gray-700 border-gray-600' : 'bg-white border-gray-200'}`}>
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h4 className="font-semibold text-lg">{component.name}</h4>
                          <div className="flex items-center space-x-2 text-sm text-gray-500">
                            <span className="capitalize">{component.type}</span>
                            <span>•</span>
                            <span className="capitalize">{component.role}</span>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(component.health.status)}`}>
                            {component.health.status}
                          </span>
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                        <div>
                          <div className="font-medium mb-1">Replicas</div>
                          <div>{component.readyReplicas}/{component.replicas}</div>
                        </div>
                        <div>
                          <div className="font-medium mb-1">Services</div>
                          <div>{component.services.length}</div>
                        </div>
                        <div>
                          <div className="font-medium mb-1">Metrics URLs</div>
                          <div>{component.metricsUrls.length}</div>
                        </div>
                      </div>

                      {component.health.issues.length > 0 && (
                        <div className="mt-3 p-2 bg-red-50 rounded text-sm">
                          <div className="font-medium text-red-700 mb-1">Issues:</div>
                          {component.health.issues.map((issue, idx) => (
                            <div key={idx} className="text-red-600">• {issue}</div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {activeTab === 'tenants' && (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold">👥 Discovered Tenants</h3>
                  <div className="text-sm text-gray-500">{Object.keys(tenants).length} tenants found</div>
                </div>
                
                <div className="grid gap-4">
                  {Object.entries(tenants).map(([tenantId, tenant]) => (
                    <div key={tenantId} className={`p-4 border rounded-lg ${darkMode ? 'bg-gray-700 border-gray-600' : 'bg-white border-gray-200'}`}>
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h4 className="font-semibold text-lg">{tenant.tenantId}</h4>
                          <div className="flex items-center space-x-2 text-sm text-gray-500">
                            <span>Source: {tenant.source}</span>
                            <span>•</span>
                            <span className={`capitalize ${getStatusColor(tenant.status)} px-2 py-1 rounded`}>
                              {tenant.status}
                            </span>
                          </div>
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                        <div>
                          <div className="font-medium mb-2">Current Limits</div>
                          <div className="space-y-1">
                            {Object.entries(tenant.limits).slice(0, 3).map(([key, value]) => (
                              <div key={key} className="flex justify-between">
                                <span className="text-gray-600">{key}:</span>
                                <span className="font-mono">{typeof value === 'number' ? value.toLocaleString() : String(value)}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                        <div>
                          <div className="font-medium mb-2">Current Usage</div>
                          <div className="space-y-1">
                            {Object.entries(tenant.currentUsage).slice(0, 3).map(([key, value]) => (
                              <div key={key} className="flex justify-between">
                                <span className="text-gray-600">{key}:</span>
                                <span className="font-mono">{value.toLocaleString()}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {activeTab === 'recommendations' && scanResult && (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold">🤖 AI-Generated Recommendations</h3>
                  <div className="text-sm text-gray-500">{scanResult.recommendations.length} recommendations</div>
                </div>
                
                <div className="grid gap-4">
                  {scanResult.recommendations.map((rec) => (
                    <div key={rec.id} className={`p-4 border rounded-lg ${darkMode ? 'bg-gray-700 border-gray-600' : 'bg-white border-gray-200'}`}>
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h4 className="font-semibold text-lg">{rec.title}</h4>
                          <div className="flex items-center space-x-2 text-sm text-gray-500 mt-1">
                            <span className="capitalize">{rec.type}</span>
                            <span>•</span>
                            <span>{rec.component}</span>
                            {rec.tenant && (
                              <>
                                <span>•</span>
                                <span>Tenant: {rec.tenant}</span>
                              </>
                            )}
                          </div>
                        </div>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getPriorityColor(rec.priority)}`}>
                          {rec.priority.toUpperCase()}
                        </span>
                      </div>
                      
                      <div className="text-sm text-gray-600 mb-3">{rec.description}</div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                        <div>
                          <div className="font-medium mb-1">Impact</div>
                          <div className="text-gray-600">{rec.impact}</div>
                        </div>
                        <div>
                          <div className="font-medium mb-1">Recommended Action</div>
                          <div className="text-gray-600">{rec.action}</div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
} 