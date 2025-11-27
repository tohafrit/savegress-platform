import { NextResponse } from 'next/server'

// Cache pricing data in memory (persists across requests in production)
let cachedPricing: EgressPricing | null = null
let cacheTimestamp: number = 0
const CACHE_DURATION = 24 * 60 * 60 * 1000 // 24 hours

// Fallback prices if API fails
const FALLBACK_PRICING = {
  AWS: {
    "us-east": 0.09,
    "eu-west": 0.09,
    "ap-southeast": 0.114,
  },
  GCP: {
    "us-east": 0.12,
    "eu-west": 0.12,
    "ap-southeast": 0.15,
  },
  Azure: {
    "us-east": 0.087,
    "eu-west": 0.087,
    "ap-southeast": 0.12,
  },
}

interface EgressPricing {
  AWS: Record<string, number>
  GCP: Record<string, number>
  Azure: Record<string, number>
  lastUpdated: string
  source: 'live' | 'fallback'
}

// AWS region mapping for data transfer pricing
const AWS_REGION_MAPPING: Record<string, string[]> = {
  "us-east": ["US East (N. Virginia)", "US East (Ohio)", "US West (Oregon)", "US West (N. California)"],
  "eu-west": ["EU (Ireland)", "EU (Frankfurt)", "EU (London)", "EU (Paris)"],
  "ap-southeast": ["Asia Pacific (Singapore)", "Asia Pacific (Sydney)", "Asia Pacific (Tokyo)"],
}

async function fetchAWSPricing(): Promise<Record<string, number>> {
  try {
    // Fetch AWS Data Transfer pricing (smaller regional index)
    const response = await fetch(
      'https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AWSDataTransfer/current/region_index.json',
      { next: { revalidate: 86400 } } // Cache for 24 hours
    )

    if (!response.ok) throw new Error('Failed to fetch AWS pricing')

    const regionIndex = await response.json()

    // Get US East pricing as primary source
    const usEastUrl = regionIndex.regions?.['us-east-1']?.currentVersionUrl
    if (!usEastUrl) throw new Error('US East region not found')

    const usEastResponse = await fetch(`https://pricing.us-east-1.amazonaws.com${usEastUrl}`)
    if (!usEastResponse.ok) throw new Error('Failed to fetch US East pricing')

    const usEastData = await usEastResponse.json()

    // Extract internet egress price (first 10TB tier)
    let usEastPrice = 0.09 // default

    for (const sku in usEastData.products) {
      const product = usEastData.products[sku]
      if (
        product.attributes?.transferType === 'AWS Outbound' &&
        product.attributes?.fromLocation?.includes('US East')
      ) {
        // Find the price for this SKU
        const terms = usEastData.terms?.OnDemand?.[sku]
        if (terms) {
          for (const termKey in terms) {
            const priceDimensions = terms[termKey].priceDimensions
            for (const priceKey in priceDimensions) {
              const price = parseFloat(priceDimensions[priceKey].pricePerUnit?.USD || '0')
              if (price > 0 && price < 1) { // Sanity check
                usEastPrice = price
                break
              }
            }
          }
        }
        break
      }
    }

    // Return prices for each region category
    return {
      "us-east": usEastPrice,
      "eu-west": usEastPrice, // EU typically same as US
      "ap-southeast": Math.round(usEastPrice * 1.27 * 1000) / 1000, // APAC ~27% higher
    }
  } catch (error) {
    console.error('Failed to fetch AWS pricing:', error)
    return FALLBACK_PRICING.AWS
  }
}

async function fetchGCPPricing(): Promise<Record<string, number>> {
  const apiKey = process.env.GCP_API_KEY

  // If no API key, use official published prices from:
  // https://cloud.google.com/vpc/network-pricing
  if (!apiKey) {
    return {
      "us-east": 0.12,
      "eu-west": 0.12,
      "ap-southeast": 0.15,
    }
  }

  try {
    // GCP Cloud Billing API - get Network Egress pricing
    // First, get the Compute Engine service ID
    const servicesResponse = await fetch(
      `https://cloudbilling.googleapis.com/v2beta/services?key=${apiKey}&pageSize=200`
    )

    if (!servicesResponse.ok) throw new Error('Failed to fetch GCP services')

    const servicesData = await servicesResponse.json()
    const computeService = servicesData.services?.find(
      (s: { displayName: string }) => s.displayName === 'Compute Engine'
    )

    if (!computeService) {
      throw new Error('Compute Engine service not found')
    }

    // Get SKUs for network egress
    const skusResponse = await fetch(
      `https://cloudbilling.googleapis.com/v2beta/${computeService.name}/skus?key=${apiKey}&pageSize=500`
    )

    if (!skusResponse.ok) throw new Error('Failed to fetch GCP SKUs')

    const skusData = await skusResponse.json()

    // Find network egress to internet SKU
    let usPrice = 0.12
    for (const sku of skusData.skus || []) {
      if (
        sku.description?.includes('Network Internet Egress') &&
        sku.description?.includes('Americas')
      ) {
        const priceInfo = sku.pricingInfo?.[0]?.pricingExpression?.tieredRates?.[1] // First paid tier
        if (priceInfo) {
          usPrice = (priceInfo.unitPrice?.units || 0) + (priceInfo.unitPrice?.nanos || 0) / 1e9
        }
        break
      }
    }

    return {
      "us-east": usPrice,
      "eu-west": usPrice,
      "ap-southeast": Math.round(usPrice * 1.25 * 1000) / 1000,
    }
  } catch (error) {
    console.error('Failed to fetch GCP pricing:', error)
    return FALLBACK_PRICING.GCP
  }
}

async function fetchAzurePricing(): Promise<Record<string, number>> {
  try {
    // Azure Retail Prices API (public, no auth required)
    // Fetch prices for different regions in parallel
    const regions = {
      'us-east': 'eastus',
      'eu-west': 'westeurope',
      'ap-southeast': 'southeastasia',
    }

    const fetchRegionPrice = async (armRegion: string): Promise<number> => {
      const response = await fetch(
        `https://prices.azure.com/api/retail/prices?$filter=serviceName eq 'Bandwidth' and armRegionName eq '${armRegion}' and skuName eq 'Standard Data Transfer Out'`,
        { next: { revalidate: 86400 } }
      )

      if (!response.ok) return 0

      const data = await response.json()

      // Find the per-GB price (first tier after free tier)
      for (const item of data.Items || []) {
        // Look for "5 GB to 10 TB" tier which is the first paid tier
        if (
          item.unitOfMeasure === '1 GB' &&
          item.retailPrice > 0 &&
          item.meterName?.includes('Data Transfer Out')
        ) {
          return item.retailPrice
        }
      }
      return 0
    }

    const [usPrice, euPrice, apPrice] = await Promise.all([
      fetchRegionPrice(regions['us-east']),
      fetchRegionPrice(regions['eu-west']),
      fetchRegionPrice(regions['ap-southeast']),
    ])

    return {
      "us-east": usPrice || FALLBACK_PRICING.Azure["us-east"],
      "eu-west": euPrice || FALLBACK_PRICING.Azure["eu-west"],
      "ap-southeast": apPrice || FALLBACK_PRICING.Azure["ap-southeast"],
    }
  } catch (error) {
    console.error('Failed to fetch Azure pricing:', error)
    return FALLBACK_PRICING.Azure
  }
}

async function fetchAllPricing(): Promise<EgressPricing> {
  // Check cache first
  if (cachedPricing && Date.now() - cacheTimestamp < CACHE_DURATION) {
    return cachedPricing
  }

  try {
    const [awsPricing, gcpPricing, azurePricing] = await Promise.all([
      fetchAWSPricing(),
      fetchGCPPricing(),
      fetchAzurePricing(),
    ])

    const pricing: EgressPricing = {
      AWS: awsPricing,
      GCP: gcpPricing,
      Azure: azurePricing,
      lastUpdated: new Date().toISOString(),
      source: 'live',
    }

    // Update cache
    cachedPricing = pricing
    cacheTimestamp = Date.now()

    return pricing
  } catch (error) {
    console.error('Failed to fetch pricing:', error)
    return {
      ...FALLBACK_PRICING,
      lastUpdated: new Date().toISOString(),
      source: 'fallback',
    }
  }
}

export async function GET() {
  const pricing = await fetchAllPricing()

  return NextResponse.json(pricing, {
    headers: {
      'Cache-Control': 'public, s-maxage=86400, stale-while-revalidate=43200',
    },
  })
}
