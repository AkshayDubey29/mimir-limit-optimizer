import React, { useState, useEffect } from 'react';
import { useApi } from '../../context/ApiContext';
import { Link, useLocation } from 'react-router-dom';

interface HeaderProps {
  onSidebarToggle: () => void;
  darkMode: boolean;
  onDarkModeToggle: () => void;
}

const Header: React.FC<HeaderProps> = ({ onSidebarToggle, darkMode, onDarkModeToggle }) => {
  const { getStatus } = useApi();
  const [status, setStatus] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const location = useLocation();

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: '🏠' },
    { name: 'Health', href: '/health', icon: '💚' },
    { name: 'Infrastructure', href: '/infrastructure', icon: '🏗️' },
    { name: 'Tenants', href: '/tenants', icon: '👥' },
    { name: 'Metrics', href: '/metrics', icon: '📊' },
    { name: 'Config', href: '/config', icon: '⚙️' },
    { name: 'Audit', href: '/audit', icon: '📜' },
    { name: 'Diff', href: '/diff', icon: '🔍' },
    { name: 'Tools', href: '/test-tools', icon: '🧪' },
  ];

  const isActive = (path: string) => {
    return location.pathname === path;
  };

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
    <>
      {/* Main Header */}
      <header className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* Left side - Logo */}
            <div className="flex items-center space-x-4">
              <h1 className="text-xl font-bold text-gray-900 dark:text-white">
                📊 Mimir Infrastructure Dashboard
              </h1>
            </div>

            {/* Right side - Controls */}
            <div className="flex items-center space-x-4">
              {/* Status indicators */}
              <div className="flex items-center space-x-2 text-sm">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                  <span className="w-2 h-2 bg-green-400 rounded-full mr-1"></span>
                  Live Data
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                  Auto-refresh
                </span>
              </div>

              {/* Dark mode toggle */}
              <button
                onClick={onDarkModeToggle}
                className="p-2 rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                title={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {darkMode ? '☀️' : '🌙'}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation Menu Bar */}
      <nav className="bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-40">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-1 overflow-x-auto py-2">
            {navigation.map((item) => (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors duration-200 whitespace-nowrap
                  ${isActive(item.href)
                    ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200 border border-blue-200 dark:border-blue-700'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white'
                  }
                `}
              >
                <span className="text-base">{item.icon}</span>
                <span>{item.name}</span>
              </Link>
            ))}
          </div>
        </div>
      </nav>
    </>
  );
};

export default Header; 