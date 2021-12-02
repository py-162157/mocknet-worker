package main

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func main() {
	handler, err := netlink.NewHandle()
	if err != nil {
		panic(err)
	}
	ens2, _ := netlink.LinkByName("ens2")
	route := &netlink.Route{
		//LinkIndex: ens2.Attrs().Index,
		Dst: &net.IPNet{
			IP:   net.IPv4(10, 3, 2, 0),
			Mask: net.CIDRMask(24, 32),
		},
		MultiPath: []*netlink.NexthopInfo{
			{
				LinkIndex: ens2.Attrs().Index,
				Gw:        net.IPv4(192, 168, 122, 100),
			},
		},
		//Src: net.IPv4(192, 168, 122, 101),
	}
	err = handler.RouteAdd(route)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("successfully add a route to linux namespace!")
	}
}
