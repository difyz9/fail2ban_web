package service

import (
	"fmt"
	"net"
)

// WhitelistService IP白名单服务
type WhitelistService struct {
	whitelistIPs   []string
	whitelistCIDRs []*net.IPNet
}

// NewWhitelistService 创建白名单服务
func NewWhitelistService() *WhitelistService {
	service := &WhitelistService{
		whitelistIPs: []string{
			"127.0.0.1",
			"::1",
		},
	}
	
	// 添加常见的内网CIDR
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12", 
		"192.168.0.0/16",
		"169.254.0.0/16", // 链路本地地址
		"::1/128",        // IPv6 localhost
		"fc00::/7",       // IPv6 ULA
		"fe80::/10",      // IPv6 链路本地
	}
	
	for _, cidr := range privateCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			service.whitelistCIDRs = append(service.whitelistCIDRs, network)
		}
	}
	
	return service
}

// IsWhitelisted 检查IP是否在白名单中
func (w *WhitelistService) IsWhitelisted(ipStr string) bool {
	// 检查精确匹配
	for _, whiteIP := range w.whitelistIPs {
		if ipStr == whiteIP {
			return true
		}
	}
	
	// 检查CIDR匹配
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	for _, network := range w.whitelistCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	
	return false
}

// AddIP 添加IP到白名单
func (w *WhitelistService) AddIP(ipStr string) error {
	if net.ParseIP(ipStr) == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}
	
	// 检查是否已存在
	for _, existingIP := range w.whitelistIPs {
		if existingIP == ipStr {
			return nil // 已存在
		}
	}
	
	w.whitelistIPs = append(w.whitelistIPs, ipStr)
	return nil
}

// AddCIDR 添加CIDR到白名单
func (w *WhitelistService) AddCIDR(cidrStr string) error {
	_, network, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", cidrStr)
	}
	
	// 检查是否已存在
	for _, existingNetwork := range w.whitelistCIDRs {
		if existingNetwork.String() == network.String() {
			return nil // 已存在
		}
	}
	
	w.whitelistCIDRs = append(w.whitelistCIDRs, network)
	return nil
}

// RemoveIP 从白名单移除IP
func (w *WhitelistService) RemoveIP(ipStr string) {
	for i, ip := range w.whitelistIPs {
		if ip == ipStr {
			w.whitelistIPs = append(w.whitelistIPs[:i], w.whitelistIPs[i+1:]...)
			break
		}
	}
}

// GetWhitelist 获取白名单列表
func (w *WhitelistService) GetWhitelist() map[string]interface{} {
	cidrs := make([]string, len(w.whitelistCIDRs))
	for i, network := range w.whitelistCIDRs {
		cidrs[i] = network.String()
	}
	
	return map[string]interface{}{
		"ips":   w.whitelistIPs,
		"cidrs": cidrs,
	}
}