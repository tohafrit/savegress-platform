"use client"

import { useState, useRef } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Turnstile, type TurnstileInstance } from "@marsidev/react-turnstile"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { CheckCircle } from "lucide-react"

const formSchema = z.object({
  email: z.string().email("Invalid email address"),
  company: z.string().min(2, "Company name is required"),
  currentSolution: z.enum(["kafka", "dms", "fivetran", "none", "other"]).optional(),
  dataVolume: z.enum(["<10GB", "10-100GB", "100GB-1TB", ">1TB"]).optional(),
  message: z.string().optional(),
})

type FormData = z.infer<typeof formSchema>

const TURNSTILE_SITE_KEY = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "0x4AAAAAACCXwGR68idJikxS"

export function EarlyAccessForm() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSuccess, setIsSuccess] = useState(false)
  const [turnstileToken, setTurnstileToken] = useState<string | null>(null)
  const turnstileRef = useRef<TurnstileInstance>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    reset,
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
  })

  const onSubmit = async (data: FormData) => {
    if (!turnstileToken) {
      alert('Please complete the verification')
      return
    }

    setIsSubmitting(true)

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/early-access`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...data, turnstileToken }),
      })

      if (!response.ok) {
        throw new Error('Failed to submit')
      }

      setIsSuccess(true)
      reset()
      setTurnstileToken(null)
      turnstileRef.current?.reset()

      setTimeout(() => {
        setIsSuccess(false)
      }, 5000)
    } catch (error) {
      console.error(error)
      alert('Something went wrong. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isSuccess) {
    return (
      <div className="text-center py-8">
        <CheckCircle className="w-16 h-16 text-accent-cyan mx-auto mb-4" />
        <h3 className="text-h4 mb-2">Thank you!</h3>
        <p className="text-content-1 text-grey">
          We&apos;ll contact you soon with early access details.
        </p>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-6">
        {/* Email */}
        <div>
          <Label htmlFor="email" className="text-content-1 text-white w-full h-[24px]">
            Work email <span className="text-cyan">*</span>
          </Label>
          <Input
            id="email"
            type="email"
            placeholder="you@company.com"
            {...register("email")}
            className="input-field mt-2 text-white placeholder:text-text-muted h-[44px]"
          />
          {errors.email && (
            <p className="text-sm text-accent-orange mt-1">{errors.email.message}</p>
          )}
        </div>

        {/* Company */}
        <div>
          <Label htmlFor="company" className="text-content-1 text-white w-full h-[24px]">
            Company name <span className="text-cyan">*</span>
          </Label>
          <Input
            id="company"
            type="text"
            placeholder="Acme Inc."
            {...register("company")}
            className="input-field mt-2 text-white placeholder:text-text-muted h-[44px]"
          />
          {errors.company && (
            <p className="text-sm text-accent-orange mt-1">{errors.company.message}</p>
          )}
        </div>

        {/* Current solution */}
        <div>
          <Label htmlFor="currentSolution" className="text-content-1 text-white w-full h-[24px]">
            Current CDC solution
          </Label>
          <Select onValueChange={(value) => setValue("currentSolution", value as FormData["currentSolution"])}>
            <SelectTrigger className="input-field mt-2 text-white h-[44px] [&>svg:last-child]:hidden">
              <SelectValue placeholder="Select..." />
              <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg" className="ml-auto">
                <path d="M0.5 0.5L5.5 5.5L10.5 0.5" stroke="#02ACD0" strokeLinecap="round"/>
              </svg>
            </SelectTrigger>
            <SelectContent className="card-dark">
              <SelectItem value="kafka" className="text-white hover:bg-white/5">Kafka/Debezium</SelectItem>
              <SelectItem value="dms" className="text-white hover:bg-white/5">AWS DMS</SelectItem>
              <SelectItem value="fivetran" className="text-white hover:bg-white/5">Fivetran</SelectItem>
              <SelectItem value="none" className="text-white hover:bg-white/5">None</SelectItem>
              <SelectItem value="other" className="text-white hover:bg-white/5">Other</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Data volume */}
        <div>
          <Label htmlFor="dataVolume" className="text-content-1 text-white w-full h-[24px]">
            Daily data volume
          </Label>
          <Select onValueChange={(value) => setValue("dataVolume", value as FormData["dataVolume"])}>
            <SelectTrigger className="input-field mt-2 text-white h-[44px] [&>svg:last-child]:hidden">
              <SelectValue placeholder="Select..." />
              <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg" className="ml-auto">
                <path d="M0.5 0.5L5.5 5.5L10.5 0.5" stroke="#02ACD0" strokeLinecap="round"/>
              </svg>
            </SelectTrigger>
            <SelectContent className="card-dark">
              <SelectItem value="<10GB" className="text-white hover:bg-white/5">&lt;10GB</SelectItem>
              <SelectItem value="10-100GB" className="text-white hover:bg-white/5">10-100GB</SelectItem>
              <SelectItem value="100GB-1TB" className="text-white hover:bg-white/5">100GB-1TB</SelectItem>
              <SelectItem value=">1TB" className="text-white hover:bg-white/5">&gt;1TB</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Message */}
        <div>
          <Label htmlFor="message" className="text-content-1 text-white w-full h-[24px]">
            Message (optional)
          </Label>
          <Textarea
            id="message"
            placeholder="Tell us about your use case..."
            {...register("message")}
            className="input-field mt-2 text-white placeholder:text-text-muted resize-none h-[106px]"
          />
        </div>
      </div>

      {/* Turnstile CAPTCHA */}
      <div className="flex justify-center py-6">
        <Turnstile
          ref={turnstileRef}
          siteKey={TURNSTILE_SITE_KEY}
          onSuccess={setTurnstileToken}
          onError={() => setTurnstileToken(null)}
          onExpire={() => setTurnstileToken(null)}
          options={{ theme: "dark" }}
        />
      </div>

      {/* Submit */}
      <div className="flex justify-center">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="btn-primary w-full sm:w-[312px] h-[60px] sm:h-[68px]"
        >
          {isSubmitting ? 'Submitting...' : 'Request Early Access  →'}
        </Button>
      </div>
    </form>
  )
}
