import type { Metadata } from "next"
import Link from "next/link"

export const metadata: Metadata = {
  title: "About - Savegress",
  description: "Learn about Savegress and our mission to make cross-cloud data replication affordable.",
}

export default function About() {
  return (
    <main className="min-h-screen bg-white">
      <div className="container-custom py-16">
        <div className="max-w-3xl mx-auto">
          <Link
            href="/"
            className="text-accent-blue hover:underline mb-8 inline-block"
          >
            &larr; Back to Home
          </Link>

          <h1 className="text-4xl font-bold text-primary mb-8">About Savegress</h1>

          <div className="prose prose-lg max-w-none">
            <section className="mb-12">
              <h2 className="text-2xl font-semibold text-primary mb-4">Our Mission</h2>
              <p className="text-gray-700 mb-4 text-lg">
                We believe that moving data between clouds shouldn&apos;t cost a fortune. Savegress was built to solve one of the biggest pain points in multi-cloud architectures: egress fees.
              </p>
              <p className="text-gray-700 mb-4">
                Cloud providers charge premium rates for data leaving their networks. For companies operating across AWS, GCP, and Azure, these costs can quickly become a significant portion of their infrastructure budget. We&apos;re here to change that.
              </p>
            </section>

            <section className="mb-12">
              <h2 className="text-2xl font-semibold text-primary mb-4">What We Do</h2>
              <p className="text-gray-700 mb-4">
                Savegress provides Change Data Capture (CDC) technology that streams only the changes from your databases, not the entire dataset. Combined with our advanced compression algorithms, we reduce the amount of data transferred by 10x–150x depending on data patterns.
              </p>
              <p className="text-gray-700 mb-4">
                This means you can replicate your PostgreSQL or MySQL databases across any cloud provider while paying a fraction of traditional egress costs.
              </p>
            </section>

            <section className="mb-12">
              <h2 className="text-2xl font-semibold text-primary mb-4">Why It Matters</h2>
              <div className="grid gap-6">
                <div className="border-l-4 border-accent-blue pl-6">
                  <h3 className="font-semibold text-primary mb-2">Cost Efficiency</h3>
                  <p className="text-gray-600">
                    Stop paying for the same data to be transferred repeatedly. Pay only for actual changes.
                  </p>
                </div>
                <div className="border-l-4 border-accent-orange pl-6">
                  <h3 className="font-semibold text-primary mb-2">Multi-Cloud Freedom</h3>
                  <p className="text-gray-600">
                    Choose the best services from each cloud without being locked in by data gravity.
                  </p>
                </div>
                <div className="border-l-4 border-green-500 pl-6">
                  <h3 className="font-semibold text-primary mb-2">Real-Time Sync</h3>
                  <p className="text-gray-600">
                    Keep your databases in sync with sub-second latency for analytics, disaster recovery, or data locality.
                  </p>
                </div>
              </div>
            </section>

            <section className="mb-12">
              <h2 className="text-2xl font-semibold text-primary mb-4">Our Approach</h2>
              <p className="text-gray-700 mb-4">
                We take a developer-first approach to data replication. Our service is designed to be easy to set up, transparent in pricing, and reliable in operation. No hidden fees, no complex configurations, just efficient data movement.
              </p>
              <ul className="list-disc pl-6 text-gray-700 space-y-2">
                <li>Simple setup with guided configuration</li>
                <li>Pay-as-you-go pricing based on actual data transferred</li>
                <li>Real-time monitoring and alerting</li>
                <li>Enterprise-grade security and compliance</li>
              </ul>
            </section>

            <section className="bg-gray-50 rounded-lg p-8 text-center">
              <h2 className="text-2xl font-bold text-primary mb-4">Get in Touch</h2>
              <p className="text-gray-600 mb-6">
                Have questions or want to learn more about how Savegress can help your organization?
              </p>
              <a
                href="mailto:hello@savegress.com"
                className="inline-block bg-accent-blue text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
              >
                Contact Us
              </a>
            </section>
          </div>
        </div>
      </div>
    </main>
  )
}
