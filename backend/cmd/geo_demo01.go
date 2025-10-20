package main

import (
	"fmt"
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)


// /https://github.com/P3TERX/GeoLite.mmdb

func main() {

	ipStr := "43.130.153.46"
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Fatalf("invalid ip: %s", ipStr)
	}

	// 打开 City 数据库文件（请把 GeoLite2-City.mmdb 放在程序运行目录或修改路径）
	cityDB, err := geoip2.Open("config/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalf("failed to open city mmdb: %v", err)
	}
	defer cityDB.Close()

	cityRecord, err := cityDB.City(ip)
	if err != nil {
		log.Fatalf("city lookup failed: %v", err)
	}

	fmt.Printf("IP: %s\n", ipStr)
	if cityRecord.Country.IsoCode != "" {
		fmt.Printf("Country: %s (%s)\n", cityRecord.Country.Names["en"], cityRecord.Country.IsoCode)
	}
	if len(cityRecord.Subdivisions) > 0 {
		fmt.Printf("Subdivision: %s\n", cityRecord.Subdivisions[0].Names["en"])
	}
	if cityRecord.City.Names != nil {
		fmt.Printf("City: %s\n", cityRecord.City.Names["en"])
	}
	fmt.Printf("Location: lat=%.6f lon=%.6f\n", cityRecord.Location.Latitude, cityRecord.Location.Longitude)

	// 打开 ASN 数据库（可选，放在单独文件 GeoLite2-ASN.mmdb）
	asnDB, err := geoip2.Open("config/GeoLite2-ASN.mmdb")
	if err == nil {
		defer asnDB.Close()
		if asn, err := asnDB.ASN(ip); err == nil {
			fmt.Printf("ASN: %d (%s)\n", asn.AutonomousSystemNumber, asn.AutonomousSystemOrganization)
		}
	} else {
		fmt.Printf("ASN db not opened: %v\n", err)
	}
}