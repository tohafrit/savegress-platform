"use client"

import { useState, useEffect, useRef } from "react"

export function useCountUp(end: number, duration: number = 500) {
  const [count, setCount] = useState(0)
  const prevEnd = useRef(end)

  useEffect(() => {
    const startValue = prevEnd.current
    const endValue = end
    prevEnd.current = end

    if (startValue === endValue) {
      setCount(endValue)
      return
    }

    const startTime = performance.now()
    const diff = endValue - startValue

    const animate = (currentTime: number) => {
      const elapsed = currentTime - startTime
      const progress = Math.min(elapsed / duration, 1)

      // Easing function (ease-out)
      const easeOut = 1 - Math.pow(1 - progress, 3)

      const currentValue = startValue + diff * easeOut
      setCount(currentValue)

      if (progress < 1) {
        requestAnimationFrame(animate)
      } else {
        setCount(endValue)
      }
    }

    requestAnimationFrame(animate)
  }, [end, duration])

  return count
}
