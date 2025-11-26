'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, DashboardStats, Instance, Pipeline } from '@/lib/api';
import {
  Activity,
  Key,
  Database,
  ArrowUpRight,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  Clock,
  Plus,
  ChevronRight,
  Zap,
  DollarSign,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
} from 'recharts';

// Mock data for charts (replace with real API data)
const mockThroughputData = Array.from({ length: 24 }, (_, i) => ({
  hour: `${i}:00`,
  events: Math.floor(Math.random() * 50000) + 10000,
  bytes: Math.floor(Math.random() * 100000000) + 10000000,
}));

const mockCompressionData = {
  original: 1200,
  compressed: 47,
  ratio: 25.5,
};

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [instances, setInstances] = useState<Instance[]>([]);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      const [statsRes, instancesRes, pipelinesRes] = await Promise.all([
        api.getDashboardStats(),
        api.getInstances(),
        api.getPipelines(),
      ]);

      if (statsRes.data) setStats(statsRes.data);
      if (instancesRes.data) setInstances(instancesRes.data.instances);
      if (pipelinesRes.data) setPipelines(pipelinesRes.data.pipelines);
      setIsLoading(false);
    }
    loadData();
  }, []);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  const runningPipelines = pipelines.filter(p => p.status === 'running').length;
  const uptime = 99.9;
  const savings = Math.floor((stats?.data_transferred_24h || 0) * 0.02 / 100);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-h4 text-white">Dashboard</h1>
          <p className="text-content-1 text-grey">Overview of your Savegress usage</p>
        </div>
        <Link href="/pipelines" className="btn-primary px-5 py-3 text-sm">
          <Plus className="w-4 h-4 mr-2" />
          New Pipeline
        </Link>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Pipelines"
          value={runningPipelines}
          subtitle="running"
          icon={Activity}
          color="text-accent-cyan"
        />
        <StatCard
          title="Uptime"
          value={`${uptime}%`}
          subtitle="this month"
          icon={TrendingUp}
          color="text-accent-cyan"
        />
        <StatCard
          title="Data Today"
          value={formatBytes(stats?.data_transferred_24h || 0)}
          subtitle="compressed"
          icon={Database}
          color="text-accent-cyan"
        />
        <StatCard
          title="Savings"
          value={`$${savings}`}
          subtitle="this month"
          icon={DollarSign}
          color="text-accent-orange"
        />
      </div>

      {/* Active Pipelines */}
      <div className="card-dark">
        <div className="p-4 border-b border-cyan-40 flex items-center justify-between">
          <h2 className="text-h5 text-white">Active Pipelines</h2>
          <Link href="/pipelines" className="text-sm text-accent-cyan hover:text-accent-cyan-bright flex items-center gap-1 transition-colors">
            View all <ChevronRight className="w-4 h-4" />
          </Link>
        </div>
        {pipelines.length === 0 ? (
          <div className="p-8 text-center">
            <Activity className="w-12 h-12 mx-auto mb-3 text-grey opacity-50" />
            <p className="text-grey">No pipelines yet</p>
            <Link href="/pipelines" className="text-accent-cyan hover:text-accent-cyan-bright text-sm mt-1 inline-block transition-colors">
              Create your first pipeline
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-primary-dark/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Name</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Source</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Target</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Lag</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyan-40/30">
                {pipelines.slice(0, 5).map((pipeline) => (
                  <tr key={pipeline.id} className="hover:bg-primary-dark/30 transition-colors">
                    <td className="px-4 py-3">
                      <Link href={`/pipelines/${pipeline.id}`} className="font-medium text-white hover:text-accent-cyan transition-colors">
                        {pipeline.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-grey">
                      {pipeline.source_connection?.type || 'Unknown'}
                    </td>
                    <td className="px-4 py-3 text-sm text-grey">{pipeline.target_type}</td>
                    <td className="px-4 py-3 text-sm text-grey">{pipeline.current_lag_ms}ms</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={pipeline.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Throughput Chart */}
        <div className="card-dark p-6">
          <h3 className="text-h5 text-white mb-4">Throughput (24h)</h3>
          <div className="h-[250px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={mockThroughputData}>
                <defs>
                  <linearGradient id="colorEvents" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#00B4D8" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#00B4D8" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#1E3A5F" />
                <XAxis dataKey="hour" tick={{ fontSize: 12, fill: '#B2BBC9' }} tickLine={false} axisLine={{ stroke: '#1E3A5F' }} />
                <YAxis tick={{ fontSize: 12, fill: '#B2BBC9' }} tickLine={false} axisLine={{ stroke: '#1E3A5F' }} tickFormatter={(v) => formatNumber(v)} />
                <Tooltip
                  formatter={(value: number) => [formatNumber(value), 'Events']}
                  contentStyle={{
                    backgroundColor: '#0F2744',
                    border: '1px solid rgba(0, 180, 216, 0.4)',
                    borderRadius: '12px',
                    color: '#fff'
                  }}
                  labelStyle={{ color: '#B2BBC9' }}
                />
                <Area type="monotone" dataKey="events" stroke="#00B4D8" fillOpacity={1} fill="url(#colorEvents)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 flex items-center justify-between text-sm text-grey">
            <span>Avg: {formatNumber(35000)} events/hr</span>
            <span>Peak: {formatNumber(89000)} events/hr</span>
          </div>
        </div>

        {/* Compression Ratio */}
        <div className="card-dark p-6">
          <h3 className="text-h5 text-white mb-4">Compression Ratio</h3>
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-grey">Original Data</span>
                <span className="text-sm font-medium text-white">{mockCompressionData.original} GB</span>
              </div>
              <div className="w-full h-3 rounded-full bg-primary-dark">
                <div className="h-3 rounded-full bg-grey/50" style={{ width: '100%' }}></div>
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-grey">Compressed</span>
                <span className="text-sm font-medium text-accent-cyan">{mockCompressionData.compressed} GB</span>
              </div>
              <div className="w-full h-3 rounded-full bg-primary-dark">
                <div
                  className="h-3 rounded-full bg-gradient-to-r from-accent-cyan to-accent-blue"
                  style={{ width: `${(mockCompressionData.compressed / mockCompressionData.original) * 100}%` }}
                ></div>
              </div>
            </div>
            <div className="pt-6 border-t border-cyan-40/30">
              <div className="flex items-center justify-center gap-3">
                <Zap className="w-6 h-6 text-accent-orange" />
                <span className="text-3xl font-bold text-white">{mockCompressionData.ratio}x</span>
                <span className="text-grey">compression ratio</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Alerts */}
      <div className="card-dark">
        <div className="p-4 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Recent Alerts</h2>
        </div>
        <div className="divide-y divide-cyan-40/30">
          <AlertItem
            type="warning"
            message="High latency on prod-to-analytics (>5s)"
            time="2 hours ago"
          />
          <AlertItem
            type="success"
            message="Schema change auto-applied: users.phone"
            time="Yesterday"
          />
          <AlertItem
            type="info"
            message="New version v1.2.0 available"
            time="3 days ago"
          />
        </div>
      </div>

      {/* Active Instances */}
      {instances.length > 0 && (
        <div className="card-dark">
          <div className="p-4 border-b border-cyan-40">
            <h2 className="text-h5 text-white">Active Instances</h2>
          </div>
          <div className="divide-y divide-cyan-40/30">
            {instances.map((instance) => (
              <div key={instance.id} className="p-4 flex items-center justify-between hover:bg-primary-dark/30 transition-colors">
                <div>
                  <p className="font-medium text-white">{instance.hostname}</p>
                  <p className="text-sm text-grey">
                    Version {instance.version} • {formatNumber(instance.events_processed)} events
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge status={instance.status} />
                  <span className="text-xs text-grey">
                    {formatRelativeTime(instance.last_seen_at)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  color,
}: {
  title: string;
  value: string | number;
  subtitle: string;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="card-dark p-5">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm text-grey">{title}</span>
        <div className="p-2 rounded-lg bg-primary-dark">
          <Icon className={`w-5 h-5 ${color}`} />
        </div>
      </div>
      <p className="text-2xl font-bold text-white">{value}</p>
      <p className="text-sm text-grey mt-1">{subtitle}</p>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    running: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
    online: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
    paused: 'bg-accent-orange/20 text-accent-orange border-accent-orange/40',
    stopped: 'bg-grey/20 text-grey border-grey/40',
    offline: 'bg-grey/20 text-grey border-grey/40',
    error: 'bg-red-500/20 text-red-400 border-red-500/40',
    created: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
  };

  return (
    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${colors[status] || 'bg-grey/20 text-grey border-grey/40'}`}>
      <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${status === 'running' || status === 'online' ? 'bg-accent-cyan' : 'bg-grey'}`}></span>
      {status}
    </span>
  );
}

function AlertItem({ type, message, time }: { type: 'warning' | 'success' | 'info'; message: string; time: string }) {
  const icons = {
    warning: AlertTriangle,
    success: CheckCircle,
    info: Clock,
  };
  const colors = {
    warning: 'text-accent-orange',
    success: 'text-accent-cyan',
    info: 'text-accent-blue',
  };
  const Icon = icons[type];

  return (
    <div className="p-4 flex items-start gap-3 hover:bg-primary-dark/30 transition-colors">
      <Icon className={`w-5 h-5 ${colors[type]} mt-0.5`} />
      <div className="flex-1">
        <p className="text-sm text-white">{message}</p>
        <p className="text-xs text-grey mt-1">{time}</p>
      </div>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="card-dark p-5">
            <div className="h-4 w-24 bg-primary-dark rounded animate-pulse mb-3" />
            <div className="h-8 w-16 bg-primary-dark rounded animate-pulse" />
          </div>
        ))}
      </div>
      <div className="card-dark p-6">
        <div className="h-6 w-40 bg-primary-dark rounded animate-pulse mb-4" />
        <div className="h-[200px] bg-primary-dark/50 rounded animate-pulse" />
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
