'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import {
  LayoutDashboard,
  Key,
  CreditCard,
  Download,
  Settings,
  LogOut,
  ChevronDown,
} from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Licenses', href: '/licenses', icon: Key },
  { name: 'Billing', href: '/billing', icon: CreditCard },
  { name: 'Downloads', href: '/downloads', icon: Download },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <div className="flex flex-col h-full bg-white border-r border-gray-200 w-64">
      {/* Logo */}
      <div className="p-4 border-b border-gray-200">
        <Link href="/dashboard" className="text-xl font-bold text-primary">
          Savegress
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navigation.map((item) => {
          const isActive = pathname === item.href || pathname?.startsWith(item.href + '/');
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-primary text-white'
                  : 'text-neutral-dark-gray hover:bg-neutral-light-gray'
              }`}
            >
              <item.icon className="w-5 h-5" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* User menu */}
      <div className="p-4 border-t border-gray-200">
        <div className="flex items-center gap-3 mb-3">
          <div className="w-10 h-10 bg-primary rounded-full flex items-center justify-center text-white font-medium">
            {user?.name?.charAt(0).toUpperCase() || 'U'}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-primary truncate">{user?.name}</p>
            <p className="text-xs text-neutral-dark-gray truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm text-neutral-dark-gray hover:bg-neutral-light-gray rounded-md transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Sign out
        </button>
      </div>
    </div>
  );
}

export function MobileNav() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <div className="lg:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-50">
      <nav className="flex justify-around p-2">
        {navigation.slice(0, 4).map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex flex-col items-center gap-1 p-2 rounded-md ${
                isActive ? 'text-primary' : 'text-neutral-dark-gray'
              }`}
            >
              <item.icon className="w-5 h-5" />
              <span className="text-xs">{item.name}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
