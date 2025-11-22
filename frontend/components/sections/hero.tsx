"use client"

import { Button } from "@/components/ui/button"
import { motion } from "framer-motion"

export function Hero() {
  const scrollToForm = () => {
    document.getElementById('early-access-form')?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <section className="section-padding bg-gradient-to-b from-neutral-light-gray to-white">
      <div className="container-custom">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="max-w-4xl mx-auto text-center"
        >
          <h1 className="text-display-md md:text-display-lg text-primary mb-6">
            Replicate data across clouds.
            <br />
            <span className="text-accent-orange">Pay less for egress.</span>
          </h1>

          <p className="text-body-lg text-neutral-dark-gray mb-8 max-w-2xl mx-auto">
            Stream database changes between AWS, GCP, and Azure.
            Compress up to 200x to cut your data transfer costs.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button
              size="lg"
              onClick={scrollToForm}
              className="bg-primary hover:bg-primary-dark"
            >
              Request Early Access
            </Button>
            <Button
              size="lg"
              variant="outline"
              onClick={() => document.getElementById('how-it-works')?.scrollIntoView({ behavior: 'smooth' })}
            >
              See How It Works
            </Button>
          </div>
        </motion.div>

        {/* Multi-cloud diagram */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          className="mt-16 max-w-5xl mx-auto"
        >
          <div className="bg-white rounded-xl border border-gray-200 p-8 shadow-lg">
            <div className="flex flex-col md:flex-row items-center justify-between gap-8">
              {/* Source */}
              <div className="flex flex-col gap-4">
                <CloudBadge name="AWS PostgreSQL" />
                <CloudBadge name="Azure MySQL" />
              </div>

              {/* Savegress */}
              <div className="flex flex-col items-center">
                <div className="bg-primary text-white px-6 py-3 rounded-lg font-semibold">
                  Savegress
                </div>
                <div className="mt-2 text-sm text-accent-orange font-mono">
                  200x smaller
                </div>
              </div>

              {/* Arrow */}
              <div className="hidden md:block">
                <span className="text-2xl text-gray-400">-&gt;</span>
              </div>

              {/* Destination */}
              <div className="flex flex-col gap-4">
                <CloudBadge name="GCP" />
                <CloudBadge name="On-prem" />
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

function CloudBadge({ name }: { name: string }) {
  return (
    <div className="flex items-center gap-2 bg-neutral-light-gray px-4 py-2 rounded-lg border border-gray-200">
      <span className="font-medium text-neutral-dark-gray">{name}</span>
    </div>
  )
}
