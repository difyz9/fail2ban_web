import apiClient from './apiClient';
import Cookies from 'js-cookie';
import { LoginRequest, LoginResponse, User } from '@/types/api';

interface AuthData {
  token: string;
  user: User;
}

class AuthService {
  private readonly TOKEN_KEY = 'auth_token';
  private readonly USER_KEY = 'user_info';
  private readonly REMEMBER_KEY = 'remember_me';
  private readonly TOKEN_EXPIRES_SHORT = 1; // 1天（不记住我）
  private readonly TOKEN_EXPIRES_LONG = 30; // 30天（记住我）

  // 登录
  async login(credentials: LoginRequest, rememberMe: boolean = false): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
      
      // 保存认证信息
      this.saveAuthData({
        token: response.token,
        user: response.user,
      }, rememberMe);
      
      return response;
    } catch (error) {
      throw error;
    }
  }

  // 登出
  async logout(): Promise<void> {
    try {
      // 调用后端登出接口
      await apiClient.post('/auth/logout');
    } catch (error) {
      console.error('Logout API call failed:', error);
      // 即使API调用失败，也要清除本地数据
    } finally {
      this.clearAuthData();
    }
  }

  // 获取用户信息
  async getProfile(): Promise<User> {
    return await apiClient.get<User>('/auth/profile');
  }

  // 刷新Token
  async refreshToken(): Promise<boolean> {
    try {
      const response = await apiClient.post<{ token: string; expires_at: number }>('/auth/refresh');
      
      // 更新本地token（保持原有的过期时间设置）
      const rememberMe = Cookies.get(this.REMEMBER_KEY) === 'true';
      const expires = rememberMe ? this.TOKEN_EXPIRES_LONG : this.TOKEN_EXPIRES_SHORT;
      
      Cookies.set(this.TOKEN_KEY, response.token, { 
        expires,
        sameSite: 'strict'
      });
      
      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      this.clearAuthData();
      return false;
    }
  }

  // 修改密码
  async changePassword(oldPassword: string, newPassword: string): Promise<void> {
    await apiClient.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
  }

  // 保存认证数据
  saveAuthData(authData: AuthData, rememberMe: boolean = false): void {
    const expires = rememberMe ? this.TOKEN_EXPIRES_LONG : this.TOKEN_EXPIRES_SHORT;
    
    Cookies.set(this.TOKEN_KEY, authData.token, { 
      expires,
      sameSite: 'strict'
    });
    Cookies.set(this.USER_KEY, JSON.stringify(authData.user), { 
      expires,
      sameSite: 'strict'
    });
    Cookies.set(this.REMEMBER_KEY, String(rememberMe), {
      expires,
      sameSite: 'strict'
    });
  }

  // 获取认证数据
  getAuthData(): AuthData | null {
    const token = Cookies.get(this.TOKEN_KEY);
    const userStr = Cookies.get(this.USER_KEY);
    
    if (!token || !userStr) {
      return null;
    }
    
    try {
      const user = JSON.parse(userStr);
      return { token, user };
    } catch (error) {
      console.error('Failed to parse user data:', error);
      this.clearAuthData();
      return null;
    }
  }

  // 清除认证数据
  clearAuthData(): void {
    Cookies.remove(this.TOKEN_KEY);
    Cookies.remove(this.USER_KEY);
    Cookies.remove(this.REMEMBER_KEY);
  }

  // 检查是否已登录
  isAuthenticated(): boolean {
    const authData = this.getAuthData();
    return authData !== null;
  }

  // 获取当前用户
  getCurrentUser(): User | null {
    const authData = this.getAuthData();
    return authData?.user || null;
  }

  // 获取当前Token
  getToken(): string | null {
    return Cookies.get(this.TOKEN_KEY) || null;
  }

  // 自动刷新Token
  async autoRefreshToken(): Promise<boolean> {
    const token = this.getToken();
    if (!token) {
      return false;
    }

    try {
      // 检查token是否即将过期（这里简化处理）
      const response = await apiClient.get('/auth/verify');
      return true;
    } catch (error) {
      // Token可能已过期，尝试刷新
      return await this.refreshToken();
    }
  }
}

// 创建全局实例
const authService = new AuthService();

export default authService;