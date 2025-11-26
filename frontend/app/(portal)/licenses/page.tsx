'use client';

import { useEffect, useState } from 'react';
import { api, License } from '@/lib/api';
import { Key, Plus, Copy, Trash2, Check } from 'lucide-react';

export default function LicensesPage() {
  const [licenses, setLicenses] = useState<License[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
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

  async function revokeLicense(id: string) {
    if (!confirm('Are you sure you want to revoke this license?')) return;
    await api.revokeLicense(id);
    loadLicenses();
  }

  if (isLoading) {
    return <LicensesSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-h4 text-white">Licenses</h1>
          <p className="text-content-1 text-grey">Manage your CDC engine licenses</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn-primary px-5 py-3 text-sm"
        >
          <Plus className="w-4 h-4 mr-2" />
          New License
        </button>
      </div>

      {licenses.length === 0 ? (
        <EmptyState onCreateClick={() => setShowCreateModal(true)} />
      ) : (
        <div className="card-dark overflow-hidden">
          <table className="w-full">
            <thead className="bg-primary-dark/50 border-b border-cyan-40">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                  License Key
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                  Edition
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                  Instances
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">
                  Expires
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-grey uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-cyan-40/30">
              {licenses.map((license) => (
                <tr key={license.id} className="hover:bg-primary-dark/30 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <code className="text-sm font-mono text-accent-cyan bg-primary-dark px-2 py-1 rounded">
                        {license.key.slice(0, 12)}...
                      </code>
                      <button
                        onClick={() => copyLicenseKey(license)}
                        className="p-1 text-grey hover:text-accent-cyan transition-colors"
                        title="Copy license key"
                      >
                        {copiedId === license.id ? (
                          <Check className="w-4 h-4 text-accent-cyan" />
                        ) : (
                          <Copy className="w-4 h-4" />
                        )}
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
                        license.edition === 'enterprise'
                          ? 'bg-purple-500/20 text-purple-400 border-purple-500/40'
                          : license.edition === 'pro'
                          ? 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40'
                          : 'bg-grey/20 text-grey border-grey/40'
                      }`}
                    >
                      {license.edition}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
                        license.status === 'active'
                          ? 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40'
                          : license.status === 'expired'
                          ? 'bg-accent-orange/20 text-accent-orange border-accent-orange/40'
                          : 'bg-red-500/20 text-red-400 border-red-500/40'
                      }`}
                    >
                      {license.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-grey">
                    {license.active_instances} / {license.max_instances}
                  </td>
                  <td className="px-4 py-3 text-sm text-grey">
                    {new Date(license.expires_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {license.status === 'active' && (
                      <button
                        onClick={() => revokeLicense(license.id)}
                        className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                        title="Revoke license"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreateModal && (
        <CreateLicenseModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            loadLicenses();
          }}
        />
      )}
    </div>
  );
}

function EmptyState({ onCreateClick }: { onCreateClick: () => void }) {
  return (
    <div className="card-dark p-12 text-center">
      <Key className="w-16 h-16 mx-auto mb-4 text-grey opacity-50" />
      <h3 className="text-h5 text-white mb-2">No licenses yet</h3>
      <p className="text-grey mb-6 max-w-md mx-auto">
        Create a license to start using Savegress CDC Engine.
      </p>
      <button onClick={onCreateClick} className="btn-primary px-6 py-3">
        <Plus className="w-4 h-4 mr-2" />
        Create License
      </button>
    </div>
  );
}

function CreateLicenseModal({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const [edition, setEdition] = useState('pro');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleCreate() {
    setIsLoading(true);
    setError('');

    const { error } = await api.createLicense({ edition });

    if (error) {
      setError(error);
      setIsLoading(false);
    } else {
      onCreated();
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="card-dark w-full max-w-md mx-4">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Create New License</h2>
        </div>

        <div className="p-5 space-y-4">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-grey mb-2">
              Edition
            </label>
            <select
              value={edition}
              onChange={(e) => setEdition(e.target.value)}
              className="input-field"
            >
              <option value="pro">Pro</option>
              <option value="enterprise">Enterprise</option>
            </select>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              onClick={onClose}
              className="btn-secondary px-5 py-2.5 text-sm"
            >
              Cancel
            </button>
            <button
              onClick={handleCreate}
              disabled={isLoading}
              className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
            >
              {isLoading ? 'Creating...' : 'Create License'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function LicensesSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
          <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
        </div>
        <div className="h-10 w-36 bg-primary-dark rounded-[20px] animate-pulse" />
      </div>
      <div className="card-dark p-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="py-4 border-b border-cyan-40/30 last:border-0">
            <div className="h-5 w-40 bg-primary-dark rounded animate-pulse mb-2" />
            <div className="h-4 w-24 bg-primary-dark rounded animate-pulse" />
          </div>
        ))}
      </div>
    </div>
  );
}
