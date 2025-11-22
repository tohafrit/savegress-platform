import type { Metadata } from "next"
import Link from "next/link"

export const metadata: Metadata = {
  title: "Terms of Service - Savegress",
  description: "Terms of Service for Savegress - the rules and guidelines for using our services.",
}

export default function TermsOfService() {
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

          <h1 className="text-4xl font-bold text-primary mb-8">Terms of Service</h1>
          <p className="text-gray-600 mb-8">Last updated: {new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}</p>

          <div className="prose prose-lg max-w-none">
            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">1. Agreement to Terms</h2>
              <p className="text-gray-700 mb-4">
                By accessing or using Savegress (&quot;Service&quot;), you agree to be bound by these Terms of Service (&quot;Terms&quot;). If you disagree with any part of these terms, you may not access the Service.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">2. Description of Service</h2>
              <p className="text-gray-700 mb-4">
                Savegress provides a Change Data Capture (CDC) and data replication service that enables streaming of database changes between cloud providers (AWS, GCP, Azure) with compression to reduce data transfer costs.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">3. Account Registration</h2>
              <p className="text-gray-700 mb-4">To use our Service, you must:</p>
              <ul className="list-disc pl-6 text-gray-700 mb-4 space-y-2">
                <li>Be at least 18 years old or have legal authority to enter into this agreement</li>
                <li>Provide accurate and complete registration information</li>
                <li>Maintain the security of your account credentials</li>
                <li>Promptly notify us of any unauthorized access to your account</li>
                <li>Be responsible for all activities that occur under your account</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">4. Acceptable Use</h2>
              <p className="text-gray-700 mb-4">You agree not to:</p>
              <ul className="list-disc pl-6 text-gray-700 mb-4 space-y-2">
                <li>Use the Service for any unlawful purpose or in violation of any laws</li>
                <li>Attempt to gain unauthorized access to the Service or its related systems</li>
                <li>Interfere with or disrupt the integrity or performance of the Service</li>
                <li>Transmit any malware, viruses, or other harmful code</li>
                <li>Use the Service to infringe on intellectual property rights of others</li>
                <li>Resell or redistribute the Service without our written consent</li>
                <li>Use the Service to store or transmit data that violates applicable privacy laws</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">5. Data and Privacy</h2>
              <p className="text-gray-700 mb-4">
                Our collection and use of personal information is governed by our <Link href="/privacy" className="text-accent-blue hover:underline">Privacy Policy</Link>. By using the Service, you consent to such processing and you warrant that all data provided by you is accurate.
              </p>
              <p className="text-gray-700 mb-4">
                You retain all rights to your data. We do not claim ownership over any data you transmit through our Service. You grant us a limited license to process your data solely for the purpose of providing the Service.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">6. Payment Terms</h2>
              <ul className="list-disc pl-6 text-gray-700 mb-4 space-y-2">
                <li>Fees are based on usage as described in our pricing documentation</li>
                <li>All fees are exclusive of taxes unless stated otherwise</li>
                <li>Payment is due according to the billing cycle you select</li>
                <li>We reserve the right to change pricing with 30 days&apos; notice</li>
                <li>Failure to pay may result in suspension or termination of your account</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">7. Service Level and Support</h2>
              <p className="text-gray-700 mb-4">
                We strive to maintain high availability of our Service. However, we do not guarantee uninterrupted access. Scheduled maintenance will be communicated in advance when possible.
              </p>
              <p className="text-gray-700 mb-4">
                Support is provided through our designated channels during business hours. Response times and support levels may vary based on your subscription plan.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">8. Intellectual Property</h2>
              <p className="text-gray-700 mb-4">
                The Service and its original content, features, and functionality are owned by Savegress and are protected by international copyright, trademark, patent, trade secret, and other intellectual property laws.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">9. Limitation of Liability</h2>
              <p className="text-gray-700 mb-4">
                TO THE MAXIMUM EXTENT PERMITTED BY LAW, SAVEGRESS SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS OR REVENUES, WHETHER INCURRED DIRECTLY OR INDIRECTLY, OR ANY LOSS OF DATA, USE, GOODWILL, OR OTHER INTANGIBLE LOSSES.
              </p>
              <p className="text-gray-700 mb-4">
                Our total liability for any claims under these Terms shall not exceed the amount you paid us in the twelve (12) months preceding the claim.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">10. Disclaimer of Warranties</h2>
              <p className="text-gray-700 mb-4">
                THE SERVICE IS PROVIDED &quot;AS IS&quot; AND &quot;AS AVAILABLE&quot; WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">11. Indemnification</h2>
              <p className="text-gray-700 mb-4">
                You agree to indemnify and hold harmless Savegress and its officers, directors, employees, and agents from any claims, damages, losses, liabilities, costs, or expenses (including reasonable attorneys&apos; fees) arising from your use of the Service or violation of these Terms.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">12. Termination</h2>
              <p className="text-gray-700 mb-4">
                We may terminate or suspend your account and access to the Service immediately, without prior notice, for conduct that we believe violates these Terms or is harmful to other users, us, or third parties, or for any other reason at our sole discretion.
              </p>
              <p className="text-gray-700 mb-4">
                You may terminate your account at any time by contacting us. Upon termination, your right to use the Service will cease immediately.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">13. Governing Law</h2>
              <p className="text-gray-700 mb-4">
                These Terms shall be governed by and construed in accordance with the laws of the jurisdiction in which Savegress is registered, without regard to its conflict of law provisions.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">14. Dispute Resolution</h2>
              <p className="text-gray-700 mb-4">
                Any disputes arising from these Terms or the Service shall first be attempted to be resolved through good-faith negotiations. If negotiations fail, disputes shall be resolved through binding arbitration in accordance with applicable arbitration rules.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">15. Changes to Terms</h2>
              <p className="text-gray-700 mb-4">
                We reserve the right to modify these Terms at any time. We will provide notice of significant changes by posting the new Terms on this page and updating the &quot;Last updated&quot; date. Your continued use of the Service after changes constitutes acceptance of the new Terms.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">16. Severability</h2>
              <p className="text-gray-700 mb-4">
                If any provision of these Terms is held to be unenforceable, the remaining provisions will continue in full force and effect.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-primary mb-4">17. Contact Us</h2>
              <p className="text-gray-700 mb-4">
                If you have any questions about these Terms, please contact us at:
              </p>
              <p className="text-gray-700">
                Email: <a href="mailto:legal@savegress.com" className="text-accent-blue hover:underline">legal@savegress.com</a>
              </p>
            </section>
          </div>
        </div>
      </div>
    </main>
  )
}
