import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { SyncIndicator } from '@/components/Sync/SyncIndicator';
import { ThemeToggle } from '@/components/Layout/ThemeToggle';
import { BottomNav } from '@/components/Layout/BottomNav';

export const MainLayout: React.FC = () => {
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    window.location.href = '/login';
  };

  const navItems = [
    { to: '/feed', label: 'Feed' },
    { to: '/program', label: 'Program' },
    { to: '/analytics', label: 'Analytics' },
    { to: '/exercises', label: 'Exercises' },
    { to: '/history', label: 'History' },
    { to: '/comp-prep', label: 'Comp Prep' },
    { to: '/coaches', label: 'Coaches' },
    { to: '/relationships', label: 'My Coaches' },
    { to: '/dm', label: 'Messages' },
    { to: '/tools', label: 'Tools' },
  ];

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 pb-16 md:pb-0">
      <header className="bg-white dark:bg-gray-800 shadow-sm border-b dark:border-gray-700 sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex justify-between items-center py-3 md:py-4">
            <h1 className="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">
              Powerlifting Coach
            </h1>
            <div className="flex items-center gap-2 md:gap-4">
              <ThemeToggle />
              <span className="hidden sm:inline text-sm text-gray-600 dark:text-gray-300">
                {user?.name || user?.email}
              </span>
              <button
                onClick={handleLogout}
                className="text-sm text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white px-2 md:px-0"
              >
                Logout
              </button>
            </div>
          </div>
          <nav className="hidden md:flex gap-4 lg:gap-8 border-t dark:border-gray-700 pt-2 overflow-x-auto">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `pb-2 px-2 text-sm font-medium transition-colors whitespace-nowrap ${
                    isActive
                      ? 'border-b-2 border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>
      <SyncIndicator />
      <main className="py-4 md:py-6 min-h-[calc(100vh-4rem)]">
        <Outlet />
      </main>
      <BottomNav />
    </div>
  );
};
