package main

import (
	"os/exec"
)

var POD_VPP_CONFIG_FILE = "/home/ubuntu/vpp.conf"

func main() {
	restart_cmd := exec.Command("service", "vpp", "restart")
	restart_cmd.Run()
}
