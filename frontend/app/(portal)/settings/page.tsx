'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api } from '@/lib/api';
import {
  PageHeader,
  InfoBanner,
  FormFieldWithHelp,
  HelpIcon,
  ExpandableSection,
} from '@/components/ui/helpers';
import {
  User,
  Lock,
  CheckCircle,
  XCircle,
  Shield,
  Eye,
  EyeOff,
  Info,
  AlertTriangle,
  Key,
  Mail,
  Building,
  UserCircle,
} from 'lucide-react';

export default function SettingsPage() {
  const { user, refreshUser } = useAuth();
  const [activeTab, setActiveTab] = useState('profile');

  const tabs = [
    { id: 'profile', name: 'Profile', icon: User, description: 'Your personal information' },
    { id: 'security', name: 'Security', icon: Lock, description: 'Password and authentication' },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Manage your account settings, profile information, and security preferences. Your data is encrypted and stored securely."
        tip="Keep your profile up to date and use a strong password to protect your account."
      />

      <div className="card-dark overflow-hidden">
        {/* Tabs with descriptions */}
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
    <div className="max-w-xl space-y-6">
      <div className="space-y-2">
        <h3 className="text-h5 text-white flex items-center gap-2">
          <UserCircle className="w-5 h-5 text-accent-cyan" />
          Profile Information
        </h3>
        <p className="text-sm text-grey">
          Update your personal information. This is how you&apos;ll appear in the Savegress dashboard.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <InfoBanner type="warning" title="Error updating profile">
            {error}
          </InfoBanner>
        )}
        {success && (
          <InfoBanner type="success" title="Profile updated!">
            Your changes have been saved successfully.
          </InfoBanner>
        )}

        <FormFieldWithHelp
          label="Email Address"
          help="Your email is used for login and notifications"
        >
          <div className="relative">
            <Mail className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="email"
              value={user?.email || ''}
              disabled
              className="input-field pl-10 bg-primary-dark/50 text-grey cursor-not-allowed"
            />
          </div>
          <p className="text-xs text-grey mt-1.5 flex items-center gap-1">
            <Info className="w-3 h-3" />
            Email cannot be changed. Contact support if you need to update it.
          </p>
        </FormFieldWithHelp>

        <FormFieldWithHelp
          label="Full Name"
          help="Your display name in the dashboard"
          tip="Use your real name so teammates can identify you"
        >
          <div className="relative">
            <User className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input-field pl-10"
              placeholder="John Doe"
            />
          </div>
        </FormFieldWithHelp>

        <FormFieldWithHelp
          label="Company"
          help="Your organization name (optional)"
          tip="Helps us understand your use case better"
        >
          <div className="relative">
            <Building className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
              className="input-field pl-10"
              placeholder="Acme Inc."
            />
          </div>
        </FormFieldWithHelp>

        <div className="pt-4">
          <button
            type="submit"
            disabled={isLoading}
            className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
          >
            {isLoading ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
}

function SecuritySettings() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  // Password strength indicator
  const getPasswordStrength = (password: string) => {
    if (!password) return { strength: 0, label: '', color: '' };
    let strength = 0;
    if (password.length >= 8) strength++;
    if (password.length >= 12) strength++;
    if (/[A-Z]/.test(password)) strength++;
    if (/[0-9]/.test(password)) strength++;
    if (/[^A-Za-z0-9]/.test(password)) strength++;

    if (strength <= 2) return { strength, label: 'Weak', color: 'text-red-400 bg-red-400' };
    if (strength <= 3) return { strength, label: 'Medium', color: 'text-accent-orange bg-accent-orange' };
    return { strength, label: 'Strong', color: 'text-green-400 bg-green-400' };
  };

  const passwordStrength = getPasswordStrength(newPassword);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess(false);

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match. Please make sure both password fields are identical.');
      return;
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters long for security.');
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
    <div className="max-w-xl space-y-6">
      <div className="space-y-2">
        <h3 className="text-h5 text-white flex items-center gap-2">
          <Key className="w-5 h-5 text-accent-cyan" />
          Change Password
        </h3>
        <p className="text-sm text-grey">
          Update your password to keep your account secure. We recommend using a unique password
          that you don&apos;t use anywhere else.
        </p>
      </div>

      <InfoBanner type="tip" title="Password security tips" dismissible>
        <ul className="list-disc list-inside text-sm space-y-1">
          <li>Use at least 12 characters for better security</li>
          <li>Mix uppercase, lowercase, numbers, and symbols</li>
          <li>Avoid common words or personal information</li>
          <li>Consider using a password manager</li>
        </ul>
      </InfoBanner>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <InfoBanner type="warning" title="Password change failed">
            {error}
          </InfoBanner>
        )}
        {success && (
          <InfoBanner type="success" title="Password changed!">
            Your password has been updated successfully. Use your new password next time you log in.
          </InfoBanner>
        )}

        <FormFieldWithHelp
          label="Current Password"
          help="Enter your current password to verify your identity"
          required
        >
          <div className="relative">
            <Lock className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type={showCurrentPassword ? 'text' : 'password'}
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
              className="input-field pl-10 pr-10"
              placeholder="Enter current password"
            />
            <button
              type="button"
              onClick={() => setShowCurrentPassword(!showCurrentPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-grey hover:text-white transition-colors"
              title={showCurrentPassword ? 'Hide password' : 'Show password'}
            >
              {showCurrentPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </FormFieldWithHelp>

        <FormFieldWithHelp
          label="New Password"
          help="Choose a strong password with at least 8 characters"
          required
        >
          <div className="relative">
            <Key className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type={showNewPassword ? 'text' : 'password'}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              className="input-field pl-10 pr-10"
              placeholder="Enter new password"
            />
            <button
              type="button"
              onClick={() => setShowNewPassword(!showNewPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-grey hover:text-white transition-colors"
              title={showNewPassword ? 'Hide password' : 'Show password'}
            >
              {showNewPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>

          {/* Password strength indicator */}
          {newPassword && (
            <div className="mt-2 space-y-2">
              <div className="flex items-center gap-2">
                <div className="flex-1 h-2 bg-primary-dark rounded-full overflow-hidden">
                  <div
                    className={`h-full ${passwordStrength.color.split(' ')[1]} transition-all`}
                    style={{ width: `${(passwordStrength.strength / 5) * 100}%` }}
                  />
                </div>
                <span className={`text-xs font-medium ${passwordStrength.color.split(' ')[0]}`}>
                  {passwordStrength.label}
                </span>
              </div>
              <ul className="text-xs text-grey space-y-0.5">
                <li className={newPassword.length >= 8 ? 'text-green-400' : ''}>
                  {newPassword.length >= 8 ? '✓' : '○'} At least 8 characters
                </li>
                <li className={/[A-Z]/.test(newPassword) ? 'text-green-400' : ''}>
                  {/[A-Z]/.test(newPassword) ? '✓' : '○'} Uppercase letter
                </li>
                <li className={/[0-9]/.test(newPassword) ? 'text-green-400' : ''}>
                  {/[0-9]/.test(newPassword) ? '✓' : '○'} Number
                </li>
                <li className={/[^A-Za-z0-9]/.test(newPassword) ? 'text-green-400' : ''}>
                  {/[^A-Za-z0-9]/.test(newPassword) ? '✓' : '○'} Special character
                </li>
              </ul>
            </div>
          )}
        </FormFieldWithHelp>

        <FormFieldWithHelp
          label="Confirm New Password"
          help="Re-enter your new password to confirm"
          required
        >
          <div className="relative">
            <Key className="w-4 h-4 text-grey absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              className="input-field pl-10"
              placeholder="Confirm new password"
            />
          </div>
          {confirmPassword && newPassword !== confirmPassword && (
            <p className="text-xs text-red-400 mt-1.5 flex items-center gap-1">
              <AlertTriangle className="w-3 h-3" />
              Passwords don&apos;t match
            </p>
          )}
          {confirmPassword && newPassword === confirmPassword && (
            <p className="text-xs text-green-400 mt-1.5 flex items-center gap-1">
              <CheckCircle className="w-3 h-3" />
              Passwords match
            </p>
          )}
        </FormFieldWithHelp>

        <div className="pt-4">
          <button
            type="submit"
            disabled={isLoading || !currentPassword || !newPassword || !confirmPassword || newPassword !== confirmPassword}
            className="btn-primary px-5 py-2.5 text-sm disabled:opacity-50"
          >
            {isLoading ? 'Changing Password...' : 'Change Password'}
          </button>
        </div>
      </form>

      {/* Additional security info */}
      <ExpandableSection title="Security information" icon={Shield}>
        <div className="space-y-3 text-sm text-grey">
          <p>
            Your password is stored using industry-standard bcrypt hashing.
            We never store your actual password - only a secure hash.
          </p>
          <p>
            <strong className="text-white">Session security:</strong> Your session expires after 24 hours
            of inactivity. You can log out from all devices by changing your password.
          </p>
          <p>
            <strong className="text-white">Two-factor authentication:</strong> Coming soon!
            We&apos;re working on adding 2FA support for additional security.
          </p>
        </div>
      </ExpandableSection>
    </div>
  );
}
