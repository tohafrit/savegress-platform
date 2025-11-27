'use client';

import { useState, useMemo } from 'react';
import {
  Zap,
  Server,
  Database,
  Clock,
  Shield,
  Copy,
  Check,
  ChevronRight,
  ChevronLeft,
  Download,
  RefreshCw,
  AlertCircle,
  Gauge,
  HardDrive,
  Lock,
} from 'lucide-react';

// Types
type WorkloadType = 'realtime' | 'streaming' | 'replication' | 'batch' | null;
type LatencyRequirement = 'ultra-low' | 'low' | 'standard' | null;
type DeliveryGuarantee = 'at-least-once' | 'at-most-once' | 'exactly-once' | null;
type VolumeLevel = 'low' | 'medium' | 'high' | null;
type RecoveryRequirement = 'minimal' | 'standard' | 'strict' | null;
type LicenseTier = 'community' | 'pro' | 'enterprise';

interface ConfigState {
  workload: WorkloadType;
  latency: LatencyRequirement;
  delivery: DeliveryGuarantee;
  volume: VolumeLevel;
  recovery: RecoveryRequirement;
}

interface GeneratedConfig {
  yaml: string;
  requiredTier: LicenseTier;
  features: string[];
  estimatedThroughput: string;
  estimatedLatency: string;
  compressionRatio: string;
}

// Step definitions
const STEPS = [
  { id: 'workload', title: 'Workload Type', description: 'What is your primary use case?' },
  { id: 'details', title: 'Requirements', description: 'Specific requirements for your workload' },
  { id: 'result', title: 'Configuration', description: 'Your optimized configuration' },
];

export default function OptimizerPage() {
  const [currentStep, setCurrentStep] = useState(0);
  const [config, setConfig] = useState<ConfigState>({
    workload: null,
    latency: null,
    delivery: null,
    volume: null,
    recovery: null,
  });
  const [copied, setCopied] = useState(false);

  const generatedConfig = useMemo(() => generateConfig(config), [config]);

  const handleWorkloadSelect = (workload: WorkloadType) => {
    setConfig({ ...config, workload, latency: null, delivery: null, volume: null, recovery: null });
    setCurrentStep(1);
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleNext = () => {
    if (currentStep < STEPS.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleReset = () => {
    setConfig({
      workload: null,
      latency: null,
      delivery: null,
      volume: null,
      recovery: null,
    });
    setCurrentStep(0);
  };

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(generatedConfig.yaml);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const downloadConfig = () => {
    const blob = new Blob([generatedConfig.yaml], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'savegress-config.yaml';
    a.click();
    URL.revokeObjectURL(url);
  };

  const canProceed = () => {
    if (currentStep === 0) return config.workload !== null;
    if (currentStep === 1) {
      switch (config.workload) {
        case 'realtime':
          return config.latency !== null;
        case 'streaming':
          return config.delivery !== null && config.volume !== null;
        case 'replication':
          return config.recovery !== null && config.volume !== null;
        case 'batch':
          return config.volume !== null;
        default:
          return false;
      }
    }
    return true;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-h4 text-white">Configuration Optimizer</h1>
        <p className="text-content-1 text-grey">
          Answer a few questions to get an optimized configuration for your use case
        </p>
      </div>

      {/* Progress */}
      <div className="card-dark p-4">
        <div className="flex items-center justify-between">
          {STEPS.map((step, index) => (
            <div key={step.id} className="flex items-center flex-1">
              <div className="flex items-center">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors ${
                    index < currentStep
                      ? 'bg-accent-cyan text-white'
                      : index === currentStep
                      ? 'bg-gradient-btn-primary text-white'
                      : 'bg-primary-dark text-grey'
                  }`}
                >
                  {index < currentStep ? <Check className="w-4 h-4" /> : index + 1}
                </div>
                <div className="ml-3 hidden sm:block">
                  <p className={`text-sm font-medium ${index <= currentStep ? 'text-white' : 'text-grey'}`}>
                    {step.title}
                  </p>
                  <p className="text-xs text-grey">{step.description}</p>
                </div>
              </div>
              {index < STEPS.length - 1 && (
                <div className={`flex-1 h-0.5 mx-4 ${index < currentStep ? 'bg-accent-cyan' : 'bg-primary-dark'}`} />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="card-dark p-6">
        {currentStep === 0 && (
          <WorkloadStep selected={config.workload} onSelect={handleWorkloadSelect} />
        )}

        {currentStep === 1 && (
          <DetailsStep
            workload={config.workload}
            config={config}
            onChange={(updates) => setConfig({ ...config, ...updates })}
          />
        )}

        {currentStep === 2 && (
          <ResultStep
            config={generatedConfig}
            onCopy={copyToClipboard}
            onDownload={downloadConfig}
            copied={copied}
          />
        )}
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <button
          onClick={currentStep === 2 ? handleReset : handleBack}
          disabled={currentStep === 0}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${
            currentStep === 0
              ? 'text-grey cursor-not-allowed'
              : 'text-white hover:bg-primary-dark'
          }`}
        >
          {currentStep === 2 ? (
            <>
              <RefreshCw className="w-4 h-4" />
              Start Over
            </>
          ) : (
            <>
              <ChevronLeft className="w-4 h-4" />
              Back
            </>
          )}
        </button>

        {currentStep < 2 && (
          <button
            onClick={handleNext}
            disabled={!canProceed()}
            className={`flex items-center gap-2 px-6 py-2.5 rounded-lg font-medium transition-all ${
              canProceed()
                ? 'bg-gradient-btn-primary text-white shadow-glow-blue hover:shadow-glow-blue-lg'
                : 'bg-primary-dark text-grey cursor-not-allowed'
            }`}
          >
            {currentStep === 1 ? 'Generate Config' : 'Next'}
            <ChevronRight className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
}

// Step 1: Workload Selection
function WorkloadStep({
  selected,
  onSelect,
}: {
  selected: WorkloadType;
  onSelect: (workload: WorkloadType) => void;
}) {
  const workloads = [
    {
      id: 'realtime' as WorkloadType,
      title: 'Real-time Analytics',
      description: 'Dashboard updates, live metrics, instant notifications',
      icon: Zap,
      examples: ['Live dashboards', 'Monitoring alerts', 'User activity tracking'],
    },
    {
      id: 'streaming' as WorkloadType,
      title: 'Event Streaming',
      description: 'Event-driven architecture, microservices communication',
      icon: Server,
      examples: ['Order processing', 'Inventory sync', 'Event sourcing'],
    },
    {
      id: 'replication' as WorkloadType,
      title: 'Data Replication',
      description: 'Database sync, disaster recovery, maintaining consistency',
      icon: Database,
      examples: ['DR setup', 'Read replicas', 'Data warehouse loading'],
    },
    {
      id: 'batch' as WorkloadType,
      title: 'Batch Processing',
      description: 'ETL pipelines, periodic sync, bulk operations',
      icon: Clock,
      examples: ['Nightly ETL', 'Report generation', 'Periodic backups'],
    },
  ];

  return (
    <div className="space-y-4">
      <h2 className="text-h5 text-white mb-6">What type of workload are you optimizing for?</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {workloads.map((workload) => (
          <button
            key={workload.id}
            onClick={() => onSelect(workload.id)}
            className={`p-5 rounded-xl border-2 text-left transition-all ${
              selected === workload.id
                ? 'border-accent-cyan bg-accent-cyan/10'
                : 'border-cyan-40 hover:border-accent-cyan/50 hover:bg-primary-dark/50'
            }`}
          >
            <div className="flex items-start gap-4">
              <div className={`p-3 rounded-lg ${selected === workload.id ? 'bg-accent-cyan/20' : 'bg-primary-dark'}`}>
                <workload.icon className={`w-6 h-6 ${selected === workload.id ? 'text-accent-cyan' : 'text-grey'}`} />
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-medium text-white">{workload.title}</h3>
                <p className="text-sm text-grey mt-1">{workload.description}</p>
                <div className="flex flex-wrap gap-2 mt-3">
                  {workload.examples.map((example) => (
                    <span
                      key={example}
                      className="text-xs px-2 py-1 rounded-full bg-primary-dark text-grey"
                    >
                      {example}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

// Step 2: Details based on workload
function DetailsStep({
  workload,
  config,
  onChange,
}: {
  workload: WorkloadType;
  config: ConfigState;
  onChange: (updates: Partial<ConfigState>) => void;
}) {
  if (workload === 'realtime') {
    return (
      <div className="space-y-6">
        <h2 className="text-h5 text-white">What is your latency requirement?</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <OptionCard
            selected={config.latency === 'ultra-low'}
            onClick={() => onChange({ latency: 'ultra-low' })}
            icon={Zap}
            title="Ultra-Low (< 10ms)"
            description="For real-time trading, gaming, instant responses"
            badge="Community"
          />
          <OptionCard
            selected={config.latency === 'low'}
            onClick={() => onChange({ latency: 'low' })}
            icon={Gauge}
            title="Low (10-100ms)"
            description="For live dashboards, notifications, monitoring"
            badge="Pro"
            badgeColor="text-accent-orange"
          />
          <OptionCard
            selected={config.latency === 'standard'}
            onClick={() => onChange({ latency: 'standard' })}
            icon={Clock}
            title="Standard (100ms-1s)"
            description="For analytics, reporting, non-critical updates"
            badge="Pro"
            badgeColor="text-accent-orange"
          />
        </div>
      </div>
    );
  }

  if (workload === 'streaming') {
    return (
      <div className="space-y-8">
        <div>
          <h2 className="text-h5 text-white mb-4">What delivery guarantee do you need?</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <OptionCard
              selected={config.delivery === 'at-least-once'}
              onClick={() => onChange({ delivery: 'at-least-once' })}
              icon={Check}
              title="At-Least-Once"
              description="Events delivered at least once, may have duplicates"
              badge="Pro"
              badgeColor="text-accent-orange"
            />
            <OptionCard
              selected={config.delivery === 'at-most-once'}
              onClick={() => onChange({ delivery: 'at-most-once' })}
              icon={Zap}
              title="At-Most-Once"
              description="Events delivered at most once, may lose events"
              badge="Community"
            />
            <OptionCard
              selected={config.delivery === 'exactly-once'}
              onClick={() => onChange({ delivery: 'exactly-once' })}
              icon={Shield}
              title="Exactly-Once"
              description="Every event delivered exactly once, no duplicates"
              badge="Enterprise"
              badgeColor="text-accent-cyan"
            />
          </div>
        </div>

        <div>
          <h2 className="text-h5 text-white mb-4">What is your event volume?</h2>
          <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} />
        </div>
      </div>
    );
  }

  if (workload === 'replication') {
    return (
      <div className="space-y-8">
        <div>
          <h2 className="text-h5 text-white mb-4">What is your Recovery Point Objective (RPO)?</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <OptionCard
              selected={config.recovery === 'minimal'}
              onClick={() => onChange({ recovery: 'minimal' })}
              icon={Zap}
              title="< 1 minute"
              description="Near-zero data loss, continuous replication"
              badge="Enterprise"
              badgeColor="text-accent-cyan"
            />
            <OptionCard
              selected={config.recovery === 'standard'}
              onClick={() => onChange({ recovery: 'standard' })}
              icon={Clock}
              title="1-15 minutes"
              description="Standard recovery window, periodic snapshots"
              badge="Pro"
              badgeColor="text-accent-orange"
            />
            <OptionCard
              selected={config.recovery === 'strict'}
              onClick={() => onChange({ recovery: 'strict' })}
              icon={HardDrive}
              title="> 15 minutes"
              description="Relaxed recovery, cost-optimized"
              badge="Community"
            />
          </div>
        </div>

        <div>
          <h2 className="text-h5 text-white mb-4">How many tables are you replicating?</h2>
          <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} volumeType="tables" />
        </div>
      </div>
    );
  }

  if (workload === 'batch') {
    return (
      <div className="space-y-6">
        <h2 className="text-h5 text-white mb-4">What is your batch volume?</h2>
        <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} />
      </div>
    );
  }

  return null;
}

// Option card component
function OptionCard({
  selected,
  onClick,
  icon: Icon,
  title,
  description,
  badge,
  badgeColor = 'text-grey',
}: {
  selected: boolean;
  onClick: () => void;
  icon: React.ElementType;
  title: string;
  description: string;
  badge?: string;
  badgeColor?: string;
}) {
  return (
    <button
      onClick={onClick}
      className={`p-4 rounded-xl border-2 text-left transition-all ${
        selected
          ? 'border-accent-cyan bg-accent-cyan/10'
          : 'border-cyan-40 hover:border-accent-cyan/50 hover:bg-primary-dark/50'
      }`}
    >
      <div className="flex items-center justify-between mb-3">
        <div className={`p-2 rounded-lg ${selected ? 'bg-accent-cyan/20' : 'bg-primary-dark'}`}>
          <Icon className={`w-5 h-5 ${selected ? 'text-accent-cyan' : 'text-grey'}`} />
        </div>
        {badge && <span className={`text-xs font-medium ${badgeColor}`}>{badge}</span>}
      </div>
      <h3 className="font-medium text-white">{title}</h3>
      <p className="text-sm text-grey mt-1">{description}</p>
    </button>
  );
}

// Volume selector component
function VolumeSelector({
  selected,
  onChange,
  volumeType = 'events',
}: {
  selected: VolumeLevel;
  onChange: (volume: VolumeLevel) => void;
  volumeType?: 'events' | 'tables';
}) {
  const options =
    volumeType === 'events'
      ? [
          { id: 'low' as VolumeLevel, title: 'Low', description: '< 1,000 events/sec', badge: 'Community' },
          { id: 'medium' as VolumeLevel, title: 'Medium', description: '1K - 50K events/sec', badge: 'Pro', badgeColor: 'text-accent-orange' },
          { id: 'high' as VolumeLevel, title: 'High', description: '> 50K events/sec', badge: 'Enterprise', badgeColor: 'text-accent-cyan' },
        ]
      : [
          { id: 'low' as VolumeLevel, title: 'Small', description: '< 10 tables', badge: 'Community' },
          { id: 'medium' as VolumeLevel, title: 'Medium', description: '10-100 tables', badge: 'Pro', badgeColor: 'text-accent-orange' },
          { id: 'high' as VolumeLevel, title: 'Large', description: '> 100 tables', badge: 'Enterprise', badgeColor: 'text-accent-cyan' },
        ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {options.map((option) => (
        <OptionCard
          key={option.id}
          selected={selected === option.id}
          onClick={() => onChange(option.id)}
          icon={option.id === 'low' ? Gauge : option.id === 'medium' ? Server : Zap}
          title={option.title}
          description={option.description}
          badge={option.badge}
          badgeColor={option.badgeColor}
        />
      ))}
    </div>
  );
}

// Step 3: Result
function ResultStep({
  config,
  onCopy,
  onDownload,
  copied,
}: {
  config: GeneratedConfig;
  onCopy: () => void;
  onDownload: () => void;
  copied: boolean;
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-h5 text-white">Your Optimized Configuration</h2>
        <div className="flex items-center gap-2">
          <span
            className={`px-3 py-1 rounded-full text-sm font-medium ${
              config.requiredTier === 'enterprise'
                ? 'bg-accent-cyan/20 text-accent-cyan'
                : config.requiredTier === 'pro'
                ? 'bg-accent-orange/20 text-accent-orange'
                : 'bg-grey/20 text-grey'
            }`}
          >
            {config.requiredTier.charAt(0).toUpperCase() + config.requiredTier.slice(1)} License
          </span>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <p className="text-xs text-grey mb-1">Est. Throughput</p>
          <p className="text-lg font-medium text-white">{config.estimatedThroughput}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <p className="text-xs text-grey mb-1">Est. Latency</p>
          <p className="text-lg font-medium text-white">{config.estimatedLatency}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <p className="text-xs text-grey mb-1">Compression</p>
          <p className="text-lg font-medium text-white">{config.compressionRatio}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <p className="text-xs text-grey mb-1">Features</p>
          <p className="text-lg font-medium text-white">{config.features.length}</p>
        </div>
      </div>

      {/* Features */}
      <div>
        <h3 className="text-sm font-medium text-grey mb-3">Enabled Features</h3>
        <div className="flex flex-wrap gap-2">
          {config.features.map((feature) => (
            <span key={feature} className="px-3 py-1 rounded-full bg-accent-cyan/10 text-accent-cyan text-sm">
              {feature}
            </span>
          ))}
        </div>
      </div>

      {/* YAML Config */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-grey">Configuration (YAML)</h3>
          <div className="flex items-center gap-2">
            <button
              onClick={onCopy}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-primary-dark hover:bg-cyan-40/20 text-grey hover:text-white transition-colors text-sm"
            >
              {copied ? <Check className="w-4 h-4 text-accent-cyan" /> : <Copy className="w-4 h-4" />}
              {copied ? 'Copied!' : 'Copy'}
            </button>
            <button
              onClick={onDownload}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-primary-dark hover:bg-cyan-40/20 text-grey hover:text-white transition-colors text-sm"
            >
              <Download className="w-4 h-4" />
              Download
            </button>
          </div>
        </div>
        <pre className="bg-[#0a1628] rounded-xl p-4 overflow-x-auto text-sm text-grey font-mono border border-cyan-40">
          <code>{config.yaml}</code>
        </pre>
      </div>

      {/* Pro/Enterprise notice */}
      {config.requiredTier !== 'community' && (
        <div className="flex items-start gap-3 p-4 rounded-lg bg-accent-orange/10 border border-accent-orange/30">
          <AlertCircle className="w-5 h-5 text-accent-orange flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-sm text-white font-medium">
              This configuration requires a {config.requiredTier.charAt(0).toUpperCase() + config.requiredTier.slice(1)} license
            </p>
            <p className="text-sm text-grey mt-1">
              {config.requiredTier === 'enterprise'
                ? 'Contact sales for Enterprise pricing and features.'
                : 'Upgrade to Pro to unlock compression, DLQ, and more.'}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

// Config generation logic
function generateConfig(state: ConfigState): GeneratedConfig {
  const { workload, latency, delivery, volume, recovery } = state;

  let yaml = '# Savegress Configuration\n# Generated by Optimizer\n\n';
  let requiredTier: LicenseTier = 'community';
  const features: string[] = [];
  let estimatedThroughput = '1K events/sec';
  let estimatedLatency = '< 1s';
  let compressionRatio = 'None';

  // Source config (placeholder)
  yaml += `source:
  type: postgres
  host: \${DB_HOST}
  port: 5432
  database: \${DB_NAME}
  user: \${DB_USER}
  password: \${DB_PASSWORD}
  slot_name: savegress_slot
  publication: savegress_pub
  tables:
    - public.*

`;

  // Workload-specific config
  if (workload === 'realtime') {
    if (latency === 'ultra-low') {
      yaml += `# Ultra-low latency configuration
compression:
  enabled: false

batching:
  max_size: 1
  max_wait: 1ms
  adaptive: false

buffer:
  type: ring
  size: 1024
  overflow_policy: drop_oldest

rate_limiting:
  algorithm: token_bucket
  tokens_per_second: 100000
  burst_size: 1000

backpressure:
  strategy: pause
  high_watermark: 0.7
  low_watermark: 0.3

replication:
  ack_mode: none
`;
      features.push('Token Bucket Rate Limiting', 'Ring Buffer', 'No Compression');
      estimatedLatency = '< 10ms';
      estimatedThroughput = '100K events/sec';
    } else if (latency === 'low') {
      requiredTier = 'pro';
      yaml += `# Low latency configuration
compression:
  enabled: true
  algorithm: lz4
  lz4:
    level: 3
  min_size: 512

batching:
  max_size: 10
  max_wait: 10ms
  adaptive: true

buffer:
  type: ring
  size: 4096
  overflow:
    enabled: true
    max_size: 1GB

rate_limiting:
  algorithm: token_bucket
  tokens_per_second: 50000
  burst_size: 500

backpressure:
  strategy: adaptive_throttle
  high_watermark: 0.75
  low_watermark: 0.4

replication:
  ack_mode: leader
`;
      features.push('LZ4 Compression', 'Adaptive Batching', 'Disk Overflow', 'Adaptive Throttle');
      estimatedLatency = '10-100ms';
      estimatedThroughput = '50K events/sec';
      compressionRatio = '2-3x';
    } else {
      requiredTier = 'pro';
      yaml += `# Standard latency configuration
compression:
  enabled: true
  algorithm: hybrid
  hybrid:
    threshold: 4096
    small_algo: lz4
    large_algo: zstd
    large_level: 3

batching:
  max_size: 100
  max_wait: 100ms
  adaptive: true
  target_latency: 50ms

rate_limiting:
  algorithm: sliding_window
  window_size: 1s
  max_requests: 10000

backpressure:
  strategy: adaptive_throttle
  high_watermark: 0.8

replication:
  ack_mode: leader
`;
      features.push('Hybrid Compression', 'Sliding Window Rate Limit', 'Adaptive Batching');
      estimatedLatency = '100ms - 1s';
      estimatedThroughput = '10K events/sec';
      compressionRatio = '4-6x';
    }
  }

  if (workload === 'streaming') {
    if (delivery === 'exactly-once') {
      requiredTier = 'enterprise';
      yaml += `# Exactly-once streaming configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 3
  simd:
    enabled: true

batching:
  max_size: 1000
  max_wait: 50ms
  adaptive: true

exactly_once:
  enabled: true
  transaction_timeout: 60s
  idempotent_producer: true

dlq:
  enabled: true
  max_retries: 10
  preserve_order: true

replication:
  ack_mode: all
  min_isr: 2

ha:
  enabled: true
`;
      features.push('Exactly-Once', 'SIMD Compression', 'DLQ', 'HA Mode', 'All-ISR ACK');
      estimatedLatency = '50-200ms';
      compressionRatio = '4-8x';
    } else if (delivery === 'at-least-once') {
      requiredTier = 'pro';
      yaml += `# At-least-once streaming configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 5

batching:
  max_size: 200
  max_wait: 100ms
  adaptive: true

dlq:
  enabled: true
  max_retries: 5
  retry_delay: 1s
  exponential_backoff: true

retry:
  max_attempts: 10
  initial_delay: 100ms
  max_delay: 30s
  multiplier: 2.0

backpressure:
  strategy: adaptive_throttle
  high_watermark: 0.8

replication:
  ack_mode: leader
`;
      features.push('ZSTD Compression', 'DLQ', 'Exponential Backoff', 'Adaptive Throttle');
      estimatedLatency = '100-500ms';
      compressionRatio = '4-10x';
    } else {
      yaml += `# At-most-once streaming configuration
compression:
  enabled: false

batching:
  max_size: 50
  max_wait: 50ms

buffer:
  overflow_policy: drop_newest

replication:
  ack_mode: none
`;
      features.push('Fire-and-Forget', 'Drop on Overflow');
      estimatedLatency = '< 50ms';
    }

    // Volume adjustments
    if (volume === 'high') {
      requiredTier = 'enterprise';
      estimatedThroughput = '100K+ events/sec';
    } else if (volume === 'medium') {
      if (requiredTier === 'community') requiredTier = 'pro';
      estimatedThroughput = '10K-50K events/sec';
    } else {
      estimatedThroughput = '< 1K events/sec';
    }
  }

  if (workload === 'replication') {
    if (recovery === 'minimal') {
      requiredTier = 'enterprise';
      yaml += `# Minimal RPO replication configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 5
  simd:
    enabled: true

checkpoint:
  interval: 10s
  sync_mode: sync

pitr:
  enabled: true
  retention: 7d
  granularity: 1m

storage:
  backend: s3
  s3:
    bucket: \${S3_BUCKET}
    region: \${AWS_REGION}
    sync_interval: 1m

replication:
  ack_mode: all
  min_isr: 2

schema:
  evolution:
    enabled: true
    approval_workflow: true

ha:
  enabled: true
  cluster:
    consensus: raft
    nodes: 3
`;
      features.push('PITR', 'Cloud Storage', 'Raft Clustering', 'Schema Approval', 'SIMD');
      estimatedLatency = '< 1 min RPO';
      compressionRatio = '4-8x';
    } else if (recovery === 'standard') {
      requiredTier = 'pro';
      yaml += `# Standard replication configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 5

checkpoint:
  interval: 1m
  sync_mode: async

snapshot:
  enabled: true
  interval: 1h
  retention_count: 24

schema:
  evolution:
    enabled: true
    compatible_changes: auto

replication:
  ack_mode: leader
`;
      features.push('Snapshots', 'Auto Schema Evolution', 'ZSTD Compression');
      estimatedLatency = '1-15 min RPO';
      compressionRatio = '4-10x';
    } else {
      yaml += `# Relaxed replication configuration
checkpoint:
  interval: 15m
  sync_mode: async

replication:
  ack_mode: leader
`;
      features.push('Basic Checkpointing');
      estimatedLatency = '> 15 min RPO';
    }

    if (volume === 'high') {
      requiredTier = 'enterprise';
      estimatedThroughput = 'Unlimited tables';
    } else if (volume === 'medium') {
      if (requiredTier === 'community') requiredTier = 'pro';
      estimatedThroughput = '10-100 tables';
    } else {
      estimatedThroughput = '< 10 tables';
    }
  }

  if (workload === 'batch') {
    if (volume === 'high') {
      requiredTier = 'enterprise';
      yaml += `# High volume batch configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 12
  simd:
    enabled: true

batching:
  mode: time
  interval: 5m
  max_size: 100000

parallel:
  table_parallelism: 32
  transaction_parallelism: 16

storage:
  segment_size: 1GB
  compaction:
    enabled: true
`;
      features.push('SIMD Compression', 'Parallel Processing', 'Segment Compaction');
      estimatedThroughput = '100K+ events/sec';
      estimatedLatency = 'Batch interval';
      compressionRatio = '8-15x';
    } else if (volume === 'medium') {
      requiredTier = 'pro';
      yaml += `# Medium volume batch configuration
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 9

batching:
  mode: time
  interval: 1h
  max_size: 50000

parallel:
  table_parallelism: 8
`;
      features.push('High Compression', 'Time-based Batching');
      estimatedThroughput = '10K-50K events/sec';
      estimatedLatency = 'Hourly';
      compressionRatio = '6-10x';
    } else {
      yaml += `# Low volume batch configuration
batching:
  mode: time
  interval: 24h
  max_size: 10000
`;
      features.push('Daily Batching');
      estimatedThroughput = '< 1K events/sec';
      estimatedLatency = 'Daily';
    }
  }

  // Add output config
  yaml += `
output:
  type: webhook  # or kafka, grpc
  url: \${WEBHOOK_URL}
  batch_size: 100
  retry:
    enabled: true
    max_attempts: 3

logging:
  level: info
  format: json

metrics:
  enabled: true
  prometheus:
    enabled: true
`;

  if (requiredTier === 'pro' || requiredTier === 'enterprise') {
    features.push('Prometheus Metrics');
  }

  return {
    yaml,
    requiredTier,
    features,
    estimatedThroughput,
    estimatedLatency,
    compressionRatio,
  };
}
