// Fail2Ban Web Panel JavaScript
class Fail2BanPanel {
    constructor() {
        this.baseURL = '/api/v1';
        this.currentPage = 'dashboard';
        this.token = localStorage.getItem('token');
        this.init();
    }

    init() {
        // 检查是否已登录
        if (!this.token) {
            window.location.href = '/login';
            return;
        }

        // 显示用户名
        const username = localStorage.getItem('username');
        if (username) {
            const usernameDisplay = document.getElementById('username-display');
            if (usernameDisplay) {
                usernameDisplay.textContent = username;
            }
        }

        this.setupNavigation();
        this.loadDashboard();
        this.startAutoRefresh();
    }

    // 获取认证头
    getAuthHeaders() {
        return {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json'
        };
    }

    // 发送认证请求
    async authenticatedFetch(url, options = {}) {
        const defaultOptions = {
            headers: this.getAuthHeaders(),
            ...options
        };

        const response = await fetch(url, defaultOptions);
        
        if (response.status === 401) {
            // Token过期，重定向到登录页
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            window.location.href = '/login';
            return;
        }

        return response;
    }

    // 处理统一响应格式
    async handleApiResponse(response) {
        if (!response || !response.ok) {
            throw new Error('API请求失败');
        }
        
        const result = await response.json();
        
        // 统一响应格式: { success: true, data: {...}, error: "", message: "" }
        if (result.success) {
            return result.data;
        } else {
            throw new Error(result.error || result.message || 'API请求失败');
        }
    }

    setupNavigation() {
        // 处理导航点击事件
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                if (e.target.getAttribute('href').startsWith('#')) {
                    e.preventDefault();
                    const page = e.target.getAttribute('href').substring(1);
                    this.navigateTo(page);
                }
            });
        });
    }

    navigateTo(page) {
        // 更新导航状态
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[href="#${page}"]`).classList.add('active');

        this.currentPage = page;

        // 根据页面加载不同内容
        switch (page) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'banned-ips':
                this.loadBannedIPs();
                break;
            case 'rules':
                this.loadRules();
                break;
            case 'logs':
                this.loadLogs();
                break;
            case 'settings':
                this.loadSettings();
                break;
            case 'logout':
                this.logout();
                break;
        }
    }

    async loadDashboard() {
        try {
            // 加载统计数据
            await this.loadStats();
            
            // 加载最近被禁IP
            await this.loadRecentBans();

            // 加载系统信息
            await this.loadSystemInfo();

        } catch (error) {
            console.error('加载仪表板数据失败:', error);
            this.showError('加载仪表板数据失败');
        }
    }

    async loadStats() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/stats`);
            const stats = await this.handleApiResponse(response);
            
            document.getElementById('banned-count').textContent = stats.bannedCount || '0';
            document.getElementById('today-blocks').textContent = stats.todayBlocks || '0';
            document.getElementById('active-rules').textContent = stats.activeRules || '0';
            document.getElementById('system-status').textContent = stats.systemStatus || '正常';
            
        } catch (error) {
            console.error('加载统计数据失败:', error);
            // 显示默认值
            document.getElementById('banned-count').textContent = '--';
            document.getElementById('today-blocks').textContent = '--';
            document.getElementById('active-rules').textContent = '--';
            document.getElementById('system-status').textContent = '未知';
        }
    }

    async loadRecentBans() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/banned-ips?limit=10`);
            const data = await this.handleApiResponse(response);
            const tableBody = document.querySelector('#recent-bans-table tbody');
            
            if (data.ips && data.ips.length > 0) {
                tableBody.innerHTML = data.ips.map(ip => `
                    <tr>
                        <td><span class="ip-address">${ip.address}</span></td>
                        <td>${ip.jail}</td>
                        <td>${new Date(ip.banTime).toLocaleString()}</td>
                        <td>
                            <button class="btn btn-sm btn-danger" onclick="app.unbanIP('${ip.address}', '${ip.jail}')">
                                <i class="fas fa-unlock"></i> 解禁
                            </button>
                        </td>
                    </tr>
                `).join('');
            } else {
                tableBody.innerHTML = '<tr><td colspan="4" class="text-center">暂无被禁IP</td></tr>';
            }
            
        } catch (error) {
            console.error('加载最近被禁IP失败:', error);
            document.querySelector('#recent-bans-table tbody').innerHTML = 
                '<tr><td colspan="4" class="text-center text-danger">加载失败</td></tr>';
        }
    }

    async loadSystemInfo() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/system-info`);
            const info = await this.handleApiResponse(response);
            
            document.getElementById('fail2ban-version').textContent = info.version || '未知';
            document.getElementById('uptime').textContent = this.formatUptime(info.uptime) || '未知';
            
        } catch (error) {
            console.error('加载系统信息失败:', error);
            document.getElementById('fail2ban-version').textContent = '获取失败';
            document.getElementById('uptime').textContent = '获取失败';
        }
    }

    async loadBannedIPs() {
        const content = `
            <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                <h1 class="h2">被禁IP管理</h1>
                <div class="btn-toolbar mb-2 mb-md-0">
                    <button type="button" class="btn btn-primary" onclick="app.refreshBannedIPs()">
                        <i class="fas fa-sync-alt"></i> 刷新
                    </button>
                </div>
            </div>
            
            <div class="card shadow">
                <div class="card-header py-3">
                    <h6 class="m-0 font-weight-bold text-primary">当前被禁IP列表</h6>
                </div>
                <div class="card-body">
                    <div class="table-responsive">
                        <table class="table table-bordered" id="banned-ips-table">
                            <thead>
                                <tr>
                                    <th>IP地址</th>
                                    <th>所属Jail</th>
                                    <th>被禁时间</th>
                                    <th>剩余时间</th>
                                    <th>操作</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr><td colspan="5" class="text-center">加载中...</td></tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        `;
        
        document.getElementById('content-area').innerHTML = content;
        await this.refreshBannedIPs();
    }

    async refreshBannedIPs() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/banned-ips`);
            const data = await this.handleApiResponse(response);
            const tableBody = document.querySelector('#banned-ips-table tbody');
            
            if (data.ips && data.ips.length > 0) {
                tableBody.innerHTML = data.ips.map(ip => `
                    <tr>
                        <td><span class="ip-address">${ip.address}</span></td>
                        <td>${ip.jail}</td>
                        <td>${new Date(ip.banTime).toLocaleString()}</td>
                        <td>${this.formatTimeRemaining(ip.remainingTime)}</td>
                        <td>
                            <button class="btn btn-sm btn-danger" onclick="app.unbanIP('${ip.address}', '${ip.jail}')">
                                <i class="fas fa-unlock"></i> 解禁
                            </button>
                        </td>
                    </tr>
                `).join('');
            } else {
                tableBody.innerHTML = '<tr><td colspan="5" class="text-center">暂无被禁IP</td></tr>';
            }
            
        } catch (error) {
            console.error('刷新被禁IP列表失败:', error);
            document.querySelector('#banned-ips-table tbody').innerHTML = 
                '<tr><td colspan="5" class="text-center text-danger">加载失败</td></tr>';
        }
    }

    async unbanIP(ip, jail) {
        if (!confirm(`确定要解禁IP ${ip} 吗？`)) {
            return;
        }

        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/unban`, {
                method: 'POST',
                body: JSON.stringify({
                    ip: ip,
                    jail: jail
                })
            });

            await this.handleApiResponse(response);
            this.showSuccess(`成功解禁IP: ${ip}`);
            
            // 刷新列表
            if (this.currentPage === 'banned-ips') {
                await this.refreshBannedIPs();
            } else {
                await this.loadRecentBans();
            }
            
            // 更新统计
            await this.loadStats();

        } catch (error) {
            console.error('解禁IP失败:', error);
            this.showError(`解禁IP失败: ${error.message}`);
        }
    }

    loadRules() {
        const content = `
            <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                <h1 class="h2">规则管理</h1>
                <div class="btn-toolbar mb-2 mb-md-0">
                    <div class="btn-group me-2">
                        <button type="button" class="btn btn-success" onclick="app.installNginxDefaults()">
                            <i class="fas fa-download"></i> 安装Nginx默认配置
                        </button>
                        <button type="button" class="btn btn-info" onclick="app.showDefaultConfigInfo()">
                            <i class="fas fa-info-circle"></i> 配置说明
                        </button>
                    </div>
                    <button type="button" class="btn btn-primary" onclick="app.refreshRules()">
                        <i class="fas fa-sync-alt"></i> 刷新
                    </button>
                </div>
            </div>
            
            <div class="row mb-4">
                <div class="col-lg-8">
                    <div class="card shadow">
                        <div class="card-header py-3">
                            <h6 class="m-0 font-weight-bold text-primary">Jail 配置列表</h6>
                        </div>
                        <div class="card-body">
                            <div class="table-responsive">
                                <table class="table table-bordered" id="jails-table">
                                    <thead>
                                        <tr>
                                            <th>名称</th>
                                            <th>状态</th>
                                            <th>端口</th>
                                            <th>最大重试</th>
                                            <th>禁止时间</th>
                                            <th>操作</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr><td colspan="6" class="text-center">加载中...</td></tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card shadow">
                        <div class="card-header py-3">
                            <h6 class="m-0 font-weight-bold text-primary">快速操作</h6>
                        </div>
                        <div class="card-body">
                            <div class="list-group">
                                <button class="list-group-item list-group-item-action" onclick="app.exportNginxConfig()">
                                    <i class="fas fa-file-export"></i> 导出Nginx配置
                                </button>
                                <button class="list-group-item list-group-item-action" onclick="app.viewFilterTemplates()">
                                    <i class="fas fa-filter"></i> 查看过滤器模板
                                </button>
                                <button class="list-group-item list-group-item-action" onclick="app.showInstallGuide()">
                                    <i class="fas fa-book"></i> 安装指南
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.getElementById('content-area').innerHTML = content;
        this.refreshRules();
    }

    loadLogs() {
        const content = `
            <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                <h1 class="h2">日志查看</h1>
            </div>
            <div class="alert alert-info" role="alert">
                <i class="fas fa-info-circle"></i> 日志查看功能正在开发中...
            </div>
        `;
        document.getElementById('content-area').innerHTML = content;
    }

    loadSettings() {
        const content = `
            <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                <h1 class="h2">系统设置</h1>
            </div>
            <div class="alert alert-info" role="alert">
                <i class="fas fa-info-circle"></i> 系统设置功能正在开发中...
            </div>
        `;
        document.getElementById('content-area').innerHTML = content;
    }

    logout() {
        if (confirm('确定要退出吗？')) {
            // 清除本地存储
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            // 跳转到登录页
            window.location.href = '/login';
        }
    }

    formatUptime(seconds) {
        if (!seconds) return '未知';
        
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        return `${days}天 ${hours}小时 ${minutes}分钟`;
    }

    formatTimeRemaining(seconds) {
        if (!seconds || seconds <= 0) return '永久';
        
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (hours > 0) {
            return `${hours}小时 ${minutes}分钟`;
        } else {
            return `${minutes}分钟`;
        }
    }

    showSuccess(message) {
        this.showAlert(message, 'success');
    }

    showError(message) {
        this.showAlert(message, 'danger');
    }

    showAlert(message, type) {
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        const container = document.querySelector('main');
        container.insertBefore(alertDiv, container.firstChild);
        
        // 3秒后自动隐藏
        setTimeout(() => {
            alertDiv.remove();
        }, 3000);
    }

    startAutoRefresh() {
        // 每30秒自动刷新统计数据
        setInterval(() => {
            if (this.currentPage === 'dashboard') {
                this.loadStats();
                this.loadRecentBans();
            }
        }, 30000);
    }

    // 刷新规则列表
    async refreshRules() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/jails`);
            const data = await this.handleApiResponse(response);
            const tableBody = document.querySelector('#jails-table tbody');
            
            if (data.jails && data.jails.length > 0) {
                tableBody.innerHTML = data.jails.map(jail => `
                    <tr>
                        <td><strong>${jail.name}</strong></td>
                        <td>
                            <span class="badge ${jail.enabled ? 'bg-success' : 'bg-secondary'}">
                                ${jail.enabled ? '启用' : '禁用'}
                            </span>
                        </td>
                        <td>${jail.port || '-'}</td>
                        <td>${jail.max_retry || '-'}</td>
                        <td>${this.formatDuration(jail.ban_time)}</td>
                        <td>
                            <div class="btn-group btn-group-sm">
                                <button class="btn btn-outline-primary" onclick="app.editJail('${jail.name}')">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-outline-${jail.enabled ? 'warning' : 'success'}" 
                                        onclick="app.toggleJail('${jail.name}', ${!jail.enabled})">
                                    <i class="fas fa-${jail.enabled ? 'pause' : 'play'}"></i>
                                </button>
                                <button class="btn btn-outline-danger" onclick="app.deleteJail('${jail.name}')">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                `).join('');
            } else {
                tableBody.innerHTML = '<tr><td colspan="6" class="text-center">暂无规则配置</td></tr>';
            }
            
        } catch (error) {
            console.error('刷新规则列表失败:', error);
            document.querySelector('#jails-table tbody').innerHTML = 
                '<tr><td colspan="6" class="text-center text-danger">加载失败</td></tr>';
        }
    }

    // 安装Nginx默认配置
    async installNginxDefaults() {
        if (!confirm('确定要安装Nginx默认配置吗？这将添加10个预配置的安全规则。')) {
            return;
        }

        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/defaults/nginx/install`, {
                method: 'POST'
            });

            const result = await this.handleApiResponse(response);
            this.showSuccess(result.message || '安装成功');
            
            // 刷新规则列表
            if (this.currentPage === 'rules') {
                await this.refreshRules();
            }

        } catch (error) {
            console.error('安装Nginx默认配置失败:', error);
            this.showError(`安装失败: ${error.message}`);
        }
    }

    // 显示默认配置信息
    async showDefaultConfigInfo() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/defaults/info`);
            const info = await this.handleApiResponse(response);
            
            let modalContent = `
                <div class="modal fade" id="configInfoModal" tabindex="-1">
                    <div class="modal-dialog modal-lg">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h5 class="modal-title">默认配置说明</h5>
                                <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                            </div>
                            <div class="modal-body">
                                <div class="accordion" id="configAccordion">
            `;

            info.nginx_jails.categories.forEach((category, index) => {
                modalContent += `
                    <div class="accordion-item">
                        <h2 class="accordion-header" id="heading${index}">
                            <button class="accordion-button ${index === 0 ? '' : 'collapsed'}" type="button" 
                                    data-bs-toggle="collapse" data-bs-target="#collapse${index}">
                                ${category.name} (${category.jails.length} 个规则)
                            </button>
                        </h2>
                        <div id="collapse${index}" class="accordion-collapse collapse ${index === 0 ? 'show' : ''}" 
                             data-bs-parent="#configAccordion">
                            <div class="accordion-body">
                                <p>${category.description}</p>
                                <ul>
                                    ${category.jails.map(jail => `<li><code>${jail}</code></li>`).join('')}
                                </ul>
                            </div>
                        </div>
                    </div>
                `;
            });

            modalContent += `
                                </div>
                                <div class="mt-4">
                                    <h6>安装步骤:</h6>
                                    <ol>
                                        ${info.installation_steps.map(step => `<li>${step}</li>`).join('')}
                                    </ol>
                                </div>
                                <div class="mt-3">
                                    <h6>建议:</h6>
                                    <ul>
                                        ${info.recommendations.map(rec => `<li>${rec}</li>`).join('')}
                                    </ul>
                                </div>
                            </div>
                            <div class="modal-footer">
                                <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                                <button type="button" class="btn btn-primary" onclick="app.installNginxDefaults()">
                                    安装配置
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            `;

            // 移除现有的modal
            const existingModal = document.getElementById('configInfoModal');
            if (existingModal) {
                existingModal.remove();
            }

            // 添加新的modal
            document.body.insertAdjacentHTML('beforeend', modalContent);
            
            // 显示modal
            const modal = new bootstrap.Modal(document.getElementById('configInfoModal'));
            modal.show();

        } catch (error) {
            console.error('获取配置信息失败:', error);
            this.showError(`获取信息失败: ${error.message}`);
        }
    }

    // 导出Nginx配置
    async exportNginxConfig() {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/defaults/nginx/export`);
            const config = await this.handleApiResponse(response);
            
            // 创建下载链接
            const dataStr = "data:text/plain;charset=utf-8," + encodeURIComponent(config.jail_config);
            const downloadAnchorNode = document.createElement('a');
            downloadAnchorNode.setAttribute("href", dataStr);
            downloadAnchorNode.setAttribute("download", "nginx-jails.conf");
            document.body.appendChild(downloadAnchorNode);
            downloadAnchorNode.click();
            downloadAnchorNode.remove();

            this.showSuccess('配置文件已下载');

        } catch (error) {
            console.error('导出配置失败:', error);
            this.showError(`导出失败: ${error.message}`);
        }
    }

    // 切换jail状态
    async toggleJail(name, enabled) {
        try {
            const response = await this.authenticatedFetch(`${this.baseURL}/jails/${name}/toggle`, {
                method: 'POST',
                body: JSON.stringify({ enabled })
            });

            const result = await this.handleApiResponse(response);
            this.showSuccess(result.message || '操作成功');
            
            // 刷新列表
            await this.refreshRules();

        } catch (error) {
            console.error('切换jail状态失败:', error);
            this.showError(`操作失败: ${error.message}`);
        }
    }

    // 格式化持续时间
    formatDuration(seconds) {
        if (!seconds) return '-';
        
        if (seconds < 60) {
            return `${seconds}秒`;
        } else if (seconds < 3600) {
            return `${Math.floor(seconds / 60)}分钟`;
        } else if (seconds < 86400) {
            return `${Math.floor(seconds / 3600)}小时`;
        } else {
            return `${Math.floor(seconds / 86400)}天`;
        }
    }

    // 编辑jail (占位符)
    editJail(name) {
        this.showAlert('编辑功能开发中...', 'info');
    }

    // 删除jail (占位符)
    deleteJail(name) {
        this.showAlert('删除功能开发中...', 'info');
    }

    // 查看过滤器模板
    viewFilterTemplates() {
        this.showAlert('过滤器模板查看功能开发中...', 'info');
    }

    // 显示安装指南
    showInstallGuide() {
        this.showAlert('安装指南功能开发中...', 'info');
    }
}

// 初始化应用
const app = new Fail2BanPanel();