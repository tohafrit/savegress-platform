"use client"

import { useState } from "react"
import { Card } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { motion } from "framer-motion"

const EGRESS_PRICING = {
  AWS: 0.09,
  GCP: 0.08,
  Azure: 0.087,
}

const COMPRESSION_RATIO = 20 // Conservative estimate

export function Calculator() {
  const [dailyGB, setDailyGB] = useState(100)
  const [sourceCloud, setSourceCloud] = useState<keyof typeof EGRESS_PRICING>("AWS")

  const monthlyGB = dailyGB * 30
  const pricePerGB = EGRESS_PRICING[sourceCloud]

  const currentMonthlyCost = monthlyGB * pricePerGB
  const compressedGB = monthlyGB / COMPRESSION_RATIO
  const newMonthlyCost = compressedGB * pricePerGB
  const monthlySavings = currentMonthlyCost - newMonthlyCost
  const yearlySavings = monthlySavings * 12

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
            Calculate your egress savings
          </h2>
        </motion.div>

        <Card className="max-w-2xl mx-auto p-8">
          <div className="space-y-6">
            {/* Daily volume slider */}
            <div>
              <Label className="text-sm font-medium text-neutral-dark-gray mb-2 block">
                Daily data volume: {dailyGB} GB
              </Label>
              <input
                type="range"
                min="10"
                max="1000"
                step="10"
                value={dailyGB}
                onChange={(e) => setDailyGB(Number(e.target.value))}
                className="w-full h-2 bg-neutral-light-gray rounded-lg appearance-none cursor-pointer accent-primary"
              />
              <div className="flex justify-between text-xs text-neutral-dark-gray mt-1">
                <span>10 GB</span>
                <span>1 TB</span>
              </div>
            </div>

            {/* Source cloud */}
            <div>
              <Label className="text-sm font-medium text-neutral-dark-gray mb-2 block">
                Source cloud
              </Label>
              <Select value={sourceCloud} onValueChange={(v) => setSourceCloud(v as keyof typeof EGRESS_PRICING)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="AWS">AWS</SelectItem>
                  <SelectItem value="GCP">GCP</SelectItem>
                  <SelectItem value="Azure">Azure</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Results */}
            <div className="mt-8 pt-6 border-t border-gray-200">
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-neutral-dark-gray">Current monthly egress cost:</span>
                  <span className="text-lg font-bold text-primary font-mono">
                    ${currentMonthlyCost.toFixed(2)}
                  </span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-sm text-neutral-dark-gray">With Savegress (at {COMPRESSION_RATIO}x):</span>
                  <span className="text-lg font-bold text-accent-orange font-mono">
                    ${newMonthlyCost.toFixed(2)}
                  </span>
                </div>

                <div className="pt-4 border-t border-gray-200">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-semibold text-neutral-dark-gray">You save:</span>
                    <span className="text-2xl font-bold text-accent-orange font-mono">
                      ${monthlySavings.toFixed(2)}/month
                    </span>
                  </div>
                  <div className="text-right">
                    <span className="text-lg font-bold text-primary font-mono">
                      ${yearlySavings.toFixed(0)}/year
                    </span>
                  </div>
                </div>
              </div>

              <p className="text-xs text-neutral-dark-gray mt-4">
                * Based on {sourceCloud} egress pricing (${pricePerGB}/GB).
                Compression typically 10-50x for CDC data.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </section>
  )
}
