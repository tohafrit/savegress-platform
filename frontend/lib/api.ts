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

  // Connections
  async getConnections() {
    return this.request<{ connections: Connection[] }>('/connections');
  }

  async getConnection(id: string) {
    return this.request<Connection>(`/connections/${id}`);
  }

  async createConnection(data: Partial<Connection>) {
    return this.request<Connection>('/connections', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateConnection(id: string, data: Partial<Connection>) {
    return this.request<Connection>(`/connections/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteConnection(id: string) {
    return this.request(`/connections/${id}`, {
      method: 'DELETE',
    });
  }

  async testConnection(id: string) {
    return this.request<{ success: boolean; message: string }>(`/connections/${id}/test`, {
      method: 'POST',
    });
  }

  async testConnectionDirect(data: Partial<Connection>) {
    return this.request<{ success: boolean; message: string }>('/connections/test', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Pipelines
  async getPipelines() {
    return this.request<{ pipelines: Pipeline[] }>('/pipelines');
  }

  async getPipeline(id: string) {
    return this.request<Pipeline>(`/pipelines/${id}`);
  }

  async createPipeline(data: Partial<Pipeline>) {
    return this.request<Pipeline>('/pipelines', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updatePipeline(id: string, data: Partial<Pipeline>) {
    return this.request<Pipeline>(`/pipelines/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deletePipeline(id: string) {
    return this.request(`/pipelines/${id}`, {
      method: 'DELETE',
    });
  }

  async getPipelineMetrics(id: string, hours: number = 24) {
    return this.request<{ metrics: PipelineMetric[] }>(`/pipelines/${id}/metrics?hours=${hours}`);
  }

  async getPipelineLogs(id: string, limit: number = 100, level?: string) {
    const params = new URLSearchParams({ limit: String(limit) });
    if (level) params.append('level', level);
    return this.request<{ logs: PipelineLog[] }>(`/pipelines/${id}/logs?${params}`);
  }

  // Config Generator
  async getConfigFormats() {
    return this.request<{ formats: ConfigFormat[] }>('/config/formats');
  }

  async generateConfig(format: string, pipelineId?: string, download?: boolean) {
    const params = new URLSearchParams({ format });
    if (pipelineId) params.append('pipeline_id', pipelineId);
    if (download) params.append('download', 'true');
    return this.request<string>(`/config/generate?${params}`);
  }

  async getQuickStart(sourceType: string = 'postgres') {
    return this.request<string>(`/config/quickstart?source_type=${sourceType}`);
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

export interface Connection {
  id: string;
  name: string;
  type: 'postgres' | 'mysql' | 'mongodb' | 'sqlserver' | 'oracle' | 'cassandra' | 'dynamodb';
  host: string;
  port: number;
  database: string;
  username: string;
  password?: string;
  ssl_mode: string;
  options?: Record<string, string>;
  last_tested_at?: string;
  test_status?: 'success' | 'failed' | 'pending';
  created_at: string;
  updated_at: string;
}

export interface Pipeline {
  id: string;
  name: string;
  description?: string;
  source_connection_id: string;
  target_connection_id?: string;
  target_type: string;
  target_config?: Record<string, string>;
  tables: string[];
  status: 'created' | 'running' | 'paused' | 'stopped' | 'error';
  license_id?: string;
  hardware_id?: string;
  events_processed: number;
  bytes_processed: number;
  current_lag_ms: number;
  last_event_at?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
  source_connection?: Connection;
  target_connection?: Connection;
}

export interface PipelineMetric {
  timestamp: string;
  events_per_second: number;
  bytes_per_second: number;
  latency_ms: number;
  errors: number;
}

export interface PipelineLog {
  id: string;
  pipeline_id: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  details?: Record<string, unknown>;
  timestamp: string;
}

export interface ConfigFormat {
  id: string;
  name: string;
  description: string;
  filename: string;
  icon: string;
}
