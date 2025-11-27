'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, Pipeline, Connection } from '@/lib/api';
import {
  PageHeader,
  InfoBanner,
  EmptyStateWithGuide,
  FormFieldWithHelp,
  HelpIcon,
  QuickGuide,
  ExpandableSection,
} from '@/components/ui/helpers';
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
  Zap,
  RefreshCw,
  Globe,
  Server,
  Cloud,
  Table2,
  Info,
  CheckCircle2,
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
    if (!confirm('Are you sure you want to delete this pipeline? This will stop all data replication for this pipeline.')) return;
    await api.deletePipeline(id);
    loadData();
  }

  if (isLoading) {
    return <PipelinesSkeleton />;
  }

  const hasNoPipelines = pipelines.length === 0;
  const hasNoConnections = connections.length === 0;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Pipelines"
        description="A pipeline connects your source database to a destination and streams changes in real-time. Each pipeline monitors specific tables and delivers every INSERT, UPDATE, and DELETE event."
        tip="Pipelines are the heart of CDC - they capture every database change as it happens!"
        action={
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn-primary px-5 py-3 text-sm"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Pipeline
          </button>
        }
      />

      {/* Helpful guide for users without pipelines */}
      {hasNoPipelines && !hasNoConnections && (
        <InfoBanner type="tip" title="Ready to create your first pipeline?">
          You have {connections.length} connection{connections.length > 1 ? 's' : ''} ready to use!
          A pipeline will start capturing changes from your database the moment you create it.
          Click &quot;New Pipeline&quot; above to get started.
        </InfoBanner>
      )}

      {/* Warning if no connections yet */}
      {hasNoConnections && (
        <InfoBanner
          type="warning"
          title="Set up a connection first"
          action={{ label: 'Create Connection', href: '/connections' }}
        >
          Before creating a pipeline, you need to connect to a source database.
          Go to Connections page to add your PostgreSQL database.
        </InfoBanner>
      )}

      {pipelines.length === 0 ? (
        <EmptyStateWithGuide
          icon={Activity}
          title="No pipelines yet"
          description="Create your first CDC pipeline to start replicating data from your source database in real-time."
          guide={{
            title: 'How Pipelines Work',
            steps: [
              'Pipeline connects to your source database (PostgreSQL)',
              'It reads the Write-Ahead Log (WAL) to capture every change',
              'Changes are transformed and sent to your destination (webhook, Kafka, etc.)',
              'You can monitor events, lag, and errors in real-time',
            ],
          }}
          action={
            hasNoConnections
              ? undefined
              : { label: 'Create Pipeline', onClick: () => setShowCreateModal(true) }
          }
        />
      ) : (
        <>
          {/* Quick status summary */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatusSummaryCard
              label="Total Pipelines"
              value={pipelines.length}
              icon={Activity}
              help="Number of pipelines you've created"
            />
            <StatusSummaryCard
              label="Running"
              value={pipelines.filter(p => p.status === 'running').length}
              icon={Play}
              color="text-green-400"
              help="Pipelines actively streaming data"
            />
            <StatusSummaryCard
              label="Paused"
              value={pipelines.filter(p => p.status === 'paused').length}
              icon={Pause}
              color="text-accent-orange"
              help="Pipelines temporarily stopped (can be resumed)"
            />
            <StatusSummaryCard
              label="Errors"
              value={pipelines.filter(p => p.status === 'error').length}
              icon={AlertCircle}
              color="text-red-400"
              help="Pipelines with issues that need attention"
            />
          </div>

          {/* Pipeline cards */}
          <div className="grid gap-4">
            {pipelines.map((pipeline) => (
              <PipelineCard
                key={pipeline.id}
                pipeline={pipeline}
                onDelete={() => deletePipeline(pipeline.id)}
              />
            ))}
          </div>

          {/* Helpful tips at the bottom */}
          <div className="grid md:grid-cols-3 gap-4">
            <InfoBanner type="tip" title="Performance tip" dismissible>
              Keep your pipeline lag low by ensuring your destination can handle the throughput.
              High lag means changes are queueing up faster than they can be delivered.
            </InfoBanner>
            <InfoBanner type="info" title="Did you know?" dismissible>
              You can pause a pipeline without losing data. When you resume,
              it will continue from exactly where it left off.
            </InfoBanner>
            <InfoBanner
              type="tip"
              title="Need help configuring?"
              action={{ label: 'Open Optimizer', href: '/optimizer' }}
              dismissible
            >
              Use our Configuration Optimizer to generate the best settings for your workload.
            </InfoBanner>
          </div>
        </>
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

function StatusSummaryCard({
  label,
  value,
  icon: Icon,
  color = 'text-accent-cyan',
  help
}: {
  label: string;
  value: number;
  icon: React.ElementType;
  color?: string;
  help: string;
}) {
  return (
    <div className="card-dark p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm text-grey">{label}</span>
          <HelpIcon text={help} />
        </div>
        <Icon className={`w-4 h-4 ${color}`} />
      </div>
      <p className="text-2xl font-bold text-white mt-2">{value}</p>
    </div>
  );
}

function PipelineCard({ pipeline, onDelete }: { pipeline: Pipeline; onDelete: () => void }) {
  const statusConfig: Record<string, {
    colors: string;
    icon: React.ElementType;
    description: string
  }> = {
    created: {
      colors: 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40',
      icon: Clock,
      description: 'Pipeline is created but not yet started',
    },
    running: {
      colors: 'bg-green-500/20 text-green-400 border-green-500/40',
      icon: Play,
      description: 'Actively streaming changes in real-time',
    },
    paused: {
      colors: 'bg-accent-orange/20 text-accent-orange border-accent-orange/40',
      icon: Pause,
      description: 'Temporarily paused - can be resumed anytime',
    },
    stopped: {
      colors: 'bg-grey/20 text-grey border-grey/40',
      icon: Square,
      description: 'Pipeline has been stopped',
    },
    error: {
      colors: 'bg-red-500/20 text-red-400 border-red-500/40',
      icon: AlertCircle,
      description: 'Something went wrong - check the error message',
    },
  };

  const status = statusConfig[pipeline.status] || statusConfig.created;
  const StatusIcon = status.icon;

  return (
    <div className="card-dark p-5">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <Link href={`/pipelines/${pipeline.id}`} className="text-h5 text-white hover:text-accent-cyan transition-colors">
              {pipeline.name}
            </Link>
            <div className="relative group">
              <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border cursor-help ${status.colors}`}>
                <StatusIcon className="w-3 h-3" />
                {pipeline.status}
              </span>
              {/* Status tooltip */}
              <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-[#0a1628] border border-cyan-40 rounded-lg text-sm text-grey whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity z-10">
                {status.description}
              </div>
            </div>
          </div>
          {pipeline.description && (
            <p className="text-sm text-grey mt-1">{pipeline.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Link
            href={`/pipelines/${pipeline.id}`}
            className="p-2 text-grey hover:text-accent-cyan hover:bg-primary-dark rounded-lg transition-colors"
            title="Configure pipeline"
          >
            <Settings className="w-4 h-4" />
          </Link>
          <button
            onClick={onDelete}
            className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
            title="Delete pipeline"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Data flow visualization */}
      <div className="mt-4 p-3 bg-primary-dark/50 rounded-lg">
        <div className="flex items-center justify-center gap-3 text-sm">
          <div className="flex items-center gap-2 px-3 py-2 bg-primary-dark rounded-lg border border-cyan-40/30">
            <Database className="w-4 h-4 text-accent-cyan" />
            <span className="text-white">{pipeline.source_connection?.name || pipeline.source_connection?.type || 'Source'}</span>
          </div>
          <div className="flex items-center gap-1 text-accent-cyan">
            <div className="w-8 h-px bg-accent-cyan/50" />
            <Zap className="w-4 h-4" />
            <div className="w-8 h-px bg-accent-cyan/50" />
          </div>
          <div className="flex items-center gap-2 px-3 py-2 bg-primary-dark rounded-lg border border-cyan-40/30">
            {getTargetIcon(pipeline.target_type)}
            <span className="text-white">{formatTargetType(pipeline.target_type)}</span>
          </div>
        </div>
      </div>

      {/* Metrics with explanations */}
      <div className="mt-4 grid grid-cols-3 gap-4">
        <MetricItem
          label="Events Processed"
          value={formatNumber(pipeline.events_processed)}
          help="Total number of database changes (INSERTs, UPDATEs, DELETEs) captured and delivered"
        />
        <MetricItem
          label="Data Transferred"
          value={formatBytes(pipeline.bytes_processed)}
          help="Total amount of data sent to your destination"
        />
        <MetricItem
          label="Current Lag"
          value={`${pipeline.current_lag_ms}ms`}
          help="Time between a change happening in your database and being delivered. Lower is better!"
          highlight={pipeline.current_lag_ms > 1000}
        />
      </div>

      {pipeline.error_message && (
        <div className="mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
          <div className="flex items-start gap-2">
            <AlertCircle className="w-4 h-4 text-red-400 mt-0.5 flex-shrink-0" />
            <div>
              <p className="text-sm font-medium text-red-400">Error encountered</p>
              <p className="text-sm text-red-300 mt-1">{pipeline.error_message}</p>
              <p className="text-xs text-grey mt-2">
                Check your connection settings and destination availability.
                <Link href={`/pipelines/${pipeline.id}`} className="text-accent-cyan ml-1 hover:underline">
                  View details →
                </Link>
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function MetricItem({
  label,
  value,
  help,
  highlight = false
}: {
  label: string;
  value: string;
  help: string;
  highlight?: boolean;
}) {
  return (
    <div>
      <div className="flex items-center gap-1.5 mb-1">
        <p className="text-xs text-grey">{label}</p>
        <HelpIcon text={help} />
      </div>
      <p className={`text-lg font-semibold ${highlight ? 'text-accent-orange' : 'text-white'}`}>
        {value}
      </p>
    </div>
  );
}

function getTargetIcon(targetType: string) {
  switch (targetType) {
    case 'http':
      return <Globe className="w-4 h-4 text-accent-cyan" />;
    case 'kafka':
      return <Server className="w-4 h-4 text-accent-cyan" />;
    case 's3':
    case 'bigquery':
      return <Cloud className="w-4 h-4 text-accent-cyan" />;
    default:
      return <Database className="w-4 h-4 text-accent-cyan" />;
  }
}

function formatTargetType(targetType: string): string {
  const names: Record<string, string> = {
    http: 'HTTP Webhook',
    kafka: 'Apache Kafka',
    s3: 'Amazon S3',
    bigquery: 'BigQuery',
    postgres: 'PostgreSQL',
  };
  return names[targetType] || targetType;
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
  const [step, setStep] = useState(1);

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

  const targetTypeDescriptions: Record<string, { description: string; placeholder: string; example: string }> = {
    http: {
      description: 'Send events as HTTP POST requests to your endpoint. Great for webhooks and custom integrations.',
      placeholder: 'https://your-api.com/events',
      example: 'Each event will be sent as a JSON POST request with the changed data.',
    },
    kafka: {
      description: 'Stream events to Apache Kafka topics. Perfect for event-driven architectures.',
      placeholder: 'kafka://localhost:9092',
      example: 'Events will be published to a topic named after your pipeline.',
    },
    s3: {
      description: 'Store events in Amazon S3 as JSON files. Ideal for data lakes and archival.',
      placeholder: 's3://your-bucket/prefix',
      example: 'Events are batched and written as JSON files in your bucket.',
    },
    postgres: {
      description: 'Replicate to another PostgreSQL database. Great for read replicas and migrations.',
      placeholder: 'postgres://user:pass@host:5432/db',
      example: 'Tables will be created automatically in the target database.',
    },
    bigquery: {
      description: 'Stream events directly to Google BigQuery. Perfect for analytics.',
      placeholder: 'bigquery://project.dataset',
      example: 'Data will be streamed to BigQuery tables in real-time.',
    },
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="card-dark w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Create New Pipeline</h2>
          <p className="text-sm text-grey mt-1">
            Set up a data flow from your source database to a destination
          </p>
        </div>

        {/* Progress indicator */}
        <div className="px-5 pt-4">
          <div className="flex items-center gap-2">
            <StepBadge number={1} active={step === 1} completed={step > 1} label="Basics" />
            <div className="flex-1 h-px bg-cyan-40/30" />
            <StepBadge number={2} active={step === 2} completed={step > 2} label="Destination" />
            <div className="flex-1 h-px bg-cyan-40/30" />
            <StepBadge number={3} active={step === 3} completed={false} label="Tables" />
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-5">
          {error && (
            <InfoBanner type="warning" title="Something went wrong">
              {error}
            </InfoBanner>
          )}

          {step === 1 && (
            <div className="space-y-4">
              <FormFieldWithHelp
                label="Pipeline Name"
                help="A unique name to identify this pipeline. Use something descriptive!"
                tip="Example: 'orders-to-warehouse' or 'users-sync-production'"
                required
              >
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  className="input-field"
                  placeholder="my-pipeline"
                />
              </FormFieldWithHelp>

              <FormFieldWithHelp
                label="Description"
                help="Optional notes about what this pipeline does"
                tip="Help your future self remember why you created this pipeline"
              >
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={2}
                  className="input-field min-h-[80px] py-3"
                  placeholder="Sync user orders to data warehouse for analytics"
                />
              </FormFieldWithHelp>

              <FormFieldWithHelp
                label="Source Connection"
                help="The database to capture changes from"
                required
              >
                {connections.length === 0 ? (
                  <InfoBanner type="warning">
                    <p>No connections available.</p>
                    <Link href="/connections" className="text-accent-orange underline">
                      Create a connection first →
                    </Link>
                  </InfoBanner>
                ) : (
                  <select
                    value={sourceConnectionId}
                    onChange={(e) => setSourceConnectionId(e.target.value)}
                    required
                    className="input-field"
                  >
                    <option value="">Select your source database...</option>
                    {connections.map((conn) => (
                      <option key={conn.id} value={conn.id}>
                        {conn.name} ({conn.type})
                      </option>
                    ))}
                  </select>
                )}
              </FormFieldWithHelp>

              <div className="flex justify-end pt-4">
                <button
                  type="button"
                  onClick={() => setStep(2)}
                  disabled={!name || !sourceConnectionId}
                  className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
                >
                  Next: Choose Destination
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <FormFieldWithHelp
                label="Destination Type"
                help="Where should the captured changes be sent?"
                required
              >
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {[
                    { value: 'http', label: 'HTTP Webhook', icon: Globe },
                    { value: 'kafka', label: 'Kafka', icon: Server },
                    { value: 's3', label: 'Amazon S3', icon: Cloud },
                    { value: 'postgres', label: 'PostgreSQL', icon: Database },
                    { value: 'bigquery', label: 'BigQuery', icon: Cloud },
                  ].map((option) => (
                    <button
                      key={option.value}
                      type="button"
                      onClick={() => setTargetType(option.value)}
                      className={`p-3 rounded-lg border transition-colors text-left ${
                        targetType === option.value
                          ? 'bg-accent-cyan/10 border-accent-cyan text-white'
                          : 'bg-primary-dark/50 border-cyan-40/30 text-grey hover:border-cyan-40'
                      }`}
                    >
                      <option.icon className={`w-5 h-5 mb-2 ${targetType === option.value ? 'text-accent-cyan' : ''}`} />
                      <span className="text-sm font-medium">{option.label}</span>
                    </button>
                  ))}
                </div>
              </FormFieldWithHelp>

              {/* Target type description */}
              <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                <div className="flex items-start gap-3">
                  <Info className="w-5 h-5 text-accent-cyan flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm text-grey">{targetTypeDescriptions[targetType].description}</p>
                    <p className="text-xs text-grey/70 mt-2">{targetTypeDescriptions[targetType].example}</p>
                  </div>
                </div>
              </div>

              {(targetType === 'http' || targetType === 'kafka' || targetType === 's3') && (
                <FormFieldWithHelp
                  label="Destination URL"
                  help="The endpoint or address where events will be sent"
                  tip={`Format: ${targetTypeDescriptions[targetType].placeholder}`}
                >
                  <input
                    type="text"
                    value={targetUrl}
                    onChange={(e) => setTargetUrl(e.target.value)}
                    className="input-field"
                    placeholder={targetTypeDescriptions[targetType].placeholder}
                  />
                </FormFieldWithHelp>
              )}

              <div className="flex justify-between pt-4">
                <button
                  type="button"
                  onClick={() => setStep(1)}
                  className="btn-secondary px-5 py-2.5 text-sm"
                >
                  Back
                </button>
                <button
                  type="button"
                  onClick={() => setStep(3)}
                  className="btn-primary px-5 py-2.5 text-sm"
                >
                  Next: Select Tables
                </button>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <FormFieldWithHelp
                label="Tables to Replicate"
                help="Specify which tables to capture changes from, or leave empty for all tables"
                tip="Use format: schema.table (e.g., public.users, public.orders)"
              >
                <input
                  type="text"
                  value={tables}
                  onChange={(e) => setTables(e.target.value)}
                  className="input-field"
                  placeholder="public.users, public.orders"
                />
              </FormFieldWithHelp>

              <ExpandableSection title="How table selection works" icon={Table2}>
                <div className="space-y-3 text-sm text-grey">
                  <p>
                    <strong className="text-white">Leave empty</strong> to replicate all tables in your database.
                    This is the simplest option for full replication.
                  </p>
                  <p>
                    <strong className="text-white">Specify tables</strong> as a comma-separated list using
                    the format <code className="text-accent-cyan">schema.table</code>.
                  </p>
                  <p>
                    <strong className="text-white">Examples:</strong>
                  </p>
                  <ul className="list-disc list-inside space-y-1 text-grey">
                    <li><code className="text-accent-cyan">public.users</code> - just the users table</li>
                    <li><code className="text-accent-cyan">public.orders, public.order_items</code> - multiple tables</li>
                    <li><code className="text-accent-cyan">analytics.*</code> - all tables in analytics schema</li>
                  </ul>
                </div>
              </ExpandableSection>

              {/* Summary */}
              <div className="p-4 bg-gradient-to-br from-accent-cyan/5 to-accent-blue/5 border border-cyan-40/30 rounded-xl">
                <h4 className="font-medium text-white mb-3 flex items-center gap-2">
                  <CheckCircle2 className="w-5 h-5 text-accent-cyan" />
                  Pipeline Summary
                </h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-grey">Name:</span>
                    <span className="text-white">{name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-grey">Source:</span>
                    <span className="text-white">
                      {connections.find(c => c.id === sourceConnectionId)?.name || 'Not selected'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-grey">Destination:</span>
                    <span className="text-white">{formatTargetType(targetType)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-grey">Tables:</span>
                    <span className="text-white">{tables || 'All tables'}</span>
                  </div>
                </div>
              </div>

              <div className="flex justify-between pt-4">
                <button
                  type="button"
                  onClick={() => setStep(2)}
                  className="btn-secondary px-5 py-2.5 text-sm"
                >
                  Back
                </button>
                <div className="flex gap-3">
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
              </div>
            </div>
          )}
        </form>
      </div>
    </div>
  );
}

function StepBadge({
  number,
  active,
  completed,
  label
}: {
  number: number;
  active: boolean;
  completed: boolean;
  label: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <div
        className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium transition-colors ${
          completed
            ? 'bg-accent-cyan text-white'
            : active
            ? 'bg-gradient-btn-primary text-white'
            : 'bg-primary-dark text-grey border border-cyan-40/30'
        }`}
      >
        {completed ? '✓' : number}
      </div>
      <span className={`text-sm ${active ? 'text-white' : 'text-grey'}`}>{label}</span>
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
