package service

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fail2ban-web/internal/model"

	"github.com/sirupsen/logrus"
)

type Fail2BanService struct {
	logger  *logrus.Logger
	useSudo bool
}

func NewFail2BanService(logger *logrus.Logger) *Fail2BanService {
	// 检查是否需要使用sudo
	useSudo := shouldUseSudo()
	
	service := &Fail2BanService{
		logger:  logger,
		useSudo: useSudo,
	}
	
	// 记录权限状态
	if useSudo {
		logger.Info("Fail2Ban service will use sudo for privileged operations")
		logger.Warn("Make sure the application user has sudo access to fail2ban-client")
	} else {
		logger.Info("Fail2Ban service running with direct privileges")
	}
	
	// 测试连接 (不阻止服务启动)
	if err := service.TestConnection(); err != nil {
		logger.WithError(err).Warn("Failed to connect to Fail2Ban service - service will continue with limited functionality")
	} else {
		logger.Info("Successfully connected to Fail2Ban service")
	}
	
	return service
}

// shouldUseSudo 检查是否需要使用sudo
func shouldUseSudo() bool {
	// 检查当前用户是否为root
	if os.Getuid() == 0 {
		return false
	}
	
	// 检查fail2ban socket是否可访问
	socketPath := "/var/run/fail2ban/fail2ban.sock"
	if _, err := os.Stat(socketPath); err == nil {
		// 尝试简单的fail2ban-client命令
		cmd := exec.Command("fail2ban-client", "ping")
		if err := cmd.Run(); err == nil {
			return false
		}
	}
	
	return true
}

// execFail2banCommand 执行fail2ban命令，如果需要会使用sudo
func (s *Fail2BanService) execFail2banCommand(args ...string) ([]byte, error) {
	var cmd *exec.Cmd
	
	if s.useSudo {
		// 使用sudo执行命令
		sudoArgs := append([]string{"fail2ban-client"}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
	} else {
		cmd = exec.Command("fail2ban-client", args...)
	}
	
	s.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"use_sudo": s.useSudo,
	}).Debug("Executing fail2ban command")
	
	return cmd.Output()
}

// execFail2banCommandCombined 执行fail2ban命令并返回合并输出
func (s *Fail2BanService) execFail2banCommandCombined(args ...string) ([]byte, error) {
	var cmd *exec.Cmd
	
	if s.useSudo {
		sudoArgs := append([]string{"fail2ban-client"}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
	} else {
		cmd = exec.Command("fail2ban-client", args...)
	}
	
	s.logger.WithFields(logrus.Fields{
		"command": cmd.String(),
		"use_sudo": s.useSudo,
	}).Debug("Executing fail2ban command (combined output)")
	
	return cmd.CombinedOutput()
}

// TestConnection 测试与Fail2Ban的连接
func (s *Fail2BanService) TestConnection() error {
	_, err := s.execFail2banCommand("ping")
	if err != nil {
		return fmt.Errorf("failed to ping fail2ban server: %w", err)
	}
	return nil
}

// GetPermissionStatus 获取权限状态信息
func (s *Fail2BanService) GetPermissionStatus() map[string]interface{} {
	status := map[string]interface{}{
		"using_sudo": s.useSudo,
		"user_id":    os.Getuid(),
		"user_name":  os.Getenv("USER"),
	}
	
	// 测试连接状态
	if err := s.TestConnection(); err != nil {
		status["connection_status"] = "failed"
		status["connection_error"] = err.Error()
	} else {
		status["connection_status"] = "ok"
	}
	
	return status
}

// GetStatus 获取 Fail2Ban 状态
func (s *Fail2BanService) GetStatus() (map[string]interface{}, error) {
	output, err := s.execFail2banCommand("status")
	if err != nil {
		s.logger.WithError(err).Error("Failed to get fail2ban status")
		return nil, fmt.Errorf("failed to get fail2ban status: %w", err)
	}

	status := make(map[string]interface{})
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Number of jail:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				status["jail_count"] = count
			}
		}
		if strings.Contains(line, "Jail list:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				jails := strings.Split(strings.TrimSpace(parts[1]), ",")
				var jailList []string
				for _, jail := range jails {
					jailList = append(jailList, strings.TrimSpace(jail))
				}
				status["jails"] = jailList
			}
		}
	}

	return status, nil
}

// GetVersion 获取 Fail2Ban 版本
func (s *Fail2BanService) GetVersion() (string, error) {
	output, err := s.execFail2banCommand("version")
	if err != nil {
		s.logger.WithError(err).Error("Failed to get fail2ban version")
		return "", fmt.Errorf("failed to get fail2ban version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// GetBannedIPs 获取被禁IP列表
func (s *Fail2BanService) GetBannedIPs() ([]model.BannedIPResponse, error) {
	status, err := s.GetStatus()
	if err != nil {
		return nil, err
	}

	jails, ok := status["jails"].([]string)
	if !ok {
		return []model.BannedIPResponse{}, nil
	}

	var bannedIPs []model.BannedIPResponse

	for _, jail := range jails {
		ips, err := s.GetBannedIPsForJail(jail)
		if err != nil {
			s.logger.WithError(err).WithField("jail", jail).Warn("Failed to get banned IPs for jail")
			continue
		}
		bannedIPs = append(bannedIPs, ips...)
	}

	return bannedIPs, nil
}

// GetBannedIPsForJail 获取指定jail的被禁IP列表
func (s *Fail2BanService) GetBannedIPsForJail(jail string) ([]model.BannedIPResponse, error) {
	output, err := s.execFail2banCommand("status", jail)
	if err != nil {
		return nil, fmt.Errorf("failed to get banned IPs for jail %s: %w", jail, err)
	}

	var bannedIPs []model.BannedIPResponse
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Banned IP list:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				ipList := strings.TrimSpace(parts[1])
				if ipList != "" {
					ips := strings.Split(ipList, " ")
					for _, ip := range ips {
						ip = strings.TrimSpace(ip)
						if ip != "" {
							// 获取ban时间信息
							banTime, remainingTime := s.getBanTimeInfo(jail, ip)
							bannedIPs = append(bannedIPs, model.BannedIPResponse{
								Address:       ip,
								Jail:          jail,
								BanTime:       banTime,
								RemainingTime: remainingTime,
							})
						}
					}
				}
			}
		}
	}

	return bannedIPs, nil
}

// getBanTimeInfo 获取IP的ban时间信息
func (s *Fail2BanService) getBanTimeInfo(jail, ip string) (time.Time, int64) {
	// 这里可以通过读取fail2ban日志文件来获取更精确的时间信息
	// 现在返回默认值
	return time.Now().Add(-time.Hour), 3600 // 假设1小时前被禁，剩余1小时
}

// UnbanIP 解禁IP
func (s *Fail2BanService) UnbanIP(jail, ip string) error {
	output, err := s.execFail2banCommandCombined("set", jail, "unbanip", ip)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"jail": jail,
			"ip":   ip,
			"output": string(output),
		}).Error("Failed to unban IP")
		return fmt.Errorf("failed to unban IP %s from jail %s: %w", ip, jail, err)
	}

	s.logger.WithFields(logrus.Fields{
		"jail": jail,
		"ip":   ip,
	}).Info("Successfully unbanned IP")

	return nil
}

// BanIP 手动禁止IP
func (s *Fail2BanService) BanIP(jail, ip string) error {
	output, err := s.execFail2banCommandCombined("set", jail, "banip", ip)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"jail": jail,
			"ip":   ip,
			"output": string(output),
		}).Error("Failed to ban IP")
		return fmt.Errorf("failed to ban IP %s in jail %s: %w", ip, jail, err)
	}

	s.logger.WithFields(logrus.Fields{
		"jail": jail,
		"ip":   ip,
	}).Info("Successfully banned IP")

	return nil
}

// GetJails 获取jail列表
func (s *Fail2BanService) GetJails() ([]string, error) {
	status, err := s.GetStatus()
	if err != nil {
		return nil, err
	}

	jails, ok := status["jails"].([]string)
	if !ok {
		return []string{}, nil
	}

	return jails, nil
}

// GetJailStatus 获取指定jail的详细状态
func (s *Fail2BanService) GetJailStatus(jail string) (map[string]interface{}, error) {
	output, err := s.execFail2banCommand("status", jail)
	if err != nil {
		return nil, fmt.Errorf("failed to get jail status for %s: %w", jail, err)
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
				
				// 尝试解析数值
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

// GetStats 获取统计信息
func (s *Fail2BanService) GetStats() (*model.StatsResponse, error) {
	stats := &model.StatsResponse{
		SystemStatus: "运行中",
	}

	// 获取被禁IP总数
	bannedIPs, err := s.GetBannedIPs()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get banned IPs for stats")
	} else {
		stats.BannedCount = len(bannedIPs)
	}

	// 获取活跃规则数
	jails, err := s.GetJails()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get jails for stats")
	} else {
		stats.ActiveRules = len(jails)
	}

	// 获取今日拦截数（这里需要解析日志文件）
	stats.TodayBlocks = s.getTodayBlocks()

	return stats, nil
}

// getTodayBlocks 获取今日拦截数量
func (s *Fail2BanService) getTodayBlocks() int {
	// 这里应该解析fail2ban日志文件来统计今日的拦截次数
	// 现在返回一个模拟值
	return 42
}

// GetSystemInfo 获取系统信息
func (s *Fail2BanService) GetSystemInfo() (*model.SystemInfoResponse, error) {
	info := &model.SystemInfoResponse{}

	// 获取版本
	version, err := s.GetVersion()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get fail2ban version")
		info.Version = "Unknown"
	} else {
		info.Version = version
	}

	// 获取运行时间（模拟值）
	info.Uptime = 86400 // 1天

	// 获取被禁IP数量
	bannedIPs, err := s.GetBannedIPs()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get banned IPs count")
		info.BannedIPs = 0
	} else {
		info.BannedIPs = len(bannedIPs)
	}

	// 获取活跃jail数量
	jails, err := s.GetJails()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get active jails count")
		info.ActiveJails = 0
	} else {
		info.ActiveJails = len(jails)
	}

	return info, nil
}

// ParseLogFile 解析日志文件
func (s *Fail2BanService) ParseLogFile(filePath string, lines int) ([]string, error) {
	cmd := exec.Command("tail", "-n", strconv.Itoa(lines), filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read log file %s: %w", filePath, err)
	}

	logLines := strings.Split(string(output), "\n")
	var filteredLines []string

	// 过滤掉空行
	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	return filteredLines, nil
}

// SearchLogs 搜索日志
func (s *Fail2BanService) SearchLogs(filePath, pattern string, lines int) ([]string, error) {
	cmd := exec.Command("grep", "-i", pattern, filePath)
	if lines > 0 {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("grep -i '%s' '%s' | tail -n %d", pattern, filePath, lines))
	}
	
	output, err := cmd.Output()
	if err != nil {
		// grep 返回1表示没有找到匹配，这不是错误
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to search logs: %w", err)
	}

	logLines := strings.Split(string(output), "\n")
	var filteredLines []string

	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	return filteredLines, nil
}