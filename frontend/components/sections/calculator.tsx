"use client"

import { useState, useEffect } from "react"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { motion } from "framer-motion"
import { useCountUp } from "@/hooks/use-count-up"

// Default pricing (used while loading and as fallback)
const DEFAULT_PRICING = {
  AWS: {
    "us-east": 0.09,
    "eu-west": 0.09,
    "ap-southeast": 0.114,
  },
  GCP: {
    "us-east": 0.12,
    "eu-west": 0.12,
    "ap-southeast": 0.15,
  },
  Azure: {
    "us-east": 0.087,
    "eu-west": 0.087,
    "ap-southeast": 0.12,
  },
}

interface PricingData {
  AWS: Record<string, number>
  GCP: Record<string, number>
  Azure: Record<string, number>
  lastUpdated?: string
  source?: 'live' | 'fallback'
}

const REGIONS = {
  "us-east": "North America",
  "eu-west": "Europe",
  "ap-southeast": "Asia Pacific",
}

const COMPRESSION_RATIOS = {
  "time-series": { ratio: 50, label: "Time-series heavy (timestamps, IDs)" },
  "mixed": { ratio: 12, label: "Mixed CDC (typical workload)" },
  "json-heavy": { ratio: 5, label: "JSON/text heavy" },
}

export function Calculator() {
  const [dailyGB, setDailyGB] = useState(100)
  const [sourceCloud, setSourceCloud] = useState<keyof typeof DEFAULT_PRICING>("AWS")
  const [region, setRegion] = useState<keyof typeof REGIONS>("us-east")
  const [dataPattern, setDataPattern] = useState<keyof typeof COMPRESSION_RATIOS>("mixed")
  const [pricing, setPricing] = useState<PricingData>(DEFAULT_PRICING)
  const [isLivePricing, setIsLivePricing] = useState(false)

  // Fetch live pricing on mount
  useEffect(() => {
    async function fetchPricing() {
      try {
        const response = await fetch('/api/pricing')
        if (response.ok) {
          const data: PricingData = await response.json()
          setPricing(data)
          setIsLivePricing(data.source === 'live')
        }
      } catch (error) {
        console.error('Failed to fetch pricing:', error)
      }
    }
    fetchPricing()
  }, [])

  const monthlyGB = dailyGB * 30
  const pricePerGB = pricing[sourceCloud]?.[region] ?? DEFAULT_PRICING[sourceCloud][region]
  const compressionRatio = COMPRESSION_RATIOS[dataPattern].ratio

  const currentMonthlyCost = monthlyGB * pricePerGB
  const compressedGB = monthlyGB / compressionRatio
  const newMonthlyCost = compressedGB * pricePerGB
  const monthlySavings = currentMonthlyCost - newMonthlyCost
  const yearlySavings = monthlySavings * 12

  // Animated values
  const animatedCurrentCost = useCountUp(currentMonthlyCost, 300)
  const animatedNewCost = useCountUp(newMonthlyCost, 300)
  const animatedSavings = useCountUp(monthlySavings, 300)
  const animatedYearly = useCountUp(yearlySavings, 300)

  return (
    <section className="pt-16 md:pt-[140px] pb-section md:pb-[140px] bg-dark-bg-secondary relative overflow-hidden">
      {/* Background image with pulse effect */}
      <motion.img
        src="/images/bg-calculator.png"
        alt=""
        className="absolute inset-0 w-full h-full pointer-events-none"
        animate={{
          opacity: [0.85, 1, 0.85],
        }}
        transition={{
          duration: 6,
          ease: "easeInOut",
          repeat: Infinity,
        }}
      />

      <div className="container-custom relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-8 md:mb-12"
        >
          <h2 className="text-h2">Calculate your egress savings</h2>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.1 }}
        >
          <div className="calculator-card w-full max-w-[624px] mx-auto md:mx-0 p-6 md:p-[40px_72px]">
            <div className="space-y-6">
              {/* Daily volume slider */}
              <div>
                <Label className="text-content-1 text-grey mb-3 block w-full max-w-[478px]">
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
                    className="slider-scale w-full max-w-[480px] appearance-none cursor-pointer
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
                <div className="flex justify-between mt-2 w-full max-w-[480px] text-content-1 text-cyan">
                  <span>10 GB</span>
                  <span>1 TB</span>
                </div>
              </div>

              {/* Source cloud and region */}
              <div className="flex flex-col sm:flex-row gap-4">
                <div className="flex-1">
                  <Label className="text-content-1 text-grey mb-3 block">Source cloud</Label>
                  <Select value={sourceCloud} onValueChange={(v) => setSourceCloud(v as keyof typeof DEFAULT_PRICING)}>
                    <SelectTrigger className="input-field w-full h-[44px] text-white [&>svg:last-child]:hidden">
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
                </div>
                <div className="flex-1">
                  <Label className="text-content-1 text-grey mb-3 block">Region</Label>
                  <Select value={region} onValueChange={(v) => setRegion(v as keyof typeof REGIONS)}>
                    <SelectTrigger className="input-field w-full h-[44px] text-white [&>svg:last-child]:hidden">
                      <SelectValue />
                      <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg" className="ml-auto">
                        <path d="M0.5 0.5L5.5 5.5L10.5 0.5" stroke="#02ACD0" strokeLinecap="round"/>
                      </svg>
                    </SelectTrigger>
                    <SelectContent className="bg-dark-bg-card border-white/10">
                      <SelectItem value="us-east" className="text-white hover:bg-white/5">North America</SelectItem>
                      <SelectItem value="eu-west" className="text-white hover:bg-white/5">Europe</SelectItem>
                      <SelectItem value="ap-southeast" className="text-white hover:bg-white/5">Asia Pacific</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Data pattern */}
              <div>
                <Label className="text-content-1 text-grey mb-3 block w-full max-w-[478px]">Data pattern</Label>
                <Select value={dataPattern} onValueChange={(v) => setDataPattern(v as keyof typeof COMPRESSION_RATIOS)}>
                  <SelectTrigger className="input-field w-full max-w-[480px] h-[44px] text-white [&>svg:last-child]:hidden">
                    <SelectValue />
                    <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg" className="ml-auto">
                      <path d="M0.5 0.5L5.5 5.5L10.5 0.5" stroke="#02ACD0" strokeLinecap="round"/>
                    </svg>
                  </SelectTrigger>
                  <SelectContent className="bg-dark-bg-card border-white/10">
                    <SelectItem value="time-series" className="text-white hover:bg-white/5">Time-series heavy (50x)</SelectItem>
                    <SelectItem value="mixed" className="text-white hover:bg-white/5">Mixed CDC typical (12x)</SelectItem>
                    <SelectItem value="json-heavy" className="text-white hover:bg-white/5">JSON/text heavy (5x)</SelectItem>
                  </SelectContent>
                </Select>

                {/* Divider */}
                <div className="w-full max-w-[480px] h-[1px] mt-10 md:mt-14 mb-6 md:mb-8 border-t border-dashed border-[#02ACD0]/50" />
              </div>

              {/* Results */}
              <div>
                <div className="flex flex-col sm:flex-row justify-between items-start gap-2 sm:gap-0">
                  <div className="text-content-2 text-grey">
                    <div>Current monthly egress cost:</div>
                    <div>With Savegress (at {compressionRatio}x):</div>
                  </div>
                  <div className="text-left sm:text-right">
                    <div className="text-h4 text-grey">${animatedCurrentCost.toFixed(2)}</div>
                    <div className="text-h5 text-cyan">${animatedNewCost.toFixed(2)}</div>
                  </div>
                </div>

                {/* Divider */}
                <div className="w-full max-w-[480px] h-[1px] mt-6 md:mt-8 mb-6 md:mb-8 border-t border-dashed border-[#02ACD0]/50" />

                <div>
                  <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-3 gap-2 sm:gap-0">
                    <span className="text-h5">You save:</span>
                    <div className="text-h4 text-cyan">${animatedSavings.toFixed(2)}/month</div>
                  </div>
                  <div className="text-h5 sm:text-right">${animatedYearly.toFixed(0)}/year</div>
                </div>
              </div>

              <p className="text-mini-3 text-grey mt-6 mb-10 w-full max-w-[480px]">
                * Egress pricing: {sourceCloud} {REGIONS[region]} at ${pricePerGB}/GB
                {isLivePricing ? (
                  <span className="text-cyan"> (live from API)</span>
                ) : null}
                . Prices from official {sourceCloud} pricing. May vary by volume tier and commitment.
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
