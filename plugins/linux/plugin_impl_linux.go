package linux

import (
	"bytes"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"mocknet/plugins/vpp"

	"github.com/vishvananda/netlink"
	"go.ligato.io/cn-infra/v2/logging"
)

type ProcessResult string

const (
	AlreadyExist ProcessResult = "AlreadyExist"
	TimesOver    ProcessResult = "TimesOver"
	Success      ProcessResult = "Success"
	NotExist     ProcessResult = "NotExist"
	Failed       ProcessResult = "Failed"
)

const (
	MAX_RETRY_TIMES     = 3
	RETRY_TIME_INTERVAL = 1 * time.Second
)

type Plugin struct {
	Deps

	PluginName   string
	LinuxHandler *netlink.Handle
}

type Deps struct {
	Log             logging.PluginLogger
	HostMainDevName string
	Vpp             *vpp.Plugin
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

func (p *Plugin) Add_Route(route Route_Info) ProcessResult {
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
			return AlreadyExist
		} else {
			return TimesOver
		}
	} else {
		p.Log.Infoln("added a route to linux namespace!")
		return Success
	}
}

func (p *Plugin) Del_Route(route Route_Info) ProcessResult {
	err := p.LinuxHandler.RouteDel(&netlink.Route{
		Dst: route.Dst.IpNetToStd(),
	})
	if err != nil {
		p.Log.Errorln("delete route failed")
		return TimesOver
	}
	p.Log.Infoln("delete route for destination", route.Dst.Ip)
	return Success
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

func (p *Plugin) Pod_Add_Route(container_id string, pod_name string, route ...Route_Info) ProcessResult {
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "exec", container_id, "ip", "route", "add", "10.1.0.0/16", "dev", "tap0")
	cmd.Stderr = &stderr

	count := 0
	for {
		err := cmd.Run()
		if err == nil {
			p.Log.Infoln("added route for pod", pod_name)
			return Success
		} else {
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to add route for pod", pod_name)
				return TimesOver
			} else {
				p.Log.Warningln("failed to add route for pod", pod_name, "err:", stderr.String(), "retry")
				if strings.Contains(stderr.String(), "find") { // can't find device tap0
					if result, tap_id := p.Vpp.Pod_Create_Tap(pod_name); result == vpp.Success || result == vpp.AlreadyExist {
						p.Vpp.Pod_Set_interface_state_up(pod_name, tap_id)
					}
				}
			}
		}
		count += 1
		time.Sleep(RETRY_TIME_INTERVAL)
	}
}

func (p *Plugin) Pod_Set_Ip(container_id string, pod_name string, ip string, masklen string) ProcessResult {
	var stderr bytes.Buffer
	cpip := strings.Split(ip, ".") // conrol plane ip
	data_plane_ip := "10.1." + cpip[2] + "." + cpip[3] + masklen
	cmd := exec.Command("docker", "exec", container_id, "ip", "addr", "add", "dev", "tap0", data_plane_ip)
	cmd.Stderr = &stderr

	count := 0
	for {
		err := cmd.Run()
		if err == nil {
			p.Log.Infoln("set ip address for pod", pod_name)
			return Success
		} else {
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set ip address for pod", pod_name)
				return TimesOver
			}
			p.Log.Warningln("failed to set ip address for pod", pod_name, "err:", stderr.String(), "retry")
		}
		count += 1
		time.Sleep(RETRY_TIME_INTERVAL)
	}
}

func (p *Plugin) Set_Receiver(container_id string, pod_name string) ProcessResult {
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "exec", container_id, "iperf3", "-s")
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err == nil {
		p.Log.Infoln("set iperf3 server for pod", pod_name)
		return Success
	} else {
		p.Log.Errorln("failed to set iperf3 server for pod", pod_name, "err:", stderr.String())
		return Failed
	}
}

func (p *Plugin) Set_Sender(container_id string, pod_name string, dst_ip string) (ProcessResult, string) {
	p.Log.Infoln("for pod", pod_name, "sender command is ", "docker", "exec", container_id, "iperf3", "-c", dst_ip)
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "exec", container_id, "iperf3", "-c", dst_ip)
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err == nil {
		//p.Log.Infoln("start iperf3 client for pod", pod_name)
		return Success, string(output)
	} else {
		p.Log.Errorln("failed to start iperf3 client for pod", pod_name, "err:", stderr.String())
		return Failed, "failed"
	}
}

func (p *Plugin) RestartVpp() ProcessResult {
	restart_cmd := exec.Command("service", "vpp", "restart")
	err := restart_cmd.Run()
	if err != nil {
		p.Log.Errorln(err)
		return Failed
	} else {
		p.Log.Infoln("service vpp restarted")
		return Success
	}
}

func (p *Plugin) Get_MAC_Addr(container_id string, pod_name string) string {
	cmd := exec.Command("docker", "exec", container_id, "ifconfig", "tap0")
	r, err := cmd.Output()
	if err != nil {
		p.Log.Errorln(err)
	}

	//p.Log.Infoln("for pod ", pod_name, " origin string =", string(r))
	split_string1 := strings.Split(string(r), "ether")[1]
	//p.Log.Infoln("string1 =", split_string1)
	split_string2 := strings.Split(split_string1, "txqueuelen")[0]
	//p.Log.Infoln("string2 =", split_string2)
	MAC_Address := strings.ReplaceAll(split_string2, " ", "")
	return MAC_Address
}

type ARP struct {
	Name string
	Ip   string
	Mac  string
}

func (p *Plugin) Set_Static_ARP(container_id string, pod_name string, podinfos map[string]ARP) error {
	for dst_name, info := range podinfos {
		if dst_name != pod_name {
			//p.Log.Infoln("for pod", info.Name, "containerid =", info.ContainerId, "dstip =", info.Ip, "macaddr =", info.Mac)
			cmd := exec.Command("docker", "exec", container_id, "arp", "-s", info.Ip, info.Mac)
			err := cmd.Run()
			if err != nil {
				p.Log.Errorln(err)
			}
		}
	}

	p.Log.Infoln("Static ARP write for pod", pod_name)

	return nil
}

// set static arp only for paired hosts
func (p *Plugin) Set_Static_ARP_Single(container_id string, pod_name string, podinfo ARP) error {
	//p.Log.Infoln("for pod", info.Name, "containerid =", info.ContainerId, "dstip =", info.Ip, "macaddr =", info.Mac)
	cmd := exec.Command("docker", "exec", container_id, "arp", "-s", podinfo.Ip, podinfo.Mac)
	err := cmd.Run()
	if err != nil {
		p.Log.Errorln(err)
	}

	p.Log.Infoln("Static ARP write for pod", pod_name)

	return nil
}
