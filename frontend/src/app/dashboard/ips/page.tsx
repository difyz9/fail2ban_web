'use client';

import DashboardLayout from '@/components/layout/DashboardLayout';
import { Globe, Search, Filter } from 'lucide-react';

export default function IPsPage() {
  return (
    <DashboardLayout>
      <div className="min-h-screen">
        <div className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-10">
          <div className="px-6 py-4 lg:px-8">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-gray-900 flex items-center">
                  <Globe className="w-7 h-7 lg:w-8 lg:h-8 text-blue-600 mr-3" />
                  IP 管理
                </h1>
                <p className="text-gray-600 mt-1 text-sm lg:text-base">
                  查看和管理被禁止的 IP 地址
                </p>
              </div>
              <div className="flex items-center space-x-2">
                <button className="flex items-center space-x-2 bg-gray-100 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-200">
                  <Filter className="w-4 h-4" />
                  <span>筛选</span>
                </button>
                <button className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">
                  <Search className="w-4 h-4" />
                  <span>搜索</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-8 lg:px-8">
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-12 text-center">
            <Globe className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              IP 管理功能开发中
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
