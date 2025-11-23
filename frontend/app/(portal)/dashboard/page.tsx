'use client';

import { useEffect, useState } from 'react';
import { api, DashboardStats, Instance } from '@/lib/api';
import { Activity, Key, Database, ArrowUpRight } from 'lucide-react';

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [instances, setInstances] = useState<Instance[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      const [statsRes, instancesRes] = await Promise.all([
        api.getDashboardStats(),
        api.getInstances(),
      ]);

      if (statsRes.data) setStats(statsRes.data);
      if (instancesRes.data) setInstances(instancesRes.data.instances);
      setIsLoading(false);
    }
    loadData();
  }, []);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-primary">Dashboard</h1>
        <p className="text-neutral-dark-gray">Overview of your Savegress usage</p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Active Licenses"
          value={stats?.total_licenses || 0}
          icon={Key}
          color="text-primary"
        />
        <StatCard
          title="Running Instances"
          value={stats?.active_instances || 0}
          icon={Activity}
          color="text-green-600"
        />
        <StatCard
          title="Events (24h)"
          value={formatNumber(stats?.events_processed_24h || 0)}
          icon={Database}
          color="text-blue-600"
        />
        <StatCard
          title="Data Transferred (24h)"
          value={formatBytes(stats?.data_transferred_24h || 0)}
          icon={ArrowUpRight}
          color="text-accent-orange"
        />
      </div>

      {/* Active instances */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-primary">Active Instances</h2>
        </div>
        <div className="divide-y divide-gray-200">
          {instances.length === 0 ? (
            <div className="p-8 text-center text-neutral-dark-gray">
              <Activity className="w-12 h-12 mx-auto mb-3 text-gray-300" />
              <p>No active instances yet</p>
              <p className="text-sm mt-1">Deploy a CDC engine to see it here</p>
            </div>
          ) : (
            instances.map((instance) => (
              <div key={instance.id} className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium text-primary">{instance.hostname}</p>
                  <p className="text-sm text-neutral-dark-gray">
                    Version {instance.version} • {formatNumber(instance.events_processed)} events
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                      instance.status === 'online'
                        ? 'bg-green-100 text-green-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {instance.status}
                  </span>
                  <span className="text-xs text-neutral-dark-gray">
                    {formatRelativeTime(instance.last_seen_at)}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function StatCard({
  title,
  value,
  icon: Icon,
  color,
}: {
  title: string;
  value: string | number;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-neutral-dark-gray">{title}</span>
        <Icon className={`w-5 h-5 ${color}`} />
      </div>
      <p className="text-2xl font-bold text-primary">{value}</p>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-gray-200 rounded animate-pulse" />
        <div className="h-4 w-48 bg-gray-200 rounded animate-pulse mt-2" />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
            <div className="h-4 w-24 bg-gray-200 rounded animate-pulse mb-2" />
            <div className="h-8 w-16 bg-gray-200 rounded animate-pulse" />
          </div>
        ))}
      </div>
    </div>
  );
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`;
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
  return num.toString();
}

function formatBytes(bytes: number): string {
  if (bytes >= 1_000_000_000) return `${(bytes / 1_000_000_000).toFixed(1)} GB`;
  if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(1)} MB`;
  if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(1)} KB`;
  return `${bytes} B`;
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}
