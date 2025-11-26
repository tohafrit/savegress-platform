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
        // Dark theme colors - exact Figma values
        dark: {
          bg: '#030E20',
          'bg-secondary': '#030E21',
          'bg-card': '#0F2744',
          'bg-card-hover': '#142d4f',
          surface: '#030E20',
        },
        primary: {
          DEFAULT: '#1E3A5F',
          dark: '#0F2744',
          light: '#2A4A73',
        },
        accent: {
          orange: '#FF6B35',
          'orange-light': '#FF8A5C',
          yellow: '#FFB800',
          blue: '#00B4D8',
          cyan: '#00B4D8',
          'cyan-bright': '#01C8EF',
        },
        neutral: {
          white: '#FFFFFF',
          'light-gray': '#F5F7FA',
          gray: '#B2BBC9',
          'dark-gray': '#6B7280',
          black: '#0A0A0A',
        },
        text: {
          primary: '#FFFFFF',
          secondary: '#B2BBC9',
          muted: '#6B7280',
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
        // Figma exact sizes
        'display-hero': ['60px', { lineHeight: '80px', fontWeight: '800', letterSpacing: '4.8px' }],
        'display-lg': ['64px', { lineHeight: '1.1', fontWeight: '700' }],
        'display-md': ['48px', { lineHeight: '1.2', fontWeight: '700' }],
        'heading-lg': ['40px', { lineHeight: '46px', fontWeight: '600', letterSpacing: '1.6px' }],
        'heading-md': ['32px', { lineHeight: '1.3', fontWeight: '600' }],
        'heading-sm': ['22px', { lineHeight: '28px', fontWeight: '700', letterSpacing: '0.88px' }],
        'body-lg': ['24px', { lineHeight: '38px', fontWeight: '300', letterSpacing: '1.92px' }],
        'body-md': ['16px', { lineHeight: '28px', fontWeight: '400', letterSpacing: '0.64px' }],
        'btn': ['16px', { lineHeight: '26px', fontWeight: '700', letterSpacing: '0.64px' }],
      },
      spacing: {
        'section': '60px',
        'section-mobile': '40px',
      },
      borderRadius: {
        lg: '12px',
        md: '8px',
        sm: '6px',
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-dark': 'linear-gradient(180deg, #030E20 0%, #030E21 100%)',
        // Figma button gradients
        'gradient-btn-primary': 'linear-gradient(90deg, #00B4D8 0%, #1E3A5F 112.68%)',
        'gradient-btn-secondary': 'linear-gradient(90deg, rgba(10, 10, 10, 0.60) 0%, #1E3A5F 100%)',
        'gradient-card': 'linear-gradient(90deg, rgba(3, 14, 32, 0.80) 0%, #1E3A5F 100%)',
        'gradient-savegress-box': 'linear-gradient(90deg, rgba(10, 10, 10, 0.60) 0%, #1E3A5F 100%)',
      },
      boxShadow: {
        'glow-orange': '0 0 20px rgba(255, 107, 53, 0.3)',
        'glow-blue': '0 0 20px rgba(0, 180, 216, 0.3)',
        'card-dark': '0 4px 20px rgba(0, 0, 0, 0.4)',
      },
      keyframes: {
        "fade-in": {
          "0%": { opacity: '0', transform: 'translateY(20px)' },
          "100%": { opacity: '1', transform: 'translateY(0)' },
        },
        "pulse-glow": {
          "0%, 100%": { opacity: '0.4' },
          "50%": { opacity: '0.8' },
        },
        "float": {
          "0%, 100%": { transform: 'translateY(0px)' },
          "50%": { transform: 'translateY(-10px)' },
        },
        "pulse": {
          "0%, 100%": { opacity: '1', transform: 'scale(1)' },
          "50%": { opacity: '0.7', transform: 'scale(1.05)' },
        },
        "glow": {
          "0%, 100%": { boxShadow: '0 0 20px rgba(0, 180, 216, 0.3)' },
          "50%": { boxShadow: '0 0 40px rgba(0, 180, 216, 0.6)' },
        },
        "gradient-x": {
          "0%, 100%": { backgroundPosition: '0% 50%' },
          "50%": { backgroundPosition: '100% 50%' },
        },
      },
      animation: {
        "fade-in": "fade-in 0.6s ease-out",
        "pulse-glow": "pulse-glow 2s ease-in-out infinite",
        "float": "float 3s ease-in-out infinite",
        "pulse": "pulse 2s ease-in-out infinite",
        "glow": "glow 2s ease-in-out infinite",
        "gradient-x": "gradient-x 3s ease infinite",
      },
    },
  },
  plugins: [],
} satisfies Config

export default config
