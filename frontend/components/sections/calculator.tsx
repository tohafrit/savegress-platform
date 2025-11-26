"use client"

import { useState } from "react"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { motion } from "framer-motion"
import { useCountUp } from "@/hooks/use-count-up"

const EGRESS_PRICING = {
  AWS: 0.09,
  GCP: 0.08,
  Azure: 0.087,
}

const COMPRESSION_RATIO = 20

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

  // Animated values
  const animatedCurrentCost = useCountUp(currentMonthlyCost, 300)
  const animatedNewCost = useCountUp(newMonthlyCost, 300)
  const animatedSavings = useCountUp(monthlySavings, 300)
  const animatedYearly = useCountUp(yearlySavings, 300)

  return (
    <section className="pt-[140px] pb-section bg-dark-bg-secondary relative overflow-hidden">
      {/* Background image */}
      <img
        src="/images/bg-calculator.png"
        alt=""
        className="absolute inset-0 w-full h-full pointer-events-none"
      />

      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-h2">Calculate your egress savings</h2>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.1 }}
        >
          <div className="calculator-card w-[624px] h-[670px] p-[40px_72px]">
            <div className="space-y-6">
              {/* Daily volume slider */}
              <div>
                <Label className="text-content-1 text-grey mb-3 block w-[478px]">
                  Daily data volume: <span className="text-cyan">{dailyGB} GB</span>
                </Label>
                <div className="relative">
                  <input
                    type="range"
                    min="10"
                    max="1000"
                    step="10"
                    value={dailyGB}
                    onChange={(e) => setDailyGB(Number(e.target.value))}
                    className="slider-scale w-[480px] appearance-none cursor-pointer
                      [&::-webkit-slider-thumb]:appearance-none
                      [&::-webkit-slider-thumb]:w-[15px]
                      [&::-webkit-slider-thumb]:h-[15px]
                      [&::-webkit-slider-thumb]:rounded-full
                      [&::-webkit-slider-thumb]:bg-dark-bg
                      [&::-webkit-slider-thumb]:border-2
                      [&::-webkit-slider-thumb]:border-solid
                      [&::-webkit-slider-thumb]:border-accent-cyan
                      [&::-webkit-slider-thumb]:cursor-pointer"
                  />
                </div>
                <div className="flex justify-between mt-2 w-[480px] text-content-1 text-cyan">
                  <span>10 GB</span>
                  <span>1 TB</span>
                </div>
              </div>

              {/* Source cloud */}
              <div>
                <Label className="text-content-1 text-grey mb-3 block w-[478px]">Source cloud</Label>
                <Select value={sourceCloud} onValueChange={(v) => setSourceCloud(v as keyof typeof EGRESS_PRICING)}>
                  <SelectTrigger className="input-field w-[480px] h-[44px] text-white [&>svg:last-child]:hidden">
                    <SelectValue />
                    <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg" className="ml-auto">
                      <path d="M0.5 0.5L5.5 5.5L10.5 0.5" stroke="#02ACD0" strokeLinecap="round"/>
                    </svg>
                  </SelectTrigger>
                  <SelectContent className="bg-dark-bg-card border-white/10">
                    <SelectItem value="AWS" className="text-white hover:bg-white/5">AWS</SelectItem>
                    <SelectItem value="GCP" className="text-white hover:bg-white/5">GCP</SelectItem>
                    <SelectItem value="Azure" className="text-white hover:bg-white/5">Azure</SelectItem>
                  </SelectContent>
                </Select>

                {/* Divider */}
                <svg width="480" height="1" viewBox="0 0 480 1" fill="none" xmlns="http://www.w3.org/2000/svg" className="mt-14 mb-8">
                  <path d="M0.25 0.25H479.75" stroke="#02ACD0" strokeWidth="0.5" strokeLinecap="round" strokeDasharray="4 4"/>
                </svg>
              </div>

              {/* Results */}
              <div>
                <div className="flex justify-between items-start">
                  <div className="text-content-2 text-grey">
                    <div>Current monthly egress cost:</div>
                    <div>With Savegress (at {COMPRESSION_RATIO}x):</div>
                  </div>
                  <div className="text-right">
                    <div className="text-h4 text-grey">${animatedCurrentCost.toFixed(2)}</div>
                    <div className="text-h5 text-cyan">${animatedNewCost.toFixed(2)}</div>
                  </div>
                </div>

                {/* Divider */}
                <svg width="480" height="1" viewBox="0 0 480 1" fill="none" xmlns="http://www.w3.org/2000/svg" className="mt-8 mb-8">
                  <path d="M0.25 0.25H479.75" stroke="#02ACD0" strokeWidth="0.5" strokeLinecap="round" strokeDasharray="4 4"/>
                </svg>

                <div>
                  <div className="flex justify-between items-center mb-3">
                    <span className="text-h5 w-[103px] h-[28px] flex flex-col justify-center">You save:</span>
                    <div className="text-h4 text-cyan text-right">${animatedSavings.toFixed(2)}/month</div>
                  </div>
                  <div className="text-h5 text-right">${animatedYearly.toFixed(0)}/year</div>
                </div>
              </div>

              <p className="text-mini-3 text-grey mt-6 w-[480px]">
                * Based on {sourceCloud} egress pricing (${pricePerGB}/GB). Compression typically 10-50x for CDC data.
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
