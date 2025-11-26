import type { Metadata } from "next"
import Link from "next/link"
import Image from "next/image"

export const metadata: Metadata = {
  title: "About - Savegress",
  description: "Learn about Savegress and our mission to make cross-cloud data replication affordable.",
}

export default function About() {
  return (
    <main className="min-h-screen bg-dark-bg">
      {/* Header */}
      <header className="bg-dark-surface border-b border-white/10">
        <div className="container-custom py-4 md:py-6 flex items-center justify-between">
          <Link href="/">
            <Image
              src="/images/logo.svg"
              alt="Savegress"
              width={160}
              height={45}
              className="h-8 md:h-10 w-auto"
            />
          </Link>
          <Link
            href="/"
            className="text-grey hover:text-white transition-colors text-sm"
          >
            &larr; Back to Home
          </Link>
        </div>
      </header>

      <div className="container-custom py-12 md:py-20">
        <div className="max-w-3xl mx-auto">
          <h1 className="text-h2 text-white mb-8">About Savegress</h1>

          <div className="space-y-12">
            <section>
              <h2 className="text-h4 text-white mb-4">Our Mission</h2>
              <p className="text-content-1 text-grey mb-4">
                We believe that moving data between clouds shouldn&apos;t cost a fortune. Savegress was built to solve one of the biggest pain points in multi-cloud architectures: egress fees.
              </p>
              <p className="text-content-1 text-grey">
                Cloud providers charge premium rates for data leaving their networks. For companies operating across AWS, GCP, and Azure, these costs can quickly become a significant portion of their infrastructure budget. We&apos;re here to change that.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">What We Do</h2>
              <p className="text-content-1 text-grey mb-4">
                Savegress provides Change Data Capture (CDC) technology that streams only the changes from your databases, not the entire dataset. Combined with our advanced compression algorithms, we reduce the amount of data transferred by 10x–150x depending on data patterns.
              </p>
              <p className="text-content-1 text-grey">
                This means you can replicate your PostgreSQL or MySQL databases across any cloud provider while paying a fraction of traditional egress costs.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">Why It Matters</h2>
              <div className="grid gap-4">
                <div className="card-dark p-6 border-l-2 border-cyan">
                  <h3 className="text-content-1 text-white mb-2">Cost Efficiency</h3>
                  <p className="text-content-1 text-grey">
                    Stop paying for the same data to be transferred repeatedly. Pay only for actual changes.
                  </p>
                </div>
                <div className="card-dark p-6 border-l-2 border-accent-orange">
                  <h3 className="text-content-1 text-white mb-2">Multi-Cloud Freedom</h3>
                  <p className="text-content-1 text-grey">
                    Choose the best services from each cloud without being locked in by data gravity.
                  </p>
                </div>
                <div className="card-dark p-6 border-l-2 border-green-500">
                  <h3 className="text-content-1 text-white mb-2">Real-Time Sync</h3>
                  <p className="text-content-1 text-grey">
                    Keep your databases in sync with sub-second latency for analytics, disaster recovery, or data locality.
                  </p>
                </div>
              </div>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">Our Approach</h2>
              <p className="text-content-1 text-grey mb-4">
                We take a developer-first approach to data replication. Our service is designed to be easy to set up, transparent in pricing, and reliable in operation. No hidden fees, no complex configurations, just efficient data movement.
              </p>
              <ul className="list-disc pl-6 text-content-1 text-grey space-y-2">
                <li>Simple setup with guided configuration</li>
                <li>Pay-as-you-go pricing based on actual data transferred</li>
                <li>Real-time monitoring and alerting</li>
                <li>Enterprise-grade security and compliance</li>
              </ul>
            </section>

            <section className="card-dark p-8 text-center">
              <h2 className="text-h4 text-white mb-4">Get in Touch</h2>
              <p className="text-content-1 text-grey mb-6">
                Have questions or want to learn more about how Savegress can help your organization?
              </p>
              <a
                href="mailto:contact@savegress.com"
                className="btn-primary inline-block px-8 py-3"
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
