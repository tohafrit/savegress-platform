'use client';

import { useEffect, useState } from 'react';
import { api, Subscription, Invoice } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { CreditCard, FileText, ExternalLink, Check, Sparkles } from 'lucide-react';

const plans = [
  {
    id: 'pro',
    name: 'Pro',
    price: '$99',
    period: '/month',
    features: [
      'Up to 5 instances',
      '50,000 events/sec',
      'Email support',
      'Standard compression',
    ],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: '$499',
    period: '/month',
    popular: true,
    features: [
      'Unlimited instances',
      'Unlimited throughput',
      'Priority support',
      'Advanced compression',
      'SSO / SAML',
      'Custom SLA',
    ],
  },
];

export default function BillingPage() {
  const { user } = useAuth();
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      const [subRes, invRes] = await Promise.all([
        api.getSubscription(),
        api.getInvoices(),
      ]);

      if (subRes.data) setSubscription(subRes.data);
      if (invRes.data) setInvoices(invRes.data.invoices);
      setIsLoading(false);
    }
    loadData();
  }, []);

  async function handleManageBilling() {
    const { data } = await api.createPortalSession();
    if (data?.url) {
      window.location.href = data.url;
    }
  }

  if (isLoading) {
    return <BillingSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-h4 text-white">Billing</h1>
        <p className="text-content-1 text-grey">Manage your subscription and billing</p>
      </div>

      {/* Current plan */}
      <div className="card-dark p-6">
        <h2 className="text-h5 text-white mb-4">Current Plan</h2>

        {subscription ? (
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-accent-cyan capitalize">
                {subscription.plan}
              </p>
              <p className="text-sm text-grey mt-1">
                {subscription.status === 'active' ? (
                  <>Next billing date: {new Date(subscription.current_period_end).toLocaleDateString()}</>
                ) : subscription.status === 'trialing' ? (
                  <>Trial ends: {new Date(subscription.current_period_end).toLocaleDateString()}</>
                ) : (
                  <>Status: {subscription.status}</>
                )}
              </p>
              {subscription.cancel_at_period_end && (
                <p className="text-sm text-accent-orange mt-1">
                  Will be canceled at period end
                </p>
              )}
            </div>
            <button
              onClick={handleManageBilling}
              className="btn-secondary px-5 py-2.5 text-sm"
            >
              <CreditCard className="w-4 h-4 mr-2" />
              Manage Billing
            </button>
          </div>
        ) : (
          <div>
            <p className="text-grey mb-4">
              You&apos;re on the free Community plan. Upgrade to unlock more features.
            </p>
          </div>
        )}
      </div>

      {/* Plans */}
      {!subscription && (
        <div className="grid md:grid-cols-2 gap-4">
          {plans.map((plan) => (
            <div
              key={plan.id}
              className={`card-dark p-6 relative ${
                plan.popular ? 'border-accent-cyan' : ''
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="inline-flex items-center gap-1 px-3 py-1 bg-accent-cyan text-white text-xs font-medium rounded-full">
                    <Sparkles className="w-3 h-3" />
                    Popular
                  </span>
                </div>
              )}
              <h3 className="text-h5 text-white">{plan.name}</h3>
              <p className="mt-2">
                <span className="text-3xl font-bold text-white">{plan.price}</span>
                <span className="text-grey">{plan.period}</span>
              </p>
              <ul className="mt-4 space-y-2">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2 text-sm text-grey">
                    <Check className="w-4 h-4 text-accent-cyan" />
                    {feature}
                  </li>
                ))}
              </ul>
              <button
                onClick={() => {/* TODO: Implement Stripe checkout */}}
                className={`mt-6 w-full py-2.5 px-4 rounded-[20px] text-sm font-medium transition-all ${
                  plan.popular
                    ? 'btn-primary'
                    : 'btn-secondary'
                }`}
              >
                Upgrade to {plan.name}
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Invoices */}
      <div className="card-dark overflow-hidden">
        <div className="p-5 border-b border-cyan-40">
          <h2 className="text-h5 text-white">Invoice History</h2>
        </div>

        {invoices.length === 0 ? (
          <div className="p-12 text-center">
            <FileText className="w-12 h-12 mx-auto mb-3 text-grey opacity-50" />
            <p className="text-grey">No invoices yet</p>
          </div>
        ) : (
          <div className="divide-y divide-cyan-40/30">
            {invoices.map((invoice) => (
              <div key={invoice.id} className="p-4 flex items-center justify-between hover:bg-primary-dark/30 transition-colors">
                <div>
                  <p className="font-medium text-white">
                    {new Date(invoice.created_at).toLocaleDateString()}
                  </p>
                  <p className="text-sm text-grey">
                    {formatCurrency(invoice.amount, invoice.currency)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
                      invoice.status === 'paid'
                        ? 'bg-accent-cyan/20 text-accent-cyan border-accent-cyan/40'
                        : invoice.status === 'open'
                        ? 'bg-accent-orange/20 text-accent-orange border-accent-orange/40'
                        : 'bg-grey/20 text-grey border-grey/40'
                    }`}
                  >
                    {invoice.status}
                  </span>
                  {invoice.pdf_url && (
                    <a
                      href={invoice.pdf_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-2 text-grey hover:text-accent-cyan transition-colors"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </a>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function BillingSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-32 bg-primary-dark rounded animate-pulse" />
        <div className="h-4 w-48 bg-primary-dark rounded animate-pulse mt-2" />
      </div>
      <div className="card-dark p-6">
        <div className="h-6 w-32 bg-primary-dark rounded animate-pulse mb-4" />
        <div className="h-8 w-24 bg-primary-dark rounded animate-pulse" />
      </div>
      <div className="grid md:grid-cols-2 gap-4">
        {[1, 2].map((i) => (
          <div key={i} className="card-dark p-6">
            <div className="h-6 w-24 bg-primary-dark rounded animate-pulse mb-2" />
            <div className="h-8 w-20 bg-primary-dark rounded animate-pulse mb-4" />
            <div className="space-y-2">
              {[1, 2, 3, 4].map((j) => (
                <div key={j} className="h-4 w-full bg-primary-dark/50 rounded animate-pulse" />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function formatCurrency(amount: number, currency: string): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency.toUpperCase(),
  }).format(amount / 100);
}
