import React, { useState, useEffect } from 'react';
import { useApi } from '../context/ApiContext';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';

interface MetricData {
  time?: string;
  count: number;
  name?: string;
}

interface Metrics {
  reconcileCount: MetricData[];
  tenantHealth: MetricData[];
}

const Dashboard: React.FC = () => {
  const { getStatus, getTenants, loading, error } = useApi();
  const [status, setStatus] = useState<any>(null);
  const [tenants, setTenants] = useState<any[]>([]);
  const [metrics, setMetrics] = useState<Metrics>({
    reconcileCount: [],
    tenantHealth: [],
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statusData, tenantsData] = await Promise.all([
          getStatus(),
          getTenants()
        ]);
        setStatus(statusData);
        setTenants(tenantsData.tenants || []);
        
        // Mock metrics data
        setMetrics({
          reconcileCount: [
            { time: '00:00', count: 12 },
            { time: '04:00', count: 15 },
            { time: '08:00', count: 18 },
            { time: '12:00', count: 22 },
            { time: '16:00', count: 19 },
            { time: '20:00', count: 16 },
          ],
          tenantHealth: [
            { name: 'Healthy', count: tenantsData.tenants?.filter(t => t.status === 'active').length || 0 },
            { name: 'Warning', count: tenantsData.tenants?.filter(t => t.spike_detected).length || 0 },
            { name: 'Error', count: 0 },
          ],
        });
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 60000); // Refresh every minute
    return () => clearInterval(interval);
  }, [getStatus, getTenants]);

  if (loading && !status) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700 rounded-lg p-4">
        <div className="text-red-800 dark:text-red-200">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">System Overview</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-300">
          Monitor your Mimir limit optimizer performance and tenant health
        </p>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatusCard
          title="Mode"
          value={status?.mode || 'Unknown'}
          icon="⚙️"
          color={status?.mode === 'prod' ? 'green' : 'blue'}
        />
        <StatusCard
          title="Total Tenants"
          value={status?.total_tenants || 0}
          icon="👥"
          color="blue"
        />
        <StatusCard
          title="Reconcile Count"
          value={status?.reconcile_count || 0}
          icon="🔄"
          color="purple"
        />
        <StatusCard
          title="Circuit Breaker"
          value={status?.circuit_breaker_state || 'Unknown'}
          icon="🔌"
          color={status?.circuit_breaker_state === 'CLOSED' ? 'green' : 'red'}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Reconcile Activity */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">
            Reconcile Activity (24h)
          </h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={metrics.reconcileCount}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="count" stroke="#3b82f6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Tenant Health */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">
            Tenant Health Status
          </h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={metrics.tenantHealth}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="count" fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Recent Activity</h3>
        </div>
        <div className="px-6 py-4">
          <div className="space-y-4">
            {tenants.slice(0, 5).map((tenant, index) => (
              <div key={index} className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className={`w-2 h-2 rounded-full ${tenant.spike_detected ? 'bg-red-400' : 'bg-green-400'}`}></div>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">{tenant.id}</span>
                  <span className="text-sm text-gray-500">
                    {tenant.ingestion_rate} samples/sec
                  </span>
                </div>
                <div className="text-sm text-gray-500">
                  {new Date(tenant.last_config_change).toLocaleTimeString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Component Health */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Component Health</h3>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {status?.components_health && Object.entries(status.components_health).map(([component, healthy]) => (
              <div key={component} className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${healthy ? 'bg-green-400' : 'bg-red-400'}`}></div>
                <span className="text-sm capitalize text-gray-700 dark:text-gray-300">
                  {component.replace('_', ' ')}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

interface StatusCardProps {
  title: string;
  value: string | number;
  icon: string;
  color: 'blue' | 'green' | 'red' | 'purple' | 'yellow';
}

const StatusCard: React.FC<StatusCardProps> = ({ title, value, icon, color }) => {
  const colorClasses = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    red: 'bg-red-500',
    purple: 'bg-purple-500',
    yellow: 'bg-yellow-500',
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-center">
        <div className={`p-3 rounded-full ${colorClasses[color]} bg-opacity-10`}>
          <span className="text-2xl">{icon}</span>
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
          <p className="text-2xl font-semibold text-gray-900 dark:text-white">{value}</p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard; 