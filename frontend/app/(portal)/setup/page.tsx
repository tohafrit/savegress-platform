'use client';

import { useEffect, useState } from 'react';
import { api, Connection, Pipeline } from '@/lib/api';
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
} from 'lucide-react';

type Step = 'method' | 'source' | 'deploy';

const INSTALLATION_METHODS = [
  { id: 'docker-compose', name: 'Docker Compose', icon: Box, description: 'Recommended for local and single-server deployments' },
  { id: 'helm', name: 'Kubernetes (Helm)', icon: Cloud, description: 'For Kubernetes clusters with Helm' },
  { id: 'env', name: 'Environment File', icon: FileCode, description: 'Environment variables for any deployment' },
  { id: 'systemd', name: 'Systemd Service', icon: Terminal, description: 'For Linux servers with systemd' },
];

export default function SetupPage() {
  const [step, setStep] = useState<Step>('method');
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-h4 text-white">Setup Wizard</h1>
        <p className="text-content-1 text-grey">Generate deployment configuration for Savegress CDC Engine</p>
      </div>

      {/* Progress Steps */}
      <div className="card-dark p-4">
        <div className="flex items-center justify-between">
          {['method', 'source', 'deploy'].map((s, i) => (
            <div key={s} className="flex items-center">
              <div className={`flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium transition-colors ${
                step === s ? 'bg-accent-cyan text-white' :
                ['method', 'source', 'deploy'].indexOf(step) > i ? 'bg-accent-cyan text-white' :
                'bg-primary-dark text-grey'
              }`}>
                {['method', 'source', 'deploy'].indexOf(step) > i ? <Check className="w-4 h-4" /> : i + 1}
              </div>
              <span className={`ml-2 text-sm ${step === s ? 'text-white font-medium' : 'text-grey'}`}>
                {s === 'method' ? 'Installation Method' : s === 'source' ? 'Select Pipeline' : 'Deploy'}
              </span>
              {i < 2 && <ChevronRight className="w-4 h-4 mx-4 text-grey/50" />}
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
  onNext
}: {
  method: string;
  setMethod: (m: string) => void;
  onNext: () => void;
}) {
  return (
    <div>
      <h2 className="text-h5 text-white mb-4">Choose Installation Method</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {INSTALLATION_METHODS.map((m) => (
          <button
            key={m.id}
            onClick={() => setMethod(m.id)}
            className={`flex items-start gap-4 p-4 rounded-lg text-left transition-all ${
              method === m.id
                ? 'bg-accent-cyan/10 border border-accent-cyan'
                : 'bg-primary-dark/50 border border-cyan-40/30 hover:border-cyan-40'
            }`}
          >
            <div className={`p-2 rounded-lg ${method === m.id ? 'bg-accent-cyan text-white' : 'bg-primary-dark text-grey'}`}>
              <m.icon className="w-6 h-6" />
            </div>
            <div>
              <h3 className="font-medium text-white">{m.name}</h3>
              <p className="text-sm text-grey mt-1">{m.description}</p>
            </div>
          </button>
        ))}
      </div>
      <div className="flex justify-end">
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
    <div>
      <h2 className="text-h5 text-white mb-2">Select Pipeline (Optional)</h2>
      <p className="text-grey mb-4">
        Select a pipeline to pre-fill the configuration, or skip to generate a generic template.
      </p>

      <div className="space-y-2 mb-6">
        <button
          onClick={() => setSelectedPipeline(null)}
          className={`flex items-center gap-3 w-full p-3 rounded-lg text-left transition-all ${
            selectedPipeline === null
              ? 'bg-accent-cyan/10 border border-accent-cyan'
              : 'bg-primary-dark/50 border border-cyan-40/30 hover:border-cyan-40'
          }`}
        >
          <div className={`w-4 h-4 rounded-full border-2 flex items-center justify-center ${
            selectedPipeline === null ? 'border-accent-cyan bg-accent-cyan' : 'border-grey'
          }`}>
            {selectedPipeline === null && <Check className="w-3 h-3 text-white" />}
          </div>
          <div>
            <span className="font-medium text-white">Generic Template</span>
            <span className="text-sm text-grey ml-2">- No pre-filled values</span>
          </div>
        </button>

        {pipelines.map((pipeline) => (
          <button
            key={pipeline.id}
            onClick={() => setSelectedPipeline(pipeline.id)}
            className={`flex items-center gap-3 w-full p-3 rounded-lg text-left transition-all ${
              selectedPipeline === pipeline.id
                ? 'bg-accent-cyan/10 border border-accent-cyan'
                : 'bg-primary-dark/50 border border-cyan-40/30 hover:border-cyan-40'
            }`}
          >
            <div className={`w-4 h-4 rounded-full border-2 flex items-center justify-center ${
              selectedPipeline === pipeline.id ? 'border-accent-cyan bg-accent-cyan' : 'border-grey'
            }`}>
              {selectedPipeline === pipeline.id && <Check className="w-3 h-3 text-white" />}
            </div>
            <div className="flex-1">
              <span className="font-medium text-white">{pipeline.name}</span>
              <div className="flex items-center gap-2 text-sm text-grey mt-1">
                <Database className="w-3 h-3" />
                <span>{pipeline.source_connection?.type || 'Unknown'}</span>
                <span>→</span>
                <span>{pipeline.target_type}</span>
              </div>
            </div>
          </button>
        ))}

        {pipelines.length === 0 && (
          <p className="text-sm text-grey p-3 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
            No pipelines configured yet. A generic template will be generated.
          </p>
        )}
      </div>

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
    <div>
      <h2 className="text-h5 text-white mb-4">Your Configuration</h2>

      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          {methodInfo && <methodInfo.icon className="w-5 h-5 text-accent-cyan" />}
          <span className="font-medium text-white">{methodInfo?.name}</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onCopy}
            className="btn-secondary px-4 py-2 text-sm"
          >
            {copied ? <Check className="w-4 h-4 mr-2 text-accent-cyan" /> : <Copy className="w-4 h-4 mr-2" />}
            {copied ? 'Copied!' : 'Copy'}
          </button>
          <button onClick={onDownload} className="btn-primary px-4 py-2 text-sm">
            <Download className="w-4 h-4 mr-2" />
            Download
          </button>
        </div>
      </div>

      <div className="bg-dark-bg rounded-lg overflow-hidden mb-6 border border-cyan-40/30">
        <pre className="p-4 text-sm text-grey overflow-x-auto max-h-[400px] overflow-y-auto font-mono">
          <code>{config}</code>
        </pre>
      </div>

      {/* Next Steps */}
      <div className="bg-accent-cyan/10 border border-accent-cyan/30 rounded-lg p-4 mb-6">
        <h3 className="font-medium text-white mb-2">Next Steps</h3>
        <ol className="list-decimal list-inside text-sm text-grey space-y-1">
          {method === 'docker-compose' && (
            <>
              <li>Save as <code className="bg-primary-dark px-1 rounded text-accent-cyan">docker-compose.yml</code></li>
              <li>Set environment variable: <code className="bg-primary-dark px-1 rounded text-accent-cyan">export SOURCE_DB_PASSWORD=your_password</code></li>
              <li>Run: <code className="bg-primary-dark px-1 rounded text-accent-cyan">docker-compose up -d</code></li>
              <li>Check logs: <code className="bg-primary-dark px-1 rounded text-accent-cyan">docker-compose logs -f</code></li>
            </>
          )}
          {method === 'helm' && (
            <>
              <li>Save as <code className="bg-primary-dark px-1 rounded text-accent-cyan">values.yaml</code></li>
              <li>Create secret for database password</li>
              <li>Run: <code className="bg-primary-dark px-1 rounded text-accent-cyan">helm install cdc-engine savegress/cdc-engine -f values.yaml</code></li>
              <li>Check status: <code className="bg-primary-dark px-1 rounded text-accent-cyan">kubectl get pods</code></li>
            </>
          )}
          {method === 'env' && (
            <>
              <li>Save as <code className="bg-primary-dark px-1 rounded text-accent-cyan">savegress.env</code></li>
              <li>Download the binary from the Downloads page</li>
              <li>Run: <code className="bg-primary-dark px-1 rounded text-accent-cyan">source savegress.env && ./cdc-engine</code></li>
            </>
          )}
          {method === 'systemd' && (
            <>
              <li>Save as <code className="bg-primary-dark px-1 rounded text-accent-cyan">/etc/systemd/system/savegress.service</code></li>
              <li>Create config at <code className="bg-primary-dark px-1 rounded text-accent-cyan">/etc/savegress/savegress.env</code></li>
              <li>Run: <code className="bg-primary-dark px-1 rounded text-accent-cyan">systemctl enable savegress && systemctl start savegress</code></li>
            </>
          )}
        </ol>
      </div>

      <div className="flex justify-between">
        <button onClick={onBack} className="btn-secondary px-5 py-2.5">
          <ChevronLeft className="w-4 h-4 mr-2" />
          Back
        </button>
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
