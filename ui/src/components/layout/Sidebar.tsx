import React from 'react';
import { Link, useLocation } from 'react-router-dom';

interface SidebarProps {
  open: boolean;
  onToggle: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ open, onToggle }) => {
  const location = useLocation();

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: '🏠' },
    { name: 'Health Dashboard', href: '/health', icon: '💚' },
    { name: 'Infrastructure Dashboard', href: '/infrastructure', icon: '🏗️' },
    { name: 'Tenants', href: '/tenants', icon: '👥' },
    { name: 'Metrics', href: '/metrics', icon: '📊' },
    { name: 'Configuration', href: '/config', icon: '⚙️' },
    { name: 'Audit Log', href: '/audit', icon: '📜' },
    { name: 'Diff Viewer', href: '/diff', icon: '🔍' },
    { name: 'Test Tools', href: '/test-tools', icon: '🧪' },
  ];

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  // Debug log to check if component is rendering
  console.log('Sidebar rendering:', { open, pathname: location.pathname });

  return (
    <>
      {/* Overlay for mobile when sidebar is open */}
      {open && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 z-40 lg:hidden"
          onClick={onToggle}
        />
      )}
      
      {/* Sidebar - Always visible on desktop, slide on mobile */}
      <div 
        className={`
          fixed top-0 left-0 h-full bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 z-50
          w-64 transition-transform duration-300 ease-in-out
          ${open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        `}
        style={{
          display: 'block',
          visibility: 'visible',
          width: '256px',
          minWidth: '256px'
        }}
      >
        {/* Logo/Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            📊 Mimir Optimizer
          </h2>
        </div>

        {/* Navigation */}
        <nav className="p-4 space-y-2 overflow-y-auto" style={{ height: 'calc(100vh - 80px)' }}>
          {navigation.map((item) => (
            <Link
              key={item.name}
              to={item.href}
              className={`
                flex items-center space-x-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-200
                ${isActive(item.href)
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                }
              `}
              onClick={() => {
                // Close sidebar on mobile after navigation
                if (window.innerWidth < 1024) {
                  onToggle();
                }
              }}
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.name}</span>
            </Link>
          ))}
        </nav>
      </div>
    </>
  );
};

export default Sidebar; 