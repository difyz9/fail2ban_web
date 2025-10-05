package service

import (
	"fail2ban-web/internal/model"
)

type DefaultSSHService struct {
	jailService *JailService
}

func NewDefaultSSHService(jailService *JailService) *DefaultSSHService {
	return &DefaultSSHService{
		jailService: jailService,
	}
}

// GetDefaultSSHJails 获取默认SSH jail配置
func (s *DefaultSSHService) GetDefaultSSHJails() []model.Fail2banJail {
	return []model.Fail2banJail{
		{
			Name:        "sshd",
			Enabled:     true,
			Port:        "ssh",
			Protocol:    "tcp",
			Filter:      "sshd",
			LogPath:     "/var/log/auth.log",
			MaxRetry:    3,
			FindTime:    600,  // 10分钟
			BanTime:     3600, // 1小时
			Action:      "iptables-multiport",
		},
		{
			Name:        "sshd-ddos",
			Enabled:     true,
			Port:        "ssh",
			Protocol:    "tcp",
			Filter:      "sshd-ddos",
			LogPath:     "/var/log/auth.log",
			MaxRetry:    2,
			FindTime:    60,   // 1分钟
			BanTime:     3600, // 1小时
			Action:      "iptables-multiport",
		},
		{
			Name:        "sshd-aggressive",
			Enabled:     false,
			Port:        "ssh",
			Protocol:    "tcp",
			Filter:      "sshd-aggressive",
			LogPath:     "/var/log/auth.log",
			MaxRetry:    1,
			FindTime:    300,  // 5分钟
			BanTime:     86400, // 24小时
			Action:      "iptables-multiport",
		},
	}
}

// GetSSHFilterTemplates 获取SSH过滤器模板
func (s *DefaultSSHService) GetSSHFilterTemplates() map[string]string {
	return map[string]string{
		"sshd": `[Definition]
# SSH失败登录检测
failregex = ^%(__prefix_line)s(?:error: PAM: )?Authentication failure for .* from <HOST>( via \S+)?\s*$
            ^%(__prefix_line)s(?:error: PAM: )?User not known to the underlying authentication module for .* from <HOST>\s*$
            ^%(__prefix_line)sFailed password for (?:invalid user |)(?P<user>\S+) from <HOST> port \d+ ssh2\s*$
            ^%(__prefix_line)sReceived disconnect from <HOST> port \d+:11: (?:Bye|Authentication failed)\s*$
            ^%(__prefix_line)sDisconnected from authenticating user (?P<user>\S+) <HOST> port \d+ \[preauth\]\s*$

ignoreregex =

[Init]
maxlines = 1`,

		"sshd-ddos": `[Definition]
# SSH DDoS攻击检测
failregex = ^%(__prefix_line)sDid not receive identification string from <HOST>\s*$
            ^%(__prefix_line)sReceived disconnect from <HOST> port \d+:11: Bye Bye \[preauth\]\s*$
            ^%(__prefix_line)sConnection closed by <HOST> port \d+ \[preauth\]\s*$
            ^%(__prefix_line)sSSH: Server;Ltype: Version;Remote: <HOST>-\d+;Name: \S+;.*$

ignoreregex =

[Init]
maxlines = 1`,

		"sshd-aggressive": `[Definition]
# 更激进的SSH检测
failregex = ^%(__prefix_line)s(?:error: PAM: )?Authentication failure for .* from <HOST>( via \S+)?\s*$
            ^%(__prefix_line)s(?:error: PAM: )?User not known to the underlying authentication module for .* from <HOST>\s*$
            ^%(__prefix_line)sFailed password for (?:invalid user |)(?P<user>\S+) from <HOST> port \d+ ssh2\s*$
            ^%(__prefix_line)sReceived disconnect from <HOST> port \d+:11: (?:Bye|Authentication failed)\s*$
            ^%(__prefix_line)sDisconnected from authenticating user (?P<user>\S+) <HOST> port \d+ \[preauth\]\s*$
            ^%(__prefix_line)sDid not receive identification string from <HOST>\s*$
            ^%(__prefix_line)sConnection closed by <HOST> port \d+ \[preauth\]\s*$
            ^%(__prefix_line)sInvalid user .* from <HOST>\s*$

ignoreregex =

[Init]
maxlines = 1`,
	}
}

// InstallSSHDefaults 安装默认SSH配置
func (s *DefaultSSHService) InstallSSHDefaults() error {
	jails := s.GetDefaultSSHJails()
	
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

// GetSSHJailConfig 获取SSH jail配置文件内容
func (s *DefaultSSHService) GetSSHJailConfig() string {
	return `# SSH相关的Fail2Ban配置

[sshd]
# 标准SSH保护
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3
findtime = 600
bantime = 3600
filter = sshd
action = iptables-multiport[name=SSH, port="ssh", protocol=tcp]

[sshd-ddos]
# SSH DDoS攻击防护
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 2
findtime = 60
bantime = 3600
filter = sshd-ddos
action = iptables-multiport[name=SSH-DDOS, port="ssh", protocol=tcp]

[sshd-aggressive]
# 激进的SSH保护（默认禁用）
enabled = false
port = ssh
logpath = /var/log/auth.log
maxretry = 1
findtime = 300
bantime = 86400
filter = sshd-aggressive
action = iptables-multiport[name=SSH-AGGRESSIVE, port="ssh", protocol=tcp]`
}

// GetSSHBestPractices 获取SSH安全最佳实践
func (s *DefaultSSHService) GetSSHBestPractices() []string {
	return []string{
		"修改SSH默认端口22到其他端口",
		"禁用root直接登录",
		"使用SSH密钥认证而不是密码认证",
		"启用SSH协议版本2",
		"限制SSH登录用户",
		"设置合理的SSH连接超时时间",
		"使用AllowUsers或AllowGroups限制登录",
		"启用SSH日志记录",
		"定期检查SSH登录日志",
		"使用强密码策略",
		"启用双因素认证(2FA)",
		"使用防火墙限制SSH访问源IP",
	}
}

// GetSSHSecurityTips 获取SSH安全提示
func (s *DefaultSSHService) GetSSHSecurityTips() map[string]string {
	return map[string]string{
		"端口修改": "编辑/etc/ssh/sshd_config，修改Port 22为其他端口",
		"禁用root登录": "在sshd_config中设置PermitRootLogin no",
		"密钥认证": "生成SSH密钥对，禁用密码认证PasswordAuthentication no",
		"协议版本": "在sshd_config中设置Protocol 2",
		"用户限制": "使用AllowUsers user1 user2限制登录用户",
		"连接限制": "设置MaxAuthTries 3和ClientAliveInterval 300",
		"日志配置": "确保rsyslog服务运行，检查/var/log/auth.log",
		"防火墙": "使用ufw或iptables限制SSH访问端口",
	}
}