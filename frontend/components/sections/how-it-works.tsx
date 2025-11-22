"use client"

import { motion } from "framer-motion"
import { Database, Minimize2, Send } from "lucide-react"

const steps = [
  {
    icon: Database,
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
    icon: Minimize2,
    title: "Compress",
    subtitle: "Shrink your data automatically",
    details: [
      "Up to 200x smaller",
      "Optimized for database patterns",
      "Less storage, lower costs",
    ],
  },
  {
    icon: Send,
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
    <section id="how-it-works" className="section-padding bg-white">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-heading-lg text-primary mb-4">
            Three steps to real-time data
          </h2>
        </motion.div>

        <div className="flex flex-col md:flex-row items-start justify-between gap-8 md:gap-4">
          {steps.map((step, index) => (
            <motion.div
              key={step.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.2 }}
              className="flex-1"
            >
              <div className="flex flex-col items-center md:items-start">
                {/* Step number */}
                <div className="flex items-center gap-4 mb-4">
                  <div className="flex items-center justify-center w-12 h-12 rounded-full bg-primary text-white font-bold text-xl">
                    {index + 1}
                  </div>
                  <step.icon className="h-10 w-10 text-accent-orange" />
                </div>

                {/* Content */}
                <h3 className="text-heading-sm text-primary mb-2">
                  {step.title}
                </h3>
                <p className="text-body-md text-neutral-dark-gray font-semibold mb-4">
                  {step.subtitle}
                </p>
                <ul className="space-y-2">
                  {step.details.map((detail) => (
                    <li key={detail} className="text-sm text-neutral-dark-gray flex items-start">
                      <span className="text-primary mr-2">*</span>
                      {detail}
                    </li>
                  ))}
                </ul>
              </div>

              {/* Arrow (except last step) */}
              {index < steps.length - 1 && (
                <div className="hidden md:flex items-center justify-center mt-8">
                  <div className="text-4xl text-gray-300">-&gt;</div>
                </div>
              )}
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
