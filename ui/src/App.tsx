import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ApiProvider } from './context/ApiContext';
import { ThemeProvider } from './context/ThemeContext';
import Sidebar from './components/layout/Sidebar';
import Header from './components/layout/Header';
import Dashboard from './pages/Dashboard';
import HealthDashboard from './pages/HealthDashboard';
import InfrastructureDashboard from './pages/InfrastructureDashboard';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import Config from './pages/Config';
import AuditLog from './pages/AuditLog';
import DiffViewer from './pages/DiffViewer';
import Metrics from './pages/Metrics';
import TestTools from './pages/TestTools';
import './App.css';

// Error Boundary Component
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
          <div className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-lg max-w-md w-full">
            <div className="text-center">
              <div className="text-6xl mb-4">❌</div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                Something went wrong
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mb-6">
                {this.state.error?.message || 'An unexpected error occurred'}
              </p>
              <button
                onClick={() => window.location.reload()}
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors"
              >
                Reload Page
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Loading Component
const LoadingSpinner: React.FC<{ message?: string }> = ({ message = 'Loading...' }) => (
  <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
      <p className="text-gray-600 dark:text-gray-400">{message}</p>
    </div>
  </div>
);

// Main App Component
const App: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [darkMode, setDarkMode] = useState(() => {
    const savedMode = localStorage.getItem('darkMode');
    return savedMode ? JSON.parse(savedMode) : false;
  });

  // Update localStorage and document class when darkMode changes
  useEffect(() => {
    localStorage.setItem('darkMode', JSON.stringify(darkMode));
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  const themeValue = {
    darkMode,
    setDarkMode,
  };

  return (
    <ErrorBoundary>
      <ThemeProvider value={themeValue}>
        <ApiProvider>
          <Router>
            <div className="flex h-screen bg-gray-50 dark:bg-gray-900">
              {/* Sidebar */}
              <Sidebar open={sidebarOpen} onToggle={toggleSidebar} />

              {/* Main Content */}
              <div className="flex flex-col flex-1 overflow-hidden">
                {/* Header */}
                <Header 
                  onSidebarToggle={toggleSidebar}
                  darkMode={darkMode}
                  onDarkModeToggle={toggleDarkMode}
                />

                {/* Page Content */}
                <main className="flex-1 overflow-y-auto p-6">
                  <Routes>
                    {/* Default redirect */}
                    <Route path="/" element={<Navigate to="/dashboard" replace />} />
                    
                    {/* Dashboard Routes */}
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/health" element={<HealthDashboard />} />
                    <Route path="/infrastructure" element={<InfrastructureDashboard />} />
                    
                    {/* Tenant Routes */}
                    <Route path="/tenants" element={<Tenants />} />
                    <Route path="/tenants/:tenantId" element={<TenantDetail />} />
                    
                    {/* Configuration Routes */}
                    <Route path="/config" element={<Config />} />
                    
                    {/* Analysis Routes */}
                    <Route path="/audit" element={<AuditLog />} />
                    <Route path="/diff" element={<DiffViewer />} />
                    <Route path="/metrics" element={<Metrics />} />
                    
                    {/* Utility Routes */}
                    <Route path="/test-tools" element={<TestTools />} />
                    
                    {/* Catch all - redirect to dashboard */}
                    <Route path="*" element={<Navigate to="/dashboard" replace />} />
                  </Routes>
                </main>
              </div>
            </div>
          </Router>
        </ApiProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
};

export default App; 