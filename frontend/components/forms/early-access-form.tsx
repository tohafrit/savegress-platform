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
import { Card } from "@/components/ui/card"

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
      <Card className="p-8 text-center">
        <div className="text-4xl mb-4">+</div>
        <h3 className="text-heading-sm text-primary mb-2">Thank you!</h3>
        <p className="text-body-md text-neutral-dark-gray">
          We&apos;ll contact you soon with early access details.
        </p>
      </Card>
    )
  }

  return (
    <Card className="p-8">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Email */}
        <div>
          <Label htmlFor="email">Work email *</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@company.com"
            {...register("email")}
            className="mt-1"
          />
          {errors.email && (
            <p className="text-sm text-red-500 mt-1">{errors.email.message}</p>
          )}
        </div>

        {/* Company */}
        <div>
          <Label htmlFor="company">Company name *</Label>
          <Input
            id="company"
            type="text"
            placeholder="Acme Inc."
            {...register("company")}
            className="mt-1"
          />
          {errors.company && (
            <p className="text-sm text-red-500 mt-1">{errors.company.message}</p>
          )}
        </div>

        {/* Current solution */}
        <div>
          <Label htmlFor="currentSolution">Current CDC solution</Label>
          <Select onValueChange={(value) => setValue("currentSolution", value as FormData["currentSolution"])}>
            <SelectTrigger className="mt-1">
              <SelectValue placeholder="Select..." />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="kafka">Kafka/Debezium</SelectItem>
              <SelectItem value="dms">AWS DMS</SelectItem>
              <SelectItem value="fivetran">Fivetran</SelectItem>
              <SelectItem value="none">None</SelectItem>
              <SelectItem value="other">Other</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Data volume */}
        <div>
          <Label htmlFor="dataVolume">Daily data volume</Label>
          <Select onValueChange={(value) => setValue("dataVolume", value as FormData["dataVolume"])}>
            <SelectTrigger className="mt-1">
              <SelectValue placeholder="Select..." />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="<10GB">&lt;10GB</SelectItem>
              <SelectItem value="10-100GB">10-100GB</SelectItem>
              <SelectItem value="100GB-1TB">100GB-1TB</SelectItem>
              <SelectItem value=">1TB">&gt;1TB</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Message */}
        <div>
          <Label htmlFor="message">Message (optional)</Label>
          <Textarea
            id="message"
            placeholder="Tell us about your use case..."
            {...register("message")}
            className="mt-1"
            rows={4}
          />
        </div>

        {/* Turnstile CAPTCHA */}
        <div className="flex justify-center">
          <Turnstile
            ref={turnstileRef}
            siteKey={TURNSTILE_SITE_KEY}
            onSuccess={setTurnstileToken}
            onError={() => setTurnstileToken(null)}
            onExpire={() => setTurnstileToken(null)}
          />
        </div>

        {/* Submit */}
        <Button
          type="submit"
          className="w-full bg-primary hover:bg-primary-dark"
          disabled={isSubmitting}
        >
          {isSubmitting ? 'Submitting...' : 'Request Early Access'}
        </Button>

        <p className="text-sm text-neutral-dark-gray text-center">
          Or{' '}
          <a href="#" className="text-primary hover:text-primary-dark underline">
            schedule a call
          </a>
        </p>
      </form>
    </Card>
  )
}
