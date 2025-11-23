'use client';

import { useEffect, useState } from 'react';
import { api, Subscription, Invoice } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { CreditCard, FileText, ExternalLink, Check } from 'lucide-react';

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
        <h1 className="text-2xl font-bold text-primary">Billing</h1>
        <p className="text-neutral-dark-gray">Manage your subscription and billing</p>
      </div>

      {/* Current plan */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-primary mb-4">Current Plan</h2>

        {subscription ? (
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-primary capitalize">
                {subscription.plan}
              </p>
              <p className="text-sm text-neutral-dark-gray">
                {subscription.status === 'active' ? (
                  <>Next billing date: {new Date(subscription.current_period_end).toLocaleDateString()}</>
                ) : subscription.status === 'trialing' ? (
                  <>Trial ends: {new Date(subscription.current_period_end).toLocaleDateString()}</>
                ) : (
                  <>Status: {subscription.status}</>
                )}
              </p>
              {subscription.cancel_at_period_end && (
                <p className="text-sm text-yellow-600 mt-1">
                  Will be canceled at period end
                </p>
              )}
            </div>
            <button
              onClick={handleManageBilling}
              className="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <CreditCard className="w-4 h-4" />
              Manage Billing
            </button>
          </div>
        ) : (
          <div>
            <p className="text-neutral-dark-gray mb-4">
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
              className="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
            >
              <h3 className="text-xl font-bold text-primary">{plan.name}</h3>
              <p className="mt-2">
                <span className="text-3xl font-bold text-primary">{plan.price}</span>
                <span className="text-neutral-dark-gray">{plan.period}</span>
              </p>
              <ul className="mt-4 space-y-2">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2 text-sm text-neutral-dark-gray">
                    <Check className="w-4 h-4 text-green-600" />
                    {feature}
                  </li>
                ))}
              </ul>
              <button
                onClick={() => {/* TODO: Implement Stripe checkout */}}
                className="mt-6 w-full py-2 px-4 bg-primary text-white rounded-md hover:bg-primary-dark"
              >
                Upgrade to {plan.name}
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Invoices */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-primary">Invoice History</h2>
        </div>

        {invoices.length === 0 ? (
          <div className="p-8 text-center text-neutral-dark-gray">
            <FileText className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p>No invoices yet</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {invoices.map((invoice) => (
              <div key={invoice.id} className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium text-primary">
                    {new Date(invoice.created_at).toLocaleDateString()}
                  </p>
                  <p className="text-sm text-neutral-dark-gray">
                    {formatCurrency(invoice.amount, invoice.currency)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                      invoice.status === 'paid'
                        ? 'bg-green-100 text-green-700'
                        : invoice.status === 'open'
                        ? 'bg-yellow-100 text-yellow-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {invoice.status}
                  </span>
                  {invoice.pdf_url && (
                    <a
                      href={invoice.pdf_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-1 text-neutral-dark-gray hover:text-primary"
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
        <div className="h-8 w-32 bg-gray-200 rounded animate-pulse" />
        <div className="h-4 w-48 bg-gray-200 rounded animate-pulse mt-2" />
      </div>
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="h-6 w-32 bg-gray-200 rounded animate-pulse mb-4" />
        <div className="h-8 w-24 bg-gray-200 rounded animate-pulse" />
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
