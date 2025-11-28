'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, Connection, Pipeline } from '@/lib/api';
import {
  PageHeader,
  InfoBanner,
  HelpIcon,
  ExpandableSection,
} from '@/components/ui/helpers';
import {
  Download,
  Copy,
  Check,
  ChevronRight,
  ChevronLeft,
  Database,
  Box,
  Cloud,
  FileCode,
  Terminal,
  Zap,
  Info,
  Lightbulb,
  BookOpen,
  ArrowRight,
} from 'lucide-react';

type Step = 'choose' | 'method' | 'source' | 'deploy';

const INSTALLATION_METHODS = [
  {
    id: 'docker-compose',
    name: 'Docker Compose',
    icon: Box,
    description: 'Recommended for local and single-server deployments',
    pros: ['Easy to set up', 'Great for development', 'Single command to start'],
    cons: ['Not for production clusters', 'Manual scaling'],
  },
  {
    id: 'helm',
    name: 'Kubernetes (Helm)',
    icon: Cloud,
    description: 'For Kubernetes clusters with Helm',
    pros: ['Production-ready', 'Auto-scaling', 'High availability'],
    cons: ['Requires K8s knowledge', 'More complex setup'],
  },
  {
    id: 'env',
    name: 'Environment File',
    icon: FileCode,
    description: 'Environment variables for any deployment',
    pros: ['Works anywhere', 'Simple configuration', 'Container-friendly'],
    cons: ['Manual management', 'No orchestration'],
  },
  {
    id: 'systemd',
    name: 'Systemd Service',
    icon: Terminal,
    description: 'For Linux servers with systemd',
    pros: ['Native Linux integration', 'Auto-restart on failure', 'Boot startup'],
    cons: ['Linux only', 'Manual updates'],
  },
];

export default function SetupPage() {
  const [step, setStep] = useState<Step>('choose');
  const [method, setMethod] = useState('docker-compose');
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [selectedPipeline, setSelectedPipeline] = useState<string | null>(null);
  const [generatedConfig, setGeneratedConfig] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    const pipelinesRes = await api.getPipelines();
    if (pipelinesRes.data) setPipelines(pipelinesRes.data.pipelines);
    setIsLoading(false);
  }

  async function generateConfig() {
    setIsGenerating(true);
    const { data } = await api.generateConfig(method, selectedPipeline || undefined);
    if (data) {
      setGeneratedConfig(typeof data === 'string' ? data : JSON.stringify(data, null, 2));
    }
    setIsGenerating(false);
    setStep('deploy');
  }

  function copyToClipboard() {
    navigator.clipboard.writeText(generatedConfig);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  function downloadConfig() {
    const filename = method === 'docker-compose'
      ? 'docker-compose.yml'
      : method === 'helm'
        ? 'values.yaml'
        : method === 'env'
          ? 'savegress.env'
          : 'savegress.service';

    const blob = new Blob([generatedConfig], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  }

  if (isLoading) {
    return <SetupSkeleton />;
  }

  // Initial choice screen
  if (step === 'choose') {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Get Started"
          description="Choose how you'd like to set up Savegress CDC Engine"
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Wizard Option */}
          <button
            onClick={() => setStep('method')}
            className="card-dark p-8 text-left hover:border-accent-cyan transition-all group"
          >
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="w-16 h-16 rounded-2xl bg-gradient-btn-primary flex items-center justify-center group-hover:shadow-glow-blue transition-shadow">
                <Zap className="w-8 h-8 text-white" />
              </div>
              <div>
                <h3 className="text-xl font-semibold text-white mb-2">Quick Setup Wizard</h3>
                <p className="text-grey text-sm">
                  Step-by-step guided setup through our web interface.
                  Perfect for getting started quickly.
                </p>
              </div>
              <ul className="text-sm text-grey space-y-2 text-left w-full">
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  Interactive configuration builder
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  Auto-generated config files
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  Best for beginners
                </li>
              </ul>
              <div className="flex items-center gap-2 text-accent-cyan font-medium">
                Start Wizard
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </div>
            </div>
          </button>

          {/* Manual Option */}
          <Link
            href="/docs#installation"
            className="card-dark p-8 text-left hover:border-accent-cyan transition-all group"
          >
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="w-16 h-16 rounded-2xl bg-primary-dark border border-cyan-40 flex items-center justify-center group-hover:border-accent-cyan transition-colors">
                <Terminal className="w-8 h-8 text-accent-cyan" />
              </div>
              <div>
                <h3 className="text-xl font-semibold text-white mb-2">Manual Setup</h3>
                <p className="text-grey text-sm">
                  Follow detailed documentation to configure via CLI,
                  Docker, or Kubernetes manually.
                </p>
              </div>
              <ul className="text-sm text-grey space-y-2 text-left w-full">
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  Full control over configuration
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  CI/CD integration examples
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  Best for DevOps teams
                </li>
              </ul>
              <div className="flex items-center gap-2 text-accent-cyan font-medium">
                View Documentation
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </div>
            </div>
          </Link>
        </div>

        {/* Additional help */}
        <InfoBanner
          type="tip"
          title="Not sure which to choose?"
          action={{ label: 'Open Optimizer', href: '/optimizer' }}
        >
          Try the Configuration Optimizer to determine the best settings for your workload before setting up.
        </InfoBanner>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Setup Wizard"
        description="Generate deployment configuration for Savegress CDC Engine. We&apos;ll help you create the right configuration file for your infrastructure."
        tip="Not sure about settings? Use the Configuration Optimizer first to determine the best options for your workload."
        action={
          <Link href="/optimizer" className="btn-secondary px-5 py-3 text-sm">
            <Zap className="w-4 h-4 mr-2" />
            Open Optimizer
          </Link>
        }
      />

      {/* Progress Steps */}
      <div className="card-dark p-4">
        <div className="flex items-center justify-between">
          {[
            { id: 'method', label: 'Installation Method', description: 'Choose how to deploy' },
            { id: 'source', label: 'Select Pipeline', description: 'Pre-fill configuration' },
            { id: 'deploy', label: 'Deploy', description: 'Get your config file' },
          ].map((s, i, arr) => (
            <div key={s.id} className="flex items-center flex-1">
              <div className="flex items-center">
                <div className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium transition-colors ${
                  step === s.id ? 'bg-gradient-btn-primary text-white' :
                  arr.findIndex(x => x.id === step) > i ? 'bg-accent-cyan text-white' :
                  'bg-primary-dark text-grey border border-cyan-40/30'
                }`}>
                  {arr.findIndex(x => x.id === step) > i ? <Check className="w-5 h-5" /> : i + 1}
                </div>
                <div className="ml-3 hidden sm:block">
                  <span className={`text-sm ${step === s.id ? 'text-white font-medium' : 'text-grey'}`}>
                    {s.label}
                  </span>
                  <p className="text-xs text-grey">{s.description}</p>
                </div>
              </div>
              {i < arr.length - 1 && <div className={`flex-1 h-0.5 mx-4 ${arr.findIndex(x => x.id === step) > i ? 'bg-accent-cyan' : 'bg-primary-dark'}`} />}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="card-dark p-6">
        {step === 'method' && (
          <StepMethod
            method={method}
            setMethod={setMethod}
            onBack={() => setStep('choose')}
            onNext={() => setStep('source')}
          />
        )}
        {step === 'source' && (
          <StepSource
            pipelines={pipelines}
            selectedPipeline={selectedPipeline}
            setSelectedPipeline={setSelectedPipeline}
            onBack={() => setStep('method')}
            onNext={generateConfig}
            isGenerating={isGenerating}
          />
        )}
        {step === 'deploy' && (
          <StepDeploy
            method={method}
            config={generatedConfig}
            onCopy={copyToClipboard}
            onDownload={downloadConfig}
            copied={copied}
            onBack={() => setStep('source')}
          />
        )}
      </div>
    </div>
  );
}

function StepMethod({
  method,
  setMethod,
  onBack,
  onNext
}: {
  method: string;
  setMethod: (m: string) => void;
  onBack: () => void;
  onNext: () => void;
}) {
  const selectedMethod = INSTALLATION_METHODS.find(m => m.id === method);

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-h5 text-white flex items-center gap-2">
          Choose Installation Method
          <HelpIcon text="Select how you want to deploy the Savegress CDC Engine. Each method has its pros and cons." />
        </h2>
        <p className="text-sm text-grey">
          Pick the deployment method that best fits your infrastructure. Don&apos;t worry, you can always change later.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {INSTALLATION_METHODS.map((m) => (
          <button
            key={m.id}
            onClick={() => setMethod(m.id)}
            className={`flex items-start gap-4 p-4 rounded-xl text-left transition-all ${
              method === m.id
                ? 'bg-accent-cyan/10 border-2 border-accent-cyan'
                : 'bg-primary-dark/50 border-2 border-cyan-40/30 hover:border-cyan-40'
            }`}
          >
            <div className={`p-3 rounded-lg ${method === m.id ? 'bg-accent-cyan/20' : 'bg-primary-dark'}`}>
              <m.icon className={`w-6 h-6 ${method === m.id ? 'text-accent-cyan' : 'text-grey'}`} />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-white">{m.name}</h3>
              <p className="text-sm text-grey mt-1">{m.description}</p>
            </div>
          </button>
        ))}
      </div>

      {/* Details about selected method */}
      {selectedMethod && (
        <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
          <div className="flex items-start gap-3">
            <Info className="w-5 h-5 text-accent-cyan flex-shrink-0 mt-0.5" />
            <div className="space-y-3">
              <div>
                <h4 className="text-sm font-medium text-white mb-1">Pros</h4>
                <ul className="text-sm text-grey space-y-0.5">
                  {selectedMethod.pros.map((pro, i) => (
                    <li key={i} className="flex items-center gap-2">
                      <Check className="w-3 h-3 text-green-400" />
                      {pro}
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h4 className="text-sm font-medium text-white mb-1">Considerations</h4>
                <ul className="text-sm text-grey space-y-0.5">
                  {selectedMethod.cons.map((con, i) => (
                    <li key={i} className="flex items-center gap-2">
                      <span className="w-3 h-3 text-accent-orange">•</span>
                      {con}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-between">
        <button onClick={onBack} className="btn-secondary px-5 py-2.5">
          <ChevronLeft className="w-4 h-4 mr-2" />
          Back
        </button>
        <button onClick={onNext} className="btn-primary px-6 py-3">
          Continue
          <ChevronRight className="w-4 h-4 ml-2" />
        </button>
      </div>
    </div>
  );
}

function StepSource({
  pipelines,
  selectedPipeline,
  setSelectedPipeline,
  onBack,
  onNext,
  isGenerating,
}: {
  pipelines: Pipeline[];
  selectedPipeline: string | null;
  setSelectedPipeline: (id: string | null) => void;
  onBack: () => void;
  onNext: () => void;
  isGenerating: boolean;
}) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-h5 text-white flex items-center gap-2">
          Select Pipeline (Optional)
          <HelpIcon text="If you select a pipeline, we&apos;ll pre-fill the configuration with your connection details. Otherwise, you&apos;ll get a generic template." />
        </h2>
        <p className="text-sm text-grey">
          Select a pipeline to pre-fill the configuration with your actual settings, or skip to generate a generic template.
        </p>
      </div>

      <div className="space-y-2">
        <button
          onClick={() => setSelectedPipeline(null)}
          className={`flex items-center gap-3 w-full p-4 rounded-lg text-left transition-all ${
            selectedPipeline === null
              ? 'bg-accent-cyan/10 border-2 border-accent-cyan'
              : 'bg-primary-dark/50 border-2 border-cyan-40/30 hover:border-cyan-40'
          }`}
        >
          <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
            selectedPipeline === null ? 'border-accent-cyan bg-accent-cyan' : 'border-grey'
          }`}>
            {selectedPipeline === null && <Check className="w-3 h-3 text-white" />}
          </div>
          <div>
            <span className="font-medium text-white">Generic Template</span>
            <p className="text-sm text-grey">Start with placeholder values - you&apos;ll fill in the details later</p>
          </div>
        </button>

        {pipelines.map((pipeline) => (
          <button
            key={pipeline.id}
            onClick={() => setSelectedPipeline(pipeline.id)}
            className={`flex items-center gap-3 w-full p-4 rounded-lg text-left transition-all ${
              selectedPipeline === pipeline.id
                ? 'bg-accent-cyan/10 border-2 border-accent-cyan'
                : 'bg-primary-dark/50 border-2 border-cyan-40/30 hover:border-cyan-40'
            }`}
          >
            <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
              selectedPipeline === pipeline.id ? 'border-accent-cyan bg-accent-cyan' : 'border-grey'
            }`}>
              {selectedPipeline === pipeline.id && <Check className="w-3 h-3 text-white" />}
            </div>
            <div className="flex-1">
              <span className="font-medium text-white">{pipeline.name}</span>
              <div className="flex items-center gap-2 text-sm text-grey mt-1">
                <Database className="w-3 h-3" />
                <span>{pipeline.source_connection?.type || 'Unknown'}</span>
                <ArrowRight className="w-3 h-3" />
                <span>{pipeline.target_type}</span>
              </div>
            </div>
          </button>
        ))}

        {pipelines.length === 0 && (
          <InfoBanner type="info" title="No pipelines yet">
            You haven&apos;t created any pipelines yet. A generic template will be generated.
            <Link href="/pipelines" className="text-accent-cyan ml-1 hover:underline">
              Create a pipeline first →
            </Link>
          </InfoBanner>
        )}
      </div>

      <InfoBanner
        type="tip"
        title="Need optimized settings?"
        action={{ label: 'Open Optimizer', href: '/optimizer' }}
      >
        Use the Configuration Optimizer to generate settings tailored to your specific workload type (real-time, streaming, batch, etc.)
      </InfoBanner>

      <div className="flex justify-between">
        <button onClick={onBack} className="btn-secondary px-5 py-2.5">
          <ChevronLeft className="w-4 h-4 mr-2" />
          Back
        </button>
        <button
          onClick={onNext}
          disabled={isGenerating}
          className="btn-primary px-6 py-3 disabled:opacity-50"
        >
          {isGenerating ? 'Generating...' : 'Generate Config'}
          <ChevronRight className="w-4 h-4 ml-2" />
        </button>
      </div>
    </div>
  );
}

function StepDeploy({
  method,
  config,
  onCopy,
  onDownload,
  copied,
  onBack,
}: {
  method: string;
  config: string;
  onCopy: () => void;
  onDownload: () => void;
  copied: boolean;
  onBack: () => void;
}) {
  const methodInfo = INSTALLATION_METHODS.find(m => m.id === method);

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-h5 text-white flex items-center gap-2">
          Your Configuration is Ready!
          <HelpIcon text="Copy or download this configuration file, then follow the deployment steps below." />
        </h2>
        <p className="text-sm text-grey">
          Save this configuration file and use it to deploy Savegress CDC Engine.
        </p>
      </div>

      <div className="flex items-center justify-between p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
        <div className="flex items-center gap-3">
          {methodInfo && <methodInfo.icon className="w-6 h-6 text-accent-cyan" />}
          <div>
            <span className="font-medium text-white">{methodInfo?.name}</span>
            <p className="text-sm text-grey">{methodInfo?.description}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={onCopy} className="btn-secondary px-4 py-2 text-sm">
            {copied ? <Check className="w-4 h-4 mr-2 text-accent-cyan" /> : <Copy className="w-4 h-4 mr-2" />}
            {copied ? 'Copied!' : 'Copy'}
          </button>
          <button onClick={onDownload} className="btn-primary px-4 py-2 text-sm">
            <Download className="w-4 h-4 mr-2" />
            Download
          </button>
        </div>
      </div>

      <div className="bg-[#0a1628] rounded-xl overflow-hidden border border-cyan-40">
        <pre className="p-4 text-sm text-grey overflow-x-auto max-h-[350px] overflow-y-auto font-mono">
          <code>{config}</code>
        </pre>
      </div>

      {/* Next Steps */}
      <ExpandableSection title="Deployment Instructions" icon={BookOpen} defaultExpanded>
        <ol className="list-decimal list-inside text-sm text-grey space-y-2">
          {method === 'docker-compose' && (
            <>
              <li>Save the file as <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">docker-compose.yml</code></li>
              <li>Set your database password: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">export SOURCE_DB_PASSWORD=your_password</code></li>
              <li>Start the engine: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">docker-compose up -d</code></li>
              <li>Check the logs: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">docker-compose logs -f</code></li>
              <li>The engine will appear in your Dashboard when it connects</li>
            </>
          )}
          {method === 'helm' && (
            <>
              <li>Save the file as <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">values.yaml</code></li>
              <li>Create a Kubernetes secret for the database password</li>
              <li>Install with Helm: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">helm install cdc-engine savegress/cdc-engine -f values.yaml</code></li>
              <li>Check pod status: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">kubectl get pods</code></li>
              <li>View logs: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">kubectl logs -f deployment/cdc-engine</code></li>
            </>
          )}
          {method === 'env' && (
            <>
              <li>Save the file as <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">savegress.env</code></li>
              <li>Download the binary from the <Link href="/downloads" className="text-accent-cyan hover:underline">Downloads page</Link></li>
              <li>Load environment and run: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">source savegress.env && ./cdc-engine</code></li>
            </>
          )}
          {method === 'systemd' && (
            <>
              <li>Save as <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">/etc/systemd/system/savegress.service</code></li>
              <li>Create config directory: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">mkdir -p /etc/savegress</code></li>
              <li>Add your environment variables to <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">/etc/savegress/savegress.env</code></li>
              <li>Enable and start: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">systemctl enable savegress && systemctl start savegress</code></li>
              <li>Check status: <code className="bg-primary-dark px-1.5 py-0.5 rounded text-accent-cyan">systemctl status savegress</code></li>
            </>
          )}
        </ol>
      </ExpandableSection>

      <InfoBanner
        type="tip"
        title="Want to optimize your configuration?"
        action={{ label: 'Open Optimizer', href: '/optimizer' }}
      >
        The Configuration Optimizer can help you fine-tune settings like compression, batching, and delivery guarantees based on your workload.
      </InfoBanner>

      <div className="flex justify-between">
        <button onClick={onBack} className="btn-secondary px-5 py-2.5">
          <ChevronLeft className="w-4 h-4 mr-2" />
          Back
        </button>
        <Link href="/dashboard" className="btn-primary px-5 py-2.5">
          Go to Dashboard
          <ChevronRight className="w-4 h-4 ml-2" />
        </Link>
      </div>
    </div>
  );
}

function SetupSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-40 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-64 bg-primary-dark rounded animate-pulse mt-2" />
      </div>
      <div className="card-dark p-6">
        <div className="h-6 w-48 bg-primary-dark rounded animate-pulse mb-4" />
        <div className="grid grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-24 bg-primary-dark/50 rounded-lg animate-pulse" />
          ))}
        </div>
      </div>
    </div>
  );
}
