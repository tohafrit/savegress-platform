import type { Metadata } from "next"
import Link from "next/link"
import Image from "next/image"

export const metadata: Metadata = {
  title: "Privacy Policy - Savegress",
  description: "Privacy Policy for Savegress - how we collect, use, and protect your data.",
}

export default function PrivacyPolicy() {
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
          <h1 className="text-h2 text-white mb-4">Privacy Policy</h1>
          <p className="text-content-1 text-grey mb-12">Last updated: November 2025</p>

          <div className="space-y-10">
            <section>
              <h2 className="text-h4 text-white mb-4">1. Introduction</h2>
              <p className="text-content-1 text-grey">
                Savegress (&quot;we&quot;, &quot;our&quot;, or &quot;us&quot;) is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you visit our website savegress.com and use our services.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">2. Information We Collect</h2>

              <h3 className="text-content-1 text-white mb-3">2.1 Information You Provide</h3>
              <ul className="list-disc pl-6 text-content-1 text-grey mb-6 space-y-2">
                <li>Account information: email address, name, company name</li>
                <li>Payment information: billing address, payment method details (processed securely through our payment providers)</li>
                <li>Communications: information you provide when contacting our support team</li>
              </ul>

              <h3 className="text-content-1 text-white mb-3">2.2 Information Collected Automatically</h3>
              <ul className="list-disc pl-6 text-content-1 text-grey space-y-2">
                <li>Usage data: pages visited, features used, time spent on the service</li>
                <li>Device information: browser type, operating system, IP address</li>
                <li>Cookies and similar technologies for analytics and service improvement</li>
              </ul>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">3. How We Use Your Information</h2>
              <p className="text-content-1 text-grey mb-4">We use the collected information to:</p>
              <ul className="list-disc pl-6 text-content-1 text-grey space-y-2">
                <li>Provide, maintain, and improve our services</li>
                <li>Process transactions and send related information</li>
                <li>Send technical notices, updates, and support messages</li>
                <li>Respond to your comments, questions, and requests</li>
                <li>Monitor and analyze trends, usage, and activities</li>
                <li>Detect, investigate, and prevent fraudulent transactions and other illegal activities</li>
              </ul>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">4. Data Sharing and Disclosure</h2>
              <p className="text-content-1 text-grey mb-4">We do not sell your personal information. We may share your information with:</p>
              <ul className="list-disc pl-6 text-content-1 text-grey space-y-2">
                <li>Service providers who assist in operating our services (hosting, payment processing, analytics)</li>
                <li>Professional advisors (lawyers, accountants) as necessary</li>
                <li>Law enforcement or government agencies when required by law</li>
                <li>Other parties in connection with a merger, acquisition, or sale of assets</li>
              </ul>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">5. Data Security</h2>
              <p className="text-content-1 text-grey">
                We implement appropriate technical and organizational measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction. This includes encryption of data in transit and at rest, regular security assessments, and access controls.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">6. Data Retention</h2>
              <p className="text-content-1 text-grey">
                We retain your personal information for as long as your account is active or as needed to provide you services. We will retain and use your information as necessary to comply with our legal obligations, resolve disputes, and enforce our agreements.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">7. Your Rights</h2>
              <p className="text-content-1 text-grey mb-4">Depending on your location, you may have the right to:</p>
              <ul className="list-disc pl-6 text-content-1 text-grey mb-4 space-y-2">
                <li>Access the personal information we hold about you</li>
                <li>Request correction of inaccurate data</li>
                <li>Request deletion of your data</li>
                <li>Object to or restrict processing of your data</li>
                <li>Data portability</li>
                <li>Withdraw consent at any time</li>
              </ul>
              <p className="text-content-1 text-grey">
                To exercise these rights, please contact us at contact@savegress.com.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">8. Cookies</h2>
              <p className="text-content-1 text-grey">
                We use cookies and similar tracking technologies to collect and track information and to improve our service. You can instruct your browser to refuse all cookies or to indicate when a cookie is being sent. However, if you do not accept cookies, you may not be able to use some portions of our service.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">9. International Data Transfers</h2>
              <p className="text-content-1 text-grey">
                Your information may be transferred to and maintained on servers located outside of your country. We ensure that such transfers comply with applicable data protection laws and that your data remains protected.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">10. Children&apos;s Privacy</h2>
              <p className="text-content-1 text-grey">
                Our service is not directed to individuals under the age of 16. We do not knowingly collect personal information from children under 16. If you become aware that a child has provided us with personal information, please contact us.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">11. Changes to This Policy</h2>
              <p className="text-content-1 text-grey">
                We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the &quot;Last updated&quot; date.
              </p>
            </section>

            <section>
              <h2 className="text-h4 text-white mb-4">12. Contact Us</h2>
              <p className="text-content-1 text-grey mb-4">
                If you have any questions about this Privacy Policy, please contact us at:
              </p>
              <p className="text-content-1">
                <a href="mailto:contact@savegress.com" className="text-cyan hover:text-cyan/80">contact@savegress.com</a>
              </p>
            </section>
          </div>
        </div>
      </div>
    </main>
  )
}
