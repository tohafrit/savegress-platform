import type { Metadata } from "next"
import { Inter, JetBrains_Mono } from "next/font/google"
import "./globals.css"

const inter = Inter({
  subsets: ["latin"],
  variable: '--font-inter',
})

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: '--font-jetbrains',
})

export const metadata: Metadata = {
  title: "Savegress - Replicate data across clouds. Pay less for egress.",
  description: "Stream database changes between AWS, GCP, and Azure. Compress up to 200x to cut your data transfer costs.",
  keywords: ["CDC", "Change Data Capture", "Multi-cloud", "Database Replication", "PostgreSQL", "MySQL"],
  authors: [{ name: "Savegress" }],
  openGraph: {
    title: "Savegress - Replicate data across clouds. Pay less for egress.",
    description: "Stream database changes between AWS, GCP, and Azure. Compress up to 200x to cut your data transfer costs.",
    url: "https://savegress.com",
    siteName: "Savegress",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Savegress - Replicate data across clouds. Pay less for egress.",
    description: "Stream database changes between AWS, GCP, and Azure. Compress up to 200x to cut your data transfer costs.",
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={`${inter.variable} ${jetbrainsMono.variable}`}>
      <body className="font-sans antialiased">{children}</body>
    </html>
  )
}
