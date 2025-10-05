package service

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"fail2ban-web/config"
	"fail2ban-web/internal/model"
	"gorm.io/gorm"
)

// IPThreatLevel 表示IP威胁等级信息
type IPThreatLevel struct {
	IP            string    `json:"ip"`
	ThreatScore   int       `json:"threat_score"`      // 威胁评分 0-100
	SSHAttempts   int       `json:"ssh_attempts"`      // SSH攻击次数
	NginxAttempts int       `json:"nginx_attempts"`    // Nginx攻击次数
	FirstSeen     time.Time `json:"first_seen"`        // 首次发现时间
	LastSeen      time.Time `json:"last_seen"`         // 最后发现时间
	ThreatLevel   string    `json:"threat_level"`      // 威胁等级
	AttackTypes   []string  `json:"attack_types"`      // 攻击类型
	IsBanned      bool      `json:"is_banned"`         // 是否已被禁止
	AutoBanned    bool      `json:"auto_banned"`       // 是否自动禁止
	Country       string    `json:"country"`           // 国家
	ISP           string    `json:"isp"`               // ISP
}

// ScanResult 扫描结果
type ScanResult struct {
	Timestamp          time.Time         `json:"timestamp"`
	SSHThreats         []IPThreatLevel   `json:"ssh_threats"`
	NginxThreats       []IPThreatLevel   `json:"nginx_threats"`
	NewBans            []string          `json:"new_bans"`
	TotalThreats       int               `json:"total_threats"`
	HighRiskIPs        []string          `json:"high_risk_ips"`
	RecommendedActions []string          `json:"recommended_actions"`
}

// IntelligentScanService 智能扫描服务
type IntelligentScanService struct {
	config            *config.Config
	db                *gorm.DB
	sshService        *SSHService
	nginxService      *NginxService
	jailService       *JailService
	fail2banService   *Fail2BanService
	whitelistService  *WhitelistService
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	suspiciousIPs     map[string]*IPThreatLevel
	ipMutex           sync.RWMutex // 保护suspiciousIPs的并发访问
	scanInterval      time.Duration
	analysisInterval  time.Duration
	logAnalysisTicker *time.Ticker // 日志分析定时器，便于关闭
}

// NewIntelligentScanService 创建新的智能扫描服务实例
func NewIntelligentScanService(cfg *config.Config, db *gorm.DB, sshService *SSHService, 
	nginxService *NginxService, jailService *JailService, fail2banService *Fail2BanService) *IntelligentScanService {
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &IntelligentScanService{
		config:           cfg,
		db:               db,
		sshService:       sshService,
		nginxService:     nginxService,
		jailService:      jailService,
		fail2banService:  fail2banService,
		whitelistService: NewWhitelistService(),
		ctx:              ctx,
		cancel:           cancel,
		suspiciousIPs:    make(map[string]*IPThreatLevel),
		scanInterval:     5 * time.Minute,  // 5分钟扫描一次
		analysisInterval: 1 * time.Minute,  // 1分钟分析一次
	}
}

// Start 启动智能扫描服务
func (s *IntelligentScanService) Start() {
	log.Println("智能扫描服务启动...")
	
	// 启动日志扫描协程
	s.wg.Add(1)
	go s.startLogScanning()
	
	// 启动智能分析协程
	s.wg.Add(1)
	go s.startIntelligentAnalysis()
	
	// 启动自动处理协程
	s.wg.Add(1)
	go s.startAutoProcessing()
	
	// 启动自动日志分析
	s.StartAutoLogAnalysis()
	
	log.Println("智能扫描服务已启动")
}

// Stop 停止智能扫描服务
func (s *IntelligentScanService) Stop() {
	log.Println("正在停止智能扫描服务...")
	s.cancel()
	
	// 停止日志分析定时器
	if s.logAnalysisTicker != nil {
		s.logAnalysisTicker.Stop()
	}
	
	s.wg.Wait()
	log.Println("智能扫描服务已停止")
}

// startLogScanning 启动日志扫描
func (s *IntelligentScanService) startLogScanning() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.scanLogs()
		}
	}
}

// startIntelligentAnalysis 启动智能分析
func (s *IntelligentScanService) startIntelligentAnalysis() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.analysisInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.analyzeThreats()
		}
	}
}

// startAutoProcessing 启动自动处理
func (s *IntelligentScanService) startAutoProcessing() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // 30秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.autoProcessThreats()
		}
	}
}

// scanLogs 扫描日志
func (s *IntelligentScanService) scanLogs() {
	log.Println("开始扫描日志...")
	
	// 扫描SSH日志
	sshCount := s.scanSSHLogs()
	
	// 扫描Nginx日志
	nginxCount := s.scanNginxLogs()
	
	// 清理过期的威胁数据
	s.cleanupOldThreats()
	
	s.ipMutex.RLock()
	defer s.ipMutex.RUnlock()
	log.Printf("日志扫描完成，SSH发现%d个事件，Nginx发现%d个事件，当前监控 %d 个可疑IP", 
		sshCount, nginxCount, len(s.suspiciousIPs))
}

// cleanupOldThreats 清理过期的威胁数据（超过24小时）
func (s *IntelligentScanService) cleanupOldThreats() {
	s.ipMutex.Lock()
	defer s.ipMutex.Unlock()
	
	expirationTime := time.Now().Add(-24 * time.Hour)
	for ip, threat := range s.suspiciousIPs {
		if threat.LastSeen.Before(expirationTime) {
			delete(s.suspiciousIPs, ip)
		}
	}
}

// scanSSHLogs 扫描SSH日志
func (s *IntelligentScanService) scanSSHLogs() int {
	if s.sshService == nil {
		log.Println("SSH服务未初始化，跳过SSH日志扫描")
		return 0
	}
	
	eventCount := 0
	logs, err := s.sshService.GetSSHLogs(200) // 获取最近200条日志
	if err != nil {
		log.Printf("获取SSH日志失败: %v", err)
		return eventCount
	}
	
	log.Printf("获取到 %d 条SSH日志", len(logs))
	
	for _, logEntry := range logs {
		if logEntry.Status == "failed" && logEntry.IP != "" {
			// 检查是否在白名单中
			if s.whitelistService.IsWhitelisted(logEntry.IP) {
				continue
			}
			
			s.updateThreatLevel(logEntry.IP, "ssh", logEntry.Event, logEntry.Timestamp)
			eventCount++
		}
	}
	
	return eventCount
}

// scanNginxLogs 扫描Nginx日志
func (s *IntelligentScanService) scanNginxLogs() int {
	if s.nginxService == nil {
		log.Println("Nginx服务未初始化，跳过Nginx日志扫描")
		return 0
	}
	
	eventCount := 0
	logs, err := s.nginxService.GetNginxLogs(200) // 获取最近200条日志
	if err != nil {
		log.Printf("获取Nginx日志失败: %v", err)
		return eventCount
	}
	
	log.Printf("获取到 %d 条Nginx日志", len(logs))
	
	for _, logEntry := range logs {
		// 检查是否有攻击迹象且不在白名单中
		if logEntry.IP != "" && !s.whitelistService.IsWhitelisted(logEntry.IP) {
			if logEntry.AttackType != "" {
				s.updateThreatLevel(logEntry.IP, "nginx", logEntry.AttackType, logEntry.Timestamp)
				eventCount++
			} else if logEntry.StatusCode >= 400 {
				// 4xx, 5xx错误也算可疑
				s.updateThreatLevel(logEntry.IP, "nginx", "http_error", logEntry.Timestamp)
				eventCount++
			}
		}
	}
	
	return eventCount
}

// updateThreatLevel 更新威胁等级
func (s *IntelligentScanService) updateThreatLevel(ip, source, attackType string, timestamp time.Time) {
	s.ipMutex.Lock()
	defer s.ipMutex.Unlock()
	
	threat, exists := s.suspiciousIPs[ip]
	if !exists {
		threat = &IPThreatLevel{
			IP:          ip,
			ThreatScore: 0,
			AttackTypes: []string{},
			FirstSeen:   timestamp,
			LastSeen:    timestamp,
		}
		s.suspiciousIPs[ip] = threat
	}
	
	// 更新最后发现时间
	if timestamp.After(threat.LastSeen) {
		threat.LastSeen = timestamp
	}
	
	// 增加攻击计数和威胁评分
	switch source {
	case "ssh":
		threat.SSHAttempts++
		threat.ThreatScore += s.calculateSSHThreatScore(attackType)
	case "nginx":
		threat.NginxAttempts++
		threat.ThreatScore += s.calculateNginxThreatScore(attackType)
	}
	
	// 添加攻击类型（避免重复）
	if !contains(threat.AttackTypes, attackType) {
		threat.AttackTypes = append(threat.AttackTypes, attackType)
	}
	
	// 限制威胁评分最大值
	if threat.ThreatScore > 100 {
		threat.ThreatScore = 100
	}
	
	// 更新威胁等级描述
	threat.ThreatLevel = s.getThreatLevelDescription(threat.ThreatScore)
}

// getThreatLevelDescription 将威胁评分转换为威胁等级描述
func (s *IntelligentScanService) getThreatLevelDescription(score int) string {
	switch {
	case score >= 80:
		return "严重"
	case score >= 60:
		return "高危"
	case score >= 40:
		return "中危"
	case score >= 20:
		return "低危"
	default:
		return "可疑"
	}
}

// calculateSSHThreatScore 计算SSH威胁评分
func (s *IntelligentScanService) calculateSSHThreatScore(attackType string) int {
	scores := map[string]int{
		"failed_password":        10,
		"invalid_user":           15,
		"disconnect":             5,
		"authentication_failure": 12,
	}
	
	if score, exists := scores[attackType]; exists {
		return score
	}
	return 8 // 默认评分
}

// calculateNginxThreatScore 计算Nginx威胁评分
func (s *IntelligentScanService) calculateNginxThreatScore(attackType string) int {
	scores := map[string]int{
		"sql_injection":   25,
		"xss":             20,
		"path_traversal":  20,
		"malicious_bot":   15,
		"directory_scan":  12,
		"auth_failure":    10,
		"rate_limit":      8,
		"http_error":      5,
	}
	
	if score, exists := scores[attackType]; exists {
		return score
	}
	return 10 // 默认评分
}

// analyzeThreats 分析威胁
func (s *IntelligentScanService) analyzeThreats() {
	s.ipMutex.RLock()
	defer s.ipMutex.RUnlock()
	
	highRiskCount := 0
	mediumRiskCount := 0
	
	for _, threat := range s.suspiciousIPs {
		// 只分析24小时内的威胁
		if time.Since(threat.LastSeen) > 24*time.Hour {
			continue
		}
		
		if threat.ThreatScore >= 80 {
			highRiskCount++
		} else if threat.ThreatScore >= 50 {
			mediumRiskCount++
		}
	}
	
	if highRiskCount > 0 || mediumRiskCount > 0 {
		log.Printf("威胁分析: 高风险IP %d 个, 中风险IP %d 个", highRiskCount, mediumRiskCount)
	}
}

// autoProcessThreats 自动处理威胁
func (s *IntelligentScanService) autoProcessThreats() {
	s.ipMutex.Lock()
	defer s.ipMutex.Unlock()
	
	bannedCount := 0
	errorCount := 0
	processedCount := 0
	
	for ip, threat := range s.suspiciousIPs {
		// 跳过已经处理的IP
		if threat.IsBanned || threat.AutoBanned {
			continue
		}
		
		// 自动封禁高威胁IP
		if s.shouldAutoBan(threat) {
			processedCount++
			if err := s.autoBanIP(ip, threat); err != nil {
				log.Printf("自动封禁IP %s 失败: %v", ip, err)
				errorCount++
			} else {
				threat.AutoBanned = true
				threat.IsBanned = true
				bannedCount++
				log.Printf("成功自动封禁高威胁IP: %s (威胁评分: %d)", ip, threat.ThreatScore)
			}
		}
	}
	
	if processedCount > 0 {
		log.Printf("自动处理完成: 处理 %d 个IP, 成功封禁 %d 个, 失败 %d 个", processedCount, bannedCount, errorCount)
	}
}

// shouldAutoBan 判断是否应该自动封禁
func (s *IntelligentScanService) shouldAutoBan(threat *IPThreatLevel) bool {
	// 高威胁评分自动封禁
	if threat.ThreatScore >= 80 {
		return true
	}
	
	// SSH暴力破解自动封禁
	if threat.SSHAttempts >= 10 {
		return true
	}
	
	// 多种攻击类型自动封禁
	if len(threat.AttackTypes) >= 3 && threat.ThreatScore >= 60 {
		return true
	}
	
	// SQL注入等严重攻击立即封禁
	for _, attackType := range threat.AttackTypes {
		if attackType == "sql_injection" || attackType == "xss" || attackType == "path_traversal" {
			return true
		}
	}
	
	return false
}

// autoBanIP 自动封禁IP
func (s *IntelligentScanService) autoBanIP(ip string, threat *IPThreatLevel) error {
	if s.fail2banService == nil {
		return fmt.Errorf("fail2ban服务未初始化")
	}
	
	// 首先检查白名单
	if s.whitelistService.IsWhitelisted(ip) {
		log.Printf("[安全] IP %s 在白名单中，跳过自动封禁", ip)
		return nil
	}
	
	// 检查是否已经被封禁
	var existingBan model.BannedIP
	if err := s.db.Where("ip_address = ? AND is_active = ?", ip, true).First(&existingBan).Error; err == nil {
		log.Printf("IP %s 已经被封禁，跳过", ip)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查IP是否已封禁失败: %w", err)
	}
	
	// 获取当前可用的jails
	availableJails, err := s.fail2banService.GetJails()
	if err != nil {
		return fmt.Errorf("获取jail列表失败: %w", err)
	}
	
	if len(availableJails) == 0 {
		return fmt.Errorf("没有可用的jail进行封禁")
	}
	
	// 选择合适的jail进行封禁
	jailUsed := ""
	
	// 优先使用SSH相关的jail (如果有SSH攻击)
	if threat.SSHAttempts > 0 {
		for _, jail := range availableJails {
			if jail == "sshd" || jail == "sshd-ddos" {
				if err := s.fail2banService.BanIP(jail, ip); err != nil {
					log.Printf("在jail %s 中封禁IP %s 失败: %v", jail, ip, err)
					continue
				}
				jailUsed = jail
				break
			}
		}
	}
	
	// 如果还没有成功封禁，尝试使用nginx相关的jail
	if jailUsed == "" && threat.NginxAttempts > 0 {
		for _, jail := range availableJails {
			if jail == "nginx-http-auth" || strings.Contains(jail, "nginx") {
				if err := s.fail2banService.BanIP(jail, ip); err != nil {
					log.Printf("在jail %s 中封禁IP %s 失败: %v", jail, ip, err)
					continue
				}
				jailUsed = jail
				break
			}
		}
	}
	
	// 如果还没有成功封禁，尝试使用第一个可用的jail
	if jailUsed == "" {
		jail := availableJails[0]
		if err := s.fail2banService.BanIP(jail, ip); err != nil {
			return fmt.Errorf("在jail %s 中封禁IP失败: %w", jail, err)
		}
		jailUsed = jail
	}
	
	// 记录到数据库
	bannedIP := &model.BannedIP{
		IPAddress: ip,
		Jail:      jailUsed,
		BanTime:   time.Now(),
		// 使用配置中的封禁时长，如果未配置则默认24小时
		UnbanTime: time.Now().Add(s.getBanDuration()),
		IsActive:  true,
		Reason:    s.generateBanReason(threat),
	}
	
	log.Printf("成功在jail %s 中封禁IP %s", jailUsed, ip)
	return s.db.Create(bannedIP).Error
}

// getBanDuration 获取封禁时长，使用默认值
func (s *IntelligentScanService) getBanDuration() time.Duration {
	// 默认24小时封禁时长
	return 24 * time.Hour
}

// IsIPWhitelisted 检查IP是否在白名单中（公开方法用于测试）
func (s *IntelligentScanService) IsIPWhitelisted(ip string) bool {
	if s.whitelistService == nil {
		return false
	}
	return s.whitelistService.IsWhitelisted(ip)
}

// generateBanReason 生成封禁原因
func (s *IntelligentScanService) generateBanReason(threat *IPThreatLevel) string {
	reason := fmt.Sprintf("智能检测: 威胁评分 %d", threat.ThreatScore)
	
	if threat.SSHAttempts > 0 {
		reason += fmt.Sprintf(", SSH攻击 %d 次", threat.SSHAttempts)
	}
	
	if threat.NginxAttempts > 0 {
		reason += fmt.Sprintf(", Web攻击 %d 次", threat.NginxAttempts)
	}
	
	if len(threat.AttackTypes) > 0 {
		reason += fmt.Sprintf(", 攻击类型: %v", threat.AttackTypes)
	}
	
	return reason
}

// GetCurrentThreats 获取当前威胁
func (s *IntelligentScanService) GetCurrentThreats() map[string]*IPThreatLevel {
	s.ipMutex.RLock()
	defer s.ipMutex.RUnlock()
	
	// 复制威胁信息，避免外部修改内部数据
	threats := make(map[string]*IPThreatLevel)
	for ip, threat := range s.suspiciousIPs {
		// 只返回24小时内的威胁
		if time.Since(threat.LastSeen) <= 24*time.Hour {
			// 深拷贝，防止外部修改
			clone := *threat
			clone.AttackTypes = append([]string(nil), threat.AttackTypes...)
			threats[ip] = &clone
		}
	}
	
	return threats
}

// GetScanResult 获取扫描结果
func (s *IntelligentScanService) GetScanResult() *ScanResult {
	threats := s.GetCurrentThreats()
	
	result := &ScanResult{
		Timestamp:          time.Now(),
		SSHThreats:         []IPThreatLevel{},
		NginxThreats:       []IPThreatLevel{},
		NewBans:            []string{},
		HighRiskIPs:        []string{},
		RecommendedActions: []string{},
	}
	
	for _, threat := range threats {
		result.TotalThreats++
		
		if threat.SSHAttempts > 0 {
			result.SSHThreats = append(result.SSHThreats, *threat)
		}
		
		if threat.NginxAttempts > 0 {
			result.NginxThreats = append(result.NginxThreats, *threat)
		}
		
		if threat.ThreatScore >= 80 {
			result.HighRiskIPs = append(result.HighRiskIPs, threat.IP)
		}
		
		if threat.AutoBanned && time.Since(threat.LastSeen) < s.analysisInterval {
			result.NewBans = append(result.NewBans, threat.IP)
		}
	}
	
	// 生成建议
	result.RecommendedActions = s.generateRecommendations(result)
	
	return result
}

// generateRecommendations 生成建议
func (s *IntelligentScanService) generateRecommendations(result *ScanResult) []string {
	var recommendations []string
	
	if len(result.HighRiskIPs) > 0 {
		recommendations = append(recommendations, "检测到高风险IP，建议立即检查")
	}
	
	if len(result.SSHThreats) > 5 {
		recommendations = append(recommendations, "SSH攻击频繁，建议更换SSH端口或启用密钥认证")
	}
	
	if len(result.NginxThreats) > 10 {
		recommendations = append(recommendations, "Web攻击较多，建议启用WAF或限流")
	}
	
	if len(result.NewBans) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("已自动封禁 %d 个恶意IP", len(result.NewBans)))
	}
	
	return recommendations
}

// ManualBanIP 手动封禁IP
func (s *IntelligentScanService) ManualBanIP(ip, reason string) error {
	if ip == "" {
		return fmt.Errorf("IP地址不能为空")
	}
	
	// 检查是否在白名单中
	if s.whitelistService.IsWhitelisted(ip) {
		return fmt.Errorf("IP %s 在白名单中，无法手动封禁", ip)
	}
	
	// 更新威胁信息
	s.ipMutex.Lock()
	threat, exists := s.suspiciousIPs[ip]
	if !exists {
		threat = &IPThreatLevel{
			IP:          ip,
			ThreatScore: 100, // 手动封禁给最高评分
			LastSeen:    time.Now(),
			FirstSeen:   time.Now(),
		}
		s.suspiciousIPs[ip] = threat
	}
	threat.IsBanned = true
	threat.ThreatLevel = "严重"
	s.ipMutex.Unlock()
	
	// 执行封禁
	if s.sshService != nil {
		if err := s.sshService.BanSSHIP(ip, "sshd"); err != nil {
			log.Printf("SSH封禁失败: %v", err)
			// 不返回错误，继续尝试其他封禁方式
		}
	}
	
	if s.nginxService != nil {
		if err := s.nginxService.BanNginxIP(ip, "nginx-http-auth"); err != nil {
			log.Printf("Nginx封禁失败: %v", err)
			// 不返回错误，继续尝试其他封禁方式
		}
	}
	
	// 记录到数据库
	bannedIP := &model.BannedIP{
		IPAddress: ip,
		Jail:      "manual",
		BanTime:   time.Now(),
		UnbanTime: time.Now().Add(s.getBanDuration()),
		IsActive:  true,
		Reason:    reason,
	}
	
	return s.db.Create(bannedIP).Error
}

// contains 检查字符串数组是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// AnalyzeLogFile 分析指定的日志文件并自动封禁恶意IP
func (s *IntelligentScanService) AnalyzeLogFile(logFilePath string) error {
	if logFilePath == "" {
		return fmt.Errorf("日志文件路径不能为空")
	}
	
	log.Printf("开始分析日志文件: %s", logFilePath)
	
	file, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("无法打开日志文件: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理长日志行
	buf := make([]byte, 1024*1024) // 1MB
	scanner.Buffer(buf, 1024*1024)
	
	// Nginx日志格式正则
	logRegex := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) HTTP/\S+" (\d+) \d+ "([^"]*)" "([^"]*)"`)
	
	maliciousIPs := make(map[string]*IPThreatLevel)
	totalLines := 0
	processedLines := 0
	
	for scanner.Scan() {
		totalLines++
		line := scanner.Text()
		matches := logRegex.FindStringSubmatch(line)
		
		if len(matches) >= 8 {
			processedLines++
			ip := matches[1]
			timeStr := matches[2]
			method := matches[3]
			url := matches[4]
			statusCode := matches[5]
			userAgent := matches[7]
			
			// 检查白名单
			if s.whitelistService.IsWhitelisted(ip) {
				continue
			}
			
			// 解析时间
			t, err := time.Parse("02/Jan/2006:15:04:05 -0700", timeStr)
			if err != nil {
				continue
			}
			
			// 检测攻击类型
			attackType := s.detectLogAttackType(method, url, userAgent, statusCode)
			
			if attackType != "" {
				// 如果是恶意请求，记录到威胁分析中
				if maliciousIPs[ip] == nil {
					maliciousIPs[ip] = &IPThreatLevel{
						IP:            ip,
						FirstSeen:     t,
						LastSeen:      t,
						ThreatScore:   0,
						AttackTypes:   []string{},
						SSHAttempts:   0,
						NginxAttempts: 1,
						ThreatLevel:   "",
						IsBanned:      false,
						AutoBanned:    false,
					}
				} else {
					maliciousIPs[ip].NginxAttempts++
					if t.After(maliciousIPs[ip].LastSeen) {
						maliciousIPs[ip].LastSeen = t
					}
					if t.Before(maliciousIPs[ip].FirstSeen) {
						maliciousIPs[ip].FirstSeen = t
					}
				}
				
				// 添加攻击类型
				if !contains(maliciousIPs[ip].AttackTypes, attackType) {
					maliciousIPs[ip].AttackTypes = append(maliciousIPs[ip].AttackTypes, attackType)
				}
				
				// 更新威胁评分
				maliciousIPs[ip].ThreatScore += s.getAttackTypeScore(attackType)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取日志文件失败: %w", err)
	}
	
	log.Printf("日志分析完成: 总行数 %d, 处理行数 %d, 发现恶意IP %d", totalLines, processedLines, len(maliciousIPs))
	
	// 分析威胁等级并自动封禁
	bannedCount := 0
	errorCount := 0
	
	// 将分析结果合并到主威胁列表
	s.ipMutex.Lock()
	for ip, threat := range maliciousIPs {
		// 评估威胁等级
		s.evaluateLogThreatLevel(threat)
		
		// 合并到主列表
		if existingThreat, exists := s.suspiciousIPs[ip]; exists {
			existingThreat.ThreatScore += threat.ThreatScore
			if existingThreat.ThreatScore > 100 {
				existingThreat.ThreatScore = 100
			}
			existingThreat.NginxAttempts += threat.NginxAttempts
			existingThreat.LastSeen = threat.LastSeen
			existingThreat.ThreatLevel = s.getThreatLevelDescription(existingThreat.ThreatScore)
			
			// 合并攻击类型
			for _, atkType := range threat.AttackTypes {
				if !contains(existingThreat.AttackTypes, atkType) {
					existingThreat.AttackTypes = append(existingThreat.AttackTypes, atkType)
				}
			}
		} else {
			s.suspiciousIPs[ip] = threat
		}
	}
	s.ipMutex.Unlock()
	
	// 处理需要封禁的IP
	for ip, threat := range maliciousIPs {
		// 检查是否已经被封禁
		var existingBan model.BannedIP
		if err := s.db.Where("ip_address = ? AND is_active = ?", ip, true).First(&existingBan).Error; err == nil {
			log.Printf("IP %s 已经被封禁，跳过", ip)
			continue
		}
		
		// 自动封禁高危和严重威胁
		if threat.ThreatLevel == "高危" || threat.ThreatLevel == "严重" {
			if err := s.autoBanIP(ip, threat); err != nil {
				log.Printf("自动封禁IP %s 失败: %v", ip, err)
				errorCount++
			} else {
				log.Printf("成功自动封禁恶意IP: %s (威胁等级: %s, 攻击类型: %s)", 
					ip, threat.ThreatLevel, strings.Join(threat.AttackTypes, ","))
				bannedCount++
			}
		} else {
			log.Printf("发现可疑IP: %s (威胁等级: %s, 攻击类型: %s)", 
				ip, threat.ThreatLevel, strings.Join(threat.AttackTypes, ","))
		}
	}
	
	log.Printf("日志文件分析完成: 成功封禁 %d 个IP, 失败 %d 个", bannedCount, errorCount)
	return nil
}

// detectLogAttackType 检测日志中的攻击类型
func (s *IntelligentScanService) detectLogAttackType(method, url, userAgent, statusCode string) string {
	url = strings.ToLower(url)
	userAgent = strings.ToLower(userAgent)
	method = strings.ToLower(method)
	
	var attackTypes []string
	
	// WordPress攻击检测
	if strings.Contains(url, "/wp-admin/setup-config.php") {
		attackTypes = append(attackTypes, "wordpress_exploitation")
	}
	if strings.Contains(url, "/wordpress/wp-admin/") || strings.Contains(url, "/wp-admin/") {
		attackTypes = append(attackTypes, "wordpress_scan")
	}
	if strings.Contains(url, "/wp-content/") || strings.Contains(url, "/wp-includes/") {
		attackTypes = append(attackTypes, "wordpress_file_access")
	}
	
	// 管理面板攻击检测
	if strings.Contains(url, "/admin/config.php") {
		attackTypes = append(attackTypes, "admin_config_exploit")
	}
	if strings.Contains(url, "/admin/login.php") || strings.Contains(url, "login.asp") {
		attackTypes = append(attackTypes, "admin_login_scan")
	}
	if strings.Contains(url, "/boaform/admin/formlogin") {
		attackTypes = append(attackTypes, "router_admin_exploit")
	}
	
	// PHP文件扫描检测
	if strings.Contains(url, ".php") {
		attackTypes = append(attackTypes, "php_file_scan")
	}
	
	// 路由器/IoT设备攻击检测
	if strings.Contains(url, "/cgi-bin/luci/") {
		attackTypes = append(attackTypes, "router_exploit")
	}
	if strings.Contains(url, "/manager/text/list") {
		attackTypes = append(attackTypes, "tomcat_manager_scan")
	}
	
	// 代理滥用检测
	if method == "connect" && strings.Contains(url, ":443") {
		attackTypes = append(attackTypes, "proxy_abuse")
	}
	if method == "propfind" {
		attackTypes = append(attackTypes, "webdav_scan")
	}
	
	// 可疑User-Agent检测
	if strings.Contains(userAgent, "xfa1") || strings.Contains(userAgent, "zgrab") {
		attackTypes = append(attackTypes, "malicious_scanner")
	}
	
	// 去重并返回
	if len(attackTypes) > 0 {
		return strings.Join(attackTypes, ",")
	}
	
	return ""
}

// getAttackTypeScore 获取攻击类型的威胁评分
func (s *IntelligentScanService) getAttackTypeScore(attackType string) int {
	scores := map[string]int{
		"wordpress_exploitation": 15,
		"admin_config_exploit":   12,
		"router_exploit":         12,
		"router_admin_exploit":   10,
		"wordpress_scan":         8,
		"wordpress_file_access":  6,
		"admin_login_scan":       8,
		"php_file_scan":          5,
		"tomcat_manager_scan":    8,
		"proxy_abuse":            7,
		"webdav_scan":            6,
		"malicious_scanner":      5,
	}
	
	// 处理组合攻击类型（用逗号分隔）
	if strings.Contains(attackType, ",") {
		totalScore := 0
		attackTypes := strings.Split(attackType, ",")
		for _, singleType := range attackTypes {
			singleType = strings.TrimSpace(singleType)
			if score, exists := scores[singleType]; exists {
				totalScore += score
			} else {
				totalScore += 3 // 未知攻击类型默认分数
			}
		}
		return totalScore
	}
	
	if score, exists := scores[attackType]; exists {
		return score
	}
	return 3 // 默认分数
}

// evaluateLogThreatLevel 评估日志分析的威胁等级
func (s *IntelligentScanService) evaluateLogThreatLevel(threat *IPThreatLevel) {
	score := threat.ThreatScore
	
	// 基于攻击次数加分
	if threat.NginxAttempts > 10 {
		score += 10
	} else if threat.NginxAttempts > 5 {
		score += 5
	} else if threat.NginxAttempts > 1 {
		score += 2
	}
	
	// 基于攻击类型多样性加分
	if len(threat.AttackTypes) > 3 {
		score += 8
	} else if len(threat.AttackTypes) > 1 {
		score += 4
	}
	
	// 限制最高分数
	if score > 100 {
		score = 100
	}
	
	// 确定威胁等级
	threat.ThreatScore = score
	threat.ThreatLevel = s.getThreatLevelDescription(score)
}

// AnalyzeAccessLog 分析access.log文件并自动处理威胁
func (s *IntelligentScanService) AnalyzeAccessLog() error {
	// 默认的access.log路径
	accessLogPaths := []string{
		"/var/log/nginx/access.log",
		"/usr/local/nginx/logs/access.log", 
		"/var/log/apache2/access.log",
		"/var/log/httpd/access_log",
	}
	
	var logFile string
	for _, path := range accessLogPaths {
		if _, err := os.Stat(path); err == nil {
			logFile = path
			break
		}
	}
	
	if logFile == "" {
		log.Printf("未找到access.log文件，跳过自动分析")
		return nil
	}
	
	log.Printf("开始自动分析access.log: %s", logFile)
	return s.AnalyzeLogFile(logFile)
}

// StartAutoLogAnalysis 启动自动日志分析
func (s *IntelligentScanService) StartAutoLogAnalysis() {
	s.wg.Add(1)
	go s.autoLogAnalysis()
}

// autoLogAnalysis 自动日志分析任务
func (s *IntelligentScanService) autoLogAnalysis() {
	defer s.wg.Done()
	
	// 每30分钟分析一次access.log
	s.logAnalysisTicker = time.NewTicker(30 * time.Minute)
	defer s.logAnalysisTicker.Stop()
	
	// 立即执行一次分析
	if err := s.AnalyzeAccessLog(); err != nil {
		log.Printf("初始日志分析失败: %v", err)
	}
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.logAnalysisTicker.C:
			if err := s.AnalyzeAccessLog(); err != nil {
				log.Printf("自动日志分析失败: %v", err)
			}
		}
	}
}
