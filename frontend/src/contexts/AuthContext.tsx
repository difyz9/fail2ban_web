'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import Cookies from 'js-cookie';
import authService from '@/services/authService';
import type { User, LoginRequest } from '@/types/api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginRequest, rememberMe?: boolean) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  // 初始化时获取用户信息
  useEffect(() => {
    const initAuth = async () => {
      const token = Cookies.get('auth_token');
      console.log('[AuthContext] Initializing auth, token exists:', !!token);
      
      if (token) {
        try {
          // 先从本地 Cookie 恢复用户信息（避免闪烁）
          const cachedUser = authService.getCurrentUser();
          console.log('[AuthContext] Cached user:', cachedUser);
          
          if (cachedUser) {
            setUser(cachedUser);
            console.log('[AuthContext] Restored user from cache');
          }
          
          // 然后从服务器验证并更新用户信息
          console.log('[AuthContext] Fetching user profile from server...');
          const userData = await authService.getProfile();
          console.log('[AuthContext] User profile fetched:', userData);
          setUser(userData);
        } catch (error) {
          console.error('[AuthContext] Failed to get user profile:', error);
          authService.clearAuthData();
          setUser(null);
        }
      } else {
        console.log('[AuthContext] No token found, user not logged in');
      }
      setLoading(false);
    };

    // 添加超时保护，确保 loading 状态不会永久卡住
    const timeout = setTimeout(() => {
      console.log('[AuthContext] Timeout reached, setting loading to false');
      setLoading(false);
    }, 3000); // 3秒超时

    initAuth().finally(() => {
      clearTimeout(timeout);
    });

    return () => clearTimeout(timeout);
  }, []);

  // 自动刷新 token
  useEffect(() => {
    if (user) {
      const interval = setInterval(async () => {
        try {
          await authService.refreshToken();
        } catch (error) {
          console.error('Failed to refresh token:', error);
          await logout();
        }
      }, 15 * 60 * 1000); // 15分钟刷新一次

      return () => clearInterval(interval);
    }
  }, [user]);

  const login = async (credentials: LoginRequest, rememberMe: boolean = false) => {
    const response = await authService.login(credentials, rememberMe);
    setUser(response.user);
    
    // 确保用户信息已保存到 Cookies
    authService.saveAuthData({
      token: response.token,
      user: response.user,
    }, rememberMe);
    
    router.push('/dashboard');
  };

  const logout = async () => {
    try {
      await authService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      router.push('/login');
    }
  };

  const refreshUser = async () => {
    const userData = await authService.getProfile();
    setUser(userData);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}