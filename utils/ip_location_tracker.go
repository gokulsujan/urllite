package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type IPInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
}

func GetIPAddressLocation(ipStr string) (map[string]string, error) {

	url := "https://ipinfo.io/" + ipStr + "/json"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var data IPInfo
	json.Unmarshal(body, &data)

	return map[string]string{"country": data.Country, "city": data.City}, nil
}
