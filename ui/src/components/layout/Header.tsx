import React, { useState, useEffect } from 'react';
import { useApi } from '../../context/ApiContext';

interface HeaderProps {
  onSidebarToggle: () => void;
  darkMode: boolean;
  onDarkModeToggle: () => void;
}

const Header: React.FC<HeaderProps> = ({ onSidebarToggle, darkMode, onDarkModeToggle }) => {
  const { getStatus } = useApi();
  const [status, setStatus] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        setLoading(true);
        const data = await getStatus();
        setStatus(data);
      } catch (error) {
        console.error('Failed to fetch status:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 30000); // Refresh every 30 seconds

    return () => clearInterval(interval);
  }, [getStatus]);

  return (
    <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <div className="flex items-center justify-between h-16 px-6">
        {/* Left side */}
        <div className="flex items-center space-x-4">
          <button
            onClick={onSidebarToggle}
            className="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 lg:hidden"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>

          {/* Status indicators */}
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <div className={`w-2 h-2 rounded-full ${loading ? 'bg-yellow-400 animate-pulse' : 'bg-green-400'}`}></div>
              <span className="text-sm text-gray-600 dark:text-gray-300">
                {status?.mode === 'dry-run' ? 'Dry Run' : 'Production'}
              </span>
            </div>

            {status && (
              <div className="flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-300">
                <div className="flex items-center space-x-1">
                  <span>Circuit Breaker:</span>
                  <span className={`font-medium ${status.circuit_breaker_state === 'CLOSED' ? 'text-green-600' : 'text-red-600'}`}>
                    {status.circuit_breaker_state}
                  </span>
                </div>
                <div className="flex items-center space-x-1">
                  <span>Tenants:</span>
                  <span className="font-medium">{status.monitored_tenants || 0}</span>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center space-x-4">
          {/* Last reconcile time */}
          {status && (
            <div className="text-sm text-gray-600 dark:text-gray-300">
              Last reconcile: {new Date(status.last_reconcile).toLocaleTimeString()}
            </div>
          )}

          {/* Dark mode toggle */}
          <button
            onClick={onDarkModeToggle}
            className="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            title={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {darkMode ? (
              <svg className="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clipRule="evenodd" />
              </svg>
            ) : (
              <svg className="w-5 h-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
              </svg>
            )}
          </button>

          {/* Notifications */}
          <button className="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 relative">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5-5V9a7 7 0 10-14 0v3l-5 5h5m9 0v1a3 3 0 01-6 0v-1m6 0H9" />
            </svg>
            {status?.panic_mode_active && (
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full animate-pulse"></span>
            )}
          </button>

          {/* User menu */}
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center">
              <span className="text-sm font-medium text-white">OP</span>
            </div>
            <span className="text-sm font-medium text-gray-700 dark:text-gray-200">Operator</span>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header; 