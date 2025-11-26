'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { api, Pipeline, PipelineLog, PipelineMetric } from '@/lib/api';
import {
  ArrowLeft,
  Play,
  Pause,
  Square,
  AlertCircle,
  Clock,
  Database,
  ArrowRight,
  Activity,
  Settings,
  FileText,
  BarChart3,
  Trash2,
  RefreshCw,
  CheckCircle,
  XCircle,
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

type Tab = 'overview' | 'metrics' | 'logs' | 'settings';

export default function PipelineDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [pipeline, setPipeline] = useState<Pipeline | null>(null);
  const [metrics, setMetrics] = useState<PipelineMetric[]>([]);
  const [logs, setLogs] = useState<PipelineLog[]>([]);
  const [activeTab, setActiveTab] = useState<Tab>('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    async function fetchData() {
      setIsLoading(true);
      const [pipelineRes, metricsRes, logsRes] = await Promise.all([
        api.getPipeline(id),
        api.getPipelineMetrics(id),
        api.getPipelineLogs(id),
      ]);

      if (pipelineRes.data) setPipeline(pipelineRes.data);
      if (metricsRes.data) setMetrics(metricsRes.data.metrics || []);
      if (logsRes.data) setLogs(logsRes.data.logs || []);
      setIsLoading(false);
    }
    fetchData();
  }, [id]);

  async function loadData() {
    setIsLoading(true);
    const [pipelineRes, metricsRes, logsRes] = await Promise.all([
      api.getPipeline(id),
      api.getPipelineMetrics(id),
      api.getPipelineLogs(id),
    ]);

    if (pipelineRes.data) setPipeline(pipelineRes.data);
    if (metricsRes.data) setMetrics(metricsRes.data.metrics || []);
    if (logsRes.data) setLogs(logsRes.data.logs || []);
    setIsLoading(false);
  }

  async function handleDelete() {
    if (!confirm('Are you sure you want to delete this pipeline? This action cannot be undone.')) return;
    setIsDeleting(true);
    await api.deletePipeline(id);
    router.push('/pipelines');
  }

  if (isLoading) {
    return <PipelineDetailSkeleton />;
  }

  if (!pipeline) {
    return (
      <div className="card-dark p-12 text-center">
        <AlertCircle className="w-12 h-12 mx-auto text-red-400 mb-4" />
        <h2 className="text-h5 text-white mb-2">Pipeline not found</h2>
        <Link href="/pipelines" className="text-accent-cyan hover:underline">
          Back to pipelines
        </Link>
      </div>
    );
  }

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'metrics', label: 'Metrics', icon: BarChart3 },
    { id: 'logs', label: 'Logs', icon: FileText },
    { id: 'settings', label: 'Settings', icon: Settings },
  ] as const;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <Link
              href="/pipelines"
              className="p-2 text-grey hover:text-accent-cyan hover:bg-primary-dark rounded-lg transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <h1 className="text-h4 text-white">{pipeline.name}</h1>
            <StatusBadge status={pipeline.status} />
          </div>
          {pipeline.description && (
            <p className="text-grey ml-11">{pipeline.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={loadData}
            className="p-2 text-grey hover:text-accent-cyan hover:bg-primary-dark rounded-lg transition-colors"
            title="Refresh"
          >
            <RefreshCw className="w-5 h-5" />
          </button>
          <button
            onClick={handleDelete}
            disabled={isDeleting}
            className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
            title="Delete pipeline"
          >
            <Trash2 className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Source → Target */}
      <div className="card-dark p-5">
        <div className="flex items-center justify-center gap-6">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-accent-cyan/20 rounded-lg">
              <Database className="w-6 h-6 text-accent-cyan" />
            </div>
            <div>
              <p className="text-sm text-grey">Source</p>
              <p className="font-medium text-white">
                {pipeline.source_connection?.name || 'Unknown'}
              </p>
              <p className="text-xs text-grey">
                {pipeline.source_connection?.type}
              </p>
            </div>
          </div>
          <ArrowRight className="w-6 h-6 text-accent-cyan" />
          <div className="flex items-center gap-3">
            <div className="p-3 bg-accent-cyan/20 rounded-lg">
              <Database className="w-6 h-6 text-accent-cyan" />
            </div>
            <div>
              <p className="text-sm text-grey">Target</p>
              <p className="font-medium text-white">{pipeline.target_type}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-cyan-40">
        <nav className="flex gap-1">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-accent-cyan text-accent-cyan'
                  : 'border-transparent text-grey hover:text-white'
              }`}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div>
        {activeTab === 'overview' && (
          <OverviewTab pipeline={pipeline} />
        )}
        {activeTab === 'metrics' && (
          <MetricsTab pipeline={pipeline} metrics={metrics} />
        )}
        {activeTab === 'logs' && (
          <LogsTab logs={logs} onRefresh={loadData} />
        )}
        {activeTab === 'settings' && (
          <SettingsTab pipeline={pipeline} onUpdate={loadData} />
        )}
      </div>
    </div>
  );
}

function OverviewTab({ pipeline }: { pipeline: Pipeline }) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Stats */}
      <div className="card-dark p-6">
        <h3 className="text-h5 text-white mb-4">Statistics</h3>
        <div className="grid grid-cols-2 gap-4">
          <StatItem
            label="Events Processed"
            value={formatNumber(pipeline.events_processed)}
          />
          <StatItem
            label="Data Transferred"
            value={formatBytes(pipeline.bytes_processed)}
          />
          <StatItem
            label="Current Lag"
            value={`${pipeline.current_lag_ms}ms`}
          />
          <StatItem
            label="Status"
            value={pipeline.status}
            valueClassName={
              pipeline.status === 'running' ? 'text-accent-cyan' :
              pipeline.status === 'error' ? 'text-red-400' :
              'text-grey'
            }
          />
        </div>
      </div>

      {/* Configuration */}
      <div className="card-dark p-6">
        <h3 className="text-h5 text-white mb-4">Configuration</h3>
        <div className="space-y-3">
          <ConfigItem label="Pipeline ID" value={pipeline.id} />
          <ConfigItem
            label="Tables"
            value={pipeline.tables?.length ? pipeline.tables.join(', ') : 'All tables'}
          />
          <ConfigItem
            label="Created"
            value={new Date(pipeline.created_at).toLocaleString()}
          />
          <ConfigItem
            label="Last Event"
            value={pipeline.last_event_at ? new Date(pipeline.last_event_at).toLocaleString() : 'Never'}
          />
        </div>
      </div>

      {/* Error Message */}
      {pipeline.error_message && (
        <div className="lg:col-span-2 bg-red-500/10 border border-red-500/30 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-400 mt-0.5" />
            <div>
              <h4 className="font-medium text-red-400">Error</h4>
              <p className="text-sm text-red-400/80 mt-1">{pipeline.error_message}</p>
            </div>
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div className="lg:col-span-2 card-dark p-6">
        <h3 className="text-h5 text-white mb-4">Quick Actions</h3>
        <div className="flex flex-wrap gap-3">
          <ActionButton
            icon={Play}
            label="Start"
            disabled={pipeline.status === 'running'}
            variant="success"
          />
          <ActionButton
            icon={Pause}
            label="Pause"
            disabled={pipeline.status !== 'running'}
            variant="warning"
          />
          <ActionButton
            icon={Square}
            label="Stop"
            disabled={pipeline.status === 'stopped'}
            variant="default"
          />
          <ActionButton
            icon={RefreshCw}
            label="Restart"
            variant="default"
          />
        </div>
      </div>
    </div>
  );
}

function MetricsTab({ pipeline, metrics }: { pipeline: Pipeline; metrics: PipelineMetric[] }) {
  // Generate mock data if no real metrics
  const chartData = metrics.length > 0
    ? metrics.map(m => ({
        time: new Date(m.timestamp).toLocaleTimeString(),
        events: m.events_per_second,
        bytes: m.bytes_per_second,
        latency: m.latency_ms,
      }))
    : Array.from({ length: 24 }, (_, i) => ({
        time: `${i}:00`,
        events: Math.floor(Math.random() * 5000) + 1000,
        bytes: Math.floor(Math.random() * 10000000) + 1000000,
        latency: Math.floor(Math.random() * 100) + 10,
      }));

  return (
    <div className="space-y-6">
      {/* Throughput Chart */}
      <div className="card-dark p-6">
        <h3 className="text-h5 text-white mb-4">Events Throughput</h3>
        <div className="h-[300px]">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="colorEvents" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#00B4D8" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#00B4D8" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a5f" />
              <XAxis dataKey="time" tick={{ fontSize: 12, fill: '#B2BBC9' }} stroke="#1e3a5f" />
              <YAxis tick={{ fontSize: 12, fill: '#B2BBC9' }} stroke="#1e3a5f" tickFormatter={(v) => formatNumber(v)} />
              <Tooltip
                formatter={(value: number) => [formatNumber(value), 'Events/s']}
                contentStyle={{
                  backgroundColor: '#0F2744',
                  borderColor: '#1e3a5f',
                  borderRadius: '8px',
                  color: '#fff'
                }}
                labelStyle={{ color: '#B2BBC9' }}
              />
              <Area
                type="monotone"
                dataKey="events"
                stroke="#00B4D8"
                strokeWidth={2}
                fillOpacity={1}
                fill="url(#colorEvents)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Latency Chart */}
      <div className="card-dark p-6">
        <h3 className="text-h5 text-white mb-4">Replication Latency</h3>
        <div className="h-[300px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a5f" />
              <XAxis dataKey="time" tick={{ fontSize: 12, fill: '#B2BBC9' }} stroke="#1e3a5f" />
              <YAxis tick={{ fontSize: 12, fill: '#B2BBC9' }} stroke="#1e3a5f" unit="ms" />
              <Tooltip
                formatter={(value: number) => [`${value}ms`, 'Latency']}
                contentStyle={{
                  backgroundColor: '#0F2744',
                  borderColor: '#1e3a5f',
                  borderRadius: '8px',
                  color: '#fff'
                }}
                labelStyle={{ color: '#B2BBC9' }}
              />
              <Line
                type="monotone"
                dataKey="latency"
                stroke="#FF6B35"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <MetricCard
          label="Avg Events/s"
          value={formatNumber(chartData.reduce((acc, d) => acc + d.events, 0) / chartData.length)}
        />
        <MetricCard
          label="Peak Events/s"
          value={formatNumber(Math.max(...chartData.map(d => d.events)))}
        />
        <MetricCard
          label="Avg Latency"
          value={`${Math.round(chartData.reduce((acc, d) => acc + d.latency, 0) / chartData.length)}ms`}
        />
        <MetricCard
          label="P99 Latency"
          value={`${Math.max(...chartData.map(d => d.latency))}ms`}
        />
      </div>
    </div>
  );
}

function LogsTab({ logs, onRefresh }: { logs: PipelineLog[]; onRefresh: () => void }) {
  const [filter, setFilter] = useState<string>('all');

  const filteredLogs = filter === 'all'
    ? logs
    : logs.filter(log => log.level === filter);

  const levelColors: Record<string, string> = {
    info: 'text-accent-cyan bg-accent-cyan/10 border-accent-cyan/30',
    warn: 'text-accent-orange bg-accent-orange/10 border-accent-orange/30',
    error: 'text-red-400 bg-red-500/10 border-red-500/30',
    debug: 'text-grey bg-grey/10 border-grey/30',
  };

  return (
    <div className="card-dark overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b border-cyan-40 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h3 className="text-h5 text-white">Pipeline Logs</h3>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="input-field py-1.5 px-3 text-sm w-auto"
          >
            <option value="all">All levels</option>
            <option value="error">Errors</option>
            <option value="warn">Warnings</option>
            <option value="info">Info</option>
            <option value="debug">Debug</option>
          </select>
        </div>
        <button
          onClick={onRefresh}
          className="inline-flex items-center gap-2 text-sm text-accent-cyan hover:text-accent-cyan/80"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Logs */}
      <div className="max-h-[500px] overflow-y-auto">
        {filteredLogs.length === 0 ? (
          <div className="p-12 text-center">
            <FileText className="w-12 h-12 mx-auto mb-3 text-grey opacity-50" />
            <p className="text-grey">No logs available</p>
          </div>
        ) : (
          <div className="divide-y divide-cyan-40/30">
            {filteredLogs.map((log) => (
              <div key={log.id} className="p-3 hover:bg-primary-dark/30 transition-colors">
                <div className="flex items-start gap-3">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium border ${levelColors[log.level] || 'text-grey bg-grey/10 border-grey/30'}`}>
                    {log.level.toUpperCase()}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-white font-mono">{log.message}</p>
                    {log.details && (
                      <pre className="text-xs text-grey mt-1 overflow-x-auto bg-primary-dark/50 p-2 rounded">
                        {typeof log.details === 'string' ? log.details : JSON.stringify(log.details, null, 2)}
                      </pre>
                    )}
                  </div>
                  <span className="text-xs text-grey whitespace-nowrap">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function SettingsTab({ pipeline, onUpdate }: { pipeline: Pipeline; onUpdate: () => void }) {
  const [name, setName] = useState(pipeline.name);
  const [description, setDescription] = useState(pipeline.description || '');
  const [tables, setTables] = useState(pipeline.tables?.join(', ') || '');
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setIsSaving(true);
    setMessage(null);

    const { error } = await api.updatePipeline(pipeline.id, {
      name,
      description,
      tables: tables.split(',').map(t => t.trim()).filter(Boolean),
    });

    if (error) {
      setMessage({ type: 'error', text: error });
    } else {
      setMessage({ type: 'success', text: 'Pipeline updated successfully' });
      onUpdate();
    }
    setIsSaving(false);
  }

  return (
    <div className="max-w-2xl">
      <form onSubmit={handleSave} className="card-dark p-6 space-y-6">
        <h3 className="text-h5 text-white">Pipeline Settings</h3>

        {message && (
          <div className={`p-3 rounded-lg flex items-center gap-2 ${
            message.type === 'success'
              ? 'bg-accent-cyan/10 border border-accent-cyan/30 text-accent-cyan'
              : 'bg-red-500/10 border border-red-500/30 text-red-400'
          }`}>
            {message.type === 'success' ? (
              <CheckCircle className="w-4 h-4" />
            ) : (
              <XCircle className="w-4 h-4" />
            )}
            {message.text}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-grey mb-2">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="input-field"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-grey mb-2">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="input-field min-h-[80px] py-3"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-grey mb-2">Tables (comma-separated)</label>
          <input
            type="text"
            value={tables}
            onChange={(e) => setTables(e.target.value)}
            className="input-field"
            placeholder="public.users, public.orders"
          />
          <p className="text-xs text-grey mt-1">Leave empty to replicate all tables</p>
        </div>

        <div className="pt-4 border-t border-cyan-40/30">
          <button
            type="submit"
            disabled={isSaving}
            className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>

      {/* Danger Zone */}
      <div className="mt-6 bg-red-500/10 border border-red-500/30 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-red-400 mb-2">Danger Zone</h3>
        <p className="text-sm text-red-400/80 mb-4">
          Once you delete a pipeline, there is no going back. Please be certain.
        </p>
        <Link
          href="/pipelines"
          className="inline-flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
        >
          <Trash2 className="w-4 h-4" />
          Delete Pipeline
        </Link>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    running: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
    paused: 'bg-accent-orange/20 text-accent-orange border-accent-orange/40',
    stopped: 'bg-grey/20 text-grey border-grey/40',
    error: 'bg-red-500/20 text-red-400 border-red-500/40',
    created: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
  };

  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${colors[status] || 'bg-grey/20 text-grey border-grey/40'}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${status === 'running' ? 'bg-accent-cyan animate-pulse' : 'bg-grey'}`}></span>
      {status}
    </span>
  );
}

function StatItem({ label, value, valueClassName = '' }: { label: string; value: string | number; valueClassName?: string }) {
  return (
    <div>
      <p className="text-sm text-grey">{label}</p>
      <p className={`text-xl font-semibold text-white ${valueClassName}`}>{value}</p>
    </div>
  );
}

function ConfigItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-grey">{label}</span>
      <span className="text-sm font-medium text-white">{value}</span>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="card-dark p-4">
      <p className="text-sm text-grey">{label}</p>
      <p className="text-2xl font-bold text-white mt-1">{value}</p>
    </div>
  );
}

function ActionButton({
  icon: Icon,
  label,
  disabled = false,
  variant = 'default',
}: {
  icon: React.ElementType;
  label: string;
  disabled?: boolean;
  variant?: 'default' | 'success' | 'warning' | 'danger';
}) {
  const variants = {
    default: 'border-cyan-40 text-grey hover:text-white hover:bg-primary-dark',
    success: 'border-accent-cyan/40 text-accent-cyan hover:bg-accent-cyan/10',
    warning: 'border-accent-orange/40 text-accent-orange hover:bg-accent-orange/10',
    danger: 'border-red-500/40 text-red-400 hover:bg-red-500/10',
  };

  return (
    <button
      disabled={disabled}
      className={`inline-flex items-center gap-2 px-4 py-2 border rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${variants[variant]}`}
    >
      <Icon className="w-4 h-4" />
      {label}
    </button>
  );
}

function PipelineDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="h-8 w-8 bg-primary-dark rounded animate-pulse" />
        <div className="h-8 w-48 bg-primary-dark rounded animate-pulse" />
      </div>
      <div className="card-dark p-4">
        <div className="h-16 bg-primary-dark/50 rounded animate-pulse" />
      </div>
      <div className="flex gap-4 border-b border-cyan-40 pb-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-6 w-24 bg-primary-dark rounded animate-pulse" />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-6">
        <div className="card-dark p-6">
          <div className="h-6 w-32 bg-primary-dark rounded animate-pulse mb-4" />
          <div className="space-y-3">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-10 bg-primary-dark/50 rounded animate-pulse" />
            ))}
          </div>
        </div>
        <div className="card-dark p-6">
          <div className="h-6 w-32 bg-primary-dark rounded animate-pulse mb-4" />
          <div className="space-y-3">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-6 bg-primary-dark/50 rounded animate-pulse" />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`;
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
  return Math.round(num).toString();
}

function formatBytes(bytes: number): string {
  if (bytes >= 1_000_000_000) return `${(bytes / 1_000_000_000).toFixed(1)} GB`;
  if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(1)} MB`;
  if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(1)} KB`;
  return `${bytes} B`;
}
