import Fastify from 'fastify'
import cors from '@fastify/cors'
import helmet from '@fastify/helmet'
import rateLimit from '@fastify/rate-limit'
import { earlyAccessRoutes } from './routes/early-access'
import { healthRoutes } from './routes/health'

export const app = Fastify({
  logger: {
    level: process.env.LOG_LEVEL || 'info',
    transport:
      process.env.NODE_ENV === 'development'
        ? {
            target: 'pino-pretty',
            options: {
              colorize: true,
              translateTime: 'HH:MM:ss Z',
              ignore: 'pid,hostname',
            },
          }
        : undefined,
  },
})

// CORS
app.register(cors, {
  origin: process.env.CORS_ORIGIN || 'http://localhost:3000',
  credentials: true,
})

// Security headers
app.register(helmet, {
  contentSecurityPolicy: false, // Handled by Caddy
})

// Rate limiting
app.register(rateLimit, {
  max: 100,
  timeWindow: '15 minutes',
  errorResponseBuilder: (request, context) => {
    return {
      statusCode: 429,
      error: 'Too Many Requests',
      message: `Rate limit exceeded, retry in ${context.after}`,
    }
  },
})

// Routes
app.register(healthRoutes)
app.register(earlyAccessRoutes)

export default app
