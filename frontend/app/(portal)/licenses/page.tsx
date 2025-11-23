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
          <h1 className="text-2xl font-bold text-primary">Licenses</h1>
          <p className="text-neutral-dark-gray">Manage your CDC engine licenses</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark"
        >
          <Plus className="w-4 h-4" />
          New License
        </button>
      </div>

      {licenses.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <Key className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p className="text-neutral-dark-gray">No licenses yet</p>
          <p className="text-sm text-neutral-dark-gray mt-1">
            Create a license to start using Savegress CDC
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark"
          >
            <Plus className="w-4 h-4" />
            Create License
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  License Key
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  Edition
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  Instances
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  Expires
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-neutral-dark-gray uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {licenses.map((license) => (
                <tr key={license.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <code className="text-sm font-mono text-primary bg-gray-100 px-2 py-1 rounded">
                        {license.key.slice(0, 12)}...
                      </code>
                      <button
                        onClick={() => copyLicenseKey(license)}
                        className="p-1 text-neutral-dark-gray hover:text-primary"
                        title="Copy license key"
                      >
                        {copiedId === license.id ? (
                          <Check className="w-4 h-4 text-green-600" />
                        ) : (
                          <Copy className="w-4 h-4" />
                        )}
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                        license.edition === 'enterprise'
                          ? 'bg-purple-100 text-purple-700'
                          : license.edition === 'pro'
                          ? 'bg-blue-100 text-blue-700'
                          : 'bg-gray-100 text-gray-700'
                      }`}
                    >
                      {license.edition}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                        license.status === 'active'
                          ? 'bg-green-100 text-green-700'
                          : license.status === 'expired'
                          ? 'bg-yellow-100 text-yellow-700'
                          : 'bg-red-100 text-red-700'
                      }`}
                    >
                      {license.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-neutral-dark-gray">
                    {license.active_instances} / {license.max_instances}
                  </td>
                  <td className="px-4 py-3 text-sm text-neutral-dark-gray">
                    {new Date(license.expires_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {license.status === 'active' && (
                      <button
                        onClick={() => revokeLicense(license.id)}
                        className="p-1 text-red-600 hover:text-red-800"
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
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
        <h2 className="text-xl font-bold text-primary mb-4">Create New License</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-600 rounded-md text-sm">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-neutral-dark-gray mb-2">
              Edition
            </label>
            <select
              value={edition}
              onChange={(e) => setEdition(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="pro">Pro</option>
              <option value="enterprise">Enterprise</option>
            </select>
          </div>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <button
            onClick={onClose}
            className="px-4 py-2 text-neutral-dark-gray hover:bg-gray-100 rounded-md"
          >
            Cancel
          </button>
          <button
            onClick={handleCreate}
            disabled={isLoading}
            className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark disabled:opacity-50"
          >
            {isLoading ? 'Creating...' : 'Create License'}
          </button>
        </div>
      </div>
    </div>
  );
}

function LicensesSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-gray-200 rounded animate-pulse" />
        <div className="h-4 w-48 bg-gray-200 rounded animate-pulse mt-2" />
      </div>
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-12 bg-gray-100 rounded animate-pulse mb-2" />
        ))}
      </div>
    </div>
  );
}
