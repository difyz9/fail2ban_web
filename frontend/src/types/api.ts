// API 响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// 用户类型
export interface User {
  id: number;
  username: string;
  email?: string;
  role: string;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
}

// 登录响应
export interface LoginResponse {
  token: string;
  user: User;
  expires_at: number;  // 后端返回的是 expires_at（Unix 时间戳）
}

// 统计数据
export interface SystemStats {
  total_banned_ips: number;
  active_jails: number;
  failed_attempts_today: number;
  system_uptime: string;
  last_ban_time?: string;
}

// 被禁IP信息
export interface BannedIP {
  id: number;
  ip: string;
  jail: string;
  ban_time: string;
  unban_time?: string;
  attempts: number;
  country?: string;
  region?: string;
  is_active: boolean;
}

// Jail配置
export interface JailConfig {
  name: string;
  enabled: boolean;
  filter: string;
  logpath: string;
  maxretry: number;
  findtime: number;
  bantime: number;
  backend: string;
  action: string;
}

// 日志条目
export interface LogEntry {
  id: number;
  timestamp: string;
  level: 'INFO' | 'WARNING' | 'ERROR' | 'DEBUG';
  jail: string;
  message: string;
  ip?: string;
}

// 白名单条目
export interface WhitelistEntry {
  id: number;
  ip: string;
  description?: string;
  created_at: string;
  created_by: string;
}

// 分页响应
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

// 查询参数
export interface QueryParams {
  page?: number;
  per_page?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  filter?: Record<string, any>;
}

// 智能分析结果
export interface ThreatAnalysis {
  ip: string;
  risk_score: number;
  threat_types: string[];
  geo_location: {
    country: string;
    region: string;
    city: string;
  };
  historical_data: {
    total_attempts: number;
    first_seen: string;
    last_seen: string;
  };
  recommendations: string[];
}

// 系统配置
export interface SystemConfig {
  fail2ban_status: boolean;
  auto_ban_enabled: boolean;
  max_retry: number;
  ban_time: number;
  find_time: number;
  email_notifications: boolean;
  log_level: string;
}