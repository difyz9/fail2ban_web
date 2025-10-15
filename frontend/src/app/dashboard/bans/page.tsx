'use client';

import DashboardLayout from '@/components/layout/DashboardLayout';
import { Ban } from 'lucide-react';

export default function BansPage() {
  return (
    <DashboardLayout>
      <div className="min-h-screen">
        <div className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-10">
          <div className="px-6 py-4 lg:px-8">
            <h1 className="text-2xl lg:text-3xl font-bold text-gray-900 flex items-center">
              <Ban className="w-7 h-7 lg:w-8 lg:h-8 text-blue-600 mr-3" />
              禁令历史
            </h1>
            <p className="text-gray-600 mt-1 text-sm lg:text-base">查看历史封禁记录</p>
          </div>
        </div>
        <div className="px-6 py-8 lg:px-8">
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-12 text-center">
            <Ban className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">功能开发中</h3>
            <p className="text-gray-600">此功能正在开发中，敬请期待...</p>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
