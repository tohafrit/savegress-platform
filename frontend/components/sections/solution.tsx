"use client"

import { Card } from "@/components/ui/card"
import { TrendingDown, Globe, Zap } from "lucide-react"
import { motion } from "framer-motion"

const solutions = [
  {
    icon: TrendingDown,
    title: "Cut Egress Costs",
    subtitle: "Up to 200x compression = up to 200x savings",
    details: [
      "1TB becomes 5-50GB after compression",
      "Pay for kilobytes, not gigabytes",
      "ROI visible on your first cloud bill",
    ],
    highlight: "$2,700/mo -> $135/mo (at 20x)",
  },
  {
    icon: Globe,
    title: "True Multi-Cloud",
    subtitle: "Replicate anywhere without lock-in",
    details: [
      "AWS <-> GCP <-> Azure <-> On-prem",
      "Same tool, any destination",
      "Freedom to choose best services",
    ],
  },
  {
    icon: Zap,
    title: "Real-Time Sync",
    subtitle: "Changes arrive in milliseconds, not minutes",
    details: [
      "Capture every INSERT, UPDATE, DELETE",
      "Stream to any cloud or region",
      "Always-fresh replicas",
    ],
  },
]

export function Solution() {
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
            Compress before you transfer
          </h2>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-8">
          {solutions.map((solution, index) => (
            <motion.div
              key={solution.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
            >
              <Card className="p-6 h-full hover:shadow-lg transition-shadow">
                <solution.icon className="h-12 w-12 text-primary mb-4" />
                <h3 className="text-heading-sm text-primary mb-2">
                  {solution.title}
                </h3>
                <p className="text-body-md text-accent-orange font-semibold mb-4">
                  {solution.subtitle}
                </p>
                <ul className="space-y-2 mb-4">
                  {solution.details.map((detail) => (
                    <li key={detail} className="text-sm text-neutral-dark-gray flex items-start">
                      <span className="text-primary mr-2">+</span>
                      {detail}
                    </li>
                  ))}
                </ul>
                {solution.highlight && (
                  <div className="mt-4 p-3 bg-accent-orange/10 rounded-lg">
                    <p className="text-sm font-mono text-primary font-semibold">
                      {solution.highlight}
                    </p>
                  </div>
                )}
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
