package linux

import (
	"net"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"go.ligato.io/cn-infra/v2/logging"
)

type Plugin struct {
	Deps

	PluginName   string
	LinuxHandler *netlink.Handle
}

type Deps struct {
	Log logging.PluginLogger
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	if new_handler, err := netlink.NewHandle(); err != nil {
		panic(err)
	} else {
		p.LinuxHandler = new_handler
	}
	p.PluginName = "linux"

	return nil
}

func (p *Plugin) String() string {
	return "linux"
}

func (p *Plugin) Close() error {
	return nil
}

type Route_Info struct {
	Dst IpNet
	Gw  IpNet
	// Dev and DevId are always synchronously be set or empty
	Dev string
	// Dev and DevId are always synchronously be set or empty
	DevId uint32
}

type IpNet struct {
	Ip   string
	Mask uint
}

func (p *Plugin) AddRoute(route Route_Info) error {
	req := &netlink.Route{
		Dst:       route.Dst.IpNetToStd(),
		MultiPath: []*netlink.NexthopInfo{{}},
	}

	if route.Gw.Ip != "" {
		req.MultiPath[0].Gw = route.Dst.IpNetToStd().IP
	}

	if route.Dev != "" {
		if intf, err := netlink.LinkByName(route.Dev); err != nil {
			p.Log.Errorln(err)
			panic(err)
		} else {
			req.MultiPath[0].LinkIndex = intf.Attrs().Index
		}
	}

	err := p.LinuxHandler.RouteAdd(req)
	if err != nil {
		p.Log.Errorln(err)
		panic(err)
	} else {
		p.Log.Infoln("successfully add a route to linux namespace!")
	}
	return nil
}

func (mynet IpNet) IpNetToStd() *net.IPNet {
	split_ips := strings.Split(mynet.Ip, ".")
	int_ips := []int{}
	for _, ip := range split_ips {
		if int_ip, err := strconv.Atoi(ip); err != nil {
			panic(err)
		} else {
			int_ips = append(int_ips, int_ip)
		}
	}
	byte_ips := []byte{
		byte(int_ips[0]),
		byte(int_ips[1]),
		byte(int_ips[2]),
		byte(int_ips[3]),
	}

	mask := []byte{}
	mask_len := mynet.Mask / 8
	for i := 0; i < 4; i++ {
		if i < int(mask_len) {
			mask = append(mask, byte(255))
		} else {
			mask = append(mask, byte(0))
		}
	}

	return &net.IPNet{
		IP: net.IPv4(
			byte_ips[0],
			byte_ips[1],
			byte_ips[2],
			byte_ips[3],
		),
		Mask: net.IPv4Mask(
			mask[0],
			mask[1],
			mask[2],
			mask[3],
		),
	}
}
