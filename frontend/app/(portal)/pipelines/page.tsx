'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, Pipeline, Connection } from '@/lib/api';
import {
  Plus,
  Database,
  ArrowRight,
  Play,
  Pause,
  Square,
  AlertCircle,
  Clock,
  Activity,
  Trash2,
  Settings,
} from 'lucide-react';

export default function PipelinesPage() {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    const [pipelinesRes, connectionsRes] = await Promise.all([
      api.getPipelines(),
      api.getConnections(),
    ]);
    if (pipelinesRes.data) setPipelines(pipelinesRes.data.pipelines);
    if (connectionsRes.data) setConnections(connectionsRes.data.connections);
    setIsLoading(false);
  }

  async function deletePipeline(id: string) {
    if (!confirm('Are you sure you want to delete this pipeline?')) return;
    await api.deletePipeline(id);
    loadData();
  }

  if (isLoading) {
    return <PipelinesSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-h4 text-white">Pipelines</h1>
          <p className="text-content-1 text-grey">Manage your CDC replication pipelines</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn-primary px-5 py-3 text-sm"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Pipeline
        </button>
      </div>

      {pipelines.length === 0 ? (
        <EmptyState onCreateClick={() => setShowCreateModal(true)} />
      ) : (
        <div className="grid gap-4">
          {pipelines.map((pipeline) => (
            <PipelineCard
              key={pipeline.id}
              pipeline={pipeline}
              onDelete={() => deletePipeline(pipeline.id)}
            />
          ))}
        </div>
      )}

      {showCreateModal && (
        <CreatePipelineModal
          connections={connections}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            loadData();
          }}
        />
      )}
    </div>
  );
}

function PipelineCard({ pipeline, onDelete }: { pipeline: Pipeline; onDelete: () => void }) {
  const statusColors: Record<string, string> = {
    created: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
    running: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
    paused: 'bg-accent-orange/20 text-accent-orange border-accent-orange/40',
    stopped: 'bg-grey/20 text-grey border-grey/40',
    error: 'bg-red-500/20 text-red-400 border-red-500/40',
  };

  const statusIcons: Record<string, React.ElementType> = {
    created: Clock,
    running: Play,
    paused: Pause,
    stopped: Square,
    error: AlertCircle,
  };

  const StatusIcon = statusIcons[pipeline.status] || Clock;

  return (
    <div className="card-dark p-5">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <Link href={`/pipelines/${pipeline.id}`} className="text-h5 text-white hover:text-accent-cyan transition-colors">
              {pipeline.name}
            </Link>
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${statusColors[pipeline.status]}`}>
              <StatusIcon className="w-3 h-3" />
              {pipeline.status}
            </span>
          </div>
          {pipeline.description && (
            <p className="text-sm text-grey mt-1">{pipeline.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Link
            href={`/pipelines/${pipeline.id}`}
            className="p-2 text-grey hover:text-accent-cyan hover:bg-primary-dark rounded-lg transition-colors"
          >
            <Settings className="w-4 h-4" />
          </Link>
          <button
            onClick={onDelete}
            className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="mt-4 flex items-center gap-3 text-sm text-grey">
        <div className="flex items-center gap-1.5">
          <Database className="w-4 h-4 text-accent-cyan" />
          <span>{pipeline.source_connection?.type || 'Source'}</span>
        </div>
        <ArrowRight className="w-4 h-4 text-accent-cyan" />
        <div className="flex items-center gap-1.5">
          <Database className="w-4 h-4 text-accent-cyan" />
          <span>{pipeline.target_type}</span>
        </div>
      </div>

      <div className="mt-4 grid grid-cols-3 gap-4">
        <div>
          <p className="text-xs text-grey">Events Processed</p>
          <p className="text-lg font-semibold text-white">{formatNumber(pipeline.events_processed)}</p>
        </div>
        <div>
          <p className="text-xs text-grey">Data Transferred</p>
          <p className="text-lg font-semibold text-white">{formatBytes(pipeline.bytes_processed)}</p>
        </div>
        <div>
          <p className="text-xs text-grey">Current Lag</p>
          <p className="text-lg font-semibold text-white">{pipeline.current_lag_ms}ms</p>
        </div>
      </div>

      {pipeline.error_message && (
        <div className="mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
          <div className="flex items-start gap-2">
            <AlertCircle className="w-4 h-4 text-red-400 mt-0.5" />
            <p className="text-sm text-red-400">{pipeline.error_message}</p>
          </div>
        </div>
      )}
    </div>
  );
}

function EmptyState({ onCreateClick }: { onCreateClick: () => void }) {
  return (
    <div className="card-dark p-12 text-center">
      <Activity className="w-16 h-16 mx-auto mb-4 text-grey opacity-50" />
      <h3 className="text-h5 text-white mb-2">No pipelines yet</h3>
      <p className="text-grey mb-6 max-w-md mx-auto">
        Create your first CDC pipeline to start replicating data from your source database.
      </p>
      <button onClick={onCreateClick} className="btn-primary px-6 py-3">
        <Plus className="w-4 h-4 mr-2" />
        Create Pipeline
      </button>
    </div>
  );
}

function CreatePipelineModal({
  connections,
  onClose,
  onCreated,
}: {
  connections: Connection[];
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [sourceConnectionId, setSourceConnectionId] = useState('');
  const [targetType, setTargetType] = useState('http');
  const [targetUrl, setTargetUrl] = useState('');
  const [tables, setTables] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsSubmitting(true);
    setError('');

    const { error: apiError } = await api.createPipeline({
      name,
      description,
      source_connection_id: sourceConnectionId,
      target_type: targetType,
      target_config: targetUrl ? { url: targetUrl } : undefined,
      tables: tables.split(',').map(t => t.trim()).filter(Boolean),
    });

    if (apiError) {
      setError(apiError);
      setIsSubmitting(false);
      return;
    }

    onCreated();
  }

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="card-dark w-full max-w-lg mx-4">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Create New Pipeline</h2>
        </div>
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
              {error}
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
              placeholder="my-pipeline"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className="input-field min-h-[80px] py-3"
              placeholder="Optional description"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Source Connection</label>
            {connections.length === 0 ? (
              <div className="p-3 bg-accent-orange/10 border border-accent-orange/30 rounded-lg">
                <p className="text-sm text-accent-orange">
                  No connections available. <Link href="/connections" className="underline">Create one first</Link>.
                </p>
              </div>
            ) : (
              <select
                value={sourceConnectionId}
                onChange={(e) => setSourceConnectionId(e.target.value)}
                required
                className="input-field"
              >
                <option value="">Select a connection</option>
                {connections.map((conn) => (
                  <option key={conn.id} value={conn.id}>
                    {conn.name} ({conn.type})
                  </option>
                ))}
              </select>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Target Type</label>
            <select
              value={targetType}
              onChange={(e) => setTargetType(e.target.value)}
              className="input-field"
            >
              <option value="http">HTTP Webhook</option>
              <option value="kafka">Kafka</option>
              <option value="s3">S3</option>
              <option value="postgres">PostgreSQL</option>
              <option value="bigquery">BigQuery</option>
            </select>
          </div>

          {(targetType === 'http' || targetType === 'kafka' || targetType === 's3') && (
            <div>
              <label className="block text-sm font-medium text-grey mb-2">Target URL</label>
              <input
                type="text"
                value={targetUrl}
                onChange={(e) => setTargetUrl(e.target.value)}
                className="input-field"
                placeholder={targetType === 'http' ? 'https://your-endpoint.com/events' : targetType === 'kafka' ? 'kafka://localhost:9092' : 's3://bucket/prefix'}
              />
            </div>
          )}

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

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="btn-secondary px-5 py-2.5 text-sm"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || connections.length === 0}
              className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
            >
              {isSubmitting ? 'Creating...' : 'Create Pipeline'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function PipelinesSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
          <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
        </div>
        <div className="h-10 w-32 bg-primary-dark rounded-[20px] animate-pulse" />
      </div>
      {[1, 2, 3].map((i) => (
        <div key={i} className="card-dark p-5">
          <div className="h-6 w-48 bg-primary-dark rounded animate-pulse mb-4" />
          <div className="h-4 w-32 bg-primary-dark rounded animate-pulse" />
        </div>
      ))}
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
