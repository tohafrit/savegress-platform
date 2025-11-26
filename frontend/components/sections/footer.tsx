"use client"

export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-dark-surface py-10 md:py-16 relative overflow-hidden">
      {/* Background image */}
      <img
        src="/images/bg-footer.png"
        alt=""
        className="absolute inset-0 w-full md:w-[1873px] h-full md:h-[1061px] opacity-20 pointer-events-none"
      />

      <div className="container-custom relative z-10">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 md:gap-8 mb-8 md:mb-12">
          {/* Logo */}
          <div className="col-span-2 md:col-span-1 mb-4 md:mb-0">
            <img
              src="/images/logo-footer.svg"
              alt="Savegress"
              className="w-[140px] md:w-[176px] h-auto"
            />
          </div>

          {/* Product */}
          <div>
            <h4 className="text-h6 mb-3 md:mb-4">Product</h4>
            <ul className="space-y-2 md:space-y-3">
              <li>
                <a href="/docs" className="footer-link">Documentation</a>
              </li>
            </ul>
          </div>

          {/* Company */}
          <div>
            <h4 className="text-h6 mb-3 md:mb-4">Company</h4>
            <ul className="space-y-2 md:space-y-3">
              <li>
                <a href="/about" className="footer-link">About</a>
              </li>
            </ul>
          </div>

          {/* Legal */}
          <div className="col-span-2 md:col-span-1">
            <h4 className="text-h6 mb-3 md:mb-4">Legal</h4>
            <ul className="space-y-2 md:space-y-3">
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
        <div className="w-full h-[1px] border-t border-dashed border-[#02ACD0]/50 mb-6 md:mb-8" />

        <div className="flex flex-col items-center pt-4 md:pt-8 gap-4">
          <p className="text-mini-1 text-cyan w-full max-w-[592px] text-center">
            &copy; {currentYear} Savegress. All rights reserved.
          </p>
{/*
          <p className="text-mini-1 text-gray-500 w-full max-w-[720px] text-center text-[11px] leading-relaxed">
            AWS is a trademark of Amazon.com, Inc. Google Cloud is a trademark of Google LLC. Microsoft Azure is a trademark of Microsoft Corporation. Savegress is not affiliated with or endorsed by these companies.
          </p>
*/}
        </div>
      </div>
    </footer>
  )
}
