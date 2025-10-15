'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '@/contexts/AuthContext';
import AuthGuard from '@/components/auth/AuthGuard';
import { 
  Shield, 
  User, 
  Lock, 
  AlertCircle, 
  Eye, 
  EyeOff, 
  CheckCircle2,
  KeyRound
} from 'lucide-react';

function LoginForm() {
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    rememberMe: false,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    
    try {
      await login({
        username: formData.username,
        password: formData.password,
      });
    } catch (err: any) {
      setError(err.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, type, checked, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 relative overflow-hidden">{/* 动态背景效果 */}
      <div className="absolute inset-0">
          <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHZpZXdCb3g9IjAgMCA0MCA0MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZGVmcz48cGF0dGVybiBpZD0iZ3JpZCIgd2lkdGg9IjQwIiBoZWlnaHQ9IjQwIiBwYXR0ZXJuVW5pdHM9InVzZXJTcGFjZU9uVXNlIj48cGF0aCBkPSJNIDQwIDAgTCAwIDAgMCA0MCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSJyZ2JhKDI1NSwgMjU1LCAyNTUsIDAuMDUpIiBzdHJva2Utd2lkdGg9IjEiLz48L3BhdHRlcm4+PC9kZWZzPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9InVybCgjZ3JpZCkiLz48L3N2Zz4=')] opacity-20"></div>
          
          {/* 浮动装饰元素 */}
          <motion.div 
            className="absolute top-20 left-20 w-32 h-32 bg-blue-500/10 rounded-full blur-xl"
            animate={{ 
              scale: [1, 1.2, 1],
              opacity: [0.3, 0.6, 0.3]
            }}
            transition={{ 
              duration: 4,
              repeat: Infinity,
              ease: "easeInOut"
            }}
          />
          <motion.div 
            className="absolute bottom-20 right-20 w-40 h-40 bg-purple-500/10 rounded-full blur-xl"
            animate={{ 
              scale: [1.2, 1, 1.2],
              opacity: [0.4, 0.7, 0.4]
            }}
            transition={{ 
              duration: 5,
              repeat: Infinity,
              ease: "easeInOut",
              delay: 1
            }}
          />
        </div>

        <div className="relative z-10 min-h-screen flex">
          {/* 左侧 - 品牌展示区域 */}
          <div className="hidden lg:flex lg:w-[60%] flex-col justify-center items-center px-12 xl:px-20 text-white">
            <motion.div 
              className="max-w-3xl"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8 }}
            >
              {/* 主图标 */}
              <motion.div 
                className="mb-12"
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ duration: 0.6, delay: 0.2 }}
              >
                <div className="w-40 h-40 mx-auto mb-8 relative">
                  <motion.div 
                    className="absolute inset-0 bg-gradient-to-br from-blue-400 to-purple-600 rounded-3xl shadow-2xl"
                    animate={{ rotate: [0, 3, -3, 0] }}
                    transition={{ 
                      duration: 6,
                      repeat: Infinity,
                      ease: "easeInOut"
                    }}
                  />
                  <div className="absolute inset-3 bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl flex items-center justify-center">
                    <Shield className="w-20 h-20 text-blue-400" />
                  </div>
                  <motion.div 
                    className="absolute -top-2 -right-2 w-10 h-10 bg-green-400 rounded-full flex items-center justify-center"
                    animate={{ scale: [1, 1.2, 1] }}
                    transition={{ 
                      duration: 2,
                      repeat: Infinity,
                      ease: "easeInOut"
                    }}
                  >
                    <CheckCircle2 className="w-6 h-6 text-white" />
                  </motion.div>
                </div>
              </motion.div>
              
              <motion.h1 
                className="text-6xl xl:text-7xl font-bold mb-6 bg-gradient-to-r from-white via-blue-200 to-purple-200 bg-clip-text text-transparent leading-tight"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.4 }}
              >
                Fail2Ban
              </motion.h1>
              <motion.h2 
                className="text-3xl xl:text-4xl font-semibold mb-8 text-blue-200"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.6 }}
              >
                安全防护管理面板
              </motion.h2>
              <motion.p 
                className="text-xl xl:text-2xl text-blue-100 leading-relaxed mb-16"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.8 }}
              >
                专业的服务器安全防护系统
                <br />
                <span className="text-purple-200">智能监控 • 实时防护 • 安全管理</span>
              </motion.p>
              
              {/* 特性列表 */}
              <motion.div
                className="grid grid-cols-2 gap-6 mt-12"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 1 }}
              >
                {[
                  { icon: Shield, label: '实时防护', desc: '24/7 监控' },
                  { icon: Eye, label: '智能分析', desc: '威胁检测' },
                  { icon: CheckCircle2, label: '自动拦截', desc: '恶意访问' },
                  { icon: KeyRound, label: '安全管理', desc: '集中控制' }
                ].map((feature, index) => (
                  <motion.div
                    key={feature.label}
                    className="flex items-start space-x-3"
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.4, delay: 1 + index * 0.1 }}
                  >
                    <div className="flex-shrink-0 w-12 h-12 bg-white/10 rounded-xl flex items-center justify-center backdrop-blur-sm">
                      <feature.icon className="w-6 h-6 text-blue-300" />
                    </div>
                    <div>
                      <div className="text-white font-semibold text-lg">{feature.label}</div>
                      <div className="text-blue-200 text-sm">{feature.desc}</div>
                    </div>
                  </motion.div>
                ))}
              </motion.div>
            </motion.div>
          </div>

          {/* 右侧 - 登录表单 */}
          <div className="flex-1 lg:w-[40%] flex items-center justify-center px-8 py-12 lg:px-12">
            <motion.div 
              className="w-full max-w-[450px]"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.8 }}
            >
              <div className="bg-white/95 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/20 overflow-hidden">
                <div className="relative z-10 p-8 flex flex-col h-[650px]">
                  {/* 表单头部 */}
                  <motion.div 
                    className="mb-8"
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.6, delay: 0.2 }}
                  >
                    <div className="flex items-center justify-center mb-6">
                      <div className="p-4 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl shadow-lg">
                        <KeyRound className="h-8 w-8 text-white" />
                      </div>
                    </div>
                    <h2 className="text-3xl font-bold text-gray-900 mb-2 text-center">欢迎回来</h2>
                    <p className="text-gray-600 text-sm text-center">请登录您的管理账户以继续</p>
                  </motion.div>

                  {/* 错误提示 */}
                  {error && (
                    <motion.div 
                      className="mb-6 bg-red-50 border border-red-200 rounded-xl p-3.5 flex items-center"
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                      transition={{ duration: 0.3 }}
                    >
                      <AlertCircle className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                      <span className="text-red-700 text-sm">{error}</span>
                    </motion.div>
                  )}

                  {/* 登录表单 */}
                  <motion.form 
                    className="space-y-6 flex-1" 
                    onSubmit={handleSubmit}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.6, delay: 0.4 }}
                  >
                    {/* 用户名输入框 */}
                    <div className="space-y-2">
                      <label htmlFor="username" className="block text-sm font-semibold text-gray-700">
                        用户名
                      </label>
                      <div className="relative group">
                        <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                          <User className="h-5 w-5 text-gray-400 group-focus-within:text-blue-500 transition-colors" />
                        </div>
                        <input
                          id="username"
                          name="username"
                          type="text"
                          required
                          value={formData.username}
                          onChange={handleChange}
                          className="block w-full pl-11 pr-4 py-3 text-base border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all hover:border-gray-400"
                          placeholder="请输入用户名"
                        />
                      </div>
                    </div>

                    {/* 密码输入框 */}
                    <div className="space-y-2">
                      <label htmlFor="password" className="block text-sm font-semibold text-gray-700">
                        密码
                      </label>
                      <div className="relative group">
                        <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                          <Lock className="h-5 w-5 text-gray-400 group-focus-within:text-blue-500 transition-colors" />
                        </div>
                        <input
                          id="password"
                          name="password"
                          type={showPassword ? "text" : "password"}
                          required
                          value={formData.password}
                          onChange={handleChange}
                          className="block w-full pl-11 pr-12 py-3 text-base border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all hover:border-gray-400"
                          placeholder="请输入密码"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3.5 flex items-center"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? (
                            <EyeOff className="h-5 w-5 text-gray-400 hover:text-gray-600 transition-colors" />
                          ) : (
                            <Eye className="h-5 w-5 text-gray-400 hover:text-gray-600 transition-colors" />
                          )}
                        </button>
                      </div>
                    </div>

                    {/* 记住我和忘记密码 */}
                    <div className="flex items-center justify-between text-sm">
                      <label className="flex items-center cursor-pointer group">
                        <input
                          type="checkbox"
                          name="rememberMe"
                          checked={formData.rememberMe}
                          onChange={handleChange}
                          className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
                        />
                        <span className="ml-2.5 font-medium text-gray-700 group-hover:text-gray-900 transition-colors">记住我</span>
                      </label>
                      <button 
                        type="button"
                        className="font-medium text-blue-600 hover:text-blue-700 hover:underline transition-colors"
                      >
                        忘记密码？
                      </button>
                    </div>

                    {/* 登录按钮 */}
                    <motion.button
                      type="submit"
                      disabled={loading}
                      className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3.5 rounded-xl hover:from-blue-700 hover:to-purple-700 font-semibold text-base shadow-lg shadow-blue-500/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300"
                      whileHover={{ scale: loading ? 1 : 1.02, boxShadow: loading ? undefined : '0 15px 35px rgba(59, 130, 246, 0.4)' }}
                      whileTap={{ scale: loading ? 1 : 0.98 }}
                    >
                      {loading ? (
                        <div className="flex items-center justify-center">
                          <motion.div 
                            className="w-5 h-5 border-2 border-white border-t-transparent rounded-full mr-3"
                            animate={{ rotate: 360 }}
                            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                          />
                          登录中...
                        </div>
                      ) : (
                        <span className="flex items-center justify-center">
                          登录到控制面板
                          <motion.span
                            className="ml-2"
                            initial={{ x: 0 }}
                            animate={{ x: [0, 5, 0] }}
                            transition={{ duration: 1.5, repeat: Infinity }}
                          >
                            →
                          </motion.span>
                        </span>
                      )}
                    </motion.button>

                    {/* 快速提示 */}
                    <motion.div
                      className="mt-6 p-4 bg-blue-50/50 border border-blue-100 rounded-xl"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.6, delay: 0.6 }}
                    >
                      <div className="flex items-start">
                        <div className="flex-shrink-0">
                          <Shield className="h-5 w-5 text-blue-600 mt-0.5" />
                        </div>
                        <div className="ml-3">
                          <p className="text-xs text-gray-600 leading-relaxed">
                            <span className="font-semibold text-gray-700">安全提示：</span>
                            首次登录建议修改默认密码，启用双因素认证以增强账户安全性。
                          </p>
                        </div>
                      </div>
                    </motion.div>
                  </motion.form>


                  {/* 版权信息 */}
                  <motion.div 
                    className="mt-6 pt-6 border-t border-gray-100 text-center"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.6, delay: 0.8 }}
                  >
                    <p className="text-xs text-gray-500">
                      © 2025 Fail2Ban 安全管理面板
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      Version 1.0.0
                    </p>
                  </motion.div>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    );
}

export default function LoginPage() {
  return (
    <AuthGuard requireAuth={false}>
      <LoginForm />
    </AuthGuard>
  );
}