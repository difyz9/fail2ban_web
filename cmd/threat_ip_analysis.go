package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IPInfo 结构体用于解析IP地理信息
type IPInfo struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
}

// 威胁IP列表
var threatIPs = []string{
	"129.213.151.201", "185.242.226.119", "152.136.209.127", "39.103.62.164", "42.193.59.214",
	"121.229.3.105", "183.204.86.11", "195.178.110.160", "43.226.79.54", "109.244.96.90",
	"112.78.3.205", "115.190.9.220", "88.210.63.67", "180.76.243.137", "203.104.42.192",
	"211.156.88.24", "39.104.52.208", "118.193.35.29", "163.227.230.172", "94.103.34.243",
	"106.55.52.106", "115.46.97.181", "39.129.46.33", "51.81.22.34", "122.51.131.143",
	"221.2.109.10", "101.126.128.106", "110.53.82.195", "34.38.145.6", "154.201.90.141",
	"20.163.16.165", "27.79.40.203", "47.83.7.5", "115.190.16.93", "211.156.80.39",
	"39.129.46.49", "106.52.170.23", "152.32.190.168", "210.1.60.243", "66.249.79.65",
}

func getIPInfo(ip string) (*IPInfo, error) {
	// 尝试多个免费API
	apis := []string{
		"http://ip-api.com/json/%s",
		"http://ipapi.co/%s/json/",
		"https://ipinfo.io/%s/json",
	}
	
	for i, apiURL := range apis {
		url := fmt.Sprintf(apiURL, ip)
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("API %d 失败: %v ", i+1, err)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			fmt.Printf("API %d HTTP错误: %d ", i+1, resp.StatusCode)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("API %d 读取失败: %v ", i+1, err)
			continue
		}
		
		// 检查是否是有效的JSON
		if len(body) == 0 || body[0] != '{' {
			fmt.Printf("API %d 返回非JSON: %s ", i+1, string(body[:min(50, len(body))]))
			continue
		}
		
		var info IPInfo
		err = json.Unmarshal(body, &info)
		if err != nil {
			fmt.Printf("API %d JSON解析失败: %v ", i+1, err)
			continue
		}
		
		// 不同API的字段映射
		if i == 0 { // ip-api.com
			info = mapIPAPI(body)
		} else if i == 2 { // ipinfo.io
			info = mapIPInfo(body, ip)
		}
		
		return &info, nil
	}
	
	return nil, fmt.Errorf("所有API都失败了")
}

func mapIPAPI(data []byte) IPInfo {
	var apiResponse struct {
		Query       string  `json:"query"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
	}
	
	json.Unmarshal(data, &apiResponse)
	
	return IPInfo{
		IP:          apiResponse.Query,
		Country:     apiResponse.Country,
		CountryCode: apiResponse.CountryCode,
		Region:      apiResponse.Region,
		City:        apiResponse.City,
		ISP:         apiResponse.ISP,
		Org:         apiResponse.Org,
		AS:          apiResponse.AS,
		Lat:         apiResponse.Lat,
		Lon:         apiResponse.Lon,
		Timezone:    apiResponse.Timezone,
	}
}

func mapIPInfo(data []byte, ip string) IPInfo {
	var apiResponse struct {
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
		Org      string `json:"org"`
		Timezone string `json:"timezone"`
		Loc      string `json:"loc"`
	}
	
	json.Unmarshal(data, &apiResponse)
	
	var lat, lon float64
	if apiResponse.Loc != "" {
		fmt.Sscanf(apiResponse.Loc, "%f,%f", &lat, &lon)
	}
	
	return IPInfo{
		IP:       ip,
		Country:  apiResponse.Country,
		Region:   apiResponse.Region,
		City:     apiResponse.City,
		Org:      apiResponse.Org,
		ISP:      apiResponse.Org,
		Lat:      lat,
		Lon:      lon,
		Timezone: apiResponse.Timezone,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("=== 威胁IP地理信息分析 ===")
	fmt.Printf("正在分析 %d 个威胁IP的地理信息...\n\n", len(threatIPs))
	
	// 统计信息
	countryStats := make(map[string]int)
	regionStats := make(map[string]int)
	ispStats := make(map[string]int)
	
	successCount := 0
	
	for i, ip := range threatIPs {
		fmt.Printf("正在查询 [%d/%d] %s ... ", i+1, len(threatIPs), ip)
		
		info, err := getIPInfo(ip)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			time.Sleep(2 * time.Second) // 失败后等待更长时间
			continue
		}
		
		successCount++
		fmt.Printf("✅\n")
		
		// 显示详细信息
		fmt.Printf("  🌍 国家: %s (%s)\n", info.Country, info.CountryCode)
		if info.Region != "" {
			fmt.Printf("  🏙️  地区: %s\n", info.Region)
		}
		if info.City != "" {
			fmt.Printf("  🏘️  城市: %s\n", info.City)
		}
		if info.ISP != "" {
			fmt.Printf("  🌐 ISP: %s\n", info.ISP)
		}
		if info.Org != "" && info.Org != info.ISP {
			fmt.Printf("  🏢 组织: %s\n", info.Org)
		}
		if info.AS != "" {
			fmt.Printf("  🔢 AS: %s\n", info.AS)
		}
		if info.Lat != 0 && info.Lon != 0 {
			fmt.Printf("  📍 坐标: %.4f, %.4f\n", info.Lat, info.Lon)
		}
		fmt.Println()
		
		// 统计
		if info.Country != "" {
			countryStats[info.Country]++
		}
		if info.Region != "" {
			key := fmt.Sprintf("%s, %s", info.Region, info.Country)
			regionStats[key]++
		}
		if info.ISP != "" {
			ispStats[info.ISP]++
		}
		
		// 避免请求过快被限制 - 增加延迟
		time.Sleep(3 * time.Second)
	}
	
	// 显示统计结果
	fmt.Println("\n=== 📊 威胁统计分析 ===")
	fmt.Printf("成功查询: %d/%d (%.1f%%)\n\n", successCount, len(threatIPs), float64(successCount)/float64(len(threatIPs))*100)
	
	fmt.Println("🏳️ 按国家统计:")
	for country, count := range countryStats {
		percentage := float64(count) / float64(successCount) * 100
		fmt.Printf("  %s: %d个IP (%.1f%%)\n", country, count, percentage)
	}
	
	fmt.Println("\n🗺️ 按地区统计 (>1个IP):")
	for region, count := range regionStats {
		if count > 1 {
			fmt.Printf("  %s: %d个IP\n", region, count)
		}
	}
	
	fmt.Println("\n🌐 主要ISP统计:")
	for isp, count := range ispStats {
		if count > 1 {
			fmt.Printf("  %s: %d个IP\n", isp, count)
		}
	}
	
	// 安全建议
	fmt.Println("\n=== 🛡️ 安全建议 ===")
	fmt.Println("基于地理分析结果:")
	
	// 识别高风险国家
	fmt.Println("\n🚨 高风险地区:")
	for country, count := range countryStats {
		if count >= 3 {
			fmt.Printf("  • %s: %d个威胁IP - 建议加强监控\n", country, count)
		}
	}
	
	fmt.Println("\n💡 建议措施:")
	fmt.Println("  1. 考虑对高风险国家实施额外的访问限制")
	fmt.Println("  2. 监控来自云服务商的异常流量")
	fmt.Println("  3. 实施基于地理位置的防火墙规则")
	fmt.Println("  4. 加强对这些IP段的实时监控")
	fmt.Println("  5. 可以将高频攻击的国家/地区加入地理封锁名单")
	
	fmt.Printf("\n✅ 分析完成！共分析了 %d 个威胁IP，成功查询 %d 个\n", len(threatIPs), successCount)
}