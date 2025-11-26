'use client';

import { useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import { CheckCircle } from 'lucide-react';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    const result = await api.forgotPassword(email);

    if (result.error) {
      setError(result.error);
    } else {
      setSuccess(true);
    }
    setIsLoading(false);
  };

  if (success) {
    return (
      <div className="card-dark p-8 md:p-10 text-center">
        <CheckCircle className="w-16 h-16 text-cyan mx-auto mb-4" />
        <h1 className="text-h3 text-white mb-2">Check your email</h1>
        <p className="text-content-1 text-grey mb-6">
          We&apos;ve sent password reset instructions to <span className="text-white">{email}</span>
        </p>
        <Link href="/login" className="text-cyan hover:text-cyan/80 transition-colors">
          Back to sign in
        </Link>
      </div>
    );
  }

  return (
    <div className="card-dark p-8 md:p-10">
      <h1 className="text-h3 text-white mb-2 text-center">
        Reset password
      </h1>
      <p className="text-content-1 text-grey mb-8 text-center">
        Enter your email and we&apos;ll send you instructions to reset your password.
      </p>

      {error && (
        <div className="mb-6 p-3 bg-accent-orange/10 border border-accent-orange/30 text-accent-orange rounded-lg text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label htmlFor="email" className="text-content-1 text-white block mb-2">
            Email address <span className="text-cyan">*</span>
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="input-field w-full h-[44px] text-white placeholder:text-text-muted"
            placeholder="you@company.com"
          />
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="btn-primary w-full h-[52px]"
        >
          {isLoading ? 'Sending...' : 'Send reset instructions  →'}
        </button>
      </form>

      <p className="mt-8 text-center text-content-1 text-grey">
        Remember your password?{' '}
        <Link href="/login" className="text-cyan hover:text-cyan/80 transition-colors">
          Sign in
        </Link>
      </p>
    </div>
  );
}
