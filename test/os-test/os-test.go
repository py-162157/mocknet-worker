package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var POD_VPP_CONFIG_FILE = "/home/ubuntu/vpp.conf"
var container_id = "c131c236ce50"

func main() {
	//cmd := exec.Command("ifconfig", "eno1", "|", "awk", `'/ether/'`, "|", "awk", `'{print $2}'`)
	//cmd := exec.Command("ifconfig", "eno1")

	cmd := exec.Command("docker", "exec", container_id, "ifconfig", "tap0")

	r, err := cmd.Output()
	if err != nil {
		fmt.Println("failed:", err)
	}

	split_string1 := strings.Split(string(r), "ether")
	split_string2 := strings.Split(split_string1[1], "txqueuelen")
	split_string3 := strings.ReplaceAll(split_string2[0], " ", "")
	fmt.Println(split_string3)
}
