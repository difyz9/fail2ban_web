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
  private readonly TOKEN_EXPIRES = 7; // 7天

  // 登录
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
      
      // 保存认证信息
      this.saveAuthData({
        token: response.token,
        user: response.user,
      });
      
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
      const response = await apiClient.post<{ token: string }>('/auth/refresh');
      
      // 更新本地token
      Cookies.set(this.TOKEN_KEY, response.token, { 
        expires: this.TOKEN_EXPIRES,
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
  saveAuthData(authData: AuthData): void {
    Cookies.set(this.TOKEN_KEY, authData.token, { 
      expires: this.TOKEN_EXPIRES,
      sameSite: 'strict'
    });
    Cookies.set(this.USER_KEY, JSON.stringify(authData.user), { 
      expires: this.TOKEN_EXPIRES,
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