import React, { useState, useEffect, useCallback } from 'react';
import { useApi } from '../context/ApiContext';
import { useTheme } from '../context/ThemeContext';
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell } from 'recharts';

interface MetricData {
  time?: string;
  count: number;
  name?: string;
}

interface Metrics {
  reconcileCount: MetricData[];
  tenantHealth: MetricData[];
}

interface ArchitectureFlowData {
  distributors: ComponentFlowInfo[];
  ingesters: ComponentFlowInfo[];
  queriers: ComponentFlowInfo[];
  query_frontends: ComponentFlowInfo[];
  compactors: ComponentFlowInfo[];
  store_gateways: ComponentFlowInfo[];
  flow: FlowStep[];
}

interface ComponentFlowInfo {
  name: string;
  status: string;
  connections: number;
  load: number;
  endpoint: string;
}

interface FlowStep {
  from: string;
  to: string;
  type: string;
  active: boolean;
  throughput: number;
  latency: number;
}

interface NamespaceInfo {
  name: string;
  namespace: string;
  status: string;
  creation_timestamp: string;
  health_score: number;
  mimir_components: ComponentInfo[];
  ingestion_rate: number;
  active_series: number;
  last_activity: string;
}

interface ComponentInfo {
  name: string;
  type: string;
  status: string;
  replicas: number;
  ready_replicas: number;
}

interface DashboardData {
  system_status: any;
  tenants: any;
  namespaces: {
    namespaces: NamespaceInfo[];
    total: number;
    scanned_at: string;
  };
  architecture_flow: any;
  metrics: any;
  timestamp: string;
}

const Dashboard: React.FC = () => {
  const { apiRequest, loading, error } = useApi();
  const { darkMode } = useTheme();
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [selectedTenant, setSelectedTenant] = useState<string>('');
  const [flowAnimation, setFlowAnimation] = useState<boolean>(true);

     const fetchDashboardData = useCallback(async () => {
     try {
       const data = await apiRequest('/api/dashboard') as DashboardData;
       setDashboardData(data);
     } catch (err) {
       console.error('Failed to fetch dashboard data:', err);
     }
   }, [apiRequest]);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [fetchDashboardData]);

  const getComponentColor = (status: string, load: number) => {
    if (status !== 'Running') return '#ef4444'; // red
    if (load > 80) return '#f59e0b'; // amber
    if (load > 60) return '#3b82f6'; // blue
    return '#10b981'; // green
  };

  const getStatusBadgeColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running': case 'active': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'pending': case 'warning': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      case 'error': case 'failed': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  const formatDuration = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diff = now.getTime() - time.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) return `${hours}h ${minutes}m ago`;
    return `${minutes}m ago`;
  };

  if (loading && !dashboardData) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700 rounded-lg p-6 m-6">
        <div className="text-red-800 dark:text-red-200">
          <h2 className="text-lg font-semibold mb-2">Dashboard Error</h2>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  const architectureFlow = dashboardData?.architecture_flow;
  const namespaces = dashboardData?.namespaces?.namespaces || [];
  const systemStatus = dashboardData?.system_status;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <div className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                Mimir Infrastructure Dashboard
              </h1>
              <p className="mt-2 text-gray-600 dark:text-gray-300">
                Real-time monitoring and visualization of your Mimir infrastructure
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
                <span className="text-sm text-gray-600 dark:text-gray-300">Live</span>
              </div>
              <select
                value={selectedTenant}
                onChange={(e) => setSelectedTenant(e.target.value)}
                className="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              >
                <option value="">All Tenants</option>
                {namespaces.map((ns) => (
                  <option key={ns.name} value={ns.name}>{ns.name}</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* System Overview Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-blue-100 dark:bg-blue-900">
                <span className="text-2xl">🏗️</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total Namespaces</p>
                <p className="text-2xl font-semibold text-gray-900 dark:text-white">
                  {dashboardData?.namespaces?.total || 0}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-green-100 dark:bg-green-900">
                <span className="text-2xl">👥</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Active Tenants</p>
                <p className="text-2xl font-semibold text-gray-900 dark:text-white">
                  {systemStatus?.monitored_tenants || 0}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-purple-100 dark:bg-purple-900">
                <span className="text-2xl">🔄</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Reconcile Count</p>
                <p className="text-2xl font-semibold text-gray-900 dark:text-white">
                  {systemStatus?.reconcile_count || 0}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-yellow-100 dark:bg-yellow-900">
                <span className="text-2xl">⚡</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Circuit Breaker</p>
                <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                  systemStatus?.circuit_breaker_state === 'CLOSED' 
                    ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                    : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                }`}>
                  {systemStatus?.circuit_breaker_state || 'Unknown'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Mimir Architecture Flow Visualization */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Mimir Architecture Flow
              </h3>
              <button
                onClick={() => setFlowAnimation(!flowAnimation)}
                className="px-3 py-1 text-sm bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-md hover:bg-blue-200 dark:hover:bg-blue-800"
              >
                {flowAnimation ? 'Pause Animation' : 'Start Animation'}
              </button>
            </div>
          </div>
          <div className="p-6">
            <MimirArchitectureVisualization 
              architectureFlow={architectureFlow} 
              animated={flowAnimation}
              darkMode={darkMode}
            />
          </div>
        </div>

        {/* Tenant Namespaces Overview */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Namespace Health Chart */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Namespace Health Distribution
              </h3>
            </div>
            <div className="p-6">
              <NamespaceHealthChart namespaces={namespaces} darkMode={darkMode} />
            </div>
          </div>

          {/* Ingestion Rate Trends */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Ingestion Rate by Namespace
              </h3>
            </div>
            <div className="p-6">
              <IngestionRateChart namespaces={namespaces} darkMode={darkMode} />
            </div>
          </div>
        </div>

        {/* Detailed Namespace Information */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Tenant Namespaces Details
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Namespace
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Health Score
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Components
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Ingestion Rate
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Active Series
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Last Activity
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {namespaces.map((namespace) => (
                  <tr key={namespace.name} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900 dark:text-white">
                        {namespace.name}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusBadgeColor(namespace.status)}`}>
                        {namespace.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="text-sm text-gray-900 dark:text-white mr-2">
                          {namespace.health_score}%
                        </div>
                        <div className="w-16 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${
                              namespace.health_score >= 90 ? 'bg-green-600' :
                              namespace.health_score >= 70 ? 'bg-yellow-600' : 'bg-red-600'
                            }`}
                            style={{ width: `${namespace.health_score}%` }}
                          ></div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex -space-x-1">
                        {namespace.mimir_components?.slice(0, 3).map((component, idx) => (
                          <div
                            key={idx}
                            className={`inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-medium border-2 border-white dark:border-gray-800 ${
                              component.status === 'Running' ? 'bg-green-500' : 'bg-red-500'
                            }`}
                            title={`${component.name} (${component.status})`}
                          >
                            {component.type.charAt(0).toUpperCase()}
                          </div>
                        ))}
                        {namespace.mimir_components?.length > 3 && (
                          <div className="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-medium bg-gray-500 text-white border-2 border-white dark:border-gray-800">
                            +{namespace.mimir_components.length - 3}
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {formatNumber(namespace.ingestion_rate || 0)}/s
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {formatNumber(namespace.active_series || 0)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {formatDuration(namespace.last_activity)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

// Mimir Architecture Visualization Component
const MimirArchitectureVisualization: React.FC<{
  architectureFlow: any;
  animated: boolean;
  darkMode: boolean;
}> = ({ architectureFlow, animated, darkMode }) => {
  if (!architectureFlow) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        <div>
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p>Loading architecture flow...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="relative h-96 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-800 dark:to-gray-900 rounded-lg p-6 overflow-hidden">
      {/* Flow Visualization */}
      <svg width="100%" height="100%" viewBox="0 0 800 300" className="absolute inset-0">
        {/* Background Grid */}
        <defs>
          <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" fill="none" stroke={darkMode ? "#374151" : "#e5e7eb"} strokeWidth="1"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />

        {/* Flow Arrows */}
        {animated && (
          <>
            {/* Prometheus to Distributor */}
            <line x1="50" y1="150" x2="150" y2="100" stroke="#3b82f6" strokeWidth="3" strokeDasharray="5,5">
              <animate attributeName="stroke-dashoffset" values="0;-10" dur="1s" repeatCount="indefinite"/>
            </line>
            
            {/* Distributor to Ingester */}
            <line x1="200" y1="100" x2="350" y2="80" stroke="#10b981" strokeWidth="3" strokeDasharray="5,5">
              <animate attributeName="stroke-dashoffset" values="0;-10" dur="1s" repeatCount="indefinite"/>
            </line>
            
            {/* Ingester to Store */}
            <line x1="400" y1="80" x2="550" y2="200" stroke="#f59e0b" strokeWidth="3" strokeDasharray="5,5">
              <animate attributeName="stroke-dashoffset" values="0;-10" dur="1.5s" repeatCount="indefinite"/>
            </line>

            {/* Query Flow */}
            <line x1="150" y1="200" x2="300" y2="180" stroke="#8b5cf6" strokeWidth="2" strokeDasharray="3,3">
              <animate attributeName="stroke-dashoffset" values="0;-6" dur="2s" repeatCount="indefinite"/>
            </line>
            <line x1="350" y1="180" x2="400" y2="120" stroke="#8b5cf6" strokeWidth="2" strokeDasharray="3,3">
              <animate attributeName="stroke-dashoffset" values="0;-6" dur="2s" repeatCount="indefinite"/>
            </line>
          </>
        )}
      </svg>

      {/* Components */}
      <div className="relative z-10 h-full">
        {/* Prometheus */}
        <ComponentNode
          x={20}
          y={130}
          name="Prometheus"
          status="External"
          load={0}
          icon="📊"
          darkMode={darkMode}
        />

        {/* Distributor */}
        <ComponentNode
          x={130}
          y={80}
          name="Distributor"
          status="Running"
          load={75}
          icon="🔄"
          darkMode={darkMode}
        />

        {/* Ingester */}
        <ComponentNode
          x={320}
          y={60}
          name="Ingester"
          status="Running"
          load={68}
          icon="📥"
          darkMode={darkMode}
        />

        {/* Store */}
        <ComponentNode
          x={520}
          y={180}
          name="Store"
          status="Running"
          load={45}
          icon="💾"
          darkMode={darkMode}
        />

        {/* Query Frontend */}
        <ComponentNode
          x={120}
          y={180}
          name="Query Frontend"
          status="Running"
          load={55}
          icon="🔍"
          darkMode={darkMode}
        />

        {/* Querier */}
        <ComponentNode
          x={280}
          y={160}
          name="Querier"
          status="Running"
          load={45}
          icon="❓"
          darkMode={darkMode}
        />

        {/* Compactor */}
        <ComponentNode
          x={620}
          y={100}
          name="Compactor"
          status="Running"
          load={30}
          icon="🗜️"
          darkMode={darkMode}
        />
      </div>

      {/* Legend */}
      <div className="absolute bottom-4 right-4 bg-white dark:bg-gray-800 rounded-lg p-3 shadow-lg">
        <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Flow Types</div>
        <div className="space-y-1">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-0.5 bg-blue-500"></div>
            <span className="text-xs text-gray-600 dark:text-gray-400">Ingestion</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-0.5 bg-purple-500"></div>
            <span className="text-xs text-gray-600 dark:text-gray-400">Query</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-0.5 bg-yellow-500"></div>
            <span className="text-xs text-gray-600 dark:text-gray-400">Storage</span>
          </div>
        </div>
      </div>
    </div>
  );
};

// Component Node for Architecture Visualization
const ComponentNode: React.FC<{
  x: number;
  y: number;
  name: string;
  status: string;
  load: number;
  icon: string;
  darkMode: boolean;
}> = ({ x, y, name, status, load, icon, darkMode }) => {
  const getStatusColor = () => {
    if (status === 'Running') {
      if (load > 80) return 'border-red-500 bg-red-50 dark:bg-red-900';
      if (load > 60) return 'border-yellow-500 bg-yellow-50 dark:bg-yellow-900';
      return 'border-green-500 bg-green-50 dark:bg-green-900';
    }
    return 'border-gray-500 bg-gray-50 dark:bg-gray-900';
  };

  return (
    <div
      className={`absolute transform -translate-x-1/2 -translate-y-1/2 ${getStatusColor()} border-2 rounded-lg p-3 shadow-lg min-w-max`}
      style={{ left: `${x}px`, top: `${y}px` }}
    >
      <div className="text-center">
        <div className="text-lg mb-1">{icon}</div>
        <div className="text-xs font-medium text-gray-900 dark:text-white whitespace-nowrap">
          {name}
        </div>
        {status === 'Running' && (
          <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
            {load}% load
          </div>
        )}
      </div>
    </div>
  );
};

// Namespace Health Chart Component
const NamespaceHealthChart: React.FC<{
  namespaces: NamespaceInfo[];
  darkMode: boolean;
}> = ({ namespaces, darkMode }) => {
  const healthData = [
    {
      name: 'Excellent (90-100%)',
      count: namespaces.filter(ns => ns.health_score >= 90).length,
      fill: '#10b981'
    },
    {
      name: 'Good (70-89%)',
      count: namespaces.filter(ns => ns.health_score >= 70 && ns.health_score < 90).length,
      fill: '#f59e0b'
    },
    {
      name: 'Poor (<70%)',
      count: namespaces.filter(ns => ns.health_score < 70).length,
      fill: '#ef4444'
    }
  ];

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={healthData}
          cx="50%"
          cy="50%"
          outerRadius={100}
          dataKey="count"
          label={({ name, count }) => `${name}: ${count}`}
        >
          {healthData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
        </Pie>
        <Tooltip />
      </PieChart>
    </ResponsiveContainer>
  );
};

// Ingestion Rate Chart Component
const IngestionRateChart: React.FC<{
  namespaces: NamespaceInfo[];
  darkMode: boolean;
}> = ({ namespaces, darkMode }) => {
  const chartData = namespaces.map(ns => ({
    name: ns.name.replace('tenant-', ''),
    rate: ns.ingestion_rate || 0,
  }));

  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" stroke={darkMode ? '#374151' : '#e5e7eb'} />
        <XAxis 
          dataKey="name" 
          tick={{ fill: darkMode ? '#d1d5db' : '#374151', fontSize: 12 }}
          axisLine={{ stroke: darkMode ? '#6b7280' : '#9ca3af' }}
        />
        <YAxis 
          tick={{ fill: darkMode ? '#d1d5db' : '#374151', fontSize: 12 }}
          axisLine={{ stroke: darkMode ? '#6b7280' : '#9ca3af' }}
        />
        <Tooltip 
          contentStyle={{
            backgroundColor: darkMode ? '#1f2937' : '#ffffff',
            border: `1px solid ${darkMode ? '#374151' : '#e5e7eb'}`,
            borderRadius: '8px',
            color: darkMode ? '#ffffff' : '#000000'
          }}
        />
        <Bar dataKey="rate" fill="#3b82f6" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
};

export default Dashboard; 