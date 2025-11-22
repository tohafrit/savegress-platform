import { Resend } from 'resend'

interface EarlyAccessData {
  email: string
  company: string
  currentSolution?: string
  dataVolume?: string
  message?: string
}

const resend = process.env.RESEND_API_KEY
  ? new Resend(process.env.RESEND_API_KEY)
  : null

const ADMIN_EMAIL = process.env.ADMIN_EMAIL || 'pakhunov.anton.n@gmail.com'
const FROM_EMAIL = process.env.EMAIL_FROM || 'Savegress <noreply@savegress.com>'

export async function sendEarlyAccessEmail(data: EarlyAccessData): Promise<void> {
  console.log('New early access request:', {
    email: data.email,
    company: data.company,
  })

  if (!resend) {
    console.log('RESEND_API_KEY not configured, skipping email')
    return
  }

  // Send notification to admin
  await resend.emails.send({
    from: FROM_EMAIL,
    to: ADMIN_EMAIL,
    subject: `New Early Access Request from ${data.company}`,
    html: `
      <h2>New Early Access Request</h2>
      <p><strong>Email:</strong> ${data.email}</p>
      <p><strong>Company:</strong> ${data.company}</p>
      <p><strong>Current Solution:</strong> ${data.currentSolution || 'Not specified'}</p>
      <p><strong>Data Volume:</strong> ${data.dataVolume || 'Not specified'}</p>
      <p><strong>Message:</strong> ${data.message || 'No message'}</p>
    `,
  })

  // Send confirmation to user
  await resend.emails.send({
    from: FROM_EMAIL,
    to: data.email,
    subject: 'Welcome to Savegress Early Access',
    html: `
      <h2>Thank you for your interest in Savegress!</h2>
      <p>Hi,</p>
      <p>We've received your early access request for <strong>${data.company}</strong>.</p>
      <p>Our team will review your application and get back to you soon with next steps.</p>
      <br>
      <p>Best regards,<br>The Savegress Team</p>
    `,
  })
}
