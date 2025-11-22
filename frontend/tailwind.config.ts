import type { Config } from "tailwindcss"

const config = {
  darkMode: ["class"],
  content: [
    './pages/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './app/**/*.{ts,tsx}',
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        primary: {
          DEFAULT: '#1E3A5F',
          dark: '#0F2744',
        },
        accent: {
          orange: '#FF6B35',
          yellow: '#FFB800',
          blue: '#00B4D8',
        },
        neutral: {
          white: '#FFFFFF',
          'light-gray': '#F5F7FA',
          'dark-gray': '#1F2937',
          black: '#0A0A0A',
        },
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Courier New', 'monospace'],
      },
      fontSize: {
        'display-lg': ['64px', { lineHeight: '1.1', fontWeight: '700' }],
        'display-md': ['48px', { lineHeight: '1.2', fontWeight: '700' }],
        'heading-lg': ['40px', { lineHeight: '1.2', fontWeight: '600' }],
        'heading-md': ['32px', { lineHeight: '1.3', fontWeight: '600' }],
        'heading-sm': ['24px', { lineHeight: '1.4', fontWeight: '600' }],
        'body-lg': ['18px', { lineHeight: '1.6', fontWeight: '400' }],
        'body-md': ['16px', { lineHeight: '1.5', fontWeight: '400' }],
      },
      spacing: {
        'section': '120px',
        'section-mobile': '80px',
      },
      borderRadius: {
        lg: '12px',
        md: '8px',
        sm: '6px',
      },
      keyframes: {
        "fade-in": {
          "0%": { opacity: '0', transform: 'translateY(20px)' },
          "100%": { opacity: '1', transform: 'translateY(0)' },
        },
      },
      animation: {
        "fade-in": "fade-in 0.6s ease-out",
      },
    },
  },
  plugins: [],
} satisfies Config

export default config
