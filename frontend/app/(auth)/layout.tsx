'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { AuthProvider, useAuth } from '@/lib/auth-context';
import { Particles } from '@/components/ui/particles';

function AuthContent({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-bg">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan"></div>
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-dark-bg flex flex-col relative overflow-hidden">
      {/* Floating particles */}
      <Particles count={30} />

      {/* Background */}
      <div className="absolute inset-0 z-0">
        <img
          src="/images/bg-hero.png"
          alt=""
          className="w-full h-full object-cover opacity-50"
        />
      </div>

      <header className="relative z-10 p-6 md:p-8">
        <Link href="/">
          <Image
            src="/images/logo.svg"
            alt="Savegress"
            width={176}
            height={49}
            className="h-10 md:h-12 w-auto"
          />
        </Link>
      </header>

      <main className="flex-1 flex items-center justify-center p-4 relative z-10">
        <div className="w-full max-w-md">{children}</div>
      </main>

      <footer className="relative z-10 p-6 text-center text-mini-1 text-grey">
        &copy; {new Date().getFullYear()} Savegress. All rights reserved.
      </footer>
    </div>
  );
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <AuthContent>{children}</AuthContent>
    </AuthProvider>
  );
}
