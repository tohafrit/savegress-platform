"use client"

import { EarlyAccessForm } from "@/components/forms/early-access-form"
import { motion } from "framer-motion"

export function CTA() {
  return (
    <section id="early-access-form" className="pt-16 md:pt-[140px] pb-section bg-dark-bg relative overflow-hidden">
      {/* Background image */}
      <img
        src="/images/bg-form.png"
        alt=""
        className="absolute inset-0 w-full h-full pointer-events-none"
      />

      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-8 md:mb-12"
        >
          <h2 className="text-h2 w-full max-w-[1216px] mx-auto mb-4">Ready to cut your egress costs?</h2>
          <p className="text-content-1 text-grey w-full max-w-[1216px] mx-auto text-center">
            Join the early access program. See compression in action on your actual data.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="flex justify-center lg:justify-end"
        >
          <div className="cta-form-card">
            <EarlyAccessForm />

            {/* Schedule call link */}
            <div className="flex justify-center mt-6">
              <a href="#" className="text-mini-2 text-cyan text-center">
                Or schedule a call
              </a>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
