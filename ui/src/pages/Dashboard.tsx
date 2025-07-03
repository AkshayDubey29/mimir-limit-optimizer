import React, { useState, useEffect } from 'react';

interface DashboardData {
  system_status?: {
    mode?: string;
    total_tenants?: number;
    monitored_tenants?: number;
    last_reconcile?: string;
  };
  tenants?: {
    total_tenants?: number;
    monitored_tenants?: number;
    skipped_tenants?: number;
  };
  architecture_flow?: {
    flow?: any[];
    components?: any[];
  };
  namespaces?: {
    total?: number;
    namespaces?: any[];
  };
  timestamp?: string;
}

const Dashboard: React.FC = () => {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/dashboard');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const result = await response.json();
      setData(result);
      setError(null);
    } catch (err) {
      console.error('Dashboard fetch error:', err);
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  if (loading && !data) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <div className="text-red-800">
            <h2 className="text-lg font-semibold mb-2">Dashboard Error</h2>
            <p className="mb-4">{error}</p>
            <button 
              onClick={fetchData}
              className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                Mimir Infrastructure Dashboard
              </h1>
              <p className="mt-2 text-gray-600">
                Real-time monitoring and management of your Mimir infrastructure
              </p>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-sm text-gray-600">Live</span>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {/* System Mode */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-blue-100">
                <span className="text-2xl">⚙️</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">System Mode</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {data?.system_status?.mode || 'Unknown'}
                </p>
              </div>
            </div>
          </div>

          {/* Total Tenants */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-green-100">
                <span className="text-2xl">👥</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Total Tenants</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {data?.tenants?.total_tenants || data?.system_status?.total_tenants || 0}
                </p>
              </div>
            </div>
          </div>

          {/* Architecture Components */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-purple-100">
                <span className="text-2xl">🏗️</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Components</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {data?.architecture_flow?.flow?.length || 0}
                </p>
              </div>
            </div>
          </div>

          {/* Namespaces */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-3 rounded-full bg-orange-100">
                <span className="text-2xl">📦</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Namespaces</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {data?.namespaces?.total || 0}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* System Status Details */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* System Status */}
          <div className="bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">System Status</h3>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                <div className="flex justify-between">
                  <span className="text-gray-600">Mode:</span>
                  <span className="font-medium text-gray-900">
                    {data?.system_status?.mode || 'Unknown'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Total Tenants:</span>
                  <span className="font-medium text-gray-900">
                    {data?.tenants?.total_tenants || 0}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Monitored:</span>
                  <span className="font-medium text-gray-900">
                    {data?.tenants?.monitored_tenants || 0}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Skipped:</span>
                  <span className="font-medium text-gray-900">
                    {data?.tenants?.skipped_tenants || 0}
                  </span>
                </div>
                {data?.system_status?.last_reconcile && (
                  <div className="flex justify-between">
                    <span className="text-gray-600">Last Update:</span>
                    <span className="font-medium text-gray-900">
                      {new Date(data.system_status.last_reconcile).toLocaleString()}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Architecture Flow */}
          <div className="bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Architecture Flow</h3>
            </div>
            <div className="p-6">
              {data?.architecture_flow?.flow && data.architecture_flow.flow.length > 0 ? (
                <div className="space-y-3">
                  {data.architecture_flow.flow.slice(0, 5).map((flow: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm font-medium text-gray-900">
                          {flow.from || `Component ${index + 1}`}
                        </span>
                        <span className="text-gray-400">→</span>
                        <span className="text-sm font-medium text-gray-900">
                          {flow.to || `Target ${index + 1}`}
                        </span>
                      </div>
                      <div className="flex items-center space-x-2">
                        <div className={`w-2 h-2 rounded-full ${
                          flow.active ? 'bg-green-400' : 'bg-gray-400'
                        }`}></div>
                        <span className="text-xs text-gray-500">
                          {flow.active ? 'Active' : 'Inactive'}
                        </span>
                      </div>
                    </div>
                  ))}
                  {data.architecture_flow.flow.length > 5 && (
                    <p className="text-sm text-gray-500 text-center">
                      +{data.architecture_flow.flow.length - 5} more components
                    </p>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <span className="text-4xl">🏗️</span>
                  <p className="mt-2 text-gray-500">No architecture flow data available</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Raw Data (for debugging) */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">System Information</h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
              <div>
                <span className="font-medium text-gray-600">Timestamp:</span>
                <p className="text-gray-900">
                  {data?.timestamp ? new Date(data.timestamp).toLocaleString() : 'Not available'}
                </p>
              </div>
              <div>
                <span className="font-medium text-gray-600">Data Status:</span>
                <p className="text-green-600">✅ Loaded successfully</p>
              </div>
              <div>
                <span className="font-medium text-gray-600">Auto-refresh:</span>
                <p className="text-blue-600">Every 30 seconds</p>
              </div>
            </div>
          </div>
        </div>

        {/* Footer with refresh button */}
        <div className="mt-8 text-center">
          <button
            onClick={fetchData}
            disabled={loading}
            className={`px-6 py-2 rounded-lg font-medium ${
              loading
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                : 'bg-blue-600 text-white hover:bg-blue-700'
            }`}
          >
            {loading ? 'Refreshing...' : 'Refresh Data'}
          </button>
        </div>
      </main>
    </div>
  );
};

export default Dashboard; 