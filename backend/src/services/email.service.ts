interface EarlyAccessData {
  email: string
  company: string
  currentSolution?: string
  dataVolume?: string
  message?: string
}

export async function sendEarlyAccessEmail(data: EarlyAccessData): Promise<void> {
  // TODO: Integrate with email service (Resend, Postmark, etc.)
  console.log('New early access request:', {
    email: data.email,
    company: data.company,
  })

  // In production, implement actual email sending:
  // - Send notification to admin
  // - Send confirmation to user
  // Example with Resend:
  // await resend.emails.send({
  //   from: 'Savegress <noreply@savegress.com>',
  //   to: process.env.ADMIN_EMAIL,
  //   subject: `New Early Access Request from ${data.company}`,
  //   html: `...`
  // })
}
