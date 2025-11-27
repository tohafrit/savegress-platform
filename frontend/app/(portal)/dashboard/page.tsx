'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, DashboardStats, Instance, Pipeline } from '@/lib/api';
import {
  Activity,
  Database,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  Clock,
  Plus,
  ChevronRight,
  Zap,
  DollarSign,
  HelpCircle,
  Lightbulb,
  ArrowRight,
  BookOpen,
  Rocket,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import {
  PageHeader,
  MetricCard,
  InfoBanner,
  WelcomeBanner,
  QuickGuide,
  HelpIcon,
} from '@/components/ui/helpers';

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
  const [showWelcome, setShowWelcome] = useState(true);

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
  const isNewUser = pipelines.length === 0;

  return (
    <div className="space-y-6">
      {/* Welcome Banner for new users */}
      {isNewUser && showWelcome && (
        <WelcomeBanner
          onDismiss={() => setShowWelcome(false)}
          onGetStarted={() => window.location.href = '/setup'}
        />
      )}

      {/* Header */}
      <PageHeader
        title="Dashboard"
        description="Monitor your data replication at a glance. See how your pipelines are performing and track key metrics."
        tip={isNewUser ? "Start by creating a connection to your database" : undefined}
        action={
          <Link href="/pipelines" className="btn-primary px-5 py-3 text-sm">
            <Plus className="w-4 h-4 mr-2" />
            New Pipeline
          </Link>
        }
      />

      {/* Getting Started Guide for new users */}
      {isNewUser && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <Link href="/connections" className="card-dark p-5 hover:border-accent-cyan/50 transition-colors group">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-xl bg-accent-cyan/10 flex items-center justify-center group-hover:bg-accent-cyan/20 transition-colors">
                <Database className="w-6 h-6 text-accent-cyan" />
              </div>
              <div>
                <h3 className="font-medium text-white group-hover:text-accent-cyan transition-colors">1. Add Connection</h3>
                <p className="text-sm text-grey">Connect your source database</p>
              </div>
              <ArrowRight className="w-5 h-5 text-grey ml-auto group-hover:text-accent-cyan transition-colors" />
            </div>
          </Link>

          <Link href="/pipelines" className="card-dark p-5 hover:border-accent-cyan/50 transition-colors group">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-xl bg-accent-blue/10 flex items-center justify-center group-hover:bg-accent-blue/20 transition-colors">
                <Activity className="w-6 h-6 text-accent-blue" />
              </div>
              <div>
                <h3 className="font-medium text-white group-hover:text-accent-cyan transition-colors">2. Create Pipeline</h3>
                <p className="text-sm text-grey">Set up data replication</p>
              </div>
              <ArrowRight className="w-5 h-5 text-grey ml-auto group-hover:text-accent-cyan transition-colors" />
            </div>
          </Link>

          <Link href="/optimizer" className="card-dark p-5 hover:border-accent-cyan/50 transition-colors group">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-xl bg-accent-orange/10 flex items-center justify-center group-hover:bg-accent-orange/20 transition-colors">
                <Zap className="w-6 h-6 text-accent-orange" />
              </div>
              <div>
                <h3 className="font-medium text-white group-hover:text-accent-cyan transition-colors">3. Optimize</h3>
                <p className="text-sm text-grey">Fine-tune performance</p>
              </div>
              <ArrowRight className="w-5 h-5 text-grey ml-auto group-hover:text-accent-cyan transition-colors" />
            </div>
          </Link>
        </div>
      )}

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Active Pipelines"
          value={runningPipelines}
          subtitle="currently running"
          icon={Activity}
          help="Number of pipelines actively replicating data right now. Paused or stopped pipelines are not counted."
          color="text-accent-cyan"
        />
        <MetricCard
          title="System Uptime"
          value={`${uptime}%`}
          subtitle="this month"
          icon={TrendingUp}
          help="Percentage of time your pipelines were operational without interruption this month."
          color="text-accent-cyan"
        />
        <MetricCard
          title="Data Transferred"
          value={formatBytes(stats?.data_transferred_24h || 0)}
          subtitle="last 24 hours (compressed)"
          icon={Database}
          help="Total amount of data transferred in the last 24 hours. This is the compressed size, so actual source data may be 5-25x larger."
          color="text-accent-cyan"
        />
        <MetricCard
          title="Est. Savings"
          value={`$${savings}`}
          subtitle="bandwidth costs"
          icon={DollarSign}
          help="Estimated cost savings from compression. Based on average cloud egress pricing of $0.02/GB."
          color="text-accent-orange"
        />
      </div>

      {/* Active Pipelines */}
      <div className="card-dark">
        <div className="p-4 border-b border-cyan-40 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h2 className="text-h5 text-white">Active Pipelines</h2>
            <HelpIcon text="Pipelines replicate data from your source database to a destination in real-time. Each pipeline can track specific tables." />
          </div>
          <Link href="/pipelines" className="text-sm text-accent-cyan hover:text-accent-cyan-bright flex items-center gap-1 transition-colors">
            View all <ChevronRight className="w-4 h-4" />
          </Link>
        </div>
        {pipelines.length === 0 ? (
          <div className="p-8">
            <div className="text-center mb-6">
              <Activity className="w-12 h-12 mx-auto mb-3 text-grey opacity-50" />
              <h3 className="text-lg font-medium text-white mb-2">No pipelines yet</h3>
              <p className="text-grey max-w-md mx-auto">
                Pipelines stream changes from your database to any destination.
                Create your first one to start replicating data in real-time.
              </p>
            </div>
            <div className="flex justify-center">
              <Link href="/pipelines" className="btn-primary px-6 py-3">
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Pipeline
              </Link>
            </div>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-primary-dark/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                    <div className="flex items-center gap-1">
                      Name
                      <HelpIcon text="A friendly name to identify this pipeline" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                    <div className="flex items-center gap-1">
                      Source
                      <HelpIcon text="The database you're replicating from" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                    <div className="flex items-center gap-1">
                      Target
                      <HelpIcon text="Where the data is being sent" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                    <div className="flex items-center gap-1">
                      Lag
                      <HelpIcon text="How far behind real-time the replication is. Lower is better. Under 1 second is excellent." />
                    </div>
                  </th>
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
                    <td className="px-4 py-3 text-sm">
                      <span className={pipeline.current_lag_ms < 1000 ? 'text-accent-cyan' : pipeline.current_lag_ms < 5000 ? 'text-accent-orange' : 'text-red-400'}>
                        {pipeline.current_lag_ms}ms
                      </span>
                    </td>
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
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <h3 className="text-h5 text-white">Throughput (24h)</h3>
              <HelpIcon text="Number of database change events processed per hour. Higher throughput means more data is being replicated." />
            </div>
          </div>
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
          <div className="mt-4 flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <span className="text-grey">Avg:</span>
              <span className="text-white font-medium">{formatNumber(35000)} events/hr</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-grey">Peak:</span>
              <span className="text-accent-cyan font-medium">{formatNumber(89000)} events/hr</span>
            </div>
          </div>
        </div>

        {/* Compression Ratio */}
        <div className="card-dark p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <h3 className="text-h5 text-white">Compression Savings</h3>
              <HelpIcon text="How much data is saved through compression. A 25x ratio means you're only transferring 4% of the original data size." />
            </div>
          </div>
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-grey">Original Data (before compression)</span>
                <span className="text-sm font-medium text-white">{mockCompressionData.original} GB</span>
              </div>
              <div className="w-full h-3 rounded-full bg-primary-dark">
                <div className="h-3 rounded-full bg-grey/50" style={{ width: '100%' }}></div>
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-grey">Compressed Data (actually transferred)</span>
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
              <p className="text-center text-sm text-grey mt-2">
                You&apos;re saving {Math.round((1 - 1/mockCompressionData.ratio) * 100)}% on bandwidth costs
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Alerts with explanations */}
      <div className="card-dark">
        <div className="p-4 border-b border-cyan-40 flex items-center gap-2">
          <h2 className="text-h5 text-white">Recent Activity</h2>
          <HelpIcon text="Important events and notifications about your pipelines. We'll alert you about issues that need attention." />
        </div>
        <div className="divide-y divide-cyan-40/30">
          <AlertItem
            type="warning"
            message="High latency detected on prod-to-analytics pipeline"
            detail="Replication lag exceeded 5 seconds. This might be due to a large transaction or network issues."
            time="2 hours ago"
            action={{ label: 'View Pipeline', href: '/pipelines' }}
          />
          <AlertItem
            type="success"
            message="Schema change automatically applied"
            detail="New column 'phone' was added to the 'users' table and automatically propagated to the target."
            time="Yesterday"
          />
          <AlertItem
            type="info"
            message="New version available"
            detail="Savegress v1.2.0 is now available with improved compression and new Kafka sink."
            time="3 days ago"
            action={{ label: 'View Changelog', href: '/docs' }}
          />
        </div>
      </div>

      {/* Active Instances with help */}
      {instances.length > 0 && (
        <div className="card-dark">
          <div className="p-4 border-b border-cyan-40 flex items-center gap-2">
            <h2 className="text-h5 text-white">Active Instances</h2>
            <HelpIcon text="Savegress engine instances running your pipelines. Each instance can handle multiple pipelines." />
          </div>
          <div className="divide-y divide-cyan-40/30">
            {instances.map((instance) => (
              <div key={instance.id} className="p-4 flex items-center justify-between hover:bg-primary-dark/30 transition-colors">
                <div>
                  <p className="font-medium text-white">{instance.hostname}</p>
                  <p className="text-sm text-grey">
                    Version {instance.version} • Processed {formatNumber(instance.events_processed)} events
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge status={instance.status} />
                  <span className="text-xs text-grey">
                    Last seen {formatRelativeTime(instance.last_seen_at)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Help Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <InfoBanner
          type="tip"
          title="Pro tip: Use the Optimizer"
          action={{ label: 'Open Optimizer', href: '/optimizer' }}
        >
          Not sure which configuration to use? Our Configuration Optimizer will help you choose the best settings for your workload type.
        </InfoBanner>

        <InfoBanner
          type="info"
          title="Need help?"
          action={{ label: 'Read Docs', href: '/docs' }}
        >
          Check out our documentation for detailed guides on setting up connections, creating pipelines, and troubleshooting issues.
        </InfoBanner>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const configs: Record<string, { color: string; dot: string; label: string }> = {
    running: { color: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40', dot: 'bg-accent-cyan', label: 'Running' },
    online: { color: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40', dot: 'bg-accent-cyan', label: 'Online' },
    paused: { color: 'bg-accent-orange/20 text-accent-orange border-accent-orange/40', dot: 'bg-accent-orange', label: 'Paused' },
    stopped: { color: 'bg-grey/20 text-grey border-grey/40', dot: 'bg-grey', label: 'Stopped' },
    offline: { color: 'bg-grey/20 text-grey border-grey/40', dot: 'bg-grey', label: 'Offline' },
    error: { color: 'bg-red-500/20 text-red-400 border-red-500/40', dot: 'bg-red-400', label: 'Error' },
    created: { color: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40', dot: 'bg-accent-cyan', label: 'Created' },
  };

  const config = configs[status] || { color: 'bg-grey/20 text-grey border-grey/40', dot: 'bg-grey', label: status };

  return (
    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${config.color}`}>
      <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${config.dot}`}></span>
      {config.label}
    </span>
  );
}

function AlertItem({
  type,
  message,
  detail,
  time,
  action,
}: {
  type: 'warning' | 'success' | 'info';
  message: string;
  detail?: string;
  time: string;
  action?: { label: string; href: string };
}) {
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
      <Icon className={`w-5 h-5 ${colors[type]} mt-0.5 flex-shrink-0`} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-white">{message}</p>
        {detail && <p className="text-sm text-grey mt-0.5">{detail}</p>}
        <div className="flex items-center gap-3 mt-2">
          <span className="text-xs text-grey">{time}</span>
          {action && (
            <Link href={action.href} className="text-xs text-accent-cyan hover:text-accent-cyan-bright transition-colors">
              {action.label} →
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-64 bg-primary-dark rounded animate-pulse mt-2" />
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
