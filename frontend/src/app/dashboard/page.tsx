'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import DashboardLayout from '@/components/layout/DashboardLayout';
import { 
  Shield, 
  Activity,
  AlertTriangle,
  Server,
  Eye,
  Globe,
  Clock,
  TrendingUp,
  TrendingDown,
  RefreshCw,
  CheckCircle,
  Info,
  MapPin,
  Ban,
  Unlock,
  Settings
} from 'lucide-react';

// 模拟数据
const mockStats = {
  totalBans: 1247,
  activeBans: 89,
  blockedAttempts: 15634,
  jailsActive: 12,
  trends: {
    bans: 12.5,
    attempts: -8.3,
    jails: 0
  }
};

const mockActivities = [
  {
    id: 1,
    type: 'ban',
    ip: '192.168.1.100',
    jail: 'ssh',
    time: '2分钟前',
    country: '中国',
    reason: 'SSH 暴力破解尝试',
    severity: 'high'
  },
  {
    id: 2,
    type: 'unban',
    ip: '10.0.0.50',
    jail: 'nginx',
    time: '5分钟前',
    country: '美国',
    reason: '禁令期满自动解除',
    severity: 'info'
  }
];

export default function Dashboard() {
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setLoading(false);
    }, 1000);
    return () => clearTimeout(timer);
  }, []);

  const handleRefresh = async () => {
    setRefreshing(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    setRefreshing(false);
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high': return 'text-red-600 bg-red-50';
      case 'medium': return 'text-orange-600 bg-orange-50';
      default: return 'text-blue-600 bg-blue-50';
    }
  };

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'ban': return <Ban className="w-4 h-4" />;
      case 'unban': return <Unlock className="w-4 h-4" />;
      case 'alert': return <AlertTriangle className="w-4 h-4" />;
      default: return <Info className="w-4 h-4" />;
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-gray-600 text-lg">加载仪表板数据...</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="min-h-screen">
        {/* 页面头部 */}
        <div className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-10">
          <div className="px-6 py-4 lg:px-8">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-gray-900 flex items-center">
                  <Shield className="w-7 h-7 lg:w-8 lg:h-8 text-blue-600 mr-3" />
                  仪表板
                </h1>
                <p className="text-gray-600 mt-1 text-sm lg:text-base">实时监控系统安全状态</p>
              </div>
              <motion.button
                onClick={handleRefresh}
                disabled={refreshing}
                className="flex items-center space-x-2 bg-blue-600 text-white px-3 py-2 lg:px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm lg:text-base"
                whileHover={{ scale: refreshing ? 1 : 1.05 }}
              >
                <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
                <span className="hidden sm:inline">{refreshing ? '刷新中...' : '刷新'}</span>
              </motion.button>
            </div>
          </div>
        </div>

        <div className="px-6 py-8 lg:px-8">
          {/* 核心指标卡片 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[
              {
                title: '总禁令数',
                value: mockStats.totalBans.toLocaleString(),
                icon: Ban,
                bgColor: 'bg-red-50',
                textColor: 'text-red-600',
                trend: mockStats.trends.bans
              },
              {
                title: '活跃禁令',
                value: mockStats.activeBans,
                icon: Shield,
                bgColor: 'bg-orange-50',
                textColor: 'text-orange-600',
                trend: mockStats.trends.attempts
              },
              {
                title: '阻止尝试',
                value: mockStats.blockedAttempts.toLocaleString(),
                icon: Eye,
                bgColor: 'bg-blue-50',
                textColor: 'text-blue-600',
                trend: mockStats.trends.bans
              },
              {
                title: '活跃监狱',
                value: mockStats.jailsActive,
                icon: Server,
                bgColor: 'bg-green-50',
                textColor: 'text-green-600',
                trend: mockStats.trends.jails
              }
            ].map((stat, index) => (
              <motion.div
                key={stat.title}
                className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-md transition-all duration-300"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.1 * index }}
                whileHover={{ y: -2 }}
              >
                <div className="flex items-center justify-between mb-4">
                  <div className={`p-3 rounded-xl ${stat.bgColor}`}>
                    <stat.icon className={`w-6 h-6 ${stat.textColor}`} />
                  </div>
                  <div className={`flex items-center text-sm ${stat.trend >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                    {stat.trend >= 0 ? <TrendingUp className="w-4 h-4 mr-1" /> : <TrendingDown className="w-4 h-4 mr-1" />}
                    {Math.abs(stat.trend)}%
                  </div>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 mb-1">{stat.value}</h3>
                <p className="text-gray-600 text-sm">{stat.title}</p>
              </motion.div>
            ))}
          </div>

          {/* 实时活动 */}
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 border-b border-gray-100">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-gray-900 flex items-center">
                  <Activity className="w-5 h-5 text-blue-600 mr-2" />
                  实时活动
                </h2>
                <span className="flex items-center text-sm text-gray-500">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></div>
                  实时更新
                </span>
              </div>
            </div>
            <div className="max-h-96 overflow-y-auto">
              {mockActivities.map((activity, index) => (
                <motion.div
                  key={activity.id}
                  className="p-4 border-b border-gray-50 last:border-b-0 hover:bg-gray-50 transition-colors"
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.4, delay: 0.1 * index }}
                >
                  <div className="flex items-start space-x-3">
                    <div className={`p-2 rounded-lg ${getSeverityColor(activity.severity)} flex-shrink-0`}>
                      {getActivityIcon(activity.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-gray-900">
                          {activity.reason}
                        </p>
                        <span className="text-xs text-gray-500">{activity.time}</span>
                      </div>
                      <div className="mt-1 flex items-center space-x-4 text-xs text-gray-500">
                        <span className="flex items-center">
                          <Globe className="w-3 h-3 mr-1" />
                          {activity.ip}
                        </span>
                        <span className="flex items-center">
                          <MapPin className="w-3 h-3 mr-1" />
                          {activity.country}
                        </span>
                        <span className="flex items-center">
                          <Server className="w-3 h-3 mr-1" />
                          {activity.jail}
                        </span>
                      </div>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}