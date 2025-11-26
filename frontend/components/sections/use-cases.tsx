"use client"

import { motion } from "framer-motion"

const useCases = [
  {
    icon: "/images/icon-sync.svg",
    title: "Multi-Cloud Data Sync",
    description: "Keep data consistent across clouds",
    details: [
      "AWS to GCP replication",
      "Azure to on-prem backup",
      "Compressed transfers save money",
    ],
  },
  {
    icon: "/images/icon-lock.svg",
    title: "Disaster Recovery",
    description: "Affordable cross-region DR",
    details: [
      "Real-time replicas in another cloud",
      "Compression cuts DR costs dramatically",
      "Failover-ready at all times",
    ],
  },
  {
    icon: "/images/icon-migration.svg",
    title: "Data Migration",
    description: "Move to a new cloud without downtime",
    details: [
      "Sync continuously during migration",
      "Switch over when ready",
      "No big-bang cutover risk",
    ],
  },
  {
    icon: "/images/icon-analytics.svg",
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
    <section className="section-padding bg-dark-bg">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-h2">Where teams use Savegress</h2>
        </motion.div>

        <div className="flex flex-wrap justify-center gap-8">
          {useCases.map((useCase, index) => (
            <motion.div
              key={useCase.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
              className="w-full md:w-auto"
            >
              <div className="usecase-card p-8">
                {/* Icon */}
                <img
                  src={useCase.icon}
                  alt=""
                  className="absolute top-6 left-6 md:top-8 md:left-8 w-[50px] h-[50px] md:w-[63px] md:h-[63px] icon-hover"
                />

                <h3 className="text-h4 max-w-full mt-[70px] md:mt-[85px]">{useCase.title}</h3>
                <p className="text-content-1 text-cyan max-w-full mt-3">{useCase.description}</p>

                <ul className="text-content-1 text-grey max-w-full mt-6">
                  {useCase.details.map((detail) => (
                    <li key={detail} className="flex items-center gap-3">
                      <span className="dot-cyan flex-shrink-0" />
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
