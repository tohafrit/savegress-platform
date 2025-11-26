"use client"

import { motion } from "framer-motion"
import Image from "next/image"

const solutions = [
  {
    image: "/images/card-costs.png",
    title: "Cut Egress Costs",
    subtitle: "10x–150x compression depending on data patterns",
    details: [
      "1TB becomes 7-100GB after compression",
      "Pay for kilobytes, not gigabytes",
      "ROI visible on your first cloud bill",
    ],
    highlight: "$2,700/mo → $225/mo (typical 12x)",
  },
  {
    image: "/images/card-multicloud.png",
    title: "True Multi-Cloud",
    subtitle: "Replicate anywhere without lock-in",
    details: [
      "AWS ↔ GCP ↔ Azure ↔ On-prem",
      "Same tool, any destination",
      "Freedom to choose best services",
    ],
  },
  {
    image: "/images/card-sync.png",
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
    <section className="section-padding bg-dark-bg relative">
      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-h2">Compress before you transfer</h2>
        </motion.div>

        <div className="flex flex-wrap justify-center gap-8">
          {solutions.map((solution, index) => (
            <motion.div
              key={solution.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
              className="w-full md:w-auto"
            >
              <div className="solution-card p-8">
                {/* Image - fixed height container */}
                <div className="h-[120px] mb-6 flex items-center justify-center">
                  <Image
                    src={solution.image}
                    alt=""
                    width={200}
                    height={150}
                    className="w-auto max-h-[120px] object-contain icon-hover"
                  />
                </div>

                {/* Title */}
                <h3 className="text-h4 text-cyan mb-2">{solution.title}</h3>

                <p className="text-content-1 text-cyan mb-6">{solution.subtitle}</p>

                <ul className="space-y-3">
                  {solution.details.map((detail) => (
                    <li key={detail} className="flex items-start gap-3 text-content-1 text-grey">
                      <span className="list-plus flex-shrink-0">+</span>
                      {detail}
                    </li>
                  ))}
                </ul>
              </div>

              {solution.highlight && (
                <div className="highlight-box mt-4">{solution.highlight}</div>
              )}
            </motion.div>
          ))}
        </div>
      </div>

    </section>
  )
}
