import { describe, it, expect } from 'vitest'
import { cn } from './utils'

describe('cn utility function', () => {
  it('should merge class names correctly', () => {
    const result = cn('foo', 'bar')
    expect(result).toBe('foo bar')
  })

  it('should handle empty inputs', () => {
    const result = cn()
    expect(result).toBe('')
  })

  it('should handle conditional classes', () => {
    const isActive = true
    const result = cn('base', isActive && 'active')
    expect(result).toBe('base active')
  })

  it('should handle false conditional classes', () => {
    const isActive = false
    const result = cn('base', isActive && 'active')
    expect(result).toBe('base')
  })

  it('should merge Tailwind classes correctly', () => {
    const result = cn('px-4 py-2', 'px-6')
    expect(result).toBe('py-2 px-6')
  })

  it('should handle arrays of classes', () => {
    const result = cn(['foo', 'bar'], 'baz')
    expect(result).toBe('foo bar baz')
  })

  it('should handle objects for conditional classes', () => {
    const result = cn({
      'text-red-500': true,
      'text-blue-500': false,
    })
    expect(result).toBe('text-red-500')
  })

  it('should handle mixed input types', () => {
    const result = cn(
      'base-class',
      { 'conditional-class': true },
      ['array-class'],
      undefined,
      null,
      false,
    )
    expect(result).toBe('base-class conditional-class array-class')
  })
})
