"use client"

import { motion } from "framer-motion"
import Image from "next/image"

const problems = [
  {
    icon: "/images/icon-dollar.svg",
    title: "Egress Fees Add Up",
    description: "Moving data between clouds is expensive",
    details: [
      "AWS charges $0.09/GB for cross-region",
      "Replicating 1TB daily = $2,700/month",
      "Multi-cloud architectures multiply costs",
    ],
  },
  {
    icon: "/images/icon-lock.svg",
    title: "Vendor Lock-in",
    description: "Staying in one cloud limits your options",
    details: [
      "Can't use best-of-breed services",
      "No leverage in pricing negotiations",
      "Disaster recovery across clouds is costly",
    ],
  },
  {
    icon: "/images/icon-waste.svg",
    title: "Uncompressed =\nWasteful",
    description: "Raw database changes are bloated",
    details: [
      "Timestamps repeat constantly",
      "Status fields rarely change",
      "You're paying to transfer redundant data",
    ],
  },
]

export function Problem() {
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
          <h2 className="text-h2">
            Cloud egress costs are killing your budget
          </h2>
        </motion.div>

        <div className="flex flex-wrap justify-center gap-8">
          {problems.map((problem, index) => (
            <motion.div
              key={problem.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
            >
              <div className="problem-card p-8">
                {/* Icon */}
                <div className="mb-6">
                  <Image
                    src={problem.icon}
                    alt=""
                    width={63}
                    height={63}
                    className="w-[63px] h-[63px] icon-hover"
                  />
                </div>

                <h3 className="text-h4 whitespace-pre-line mb-3">
                  {problem.title}
                </h3>
                <p className="text-content-1 text-cyan mb-6">
                  {problem.description}
                </p>
                <ul className="space-y-3">
                  {problem.details.map((detail) => (
                    <li key={detail} className="flex items-start gap-3 text-content-1 text-grey">
                      <span className="dot-cyan flex-shrink-0 mt-[10px]" />
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
