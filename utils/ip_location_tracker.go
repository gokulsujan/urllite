package utils

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

func GetIPAddressLocation(ipStr string) (map[string]string, error) {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ip := net.ParseIP(ipStr)
	record, err := db.City(ip)
	if err != nil {
		return nil, err
	}
	return map[string]string{"country": record.Country.Names["en"], "city": record.City.Names["en"]}, nil
}
