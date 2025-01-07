package nginx_utils

import (
	logs "SuperNet-Node/utils/log_utils"
	"fmt"
	"os"
)

func GenNginxConfig(nginxPort, workPort, serverPort, modleCreatePath string) error {
	logs.Normal(fmt.Sprintf("Start nginx. nginxPort: %v, workPort: %v, serverPort: %v",
		nginxPort, workPort, serverPort))
	nginxDir := "/etc/nginx/sites-enabled"

	os.Remove(nginxDir + "/super.conf")

	nginxConfig := fmt.Sprintf(`server {
	listen %v;
	listen [::]:%v;

	server_name super-net-node;

	location ^~ /super/ {
		proxy_pass http://127.0.0.1:%v/;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
	}

	location /uploadfiles {
		alias %v;
		index index.html;
	}

	location / {
		proxy_pass http://127.0.0.1:%v;
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`, nginxPort, nginxPort, serverPort, modleCreatePath, workPort)

	err := os.WriteFile(nginxDir+"/super.conf", []byte(nginxConfig), 0644)
	if err != nil {
		return fmt.Errorf("> WriteFile: %v", err)
	}

	return nil
}
