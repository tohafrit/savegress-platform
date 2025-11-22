"use client"

import { Card } from "@/components/ui/card"
import { DollarSign, Lock, Database } from "lucide-react"
import { motion } from "framer-motion"

const problems = [
  {
    icon: DollarSign,
    title: "Egress Fees Add Up",
    description: "Moving data between clouds is expensive",
    details: [
      "AWS charges $0.09/GB for cross-region",
      "Replicating 1TB daily = $2,700/month",
      "Multi-cloud architectures multiply costs",
    ],
  },
  {
    icon: Lock,
    title: "Vendor Lock-in",
    description: "Staying in one cloud limits your options",
    details: [
      "Can't use best-of-breed services",
      "No leverage in pricing negotiations",
      "Disaster recovery across clouds is costly",
    ],
  },
  {
    icon: Database,
    title: "Uncompressed = Wasteful",
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
    <section className="section-padding bg-white">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-heading-lg text-primary mb-4">
            Cloud egress costs are killing your budget
          </h2>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-8">
          {problems.map((problem, index) => (
            <motion.div
              key={problem.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
            >
              <Card className="p-6 h-full hover:shadow-lg transition-shadow">
                <problem.icon className="h-12 w-12 text-accent-orange mb-4" />
                <h3 className="text-heading-sm text-primary mb-2">
                  {problem.title}
                </h3>
                <p className="text-body-md text-neutral-dark-gray mb-4">
                  {problem.description}
                </p>
                <ul className="space-y-2">
                  {problem.details.map((detail) => (
                    <li key={detail} className="text-sm text-neutral-dark-gray flex items-start">
                      <span className="text-accent-orange mr-2">*</span>
                      {detail}
                    </li>
                  ))}
                </ul>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
