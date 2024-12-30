package nginx

import (
	"SuperNet-Node/pattern"
	"SuperNet-Node/utils/nginx_utils"
	"fmt"
	"os/exec"
)

// StartNginx configures and restarts Nginx with specified ports.
func StartNginx(nginxPort, workPort, serverPort string) error {

	// Generate Nginx configuration file
	err := nginx_utils.GenNginxConfig(nginxPort, workPort, serverPort, pattern.ModleCreatePath)
	if err != nil {
		return fmt.Errorf("> gen nginx config: %v", err)
	}

	// Restart Nginx service
	cmd := exec.Command("sudo", "service", "nginx", "restart")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("> nginx restart: %v", err)
	}
	return nil
}

// StopNginx attempts to stop the Nginx service using the 'service' command.
func StopNginx() error {
	cmd := exec.Command("sudo", "service", "nginx", "stop")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("> nginx stop: %v", err)
	}
	return nil
}
