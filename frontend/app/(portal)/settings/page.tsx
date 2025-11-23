'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api } from '@/lib/api';
import { User, Lock, Bell, Shield } from 'lucide-react';

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
        <h1 className="text-2xl font-bold text-primary">Settings</h1>
        <p className="text-neutral-dark-gray">Manage your account settings</p>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        {/* Tabs */}
        <div className="border-b border-gray-200">
          <nav className="flex -mb-px">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-3 border-b-2 text-sm font-medium ${
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-neutral-dark-gray hover:text-primary hover:border-gray-300'
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
        <div className="p-3 bg-red-50 border border-red-200 text-red-600 rounded-md text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="p-3 bg-green-50 border border-green-200 text-green-600 rounded-md text-sm">
          Profile updated successfully
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          Email
        </label>
        <input
          type="email"
          value={user?.email || ''}
          disabled
          className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-neutral-dark-gray"
        />
        <p className="text-xs text-neutral-dark-gray mt-1">Email cannot be changed</p>
      </div>

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          Full name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          Company
        </label>
        <input
          type="text"
          value={company}
          onChange={(e) => setCompany(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
          placeholder="Optional"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark disabled:opacity-50"
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
      <h3 className="text-lg font-medium text-primary">Change Password</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 text-red-600 rounded-md text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="p-3 bg-green-50 border border-green-200 text-green-600 rounded-md text-sm">
          Password changed successfully
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          Current password
        </label>
        <input
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          New password
        </label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-neutral-dark-gray mb-1">
          Confirm new password
        </label>
        <input
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark disabled:opacity-50"
      >
        {isLoading ? 'Changing...' : 'Change Password'}
      </button>
    </form>
  );
}
