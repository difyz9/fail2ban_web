import apiClient from './apiClient';
import { 
  SystemStats, 
  BannedIP, 
  JailConfig, 
  LogEntry, 
  WhitelistEntry, 
  ThreatAnalysis,
  PaginatedResponse,
  QueryParams 
} from '@/types/api';

// 统计服务
export const statsService = {
  // 获取系统统计信息
  getSystemStats: (): Promise<SystemStats> => 
    apiClient.get('/api/stats'),

  // 获取今日统计
  getTodayStats: (): Promise<any> => 
    apiClient.get('/api/stats/today'),

  // 获取历史统计数据
  getHistoryStats: (days: number = 7): Promise<any> => 
    apiClient.get(`/api/stats/history?days=${days}`),
};

// IP管理服务
export const ipService = {
  // 获取被禁IP列表
  getBannedIPs: (params?: QueryParams): Promise<PaginatedResponse<BannedIP>> => 
    apiClient.get('/api/banned-ips', params),

  // 解禁IP
  unbanIP: (ip: string): Promise<void> => 
    apiClient.post(`/api/banned-ips/${ip}/unban`),

  // 手动封禁IP
  banIP: (ip: string, jail: string, banTime?: number): Promise<void> => 
    apiClient.post('/api/banned-ips/ban', { ip, jail, ban_time: banTime }),

  // 获取IP详细信息
  getIPDetails: (ip: string): Promise<BannedIP> => 
    apiClient.get(`/api/banned-ips/${ip}`),

  // 批量解禁IP
  batchUnban: (ips: string[]): Promise<void> => 
    apiClient.post('/api/banned-ips/batch-unban', { ips }),
};

// Jail管理服务
export const jailService = {
  // 获取所有Jail配置
  getJails: (): Promise<JailConfig[]> => 
    apiClient.get('/api/jails'),

  // 获取单个Jail配置
  getJail: (name: string): Promise<JailConfig> => 
    apiClient.get(`/api/jails/${name}`),

  // 更新Jail配置
  updateJail: (name: string, config: Partial<JailConfig>): Promise<void> => 
    apiClient.put(`/api/jails/${name}`, config),

  // 启用/禁用Jail
  toggleJail: (name: string, enabled: boolean): Promise<void> => 
    apiClient.post(`/api/jails/${name}/toggle`, { enabled }),

  // 重启Jail
  restartJail: (name: string): Promise<void> => 
    apiClient.post(`/api/jails/${name}/restart`),

  // 获取Jail状态
  getJailStatus: (name: string): Promise<any> => 
    apiClient.get(`/api/jails/${name}/status`),
};

// 日志服务
export const logService = {
  // 获取日志列表
  getLogs: (params?: QueryParams): Promise<PaginatedResponse<LogEntry>> => 
    apiClient.get('/api/logs', params),

  // 获取实时日志
  getRealtimeLogs: (jail?: string): Promise<LogEntry[]> => 
    apiClient.get('/api/logs/realtime', { jail }),

  // 下载日志文件
  downloadLogs: (startDate: string, endDate: string): Promise<Blob> => 
    apiClient.get('/api/logs/download', { start_date: startDate, end_date: endDate }),

  // 清理日志
  clearLogs: (beforeDate: string): Promise<void> => 
    apiClient.delete(`/api/logs?before_date=${beforeDate}`),
};

// 白名单服务
export const whitelistService = {
  // 获取白名单
  getWhitelist: (params?: QueryParams): Promise<PaginatedResponse<WhitelistEntry>> => 
    apiClient.get('/api/whitelist', params),

  // 添加白名单条目
  addToWhitelist: (ip: string, description?: string): Promise<WhitelistEntry> => 
    apiClient.post('/api/whitelist', { ip, description }),

  // 删除白名单条目
  removeFromWhitelist: (id: number): Promise<void> => 
    apiClient.delete(`/api/whitelist/${id}`),

  // 更新白名单条目
  updateWhitelistEntry: (id: number, data: Partial<WhitelistEntry>): Promise<WhitelistEntry> => 
    apiClient.put(`/api/whitelist/${id}`, data),

  // 批量添加白名单
  batchAddToWhitelist: (entries: Array<{ ip: string; description?: string }>): Promise<void> => 
    apiClient.post('/api/whitelist/batch', { entries }),
};

// 智能分析服务
export const analysisService = {
  // 获取威胁分析
  getThreatAnalysis: (ip: string): Promise<ThreatAnalysis> => 
    apiClient.get(`/api/analysis/threat/${ip}`),

  // 获取攻击趋势
  getAttackTrends: (period: string = '24h'): Promise<any> => 
    apiClient.get(`/api/analysis/trends?period=${period}`),

  // 获取地理位置统计
  getGeoStats: (): Promise<any> => 
    apiClient.get('/api/analysis/geo-stats'),

  // 获取攻击类型分析
  getAttackTypeAnalysis: (): Promise<any> => 
    apiClient.get('/api/analysis/attack-types'),

  // 生成安全报告
  generateSecurityReport: (startDate: string, endDate: string): Promise<any> => 
    apiClient.post('/api/analysis/security-report', { start_date: startDate, end_date: endDate }),
};

// 系统服务
export const systemService = {
  // 获取系统信息
  getSystemInfo: (): Promise<any> => 
    apiClient.get('/api/system/info'),

  // 获取系统配置
  getSystemConfig: (): Promise<any> => 
    apiClient.get('/api/system/config'),

  // 更新系统配置
  updateSystemConfig: (config: any): Promise<void> => 
    apiClient.put('/api/system/config', config),

  // 重启Fail2ban服务
  restartService: (): Promise<void> => 
    apiClient.post('/api/system/restart'),

  // 获取服务状态
  getServiceStatus: (): Promise<any> => 
    apiClient.get('/api/system/status'),

  // 测试配置
  testConfig: (): Promise<any> => 
    apiClient.post('/api/system/test-config'),
};