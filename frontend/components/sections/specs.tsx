"use client"

import { motion } from "framer-motion"

const specs = [
  { label: "Compression", value: "10x–150x depending on data" },
  { label: "Throughput", value: "50,000+ events per second" },
  { label: "Latency", value: "Under 15ms end-to-end" },
  { label: "Memory", value: "~200MB footprint" },
  { label: "Built with", value: "Go + Rust" },
]

const databases = ["PostgreSQL 14+", "MySQL 8.0+", "MongoDB 4.0+", "SQL Server 2016+", "Oracle 12c+", "MariaDB 10.3+"]

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

const cards = [
  { title: "Key Specifications", content: "specs" },
  { title: "Supported Databases", content: "databases" },
  { title: "Destinations", content: "destinations" },
  { title: "Reliability", content: "reliability" },
]

export function Specs() {
  return (
    <section className="section-padding bg-dark-bg relative">
      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-h2">Built for production</h2>
        </motion.div>

        <div className="flex flex-wrap justify-center gap-8">
          {cards.map((card, index) => (
            <motion.div
              key={card.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
              className="w-full md:w-auto"
            >
              <div className="specs-card p-8">
                {/* Background images */}
                {card.content === 'specs' && (
                  <img src="/images/spec-1.png" alt="" className="absolute bottom-0 left-1/2 -translate-x-1/2 pointer-events-none" />
                )}
                {card.content === 'databases' && (
                  <img src="/images/spec-2.png" alt="" className="absolute bottom-0 right-0 pointer-events-none" />
                )}
                {card.content === 'destinations' && (
                  <img src="/images/spec-3.png" alt="" className="absolute top-0 right-0 pointer-events-none" />
                )}
                {card.content === 'reliability' && (
                  <img src="/images/spec-4.png" alt="" className="absolute right-0 top-1/2 -translate-y-1/2 pointer-events-none" />
                )}

                {/* Header */}
                <h3 className="text-h4 mb-6 relative z-10">{card.title}</h3>

                {/* Content */}
                {card.content === 'specs' && (
                  <div className="flex justify-between relative z-10">
                    <div className="text-content-2 text-cyan w-[230px]">
                      {specs.map((spec) => <div key={spec.label}>{spec.label}</div>)}
                    </div>
                    <div className="text-content-2 text-grey text-right">
                      {specs.map((spec) => <div key={spec.value}>{spec.value}</div>)}
                    </div>
                  </div>
                )}

                {card.content === 'databases' && (
                  <div>
                    {databases.map((db) => (
                      <div key={db} className="flex items-center gap-3">
                        <span className="list-plus">+</span>
                        <span className="text-content-2 text-grey">{db}</span>
                      </div>
                    ))}
                  </div>
                )}

                {card.content === 'destinations' && (
                  <div>
                    {destinations.map((dest) => (
                      <div key={dest} className="flex items-center gap-3">
                        <svg width="8" height="9" viewBox="0 0 8 9" fill="none" xmlns="http://www.w3.org/2000/svg" className="flex-shrink-0">
                          <path d="M7.5 4.33008L0 -4.95911e-05V8.66021L7.5 4.33008Z" fill="#00B4D8"/>
                        </svg>
                        <span className="text-content-2 text-grey">{dest}</span>
                      </div>
                    ))}
                  </div>
                )}

                {card.content === 'reliability' && (
                  <div>
                    {reliability.map((item) => (
                      <div key={item} className="flex items-center gap-3">
                        <span className="list-plus">+</span>
                        <span className="text-content-2 text-grey">{item}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
