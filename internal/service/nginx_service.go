package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fail2ban-web/config"

	"gorm.io/gorm"
)

type NginxService struct {
	config *config.Config
	db     *gorm.DB
}

type NginxStats struct {
	TotalRequests     int       `json:"total_requests"`
	AttackRequests    int       `json:"attack_requests"`
	BannedIPs         int       `json:"banned_ips"`
	BlockedRequests   int       `json:"blocked_requests"`
	LastAttack        time.Time `json:"last_attack"`
	TopAttackerIPs    []string  `json:"top_attacker_ips"`
	AttackTypes       []AttackType `json:"attack_types"`
	StatusCodes       map[string]int `json:"status_codes"`
}

type AttackType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type NginxLog struct {
	Timestamp   time.Time `json:"timestamp"`
	IP          string    `json:"ip"`
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	StatusCode  int       `json:"status_code"`
	UserAgent   string    `json:"user_agent"`
	AttackType  string    `json:"attack_type"`
	IsBlocked   bool      `json:"is_blocked"`
}

func NewNginxService(cfg *config.Config, db *gorm.DB) *NginxService {
	return &NginxService{
		config: cfg,
		db:     db,
	}
}

// GetNginxStats 获取Nginx统计信息
func (s *NginxService) GetNginxStats() (*NginxStats, error) {
	stats := &NginxStats{
		StatusCodes: make(map[string]int),
	}
	
	// 获取Nginx jail状态
	bannedCount, err := s.getNginxBannedCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get banned count: %v", err)
	}
	stats.BannedIPs = bannedCount

	// 分析Nginx访问日志
	logStats, err := s.analyzeNginxLogs()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze Nginx logs: %v", err)
	}
	
	stats.TotalRequests = logStats.TotalRequests
	stats.AttackRequests = logStats.AttackRequests
	stats.BlockedRequests = logStats.BlockedRequests
	stats.LastAttack = logStats.LastAttack
	stats.TopAttackerIPs = logStats.TopAttackerIPs
	stats.AttackTypes = logStats.AttackTypes
	stats.StatusCodes = logStats.StatusCodes

	return stats, nil
}

// getNginxBannedCount 获取Nginx被禁IP数量
func (s *NginxService) getNginxBannedCount() (int, error) {
	totalBanned := 0
	jails := []string{"nginx-http-auth", "nginx-botsearch", "nginx-bad-request", "nginx-limit-req"}
	
	for _, jail := range jails {
		cmd := exec.Command("fail2ban-client", "status", jail)
		output, err := cmd.Output()
		if err != nil {
			continue // 如果jail不存在，跳过
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Currently banned:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					countStr := strings.TrimSpace(parts[1])
					if count, err := strconv.Atoi(countStr); err == nil {
						totalBanned += count
					}
				}
			}
		}
	}
	
	return totalBanned, nil
}

// analyzeNginxLogs 分析Nginx日志
func (s *NginxService) analyzeNginxLogs() (*NginxStats, error) {
	stats := &NginxStats{
		StatusCodes: make(map[string]int),
	}
	ipCount := make(map[string]int)
	attackTypeCount := make(map[string]int)
	
	// 读取Nginx访问日志
	accessLogPath := "/var/log/nginx/access.log"
	if s.config.Fail2Ban.NginxAccessLog != "" {
		accessLogPath = s.config.Fail2Ban.NginxAccessLog
	}
	
	file, err := os.Open(accessLogPath)
	if err != nil {
		// 如果无法读取日志，返回空统计
		return stats, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// Nginx访问日志的正则表达式（Common Log Format）
	logRegex := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*)" (\d+) (\d+) "([^"]*)" "([^"]*)"`)

	for scanner.Scan() {
		line := scanner.Text()
		
		if matches := logRegex.FindStringSubmatch(line); matches != nil {
			stats.TotalRequests++
			
			ip := matches[1]
			method := matches[3]
			url := matches[4]
			statusCode := matches[5]
			userAgent := matches[8]
			
			// 统计状态码
			stats.StatusCodes[statusCode]++
			
			// 检测攻击类型
			attackType := s.detectAttackType(method, url, userAgent, statusCode)
			if attackType != "" {
				stats.AttackRequests++
				ipCount[ip]++
				attackTypeCount[attackType]++
				
				// 尝试解析时间戳
				if timestamp, err := parseNginxTimestamp(matches[2]); err == nil {
					if timestamp.After(stats.LastAttack) {
						stats.LastAttack = timestamp
					}
				}
			}
			
			// 检查是否被阻止（4xx状态码）
			if strings.HasPrefix(statusCode, "4") {
				stats.BlockedRequests++
			}
		}
	}

	// 获取攻击最多的IP
	stats.TopAttackerIPs = getTopIPs(ipCount, 5)
	
	// 转换攻击类型统计
	for attackType, count := range attackTypeCount {
		stats.AttackTypes = append(stats.AttackTypes, AttackType{
			Type:  attackType,
			Count: count,
		})
	}

	return stats, nil
}

// detectAttackType 检测攻击类型
func (s *NginxService) detectAttackType(method, url, userAgent, statusCode string) string {
	url = strings.ToLower(url)
	userAgent = strings.ToLower(userAgent)
	
	// WordPress攻击检测 (最高优先级 - 根据真实日志分析)
	if strings.Contains(url, "/wp-admin/setup-config.php") {
		return "wordpress_exploitation"
	}
	if strings.Contains(url, "/wordpress/wp-admin/") || strings.Contains(url, "/wp-admin/") {
		return "wordpress_scan"
	}
	if strings.Contains(url, "/wp-content/") || strings.Contains(url, "/wp-includes/") {
		return "wordpress_file_access"
	}
	
	// 管理面板攻击检测
	if strings.Contains(url, "/admin/config.php") {
		return "admin_config_exploit"
	}
	if strings.Contains(url, "/admin/login.php") || strings.Contains(url, "/admin/login.asp") {
		return "admin_login_scan"
	}
	if strings.Contains(url, "/boaform/admin/formlogin") {
		return "router_admin_exploit"
	}
	
	// PHP文件扫描检测
	if strings.Contains(url, ".php") {
		phpScanPatterns := []string{
			"config.php", "admin.php", "login.php", "test.php", 
			"info.php", "shell.php", "upload.php", "index.php",
		}
		for _, pattern := range phpScanPatterns {
			if strings.Contains(url, pattern) {
				return "php_file_scan"
			}
		}
		return "php_access"
	}
	
	// 路由器/IoT设备攻击检测
	if strings.Contains(url, "/cgi-bin/luci/") {
		return "router_exploit"
	}
	if strings.Contains(url, "/manager/text/list") {
		return "tomcat_manager_scan"
	}
	
	// 代理滥用检测
	if method == "CONNECT" && strings.Contains(url, ":443") {
		return "proxy_abuse"
	}
	if method == "PROPFIND" {
		return "webdav_scan"
	}
	
	// SQL注入检测
	sqlPatterns := []string{"union", "select", "insert", "delete", "drop", "alter", "'", "\"", "--", "/*"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(url, pattern) {
			return "sql_injection"
		}
	}
	
	// XSS检测
	xssPatterns := []string{"<script", "javascript:", "onerror=", "onload=", "alert(", "document.cookie"}
	for _, pattern := range xssPatterns {
		if strings.Contains(url, pattern) {
			return "xss"
		}
	}
	
	// 路径遍历检测
	if strings.Contains(url, "../") || strings.Contains(url, "..\\") {
		return "path_traversal"
	}
	
	// 可疑User-Agent检测
	suspiciousAgents := []string{"xfa1", "zgrab", "masscan", "nmap", "nikto", "sqlmap"}
	for _, agent := range suspiciousAgents {
		if strings.Contains(userAgent, agent) {
			return "malicious_scanner"
		}
	}
	
	// 恶意机器人检测
	botPatterns := []string{"bot", "crawler", "spider", "scraper", "scanner"}
	for _, pattern := range botPatterns {
		if strings.Contains(userAgent, pattern) {
			return "malicious_bot"
		}
	}
	
	// SSL探测检测
	if strings.Contains(userAgent, "\\x16\\x03\\x01") {
		return "ssl_probe"
	}
	
	// 403/404 扫描检测
	if statusCode == "403" || statusCode == "404" {
		scanPatterns := []string{".php", ".asp", ".jsp", "admin", "login", "config", ".env", ".git"}
		for _, pattern := range scanPatterns {
			if strings.Contains(url, pattern) {
				return "directory_scan"
			}
		}
	}
	
	// HTTP认证失败
	if statusCode == "401" {
		return "auth_failure"
	}
	
	// 频率限制
	if statusCode == "429" {
		return "rate_limit"
	}
	
	return ""
}

// GetNginxLogs 获取Nginx日志
func (s *NginxService) GetNginxLogs(limit int) ([]NginxLog, error) {
	var logs []NginxLog
	
	// 开发模式使用测试日志
	if s.config.Fail2Ban.DevMode {
		return s.getTestNginxLogs(limit), nil
	}
	
	accessLogPath := "/var/log/nginx/access.log"
	if s.config.Fail2Ban.NginxAccessLog != "" {
		accessLogPath = s.config.Fail2Ban.NginxAccessLog
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(accessLogPath); os.IsNotExist(err) {
		// 尝试其他常见的Nginx日志路径
		alternativePaths := []string{
			"/var/log/nginx/access.log",
			"/usr/local/nginx/logs/access.log",
			"/var/log/nginx/default.access.log",
			"/etc/nginx/logs/access.log",
		}
		
		found := false
		for _, path := range alternativePaths {
			if _, err := os.Stat(path); err == nil {
				accessLogPath = path
				found = true
				break
			}
		}
		
		if !found {
			return logs, fmt.Errorf("nginx access log not found at %s or alternative paths", accessLogPath)
		}
	}
	
	file, err := os.Open(accessLogPath)
	if err != nil {
		return logs, fmt.Errorf("failed to open nginx log file %s: %w", accessLogPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	
	// 读取所有行
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return logs, fmt.Errorf("error reading nginx log file: %w", err)
	}

	// 从最后开始处理，获取最新的日志
	count := 0
	parsed := 0
	for i := len(lines) - 1; i >= 0 && count < limit; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		if log := parseNginxLogLine(line); log != nil {
			logs = append([]NginxLog{*log}, logs...)
			count++
		}
		parsed++
	}
	
	// 记录解析结果
	fmt.Printf("Nginx日志解析完成: 读取%d行，解析成功%d条，文件: %s\n", parsed, count, accessLogPath)

	return logs, nil
}

// getTestNginxLogs 获取测试Nginx日志
func (s *NginxService) getTestNginxLogs(limit int) []NginxLog {
	testLogs := []string{
		`192.168.1.100 - - [03/Oct/2025:22:45:01 +0000] "GET / HTTP/1.1" 200 615 "-" "Mozilla/5.0"`,
		`10.0.0.1 - - [03/Oct/2025:22:45:02 +0000] "POST /login HTTP/1.1" 401 82 "-" "curl/7.68.0"`,
		`192.168.1.200 - - [03/Oct/2025:22:45:03 +0000] "GET /admin HTTP/1.1" 404 162 "-" "sqlmap/1.0"`,
		`172.16.0.1 - - [03/Oct/2025:22:45:04 +0000] "GET /?id=1' OR '1'='1 HTTP/1.1" 400 400 "-" "BadBot/1.0"`,
		`203.0.113.1 - - [03/Oct/2025:22:45:05 +0000] "GET /wp-admin HTTP/1.1" 404 162 "-" "scanner"`,
		`198.51.100.1 - - [03/Oct/2025:22:45:06 +0000] "GET /admin.php HTTP/1.1" 404 162 "-" "BadBot/2.0"`,
		`203.0.113.2 - - [03/Oct/2025:22:45:07 +0000] "POST /xmlrpc.php HTTP/1.1" 404 162 "-" "WordPress/5.0"`,
		`192.168.1.300 - - [03/Oct/2025:22:45:08 +0000] "GET /<script>alert(1)</script> HTTP/1.1" 400 400 "-" "XSSBot"`,
	}
	
	var logs []NginxLog
	for i, line := range testLogs {
		if i >= limit {
			break
		}
		if log := parseNginxLogLine(line); log != nil {
			logs = append(logs, *log)
		}
	}
	
	fmt.Printf("开发模式: 生成了 %d 条测试Nginx日志\n", len(logs))
	return logs
}

// GetNginxJailStatus 获取Nginx jail状态
func (s *NginxService) GetNginxJailStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})
	jails := []string{"nginx-http-auth", "nginx-botsearch", "nginx-bad-request", "nginx-limit-req"}
	
	for _, jail := range jails {
		jailStatus, err := s.getJailStatus(jail)
		if err == nil {
			status[jail] = jailStatus
		}
	}

	return status, nil
}

// getJailStatus 获取指定jail的状态
func (s *NginxService) getJailStatus(jailName string) (map[string]interface{}, error) {
	cmd := exec.Command("fail2ban-client", "status", jailName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	status := make(map[string]interface{})
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// 尝试转换为数字
				if intVal, err := strconv.Atoi(value); err == nil {
					status[key] = intVal
				} else {
					status[key] = value
				}
			}
		}
	}

	return status, nil
}

// BanNginxIP 手动禁止Nginx IP
func (s *NginxService) BanNginxIP(ip string, jail string) error {
	if jail == "" {
		jail = "nginx-http-auth"
	}
	
	// 使用fail2ban服务的统一命令执行
	fail2banSvc := NewFail2BanService(nil)
	return fail2banSvc.BanIP(jail, ip)
}

// UnbanNginxIP 解禁Nginx IP
func (s *NginxService) UnbanNginxIP(ip string, jail string) error {
	if jail == "" {
		jail = "nginx-http-auth"
	}
	
	// 使用fail2ban服务的统一命令执行
	fail2banSvc := NewFail2BanService(nil)
	return fail2banSvc.UnbanIP(jail, ip)
}

// parseNginxTimestamp 解析Nginx时间戳
func parseNginxTimestamp(timeStr string) (time.Time, error) {
	// Nginx日志时间格式: "02/Jan/2006:15:04:05 +0000"
	return time.Parse("02/Jan/2006:15:04:05 -0700", timeStr)
}

// parseNginxLogLine 解析Nginx日志行
func parseNginxLogLine(line string) *NginxLog {
	// 支持多种Nginx日志格式
	logFormats := []*regexp.Regexp{
		// 标准格式: IP - - [timestamp] "method url" status size "referer" "user-agent"
		regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) [^"]*" (\d+) (\d+) "([^"]*)" "([^"]*)"`),
		// 简化格式: IP - - [timestamp] "method url" status size
		regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*)" (\d+) (\d+)`),
		// Combined格式变体
		regexp.MustCompile(`^(\S+) - - \[([^\]]+)\] "([A-Z]+) ([^"]*) HTTP/[^"]*" (\d+) (\d+) "([^"]*)" "([^"]*)"`),
	}
	
	for _, regex := range logFormats {
		if matches := regex.FindStringSubmatch(line); matches != nil {
			log := &NginxLog{}
			
			log.IP = matches[1]
			log.Method = matches[3]
			log.URL = matches[4]
			
			// 用户代理可能不存在
			if len(matches) > 8 {
				log.UserAgent = matches[8]
			} else {
				log.UserAgent = ""
			}
			
			// 解析状态码
			if statusCode, err := strconv.Atoi(matches[5]); err == nil {
				log.StatusCode = statusCode
			}
			
			// 解析时间戳
			if timestamp, err := parseNginxTimestamp(matches[2]); err == nil {
				log.Timestamp = timestamp
			} else {
				log.Timestamp = time.Now()
			}
			
			// 检测攻击类型 (创建临时服务实例)
			tmpService := &NginxService{}
			log.AttackType = tmpService.detectAttackType(log.Method, log.URL, log.UserAgent, matches[5])
			
			// 检查是否被阻止 (4xx, 5xx状态码)
			log.IsBlocked = log.StatusCode >= 400
			
			return log
		}
	}
	
	// 如果所有格式都匹配失败，尝试简单解析IP
	if ipRegex := regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+)`); ipRegex.MatchString(line) {
		matches := ipRegex.FindStringSubmatch(line)
		return &NginxLog{
			IP:         matches[1],
			Timestamp:  time.Now(),
			StatusCode: 200,
			Method:     "GET",
			URL:        "/",
		}
	}
	
	return nil
}