'use client';

import { useState } from 'react';
import {
  Book, Database, Settings, Terminal, Server,
  FileCode, Zap, Shield, Box, ChevronRight,
  Copy, Check, Play, Pause, RefreshCw, Layers
} from 'lucide-react';

// Code block component with copy button
function CodeBlock({ code, language = 'bash' }: { code: string; language?: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group">
      <pre className="bg-[#0a1628] border border-cyan-40 rounded-lg p-4 overflow-x-auto text-sm">
        <code className="text-grey-light">{code}</code>
      </pre>
      <button
        onClick={handleCopy}
        className="absolute top-3 right-3 p-2 rounded-md bg-cyan-40/50 hover:bg-cyan-40 transition-colors opacity-0 group-hover:opacity-100"
      >
        {copied ? <Check className="w-4 h-4 text-accent-cyan" /> : <Copy className="w-4 h-4 text-grey" />}
      </button>
    </div>
  );
}

// Section component
function Section({ id, title, icon: Icon, children }: {
  id: string;
  title: string;
  icon: React.ElementType;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-24">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-2 rounded-lg bg-accent-cyan/10 border border-accent-cyan/30">
          <Icon className="w-5 h-5 text-accent-cyan" />
        </div>
        <h2 className="text-h4 text-white">{title}</h2>
      </div>
      <div className="space-y-6">
        {children}
      </div>
    </section>
  );
}

// Subsection component
function Subsection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-4">
      <h3 className="text-h5 text-white">{title}</h3>
      {children}
    </div>
  );
}

// Database card component
function DatabaseCard({
  name,
  type,
  port,
  description,
  options,
  tier
}: {
  name: string;
  type: string;
  port: string;
  description: string;
  options: { name: string; default?: string; description: string }[];
  tier: 'community' | 'pro' | 'enterprise';
}) {
  const tierColors = {
    community: 'bg-green-500/10 text-green-400 border-green-500/30',
    pro: 'bg-accent-cyan/10 text-accent-cyan border-accent-cyan/30',
    enterprise: 'bg-purple-500/10 text-purple-400 border-purple-500/30'
  };

  return (
    <div className="card-dark p-6 space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h4 className="text-h5 text-white">{name}</h4>
          <p className="text-content-2 text-grey mt-1">Type: <code className="text-accent-cyan">{type}</code> | Port: <code className="text-accent-cyan">{port}</code></p>
        </div>
        <span className={`px-3 py-1 rounded-full text-xs font-medium border ${tierColors[tier]}`}>
          {tier}
        </span>
      </div>
      <p className="text-content-1 text-grey">{description}</p>
      {options.length > 0 && (
        <div className="space-y-2">
          <p className="text-content-2 text-grey-light font-medium">Options:</p>
          <div className="space-y-1">
            {options.map((opt) => (
              <div key={opt.name} className="flex items-start gap-2 text-sm">
                <code className="text-accent-cyan bg-cyan-40/30 px-2 py-0.5 rounded">{opt.name}</code>
                {opt.default && <span className="text-grey">(default: {opt.default})</span>}
                <span className="text-grey">— {opt.description}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Navigation items
const navItems = [
  { id: 'quick-start', label: 'Quick Start', icon: Zap },
  { id: 'modes', label: 'Operation Modes', icon: Layers },
  { id: 'installation', label: 'Installation', icon: Box },
  { id: 'configuration', label: 'Configuration', icon: Settings },
  { id: 'databases', label: 'Databases', icon: Database },
  { id: 'outputs', label: 'Outputs', icon: Server },
  { id: 'snapshots', label: 'Snapshots', icon: RefreshCw },
  { id: 'checkpoints', label: 'Checkpoints', icon: Play },
  { id: 'cli', label: 'CLI Reference', icon: Terminal },
  { id: 'licensing', label: 'Licensing', icon: Shield },
  { id: 'api', label: 'Event Format', icon: FileCode },
];

export default function DocsPage() {
  const [activeSection, setActiveSection] = useState('quick-start');

  return (
    <div className="flex gap-8">
      {/* Sidebar Navigation */}
      <aside className="hidden lg:block w-64 flex-shrink-0">
        <div className="sticky top-24 space-y-1">
          <div className="flex items-center gap-2 mb-6">
            <Book className="w-5 h-5 text-accent-cyan" />
            <span className="text-h5 text-white">Documentation</span>
          </div>
          <nav className="space-y-1">
            {navItems.map((item) => (
              <a
                key={item.id}
                href={`#${item.id}`}
                onClick={() => setActiveSection(item.id)}
                className={`flex items-center gap-3 px-4 py-2.5 rounded-lg transition-colors ${
                  activeSection === item.id
                    ? 'bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/30'
                    : 'text-grey hover:text-white hover:bg-cyan-40/30'
                }`}
              >
                <item.icon className="w-4 h-4" />
                <span className="text-sm">{item.label}</span>
              </a>
            ))}
          </nav>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 max-w-4xl space-y-16 pb-24">
        {/* Header */}
        <div>
          <h1 className="text-h3 text-white mb-4">Savegress CDC Engine</h1>
          <p className="text-content-1 text-grey">
            Complete documentation for setting up and using the CDC engine for data migration and real-time replication.
          </p>
        </div>

        {/* Quick Start */}
        <Section id="quick-start" title="Quick Start" icon={Zap}>
          <p className="text-content-1 text-grey">
            Get started with Savegress in 5 minutes. This example shows basic setup for PostgreSQL.
          </p>

          <Subsection title="1. Download the Engine">
            <p className="text-content-1 text-grey mb-3">
              Download the latest version for your platform from the Downloads page in your dashboard.
            </p>
            <CodeBlock code={`# Linux/macOS
chmod +x cdc-engine
./cdc-engine --help`} />
          </Subsection>

          <Subsection title="2. Configure PostgreSQL">
            <p className="text-content-1 text-grey mb-3">
              Enable logical replication in PostgreSQL:
            </p>
            <CodeBlock language="sql" code={`-- postgresql.conf
wal_level = logical
max_replication_slots = 4
max_wal_senders = 4

-- Create a user for CDC
CREATE USER cdc_user WITH REPLICATION PASSWORD 'secure_password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;

-- Create publication
CREATE PUBLICATION cdc_pub FOR ALL TABLES;`} />
          </Subsection>

          <Subsection title="3. Run the Engine">
            <CodeBlock code={`# Using environment variables
export CDC_SOURCE_TYPE=postgres
export CDC_SOURCE_HOST=localhost
export CDC_SOURCE_DATABASE=mydb
export CDC_SOURCE_USER=cdc_user
export CDC_SOURCE_PASSWORD=secure_password
export CDC_OUTPUT_TYPE=stdout

./cdc-engine`} />
            <p className="text-content-1 text-grey mt-3">
              Or use a configuration file:
            </p>
            <CodeBlock language="yaml" code={`# config.yaml
source:
  type: postgres
  host: localhost
  port: 5432
  database: mydb
  user: cdc_user
  password: secure_password
  tables: [users, orders]
  options:
    slot_name: savegress_slot
    publication: cdc_pub

output:
  type: stdout

checkpoint:
  type: file
  path: ./checkpoints`} />
            <CodeBlock code={`./cdc-engine --config config.yaml`} />
          </Subsection>
        </Section>

        {/* Operation Modes */}
        <Section id="modes" title="Operation Modes" icon={Layers}>
          <p className="text-content-1 text-grey">
            Savegress supports three operation modes depending on your use case: migration, replication, or both.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
            <div className="card-dark p-6 border-purple-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 rounded-lg bg-purple-500/10">
                  <RefreshCw className="w-5 h-5 text-purple-400" />
                </div>
                <h4 className="text-h5 text-white">Migration Only</h4>
              </div>
              <p className="text-content-2 text-grey mb-4">
                One-time data transfer from source to target. Engine stops after snapshot completion.
              </p>
              <ul className="text-sm text-grey space-y-1">
                <li>• Initial data load</li>
                <li>• Database migrations</li>
                <li>• Data warehouse seeding</li>
              </ul>
            </div>

            <div className="card-dark p-6 border-accent-cyan/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 rounded-lg bg-accent-cyan/10">
                  <Play className="w-5 h-5 text-accent-cyan" />
                </div>
                <h4 className="text-h5 text-white">Replication Only</h4>
              </div>
              <p className="text-content-2 text-grey mb-4">
                Continuous streaming of changes without initial snapshot. Start from specific position.
              </p>
              <ul className="text-sm text-grey space-y-1">
                <li>• Real-time sync</li>
                <li>• Event streaming</li>
                <li>• Resume from checkpoint</li>
              </ul>
            </div>

            <div className="card-dark p-6 border-green-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 rounded-lg bg-green-500/10">
                  <Layers className="w-5 h-5 text-green-400" />
                </div>
                <h4 className="text-h5 text-white">Migration + Replication</h4>
              </div>
              <p className="text-content-2 text-grey mb-4">
                Full snapshot followed by continuous replication. Best for complete data sync.
              </p>
              <ul className="text-sm text-grey space-y-1">
                <li>• Zero-downtime migration</li>
                <li>• Full data consistency</li>
                <li>• Recommended approach</li>
              </ul>
            </div>
          </div>

          <Subsection title="Configuration Examples">
            <p className="text-content-1 text-grey mb-3">
              <strong>Migration Only</strong> — snapshot then exit:
            </p>
            <CodeBlock language="yaml" code={`snapshot:
  enabled: true
  exit_after: true    # Stop after snapshot completes

# No streaming after snapshot`} />

            <p className="text-content-1 text-grey mb-3 mt-6">
              <strong>Replication Only</strong> — skip snapshot, stream changes:
            </p>
            <CodeBlock language="yaml" code={`snapshot:
  enabled: false      # Skip initial snapshot

source:
  options:
    start_lsn: "0/16B3748"  # Start from specific position`} />

            <p className="text-content-1 text-grey mb-3 mt-6">
              <strong>Migration + Replication</strong> — full sync (default):
            </p>
            <CodeBlock language="yaml" code={`snapshot:
  enabled: true       # Perform initial snapshot
  exit_after: false   # Continue with streaming (default)
  parallel_workers: 4
  batch_size: 10000

# After snapshot, automatically switches to streaming mode`} />
          </Subsection>

          <Subsection title="How It Works">
            <div className="bg-primary-dark/50 rounded-lg p-6 border border-cyan-40/30">
              <div className="flex flex-col md:flex-row items-start md:items-center gap-4">
                <div className="flex-1 text-center p-4">
                  <div className="w-12 h-12 rounded-full bg-purple-500/20 flex items-center justify-center mx-auto mb-2">
                    <span className="text-purple-400 font-bold">1</span>
                  </div>
                  <p className="text-sm text-white font-medium">Snapshot</p>
                  <p className="text-xs text-grey">Read existing data</p>
                </div>
                <div className="hidden md:block text-grey">→</div>
                <div className="flex-1 text-center p-4">
                  <div className="w-12 h-12 rounded-full bg-accent-cyan/20 flex items-center justify-center mx-auto mb-2">
                    <span className="text-accent-cyan font-bold">2</span>
                  </div>
                  <p className="text-sm text-white font-medium">Checkpoint</p>
                  <p className="text-xs text-grey">Save position</p>
                </div>
                <div className="hidden md:block text-grey">→</div>
                <div className="flex-1 text-center p-4">
                  <div className="w-12 h-12 rounded-full bg-green-500/20 flex items-center justify-center mx-auto mb-2">
                    <span className="text-green-400 font-bold">3</span>
                  </div>
                  <p className="text-sm text-white font-medium">Streaming</p>
                  <p className="text-xs text-grey">Real-time changes</p>
                </div>
              </div>
            </div>
          </Subsection>
        </Section>

        {/* Installation */}
        <Section id="installation" title="Installation" icon={Box}>
          <p className="text-content-1 text-grey">
            Savegress CDC Engine is available for Linux, macOS, and Windows. Choose your preferred installation method.
          </p>

          <Subsection title="Binary Downloads">
            <p className="text-content-1 text-grey mb-3">
              Download pre-built binaries from the Downloads page:
            </p>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {['Linux (amd64)', 'macOS (arm64)', 'Windows (amd64)'].map((platform) => (
                <div key={platform} className="card-dark p-4 text-center">
                  <p className="text-content-1 text-white">{platform}</p>
                </div>
              ))}
            </div>
          </Subsection>

          <Subsection title="Docker">
            <CodeBlock code={`docker pull savegress/cdc-engine:latest

docker run -d \\
  -e CDC_SOURCE_TYPE=postgres \\
  -e CDC_SOURCE_HOST=host.docker.internal \\
  -e CDC_SOURCE_DATABASE=mydb \\
  -e CDC_SOURCE_USER=cdc_user \\
  -e CDC_SOURCE_PASSWORD=secret \\
  -e CDC_OUTPUT_TYPE=http \\
  -e CDC_OUTPUT_URL=http://your-server:8080/events \\
  -v ./checkpoints:/app/checkpoints \\
  savegress/cdc-engine:latest`} />
          </Subsection>

          <Subsection title="Kubernetes">
            <CodeBlock language="yaml" code={`apiVersion: apps/v1
kind: Deployment
metadata:
  name: savegress-cdc
spec:
  replicas: 1
  selector:
    matchLabels:
      app: savegress-cdc
  template:
    metadata:
      labels:
        app: savegress-cdc
    spec:
      containers:
      - name: cdc-engine
        image: savegress/cdc-engine:latest
        env:
        - name: CDC_SOURCE_TYPE
          value: postgres
        - name: CDC_SOURCE_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        # ... other env vars
        volumeMounts:
        - name: checkpoints
          mountPath: /app/checkpoints
      volumes:
      - name: checkpoints
        persistentVolumeClaim:
          claimName: cdc-checkpoints`} />
          </Subsection>
        </Section>

        {/* Configuration */}
        <Section id="configuration" title="Configuration" icon={Settings}>
          <p className="text-content-1 text-grey">
            The engine can be configured via YAML file or environment variables. Environment variables take precedence.
          </p>

          <Subsection title="Full Configuration Example">
            <CodeBlock language="yaml" code={`# config.yaml - Complete configuration
source:
  type: postgres              # postgres, mysql, mariadb, mongodb, sqlserver, oracle, cassandra, dynamodb
  host: localhost
  port: 5432
  database: production
  user: cdc_user
  password: secure_password
  tables:                     # Empty = all tables
    - users
    - orders
    - products
  options:                    # Database-specific options
    slot_name: savegress_slot
    publication: cdc_pub
  heartbeat_enabled: true     # Enable heartbeat events
  heartbeat_interval_seconds: 10

snapshot:
  enabled: true               # Perform initial snapshot
  parallel_workers: 4         # Parallel workers for speed
  batch_size: 10000           # Rows per batch

output:
  type: http                  # stdout, file, http
  options:
    url: https://your-api.com/events
    compression: true         # Enable compression
  headers:
    Authorization: "Bearer \${AUTH_TOKEN}"  # Env var substitution
    X-Source: savegress

checkpoint:
  type: file                  # memory, file
  path: ./checkpoints

monitoring:
  prometheus:
    enabled: true
    port: 9090
  logging:
    level: info               # debug, info, warn, error

license:
  offline: false              # Offline mode (no telemetry)
  grace_period_days: 7`} />
          </Subsection>

          <Subsection title="Environment Variables">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-cyan-40">
                    <th className="text-left py-3 px-4 text-grey-light">Variable</th>
                    <th className="text-left py-3 px-4 text-grey-light">Description</th>
                    <th className="text-left py-3 px-4 text-grey-light">Example</th>
                  </tr>
                </thead>
                <tbody className="text-grey">
                  {[
                    ['CDC_SOURCE_TYPE', 'Database type', 'postgres'],
                    ['CDC_SOURCE_HOST', 'Database host', 'localhost'],
                    ['CDC_SOURCE_PORT', 'Database port', '5432'],
                    ['CDC_SOURCE_DATABASE', 'Database name', 'production'],
                    ['CDC_SOURCE_USER', 'Username', 'cdc_user'],
                    ['CDC_SOURCE_PASSWORD', 'Password', 'secret'],
                    ['CDC_SOURCE_TABLES', 'Tables (comma-separated)', 'users,orders'],
                    ['CDC_OUTPUT_TYPE', 'Output type', 'http'],
                    ['CDC_OUTPUT_URL', 'URL for HTTP output', 'https://api.com/events'],
                    ['CDC_SNAPSHOT_ENABLED', 'Enable snapshot', 'true'],
                    ['CDC_CHECKPOINT_TYPE', 'Checkpoint type', 'file'],
                    ['CDC_CHECKPOINT_PATH', 'Checkpoint path', './checkpoints'],
                    ['CDC_LOG_LEVEL', 'Log level', 'info'],
                    ['SAVEGRESS_LICENSE_KEY', 'License key', 'eyJ...'],
                    ['SAVEGRESS_OFFLINE', 'Offline mode', 'true'],
                  ].map(([name, desc, example]) => (
                    <tr key={name} className="border-b border-cyan-40/30">
                      <td className="py-3 px-4"><code className="text-accent-cyan">{name}</code></td>
                      <td className="py-3 px-4">{desc}</td>
                      <td className="py-3 px-4"><code>{example}</code></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Subsection>
        </Section>

        {/* Databases */}
        <Section id="databases" title="Supported Databases" icon={Database}>
          <p className="text-content-1 text-grey mb-6">
            Savegress supports 8 databases. Availability depends on your license tier.
          </p>

          <div className="space-y-6">
            <DatabaseCard
              name="PostgreSQL"
              type="postgres"
              port="5432"
              tier="community"
              description="Uses logical replication via WAL (Write-Ahead Log). Requires wal_level = logical."
              options={[
                { name: 'slot_name', default: 'cdc_slot', description: 'Replication slot name' },
                { name: 'publication', default: 'cdc_pub', description: 'Publication name' },
                { name: 'start_lsn', description: 'Starting LSN for resume' },
              ]}
            />

            <DatabaseCard
              name="MySQL"
              type="mysql"
              port="3306"
              tier="community"
              description="Reads changes from binlog. Requires binlog enabled with ROW format."
              options={[
                { name: 'server_id', default: '100', description: 'Server ID for replication' },
                { name: 'use_gtid', default: 'false', description: 'Use GTID instead of position' },
                { name: 'start_gtid', description: 'Starting GTID' },
                { name: 'start_pos_name', description: 'Binlog file name' },
                { name: 'start_pos', description: 'Binlog position' },
              ]}
            />

            <DatabaseCard
              name="MariaDB"
              type="mariadb"
              port="3306"
              tier="community"
              description="Similar to MySQL, uses binlog replication with GTID support."
              options={[
                { name: 'server_id', default: '100', description: 'Server ID for replication' },
                { name: 'use_gtid', default: 'false', description: 'Use GTID' },
              ]}
            />

            <DatabaseCard
              name="MongoDB"
              type="mongodb"
              port="27017"
              tier="pro"
              description="Uses Change Streams API. Requires replica set or sharded cluster."
              options={[
                { name: 'uri', description: 'Connection URI (alternative to host/port)' },
                { name: 'replica_set', description: 'Replica set name' },
                { name: 'auth_source', default: 'admin', description: 'Authentication database' },
                { name: 'full_document', default: 'updateLookup', description: 'Full document retrieval strategy' },
                { name: 'batch_size', default: '1000', description: 'Batch size' },
                { name: 'resume_token', description: 'Token for resume' },
              ]}
            />

            <DatabaseCard
              name="SQL Server"
              type="sqlserver"
              port="1433"
              tier="pro"
              description="Uses built-in SQL Server CDC. Requires CDC enabled on database and tables."
              options={[
                { name: 'instance', description: 'SQL Server instance name' },
                { name: 'start_lsn', description: 'Starting LSN' },
                { name: 'poll_interval', description: 'Poll interval' },
              ]}
            />

            <DatabaseCard
              name="Oracle"
              type="oracle"
              port="1521"
              tier="enterprise"
              description="Uses LogMiner to read redo logs. Requires ARCHIVELOG mode."
              options={[
                { name: 'start_scn', description: 'Starting SCN (System Change Number)' },
                { name: 'poll_interval', description: 'Poll interval' },
                { name: 'use_sid', default: 'false', description: 'Use SID instead of service name' },
              ]}
            />

            <DatabaseCard
              name="Cassandra"
              type="cassandra"
              port="9042"
              tier="pro"
              description="Reads Cassandra CDC logs. Requires CDC enabled on tables."
              options={[
                { name: 'hosts', description: 'Comma-separated list of hosts' },
                { name: 'cdc_log_path', description: 'Path to CDC logs' },
                { name: 'poll_interval', description: 'Poll interval' },
                { name: 'consistency', default: 'LOCAL_QUORUM', description: 'Consistency level' },
                { name: 'local_dc', description: 'Local datacenter' },
              ]}
            />

            <DatabaseCard
              name="DynamoDB"
              type="dynamodb"
              port="N/A"
              tier="pro"
              description="Uses DynamoDB Streams. Requires Streams enabled on tables."
              options={[
                { name: 'region', default: 'us-east-1', description: 'AWS region' },
                { name: 'endpoint', description: 'Custom endpoint (for LocalStack)' },
              ]}
            />
          </div>
        </Section>

        {/* Outputs */}
        <Section id="outputs" title="Output Types" icon={Server}>
          <p className="text-content-1 text-grey mb-6">
            Savegress supports multiple output types for sending captured events.
          </p>

          <Subsection title="stdout">
            <p className="text-content-1 text-grey mb-3">
              Output to console. Ideal for development and debugging.
            </p>
            <CodeBlock language="yaml" code={`output:
  type: stdout`} />
          </Subsection>

          <Subsection title="file">
            <p className="text-content-1 text-grey mb-3">
              Write to file in JSONL format (one JSON event per line). Supports compression.
            </p>
            <CodeBlock language="yaml" code={`output:
  type: file
  options:
    path: /var/log/cdc/events.jsonl
    compression: true  # Enable Rust compression (5x-150x)`} />
          </Subsection>

          <Subsection title="http">
            <p className="text-content-1 text-grey mb-3">
              Send events to HTTP endpoint. Supports batching, compression, and custom headers.
            </p>
            <CodeBlock language="yaml" code={`output:
  type: http
  options:
    url: https://your-api.com/cdc/events
    compression: true
  headers:
    Authorization: "Bearer \${AUTH_TOKEN}"
    Content-Type: application/json
    X-CDC-Source: savegress`} />
            <div className="mt-4 p-4 bg-accent-cyan/10 border border-accent-cyan/30 rounded-lg">
              <p className="text-content-2 text-accent-cyan">
                <strong>Tip:</strong> Use <code>$&#123;ENV_VAR&#125;</code> for environment variable substitution in headers.
              </p>
            </div>
          </Subsection>
        </Section>

        {/* Snapshots */}
        <Section id="snapshots" title="Snapshots" icon={RefreshCw}>
          <p className="text-content-1 text-grey">
            Snapshots capture the initial state of data before starting CDC streaming.
          </p>

          <Subsection title="Initial Snapshot">
            <p className="text-content-1 text-grey mb-3">
              Executed once on first run. Blocking — streaming starts after completion.
            </p>
            <CodeBlock language="yaml" code={`snapshot:
  enabled: true
  parallel_workers: 4    # Parallel workers for speed
  batch_size: 10000      # Rows per batch`} />
          </Subsection>

          <Subsection title="Snapshot Process">
            <div className="space-y-3 text-content-1 text-grey">
              <div className="flex items-start gap-3">
                <div className="w-6 h-6 rounded-full bg-accent-cyan/20 flex items-center justify-center text-accent-cyan text-sm">1</div>
                <p>Engine starts parallel workers to read tables</p>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-6 h-6 rounded-full bg-accent-cyan/20 flex items-center justify-center text-accent-cyan text-sm">2</div>
                <p>Data is read in batches of 10,000 rows (configurable)</p>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-6 h-6 rounded-full bg-accent-cyan/20 flex items-center justify-center text-accent-cyan text-sm">3</div>
                <p>Events are sent to output with <code className="text-accent-cyan">snapshot</code> operation</p>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-6 h-6 rounded-full bg-accent-cyan/20 flex items-center justify-center text-accent-cyan text-sm">4</div>
                <p>After completion, checkpoint is saved and streaming begins</p>
              </div>
            </div>
          </Subsection>
        </Section>

        {/* Checkpoints */}
        <Section id="checkpoints" title="Checkpoints" icon={Play}>
          <p className="text-content-1 text-grey">
            Checkpoints save the replication position for resuming after restarts.
          </p>

          <Subsection title="Checkpoint Types">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="card-dark p-4">
                <h4 className="text-h5 text-white mb-2">memory</h4>
                <p className="text-content-2 text-grey">
                  Stores position in memory only. Lost on restart. For testing.
                </p>
              </div>
              <div className="card-dark p-4">
                <h4 className="text-h5 text-white mb-2">file</h4>
                <p className="text-content-2 text-grey">
                  Persists position to disk. Recommended for production.
                </p>
              </div>
            </div>
          </Subsection>

          <Subsection title="Configuration">
            <CodeBlock language="yaml" code={`checkpoint:
  type: file
  path: ./checkpoints  # Storage directory`} />
          </Subsection>

          <Subsection title="How It Works">
            <ul className="list-disc list-inside space-y-2 text-content-1 text-grey">
              <li>Position is saved every <strong>10 seconds</strong></li>
              <li>Also saved on each heartbeat event</li>
              <li>Final position saved on graceful shutdown</li>
              <li>On startup, last position is automatically loaded</li>
            </ul>
          </Subsection>
        </Section>

        {/* CLI Reference */}
        <Section id="cli" title="CLI Reference" icon={Terminal}>
          <CodeBlock code={`./cdc-engine [flags]

Flags:
  -config string
        Path to YAML configuration file
        (optional, uses env vars if not specified)

  -offline
        Run in offline mode - no license server or telemetry calls
        (can also set SAVEGRESS_OFFLINE=true)

  -pprof string
        Enable pprof profiling server on specified address
        Example: -pprof localhost:6060`} />

          <Subsection title="Usage Examples">
            <CodeBlock code={`# With configuration file
./cdc-engine --config /etc/savegress/config.yaml

# With environment variables
CDC_SOURCE_TYPE=postgres CDC_OUTPUT_TYPE=stdout ./cdc-engine

# In offline mode (no telemetry)
./cdc-engine --offline --config config.yaml

# With profiling
./cdc-engine --config config.yaml --pprof localhost:6060`} />
          </Subsection>
        </Section>

        {/* Licensing */}
        <Section id="licensing" title="Licensing" icon={Shield}>
          <p className="text-content-1 text-grey mb-6">
            Savegress uses licensing to unlock features. Licenses are validated locally via Ed25519 signature.
          </p>

          <Subsection title="License Tiers">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="card-dark p-6 border-green-500/30">
                <h4 className="text-h5 text-green-400 mb-3">Community</h4>
                <ul className="space-y-2 text-content-2 text-grey">
                  <li>✓ PostgreSQL</li>
                  <li>✓ MySQL</li>
                  <li>✓ MariaDB</li>
                  <li>✓ 1 source</li>
                  <li>✓ 10 tables</li>
                  <li>✓ 1,000 events/sec</li>
                </ul>
              </div>
              <div className="card-dark p-6 border-accent-cyan/30">
                <h4 className="text-h5 text-accent-cyan mb-3">Pro</h4>
                <ul className="space-y-2 text-content-2 text-grey">
                  <li>✓ Everything in Community</li>
                  <li>✓ MongoDB</li>
                  <li>✓ SQL Server</li>
                  <li>✓ Cassandra</li>
                  <li>✓ DynamoDB</li>
                  <li>✓ 10 sources</li>
                  <li>✓ 100 tables</li>
                  <li>✓ 50,000 events/sec</li>
                  <li>✓ Kafka/gRPC output</li>
                </ul>
              </div>
              <div className="card-dark p-6 border-purple-500/30">
                <h4 className="text-h5 text-purple-400 mb-3">Enterprise</h4>
                <ul className="space-y-2 text-content-2 text-grey">
                  <li>✓ Everything in Pro</li>
                  <li>✓ Oracle</li>
                  <li>✓ Unlimited</li>
                  <li>✓ HA / Raft cluster</li>
                  <li>✓ SSO / LDAP</li>
                  <li>✓ Audit log</li>
                  <li>✓ Multi-tenant</li>
                </ul>
              </div>
            </div>
          </Subsection>

          <Subsection title="Setting Up License">
            <CodeBlock code={`# Via environment variable
export SAVEGRESS_LICENSE_KEY="eyJ0eXAi..."

# Via file
export SAVEGRESS_LICENSE_FILE="/etc/savegress/license.key"

# Offline mode (no phone-home)
export SAVEGRESS_OFFLINE=true`} />
          </Subsection>
        </Section>

        {/* Event Format */}
        <Section id="api" title="Event Format" icon={FileCode}>
          <p className="text-content-1 text-grey mb-4">
            All events are transmitted in a unified JSON format.
          </p>

          <Subsection title="Event Structure">
            <CodeBlock language="json" code={`{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "operation": "insert",
  "timestamp": "2024-01-15T10:30:00.123456Z",
  "source": {
    "type": "postgres",
    "database": "production",
    "schema": "public",
    "table": "users",
    "position": "0/16B3748"
  },
  "before": null,
  "after": {
    "id": 123,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "metadata": {
    "tx_id": "12345",
    "commit_ts": "2024-01-15T10:30:00.123456Z"
  }
}`} />
          </Subsection>

          <Subsection title="Operation Types">
            <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
              {['insert', 'update', 'delete', 'snapshot', 'heartbeat'].map((op) => (
                <div key={op} className="card-dark p-3 text-center">
                  <code className="text-accent-cyan">{op}</code>
                </div>
              ))}
            </div>
          </Subsection>

          <Subsection title="Event Fields">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-cyan-40">
                    <th className="text-left py-3 px-4 text-grey-light">Field</th>
                    <th className="text-left py-3 px-4 text-grey-light">Type</th>
                    <th className="text-left py-3 px-4 text-grey-light">Description</th>
                  </tr>
                </thead>
                <tbody className="text-grey">
                  {[
                    ['id', 'string', 'Unique event ID (UUID)'],
                    ['operation', 'string', 'Operation type: insert, update, delete, snapshot, heartbeat'],
                    ['timestamp', 'string', 'ISO 8601 timestamp'],
                    ['source', 'object', 'Source information'],
                    ['source.type', 'string', 'Database type'],
                    ['source.database', 'string', 'Database name'],
                    ['source.schema', 'string', 'Schema (if applicable)'],
                    ['source.table', 'string', 'Table name'],
                    ['source.position', 'string', 'Replication position'],
                    ['before', 'object|null', 'State BEFORE change (for update/delete)'],
                    ['after', 'object|null', 'State AFTER change (for insert/update)'],
                    ['metadata', 'object', 'Additional metadata'],
                  ].map(([field, type, desc]) => (
                    <tr key={field} className="border-b border-cyan-40/30">
                      <td className="py-3 px-4"><code className="text-accent-cyan">{field}</code></td>
                      <td className="py-3 px-4"><code>{type}</code></td>
                      <td className="py-3 px-4">{desc}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Subsection>
        </Section>

      </main>
    </div>
  );
}
