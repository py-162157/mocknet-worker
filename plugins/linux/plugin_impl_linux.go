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
	Log             logging.PluginLogger
	HostMainDevName string
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

	p.get_host_main_net_dev()

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

func (p *Plugin) Add_Route(route Route_Info) error {
	//p.Log.Infoln("dst:", route.Dst, "gw:", route.Gw, "dev:", route.Dev)
	req := &netlink.Route{
		Dst:       route.Dst.IpNetToStd(),
		MultiPath: []*netlink.NexthopInfo{{}},
	}

	if route.Gw.Ip != "" {
		req.MultiPath[0].Gw = route.Gw.IpNetToStd().IP
		req.Gw = route.Gw.IpNetToStd().IP
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
		if err.Error() == "file exists" {
			p.Log.Warningln("route already exist, so automaticly delete it and retry again")
			p.Del_Route(route)
			p.Add_Route(route)
		} else {
			panic(err)
		}
	} else {
		p.Log.Infoln("successfully add a route to linux namespace!")
	}

	return nil
}

func (p *Plugin) Del_Route(route Route_Info) error {
	err := p.LinuxHandler.RouteDel(&netlink.Route{
		Dst: route.Dst.IpNetToStd(),
	})
	if err != nil {
		p.Log.Errorln("delete route failed")
		panic(err)
	}
	p.Log.Infoln("delete route for destination", route.Dst.Ip)

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

func (p *Plugin) get_host_main_net_dev() {
	devs, err := p.LinuxHandler.LinkList()
	if err != nil {
		panic(err)
	}
	for _, dev := range devs {
		if string([]byte(dev.Attrs().Name)[:1]) == "e" {
			p.HostMainDevName = dev.Attrs().Name
			break
		}
	}
}
