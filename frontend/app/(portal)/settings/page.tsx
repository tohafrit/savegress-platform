'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api } from '@/lib/api';
import { User, Lock, CheckCircle, XCircle } from 'lucide-react';

export default function SettingsPage() {
  const { user, refreshUser } = useAuth();
  const [activeTab, setActiveTab] = useState('profile');

  const tabs = [
    { id: 'profile', name: 'Profile', icon: User },
    { id: 'security', name: 'Security', icon: Lock },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-h4 text-white">Settings</h1>
        <p className="text-content-1 text-grey">Manage your account settings</p>
      </div>

      <div className="card-dark overflow-hidden">
        {/* Tabs */}
        <div className="border-b border-cyan-40">
          <nav className="flex -mb-px">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-5 py-4 border-b-2 text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'border-accent-cyan text-accent-cyan'
                    : 'border-transparent text-grey hover:text-white hover:border-cyan-40'
                }`}
              >
                <tab.icon className="w-4 h-4" />
                {tab.name}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="p-6">
          {activeTab === 'profile' && <ProfileSettings user={user} onUpdate={refreshUser} />}
          {activeTab === 'security' && <SecuritySettings />}
        </div>
      </div>
    </div>
  );
}

function ProfileSettings({
  user,
  onUpdate,
}: {
  user: ReturnType<typeof useAuth>['user'];
  onUpdate: () => void;
}) {
  const [name, setName] = useState(user?.name || '');
  const [company, setCompany] = useState(user?.company || '');
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setSuccess(false);

    const { error } = await api.updateProfile({ name, company });

    if (error) {
      setError(error);
    } else {
      setSuccess(true);
      onUpdate();
    }
    setIsLoading(false);
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-md space-y-4">
      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center gap-2 text-red-400 text-sm">
          <XCircle className="w-4 h-4" />
          {error}
        </div>
      )}
      {success && (
        <div className="p-3 bg-accent-cyan/10 border border-accent-cyan/30 rounded-lg flex items-center gap-2 text-accent-cyan text-sm">
          <CheckCircle className="w-4 h-4" />
          Profile updated successfully
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          Email
        </label>
        <input
          type="email"
          value={user?.email || ''}
          disabled
          className="input-field bg-primary-dark/50 text-grey cursor-not-allowed"
        />
        <p className="text-xs text-grey mt-1">Email cannot be changed</p>
      </div>

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          Full name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="input-field"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          Company
        </label>
        <input
          type="text"
          value={company}
          onChange={(e) => setCompany(e.target.value)}
          className="input-field"
          placeholder="Optional"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
      >
        {isLoading ? 'Saving...' : 'Save Changes'}
      </button>
    </form>
  );
}

function SecuritySettings() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess(false);

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setIsLoading(true);
    const { error } = await api.changePassword(currentPassword, newPassword);

    if (error) {
      setError(error);
    } else {
      setSuccess(true);
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    }
    setIsLoading(false);
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-md space-y-4">
      <h3 className="text-h5 text-white">Change Password</h3>

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center gap-2 text-red-400 text-sm">
          <XCircle className="w-4 h-4" />
          {error}
        </div>
      )}
      {success && (
        <div className="p-3 bg-accent-cyan/10 border border-accent-cyan/30 rounded-lg flex items-center gap-2 text-accent-cyan text-sm">
          <CheckCircle className="w-4 h-4" />
          Password changed successfully
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          Current password
        </label>
        <input
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          required
          className="input-field"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          New password
        </label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          required
          className="input-field"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-grey mb-2">
          Confirm new password
        </label>
        <input
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          required
          className="input-field"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
      >
        {isLoading ? 'Changing...' : 'Change Password'}
      </button>
    </form>
  );
}
