"use client"

export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-dark-surface py-16 relative overflow-hidden">
      {/* Background image */}
      <img
        src="/images/bg-footer.png"
        alt=""
        className="absolute inset-0 w-[1873px] h-[1061px] opacity-20 pointer-events-none"
      />

      <div className="container-custom relative z-10">
        <div className="grid md:grid-cols-4 gap-8 mb-12">
          {/* Logo */}
          <div className="md:col-span-1">
            <img
              src="/images/logo-footer.svg"
              alt="Savegress"
              className="w-[176px] h-[49px]"
            />
          </div>

          {/* Product */}
          <div>
            <h4 className="text-h6 mb-4">Product</h4>
            <ul className="space-y-3">
              <li>
                <a href="/docs" className="footer-link">Documentation</a>
              </li>
            </ul>
          </div>

          {/* Company */}
          <div>
            <h4 className="text-h6 mb-4">Company</h4>
            <ul className="space-y-3">
              <li>
                <a href="/about" className="footer-link">About</a>
              </li>
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h4 className="text-h6 mb-4">Legal</h4>
            <ul className="space-y-3">
              <li>
                <a href="/privacy" className="footer-link">Privacy Policy</a>
              </li>
              <li>
                <a href="/terms" className="footer-link">Terms of Service</a>
              </li>
            </ul>
          </div>
        </div>

        {/* Divider line */}
        <div className="flex justify-center mb-8">
          <svg width="1216" height="1" viewBox="0 0 1216 1" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M0.300003 0.299805H1215.3" stroke="#02ACD0" strokeWidth="0.6" strokeLinecap="round" strokeDasharray="4 4"/>
          </svg>
        </div>

        <div className="flex justify-center pt-8">
          <p className="text-mini-1 text-cyan w-[592px] h-[21px] text-center">
            &copy; {currentYear} Savegress. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  )
}
