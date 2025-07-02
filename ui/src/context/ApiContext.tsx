import React, { createContext, useContext, useState, useCallback } from 'react';
import axios, { AxiosError } from 'axios';

// API base URL - will use proxy in development
const API_BASE = '/api';

// Types
export interface SystemStatus {
  mode: string;
  last_reconcile: string;
  reconcile_count: number;
  update_interval: string;
  components_health: Record<string, boolean>;
  circuit_breaker_state: string;
  spike_detection_state: string;
  panic_mode_active: boolean;
  total_tenants: number;
  monitored_tenants: number;
  skipped_tenants: number;
  config_map_name: string;
  version: string;
  build_info: {
    version: string;
    commit: string;
    build_date: string;
  };
}

export interface TenantInfo {
  id: string;
  ingestion_rate: number;
  active_series: number;
  applied_limits: Record<string, any>;
  suggested_limits: Record<string, any>;
  spike_detected: boolean;
  last_config_change: string;
  buffer_usage_percent: number;
  usage_sparkline: number[];
  status: string;
}

export interface ConfigUpdateRequest {
  mode?: string;
  buffer_percentage?: number;
  spike_threshold?: number;
  update_interval?: string;
  circuit_breaker_enabled?: boolean;
  auto_discovery_enabled?: boolean;
  skip_list?: string[];
  include_list?: string[];
  enabled_limits?: string[];
}

export interface DiffItem {
  limit_name: string;
  dry_run_value: any;
  applied_value: any;
  delta: any;
  status: 'identical' | 'mismatched' | 'dry_run_only';
  tenant_id: string;
}

export interface AuditEntry {
  id: string;
  timestamp: string;
  action: string;
  tenant_id?: string;
  user?: string;
  reason: string;
  changes: Record<string, any>;
  success: boolean;
  error?: string;
}

interface ApiContextType {
  // Generic API request method
  apiRequest: (endpoint: string, options?: Record<string, any>) => Promise<any>;
  
  // System
  getStatus: () => Promise<SystemStatus>;
  getConfig: () => Promise<any>;
  updateConfig: (config: ConfigUpdateRequest) => Promise<void>;
  
  // Tenants
  getTenants: () => Promise<{ tenants: TenantInfo[]; total_tenants: number; monitored_count: number; skipped_count: number; }>;
  getTenantDetail: (tenantId: string) => Promise<any>;
  
  // Analysis
  getDiff: () => Promise<{ differences: DiffItem[]; total_diffs: number; identical_count: number; mismatched_count: number; }>;
  getAudit: (filters?: Record<string, any>) => Promise<{ entries: AuditEntry[]; total: number; }>;
  
  // Testing
  triggerTestSpike: (tenantId: string, multiplier: number, duration: string) => Promise<void>;
  triggerTestAlert: (channel: string, message: string) => Promise<void>;
  triggerReconcile: () => Promise<void>;
  
  // State
  loading: boolean;
  error: string | null;
  clearError: () => void;
}

const ApiContext = createContext<ApiContextType | undefined>(undefined);

export const ApiProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleError = useCallback((err: unknown) => {
    if (axios.isAxiosError(err)) {
      const axiosError = err as AxiosError<{ message?: string }>;
      setError(axiosError.response?.data?.message || axiosError.message || 'An error occurred');
    } else if (err instanceof Error) {
      setError(err.message);
    } else {
      setError('An unknown error occurred');
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const apiCall = useCallback(async (fn: () => Promise<any>): Promise<any> => {
    setLoading(true);
    setError(null);
    try {
      const result = await fn();
      return result;
    } catch (err) {
      handleError(err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [handleError]);

  // System APIs
  const getStatus = useCallback(() => 
    apiCall(() => axios.get(`${API_BASE}/status`).then(res => res.data)), 
    [apiCall]
  );

  const getConfig = useCallback(() => 
    apiCall(() => axios.get(`${API_BASE}/config`).then(res => res.data)), 
    [apiCall]
  );

  const updateConfig = useCallback((config: ConfigUpdateRequest) => 
    apiCall(() => axios.post(`${API_BASE}/config`, config).then(res => res.data)), 
    [apiCall]
  );

  // Tenant APIs
  const getTenants = useCallback(() => 
    apiCall(() => axios.get(`${API_BASE}/tenants`).then(res => res.data)), 
    [apiCall]
  );

  const getTenantDetail = useCallback((tenantId: string) => 
    apiCall(() => axios.get(`${API_BASE}/tenants/${tenantId}`).then(res => res.data)), 
    [apiCall]
  );

  // Analysis APIs
  const getDiff = useCallback(() => 
    apiCall(() => axios.get(`${API_BASE}/diff`).then(res => res.data)), 
    [apiCall]
  );

  const getAudit = useCallback((filters?: Record<string, any>) => 
    apiCall(() => axios.get(`${API_BASE}/audit`, { params: filters }).then(res => res.data)), 
    [apiCall]
  );

  // Test APIs
  const triggerTestSpike = useCallback((tenantId: string, multiplier: number, duration: string) => 
    apiCall(() => axios.post(`${API_BASE}/test/spike`, { tenant_id: tenantId, multiplier, duration }).then(res => res.data)), 
    [apiCall]
  );

  const triggerTestAlert = useCallback((channel: string, message: string) => 
    apiCall(() => axios.post(`${API_BASE}/test/alert`, { channel, message }).then(res => res.data)), 
    [apiCall]
  );

  const triggerReconcile = useCallback(() => 
    apiCall(() => axios.post(`${API_BASE}/test/reconcile`).then(res => res.data)), 
    [apiCall]
  );

  // Generic API request method
  const apiRequest = useCallback((endpoint: string, options?: Record<string, any>) => 
    apiCall(() => axios.get(`${API_BASE}${endpoint}`, options).then(res => res.data)), 
    [apiCall]
  );

  const value: ApiContextType = {
    apiRequest,
    getStatus,
    getConfig,
    updateConfig,
    getTenants,
    getTenantDetail,
    getDiff,
    getAudit,
    triggerTestSpike,
    triggerTestAlert,
    triggerReconcile,
    loading,
    error,
    clearError,
  };

  return <ApiContext.Provider value={value}>{children}</ApiContext.Provider>;
};

export const useApi = () => {
  const context = useContext(ApiContext);
  if (context === undefined) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return context;
}; 