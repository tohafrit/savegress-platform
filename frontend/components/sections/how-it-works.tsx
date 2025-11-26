"use client"

import { motion } from "framer-motion"

const steps = [
  {
    number: 1,
    title: "Capture",
    subtitle: "Connect to your database",
    details: [
      "PostgreSQL and MySQL supported",
      "Every change captured in real-time",
      "Schema changes tracked automatically",
      "Guaranteed delivery - no data loss",
    ],
  },
  {
    number: 2,
    title: "Compress",
    subtitle: "Shrink your data automatically",
    details: [
      "Up to 200x smaller",
      "Optimized for database patterns",
      "Less storage, lower costs",
    ],
  },
  {
    number: 3,
    title: "Deliver",
    subtitle: "Send anywhere",
    details: [
      "Any HTTP endpoint",
      "Message brokers (Kafka, NATS, Redis)",
      "File export for batch processing",
      "Your custom destination via plugins",
    ],
  },
]

export function HowItWorks() {
  return (
    <section id="how-it-works" className="section-padding bg-dark-bg-secondary relative overflow-hidden">
      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-h2">Three steps to real-time data</h2>
        </motion.div>

        <div className="flex flex-wrap justify-center gap-8">
          {steps.map((step, index) => (
            <motion.div
              key={step.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.2 }}
            >
              <div className="step-card p-8 flex flex-col">
                {/* Background number */}
                <span className="step-number-bg">{step.number}</span>

                {/* Arrow icon */}
                <svg
                  className="absolute bottom-6 right-6"
                  width="50"
                  height="50"
                  viewBox="0 0 50 50"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <rect x="0.5" y="0.5" width="49" height="49" rx="24.5" stroke="#00B4D8"/>
                  <path d="M18 32.5076L31.5076 19M31.5076 19V30.3888M31.5076 19H20.3837" stroke="#00B4D8" strokeLinecap="round"/>
                </svg>

                {/* Content */}
                <h3 className="text-h3 mb-2">{step.title}</h3>
                <p className="text-content-1 text-cyan mb-6">{step.subtitle}</p>

                <ul className="space-y-3 flex-1">
                  {step.details.map((detail) => (
                    <li key={detail} className="flex items-start gap-3 text-content-1 text-grey">
                      <svg
                        className="flex-shrink-0 mt-[10px]"
                        width="8"
                        height="9"
                        viewBox="0 0 8 9"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path d="M7.5 4.33008L0 -4.95911e-05V8.66021L7.5 4.33008Z" fill="#00B4D8"/>
                      </svg>
                      {detail}
                    </li>
                  ))}
                </ul>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
