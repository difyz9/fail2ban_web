'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuth } from '@/contexts/AuthContext';
import {
  LayoutDashboard,
  Shield,
  Ban,
  FileText,
  Settings,
  LogOut,
  ChevronLeft,
  User,
  Activity,
  Globe,
  Database,
  AlertCircle,
  Menu,
  X
} from 'lucide-react';

interface NavItem {
  name: string;
  href: string;
  icon: React.ElementType;
  badge?: number;
}

const navigationItems: NavItem[] = [
  { name: '仪表板', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Jails 管理', href: '/dashboard/jails', icon: Shield },
  { name: 'IP 管理', href: '/dashboard/ips', icon: Globe },
  { name: '禁令历史', href: '/dashboard/bans', icon: Ban },
  { name: '活动日志', href: '/dashboard/logs', icon: FileText },
  { name: '威胁分析', href: '/dashboard/threats', icon: AlertCircle },
  { name: '数据库', href: '/dashboard/database', icon: Database },
  { name: '系统设置', href: '/dashboard/settings', icon: Settings },
];

export default function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
  };

  const SidebarContent = () => (
    <div className="h-full flex flex-col bg-gradient-to-b from-slate-900 to-slate-800 text-white">
      {/* Logo & Brand */}
      <div className="p-6 border-b border-slate-700">
        <Link href="/dashboard" className="flex items-center space-x-3">
          <div className="bg-blue-600 p-2 rounded-lg">
            <Shield className="w-6 h-6" />
          </div>
          {!collapsed && (
            <div>
              <h1 className="text-xl font-bold">Fail2Ban</h1>
              <p className="text-xs text-slate-400">Web管理平台</p>
            </div>
          )}
        </Link>
      </div>


      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto p-4 space-y-1">
        {navigationItems.map((item) => {
          const isActive = pathname === item.href;
          const Icon = item.icon;

          return (
            <Link key={item.name} href={item.href}>
              <motion.div
                className={`
                  flex items-center space-x-3 px-4 py-3 rounded-lg cursor-pointer
                  transition-all duration-200
                  ${isActive 
                    ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/50' 
                    : 'text-slate-300 hover:bg-slate-700/50 hover:text-white'
                  }
                  ${collapsed ? 'justify-center' : ''}
                `}
                whileHover={{ x: collapsed ? 0 : 4 }}
                whileTap={{ scale: 0.95 }}
              >
                <Icon className={`w-5 h-5 flex-shrink-0 ${isActive ? 'text-white' : ''}`} />
                {!collapsed && (
                  <>
                    <span className="flex-1 text-sm font-medium">{item.name}</span>
                    {item.badge && (
                      <span className="bg-red-500 text-white text-xs px-2 py-0.5 rounded-full">
                        {item.badge}
                      </span>
                    )}
                  </>
                )}
              </motion.div>
            </Link>
          );
        })}
      </nav>

      {/* System Status */}
      {!collapsed && (
        <div className="p-4 border-t border-slate-700">
          <div className="bg-slate-700/50 rounded-lg p-3 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-400">系统状态</span>
              <span className="flex items-center text-green-400">
                <Activity className="w-3 h-3 mr-1" />
                正常
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-400">活跃监狱</span>
              <span className="text-white font-medium">12</span>
            </div>
          </div>
        </div>
      )}

      {/* Logout Button */}
      <div className="p-4 border-t border-slate-700">
        <motion.button
          onClick={handleLogout}
          className={`
            w-full flex items-center space-x-3 px-4 py-3 rounded-lg
            bg-red-600/10 hover:bg-red-600 text-red-400 hover:text-white
            transition-all duration-200
            ${collapsed ? 'justify-center' : ''}
          `}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          <LogOut className="w-5 h-5" />
          {!collapsed && <span className="text-sm font-medium">退出登录</span>}
        </motion.button>
      </div>

      {/* Collapse Toggle - Desktop */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="hidden lg:flex absolute -right-3 top-20 bg-slate-800 border border-slate-700 rounded-full p-1.5 hover:bg-slate-700 transition-colors"
      >
        <ChevronLeft className={`w-4 h-4 transition-transform ${collapsed ? 'rotate-180' : ''}`} />
      </button>
    </div>
  );

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={() => setMobileOpen(!mobileOpen)}
        className="lg:hidden fixed top-4 left-4 z-50 bg-slate-900 text-white p-2 rounded-lg shadow-lg"
      >
        {mobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
      </button>

      {/* Desktop Sidebar */}
      <aside
        className={`
          hidden lg:block fixed left-0 top-0 h-screen
          transition-all duration-300 ease-in-out
          ${collapsed ? 'w-20' : 'w-72'}
          shadow-xl z-40
        `}
      >
        <SidebarContent />
      </aside>

      {/* Mobile Sidebar */}
      <AnimatePresence>
        {mobileOpen && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="lg:hidden fixed inset-0 bg-black/50 z-40"
              onClick={() => setMobileOpen(false)}
            />

            {/* Sidebar */}
            <motion.aside
              initial={{ x: -300 }}
              animate={{ x: 0 }}
              exit={{ x: -300 }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              className="lg:hidden fixed left-0 top-0 h-screen w-72 z-50 shadow-2xl"
            >
              <SidebarContent />
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Spacer for desktop */}
      <div className={`hidden lg:block ${collapsed ? 'w-20' : 'w-72'} flex-shrink-0 transition-all duration-300`} />
    </>
  );
}
