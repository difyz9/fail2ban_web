import axios, { AxiosInstance, AxiosResponse } from 'axios';
import Cookies from 'js-cookie';
import { ApiResponse } from '@/types/api';

class ApiClient {
  private client: AxiosInstance;
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // 请求拦截器 - 添加认证token
    this.client.interceptors.request.use(
      (config) => {
        const token = Cookies.get('auth_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器 - 处理统一错误
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        return response;
      },
      (error) => {
        if (error.response?.status === 401) {
          // Token过期，清除本地存储并跳转到登录页
          Cookies.remove('auth_token');
          Cookies.remove('user_info');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
        }
        
        const errorMessage = error.response?.data?.error || 
                           error.response?.data?.message || 
                           error.message || 
                           '网络请求失败';
        
        return Promise.reject(new Error(errorMessage));
      }
    );
  }

  // GET 请求
  async get<T = any>(url: string, params?: any): Promise<T> {
    const response = await this.client.get<ApiResponse<T>>(url, { params });
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '请求失败');
    }
  }

  // POST 请求
  async post<T = any>(url: string, data?: any): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, data);
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '请求失败');
    }
  }

  // PUT 请求
  async put<T = any>(url: string, data?: any): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(url, data);
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '请求失败');
    }
  }

  // DELETE 请求
  async delete<T = any>(url: string): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(url);
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '请求失败');
    }
  }

  // PATCH 请求
  async patch<T = any>(url: string, data?: any): Promise<T> {
    const response = await this.client.patch<ApiResponse<T>>(url, data);
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '请求失败');
    }
  }

  // 文件上传
  async upload<T = any>(url: string, formData: FormData): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.error || response.data.message || '上传失败');
    }
  }

  // 获取基础URL
  getBaseURL(): string {
    return this.baseURL;
  }
}

// 创建全局实例
const apiClient = new ApiClient();

export default apiClient;