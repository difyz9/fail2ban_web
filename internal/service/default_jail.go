package service

import (
	"fail2ban-web/internal/model"
)

// DefaultJailService 默认配置服务
type DefaultJailService struct {
	jailService *JailService
}

func NewDefaultJailService(jailService *JailService) *DefaultJailService {
	return &DefaultJailService{
		jailService: jailService,
	}
}

// CreateDefaultNginxJails 创建默认的 Nginx 相关 jail 配置
func (s *DefaultJailService) CreateDefaultNginxJails() error {
	defaultJails := s.getDefaultNginxJails()
	
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

// getDefaultNginxJails 获取默认的 Nginx jail 配置
func (s *DefaultJailService) getDefaultNginxJails() []model.Fail2banJail {
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
			Action:      "iptables-multiport[name=nginx-http-auth, port=\"http,https\", protocol=tcp]",
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
			Action:      "iptables-multiport[name=nginx-botsearch, port=\"http,https\", protocol=tcp]",
		},
	}
}