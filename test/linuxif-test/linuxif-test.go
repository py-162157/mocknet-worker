package main

import (
	"fmt"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"

	//"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	linux_intf "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	vpp_intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func main() {
	//VPP := app.DefaultVPP()
	//Linux := app.DefaultLinux()
	err := localclient.DataResyncRequest("linuxif_test").
		VppInterface(initialTap1()).
		LinuxInterface(initialLinuxTap1()).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("successfully create tap pair!")
	}
}

func initialTap1() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:    "tap1",
		Type:    vpp_intf.Interface_TAP,
		Enabled: true,
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func initialLinuxTap1() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap1",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"10.0.0.2/24",
		},
		HostIfName: "tap_to_vpp1",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap1",
			},
		},
	}
}
