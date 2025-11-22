"use client"

import { EarlyAccessForm } from "@/components/forms/early-access-form"
import { motion } from "framer-motion"

export function CTA() {
  return (
    <section id="early-access-form" className="section-padding bg-gradient-to-b from-neutral-light-gray to-white">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-heading-lg text-primary mb-4">
            Ready to cut your egress costs?
          </h2>
          <p className="text-body-lg text-neutral-dark-gray max-w-2xl mx-auto">
            Join the early access program. See compression
            in action on your actual data.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="max-w-2xl mx-auto"
        >
          <EarlyAccessForm />
        </motion.div>
      </div>
    </section>
  )
}
