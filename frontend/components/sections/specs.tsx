"use client"

import { Card } from "@/components/ui/card"
import { motion } from "framer-motion"

const specs = [
  { label: "Compression", value: "Up to 200x smaller transfers" },
  { label: "Throughput", value: "50,000+ events per second" },
  { label: "Latency", value: "Under 15ms end-to-end" },
  { label: "Memory", value: "~200MB footprint" },
  { label: "Built with", value: "Go + Rust" },
]

const databases = ["PostgreSQL 12+", "MySQL 5.7+ / 8.0+"]

const destinations = [
  "Any cloud (AWS, GCP, Azure)",
  "On-premises data centers",
  "Message brokers (Kafka, NATS, Redis)",
  "HTTP endpoints and webhooks",
]

const reliability = [
  "Exactly-once delivery",
  "Automatic recovery after restarts",
  "No data loss, guaranteed",
]

export function Specs() {
  return (
    <section className="section-padding bg-neutral-light-gray">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-heading-lg text-primary mb-4">
            Built for production
          </h2>
        </motion.div>

        <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          {/* Key specs */}
          <Card className="p-6">
            <h3 className="text-heading-sm text-primary mb-4">Key Specifications</h3>
            <div className="space-y-3">
              {specs.map((spec) => (
                <div key={spec.label} className="flex justify-between items-center">
                  <span className="text-sm font-medium text-neutral-dark-gray">{spec.label}</span>
                  <span className="text-sm font-mono text-primary">{spec.value}</span>
                </div>
              ))}
            </div>
          </Card>

          {/* Supported databases */}
          <Card className="p-6">
            <h3 className="text-heading-sm text-primary mb-4">Supported Databases</h3>
            <ul className="space-y-2">
              {databases.map((db) => (
                <li key={db} className="text-sm text-neutral-dark-gray flex items-start">
                  <span className="text-primary mr-2">+</span>
                  {db}
                </li>
              ))}
            </ul>
          </Card>

          {/* Destinations */}
          <Card className="p-6">
            <h3 className="text-heading-sm text-primary mb-4">Destinations</h3>
            <ul className="space-y-2">
              {destinations.map((dest) => (
                <li key={dest} className="text-sm text-neutral-dark-gray flex items-start">
                  <span className="text-primary mr-2">-&gt;</span>
                  {dest}
                </li>
              ))}
            </ul>
          </Card>

          {/* Reliability */}
          <Card className="p-6">
            <h3 className="text-heading-sm text-primary mb-4">Reliability</h3>
            <ul className="space-y-2">
              {reliability.map((item) => (
                <li key={item} className="text-sm text-neutral-dark-gray flex items-start">
                  <span className="text-primary mr-2">+</span>
                  {item}
                </li>
              ))}
            </ul>
          </Card>
        </div>
      </div>
    </section>
  )
}
