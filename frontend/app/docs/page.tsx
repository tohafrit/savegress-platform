import type { Metadata } from "next"
import Link from "next/link"

export const metadata: Metadata = {
  title: "Documentation - Savegress",
  description: "Learn how to use Savegress for cross-cloud database replication with CDC.",
}

export default function Documentation() {
  return (
    <main className="min-h-screen bg-white">
      <div className="container-custom py-16">
        <div className="max-w-4xl mx-auto">
          <Link
            href="/"
            className="text-accent-blue hover:underline mb-8 inline-block"
          >
            &larr; Back to Home
          </Link>

          <h1 className="text-4xl font-bold text-primary mb-4">Documentation</h1>
          <p className="text-xl text-gray-600 mb-12">
            Everything you need to get started with Savegress and optimize your cross-cloud data replication.
          </p>

          <div className="grid md:grid-cols-2 gap-8 mb-16">
            <div className="border border-gray-200 rounded-lg p-6 hover:border-accent-blue transition-colors">
              <h2 className="text-xl font-semibold text-primary mb-3">Quick Start</h2>
              <p className="text-gray-600 mb-4">
                Get up and running with Savegress in minutes. Learn how to set up your first replication pipeline.
              </p>
              <a href="#quick-start" className="text-accent-blue hover:underline">Get started &rarr;</a>
            </div>

            <div className="border border-gray-200 rounded-lg p-6 hover:border-accent-blue transition-colors">
              <h2 className="text-xl font-semibold text-primary mb-3">Concepts</h2>
              <p className="text-gray-600 mb-4">
                Understand how CDC works and how Savegress achieves up to 200x compression.
              </p>
              <a href="#concepts" className="text-accent-blue hover:underline">Learn more &rarr;</a>
            </div>

            <div className="border border-gray-200 rounded-lg p-6 hover:border-accent-blue transition-colors">
              <h2 className="text-xl font-semibold text-primary mb-3">Supported Databases</h2>
              <p className="text-gray-600 mb-4">
                See which databases are supported and cloud-specific configuration details.
              </p>
              <a href="#databases" className="text-accent-blue hover:underline">View databases &rarr;</a>
            </div>

            <div className="border border-gray-200 rounded-lg p-6 hover:border-accent-blue transition-colors">
              <h2 className="text-xl font-semibold text-primary mb-3">API Reference</h2>
              <p className="text-gray-600 mb-4">
                Complete API documentation for programmatic access to Savegress.
              </p>
              <a href="#api" className="text-accent-blue hover:underline">View API docs &rarr;</a>
            </div>
          </div>

          {/* Quick Start Section */}
          <section id="quick-start" className="mb-16 scroll-mt-8">
            <h2 className="text-3xl font-bold text-primary mb-6">Quick Start</h2>

            <div className="space-y-8">
              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">1. Sign Up</h3>
                <p className="text-gray-700 mb-4">
                  Create your Savegress account to get access to the dashboard and API credentials.
                </p>
              </div>

              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">2. Connect Your Source Database</h3>
                <p className="text-gray-700 mb-4">
                  Configure your source database (PostgreSQL or MySQL) to enable logical replication. Savegress will guide you through the necessary permissions and settings.
                </p>
                <div className="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-sm overflow-x-auto">
                  <pre>{`-- PostgreSQL: Enable logical replication
ALTER SYSTEM SET wal_level = logical;
ALTER SYSTEM SET max_replication_slots = 4;

-- Create a replication user
CREATE USER savegress_replication REPLICATION LOGIN;`}</pre>
                </div>
              </div>

              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">3. Set Up Your Target</h3>
                <p className="text-gray-700 mb-4">
                  Connect your target database in any supported cloud. Savegress handles the schema migration and initial data sync automatically.
                </p>
              </div>

              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">4. Start Replicating</h3>
                <p className="text-gray-700 mb-4">
                  Once configured, Savegress continuously captures changes and replicates them to your target with minimal latency and maximum compression.
                </p>
              </div>
            </div>
          </section>

          {/* Concepts Section */}
          <section id="concepts" className="mb-16 scroll-mt-8">
            <h2 className="text-3xl font-bold text-primary mb-6">Core Concepts</h2>

            <div className="space-y-8">
              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">Change Data Capture (CDC)</h3>
                <p className="text-gray-700 mb-4">
                  CDC is a technique that identifies and captures changes made to data in a database. Instead of querying the entire dataset, Savegress reads the database&apos;s transaction log to capture only the changes (inserts, updates, deletes).
                </p>
              </div>

              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">Delta Compression</h3>
                <p className="text-gray-700 mb-4">
                  Savegress uses advanced delta compression algorithms to minimize the amount of data transferred. By sending only the differences between states, we achieve compression ratios up to 200x compared to full data transfers.
                </p>
              </div>

              <div>
                <h3 className="text-xl font-semibold text-primary mb-3">Exactly-Once Delivery</h3>
                <p className="text-gray-700 mb-4">
                  Our replication guarantees exactly-once semantics, ensuring that every change is applied exactly once to the target database, even in the event of network failures or restarts.
                </p>
              </div>
            </div>
          </section>

          {/* Supported Databases Section */}
          <section id="databases" className="mb-16 scroll-mt-8">
            <h2 className="text-3xl font-bold text-primary mb-6">Supported Databases</h2>

            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-3 px-4 font-semibold text-primary">Database</th>
                    <th className="text-left py-3 px-4 font-semibold text-primary">As Source</th>
                    <th className="text-left py-3 px-4 font-semibold text-primary">As Target</th>
                    <th className="text-left py-3 px-4 font-semibold text-primary">Min Version</th>
                  </tr>
                </thead>
                <tbody className="text-gray-700">
                  <tr className="border-b border-gray-100">
                    <td className="py-3 px-4">PostgreSQL</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4">10+</td>
                  </tr>
                  <tr className="border-b border-gray-100">
                    <td className="py-3 px-4">MySQL</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4">5.7+</td>
                  </tr>
                  <tr className="border-b border-gray-100">
                    <td className="py-3 px-4">Amazon RDS</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4">-</td>
                  </tr>
                  <tr className="border-b border-gray-100">
                    <td className="py-3 px-4">Google Cloud SQL</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4">-</td>
                  </tr>
                  <tr className="border-b border-gray-100">
                    <td className="py-3 px-4">Azure Database</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4 text-green-600">&#10003;</td>
                    <td className="py-3 px-4">-</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          {/* API Section */}
          <section id="api" className="mb-16 scroll-mt-8">
            <h2 className="text-3xl font-bold text-primary mb-6">API Reference</h2>

            <p className="text-gray-700 mb-6">
              The Savegress API allows you to programmatically manage your replication pipelines, monitor status, and retrieve metrics.
            </p>

            <div className="space-y-6">
              <div className="border border-gray-200 rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <span className="bg-green-100 text-green-800 px-2 py-1 rounded text-sm font-mono">GET</span>
                  <code className="text-primary">/api/v1/pipelines</code>
                </div>
                <p className="text-gray-600">List all replication pipelines in your account.</p>
              </div>

              <div className="border border-gray-200 rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-sm font-mono">POST</span>
                  <code className="text-primary">/api/v1/pipelines</code>
                </div>
                <p className="text-gray-600">Create a new replication pipeline.</p>
              </div>

              <div className="border border-gray-200 rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <span className="bg-green-100 text-green-800 px-2 py-1 rounded text-sm font-mono">GET</span>
                  <code className="text-primary">/api/v1/pipelines/:id/metrics</code>
                </div>
                <p className="text-gray-600">Get real-time metrics for a specific pipeline including throughput, latency, and compression ratio.</p>
              </div>
            </div>
          </section>

          {/* Contact Section */}
          <section className="bg-gray-50 rounded-lg p-8 text-center">
            <h2 className="text-2xl font-bold text-primary mb-4">Need Help?</h2>
            <p className="text-gray-600 mb-6">
              Our team is here to help you get the most out of Savegress.
            </p>
            <a
              href="mailto:support@savegress.com"
              className="inline-block bg-accent-blue text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Contact Support
            </a>
          </section>
        </div>
      </div>
    </main>
  )
}
