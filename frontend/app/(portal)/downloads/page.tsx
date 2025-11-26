'use client';

import { useEffect, useState } from 'react';
import { api, Download } from '@/lib/api';
import { Download as DownloadIcon, ExternalLink, FileCode, Package } from 'lucide-react';

const platforms = [
  { id: 'linux-amd64', name: 'Linux (x64)', icon: '🐧' },
  { id: 'linux-arm64', name: 'Linux (ARM64)', icon: '🐧' },
  { id: 'darwin-amd64', name: 'macOS (Intel)', icon: '🍎' },
  { id: 'darwin-arm64', name: 'macOS (Apple Silicon)', icon: '🍎' },
  { id: 'windows-amd64', name: 'Windows (x64)', icon: '🪟' },
];

export default function DownloadsPage() {
  const [downloads, setDownloads] = useState<Download[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  useEffect(() => {
    async function loadDownloads() {
      const { data } = await api.getDownloads();
      if (data) setDownloads(data.downloads);
      setIsLoading(false);
    }
    loadDownloads();
  }, []);

  async function handleDownload(product: string, version: string, platform: string) {
    const id = `${product}-${version}-${platform}`;
    setDownloadingId(id);

    const { data } = await api.getDownloadURL(product, version, platform);

    if (data?.url) {
      window.location.href = data.url;
    }

    setTimeout(() => setDownloadingId(null), 1000);
  }

  if (isLoading) {
    return <DownloadsSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-h4 text-white">Downloads</h1>
        <p className="text-content-1 text-grey">Download Savegress CDC engine binaries</p>
      </div>

      {downloads.length === 0 ? (
        <div className="card-dark p-12 text-center">
          <Package className="w-16 h-16 mx-auto mb-4 text-grey opacity-50" />
          <p className="text-grey">No downloads available</p>
        </div>
      ) : (
        downloads.map((download) => (
          <div key={`${download.product}-${download.version}`} className="card-dark">
            <div className="p-5 border-b border-cyan-40">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-h5 text-white">{download.product}</h2>
                  <p className="text-sm text-grey">
                    Version {download.version}
                    {download.release_date && (
                      <> • Released {new Date(download.release_date).toLocaleDateString()}</>
                    )}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {download.changelog_url && (
                    <a
                      href={download.changelog_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-accent-cyan hover:text-accent-cyan-bright transition-colors"
                    >
                      <FileCode className="w-4 h-4" />
                      Changelog
                    </a>
                  )}
                </div>
              </div>
            </div>

            <div className="p-5">
              <h3 className="text-sm font-medium text-grey mb-3">Choose your platform</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {platforms
                  .filter((p) => download.platforms.includes(p.id))
                  .map((platform) => {
                    const id = `${download.product}-${download.version}-${platform.id}`;
                    const isDownloading = downloadingId === id;

                    return (
                      <button
                        key={platform.id}
                        onClick={() => handleDownload(download.product, download.version, platform.id)}
                        disabled={isDownloading}
                        className="flex items-center gap-3 p-3 bg-primary-dark/50 border border-cyan-40/30 rounded-lg hover:border-accent-cyan hover:bg-accent-cyan/5 transition-all disabled:opacity-50"
                      >
                        <span className="text-2xl">{platform.icon}</span>
                        <div className="flex-1 text-left">
                          <p className="font-medium text-white">{platform.name}</p>
                          <p className="text-xs text-grey">{platform.id}</p>
                        </div>
                        <DownloadIcon className={`w-5 h-5 text-grey ${isDownloading ? 'animate-bounce' : ''}`} />
                      </button>
                    );
                  })}
              </div>
            </div>

            {/* Edition info */}
            <div className="p-4 border-t border-cyan-40/30 bg-primary-dark/30">
              <p className="text-sm text-grey">
                Available editions:{' '}
                {download.editions.map((edition, i) => (
                  <span key={edition}>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                      edition === 'enterprise'
                        ? 'bg-accent-orange/10 text-accent-orange border-accent-orange/30'
                        : edition === 'pro'
                        ? 'bg-accent-cyan/10 text-accent-cyan border-accent-cyan/30'
                        : 'bg-grey/10 text-grey border-grey/30'
                    }`}>
                      {edition}
                    </span>
                    {i < download.editions.length - 1 && ' '}
                  </span>
                ))}
              </p>
            </div>
          </div>
        ))
      )}

      {/* Documentation */}
      <div className="card-dark p-6">
        <h2 className="text-h5 text-white mb-4">Installation</h2>
        <div className="space-y-4">
          <div>
            <h3 className="text-sm font-medium text-grey mb-2">Linux / macOS</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`# Download and extract
curl -L https://releases.savegress.io/cdc-engine/latest/cdc-engine-linux-amd64.tar.gz | tar xz

# Move to path
sudo mv cdc-engine /usr/local/bin/

# Verify installation
cdc-engine --version`}</code>
            </pre>
          </div>

          <div>
            <h3 className="text-sm font-medium text-grey mb-2">Docker</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`docker pull savegress/cdc-engine:latest

docker run -d \\
  -e LICENSE_KEY=your-license-key \\
  -e DATABASE_URL=postgres://... \\
  savegress/cdc-engine:latest`}</code>
            </pre>
          </div>

          <div>
            <h3 className="text-sm font-medium text-grey mb-2">Helm (Kubernetes)</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`helm repo add savegress https://charts.savegress.io
helm install cdc-engine savegress/cdc-engine \\
  --set licenseKey=your-license-key \\
  --set database.url=postgres://...`}</code>
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}

function DownloadsSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
      </div>
      <div className="card-dark p-6">
        <div className="h-6 w-48 bg-primary-dark rounded animate-pulse mb-4" />
        <div className="grid grid-cols-3 gap-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 bg-primary-dark/50 rounded-lg animate-pulse" />
          ))}
        </div>
      </div>
    </div>
  );
}
