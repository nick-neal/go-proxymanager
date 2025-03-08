package nginx

import (
	"os/exec"
)

// check to make sure nginx config is in good state
func CheckNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

// restart nginx
func RestartNginx() error {
	err := CheckNginxConfig()
	if err == nil {
		cmd := exec.Command("systemctl", "restart", "nginx")
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err2 := cmd.Run()

		return err2
	}

	return err
}
