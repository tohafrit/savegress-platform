'use client';

import { useState, useMemo } from 'react';
import {
  PageHeader,
  InfoBanner,
  HelpIcon,
  ExpandableSection,
  QuickGuide,
} from '@/components/ui/helpers';
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
  Lightbulb,
  Info,
  BookOpen,
  HelpCircle,
  Sparkles,
  ArrowRight,
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
  { id: 'details', title: 'Requirements', description: 'Fine-tune your needs' },
  { id: 'result', title: 'Configuration', description: 'Your optimized config' },
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
      {/* Header with detailed explanation */}
      <PageHeader
        title="Configuration Optimizer"
        description="Not sure how to configure Savegress for your use case? This wizard analyzes your requirements and generates an optimized configuration file with all the right settings."
        tip="Answer a few simple questions and we'll create a production-ready config for you!"
      />

      {/* Intro banner for first-time users */}
      {currentStep === 0 && (
        <InfoBanner type="info" title="How the Optimizer works" dismissible>
          <div className="space-y-2">
            <p>The optimizer asks about your workload type, performance requirements, and data volume to generate a configuration tailored to your needs.</p>
            <ul className="list-disc list-inside text-sm space-y-1 mt-2">
              <li><strong>Step 1:</strong> Choose your primary use case (real-time, streaming, etc.)</li>
              <li><strong>Step 2:</strong> Specify detailed requirements (latency, delivery guarantees)</li>
              <li><strong>Step 3:</strong> Get your optimized YAML configuration</li>
            </ul>
          </div>
        </InfoBanner>
      )}

      {/* Progress */}
      <div className="card-dark p-4">
        <div className="flex items-center justify-between">
          {STEPS.map((step, index) => (
            <div key={step.id} className="flex items-center flex-1">
              <div className="flex items-center">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium transition-colors ${
                    index < currentStep
                      ? 'bg-accent-cyan text-white'
                      : index === currentStep
                      ? 'bg-gradient-btn-primary text-white'
                      : 'bg-primary-dark text-grey border border-cyan-40/30'
                  }`}
                >
                  {index < currentStep ? <Check className="w-5 h-5" /> : index + 1}
                </div>
                <div className="ml-3 hidden sm:block">
                  <p className={`text-sm font-medium ${index <= currentStep ? 'text-white' : 'text-grey'}`}>
                    {step.title}
                  </p>
                  <p className="text-xs text-grey">{step.description}</p>
                </div>
              </div>
              {index < STEPS.length - 1 && (
                <div className={`flex-1 h-0.5 mx-4 transition-colors ${index < currentStep ? 'bg-accent-cyan' : 'bg-primary-dark'}`} />
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
            {currentStep === 1 ? (
              <>
                <Sparkles className="w-4 h-4" />
                Generate Config
              </>
            ) : (
              <>
                Next
                <ChevronRight className="w-4 h-4" />
              </>
            )}
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
      description: 'For dashboards that update instantly, live monitoring, and immediate notifications. Every change appears within milliseconds.',
      icon: Zap,
      examples: ['Live dashboards', 'Monitoring alerts', 'User activity tracking'],
      tip: 'Choose this if you need sub-second updates',
    },
    {
      id: 'streaming' as WorkloadType,
      title: 'Event Streaming',
      description: 'For microservices that need to react to database changes, event-driven architectures, and async processing.',
      icon: Server,
      examples: ['Order processing', 'Inventory sync', 'Event sourcing'],
      tip: 'Ideal for event-driven microservices',
    },
    {
      id: 'replication' as WorkloadType,
      title: 'Data Replication',
      description: 'For maintaining database copies, disaster recovery setups, or keeping data warehouses in sync with production.',
      icon: Database,
      examples: ['DR setup', 'Read replicas', 'Data warehouse loading'],
      tip: 'Best for database synchronization',
    },
    {
      id: 'batch' as WorkloadType,
      title: 'Batch Processing',
      description: 'For nightly ETL jobs, periodic data exports, or when real-time isn\'t necessary and you want to optimize for throughput.',
      icon: Clock,
      examples: ['Nightly ETL', 'Report generation', 'Periodic backups'],
      tip: 'Most cost-effective for non-urgent data',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-h5 text-white">What type of workload are you optimizing for?</h2>
        <p className="text-sm text-grey">
          Select the option that best describes your primary use case. This determines which settings matter most for your configuration.
        </p>
      </div>

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
                <p className="text-xs text-accent-cyan mt-3 flex items-center gap-1">
                  <Lightbulb className="w-3 h-3" />
                  {workload.tip}
                </p>
              </div>
            </div>
          </button>
        ))}
      </div>

      <ExpandableSection title="Not sure which to choose?" icon={HelpCircle}>
        <div className="space-y-4 text-sm text-grey">
          <p>Here&apos;s a quick decision guide:</p>
          <ul className="space-y-2">
            <li className="flex items-start gap-2">
              <ArrowRight className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
              <span><strong className="text-white">Real-time Analytics</strong>: You need data to appear in dashboards or trigger alerts within milliseconds of the change.</span>
            </li>
            <li className="flex items-start gap-2">
              <ArrowRight className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
              <span><strong className="text-white">Event Streaming</strong>: Your services communicate via events and you care about delivery guarantees (at-least-once, exactly-once).</span>
            </li>
            <li className="flex items-start gap-2">
              <ArrowRight className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
              <span><strong className="text-white">Data Replication</strong>: You&apos;re syncing databases together (e.g., to a read replica or data warehouse) and care about consistency.</span>
            </li>
            <li className="flex items-start gap-2">
              <ArrowRight className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
              <span><strong className="text-white">Batch Processing</strong>: Real-time isn&apos;t required. You&apos;re running ETL jobs on a schedule and want maximum throughput.</span>
            </li>
          </ul>
        </div>
      </ExpandableSection>
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
        <div className="space-y-2">
          <h2 className="text-h5 text-white">What is your latency requirement?</h2>
          <p className="text-sm text-grey">
            Latency is the time between a change happening in your database and it being delivered to your destination.
            Lower latency requires more resources but gives faster updates.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <OptionCard
            selected={config.latency === 'ultra-low'}
            onClick={() => onChange({ latency: 'ultra-low' })}
            icon={Zap}
            title="Ultra-Low (< 10ms)"
            description="For high-frequency trading, gaming, or anything requiring instant response"
            badge="Community"
            details="Minimal batching, no compression, ring buffer"
          />
          <OptionCard
            selected={config.latency === 'low'}
            onClick={() => onChange({ latency: 'low' })}
            icon={Gauge}
            title="Low (10-100ms)"
            description="For live dashboards and monitoring where sub-second updates are important"
            badge="Pro"
            badgeColor="text-accent-orange"
            details="LZ4 compression, adaptive batching, disk overflow"
          />
          <OptionCard
            selected={config.latency === 'standard'}
            onClick={() => onChange({ latency: 'standard' })}
            icon={Clock}
            title="Standard (100ms-1s)"
            description="For analytics and reporting where near-real-time is sufficient"
            badge="Pro"
            badgeColor="text-accent-orange"
            details="Hybrid compression, larger batches, best throughput"
          />
        </div>

        <InfoBanner type="tip" title="Latency vs Throughput tradeoff">
          Lower latency means smaller batches and less compression, which can reduce throughput.
          If you don&apos;t need sub-10ms updates, choosing &quot;Low&quot; or &quot;Standard&quot; will give you better resource efficiency.
        </InfoBanner>
      </div>
    );
  }

  if (workload === 'streaming') {
    return (
      <div className="space-y-8">
        <div className="space-y-4">
          <div className="space-y-2">
            <h2 className="text-h5 text-white">What delivery guarantee do you need?</h2>
            <p className="text-sm text-grey">
              Delivery guarantees determine how Savegress handles failures. This is one of the most important decisions for event-driven systems.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <OptionCard
              selected={config.delivery === 'at-most-once'}
              onClick={() => onChange({ delivery: 'at-most-once' })}
              icon={Zap}
              title="At-Most-Once"
              description="Events may be lost but will never be duplicated. Fastest option."
              badge="Community"
              details="Fire-and-forget, no retries"
            />
            <OptionCard
              selected={config.delivery === 'at-least-once'}
              onClick={() => onChange({ delivery: 'at-least-once' })}
              icon={Check}
              title="At-Least-Once"
              description="Events will be delivered at least once. Some duplicates possible."
              badge="Pro"
              badgeColor="text-accent-orange"
              details="Retries with backoff, DLQ support"
            />
            <OptionCard
              selected={config.delivery === 'exactly-once'}
              onClick={() => onChange({ delivery: 'exactly-once' })}
              icon={Shield}
              title="Exactly-Once"
              description="Every event delivered exactly once. Most reliable but complex."
              badge="Enterprise"
              badgeColor="text-accent-cyan"
              details="Transactions, idempotency, full HA"
            />
          </div>

          <ExpandableSection title="What do these guarantees mean?" icon={BookOpen}>
            <div className="space-y-3 text-sm text-grey">
              <p><strong className="text-white">At-Most-Once:</strong> If delivery fails, the event is dropped. Use when losing an occasional event is acceptable (e.g., analytics, metrics).</p>
              <p><strong className="text-white">At-Least-Once:</strong> Failed events are retried. Your consumer needs to handle duplicates (e.g., using idempotent operations or deduplication).</p>
              <p><strong className="text-white">Exactly-Once:</strong> Complex coordination ensures no data loss or duplicates. Required for financial transactions, inventory, or any business-critical data.</p>
            </div>
          </ExpandableSection>
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <h2 className="text-h5 text-white">What is your event volume?</h2>
            <p className="text-sm text-grey">
              How many events per second does your database typically produce? This helps us tune buffer sizes, parallelism, and compression.
            </p>
          </div>
          <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} />
        </div>
      </div>
    );
  }

  if (workload === 'replication') {
    return (
      <div className="space-y-8">
        <div className="space-y-4">
          <div className="space-y-2">
            <h2 className="text-h5 text-white flex items-center gap-2">
              What is your Recovery Point Objective (RPO)?
              <HelpIcon text="RPO is the maximum acceptable data loss in case of failure, measured in time" />
            </h2>
            <p className="text-sm text-grey">
              RPO determines how much data you could lose in a disaster. Shorter RPO requires more frequent checkpoints and replication.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <OptionCard
              selected={config.recovery === 'minimal'}
              onClick={() => onChange({ recovery: 'minimal' })}
              icon={Zap}
              title="< 1 minute"
              description="Near-zero data loss. Continuous replication with frequent checkpoints."
              badge="Enterprise"
              badgeColor="text-accent-cyan"
              details="PITR, Raft clustering, cloud storage"
            />
            <OptionCard
              selected={config.recovery === 'standard'}
              onClick={() => onChange({ recovery: 'standard' })}
              icon={Clock}
              title="1-15 minutes"
              description="Standard recovery window. Good balance of safety and resources."
              badge="Pro"
              badgeColor="text-accent-orange"
              details="Hourly snapshots, auto schema evolution"
            />
            <OptionCard
              selected={config.recovery === 'strict'}
              onClick={() => onChange({ recovery: 'strict' })}
              icon={HardDrive}
              title="> 15 minutes"
              description="Relaxed recovery. Cost-optimized for non-critical data."
              badge="Community"
              details="Async checkpoints, basic recovery"
            />
          </div>

          <InfoBanner type="info" title="What is RPO?">
            If your RPO is 1 minute, that means in the worst case (complete system failure), you could lose up to 1 minute of data.
            For critical systems like financial data, you want minimal RPO. For analytics or logs, longer RPO is often acceptable.
          </InfoBanner>
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <h2 className="text-h5 text-white">How many tables are you replicating?</h2>
            <p className="text-sm text-grey">
              The number of tables affects parallelism settings and resource requirements.
            </p>
          </div>
          <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} volumeType="tables" />
        </div>
      </div>
    );
  }

  if (workload === 'batch') {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <h2 className="text-h5 text-white">What is your batch volume?</h2>
          <p className="text-sm text-grey">
            For batch processing, we optimize for maximum throughput rather than low latency.
            Higher volumes benefit from more aggressive compression and parallelism.
          </p>
        </div>

        <VolumeSelector selected={config.volume} onChange={(volume) => onChange({ volume })} />

        <InfoBanner type="tip" title="Batch processing benefits">
          Batch mode allows for maximum compression (up to 15x), parallel processing across tables,
          and optimal resource utilization. Perfect for nightly ETL or periodic sync jobs.
        </InfoBanner>
      </div>
    );
  }

  return null;
}

// Option card component with more details
function OptionCard({
  selected,
  onClick,
  icon: Icon,
  title,
  description,
  badge,
  badgeColor = 'text-grey',
  details,
}: {
  selected: boolean;
  onClick: () => void;
  icon: React.ElementType;
  title: string;
  description: string;
  badge?: string;
  badgeColor?: string;
  details?: string;
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
        {badge && (
          <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${
            badgeColor === 'text-accent-cyan' ? 'bg-accent-cyan/20' :
            badgeColor === 'text-accent-orange' ? 'bg-accent-orange/20' :
            'bg-grey/20'
          } ${badgeColor}`}>
            {badge}
          </span>
        )}
      </div>
      <h3 className="font-medium text-white">{title}</h3>
      <p className="text-sm text-grey mt-1">{description}</p>
      {details && (
        <p className="text-xs text-accent-cyan/70 mt-2 flex items-center gap-1">
          <Info className="w-3 h-3" />
          {details}
        </p>
      )}
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
          { id: 'low' as VolumeLevel, title: 'Low Volume', description: '< 1,000 events/sec', badge: 'Community', tip: 'Small apps, development' },
          { id: 'medium' as VolumeLevel, title: 'Medium Volume', description: '1K - 50K events/sec', badge: 'Pro', badgeColor: 'text-accent-orange', tip: 'Production workloads' },
          { id: 'high' as VolumeLevel, title: 'High Volume', description: '> 50K events/sec', badge: 'Enterprise', badgeColor: 'text-accent-cyan', tip: 'Large-scale, multi-region' },
        ]
      : [
          { id: 'low' as VolumeLevel, title: 'Small', description: '< 10 tables', badge: 'Community', tip: 'Single service database' },
          { id: 'medium' as VolumeLevel, title: 'Medium', description: '10-100 tables', badge: 'Pro', badgeColor: 'text-accent-orange', tip: 'Typical production app' },
          { id: 'high' as VolumeLevel, title: 'Large', description: '> 100 tables', badge: 'Enterprise', badgeColor: 'text-accent-cyan', tip: 'Enterprise monolith or multi-tenant' },
        ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {options.map((option) => (
        <button
          key={option.id}
          onClick={() => onChange(option.id)}
          className={`p-4 rounded-xl border-2 text-left transition-all ${
            selected === option.id
              ? 'border-accent-cyan bg-accent-cyan/10'
              : 'border-cyan-40 hover:border-accent-cyan/50 hover:bg-primary-dark/50'
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className={`p-2 rounded-lg ${selected === option.id ? 'bg-accent-cyan/20' : 'bg-primary-dark'}`}>
              {option.id === 'low' ? <Gauge className={`w-5 h-5 ${selected === option.id ? 'text-accent-cyan' : 'text-grey'}`} /> :
               option.id === 'medium' ? <Server className={`w-5 h-5 ${selected === option.id ? 'text-accent-cyan' : 'text-grey'}`} /> :
               <Zap className={`w-5 h-5 ${selected === option.id ? 'text-accent-cyan' : 'text-grey'}`} />}
            </div>
            <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${
              option.badgeColor === 'text-accent-cyan' ? 'bg-accent-cyan/20' :
              option.badgeColor === 'text-accent-orange' ? 'bg-accent-orange/20' :
              'bg-grey/20'
            } ${option.badgeColor || 'text-grey'}`}>
              {option.badge}
            </span>
          </div>
          <h3 className="font-medium text-white">{option.title}</h3>
          <p className="text-sm text-grey">{option.description}</p>
          <p className="text-xs text-accent-cyan/70 mt-2">{option.tip}</p>
        </button>
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
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-h5 text-white flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-accent-cyan" />
            Your Optimized Configuration
          </h2>
          <p className="text-sm text-grey mt-1">
            This configuration is tailored to your requirements. Copy it or download to get started.
          </p>
        </div>
        <span
          className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap ${
            config.requiredTier === 'enterprise'
              ? 'bg-accent-cyan/20 text-accent-cyan border border-accent-cyan/30'
              : config.requiredTier === 'pro'
              ? 'bg-accent-orange/20 text-accent-orange border border-accent-orange/30'
              : 'bg-grey/20 text-grey border border-grey/30'
          }`}
        >
          {config.requiredTier.charAt(0).toUpperCase() + config.requiredTier.slice(1)} License Required
        </span>
      </div>

      {/* Stats with explanations */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-1">
            <p className="text-xs text-grey">Est. Throughput</p>
            <HelpIcon text="Maximum number of events that can be processed per second" />
          </div>
          <p className="text-lg font-medium text-white">{config.estimatedThroughput}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-1">
            <p className="text-xs text-grey">Est. Latency</p>
            <HelpIcon text="Expected delay between a change and its delivery" />
          </div>
          <p className="text-lg font-medium text-white">{config.estimatedLatency}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-1">
            <p className="text-xs text-grey">Compression</p>
            <HelpIcon text="How much smaller your data will be after compression" />
          </div>
          <p className="text-lg font-medium text-white">{config.compressionRatio}</p>
        </div>
        <div className="bg-primary-dark/50 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-1">
            <p className="text-xs text-grey">Features</p>
            <HelpIcon text="Number of optimizations enabled in this config" />
          </div>
          <p className="text-lg font-medium text-white">{config.features.length}</p>
        </div>
      </div>

      {/* Features explanation */}
      <div>
        <h3 className="text-sm font-medium text-grey mb-3 flex items-center gap-2">
          Enabled Features
          <HelpIcon text="These optimizations are included based on your requirements" />
        </h3>
        <div className="flex flex-wrap gap-2">
          {config.features.map((feature) => (
            <span key={feature} className="px-3 py-1.5 rounded-full bg-accent-cyan/10 text-accent-cyan text-sm border border-accent-cyan/20">
              {feature}
            </span>
          ))}
        </div>
      </div>

      {/* YAML Config */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-grey flex items-center gap-2">
            Configuration File (YAML)
            <HelpIcon text="Save this as savegress-config.yaml in your project" />
          </h3>
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
        <pre className="bg-[#0a1628] rounded-xl p-4 overflow-x-auto text-sm text-grey font-mono border border-cyan-40 max-h-96">
          <code>{config.yaml}</code>
        </pre>
      </div>

      {/* License notice with more context */}
      {config.requiredTier !== 'community' && (
        <InfoBanner
          type={config.requiredTier === 'enterprise' ? 'info' : 'warning'}
          title={`This configuration requires a ${config.requiredTier.charAt(0).toUpperCase() + config.requiredTier.slice(1)} license`}
          action={config.requiredTier === 'enterprise' ? { label: 'Contact Sales', href: '/contact' } : { label: 'Upgrade to Pro', href: '/pricing' }}
        >
          <p>
            {config.requiredTier === 'enterprise'
              ? 'Enterprise features include exactly-once delivery, PITR, HA clustering, and unlimited scale. Contact us for custom pricing.'
              : 'Pro includes compression, DLQ, adaptive batching, and more. Upgrade to unlock these features.'}
          </p>
        </InfoBanner>
      )}

      {/* Next steps */}
      <ExpandableSection title="What's next?" icon={BookOpen} defaultExpanded>
        <div className="space-y-3 text-sm text-grey">
          <p>Now that you have your configuration:</p>
          <ol className="list-decimal list-inside space-y-2">
            <li>Save this file as <code className="text-accent-cyan bg-primary-dark px-1 rounded">savegress-config.yaml</code> in your project</li>
            <li>Replace the placeholder values (<code className="text-accent-cyan bg-primary-dark px-1 rounded">${'{'}DB_HOST{'}'}</code>, etc.) with your actual credentials</li>
            <li>Create a connection in Savegress pointing to your database</li>
            <li>Create a pipeline using this configuration</li>
            <li>Start your pipeline and watch the data flow!</li>
          </ol>
        </div>
      </ExpandableSection>
    </div>
  );
}

// Config generation logic
function generateConfig(state: ConfigState): GeneratedConfig {
  const { workload, latency, delivery, volume, recovery } = state;

  let yaml = '# Savegress Configuration\n# Generated by Optimizer\n# See docs at https://savegress.com/docs/config\n\n';
  let requiredTier: LicenseTier = 'community';
  const features: string[] = [];
  let estimatedThroughput = '1K events/sec';
  let estimatedLatency = '< 1s';
  let compressionRatio = 'None';

  // Source config (placeholder)
  yaml += `# Source Database Connection
# Replace these environment variables with your actual values
source:
  type: postgres
  host: \${DB_HOST}        # e.g., db.example.com
  port: 5432
  database: \${DB_NAME}    # e.g., myapp
  user: \${DB_USER}        # e.g., replication_user
  password: \${DB_PASSWORD}
  slot_name: savegress_slot
  publication: savegress_pub
  tables:
    - public.*             # Replicate all tables in public schema

`;

  // Workload-specific config
  if (workload === 'realtime') {
    if (latency === 'ultra-low') {
      yaml += `# Ultra-low latency configuration
# Optimized for < 10ms end-to-end latency

compression:
  enabled: false           # No compression for minimum latency

batching:
  max_size: 1              # Send each event immediately
  max_wait: 1ms
  adaptive: false

buffer:
  type: ring               # In-memory ring buffer
  size: 1024
  overflow_policy: drop_oldest  # Drop old events if buffer fills

rate_limiting:
  algorithm: token_bucket
  tokens_per_second: 100000
  burst_size: 1000

backpressure:
  strategy: pause          # Pause source if destination is slow
  high_watermark: 0.7
  low_watermark: 0.3

replication:
  ack_mode: none           # Fire and forget for speed
`;
      features.push('Token Bucket Rate Limiting', 'Ring Buffer', 'No Compression');
      estimatedLatency = '< 10ms';
      estimatedThroughput = '100K events/sec';
    } else if (latency === 'low') {
      requiredTier = 'pro';
      yaml += `# Low latency configuration
# Optimized for 10-100ms latency with good throughput

compression:
  enabled: true
  algorithm: lz4           # Fast compression
  lz4:
    level: 3               # Balance of speed and ratio
  min_size: 512            # Only compress larger messages

batching:
  max_size: 10
  max_wait: 10ms
  adaptive: true           # Automatically adjust batch size

buffer:
  type: ring
  size: 4096
  overflow:
    enabled: true
    max_size: 1GB          # Spill to disk if needed

rate_limiting:
  algorithm: token_bucket
  tokens_per_second: 50000
  burst_size: 500

backpressure:
  strategy: adaptive_throttle
  high_watermark: 0.75
  low_watermark: 0.4

replication:
  ack_mode: leader         # Wait for leader acknowledgment
`;
      features.push('LZ4 Compression', 'Adaptive Batching', 'Disk Overflow', 'Adaptive Throttle');
      estimatedLatency = '10-100ms';
      estimatedThroughput = '50K events/sec';
      compressionRatio = '2-3x';
    } else {
      requiredTier = 'pro';
      yaml += `# Standard latency configuration
# Optimized for best throughput with acceptable latency

compression:
  enabled: true
  algorithm: hybrid        # Use different compression based on size
  hybrid:
    threshold: 4096
    small_algo: lz4        # LZ4 for small messages
    large_algo: zstd       # ZSTD for large messages
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
# Guaranteed exactly-once delivery - no duplicates, no data loss

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 3
  simd:
    enabled: true          # Use SIMD for faster compression

batching:
  max_size: 1000
  max_wait: 50ms
  adaptive: true

exactly_once:
  enabled: true
  transaction_timeout: 60s
  idempotent_producer: true  # Enable idempotent writes

dlq:
  enabled: true            # Dead letter queue for failed events
  max_retries: 10
  preserve_order: true     # Maintain event ordering

replication:
  ack_mode: all            # Wait for all replicas
  min_isr: 2               # Minimum in-sync replicas

ha:
  enabled: true            # High availability mode
`;
      features.push('Exactly-Once', 'SIMD Compression', 'DLQ', 'HA Mode', 'All-ISR ACK');
      estimatedLatency = '50-200ms';
      compressionRatio = '4-8x';
    } else if (delivery === 'at-least-once') {
      requiredTier = 'pro';
      yaml += `# At-least-once streaming configuration
# Events guaranteed to be delivered, may have duplicates

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
  enabled: true            # Dead letter queue
  max_retries: 5
  retry_delay: 1s
  exponential_backoff: true

retry:
  max_attempts: 10
  initial_delay: 100ms
  max_delay: 30s
  multiplier: 2.0          # Double delay each retry

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
# Fastest delivery, events may be lost on failure

compression:
  enabled: false

batching:
  max_size: 50
  max_wait: 50ms

buffer:
  overflow_policy: drop_newest  # Drop new events if buffer is full

replication:
  ack_mode: none           # Fire and forget
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
# Near-zero data loss with Point-in-Time Recovery

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 5
  simd:
    enabled: true

checkpoint:
  interval: 10s            # Checkpoint every 10 seconds
  sync_mode: sync          # Synchronous checkpointing

pitr:
  enabled: true            # Point-in-Time Recovery
  retention: 7d            # Keep 7 days of history
  granularity: 1m          # 1-minute recovery granularity

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
    approval_workflow: true  # Require approval for schema changes

ha:
  enabled: true
  cluster:
    consensus: raft        # Raft consensus for coordination
    nodes: 3
`;
      features.push('PITR', 'Cloud Storage', 'Raft Clustering', 'Schema Approval', 'SIMD');
      estimatedLatency = '< 1 min RPO';
      compressionRatio = '4-8x';
    } else if (recovery === 'standard') {
      requiredTier = 'pro';
      yaml += `# Standard replication configuration
# Good balance of data safety and resource usage

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
  interval: 1h             # Hourly snapshots
  retention_count: 24      # Keep 24 snapshots

schema:
  evolution:
    enabled: true
    compatible_changes: auto  # Auto-apply compatible changes

replication:
  ack_mode: leader
`;
      features.push('Snapshots', 'Auto Schema Evolution', 'ZSTD Compression');
      estimatedLatency = '1-15 min RPO';
      compressionRatio = '4-10x';
    } else {
      yaml += `# Relaxed replication configuration
# Cost-optimized for non-critical data

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
# Maximum throughput for large-scale ETL

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 12              # Maximum compression
  simd:
    enabled: true

batching:
  mode: time
  interval: 5m             # Process every 5 minutes
  max_size: 100000

parallel:
  table_parallelism: 32    # Process 32 tables in parallel
  transaction_parallelism: 16

storage:
  segment_size: 1GB
  compaction:
    enabled: true          # Compact storage periodically
`;
      features.push('SIMD Compression', 'Parallel Processing', 'Segment Compaction');
      estimatedThroughput = '100K+ events/sec';
      estimatedLatency = 'Batch interval';
      compressionRatio = '8-15x';
    } else if (volume === 'medium') {
      requiredTier = 'pro';
      yaml += `# Medium volume batch configuration
# Good throughput with hourly batches

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
# Simple daily processing

batching:
  mode: time
  interval: 24h            # Daily batch
  max_size: 10000
`;
      features.push('Daily Batching');
      estimatedThroughput = '< 1K events/sec';
      estimatedLatency = 'Daily';
    }
  }

  // Add output config
  yaml += `
# Output Configuration
# Configure where events should be sent
output:
  type: webhook            # Options: webhook, kafka, grpc, s3
  url: \${WEBHOOK_URL}     # Your destination endpoint
  batch_size: 100
  retry:
    enabled: true
    max_attempts: 3

# Observability
logging:
  level: info
  format: json

metrics:
  enabled: true
  prometheus:
    enabled: true
    port: 9090
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
