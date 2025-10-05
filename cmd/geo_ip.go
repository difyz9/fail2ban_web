package main

import (
	"fmt"
	"log"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

func main() {
	db, err := geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, use netip.ParseAddr and check for errors
	ip, err := netip.ParseAddr("81.2.69.142")
	if err != nil {
		log.Fatal(err)
	}
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	if !record.HasData() {
		fmt.Println("No data found for this IP")
		return
	}
	fmt.Printf("Portuguese (BR) city name: %v\n", record.City.Names.BrazilianPortuguese)
	if len(record.Subdivisions) > 0 {
		fmt.Printf("English subdivision name: %v\n", record.Subdivisions[0].Names.English)
	}
	fmt.Printf("Russian country name: %v\n", record.Country.Names.Russian)
	fmt.Printf("ISO country code: %v\n", record.Country.ISOCode)
	fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
	if record.Location.HasCoordinates() {
		fmt.Printf("Coordinates: %v, %v\n", *record.Location.Latitude, *record.Location.Longitude)
	}
	// Output:
	// Portuguese (BR) city name: Londres
	// English subdivision name: England
	// Russian country name: Великобритания
	// ISO country code: GB
	// Time zone: Europe/London
	// Coordinates: 51.5142, -0.0931
}
