"use client"

import { motion } from "framer-motion"

export function Trust() {
  return (
    <section className="section-padding bg-dark-bg-secondary relative overflow-hidden">
      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center flex flex-col items-center"
        >
          <h2 className="text-h2 w-[1216px] mb-6">Early Access Program</h2>

          <div>
            <p className="text-content-1 text-cyan w-[1215px] h-[48px] text-center mb-8">
              Savegress is currently in early access. We&apos;re working with
              design partners to refine the product for general availability.
            </p>

            {/* GO and RUST buttons */}
            <div className="flex justify-center gap-4 mb-8">
              <span className="trust-btn-primary">GO</span>
              <span className="trust-btn-secondary">RUST</span>
            </div>

            {/* Feature badges */}
            <div className="flex flex-wrap justify-center gap-4">
              <span className="trust-badge">Production-tested</span>
              <span className="trust-badge">Battle-hardened</span>
              <span className="trust-badge">Enterprise-ready</span>
            </div>

            {/* Quote */}
            <p className="text-content-1 text-grey w-[1216px] h-[28px] text-center mt-8">
              Built by infrastructure engineers who understand the pain of scaling data pipelines
            </p>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
