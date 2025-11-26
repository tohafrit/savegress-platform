'use client';

import { useEffect, useState } from 'react';
import { api, Connection } from '@/lib/api';
import {
  Plus,
  Database,
  CheckCircle,
  XCircle,
  Clock,
  Trash2,
  RefreshCw,
  Eye,
  EyeOff
} from 'lucide-react';

const DATABASE_TYPES = [
  { value: 'postgres', label: 'PostgreSQL', port: 5432 },
  { value: 'mysql', label: 'MySQL', port: 3306 },
  { value: 'mariadb', label: 'MariaDB', port: 3306 },
  { value: 'mongodb', label: 'MongoDB', port: 27017 },
  { value: 'sqlserver', label: 'SQL Server', port: 1433 },
  { value: 'oracle', label: 'Oracle', port: 1521 },
  { value: 'cassandra', label: 'Cassandra', port: 9042 },
  { value: 'dynamodb', label: 'DynamoDB', port: 443 },
];

export default function ConnectionsPage() {
  const [connections, setConnections] = useState<Connection[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);

  useEffect(() => {
    loadConnections();
  }, []);

  async function loadConnections() {
    const { data } = await api.getConnections();
    if (data) setConnections(data.connections);
    setIsLoading(false);
  }

  async function testConnection(id: string) {
    setTestingId(id);
    await api.testConnection(id);
    await loadConnections();
    setTestingId(null);
  }

  async function deleteConnection(id: string) {
    if (!confirm('Are you sure you want to delete this connection?')) return;
    await api.deleteConnection(id);
    loadConnections();
  }

  if (isLoading) {
    return <ConnectionsSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-h4 text-white">Connections</h1>
          <p className="text-content-1 text-grey">Manage your database connections</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn-primary px-5 py-3 text-sm"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Connection
        </button>
      </div>

      {connections.length === 0 ? (
        <EmptyState onCreateClick={() => setShowCreateModal(true)} />
      ) : (
        <div className="card-dark overflow-hidden">
          <table className="w-full">
            <thead className="bg-primary-dark/50 border-b border-cyan-40">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Host</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Database</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-grey uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-grey uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-cyan-40/30">
              {connections.map((conn) => (
                <tr key={conn.id} className="hover:bg-primary-dark/30 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4 text-accent-cyan" />
                      <span className="font-medium text-white">{conn.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm text-grey">
                    {DATABASE_TYPES.find(t => t.value === conn.type)?.label || conn.type}
                  </td>
                  <td className="px-4 py-3 text-sm text-grey">
                    {conn.host}:{conn.port}
                  </td>
                  <td className="px-4 py-3 text-sm text-grey">{conn.database}</td>
                  <td className="px-4 py-3">
                    <TestStatus status={conn.test_status} />
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => testConnection(conn.id)}
                        disabled={testingId === conn.id}
                        className="p-2 text-grey hover:text-accent-cyan hover:bg-primary-dark rounded-lg transition-colors disabled:opacity-50"
                        title="Test connection"
                      >
                        <RefreshCw className={`w-4 h-4 ${testingId === conn.id ? 'animate-spin' : ''}`} />
                      </button>
                      <button
                        onClick={() => deleteConnection(conn.id)}
                        className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                        title="Delete connection"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreateModal && (
        <CreateConnectionModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            loadConnections();
          }}
        />
      )}
    </div>
  );
}

function TestStatus({ status }: { status?: string }) {
  if (!status) {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-grey">
        <Clock className="w-3 h-3" />
        Not tested
      </span>
    );
  }
  if (status === 'success') {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-accent-cyan">
        <CheckCircle className="w-3 h-3" />
        Connected
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-xs text-red-400">
      <XCircle className="w-3 h-3" />
      Failed
    </span>
  );
}

function EmptyState({ onCreateClick }: { onCreateClick: () => void }) {
  return (
    <div className="card-dark p-12 text-center">
      <Database className="w-16 h-16 mx-auto mb-4 text-grey opacity-50" />
      <h3 className="text-h5 text-white mb-2">No connections yet</h3>
      <p className="text-grey mb-6 max-w-md mx-auto">
        Add your first database connection to start creating pipelines.
      </p>
      <button onClick={onCreateClick} className="btn-primary px-6 py-3">
        <Plus className="w-4 h-4 mr-2" />
        Add Connection
      </button>
    </div>
  );
}

function CreateConnectionModal({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState('');
  const [type, setType] = useState('postgres');
  const [host, setHost] = useState('');
  const [port, setPort] = useState(5432);
  const [database, setDatabase] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [sslMode, setSslMode] = useState('prefer');
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [error, setError] = useState('');

  function handleTypeChange(newType: string) {
    setType(newType);
    const dbType = DATABASE_TYPES.find(t => t.value === newType);
    if (dbType) {
      setPort(dbType.port);
    }
  }

  async function testConnection() {
    setIsTesting(true);
    setTestResult(null);
    const { data, error } = await api.testConnectionDirect({
      type: type as Connection['type'],
      host,
      port,
      database,
      username,
      password,
      ssl_mode: sslMode,
    });
    if (error) {
      setTestResult({ success: false, message: error });
    } else if (data) {
      setTestResult(data);
    }
    setIsTesting(false);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsSubmitting(true);
    setError('');

    const { error: apiError } = await api.createConnection({
      name,
      type: type as Connection['type'],
      host,
      port,
      database,
      username,
      password,
      ssl_mode: sslMode,
    });

    if (apiError) {
      setError(apiError);
      setIsSubmitting(false);
      return;
    }

    onCreated();
  }

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="card-dark w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Add Database Connection</h2>
        </div>
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Connection Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="input-field"
              placeholder="production-db"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Database Type</label>
            <select
              value={type}
              onChange={(e) => handleTypeChange(e.target.value)}
              className="input-field"
            >
              {DATABASE_TYPES.map((db) => (
                <option key={db.value} value={db.value}>{db.label}</option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2">
              <label className="block text-sm font-medium text-grey mb-2">Host</label>
              <input
                type="text"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                required
                className="input-field"
                placeholder="db.example.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-grey mb-2">Port</label>
              <input
                type="number"
                value={port}
                onChange={(e) => setPort(parseInt(e.target.value))}
                required
                className="input-field"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Database</label>
            <input
              type="text"
              value={database}
              onChange={(e) => setDatabase(e.target.value)}
              required
              className="input-field"
              placeholder="myapp"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              className="input-field"
              placeholder="replication_user"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">Password</label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="input-field pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-grey hover:text-white transition-colors"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-grey mb-2">SSL Mode</label>
            <select
              value={sslMode}
              onChange={(e) => setSslMode(e.target.value)}
              className="input-field"
            >
              <option value="disable">Disable</option>
              <option value="prefer">Prefer</option>
              <option value="require">Require</option>
              <option value="verify-ca">Verify CA</option>
              <option value="verify-full">Verify Full</option>
            </select>
          </div>

          {/* Test Result */}
          {testResult && (
            <div className={`p-3 rounded-lg border ${testResult.success ? 'bg-accent-cyan/10 border-accent-cyan/30' : 'bg-red-500/10 border-red-500/30'}`}>
              <div className="flex items-center gap-2">
                {testResult.success ? (
                  <CheckCircle className="w-4 h-4 text-accent-cyan" />
                ) : (
                  <XCircle className="w-4 h-4 text-red-400" />
                )}
                <span className={`text-sm ${testResult.success ? 'text-accent-cyan' : 'text-red-400'}`}>
                  {testResult.message}
                </span>
              </div>
            </div>
          )}

          <div className="flex justify-between pt-4">
            <button
              type="button"
              onClick={testConnection}
              disabled={isTesting || !host || !port || !database || !username || !password}
              className="btn-secondary px-5 py-2.5 text-sm disabled:opacity-50"
            >
              {isTesting ? 'Testing...' : 'Test Connection'}
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
                disabled={isSubmitting}
                className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
              >
                {isSubmitting ? 'Saving...' : 'Save Connection'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}

function ConnectionsSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
          <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
        </div>
        <div className="h-10 w-40 bg-primary-dark rounded-[20px] animate-pulse" />
      </div>
      <div className="card-dark p-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="py-4 border-b border-cyan-40/30 last:border-0">
            <div className="h-5 w-32 bg-primary-dark rounded animate-pulse mb-2" />
            <div className="h-4 w-48 bg-primary-dark rounded animate-pulse" />
          </div>
        ))}
      </div>
    </div>
  );
}
