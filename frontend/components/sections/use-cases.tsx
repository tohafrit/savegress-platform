"use client"

import { Card } from "@/components/ui/card"
import { Cloud, Shield, TruckIcon, BarChart } from "lucide-react"
import { motion } from "framer-motion"

const useCases = [
  {
    icon: Cloud,
    title: "Multi-Cloud Data Sync",
    description: "Keep data consistent across clouds",
    details: [
      "AWS to GCP replication",
      "Azure to on-prem backup",
      "Compressed transfers save money",
    ],
  },
  {
    icon: Shield,
    title: "Disaster Recovery",
    description: "Affordable cross-region DR",
    details: [
      "Real-time replicas in another cloud",
      "Compression cuts DR costs dramatically",
      "Failover-ready at all times",
    ],
  },
  {
    icon: TruckIcon,
    title: "Data Migration",
    description: "Move to a new cloud without downtime",
    details: [
      "Sync continuously during migration",
      "Switch over when ready",
      "No big-bang cutover risk",
    ],
  },
  {
    icon: BarChart,
    title: "Analytics Pipeline",
    description: "Feed your data warehouse in real-time",
    details: [
      "Stream changes as they happen",
      "Compressed data = faster transfers",
      "Fresh data for better decisions",
    ],
  },
]

export function UseCases() {
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
            Where teams use Savegress
          </h2>
        </motion.div>

        <div className="grid md:grid-cols-2 gap-8">
          {useCases.map((useCase, index) => (
            <motion.div
              key={useCase.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
            >
              <Card className="p-6 h-full hover:shadow-lg transition-shadow">
                <useCase.icon className="h-10 w-10 text-accent-orange mb-4" />
                <h3 className="text-heading-sm text-primary mb-2">
                  {useCase.title}
                </h3>
                <p className="text-body-md text-neutral-dark-gray mb-4">
                  {useCase.description}
                </p>
                <ul className="space-y-2">
                  {useCase.details.map((detail) => (
                    <li key={detail} className="text-sm text-neutral-dark-gray flex items-start">
                      <span className="text-primary mr-2">*</span>
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
