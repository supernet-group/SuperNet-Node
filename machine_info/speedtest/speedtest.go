package speedtest

import (
	"SuperNet-Node/utils/log_utils"
	"fmt"
	"os/exec"
	"strings"
)

type InfoSpeed struct {
	Download string `json:"Download"`
	Upload   string `json:"Upload"`
}

// GetSpeedInfo retrieves the current network speed information using the speedtest-cli command.
func GetSpeedInfo() (InfoSpeed, error) {
	logs.Normal("Getting network speed info...")

	out, err := exec.Command("speedtest-cli").Output()

	if err != nil {
		return InfoSpeed{
			Download: "- Mbit/s",
			Upload:   "- Mbit/s",
		}, fmt.Errorf("speedtest-cli, Failed to execute command: %v\n %v", err, string(out))
	}

	output := string(out)
	lines := strings.Split(output, "\n")

	var netSpeed InfoSpeed

	for _, line := range lines {
		if strings.Contains(line, "Download") {
			netSpeed.Download = strings.TrimSpace(strings.TrimPrefix(line, "Download:"))
		}

		if strings.Contains(line, "Upload") {
			netSpeed.Upload = strings.TrimSpace(strings.TrimPrefix(line, "Upload:"))
		}
	}
	return netSpeed, nil
}
