package service

import (
	"fail2ban-web/internal/model"
)

type DefaultNginxAdvancedService struct {
	jailService *JailService
}

func NewDefaultNginxAdvancedService(jailService *JailService) *DefaultNginxAdvancedService {
	return &DefaultNginxAdvancedService{
		jailService: jailService,
	}
}

// GetAdvancedNginxJails 获取高级Nginx jail配置
func (s *DefaultNginxAdvancedService) GetAdvancedNginxJails() []model.Fail2banJail {
	return []model.Fail2banJail{
		{
			Name:        "nginx-xmlrpc",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-xmlrpc",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    3,
			FindTime:    300,
			BanTime:     7200,
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-wordpress",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-wordpress",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    5,
			FindTime:    300,
			BanTime:     3600,
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-dos",
			Enabled:     false,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-dos",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    100,
			FindTime:    60,
			BanTime:     600,
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-slowloris",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-slowloris",
			LogPath:     "/var/log/nginx/error.log",
			MaxRetry:    2,
			FindTime:    60,
			BanTime:     1800,
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-phpmyadmin",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-phpmyadmin",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    3,
			FindTime:    300,
			BanTime:     3600,
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-scan",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-scan",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    10,
			FindTime:    600,
			BanTime:     86400,
			Action:      "iptables-multiport",
		},
	}
}

// GetAdvancedNginxFilterTemplates 获取高级Nginx过滤器模板
func (s *DefaultNginxAdvancedService) GetAdvancedNginxFilterTemplates() map[string]string {
	return map[string]string{
		"nginx-xmlrpc": `# WordPress XMLRPC攻击过滤器
[Definition]
failregex = ^<HOST> -.*"POST.*xmlrpc\.php.*" [2-4][0-9][0-9] .*$
            ^<HOST> -.*"POST.*xmlrpc\.php.*" 200 .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-wordpress": `# WordPress登录攻击过滤器
[Definition]
failregex = ^<HOST> -.*"POST.*wp-login\.php.*" 200 .*$
            ^<HOST> -.*"POST.*wp-admin.*" 403 .*$
            ^<HOST> -.*"GET.*wp-admin.*" 302 .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-dos": `# DOS攻击过滤器（谨慎使用）
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*" [2-5][0-9][0-9] .*$
ignoreregex = ^<HOST> -.*"(GET|POST|HEAD).*(\.css|\.js|\.png|\.jpg|\.gif|\.ico|\.woff).*" [2-3][0-9][0-9] .*$

[Init]
maxlines = 1`,

		"nginx-slowloris": `# Slowloris攻击过滤器
[Definition]
failregex = client <HOST> timed out \(110: Connection timed out\)
            upstream timed out \(110: Connection timed out\) while connecting to upstream, client: <HOST>
            client <HOST> closed connection while waiting for request
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-phpmyadmin": `# phpMyAdmin攻击过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST).*(phpmyadmin|pma|myadmin|mysql|sql).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-scan": `# 扫描器检测过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(\.git|\.svn|\.env|config\.php|wp-config|admin|login|setup|install|backup|test|temp|\.bak|\.old).*" [2-4][0-9][0-9] .*$
            ^<HOST> -.*".*User-Agent.*(?:nmap|sqlmap|nikto|dirbuster|gobuster|wpscan|masscan|zap|burp|acunetix).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-sql-injection": `# SQL注入攻击过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(union|select|insert|delete|drop|alter|create|exec|script|alert|document\.cookie|javascript:).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-xss": `# XSS攻击过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(script|alert|onerror|onload|document\.cookie|javascript:|<iframe|<object|<embed).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-path-traversal": `# 路径遍历攻击过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(\.\.\/|\.\.\\\\|%2e%2e%2f|%2e%2e%5c|etc\/passwd|windows\/system32).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,

		"nginx-shell-injection": `# Shell注入攻击过滤器
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(;|&&|\|\||` + "`" + `|nc\\s|sh\\s|bash\\s|cmd\\s|eval\\s|system\\s|exec\\s).*" [2-4][0-9][0-9] .*$
ignoreregex =

[Init]
maxlines = 1`,
	}
}

// InstallAdvancedNginxDefaults 安装高级Nginx配置
func (s *DefaultNginxAdvancedService) InstallAdvancedNginxDefaults() error {
	jails := s.GetAdvancedNginxJails()
	
	for _, jail := range jails {
		// 检查是否已存在
		if existingJail, err := s.jailService.GetJailByName(jail.Name); err == nil {
			// 如果存在，更新配置
			jail.ID = existingJail.ID
			if err := s.jailService.UpdateJail(&jail); err != nil {
				return err
			}
		} else {
			// 如果不存在，创建新的
			if err := s.jailService.CreateJail(&jail); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// GetNginxSecurityConfig 获取Nginx安全配置建议
func (s *DefaultNginxAdvancedService) GetNginxSecurityConfig() string {
	return `# Nginx安全配置建议

# 1. 隐藏Nginx版本号
server_tokens off;

# 2. 设置请求大小限制
client_max_body_size 10M;
client_body_buffer_size 128k;
client_header_buffer_size 1k;
large_client_header_buffers 2 1k;

# 3. 设置超时时间
client_body_timeout 12;
client_header_timeout 12;
keepalive_timeout 15;
send_timeout 10;

# 4. 限制请求频率
http {
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=global:10m rate=5r/s;
}

server {
    # 对登录页面限流
    location /login {
        limit_req zone=login burst=2 nodelay;
    }
    
    # 对API接口限流
    location /api {
        limit_req zone=api burst=20 nodelay;
    }
    
    # 全局限流
    limit_req zone=global burst=10 nodelay;
}

# 5. 限制连接数
http {
    limit_conn_zone $binary_remote_addr zone=conn_limit_per_ip:10m;
}

server {
    limit_conn conn_limit_per_ip 10;
}

# 6. 禁止访问敏感文件
location ~* \.(git|svn|env|config|bak|old|sql|log)$ {
    deny all;
    return 404;
}

# 7. 防止SQL注入
location ~* (union|select|insert|delete|drop|alter|script|alert) {
    deny all;
    return 403;
}

# 8. 安全头部
add_header X-Frame-Options SAMEORIGIN;
add_header X-Content-Type-Options nosniff;
add_header X-XSS-Protection "1; mode=block";
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;`
}

// GetNginxBestPractices 获取Nginx安全最佳实践
func (s *DefaultNginxAdvancedService) GetNginxBestPractices() []string {
	return []string{
		"隐藏Nginx版本信息，设置server_tokens off",
		"配置适当的请求大小限制",
		"设置合理的超时时间",
		"启用请求频率限制(limit_req)",
		"启用连接数限制(limit_conn)",
		"禁止访问敏感文件和目录",
		"配置安全HTTP头部",
		"使用HTTPS并强制重定向",
		"定期更新Nginx版本",
		"配置适当的日志记录",
		"使用ModSecurity等WAF工具",
		"监控异常访问模式",
		"定期备份配置文件",
		"使用fail2ban保护Web服务",
		"配置适当的错误页面",
	}
}