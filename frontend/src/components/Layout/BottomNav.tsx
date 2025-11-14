import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  HomeIcon,
  ChartBarIcon,
  UserGroupIcon,
  TrophyIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline';
import {
  HomeIcon as HomeIconSolid,
  ChartBarIcon as ChartBarIconSolid,
  UserGroupIcon as UserGroupIconSolid,
  TrophyIcon as TrophyIconSolid,
  Cog6ToothIcon as Cog6ToothIconSolid,
} from '@heroicons/react/24/solid';

interface NavItem {
  path: string;
  label: string;
  icon: React.ForwardRefExoticComponent<any>;
  iconSolid: React.ForwardRefExoticComponent<any>;
}

const navItems: NavItem[] = [
  { path: '/feed', label: 'Home', icon: HomeIcon, iconSolid: HomeIconSolid },
  { path: '/analytics', label: 'Stats', icon: ChartBarIcon, iconSolid: ChartBarIconSolid },
  { path: '/program', label: 'Train', icon: TrophyIconSolid, iconSolid: TrophyIconSolid },
  { path: '/coaches', label: 'Coaches', icon: UserGroupIcon, iconSolid: UserGroupIconSolid },
  { path: '/tools', label: 'More', icon: Cog6ToothIcon, iconSolid: Cog6ToothIconSolid },
];

export const BottomNav: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 safe-area-bottom z-40">
      <div className="flex justify-around items-center h-16">
        {navItems.map((item) => {
          const active = isActive(item.path);
          const Icon = active ? item.iconSolid : item.icon;

          return (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className={`flex flex-col items-center justify-center flex-1 h-full transition-colors ${
                active
                  ? 'text-blue-600 dark:text-blue-400'
                  : 'text-gray-600 dark:text-gray-400 active:text-blue-600 dark:active:text-blue-400'
              }`}
              aria-label={item.label}
            >
              <Icon className="w-6 h-6" />
              <span className="text-xs mt-1 font-medium">{item.label}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
};
