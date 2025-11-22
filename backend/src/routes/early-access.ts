import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify'
import { z } from 'zod'
import { prisma } from '../services/database.service'
import { sendEarlyAccessEmail } from '../services/email.service'

const TURNSTILE_SECRET_KEY = process.env.TURNSTILE_SECRET_KEY || '0x4AAAAAACCXwEK23TURNS62usztaHWOSnE'

async function verifyTurnstileToken(token: string, ip?: string): Promise<boolean> {
  // Skip verification if using test key
  if (TURNSTILE_SECRET_KEY.startsWith('1x0000')) {
    return true
  }

  const response = await fetch('https://challenges.cloudflare.com/turnstile/v0/siteverify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      secret: TURNSTILE_SECRET_KEY,
      response: token,
      ...(ip && { remoteip: ip }),
    }),
  })

  const data = await response.json() as { success: boolean }
  return data.success
}

const earlyAccessSchema = z.object({
  email: z.string().email(),
  company: z.string().min(2),
  currentSolution: z
    .enum(['kafka', 'dms', 'fivetran', 'none', 'other'])
    .optional(),
  dataVolume: z.enum(['<10GB', '10-100GB', '100GB-1TB', '>1TB']).optional(),
  message: z.string().optional(),
  turnstileToken: z.string().min(1, 'Captcha verification required'),
})

type EarlyAccessBody = z.infer<typeof earlyAccessSchema>

export async function earlyAccessRoutes(fastify: FastifyInstance) {
  fastify.post<{ Body: EarlyAccessBody }>(
    '/api/early-access',
    {
      schema: {
        body: {
          type: 'object',
          required: ['email', 'company', 'turnstileToken'],
          properties: {
            email: { type: 'string', format: 'email' },
            company: { type: 'string', minLength: 2 },
            currentSolution: {
              type: 'string',
              enum: ['kafka', 'dms', 'fivetran', 'none', 'other'],
            },
            dataVolume: {
              type: 'string',
              enum: ['<10GB', '10-100GB', '100GB-1TB', '>1TB'],
            },
            message: { type: 'string' },
            turnstileToken: { type: 'string' },
          },
        },
      },
    },
    async (request: FastifyRequest<{ Body: EarlyAccessBody }>, reply: FastifyReply) => {
      try {
        // Validate with Zod
        const data = earlyAccessSchema.parse(request.body)

        // Verify Turnstile token
        const isValidCaptcha = await verifyTurnstileToken(data.turnstileToken, request.ip)
        if (!isValidCaptcha) {
          return reply.status(400).send({
            success: false,
            message: 'Captcha verification failed',
          })
        }

        // Save to database
        const earlyAccessRequest = await prisma.earlyAccessRequest.create({
          data: {
            email: data.email,
            company: data.company,
            currentSolution: data.currentSolution,
            dataVolume: data.dataVolume,
            message: data.message,
            ipAddress: request.ip,
            userAgent: request.headers['user-agent'],
          },
        })

        // Send notification email (async, don't wait)
        sendEarlyAccessEmail(data).catch((err) =>
          fastify.log.error(err, 'Failed to send email')
        )

        fastify.log.info(
          { requestId: earlyAccessRequest.id, email: data.email },
          'New early access request'
        )

        return reply.status(201).send({
          success: true,
          message: 'Request submitted successfully',
        })
      } catch (error) {
        if (error instanceof z.ZodError) {
          return reply.status(400).send({
            success: false,
            errors: error.errors.map((e) => ({
              path: e.path.join('.'),
              message: e.message,
            })),
          })
        }

        fastify.log.error(error, 'Early access request failed')
        return reply.status(500).send({
          success: false,
          message: 'Internal server error',
        })
      }
    }
  )
}
