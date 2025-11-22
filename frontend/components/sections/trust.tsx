"use client"

import { motion } from "framer-motion"

export function Trust() {
  return (
    <section className="section-padding bg-white">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center max-w-3xl mx-auto"
        >
          <h2 className="text-heading-lg text-primary mb-6">
            Early Access Program
          </h2>

          <p className="text-body-lg text-neutral-dark-gray mb-8">
            Savegress is currently in early access. We&apos;re working with
            design partners to refine the product for general availability.
          </p>

          <div className="flex items-center justify-center gap-8 mb-8">
            <div className="flex items-center gap-2">
              <div className="w-12 h-12 bg-neutral-light-gray rounded-lg flex items-center justify-center">
                <span className="font-mono font-bold text-primary">Go</span>
              </div>
              <div className="w-12 h-12 bg-neutral-light-gray rounded-lg flex items-center justify-center">
                <span className="font-mono font-bold text-accent-orange">Rs</span>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap justify-center gap-4 text-sm text-neutral-dark-gray">
            <span className="px-3 py-1 bg-neutral-light-gray rounded-full">Production-tested</span>
            <span className="px-3 py-1 bg-neutral-light-gray rounded-full">Battle-hardened</span>
            <span className="px-3 py-1 bg-neutral-light-gray rounded-full">Enterprise-ready</span>
          </div>

          <p className="mt-6 text-sm text-neutral-dark-gray italic">
            &quot;Built by infrastructure engineers who understand
            the pain of scaling data pipelines&quot;
          </p>
        </motion.div>
      </div>
    </section>
  )
}
