import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import './App.css';

// Components
import Sidebar from './components/layout/Sidebar';
import Header from './components/layout/Header';
import Dashboard from './pages/Dashboard';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import Config from './pages/Config';
import AuditLog from './pages/AuditLog';
import DiffViewer from './pages/DiffViewer';
import Metrics from './pages/Metrics';
import TestTools from './pages/TestTools';

// Context
import { ThemeProvider } from './context/ThemeContext';
import { ApiProvider } from './context/ApiContext';

function App() {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [darkMode, setDarkMode] = useState(() => {
    const saved = localStorage.getItem('darkMode');
    return saved ? JSON.parse(saved) : false;
  });

  useEffect(() => {
    localStorage.setItem('darkMode', JSON.stringify(darkMode));
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  return (
    <ThemeProvider value={{ darkMode, setDarkMode }}>
      <ApiProvider>
        <Router>
          <div className={`flex h-screen bg-gray-100 dark:bg-gray-900 ${darkMode ? 'dark' : ''}`}>
            <Sidebar open={sidebarOpen} onToggle={() => setSidebarOpen(!sidebarOpen)} />
            
            <div className="flex-1 flex flex-col overflow-hidden">
              <Header 
                onSidebarToggle={() => setSidebarOpen(!sidebarOpen)}
                darkMode={darkMode}
                onDarkModeToggle={() => setDarkMode(!darkMode)}
              />
              
              <main className="flex-1 overflow-x-hidden overflow-y-auto bg-gray-100 dark:bg-gray-900">
                <div className="container mx-auto px-6 py-8">
                  <Routes>
                    <Route path="/" element={<Dashboard />} />
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/tenants" element={<Tenants />} />
                    <Route path="/tenants/:tenantId" element={<TenantDetail />} />
                    <Route path="/config" element={<Config />} />
                    <Route path="/audit" element={<AuditLog />} />
                    <Route path="/diff" element={<DiffViewer />} />
                    <Route path="/metrics" element={<Metrics />} />
                    <Route path="/test-tools" element={<TestTools />} />
                  </Routes>
                </div>
              </main>
            </div>
          </div>
        </Router>
      </ApiProvider>
    </ThemeProvider>
  );
}

export default App; 