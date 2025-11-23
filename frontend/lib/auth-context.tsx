'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { api, User } from './api';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<{ error?: string }>;
  register: (email: string, password: string, name: string) => Promise<{ error?: string }>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    const token = api.getToken();
    if (!token) {
      setUser(null);
      setIsLoading(false);
      return;
    }

    const { data, error } = await api.getProfile();
    if (error) {
      api.setToken(null);
      setUser(null);
    } else if (data) {
      setUser(data);
    }
    setIsLoading(false);
  }, []);

  useEffect(() => {
    refreshUser();
  }, [refreshUser]);

  const login = async (email: string, password: string) => {
    const { data, error } = await api.login(email, password);
    if (error) {
      return { error };
    }
    if (data) {
      api.setToken(data.tokens.access_token);
      if (data.tokens.refresh_token) {
        localStorage.setItem('refresh_token', data.tokens.refresh_token);
      }
      setUser(data.user);
    }
    return {};
  };

  const register = async (email: string, password: string, name: string) => {
    const { data, error } = await api.register(email, password, name);
    if (error) {
      return { error };
    }
    if (data) {
      api.setToken(data.tokens.access_token);
      if (data.tokens.refresh_token) {
        localStorage.setItem('refresh_token', data.tokens.refresh_token);
      }
      setUser(data.user);
    }
    return {};
  };

  const logout = () => {
    api.setToken(null);
    localStorage.removeItem('refresh_token');
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
