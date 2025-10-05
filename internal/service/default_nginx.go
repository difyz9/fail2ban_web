package service

import (
	"fmt"
	"fail2ban-web/internal/model"
)

// DefaultNginxService 默认Nginx配置服务
type DefaultNginxService struct {
	jailService *JailService
}

func NewDefaultNginxService() *DefaultNginxService {
	return &DefaultNginxService{}
}

func NewDefaultNginxServiceWithJail(jailService *JailService) *DefaultNginxService {
	return &DefaultNginxService{
		jailService: jailService,
	}
}

// InstallNginxDefaults 安装Nginx默认配置
func (s *DefaultNginxService) InstallNginxDefaults() error {
	if s.jailService == nil {
		return fmt.Errorf("jail service not initialized")
	}
	
	defaultJails := s.GetDefaultNginxJails()
	
	for _, jail := range defaultJails {
		// 检查是否已存在
		existingJail, err := s.jailService.GetJailByName(jail.Name)
		if err == nil && existingJail != nil {
			// 如果已存在，跳过
			continue
		}
		
		// 创建新的 jail 配置
		if err := s.jailService.CreateJail(&jail); err != nil {
			return err
		}
	}
	
	return nil
}

// GetDefaultNginxJails 获取默认的 Nginx jail 配置
func (s *DefaultNginxService) GetDefaultNginxJails() []model.Fail2banJail {
	return []model.Fail2banJail{
		{
			Name:        "nginx-http-auth",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-http-auth",
			LogPath:     "/var/log/nginx/error.log",
			MaxRetry:    5,
			FindTime:    600,  // 10分钟
			BanTime:     3600, // 1小时
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-botsearch",
			Enabled:     true,
			Port:        "http,https", 
			Protocol:    "tcp",
			Filter:      "nginx-botsearch",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    2,
			FindTime:    600,  // 10分钟
			BanTime:     86400, // 24小时
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-bad-request",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-bad-request",
			LogPath:     "/var/log/nginx/access.log",
			MaxRetry:    3,
			FindTime:    300,  // 5分钟
			BanTime:     1800, // 30分钟
			Action:      "iptables-multiport",
		},
		{
			Name:        "nginx-limit-req",
			Enabled:     true,
			Port:        "http,https",
			Protocol:    "tcp",
			Filter:      "nginx-limit-req",
			LogPath:     "/var/log/nginx/error.log",
			MaxRetry:    5,
			FindTime:    60,   // 1分钟
			BanTime:     300,  // 5分钟
			Action:      "iptables-multiport",
		},
	}
}

// GetNginxFilterTemplates 获取 Nginx 过滤器模板
func (s *DefaultNginxService) GetNginxFilterTemplates() map[string]string {
	return map[string]string{
		"nginx-http-auth": `# Fail2Ban filter for nginx http auth failures
[Definition]
failregex = ^ \[error\] \d+#\d+: \*\d+ user "\S+":? (password mismatch|was not found in ".*"), client: <HOST>, server: \S+, request: "\S+ \S+ HTTP/\d+\.\d+", host: "\S+"$
            ^ \[error\] \d+#\d+: \*\d+ no user/password was provided for basic authentication, client: <HOST>, server: \S+, request: "\S+ \S+ HTTP/\d+\.\d+", host: "\S+"$

ignoreregex =`,

		"nginx-botsearch": `# Fail2Ban filter for nginx bot search
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*(\.php|\.asp|\.exe|\.pl|\.cgi|\.scgi).*" [2-4][0-9][0-9] .*$

ignoreregex =`,

		"nginx-bad-request": `# Fail2Ban filter for nginx bad requests
[Definition]
failregex = ^<HOST> -.*"(GET|POST|HEAD).*HTTP.*" (4|5)[0-9][0-9] .*$

ignoreregex =`,

		"nginx-limit-req": `# Fail2Ban filter for nginx limit_req
[Definition]
failregex = limiting requests, excess: .* by zone .*, client: <HOST>

ignoreregex =`,
	}
}