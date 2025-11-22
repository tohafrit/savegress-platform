export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-primary text-white py-12">
      <div className="container-custom">
        <div className="grid md:grid-cols-4 gap-8 mb-8">
          {/* Product */}
          <div>
            <h4 className="font-semibold mb-4">Product</h4>
            <ul className="space-y-2 text-sm">
              <li><a href="#" className="hover:text-accent-orange transition-colors">Documentation</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">GitHub</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Status Page</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Changelog</a></li>
            </ul>
          </div>

          {/* Company */}
          <div>
            <h4 className="font-semibold mb-4">Company</h4>
            <ul className="space-y-2 text-sm">
              <li><a href="#" className="hover:text-accent-orange transition-colors">About</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Blog</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Contact</a></li>
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h4 className="font-semibold mb-4">Legal</h4>
            <ul className="space-y-2 text-sm">
              <li><a href="#" className="hover:text-accent-orange transition-colors">Privacy Policy</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Terms of Service</a></li>
            </ul>
          </div>

          {/* Connect */}
          <div>
            <h4 className="font-semibold mb-4">Connect</h4>
            <ul className="space-y-2 text-sm">
              <li><a href="#" className="hover:text-accent-orange transition-colors">LinkedIn</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">Twitter/X</a></li>
              <li><a href="#" className="hover:text-accent-orange transition-colors">GitHub</a></li>
            </ul>
          </div>
        </div>

        <div className="border-t border-white/20 pt-8 text-center text-sm">
          <p>&copy; {currentYear} Savegress. All rights reserved.</p>
        </div>
      </div>
    </footer>
  )
}
