package location

import (
	"SuperNet-Node/utils/log_utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type InfoLocation struct {
	Country string `json:"Country"`
	Region  string `json:"RegionName"`
	City    string `json:"City"`
	IP      string `json:"query"`
}

func GetLocationInfo(ip string) (InfoLocation, error) {
	logs.Normal("Getting location info...")

	url := fmt.Sprintf("http://ip-api.com/json/%v", ip)
	resp, err := http.Get(url)
	if err != nil {
		return InfoLocation{}, fmt.Errorf("> http.Get: %v", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return InfoLocation{}, fmt.Errorf("> io.ReadAll: %v", err)
	}

	// Easy debugging
	var response InfoLocation
	json.Unmarshal(body, &response)
	// response.Country = "South Korea"
	// response.Region = "South Korea"
	// response.City = "Anyang-si"
	// response.IP = "67.185.75.5"
	return response, nil
}
