'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, Download, License } from '@/lib/api';
import {
  Download as DownloadIcon,
  ExternalLink,
  FileCode,
  Package,
  Key,
  Lock,
  Check,
  Zap,
} from 'lucide-react';

const platforms = [
  { id: 'linux-amd64', name: 'Linux (x64)', icon: '🐧' },
  { id: 'linux-arm64', name: 'Linux (ARM64)', icon: '🐧' },
  { id: 'darwin-amd64', name: 'macOS (Intel)', icon: '🍎' },
  { id: 'darwin-arm64', name: 'macOS (Apple Silicon)', icon: '🍎' },
  { id: 'windows-amd64', name: 'Windows (x64)', icon: '🪟' },
];

const editionInfo: Record<string, { name: string; color: string; bgColor: string; borderColor: string }> = {
  community: {
    name: 'Community',
    color: 'text-green-400',
    bgColor: 'bg-green-500/10',
    borderColor: 'border-green-500/30',
  },
  pro: {
    name: 'Pro',
    color: 'text-accent-cyan',
    bgColor: 'bg-accent-cyan/10',
    borderColor: 'border-accent-cyan/30',
  },
  enterprise: {
    name: 'Enterprise',
    color: 'text-purple-400',
    bgColor: 'bg-purple-500/10',
    borderColor: 'border-purple-500/30',
  },
};

export default function DownloadsPage() {
  const [downloads, setDownloads] = useState<Download[]>([]);
  const [license, setLicense] = useState<License | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  useEffect(() => {
    async function loadData() {
      const [downloadsRes, licensesRes] = await Promise.all([
        api.getDownloads(),
        api.getLicenses(),
      ]);

      if (downloadsRes.data) setDownloads(downloadsRes.data.downloads);
      if (licensesRes.data) {
        const activeLicense = licensesRes.data.licenses.find((l: License) => l.status === 'active');
        setLicense(activeLicense || null);
      }
      setIsLoading(false);
    }
    loadData();
  }, []);

  async function handleDownload(product: string, version: string, platform: string, edition: string) {
    const id = `${product}-${version}-${platform}-${edition}`;
    setDownloadingId(id);

    const { data } = await api.getDownloadURL(product, version, platform, edition);

    if (data?.url) {
      window.location.href = data.url;
    }

    setTimeout(() => setDownloadingId(null), 1000);
  }

  function canDownloadEdition(edition: string): boolean {
    if (!license) return edition === 'community';

    const tierHierarchy: Record<string, number> = {
      community: 0,
      pro: 1,
      enterprise: 2,
    };

    const userTier = tierHierarchy[license.edition] ?? 0;
    const editionTier = tierHierarchy[edition] ?? 0;

    return userTier >= editionTier;
  }

  if (isLoading) {
    return <DownloadsSkeleton />;
  }

  // No license at all
  if (!license) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-h4 text-white">Downloads</h1>
          <p className="text-content-1 text-grey">Download Savegress CDC engine binaries</p>
        </div>

        <NoLicenseState />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-h4 text-white">Downloads</h1>
          <p className="text-content-1 text-grey">Download Savegress CDC engine binaries</p>
        </div>

        {/* Current license badge */}
        <div className={`flex items-center gap-2 px-4 py-2 rounded-full ${editionInfo[license.edition]?.bgColor} ${editionInfo[license.edition]?.borderColor} border`}>
          <Key className={`w-4 h-4 ${editionInfo[license.edition]?.color}`} />
          <span className={`text-sm font-medium ${editionInfo[license.edition]?.color}`}>
            {editionInfo[license.edition]?.name} License
          </span>
        </div>
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

            {/* Editions */}
            <div className="p-5 space-y-6">
              {download.editions.map((edition) => {
                const canDownload = canDownloadEdition(edition);
                const info = editionInfo[edition];

                return (
                  <div key={edition} className={`rounded-xl border ${canDownload ? info?.borderColor : 'border-grey/20'} overflow-hidden`}>
                    {/* Edition Header */}
                    <div className={`p-4 ${canDownload ? info?.bgColor : 'bg-grey/5'} flex items-center justify-between`}>
                      <div className="flex items-center gap-3">
                        <span className={`text-lg font-semibold ${canDownload ? info?.color : 'text-grey'}`}>
                          {info?.name} Edition
                        </span>
                        {canDownload ? (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-green-500/20 text-green-400">
                            <Check className="w-3 h-3" />
                            Included
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-grey/20 text-grey">
                            <Lock className="w-3 h-3" />
                            Upgrade required
                          </span>
                        )}
                      </div>

                      {!canDownload && (
                        <Link
                          href="/billing"
                          className="inline-flex items-center gap-1 text-sm text-accent-cyan hover:underline"
                        >
                          <Zap className="w-4 h-4" />
                          Upgrade to {edition === 'enterprise' ? 'Enterprise' : 'Pro'}
                        </Link>
                      )}
                    </div>

                    {/* Platforms */}
                    <div className="p-4">
                      {canDownload ? (
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                          {platforms
                            .filter((p) => download.platforms.includes(p.id))
                            .map((platform) => {
                              const id = `${download.product}-${download.version}-${platform.id}-${edition}`;
                              const isDownloading = downloadingId === id;

                              return (
                                <button
                                  key={platform.id}
                                  onClick={() => handleDownload(download.product, download.version, platform.id, edition)}
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
                      ) : (
                        <div className="text-center py-6 text-grey">
                          <Lock className="w-8 h-8 mx-auto mb-2 opacity-50" />
                          <p className="text-sm">
                            {edition === 'enterprise'
                              ? 'Enterprise features: Oracle, HA cluster, SSO, audit logs'
                              : 'Pro features: MongoDB, SQL Server, Kafka output, compression'}
                          </p>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ))
      )}

      {/* License Embedded Notice */}
      <div className="card-dark p-5 border-green-500/30 bg-green-500/5">
        <div className="flex items-start gap-4">
          <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center flex-shrink-0">
            <Check className="w-5 h-5 text-green-400" />
          </div>
          <div>
            <h3 className="font-medium text-white mb-1">License Embedded</h3>
            <p className="text-sm text-grey">
              Your license key is automatically embedded in downloaded binaries.
              No manual configuration needed - just download and run!
            </p>
          </div>
        </div>
      </div>

      {/* Installation Guide */}
      <div className="card-dark p-6">
        <h2 className="text-h5 text-white mb-4">Quick Start</h2>
        <div className="space-y-4">
          <div>
            <h3 className="text-sm font-medium text-grey mb-2">One-line Install (Linux / macOS)</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`curl -fsSL https://savegress.io/install | bash`}</code>
            </pre>
            <p className="text-xs text-grey mt-2">This script downloads and installs the CDC engine with your license already embedded.</p>
          </div>

          <div>
            <h3 className="text-sm font-medium text-grey mb-2">Manual Download</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`# Download from portal (license embedded)
# Just click the download button above!

# Make executable
chmod +x cdc-engine

# Move to path
sudo mv cdc-engine /usr/local/bin/

# Verify - license is already configured!
cdc-engine --version
cdc-engine license status`}</code>
            </pre>
          </div>

          <div>
            <h3 className="text-sm font-medium text-grey mb-2">Docker</h3>
            <pre className="bg-dark-bg border border-cyan-40/30 p-4 rounded-lg text-sm text-grey overflow-x-auto font-mono">
              <code>{`# For Docker, use your license key as environment variable:
docker run -d \\
  -e SAVEGRESS_LICENSE_KEY="${license.key}" \\
  -v $(pwd)/config.yaml:/etc/savegress/config.yaml \\
  savegress/cdc-engine:latest`}</code>
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}

function NoLicenseState() {
  return (
    <div className="card-dark p-12 text-center">
      <div className="w-20 h-20 rounded-full bg-primary-dark flex items-center justify-center mx-auto mb-6">
        <Key className="w-10 h-10 text-grey" />
      </div>
      <h3 className="text-h4 text-white mb-3">License Required</h3>
      <p className="text-grey mb-8 max-w-md mx-auto">
        You need an active license to download Savegress CDC Engine.
        Choose a plan to get started.
      </p>

      <div className="flex items-center justify-center gap-4">
        <Link href="/licenses" className="btn-primary px-8 py-3">
          <Key className="w-4 h-4 mr-2" />
          Get a License
        </Link>
        <Link href="/billing" className="btn-secondary px-8 py-3">
          View Plans
        </Link>
      </div>
    </div>
  );
}

function DownloadsSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
          <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
        </div>
        <div className="h-10 w-40 bg-primary-dark rounded-full animate-pulse" />
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
