import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { motion } from 'framer-motion';

interface SidebarProps {
  open: boolean;
  onToggle: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ open, onToggle }) => {
  const location = useLocation();

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: '📊' },
    { name: 'Health Monitor', href: '/health', icon: '💚' },
    { name: 'Tenants', href: '/tenants', icon: '👥' },
    { name: 'Configuration', href: '/config', icon: '⚙️' },
    { name: 'Audit Log', href: '/audit', icon: '📝' },
    { name: 'Diff Viewer', href: '/diff', icon: '🔍' },
    { name: 'Metrics', href: '/metrics', icon: '📈' },
    { name: 'Test Tools', href: '/test-tools', icon: '🧪' },
  ];

  const isActive = (href: string) => {
    return location.pathname === href;
  };

  return (
    <>
      {/* Mobile backdrop */}
      {open && (
        <div
          className="fixed inset-0 z-40 bg-gray-600 bg-opacity-75 lg:hidden"
          onClick={onToggle}
        />
      )}

      {/* Sidebar */}
      <motion.div
        initial={false}
        animate={{
          width: open ? (window.innerWidth < 1024 ? '16rem' : '16rem') : '0rem',
        }}
        transition={{ duration: 0.2, ease: 'easeInOut' }}
        className={`fixed inset-y-0 left-0 z-50 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 lg:relative lg:z-auto overflow-hidden`}
      >
        <div className="flex flex-col h-full">
          {/* Logo */}
          <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center space-x-2">
              <div className="text-2xl">🎯</div>
              <span className="text-xl font-bold text-gray-900 dark:text-white">
                Mimir Optimizer
              </span>
            </div>
            <button
              onClick={onToggle}
              className="lg:hidden p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 py-4 space-y-2 overflow-y-auto">
            {navigation.map((item) => {
              const active = isActive(item.href);
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  onClick={() => {
                    if (window.innerWidth < 1024) {
                      onToggle();
                    }
                  }}
                  className={`
                    flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors duration-200
                    ${active
                      ? 'bg-primary-100 text-primary-900 dark:bg-primary-900 dark:text-primary-100'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-700 dark:hover:text-white'
                    }
                  `}
                >
                  <span className="text-lg mr-3">{item.icon}</span>
                  {item.name}
                </Link>
              );
            })}
          </nav>

          {/* Footer */}
          <div className="p-4 border-t border-gray-200 dark:border-gray-700">
            <div className="text-xs text-gray-500 dark:text-gray-400">
              <div>Mimir Limit Optimizer</div>
              <div>v2.4.0</div>
            </div>
          </div>
        </div>
      </motion.div>
    </>
  );
};

export default Sidebar; 