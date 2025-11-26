"use client"

import { motion } from "framer-motion"

const features = [
  {
    icon: "security",
    title: "Enterprise Security",
    items: [
      { label: "RBAC", value: "Admin, Operator, Viewer, Developer roles" },
      { label: "Encryption", value: "AES-256-GCM end-to-end" },
      { label: "Vault", value: "HashiCorp Vault integration" },
      { label: "Audit", value: "Immutable audit trail" },
      { label: "TLS/mTLS", value: "Mutual TLS for all connections" },
    ],
  },
  {
    icon: "ha",
    title: "High Availability",
    items: [
      { label: "Consensus", value: "Raft-based cluster coordination" },
      { label: "Failover", value: "Automatic leader election" },
      { label: "PITR", value: "Point-in-Time Recovery" },
      { label: "Split-brain", value: "Detection and prevention" },
    ],
  },
  {
    icon: "databases",
    title: "8 Database Sources",
    items: [
      { label: "PostgreSQL", value: "14+ (Logical Replication)" },
      { label: "MySQL", value: "8.0+ (Binlog)" },
      { label: "MongoDB", value: "4.0+ (Change Streams)" },
      { label: "SQL Server", value: "2016+ (CDC Tables)" },
      { label: "Oracle", value: "12c+ (LogMiner)" },
      { label: "More", value: "MariaDB, Cassandra, DynamoDB" },
    ],
  },
  {
    icon: "observability",
    title: "Full Observability",
    items: [
      { label: "Metrics", value: "Prometheus /metrics endpoint" },
      { label: "Tracing", value: "OpenTelemetry distributed tracing" },
      { label: "Health", value: "Liveness & Readiness probes" },
      { label: "Alerts", value: "PagerDuty, Slack integration" },
    ],
  },
]

const IconSecurity = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M24 4L6 12V22C6 33.1 13.68 43.34 24 46C34.32 43.34 42 33.1 42 22V12L24 4Z" stroke="#00B4D8" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M18 24L22 28L30 20" stroke="#00B4D8" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const IconHA = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="24" cy="12" r="6" stroke="#00B4D8" strokeWidth="2"/>
    <circle cx="12" cy="36" r="6" stroke="#00B4D8" strokeWidth="2"/>
    <circle cx="36" cy="36" r="6" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M24 18V24L18 30" stroke="#00B4D8" strokeWidth="2" strokeLinecap="round"/>
    <path d="M24 24L30 30" stroke="#00B4D8" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const IconDatabases = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <ellipse cx="24" cy="10" rx="16" ry="6" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M8 10V24C8 27.31 15.16 30 24 30C32.84 30 40 27.31 40 24V10" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M8 24V38C8 41.31 15.16 44 24 44C32.84 44 40 41.31 40 38V24" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M8 17C8 20.31 15.16 23 24 23C32.84 23 40 20.31 40 17" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M8 31C8 34.31 15.16 37 24 37C32.84 37 40 34.31 40 31" stroke="#00B4D8" strokeWidth="2"/>
  </svg>
)

const IconObservability = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="6" y="6" width="36" height="36" rx="4" stroke="#00B4D8" strokeWidth="2"/>
    <path d="M12 30L18 24L24 28L30 18L36 22" stroke="#00B4D8" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <circle cx="18" cy="24" r="2" fill="#00B4D8"/>
    <circle cx="24" cy="28" r="2" fill="#00B4D8"/>
    <circle cx="30" cy="18" r="2" fill="#00B4D8"/>
  </svg>
)

const icons: Record<string, () => JSX.Element> = {
  security: IconSecurity,
  ha: IconHA,
  databases: IconDatabases,
  observability: IconObservability,
}

export function EnterpriseFeatures() {
  return (
    <section className="section-padding bg-dark-bg-secondary">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-h2">Enterprise-grade capabilities</h2>
          <p className="text-content-1 text-grey mt-4 max-w-2xl mx-auto">
            Built for production workloads with security, reliability, and observability from day one
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-[1232px] mx-auto">
          {features.map((feature, index) => {
            const Icon = icons[feature.icon]
            return (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: index * 0.1 }}
                className="h-full"
              >
                <div className="enterprise-card p-8 h-full flex flex-col">
                  {/* Icon */}
                  <div className="mb-6">
                    <Icon />
                  </div>

                  {/* Title */}
                  <h3 className="text-h4 mb-6">{feature.title}</h3>

                  {/* Features list */}
                  <div className="space-y-3 flex-1">
                    {feature.items.map((item) => (
                      <div key={item.label} className="flex justify-between gap-4">
                        <span className="text-content-2 text-cyan whitespace-nowrap">{item.label}</span>
                        <span className="text-content-2 text-grey text-right">{item.value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </motion.div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
