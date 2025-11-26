'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';

export default function RegisterPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { register } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setIsLoading(true);
    const result = await register(email, password, name);

    if (result.error) {
      setError(result.error);
      setIsLoading(false);
    } else {
      router.push('/dashboard');
    }
  };

  return (
    <div className="card-dark p-8 md:p-10">
      <h1 className="text-h3 text-white mb-2 text-center">
        Create account
      </h1>
      <p className="text-content-1 text-grey mb-8 text-center">
        Get started with Savegress
      </p>

      {error && (
        <div className="mb-6 p-3 bg-accent-orange/10 border border-accent-orange/30 text-accent-orange rounded-lg text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="name" className="text-content-1 text-white block mb-2">
            Full name <span className="text-cyan">*</span>
          </label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="input-field w-full h-[44px] text-white placeholder:text-text-muted"
            placeholder="John Doe"
          />
        </div>

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

        <div>
          <label htmlFor="password" className="text-content-1 text-white block mb-2">
            Password <span className="text-cyan">*</span>
          </label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="input-field w-full h-[44px] text-white placeholder:text-text-muted"
            placeholder="••••••••"
          />
        </div>

        <div>
          <label htmlFor="confirmPassword" className="text-content-1 text-white block mb-2">
            Confirm password <span className="text-cyan">*</span>
          </label>
          <input
            id="confirmPassword"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            className="input-field w-full h-[44px] text-white placeholder:text-text-muted"
            placeholder="••••••••"
          />
        </div>

        <p className="text-mini-1 text-grey">
          By creating an account, you agree to our{' '}
          <Link href="/terms" className="text-cyan hover:text-cyan/80">Terms of Service</Link>
          {' '}and{' '}
          <Link href="/privacy" className="text-cyan hover:text-cyan/80">Privacy Policy</Link>.
        </p>

        <button
          type="submit"
          disabled={isLoading}
          className="btn-primary w-full h-[52px]"
        >
          {isLoading ? 'Creating account...' : 'Create account  →'}
        </button>
      </form>

      <p className="mt-8 text-center text-content-1 text-grey">
        Already have an account?{' '}
        <Link href="/login" className="text-cyan hover:text-cyan/80 transition-colors">
          Sign in
        </Link>
      </p>
    </div>
  );
}
