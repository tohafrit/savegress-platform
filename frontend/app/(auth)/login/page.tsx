'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    const result = await login(email, password);

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
        Sign in
      </h1>
      <p className="text-content-1 text-grey mb-8 text-center">
        Access your Savegress dashboard
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

        <div className="flex items-center justify-between text-sm">
          <label className="flex items-center cursor-pointer">
            <input
              type="checkbox"
              className="w-4 h-4 rounded border-white/20 bg-transparent text-cyan focus:ring-cyan focus:ring-offset-0"
            />
            <span className="ml-2 text-grey">Remember me</span>
          </label>
          <Link href="/forgot-password" className="text-cyan hover:text-cyan/80 transition-colors">
            Forgot password?
          </Link>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="btn-primary w-full h-[52px]"
        >
          {isLoading ? 'Signing in...' : 'Sign in  →'}
        </button>
      </form>

      <p className="mt-8 text-center text-content-1 text-grey">
        Don&apos;t have an account?{' '}
        <Link href="/register" className="text-cyan hover:text-cyan/80 transition-colors">
          Request access
        </Link>
      </p>
    </div>
  );
}
