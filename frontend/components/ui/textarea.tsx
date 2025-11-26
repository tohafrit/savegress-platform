import * as React from "react"

import { cn } from "@/lib/utils"

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={cn(
          "flex min-h-[80px] w-full rounded-[15px] border border-[rgba(0,180,216,0.4)] bg-gradient-to-r from-[rgba(3,14,32,0.8)] to-[#1E3A5F] px-5 py-3 text-base text-white ring-offset-background placeholder:text-[#b2bbc9] placeholder:opacity-60 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#00b4d8] disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Textarea.displayName = "Textarea"

export { Textarea }
