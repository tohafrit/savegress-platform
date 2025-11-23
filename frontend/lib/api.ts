const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class ApiClient {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken(): string | null {
    if (typeof window === 'undefined') return null;
    if (!this.token) {
      this.token = localStorage.getItem('token');
    }
    return this.token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = this.getToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${API_URL}/api/v1${endpoint}`, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        return { error: data.error || 'Request failed' };
      }

      return { data };
    } catch (error) {
      return { error: 'Network error' };
    }
  }

  // Auth
  async register(email: string, password: string, name: string) {
    return this.request<{ user: User; tokens: { access_token: string; refresh_token: string } }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    });
  }

  async login(email: string, password: string) {
    return this.request<{ user: User; tokens: { access_token: string; refresh_token: string } }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async refreshToken(refreshToken: string) {
    return this.request<{ token: string; refresh_token: string }>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  }

  async forgotPassword(email: string) {
    return this.request('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  }

  async resetPassword(token: string, password: string) {
    return this.request('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    });
  }

  // User
  async getProfile() {
    return this.request<User>('/user');
  }

  async updateProfile(data: Partial<User>) {
    return this.request<User>('/user', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async changePassword(currentPassword: string, newPassword: string) {
    return this.request('/user/password', {
      method: 'PUT',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    });
  }

  // Licenses
  async getLicenses() {
    return this.request<{ licenses: License[] }>('/licenses');
  }

  async getLicense(id: string) {
    return this.request<License>(`/licenses/${id}`);
  }

  async createLicense(data: { edition: string; max_instances?: number }) {
    return this.request<License>('/licenses', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async revokeLicense(id: string) {
    return this.request(`/licenses/${id}`, {
      method: 'DELETE',
    });
  }

  async getLicenseActivations(id: string) {
    return this.request<{ activations: Activation[] }>(`/licenses/${id}/activations`);
  }

  // Billing
  async getSubscription() {
    return this.request<Subscription>('/billing/subscription');
  }

  async createSubscription(priceId: string, paymentMethodId?: string) {
    return this.request<{ subscription: Subscription; client_secret?: string }>('/billing/subscription', {
      method: 'POST',
      body: JSON.stringify({ price_id: priceId, payment_method_id: paymentMethodId }),
    });
  }

  async cancelSubscription() {
    return this.request('/billing/subscription', {
      method: 'DELETE',
    });
  }

  async getInvoices() {
    return this.request<{ invoices: Invoice[] }>('/billing/invoices');
  }

  async createPortalSession() {
    return this.request<{ url: string }>('/billing/portal-session', {
      method: 'POST',
    });
  }

  // Dashboard
  async getDashboardStats() {
    return this.request<DashboardStats>('/dashboard/stats');
  }

  async getUsage() {
    return this.request<UsageData>('/dashboard/usage');
  }

  async getInstances() {
    return this.request<{ instances: Instance[] }>('/dashboard/instances');
  }

  // Downloads
  async getDownloads() {
    return this.request<{ downloads: Download[] }>('/downloads');
  }

  async getDownloadURL(product: string, version: string, platform?: string) {
    const params = platform ? `?platform=${platform}` : '';
    return this.request<{ url: string; expires_in: string }>(`/downloads/${product}/${version}${params}`);
  }
}

export const api = new ApiClient();

// Types
export interface User {
  id: string;
  email: string;
  name: string;
  company?: string;
  role: 'user' | 'admin';
  email_verified?: boolean;
  subscription_tier?: 'free' | 'pro' | 'enterprise';
  created_at: string;
  updated_at?: string;
  last_login_at?: string;
}

export interface License {
  id: string;
  key: string;
  edition: 'community' | 'pro' | 'enterprise';
  status: 'active' | 'expired' | 'revoked';
  max_instances: number;
  active_instances: number;
  expires_at: string;
  created_at: string;
}

export interface Activation {
  id: string;
  license_id: string;
  hardware_id: string;
  hostname: string;
  ip_address: string;
  activated_at: string;
  last_seen_at: string;
}

export interface Subscription {
  id: string;
  status: 'active' | 'canceled' | 'past_due' | 'trialing';
  plan: 'pro' | 'enterprise';
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
}

export interface Invoice {
  id: string;
  amount: number;
  currency: string;
  status: 'paid' | 'open' | 'void';
  created_at: string;
  pdf_url?: string;
}

export interface DashboardStats {
  total_licenses: number;
  active_instances: number;
  events_processed_24h: number;
  data_transferred_24h: number;
}

export interface UsageData {
  events: { date: string; count: number }[];
  data_transfer: { date: string; bytes: number }[];
}

export interface Instance {
  id: string;
  license_id: string;
  hostname: string;
  version: string;
  status: 'online' | 'offline';
  last_seen_at: string;
  events_processed: number;
}

export interface Download {
  product: string;
  version: string;
  editions: string[];
  platforms: string[];
  release_date: string;
  changelog_url?: string;
}
