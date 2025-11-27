import { MetadataRoute } from 'next'

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: [
        '/dashboard',
        '/settings',
        '/billing',
        '/licenses',
        '/connections',
        '/pipelines',
        '/downloads',
        '/setup',
      ],
    },
    sitemap: 'https://savegress.com/sitemap.xml',
  }
}
