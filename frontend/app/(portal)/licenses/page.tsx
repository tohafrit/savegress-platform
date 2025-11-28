'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, License } from '@/lib/api';
import {
  Key,
  Copy,
  Check,
  Shield,
  Zap,
  Database,
  Clock,
  Server,
  AlertTriangle,
  CreditCard,
  Download,
  ChevronRight,
} from 'lucide-react';

export default function LicensesPage() {
  const [licenses, setLicenses] = useState<License[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  useEffect(() => {
    loadLicenses();
  }, []);

  async function loadLicenses() {
    const { data } = await api.getLicenses();
    if (data) setLicenses(data.licenses);
    setIsLoading(false);
  }

  async function copyLicenseKey(license: License) {
    await navigator.clipboard.writeText(license.key);
    setCopiedId(license.id);
    setTimeout(() => setCopiedId(null), 2000);
  }

  if (isLoading) {
    return <LicensesSkeleton />;
  }

  const activeLicense = licenses.find((l) => l.status === 'active');
  const hasLicense = licenses.length > 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-h4 text-white">Licenses</h1>
        <p className="text-content-1 text-grey">Your CDC engine license and usage</p>
      </div>

      {!hasLicense ? (
        <NoLicenseState />
      ) : activeLicense ? (
        <>
          {/* Active License Card */}
          <div className="card-dark overflow-hidden">
            <div className="p-6 border-b border-cyan-40 bg-gradient-to-r from-accent-cyan/5 to-transparent">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-btn-primary flex items-center justify-center shadow-glow-blue">
                    <Key className="w-7 h-7 text-white" />
                  </div>
                  <div>
                    <div className="flex items-center gap-3">
                      <h2 className="text-h4 text-white capitalize">{activeLicense.edition}</h2>
                      <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400 border border-green-500/40">
                        Active
                      </span>
                    </div>
                    <p className="text-sm text-grey mt-1">
                      Licensed to your account
                    </p>
                  </div>
                </div>

                {activeLicense.edition !== 'enterprise' && (
                  <Link href="/billing" className="btn-secondary px-4 py-2 text-sm">
                    <Zap className="w-4 h-4 mr-2" />
                    Upgrade
                  </Link>
                )}
              </div>
            </div>

            <div className="p-6 space-y-6">
              {/* License Key */}
              <div>
                <label className="block text-sm font-medium text-grey mb-2">License Key</label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-sm font-mono text-accent-cyan bg-primary-dark px-4 py-3 rounded-lg border border-cyan-40/30 overflow-x-auto">
                    {activeLicense.key}
                  </code>
                  <button
                    onClick={() => copyLicenseKey(activeLicense)}
                    className="p-3 bg-primary-dark border border-cyan-40/30 rounded-lg text-grey hover:text-accent-cyan hover:border-accent-cyan transition-colors"
                    title="Copy license key"
                  >
                    {copiedId === activeLicense.id ? (
                      <Check className="w-5 h-5 text-accent-cyan" />
                    ) : (
                      <Copy className="w-5 h-5" />
                    )}
                  </button>
                </div>
                <p className="text-xs text-grey mt-2">
                  Use this key in your CDC engine configuration or as LICENSE_KEY environment variable
                </p>
              </div>

              {/* License Details Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                  <div className="flex items-center gap-2 text-grey mb-1">
                    <Server className="w-4 h-4" />
                    <span className="text-xs">Instances</span>
                  </div>
                  <p className="text-lg font-semibold text-white">
                    {activeLicense.active_instances} / {activeLicense.max_instances === 0 ? '∞' : activeLicense.max_instances}
                  </p>
                </div>

                <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                  <div className="flex items-center gap-2 text-grey mb-1">
                    <Database className="w-4 h-4" />
                    <span className="text-xs">Tables</span>
                  </div>
                  <p className="text-lg font-semibold text-white">
                    {activeLicense.max_tables === 0 ? 'Unlimited' : activeLicense.max_tables}
                  </p>
                </div>

                <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                  <div className="flex items-center gap-2 text-grey mb-1">
                    <Zap className="w-4 h-4" />
                    <span className="text-xs">Throughput</span>
                  </div>
                  <p className="text-lg font-semibold text-white">
                    {activeLicense.max_throughput === 0 ? 'Unlimited' : `${(activeLicense.max_throughput / 1000).toFixed(0)}k/s`}
                  </p>
                </div>

                <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                  <div className="flex items-center gap-2 text-grey mb-1">
                    <Clock className="w-4 h-4" />
                    <span className="text-xs">Expires</span>
                  </div>
                  <p className="text-lg font-semibold text-white">
                    {new Date(activeLicense.expires_at).toLocaleDateString()}
                  </p>
                </div>
              </div>

              {/* Features */}
              {activeLicense.features && activeLicense.features.length > 0 && (
                <div>
                  <label className="block text-sm font-medium text-grey mb-3">Included Features</label>
                  <div className="flex flex-wrap gap-2">
                    {activeLicense.features.slice(0, 12).map((feature) => (
                      <span
                        key={feature}
                        className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/30"
                      >
                        {feature.replace(/_/g, ' ')}
                      </span>
                    ))}
                    {activeLicense.features.length > 12 && (
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-grey/10 text-grey border border-grey/30">
                        +{activeLicense.features.length - 12} more
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Link
              href="/downloads"
              className="card-dark p-5 hover:border-accent-cyan transition-colors group"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-xl bg-accent-cyan/10 flex items-center justify-center group-hover:bg-accent-cyan/20 transition-colors">
                    <Download className="w-6 h-6 text-accent-cyan" />
                  </div>
                  <div>
                    <h3 className="font-medium text-white">Download Engine</h3>
                    <p className="text-sm text-grey">Get the CDC engine for your platform</p>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-grey group-hover:text-accent-cyan transition-colors" />
              </div>
            </Link>

            <Link
              href="/docs"
              className="card-dark p-5 hover:border-accent-cyan transition-colors group"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-xl bg-purple-500/10 flex items-center justify-center group-hover:bg-purple-500/20 transition-colors">
                    <Shield className="w-6 h-6 text-purple-400" />
                  </div>
                  <div>
                    <h3 className="font-medium text-white">Setup Guide</h3>
                    <p className="text-sm text-grey">Configure your license in the engine</p>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-grey group-hover:text-accent-cyan transition-colors" />
              </div>
            </Link>
          </div>

          {/* License History */}
          {licenses.length > 1 && (
            <div className="card-dark overflow-hidden">
              <div className="p-5 border-b border-cyan-40">
                <h2 className="text-h5 text-white">License History</h2>
              </div>
              <div className="divide-y divide-cyan-40/30">
                {licenses.filter((l) => l.id !== activeLicense.id).map((license) => (
                  <div key={license.id} className="p-4 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
                        license.edition === 'enterprise'
                          ? 'bg-purple-500/20 text-purple-400 border-purple-500/40'
                          : license.edition === 'pro'
                          ? 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40'
                          : 'bg-grey/20 text-grey border-grey/40'
                      }`}>
                        {license.edition}
                      </span>
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
                        license.status === 'expired'
                          ? 'bg-accent-orange/20 text-accent-orange border-accent-orange/40'
                          : 'bg-red-500/20 text-red-400 border-red-500/40'
                      }`}>
                        {license.status}
                      </span>
                    </div>
                    <span className="text-sm text-grey">
                      Expired {new Date(license.expires_at).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      ) : (
        <ExpiredLicenseState license={licenses[0]} />
      )}
    </div>
  );
}

function NoLicenseState() {
  return (
    <div className="card-dark p-12 text-center">
      <div className="w-20 h-20 rounded-full bg-primary-dark flex items-center justify-center mx-auto mb-6">
        <Key className="w-10 h-10 text-grey" />
      </div>
      <h3 className="text-h4 text-white mb-3">No Active License</h3>
      <p className="text-grey mb-8 max-w-md mx-auto">
        Choose a plan to get your license key and start using Savegress CDC Engine.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-3xl mx-auto mb-8">
        {/* Community */}
        <div className="card-dark p-6 border-grey/30">
          <h4 className="text-lg font-semibold text-white mb-2">Community</h4>
          <p className="text-2xl font-bold text-white mb-4">Free</p>
          <ul className="text-sm text-grey space-y-2 mb-6 text-left">
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-green-400" />
              PostgreSQL, MySQL, MariaDB
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-green-400" />
              1 source, 10 tables
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-green-400" />
              1,000 events/sec
            </li>
          </ul>
          <Link href="/billing?plan=community" className="btn-secondary w-full py-2.5 text-sm">
            Get Started
          </Link>
        </div>

        {/* Pro */}
        <div className="card-dark p-6 border-accent-cyan relative">
          <div className="absolute -top-3 left-1/2 -translate-x-1/2">
            <span className="px-3 py-1 bg-accent-cyan text-white text-xs font-medium rounded-full">
              Popular
            </span>
          </div>
          <h4 className="text-lg font-semibold text-white mb-2">Pro</h4>
          <p className="text-2xl font-bold text-white mb-4">$99<span className="text-sm text-grey font-normal">/mo</span></p>
          <ul className="text-sm text-grey space-y-2 mb-6 text-left">
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-accent-cyan" />
              + MongoDB, SQL Server, DynamoDB
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-accent-cyan" />
              10 sources, 100 tables
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-accent-cyan" />
              50,000 events/sec
            </li>
          </ul>
          <Link href="/billing?plan=pro" className="btn-primary w-full py-2.5 text-sm">
            Subscribe
          </Link>
        </div>

        {/* Enterprise */}
        <div className="card-dark p-6 border-purple-500/30">
          <h4 className="text-lg font-semibold text-white mb-2">Enterprise</h4>
          <p className="text-2xl font-bold text-white mb-4">$499<span className="text-sm text-grey font-normal">/mo</span></p>
          <ul className="text-sm text-grey space-y-2 mb-6 text-left">
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-purple-400" />
              + Oracle
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-purple-400" />
              Unlimited everything
            </li>
            <li className="flex items-center gap-2">
              <Check className="w-4 h-4 text-purple-400" />
              HA, SSO, Audit logs
            </li>
          </ul>
          <Link href="/billing?plan=enterprise" className="btn-secondary w-full py-2.5 text-sm border-purple-500/50 hover:border-purple-500">
            Subscribe
          </Link>
        </div>
      </div>

      <p className="text-sm text-grey">
        <Link href="/billing" className="text-accent-cyan hover:underline">
          View full plan comparison →
        </Link>
      </p>
    </div>
  );
}

function ExpiredLicenseState({ license }: { license: License }) {
  return (
    <div className="card-dark p-12 text-center">
      <div className="w-20 h-20 rounded-full bg-accent-orange/10 flex items-center justify-center mx-auto mb-6">
        <AlertTriangle className="w-10 h-10 text-accent-orange" />
      </div>
      <h3 className="text-h4 text-white mb-3">License Expired</h3>
      <p className="text-grey mb-2">
        Your <span className="text-white capitalize">{license.edition}</span> license expired on{' '}
        {new Date(license.expires_at).toLocaleDateString()}
      </p>
      <p className="text-grey mb-8 max-w-md mx-auto">
        Renew your subscription to continue using Savegress CDC Engine with all features.
      </p>
      <div className="flex items-center justify-center gap-4">
        <Link href="/billing" className="btn-primary px-8 py-3">
          <CreditCard className="w-4 h-4 mr-2" />
          Renew Subscription
        </Link>
      </div>
    </div>
  );
}

function LicensesSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
      </div>
      <div className="card-dark p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-14 h-14 rounded-2xl bg-primary-dark animate-pulse" />
          <div>
            <div className="h-6 w-24 bg-primary-dark rounded animate-pulse mb-2" />
            <div className="h-4 w-32 bg-primary-dark rounded animate-pulse" />
          </div>
        </div>
        <div className="h-12 w-full bg-primary-dark/50 rounded-lg animate-pulse mb-6" />
        <div className="grid grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-20 bg-primary-dark/50 rounded-lg animate-pulse" />
          ))}
        </div>
      </div>
    </div>
  );
}
