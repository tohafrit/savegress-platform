'use client';

import { useEffect, useState } from 'react';
import { api, Connection } from '@/lib/api';
import {
  PageHeader,
  InfoBanner,
  EmptyStateWithGuide,
  FormFieldWithHelp,
  HelpIcon,
  ExpandableSection,
  QuickGuide,
} from '@/components/ui/helpers';
import {
  Plus,
  Database,
  CheckCircle,
  XCircle,
  Clock,
  Trash2,
  RefreshCw,
  Eye,
  EyeOff,
  Shield,
  Server,
  Key,
  AlertTriangle,
  Info,
  Zap,
  Lock,
} from 'lucide-react';

const DATABASE_TYPES = [
  {
    value: 'postgres',
    label: 'PostgreSQL',
    port: 5432,
    description: 'Most popular open-source database. Full CDC support via logical replication.',
    requirements: 'Requires PostgreSQL 10+ with logical replication enabled',
    available: true,
  },
  {
    value: 'mysql',
    label: 'MySQL',
    port: 3306,
    description: 'Widely used relational database. CDC via binlog.',
    requirements: 'Requires MySQL 5.6+ with binlog enabled',
    available: false,
  },
  {
    value: 'mariadb',
    label: 'MariaDB',
    port: 3306,
    description: 'MySQL-compatible database with enhanced features.',
    requirements: 'Requires MariaDB 10.0+ with binlog enabled',
    available: false,
  },
  {
    value: 'mongodb',
    label: 'MongoDB',
    port: 27017,
    description: 'Document database. CDC via change streams.',
    requirements: 'Requires MongoDB 4.0+ replica set',
    available: false,
  },
  {
    value: 'sqlserver',
    label: 'SQL Server',
    port: 1433,
    description: 'Microsoft enterprise database. CDC via SQL Server CDC.',
    requirements: 'Requires SQL Server 2016+ with CDC enabled',
    available: false,
  },
  {
    value: 'oracle',
    label: 'Oracle',
    port: 1521,
    description: 'Enterprise database. CDC via LogMiner.',
    requirements: 'Requires Oracle 12c+ with LogMiner',
    available: false,
  },
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
    if (!confirm('Are you sure you want to delete this connection? Any pipelines using this connection will stop working.')) return;
    await api.deleteConnection(id);
    loadConnections();
  }

  if (isLoading) {
    return <ConnectionsSkeleton />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Connections"
        description="A connection stores the credentials to access your source database. Savegress uses these credentials to read the change log and capture every INSERT, UPDATE, and DELETE."
        tip="Your credentials are encrypted and never shared. We only read from the database - never write!"
        action={
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn-primary px-5 py-3 text-sm"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Connection
          </button>
        }
      />

      {connections.length === 0 ? (
        <EmptyStateWithGuide
          icon={Database}
          title="No connections yet"
          description="Add your first database connection to start capturing changes. This is the first step to setting up CDC."
          guide={{
            title: 'Getting Started with Connections',
            steps: [
              'Gather your database credentials (host, port, username, password)',
              'Make sure your database has logical replication enabled (for PostgreSQL)',
              'Create a dedicated replication user with read-only permissions',
              'Test the connection before creating pipelines',
            ],
          }}
          action={{ label: 'Add Connection', onClick: () => setShowCreateModal(true) }}
        />
      ) : (
        <>
          {/* Connection summary */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <SummaryCard
              label="Total Connections"
              value={connections.length}
              icon={Database}
              help="Number of database connections configured"
            />
            <SummaryCard
              label="Connected"
              value={connections.filter(c => c.test_status === 'success').length}
              icon={CheckCircle}
              color="text-green-400"
              help="Connections successfully tested and ready to use"
            />
            <SummaryCard
              label="Failed"
              value={connections.filter(c => c.test_status === 'failed').length}
              icon={XCircle}
              color="text-red-400"
              help="Connections that failed the last test - check credentials"
            />
            <SummaryCard
              label="Not Tested"
              value={connections.filter(c => !c.test_status).length}
              icon={Clock}
              color="text-grey"
              help="Connections that haven't been tested yet"
            />
          </div>

          {/* Connections table */}
          <div className="card-dark overflow-hidden">
            <table className="w-full">
              <thead className="bg-primary-dark/50 border-b border-cyan-40">
                <tr>
                  <th className="px-4 py-3 text-left">
                    <div className="flex items-center gap-2 text-xs font-medium text-grey uppercase tracking-wider">
                      Name
                      <HelpIcon text="A friendly name to identify this connection" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left">
                    <div className="flex items-center gap-2 text-xs font-medium text-grey uppercase tracking-wider">
                      Type
                      <HelpIcon text="The database engine type (PostgreSQL, MySQL, etc.)" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left">
                    <div className="flex items-center gap-2 text-xs font-medium text-grey uppercase tracking-wider">
                      Host
                      <HelpIcon text="Database server address and port" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left">
                    <div className="flex items-center gap-2 text-xs font-medium text-grey uppercase tracking-wider">
                      Database
                      <HelpIcon text="The specific database to connect to" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-left">
                    <div className="flex items-center gap-2 text-xs font-medium text-grey uppercase tracking-wider">
                      Status
                      <HelpIcon text="Connection test result - green means ready to use" />
                    </div>
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-grey uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyan-40/30">
                {connections.map((conn) => (
                  <tr key={conn.id} className="hover:bg-primary-dark/30 transition-colors group">
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
                          title="Test connection - verify credentials work"
                        >
                          <RefreshCw className={`w-4 h-4 ${testingId === conn.id ? 'animate-spin' : ''}`} />
                        </button>
                        <button
                          onClick={() => deleteConnection(conn.id)}
                          className="p-2 text-grey hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
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

          {/* Helpful tips */}
          <div className="grid md:grid-cols-3 gap-4">
            <InfoBanner type="tip" title="Security best practice" dismissible>
              Create a dedicated database user for CDC with minimal permissions.
              Only SELECT and REPLICATION privileges are needed - no write access required!
            </InfoBanner>
            <InfoBanner type="info" title="Connection testing" dismissible>
              Always test your connection after creating it. A green status means
              Savegress can successfully connect and read from your database.
            </InfoBanner>
            <InfoBanner
              type="tip"
              title="What comes next?"
              action={{ label: 'Open Optimizer', href: '/optimizer' }}
              dismissible
            >
              After connecting your database, use the Configuration Optimizer to find the best settings for your use case.
            </InfoBanner>
          </div>
        </>
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

function SummaryCard({
  label,
  value,
  icon: Icon,
  color = 'text-accent-cyan',
  help
}: {
  label: string;
  value: number;
  icon: React.ElementType;
  color?: string;
  help: string;
}) {
  return (
    <div className="card-dark p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm text-grey">{label}</span>
          <HelpIcon text={help} />
        </div>
        <Icon className={`w-4 h-4 ${color}`} />
      </div>
      <p className="text-2xl font-bold text-white mt-2">{value}</p>
    </div>
  );
}

function TestStatus({ status }: { status?: string }) {
  if (!status) {
    return (
      <div className="relative group">
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-grey/20 text-grey border border-grey/40 cursor-help">
          <Clock className="w-3 h-3" />
          Not tested
        </span>
        <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-[#0a1628] border border-cyan-40 rounded-lg text-sm text-grey whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity z-10">
          Click the refresh button to test this connection
        </div>
      </div>
    );
  }
  if (status === 'success') {
    return (
      <div className="relative group">
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400 border border-green-500/40 cursor-help">
          <CheckCircle className="w-3 h-3" />
          Connected
        </span>
        <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-[#0a1628] border border-cyan-40 rounded-lg text-sm text-grey whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity z-10">
          Connection successful! Ready to use in pipelines.
        </div>
      </div>
    );
  }
  return (
    <div className="relative group">
      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400 border border-red-500/40 cursor-help">
        <XCircle className="w-3 h-3" />
        Failed
      </span>
      <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-[#0a1628] border border-cyan-40 rounded-lg text-sm text-grey whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity z-10">
        Connection failed - check your credentials and try again
      </div>
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
  const [step, setStep] = useState(1);

  const selectedDbType = DATABASE_TYPES.find(t => t.value === type);

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

  const sslModeDescriptions: Record<string, string> = {
    disable: 'No encryption - only use for local development',
    prefer: 'Try SSL first, fall back to unencrypted if unavailable',
    require: 'Always use SSL, but don\'t verify the certificate',
    'verify-ca': 'Use SSL and verify the server certificate is signed by a trusted CA',
    'verify-full': 'Use SSL, verify certificate, and check hostname matches (most secure)',
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="card-dark w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Add Database Connection</h2>
          <p className="text-sm text-grey mt-1">
            Configure access to your source database for change data capture
          </p>
        </div>

        {/* Progress steps */}
        <div className="px-5 pt-4">
          <div className="flex items-center gap-2">
            <StepBadge number={1} active={step === 1} completed={step > 1} label="Database" />
            <div className="flex-1 h-px bg-cyan-40/30" />
            <StepBadge number={2} active={step === 2} completed={step > 2} label="Credentials" />
            <div className="flex-1 h-px bg-cyan-40/30" />
            <StepBadge number={3} active={step === 3} completed={false} label="Test" />
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-5">
          {error && (
            <InfoBanner type="warning" title="Error saving connection">
              {error}
            </InfoBanner>
          )}

          {step === 1 && (
            <div className="space-y-4">
              <FormFieldWithHelp
                label="Connection Name"
                help="A friendly name to identify this connection in the dashboard"
                tip="Use something descriptive like 'production-postgres' or 'staging-db'"
                required
              >
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  className="input-field"
                  placeholder="production-db"
                />
              </FormFieldWithHelp>

              <FormFieldWithHelp
                label="Database Type"
                help="Select the type of database you want to connect to"
                required
              >
                <div className="grid grid-cols-2 gap-3">
                  {DATABASE_TYPES.map((db) => (
                    <button
                      key={db.value}
                      type="button"
                      onClick={() => db.available && handleTypeChange(db.value)}
                      disabled={!db.available}
                      className={`p-3 rounded-lg border transition-colors text-left ${
                        type === db.value
                          ? 'bg-accent-cyan/10 border-accent-cyan'
                          : db.available
                          ? 'bg-primary-dark/50 border-cyan-40/30 hover:border-cyan-40'
                          : 'bg-primary-dark/30 border-cyan-40/20 opacity-50 cursor-not-allowed'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className={`font-medium ${type === db.value ? 'text-white' : 'text-grey'}`}>
                          {db.label}
                        </span>
                        {!db.available && (
                          <span className="text-xs px-2 py-0.5 bg-grey/20 text-grey rounded-full">
                            Coming soon
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-grey mt-1">{db.description}</p>
                    </button>
                  ))}
                </div>
              </FormFieldWithHelp>

              {selectedDbType && (
                <InfoBanner type="info" title="Requirements">
                  {selectedDbType.requirements}
                </InfoBanner>
              )}

              <div className="flex justify-end pt-4">
                <button
                  type="button"
                  onClick={() => setStep(2)}
                  disabled={!name || !type}
                  className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
                >
                  Next: Enter Credentials
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div className="col-span-2">
                  <FormFieldWithHelp
                    label="Host"
                    help="The hostname or IP address of your database server"
                    tip="Examples: db.example.com, 192.168.1.100, localhost"
                    required
                  >
                    <input
                      type="text"
                      value={host}
                      onChange={(e) => setHost(e.target.value)}
                      required
                      className="input-field"
                      placeholder="db.example.com"
                    />
                  </FormFieldWithHelp>
                </div>
                <div>
                  <FormFieldWithHelp
                    label="Port"
                    help={`Default port for ${selectedDbType?.label || 'database'}: ${selectedDbType?.port || port}`}
                    required
                  >
                    <input
                      type="number"
                      value={port}
                      onChange={(e) => setPort(parseInt(e.target.value))}
                      required
                      className="input-field"
                    />
                  </FormFieldWithHelp>
                </div>
              </div>

              <FormFieldWithHelp
                label="Database Name"
                help="The name of the specific database to connect to"
                tip="This is the database containing the tables you want to replicate"
                required
              >
                <input
                  type="text"
                  value={database}
                  onChange={(e) => setDatabase(e.target.value)}
                  required
                  className="input-field"
                  placeholder="myapp"
                />
              </FormFieldWithHelp>

              <FormFieldWithHelp
                label="Username"
                help="Database username with replication permissions"
                tip="We recommend creating a dedicated user with only SELECT and REPLICATION privileges"
                required
              >
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  className="input-field"
                  placeholder="replication_user"
                />
              </FormFieldWithHelp>

              <FormFieldWithHelp
                label="Password"
                help="Password for the database user"
                required
              >
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    className="input-field pr-10"
                    placeholder="Enter your password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-grey hover:text-white transition-colors"
                    title={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </FormFieldWithHelp>

              <ExpandableSection title="Advanced: SSL Settings" icon={Lock}>
                <div className="space-y-4">
                  <FormFieldWithHelp
                    label="SSL Mode"
                    help="How to handle SSL/TLS encryption for the connection"
                  >
                    <select
                      value={sslMode}
                      onChange={(e) => setSslMode(e.target.value)}
                      className="input-field"
                    >
                      <option value="disable">Disable - No encryption</option>
                      <option value="prefer">Prefer - Try SSL, fallback if unavailable</option>
                      <option value="require">Require - Always use SSL</option>
                      <option value="verify-ca">Verify CA - Validate certificate</option>
                      <option value="verify-full">Verify Full - Validate certificate + hostname</option>
                    </select>
                  </FormFieldWithHelp>
                  <p className="text-sm text-grey flex items-start gap-2">
                    <Info className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
                    {sslModeDescriptions[sslMode]}
                  </p>
                </div>
              </ExpandableSection>

              <InfoBanner type="tip" title="Security tip">
                <div className="flex items-start gap-2">
                  <Shield className="w-4 h-4 text-accent-cyan flex-shrink-0 mt-0.5" />
                  <span>
                    Your credentials are encrypted at rest and in transit. We recommend using
                    SSL mode &quot;require&quot; or higher for production databases.
                  </span>
                </div>
              </InfoBanner>

              <div className="flex justify-between pt-4">
                <button
                  type="button"
                  onClick={() => setStep(1)}
                  className="btn-secondary px-5 py-2.5 text-sm"
                >
                  Back
                </button>
                <button
                  type="button"
                  onClick={() => setStep(3)}
                  disabled={!host || !port || !database || !username || !password}
                  className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
                >
                  Next: Test Connection
                </button>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              {/* Connection summary */}
              <div className="p-4 bg-gradient-to-br from-accent-cyan/5 to-accent-blue/5 border border-cyan-40/30 rounded-xl">
                <h4 className="font-medium text-white mb-3 flex items-center gap-2">
                  <Database className="w-5 h-5 text-accent-cyan" />
                  Connection Summary
                </h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-grey">Name:</span>
                    <span className="text-white ml-2">{name}</span>
                  </div>
                  <div>
                    <span className="text-grey">Type:</span>
                    <span className="text-white ml-2">{selectedDbType?.label}</span>
                  </div>
                  <div>
                    <span className="text-grey">Host:</span>
                    <span className="text-white ml-2">{host}:{port}</span>
                  </div>
                  <div>
                    <span className="text-grey">Database:</span>
                    <span className="text-white ml-2">{database}</span>
                  </div>
                  <div>
                    <span className="text-grey">Username:</span>
                    <span className="text-white ml-2">{username}</span>
                  </div>
                  <div>
                    <span className="text-grey">SSL Mode:</span>
                    <span className="text-white ml-2">{sslMode}</span>
                  </div>
                </div>
              </div>

              {/* Test connection button */}
              <div className="p-4 bg-primary-dark/50 rounded-lg border border-cyan-40/30">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium text-white flex items-center gap-2">
                      <Zap className="w-4 h-4 text-accent-cyan" />
                      Test Your Connection
                    </h4>
                    <p className="text-sm text-grey mt-1">
                      Verify that Savegress can connect to your database before saving
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={testConnection}
                    disabled={isTesting}
                    className="btn-secondary px-4 py-2 text-sm"
                  >
                    {isTesting ? (
                      <>
                        <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                        Testing...
                      </>
                    ) : (
                      <>
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Test Now
                      </>
                    )}
                  </button>
                </div>

                {/* Test result */}
                {testResult && (
                  <div className={`mt-4 p-3 rounded-lg ${testResult.success ? 'bg-green-500/10 border border-green-500/30' : 'bg-red-500/10 border border-red-500/30'}`}>
                    <div className="flex items-start gap-2">
                      {testResult.success ? (
                        <CheckCircle className="w-5 h-5 text-green-400 flex-shrink-0" />
                      ) : (
                        <XCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
                      )}
                      <div>
                        <p className={`font-medium ${testResult.success ? 'text-green-400' : 'text-red-400'}`}>
                          {testResult.success ? 'Connection successful!' : 'Connection failed'}
                        </p>
                        <p className={`text-sm mt-1 ${testResult.success ? 'text-green-300' : 'text-red-300'}`}>
                          {testResult.message}
                        </p>
                        {!testResult.success && (
                          <p className="text-xs text-grey mt-2">
                            Check your credentials and make sure the database is accessible from the internet.
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {!testResult && (
                <InfoBanner type="warning" title="Don't skip the test!">
                  We strongly recommend testing your connection before saving.
                  This ensures everything is configured correctly.
                </InfoBanner>
              )}

              <div className="flex justify-between pt-4">
                <button
                  type="button"
                  onClick={() => setStep(2)}
                  className="btn-secondary px-5 py-2.5 text-sm"
                >
                  Back
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
            </div>
          )}
        </form>
      </div>
    </div>
  );
}

function StepBadge({
  number,
  active,
  completed,
  label
}: {
  number: number;
  active: boolean;
  completed: boolean;
  label: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <div
        className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium transition-colors ${
          completed
            ? 'bg-accent-cyan text-white'
            : active
            ? 'bg-gradient-btn-primary text-white'
            : 'bg-primary-dark text-grey border border-cyan-40/30'
        }`}
      >
        {completed ? '✓' : number}
      </div>
      <span className={`text-sm ${active ? 'text-white' : 'text-grey'}`}>{label}</span>
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
