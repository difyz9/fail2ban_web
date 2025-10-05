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

type SSHService struct {
	config *config.Config
	db     *gorm.DB
}

type SSHStats struct {
	TotalAttempts    int       `json:"total_attempts"`
	FailedAttempts   int       `json:"failed_attempts"`
	BannedIPs        int       `json:"banned_ips"`
	ActiveBans       int       `json:"active_bans"`
	LastAttack       time.Time `json:"last_attack"`
	TopAttackerIPs   []string  `json:"top_attacker_ips"`
	AttacksByCountry []CountryAttack `json:"attacks_by_country"`
}

type CountryAttack struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

type SSHLog struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	User      string    `json:"user"`
	Event     string    `json:"event"`
	Status    string    `json:"status"`
}

func NewSSHService(cfg *config.Config, db *gorm.DB) *SSHService {
	return &SSHService{
		config: cfg,
		db:     db,
	}
}

// GetSSHStats 获取SSH统计信息
func (s *SSHService) GetSSHStats() (*SSHStats, error) {
	stats := &SSHStats{}
	
	// 获取SSH jail状态
	bannedCount, err := s.getSSHBannedCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get banned count: %v", err)
	}
	stats.BannedIPs = bannedCount
	stats.ActiveBans = bannedCount

	// 分析SSH日志
	logStats, err := s.analyzeSSHLogs()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze SSH logs: %v", err)
	}
	
	stats.TotalAttempts = logStats.TotalAttempts
	stats.FailedAttempts = logStats.FailedAttempts
	stats.LastAttack = logStats.LastAttack
	stats.TopAttackerIPs = logStats.TopAttackerIPs

	return stats, nil
}

// getSSHBannedCount 获取SSH被禁IP数量
func (s *SSHService) getSSHBannedCount() (int, error) {
	cmd := exec.Command("fail2ban-client", "status", "sshd")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 解析输出获取被禁IP数量
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Currently banned:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				countStr := strings.TrimSpace(parts[1])
				return strconv.Atoi(countStr)
			}
		}
	}
	return 0, nil
}

// analyzeSSHLogs 分析SSH日志
func (s *SSHService) analyzeSSHLogs() (*SSHStats, error) {
	stats := &SSHStats{}
	ipCount := make(map[string]int)
	
	// 读取SSH日志文件
	file, err := os.Open("/var/log/auth.log")
	if err != nil {
		// 如果无法读取系统日志，返回空统计
		return stats, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// SSH失败登录的正则表达式
	failedRegex := regexp.MustCompile(`Failed password for (?:invalid user )?(\S+) from (\d+\.\d+\.\d+\.\d+)`)
	acceptedRegex := regexp.MustCompile(`Accepted password for (\S+) from (\d+\.\d+\.\d+\.\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// 匹配失败登录
		if matches := failedRegex.FindStringSubmatch(line); matches != nil {
			stats.FailedAttempts++
			ip := matches[2]
			ipCount[ip]++
			
			// 尝试解析时间戳
			if timestamp, err := parseLogTimestamp(line); err == nil {
				if timestamp.After(stats.LastAttack) {
					stats.LastAttack = timestamp
				}
			}
		}
		
		// 匹配成功登录
		if matches := acceptedRegex.FindStringSubmatch(line); matches != nil {
			stats.TotalAttempts++
		}
	}

	stats.TotalAttempts += stats.FailedAttempts

	// 获取攻击最多的IP
	stats.TopAttackerIPs = getTopIPs(ipCount, 5)

	return stats, nil
}

// GetSSHLogs 获取SSH日志
func (s *SSHService) GetSSHLogs(limit int) ([]SSHLog, error) {
	var logs []SSHLog
	
	// 开发模式使用测试日志
	if s.config.Fail2Ban.DevMode {
		return s.getTestSSHLogs(limit), nil
	}
	
	// 尝试多个可能的SSH日志路径
	logPaths := []string{
		"/var/log/auth.log",       // Ubuntu/Debian
		"/var/log/secure",         // CentOS/RHEL
		"/var/log/messages",       // 一些系统
		"/var/log/syslog",         // 备选路径
	}
	
	if s.config.Fail2Ban.SSHLogPath != "" {
		logPaths = append([]string{s.config.Fail2Ban.SSHLogPath}, logPaths...)
	}
	
	var file *os.File
	var err error
	var usedPath string
	
	for _, path := range logPaths {
		file, err = os.Open(path)
		if err == nil {
			usedPath = path
			break
		}
	}
	
	if file == nil {
		return logs, fmt.Errorf("no SSH log file found in paths: %v", logPaths)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	
	// 读取所有行
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return logs, fmt.Errorf("error reading SSH log file: %w", err)
	}

	// 从最后开始处理，获取最新的日志
	count := 0
	parsed := 0
	for i := len(lines) - 1; i >= 0 && count < limit; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		if log := parseSSHLogLine(line); log != nil {
			logs = append([]SSHLog{*log}, logs...)
			count++
		}
		parsed++
	}
	
	// 记录解析结果
	fmt.Printf("SSH日志解析完成: 读取%d行，解析成功%d条，文件: %s\n", parsed, count, usedPath)

	return logs, nil
}

// getTestSSHLogs 获取测试SSH日志
func (s *SSHService) getTestSSHLogs(limit int) []SSHLog {
	testLogs := []string{
		`Oct  3 22:45:01 ubuntu sshd[1234]: Failed password for admin from 192.168.1.100 port 22 ssh2`,
		`Oct  3 22:45:02 ubuntu sshd[1235]: Failed password for root from 10.0.0.1 port 22 ssh2`,
		`Oct  3 22:45:03 ubuntu sshd[1236]: Failed password for admin from 192.168.1.100 port 22 ssh2`,
		`Oct  3 22:45:04 ubuntu sshd[1237]: Failed password for test from 172.16.0.1 port 22 ssh2`,
		`Oct  3 22:45:05 ubuntu sshd[1238]: Accepted password for ubuntu from 192.168.1.50 port 22 ssh2`,
		`Oct  3 22:45:06 ubuntu sshd[1239]: Failed password for root from 203.0.113.1 port 22 ssh2`,
		`Oct  3 22:45:07 ubuntu sshd[1240]: Failed password for admin from 203.0.113.1 port 22 ssh2`,
		`Oct  3 22:45:08 ubuntu sshd[1241]: Invalid user hacker from 198.51.100.1 port 22`,
	}
	
	var logs []SSHLog
	for i, line := range testLogs {
		if i >= limit {
			break
		}
		if log := parseSSHLogLine(line); log != nil {
			logs = append(logs, *log)
		}
	}
	
	fmt.Printf("开发模式: 生成了 %d 条测试SSH日志\n", len(logs))
	return logs
}

// GetSSHJailStatus 获取SSH jail状态
func (s *SSHService) GetSSHJailStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})
	
	// 获取sshd jail状态
	sshdStatus, err := s.getJailStatus("sshd")
	if err == nil {
		status["sshd"] = sshdStatus
	}

	// 获取sshd-ddos jail状态
	sshdDdosStatus, err := s.getJailStatus("sshd-ddos")
	if err == nil {
		status["sshd-ddos"] = sshdDdosStatus
	}

	return status, nil
}

// getJailStatus 获取指定jail的状态
func (s *SSHService) getJailStatus(jailName string) (map[string]interface{}, error) {
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

// BanSSHIP 手动禁止SSH IP
func (s *SSHService) BanSSHIP(ip string, jail string) error {
	if jail == "" {
		jail = "sshd"
	}
	
	// 使用fail2ban服务的统一命令执行
	fail2banSvc := NewFail2BanService(nil)
	return fail2banSvc.BanIP(jail, ip)
}

// UnbanSSHIP 解禁SSH IP
func (s *SSHService) UnbanSSHIP(ip string, jail string) error {
	if jail == "" {
		jail = "sshd"
	}
	
	// 使用fail2ban服务的统一命令执行
	fail2banSvc := NewFail2BanService(nil)
	return fail2banSvc.UnbanIP(jail, ip)
}

// parseLogTimestamp 解析日志时间戳
func parseLogTimestamp(line string) (time.Time, error) {
	// 简单的时间戳解析，实际可能需要更复杂的逻辑
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid log format")
	}
	
	// 假设格式为 "Dec 18 10:45:23"
	timeStr := strings.Join(parts[0:3], " ")
	currentYear := time.Now().Year()
	fullTimeStr := fmt.Sprintf("%d %s", currentYear, timeStr)
	
	return time.Parse("2006 Jan 2 15:04:05", fullTimeStr)
}

// parseSSHLogLine 解析SSH日志行
func parseSSHLogLine(line string) *SSHLog {
	// SSH登录相关的正则表达式
	patterns := map[string]*regexp.Regexp{
		"failed_password": regexp.MustCompile(`Failed password for (?:invalid user )?(\S+) from (\d+\.\d+\.\d+\.\d+)`),
		"accepted_password": regexp.MustCompile(`Accepted password for (\S+) from (\d+\.\d+\.\d+\.\d+)`),
		"invalid_user": regexp.MustCompile(`Invalid user (\S+) from (\d+\.\d+\.\d+\.\d+)`),
		"disconnect": regexp.MustCompile(`Received disconnect from (\d+\.\d+\.\d+\.\d+)`),
	}

	for event, regex := range patterns {
		if matches := regex.FindStringSubmatch(line); matches != nil {
			log := &SSHLog{
				Event: event,
			}
			
			// 解析时间戳
			if timestamp, err := parseLogTimestamp(line); err == nil {
				log.Timestamp = timestamp
			}

			switch event {
			case "failed_password", "accepted_password", "invalid_user":
				if len(matches) > 2 {
					log.User = matches[1]
					log.IP = matches[2]
				}
			case "disconnect":
				if len(matches) > 1 {
					log.IP = matches[1]
				}
			}

			// 设置状态
			if strings.Contains(event, "failed") || strings.Contains(event, "invalid") {
				log.Status = "failed"
			} else if strings.Contains(event, "accepted") {
				log.Status = "success"
			} else {
				log.Status = "info"
			}

			return log
		}
	}

	return nil
}

// getTopIPs 获取攻击次数最多的IP
func getTopIPs(ipCount map[string]int, limit int) []string {
	type ipStat struct {
		ip    string
		count int
	}

	var stats []ipStat
	for ip, count := range ipCount {
		stats = append(stats, ipStat{ip: ip, count: count})
	}

	// 简单排序（冒泡排序）
	for i := 0; i < len(stats)-1; i++ {
		for j := 0; j < len(stats)-i-1; j++ {
			if stats[j].count < stats[j+1].count {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}

	var topIPs []string
	for i := 0; i < len(stats) && i < limit; i++ {
		topIPs = append(topIPs, stats[i].ip)
	}

	return topIPs
}