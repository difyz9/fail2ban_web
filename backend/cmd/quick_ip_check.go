package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 简化的IP信息结构
type SimpleIPInfo struct {
	IP      string `json:"query"`
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
	AS      string `json:"as"`
	Status  string `json:"status"`
}

// 前10个威胁IP进行测试
var testIPs = []string{
	"129.213.151.201", "185.242.226.119", "152.136.209.127", "39.103.62.164", "42.193.59.214",
	"121.229.3.105", "183.204.86.11", "195.178.110.160", "43.226.79.54", "109.244.96.90",
}

func getSimpleIPInfo(ip string) (*SimpleIPInfo, error) {
	// 使用ip-api.com，每分钟限制45次请求
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var info SimpleIPInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 响应: %s", err, string(body[:min(100, len(body))]))
	}
	
	if info.Status == "fail" {
		return nil, fmt.Errorf("API返回失败状态")
	}
	
	return &info, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("=== 威胁IP快速地理分析 ===")
	fmt.Printf("分析前 %d 个威胁IP...\n\n", len(testIPs))
	
	countryCount := make(map[string]int)
	
	for i, ip := range testIPs {
		fmt.Printf("[%d/%d] 查询 %s ... ", i+1, len(testIPs), ip)
		
		info, err := getSimpleIPInfo(ip)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		
		fmt.Printf("✅\n")
		fmt.Printf("  🌍 %s (%s)\n", info.Country, info.Region)
		if info.City != "" {
			fmt.Printf("  🏘️  %s\n", info.City)
		}
		if info.ISP != "" {
			fmt.Printf("  🌐 %s\n", info.ISP)
		}
		if info.AS != "" {
			fmt.Printf("  🔢 %s\n", info.AS)
		}
		fmt.Println()
		
		countryCount[info.Country]++
		
		// 等待2秒避免限制
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("=== 国家统计 ===")
	for country, count := range countryCount {
		fmt.Printf("%s: %d个IP\n", country, count)
	}
	
	fmt.Println("\n🎯 这些是我们智能分析系统识别出的威胁IP！")
	fmt.Println("可以看到攻击来源分布在多个国家，说明这是全球性的威胁。")
}