import * as React from "react"

import { cn } from "@/lib/utils"

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex w-full rounded-[15px] border border-[rgba(0,180,216,0.4)] bg-gradient-to-r from-[rgba(3,14,32,0.8)] to-[#1E3A5F] px-5 py-2 text-base text-white ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-[#b2bbc9] placeholder:opacity-60 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#00b4d8] disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }
