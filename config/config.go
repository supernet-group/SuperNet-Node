package config

import (
	"SuperNet-Node/pattern"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Base struct {
		Rpc           string `yaml:"rpc"`
		PrivateKey    string `yaml:"privateKey"`
		SecurityLevel string `yaml:"securityLevel"`
	} `yaml:"base"`
	Console struct {
		WorkDirectory string `yaml:"workDirectory"`
		IpfsNodeUrl   string `yaml:"ipfsNodeUrl"`
		PublicIP      string `yaml:"publicIP"`
		DistriPort    string `yaml:"distriPort"`
		WorkPort      string `yaml:"workPort"`
		ServerPort    string `yaml:"serverPort"`
		ExpandPort1   string `yaml:"publicPortExpand1"`
		ExpandPort2   string `yaml:"publicPortExpand2"`
		ExpandPort3   string `yaml:"publicPortExpand3"`
	} `yaml:"console"`
}

var GlobalConfig Config

func InitializeConfig() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		logs.Error(fmt.Sprintf("Error reading config file: %v", err))
	}
	err = yaml.Unmarshal([]byte(data), &GlobalConfig)
	if err != nil {
		logs.Error(fmt.Sprintf("Error reading config file: %v", err))
	}

	if GlobalConfig.Console.WorkDirectory != "" {
		GlobalConfig.Console.WorkDirectory = utils.RemoveTrailingSlash(GlobalConfig.Console.WorkDirectory)
		GlobalConfig.Console.WorkDirectory = utils.EnsureLeadingSlash(GlobalConfig.Console.WorkDirectory)
	}

	if GlobalConfig.Console.IpfsNodeUrl == "" {
		GlobalConfig.Console.IpfsNodeUrl = pattern.DefaultIpfsNode
	} else {
		GlobalConfig.Console.IpfsNodeUrl = utils.RemoveTrailingSlash(GlobalConfig.Console.IpfsNodeUrl)
	}
	if GlobalConfig.Console.ServerPort == "" {
		GlobalConfig.Console.ServerPort = "13012"
	}
	if GlobalConfig.Console.WorkPort == "" {
		GlobalConfig.Console.WorkPort = "13011"
	}
	if GlobalConfig.Console.DistriPort == "" {
		GlobalConfig.Console.DistriPort = "13010"
	}
	if GlobalConfig.Base.Rpc == "" {
		GlobalConfig.Base.Rpc = pattern.RPC
	}
}

type SolanaConfig struct {
	Key string
	RPC string
}

func NewConfig(key string, rpc string) *SolanaConfig {
	return &SolanaConfig{
		Key: key,
		RPC: rpc,
	}
}
