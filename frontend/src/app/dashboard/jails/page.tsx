'use client';

import DashboardLayout from '@/components/layout/DashboardLayout';
import { Shield, Plus, Power, Settings } from 'lucide-react';

export default function JailsPage() {
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
                  Jails 管理
                </h1>
                <p className="text-gray-600 mt-1 text-sm lg:text-base">
                  配置和管理 Fail2Ban 监狱规则
                </p>
              </div>
              <button className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">
                <Plus className="w-4 h-4" />
                <span>新建 Jail</span>
              </button>
            </div>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 py-8 lg:px-8">
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-12 text-center">
            <Shield className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              Jails 管理功能开发中
            </h3>
            <p className="text-gray-600">
              此功能正在开发中，敬请期待...
            </p>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
