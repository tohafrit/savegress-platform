'use client';

import Link from 'next/link';
import Image from 'next/image';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import {
  LayoutDashboard,
  Key,
  CreditCard,
  Download,
  Settings,
  LogOut,
  GitBranch,
  Database,
  Rocket,
  Book,
  Sliders,
  Activity,
  BookOpen,
  User,
} from 'lucide-react';

interface NavItem {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavGroup {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  items: NavItem[];
}

const navigationGroups: NavGroup[] = [
  {
    title: 'CDC',
    icon: Activity,
    items: [
      { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
      { name: 'Connections', href: '/connections', icon: Database },
      { name: 'Pipelines', href: '/pipelines', icon: GitBranch },
    ],
  },
  {
    title: 'Getting Started',
    icon: BookOpen,
    items: [
      { name: 'Setup', href: '/setup', icon: Rocket },
      { name: 'Optimizer', href: '/optimizer', icon: Sliders },
      { name: 'Docs', href: '/docs', icon: Book },
    ],
  },
  {
    title: 'Account',
    icon: User,
    items: [
      { name: 'Licenses', href: '/licenses', icon: Key },
      { name: 'Downloads', href: '/downloads', icon: Download },
      { name: 'Billing', href: '/billing', icon: CreditCard },
      { name: 'Settings', href: '/settings', icon: Settings },
    ],
  },
];

// Flat list for mobile nav (first 5 most important)
const mobileNavItems: NavItem[] = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Connections', href: '/connections', icon: Database },
  { name: 'Pipelines', href: '/pipelines', icon: GitBranch },
  { name: 'Setup', href: '/setup', icon: Rocket },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <div className="flex flex-col h-screen w-64 bg-dark-bg-card border-r border-cyan-40 sticky top-0">
      {/* Logo */}
      <div className="p-4 border-b border-cyan-40">
        <Link href="/dashboard" className="block">
          <Image
            src="/images/logo.svg"
            alt="Savegress"
            width={160}
            height={44}
            className="h-10 w-auto"
          />
        </Link>
      </div>

      {/* Navigation Groups */}
      <nav className="flex-1 p-4 space-y-6 overflow-y-auto">
        {navigationGroups.map((group) => (
          <div key={group.title}>
            {/* Group Header */}
            <div className="flex items-center gap-2 px-3 mb-2">
              <group.icon className="w-4 h-4 text-cyan-40" />
              <span className="text-xs font-semibold text-cyan-40 uppercase tracking-wider">
                {group.title}
              </span>
            </div>

            {/* Group Items */}
            <div className="space-y-1">
              {group.items.map((item) => {
                const isActive = pathname === item.href || pathname?.startsWith(item.href + '/');
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all ${
                      isActive
                        ? 'bg-gradient-btn-primary text-white shadow-glow-blue'
                        : 'text-grey hover:text-white hover:bg-primary-dark'
                    }`}
                  >
                    <item.icon className="w-4 h-4" />
                    {item.name}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* User menu */}
      <div className="p-4 border-t border-cyan-40">
        <div className="flex items-center gap-3 mb-3">
          <div className="w-10 h-10 bg-gradient-btn-primary rounded-full flex items-center justify-center text-white font-medium">
            {user?.name?.charAt(0).toUpperCase() || 'U'}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-white truncate">{user?.name}</p>
            <p className="text-xs text-grey truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm text-grey hover:text-white hover:bg-primary-dark rounded-lg transition-colors"
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

  return (
    <div className="lg:hidden fixed bottom-0 left-0 right-0 bg-dark-bg-card border-t border-cyan-40 z-50">
      <nav className="flex justify-around p-2">
        {mobileNavItems.map((item) => {
          const isActive = pathname === item.href || pathname?.startsWith(item.href + '/');
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex flex-col items-center gap-1 p-2 rounded-lg transition-colors ${
                isActive ? 'text-accent-cyan' : 'text-grey'
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
