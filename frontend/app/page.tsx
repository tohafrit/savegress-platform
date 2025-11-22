import { Hero } from "@/components/sections/hero"
import { Problem } from "@/components/sections/problem"
import { Solution } from "@/components/sections/solution"
import { HowItWorks } from "@/components/sections/how-it-works"
import { Specs } from "@/components/sections/specs"
import { Calculator } from "@/components/sections/calculator"
import { UseCases } from "@/components/sections/use-cases"
import { Trust } from "@/components/sections/trust"
import { CTA } from "@/components/sections/cta"
import { Footer } from "@/components/sections/footer"

export default function Home() {
  return (
    <main className="min-h-screen">
      <Hero />
      <Problem />
      <Solution />
      <HowItWorks />
      <Specs />
      <Calculator />
      <UseCases />
      <Trust />
      <CTA />
      <Footer />
    </main>
  )
}
