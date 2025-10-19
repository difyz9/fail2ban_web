import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  eslint: {
    // 生产构建时忽略 ESLint 错误
    ignoreDuringBuilds: true,
  },
  typescript: {
    // 生产构建时忽略 TypeScript 错误（不推荐长期使用）
    ignoreBuildErrors: true,
  },
  output: 'export',  // 导出静态 HTML
  distDir: 'out',    // 输出目录为 out
};

export default nextConfig;
