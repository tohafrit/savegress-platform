"use client"

import { motion } from "framer-motion"
import Image from "next/image"
import { Particles } from "@/components/ui/particles"

export function Hero() {
  const scrollToForm = () => {
    document.getElementById('early-access-form')?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <section className="relative overflow-hidden bg-dark-bg">
      {/* Floating particles */}
      <Particles count={40} />

      {/* Background wave */}
      <img
        src="/images/bg-hero.png"
        alt=""
        className="absolute left-0 top-0 w-full h-[814px] object-cover z-[9]"
      />

      {/* Diagonal overlay at bottom */}
      <div
        className="absolute bottom-0 left-0 right-0 z-[9] h-[150px] bg-dark-bg"
        style={{ clipPath: 'polygon(0 100%, 100% 0, 100% 100%)' }}
      />

      {/* Header/Logo */}
      <header className="relative z-10 pt-[52px] pb-[50px]">
        <div className="container-custom">
          <Image
            src="/images/logo.svg"
            alt="Savegress"
            width={208}
            height={58}
            className="h-[58px] w-auto"
          />
        </div>
      </header>

      {/* Hero content */}
      <div className="container-custom relative z-10 pb-24">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          {/* Main heading */}
          <h1 className="text-h1 uppercase mb-8 max-w-[1219px]">
            <span className="text-white">REPLICATE DATA ACROSS CLOUDS. </span>
            <span className="text-gradient-cyan">PAY LESS FOR EGRESS.</span>
          </h1>

          {/* Subtitle */}
          <p className="text-subtitle-1 text-grey w-[800px] mb-10">
            Stream database changes between AWS, GCP, and Azure.<br />
            Compress up to 200x to cut your data transfer costs.
          </p>

          {/* Buttons */}
          <div className="flex flex-col sm:flex-row gap-4">
            <button onClick={scrollToForm} className="btn-primary w-[352px]">
              Request Early Access  →
            </button>
            <button
              onClick={() => document.getElementById('how-it-works')?.scrollIntoView({ behavior: 'smooth' })}
              className="btn-secondary w-[240px] h-[68px]"
            >
              See How It Works
            </button>
          </div>
        </motion.div>
      </div>

      {/* Savegress Schema */}
      <div className="container-custom relative z-10 pb-20">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="schema-box">
            <div className="flex flex-col md:flex-row items-center justify-between gap-6 md:gap-4">
              {/* Left side - Source databases */}
              <div className="flex flex-col gap-4">
                <div className="schema-badge">AWS PostgreSQL</div>
                <div className="schema-badge">Azure MySQL</div>
              </div>

              {/* Arrows left */}
              <div className="hidden md:block flex-shrink-0">
                <svg width="146" height="97" viewBox="0 0 146 97" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path opacity="0.6" d="M0.396484 0.304626C31.8965 41.3046 58.3965 50.3046 143.896 47.8046" stroke="#01C8EF" strokeDasharray="6 6"/>
                  <path opacity="0.6" d="M0.396484 96.1891C31.8965 55.1891 58.3965 46.1891 143.896 48.6891" stroke="#01C8EF" strokeDasharray="6 6"/>
                  <path d="M140.896 44.8046L144.896 48.3046L140.896 51.3046" stroke="#01C8EF"/>
                </svg>
              </div>

              {/* Center - Savegress */}
              <div className="relative flex flex-col items-center flex-shrink-0">
                <div className="schema-center">Savegress</div>
                <div className="absolute top-full mt-3 w-[176px] text-center text-content-1 text-cyan">
                  200x smaller
                </div>
              </div>

              {/* Arrows right */}
              <div className="hidden md:block flex-shrink-0">
                <svg width="149" height="86" viewBox="0 0 149 86" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path opacity="0.6" d="M143.515 5.00012C112.015 46.0001 85.5146 45.0001 0.0146484 42.5001" stroke="#01C8EF" strokeDasharray="6 6"/>
                  <path opacity="0.6" d="M143.515 80.8847C112.015 39.8847 85.5146 40.8847 0.0146484 43.3847" stroke="#01C8EF" strokeDasharray="6 6"/>
                  <path d="M145.611 78.0002L145.964 83.3035L141.015 82.5964" stroke="#01C8EF"/>
                  <path d="M141.015 2.82837L146.318 2.47482L145.611 7.42456" stroke="#01C8EF"/>
                </svg>
              </div>

              {/* Right side - Destinations */}
              <div className="flex flex-col gap-4">
                <div className="schema-badge">GCP</div>
                <div className="schema-badge">On-prem</div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
