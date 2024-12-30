package ip

import (
	"SuperNet-Node/config"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type InfoIP struct {
	IP         string   `json:"ip"`
	Port       string   `json:"port"`
	ExpandPort []string `json:"expandPort"`
}

func GetIpInfo() (InfoIP, error) {
	logs.Normal("Getting public ip info...")

	var response InfoIP
	response.IP = config.GlobalConfig.Console.PublicIP
	if response.IP == "" {
		resp, err := http.Get("https://ipinfo.io")
		if err != nil {
			return InfoIP{}, fmt.Errorf("> http.Get: %v", err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return InfoIP{}, fmt.Errorf("> io.ReadAll: %v", err)
		}
		json.Unmarshal(body, &response)

		config.GlobalConfig.Console.PublicIP = response.IP
	}

	if !utils.CheckPort(config.GlobalConfig.Console.DistriPort) {
		return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.DistriPort)
	}
	if !utils.CheckPort(config.GlobalConfig.Console.WorkPort) {
		return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.WorkPort)
	}
	if !utils.CheckPort(config.GlobalConfig.Console.ServerPort) {
		return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.ServerPort)
	}

	if config.GlobalConfig.Console.ExpandPort1 != "" {
		if !utils.CheckPort(config.GlobalConfig.Console.ExpandPort1) {
			return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.ExpandPort1)
		}
		response.ExpandPort = append(response.ExpandPort, config.GlobalConfig.Console.ExpandPort1)
	}
	if config.GlobalConfig.Console.ExpandPort2 != "" {
		if !utils.CheckPort(config.GlobalConfig.Console.ExpandPort2) {
			return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.ExpandPort2)
		}
		response.ExpandPort = append(response.ExpandPort, config.GlobalConfig.Console.ExpandPort2)
	}
	if config.GlobalConfig.Console.ExpandPort3 != "" {
		if !utils.CheckPort(config.GlobalConfig.Console.ExpandPort3) {
			return InfoIP{}, fmt.Errorf("> port %s is not available", config.GlobalConfig.Console.ExpandPort3)
		}
		response.ExpandPort = append(response.ExpandPort, config.GlobalConfig.Console.ExpandPort3)
	}

	response.Port = config.GlobalConfig.Console.DistriPort
	return response, nil
}
