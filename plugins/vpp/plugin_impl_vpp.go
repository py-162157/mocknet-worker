package vpp

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"

	"flag"
	"os"
	"strconv"
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter/socketclient"

	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"

	fib_types_2009 "mocknet/binapi/vpp2009/fib_types"
	interfaces_2009 "mocknet/binapi/vpp2009/interface"
	interface_types_2009 "mocknet/binapi/vpp2009/interface_types"
	ip_2009 "mocknet/binapi/vpp2009/ip"
	ip_types_2009 "mocknet/binapi/vpp2009/ip_types"
	l2_2009 "mocknet/binapi/vpp2009/l2"
	memif_2009 "mocknet/binapi/vpp2009/memif"
	tap_2009 "mocknet/binapi/vpp2009/tapv2"

	fib_types_2110 "mocknet/binapi/vpp2110/fib_types"
	interfaces_2110 "mocknet/binapi/vpp2110/interface"
	interface_types_2110 "mocknet/binapi/vpp2110/interface_types"
	ip_2110 "mocknet/binapi/vpp2110/ip"
	ip_types_2110 "mocknet/binapi/vpp2110/ip_types"
	l2_2110 "mocknet/binapi/vpp2110/l2"
	memif_2110 "mocknet/binapi/vpp2110/memif"
	tap_2110 "mocknet/binapi/vpp2110/tapv2"
	vxlan_2110 "mocknet/binapi/vpp2110/vxlan"
)

const (
	SOCKET_NAME         = "memif.sock"
	RETRY_TIME_INTERVAL = 1 * time.Second
	MAX_RETRY_TIMES     = 3
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
	pod_sock = "/var/run/mocknet/"
)

const (
	// NextHopViaLabelUnset constant has to be assigned into the field next hop
	// via label in ip_add_del_route binary message if next hop via label is not defined.
	// Equals to MPLS_LABEL_INVALID defined in VPP
	NextHopViaLabelUnset uint32 = 0xfffff + 1

	// ClassifyTableIndexUnset is a default value for field classify_table_index in ip_add_del_route binary message.
	ClassifyTableIndexUnset = ^uint32(0)

	// NextHopOutgoingIfUnset constant has to be assigned into the field next_hop_outgoing_interface
	// in ip_add_del_route binary message if outgoing interface for next hop is not defined.
	NextHopOutgoingIfUnset = ^uint32(0)
)

type ProcessResult string

const (
	AlreadyExist ProcessResult = "AlreadyExist"
	TimesOver    ProcessResult = "TimesOver"
	Success      ProcessResult = "Success"
	NotExist     ProcessResult = "NotExist"
)

type Plugin struct {
	Deps

	PluginName   string
	K8sNamespace string
	EtcdClient   *clientv3.Client
	Channel      api.Channel
}

type Deps struct {
	Log logging.PluginLogger
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	p.PluginName = "vpp"

	conn, conev, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		p.Log.Errorln("ERROR:", err)
	}

	select {
	case e := <-conev:
		if e.State != core.Connected {
			p.Log.Errorln("ERROR: connecting to VPP failed or interrupted:", e.Error)
		}
	}

	ch, err := conn.NewAPIChannel()
	if err != nil {
		p.Log.Errorln("error when connect to vpp")
		panic(err)
	} else {
		p.Channel = ch
	}

	return nil
}

func (p *Plugin) String() string {
	return "vpp"
}

func (p *Plugin) Close() error {
	return nil
}

func (p *Plugin) CreateSocket(id uint32, filedir string) ProcessResult {
	filename := filedir + "/" + SOCKET_NAME
	_, err := os.Stat(filedir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(filename, 0777)
		}
	}
	//p.Log.Infoln("createing socket:", id, filename)

	req := &memif_2110.MemifSocketFilenameAddDel{
		IsAdd:          true,
		SocketID:       id,
		SocketFilename: filename,
	}
	reply := &memif_2110.MemifSocketFilenameAddDelReply{}

	count := 0
	for {
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to create memif socket, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Infoln("max retry times up to create memif socket")
				return TimesOver
			}
		} else {
			p.Log.Infoln("created socket:", id, filename)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) DeleteSocket(id uint32) {
	req := &memif_2110.MemifSocketFilenameAddDel{
		IsAdd:    false,
		SocketID: id,
	}
	reply := &memif_2110.MemifSocketFilenameAddDelReply{}

	if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to delete memif socket, id =", id)
		panic(err)
	}

	p.Log.Infoln("deleted socket:", id)
}

func (p *Plugin) Create_Memif_Interface(role_string string, id uint32, socket_id uint32) (ProcessResult, uint32) {
	var role memif_2110.MemifRole
	if role_string == "master" {
		role = memif_2110.MEMIF_ROLE_API_MASTER
	} else {
		role = memif_2110.MEMIF_ROLE_API_SLAVE
	}

	req := &memif_2110.MemifCreate{
		Role:     role,
		ID:       id,
		SocketID: socket_id,
	}
	reply := &memif_2110.MemifCreateReply{}

	count := 0
	for {
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				p.Log.Warningln("the memif interface already exist")
				return AlreadyExist, 0
			} else {
				p.Log.Warningln("failed to create memif interface, retrying", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to create memif interface")
					return TimesOver, 0
				}
			}
		} else {
			p.Log.Infoln("created interface id:", id, "socket-id:", socket_id, "role:", role_string)
			return Success, uint32(reply.SwIfIndex)
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Delete_Memif_Interface(id uint32) ProcessResult {
	req := &memif_2110.MemifDelete{
		SwIfIndex: interface_types_2110.InterfaceIndex(id),
	}
	reply := &memif_2110.MemifDeleteReply{}

	if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
		if strings.Contains(err.Error(), "Invalid") {
			p.Log.Warningln("the memif interface to be deleted not exist")
			return NotExist
		} else {
			p.Log.Errorln("failed to delete memif interface, id =", id)
			return TimesOver
		}
	} else {
		p.Log.Infoln("deleted memif interface, id =", id)
		return Success
	}
}

func (p *Plugin) Pod_Create_Memif_Interface(pod_name string, role_string string, name string, id uint32) (ProcessResult, uint32) {
	var role memif_2009.MemifRole
	if role_string == "master" {
		role = memif_2009.MEMIF_ROLE_API_MASTER
	} else {
		role = memif_2009.MEMIF_ROLE_API_SLAVE
	}

	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &memif_2009.MemifCreate{
		Role:     role,
		ID:       id,
		SocketID: 0,
	}
	reply := &memif_2009.MemifCreateReply{}

	count := 0
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to create memif interface for pod", pod_name, ", retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to create memif interface for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver, 0
			}
		} else {
			p.Log.Infoln("created memif interface for pod", pod_name, ",id:", id, ",role:", role_string)
			return Success, uint32(reply.SwIfIndex)
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Set_interface_state_up(id uint32) ProcessResult {
	req := &interfaces_2110.SwInterfaceSetFlags{
		SwIfIndex: interface_types_2110.InterfaceIndex(id),
		Flags:     1,
	}
	reply := &interfaces_2110.SwInterfaceSetFlagsReply{}

	count := 0
	for {
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to set interface state up, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface state up")
				return TimesOver
			}
		} else {
			p.Log.Infoln("set interface id", id, "state up")
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Set_interface_state_up(pod_name string, id uint32) ProcessResult {
	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &interfaces_2009.SwInterfaceSetFlags{
		SwIfIndex: interface_types_2009.InterfaceIndex(id),
		Flags:     1,
	}
	reply := &interfaces_2009.SwInterfaceSetFlagsReply{}

	count := 0
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to set interface state up for pod", pod_name, ", retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface state up for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver
			}
		} else {
			p.Log.Infoln("set interface id", id, "state up for pod", pod_name)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Create_Vxlan_Tunnel(src string, dst string, vni uint32, instance uint32) (ProcessResult, uint32) {
	//p.Log.Infoln("src =", src, "dst =", dst, "vni =", vni, "instance =", instance)
	src_addr := strings.Split(src, ".")
	src_addr_slice := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		conv, err := strconv.Atoi(src_addr[i])
		if err != nil {
			panic("error parsing src ip address string to int")
		}
		src_addr_slice[i] = uint8(conv)
	}

	dst_addr := strings.Split(dst, ".")
	dst_addr_slice := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		conv, err := strconv.Atoi(dst_addr[i])
		if err != nil {
			p.Log.Error("error parsing dst ip address string to int, error is", err)
			panic(err)
		}
		dst_addr_slice[i] = uint8(conv)
	}

	req := &vxlan_2110.VxlanAddDelTunnel{
		IsAdd:    true,
		Instance: instance,
		SrcAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				src_addr_slice[0],
				src_addr_slice[1],
				src_addr_slice[2],
				src_addr_slice[3],
			}),
		},
		DstAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				dst_addr_slice[0],
				dst_addr_slice[1],
				dst_addr_slice[2],
				dst_addr_slice[3],
			}),
		},
		Vni:            vni,
		EncapVrfID:     0,
		DecapNextIndex: 1,
	}
	reply := &vxlan_2110.VxlanAddDelTunnelReply{}

	count := 0
	for {
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			if strings.Contains(err.Error(), "already exist") {
				p.Log.Info("vxlan tunnel already exist:", "src:", src, "dst:", dst, "vni:", vni)
				return AlreadyExist, 0
			} else {
				p.Log.Warningln("failed to create vxlan tunnel, retrying,", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to create vxlan tunnel")
					return TimesOver, 0
				}
			}
		} else {
			p.Log.Infoln("created vxlan tunnel:", "src:", src, "dst:", dst, "vni:", vni)
			return Success, uint32(reply.SwIfIndex)
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Delete_Vxlan_Tunnel(id uint32) ProcessResult {
	req := &vxlan_2110.VxlanAddDelTunnel{
		IsAdd:          false,
		McastSwIfIndex: interface_types_2110.InterfaceIndex(id),
	}
	reply := &vxlan_2110.VxlanAddDelTunnelReply{}
	for {
		// todo: unconfig of timesover
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to delete vxlan tunnel, retry, err:", err)
		} else {
			p.Log.Infoln("deleted vxlan tunnel:", "id:", id)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
	}
}

func (p *Plugin) XConnect(tx_id uint32, rx_id uint32) ProcessResult {
	req := &l2_2110.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: interface_types_2110.InterfaceIndex(rx_id),
		TxSwIfIndex: interface_types_2110.InterfaceIndex(tx_id),
		Enable:      true,
	}
	reply := &l2_2110.SwInterfaceSetL2XconnectReply{}

	count := 0
	for {
		if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to xconnect interfaces, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to xconnect interfaces")
				return TimesOver
			}
		} else {
			p.Log.Infoln("xconnect:", "tx_id:", tx_id, "rx_id:", rx_id)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Xconnect(pod_name string, tx_id uint32, rx_id uint32) ProcessResult {
	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &l2_2009.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: interface_types_2009.InterfaceIndex(rx_id),
		TxSwIfIndex: interface_types_2009.InterfaceIndex(tx_id),
		Enable:      true,
	}
	reply := &l2_2009.SwInterfaceSetL2XconnectReply{}

	count := 0
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to xconnect pod-side interfaces for", pod_name, ", retrying, message is:", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to xconnect pod-side interfaces for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver
			}
		} else {
			p.Log.Infoln("xconnected for", pod_name, ": tx_id:", tx_id, "rx_id:", rx_id)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Bridge(pod_name string, ints_id []uint32, bridge_id uint32) ProcessResult {
	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &l2_2009.BridgeDomainAddDel{
		BdID:    bridge_id,
		Learn:   true,
		IsAdd:   true,
		Flood:   true,
		UuFlood: true,
		Forward: true,
	}
	reply := &l2_2009.BridgeDomainAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to create bridge domain")
		panic(err)
	}

	p.Log.Infoln("created bridge domain id:", bridge_id, "in pod:", pod_name)

	for _, id := range ints_id {
		req := &l2_2009.SwInterfaceSetL2Bridge{
			RxSwIfIndex: interface_types_2009.InterfaceIndex(id),
			BdID:        bridge_id,
			Enable:      true,
		}
		reply := &l2_2009.SwInterfaceSetL2BridgeReply{}

		count := 0
		for {
			if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
				p.Log.Warningln("failed to bridge interface ", id, "to domain", bridge_id, "in pod:", pod_name, "err:", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to bridge interface for pod", pod_name, ", imply pod has been restarted, skip for now")
					return TimesOver
				}
			} else {
				p.Log.Infoln("bridged interface ", id, "to domain", bridge_id, "in pod:", pod_name)
				break
			}
			time.Sleep(RETRY_TIME_INTERVAL)
			count += 1
		}
	}
	return Success
}

func (p *Plugin) Set_Interface_Ip(int_id uint32, ip IpNet) ProcessResult {
	req := &interfaces_2110.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types_2110.InterfaceIndex(int_id),
		IsAdd:     true,
		Prefix: ip_types_2110.AddressWithPrefix(
			ip_types_2110.Prefix{
				Address: ip_types_2110.Address{
					Af: 0,
					Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(ip.parse_ipv4_address())),
				},
				Len: uint8(ip.Mask),
			},
		),
	}
	reply := &interfaces_2110.SwInterfaceAddDelAddressReply{}

	if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to set interface address")
		return TimesOver
	}

	p.Log.Infoln("set interface address")
	return Success
}

func (p *Plugin) Create_Tap() ProcessResult {
	req := &tap_2110.TapCreateV2{
		ID: 0,
	}
	reply := &tap_2110.TapCreateV2Reply{}

	if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to create host tap interface")
		return TimesOver
	}

	p.Log.Infoln("created host tap interface")
	return Success
}

func (p *Plugin) Pod_Create_Tap(pod_name string) (ProcessResult, uint32) {
	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &tap_2009.TapCreateV2{
		ID:  0,
		Tag: "test",
	}
	reply := &tap_2009.TapCreateV2Reply{}

	count := 0
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if strings.Contains(err.Error(), "already exist") {
				p.Log.Info("tap interface already exist")
				return ProcessResult("AlreadyExist"), 0
			} else {
				p.Log.Warningln("failed to create pod tap interface, retrying", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to create tap interface for pod", pod_name)
					return ProcessResult("TimesOver"), 0
				}
			}
		} else {
			p.Log.Infoln("created pod tap interface for pod", pod_name)
			return ProcessResult("Success"), uint32(reply.SwIfIndex)
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
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

func (ip IpNet) parse_ipv4_address() [4]uint8 {
	ip_addr := strings.Split(ip.Ip, ".")
	ip_addr_slice := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		conv, err := strconv.Atoi(ip_addr[i])
		if err != nil {
			panic("error parsing dst ip address string to int")
		}
		ip_addr_slice[i] = uint8(conv)
	}
	return [4]uint8{
		ip_addr_slice[0],
		ip_addr_slice[1],
		ip_addr_slice[2],
		ip_addr_slice[3],
	}
}

func (p *Plugin) Add_Route(route Route_Info) ProcessResult {
	req := &ip_2110.IPRouteAddDel{
		// Multi path is always true
		IsMultipath: true,
		IsAdd:       true,
	}

	fibPath := fib_types_2110.FibPath{}

	if route.Gw.Ip != "" {
		fibPath.Nh = fib_types_2110.FibPathNh{
			Address:            ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(route.Gw.parse_ipv4_address())),
			ClassifyTableIndex: ClassifyTableIndexUnset,
		}
		fibPath.Proto = fib_types_2110.FIB_API_PATH_NH_PROTO_IP4
	}

	prefix := ip_types_2110.Prefix{
		Address: ip_types_2110.Address{
			Af: ip_types_2110.ADDRESS_IP4,
			Un: ip_types_2110.AddressUnionIP4(route.Dst.parse_ipv4_address()),
		},
		Len: uint8(route.Dst.Mask),
	}

	if route.Dev != "" {
		fibPath.SwIfIndex = route.DevId
	}

	req.Route = ip_2110.IPRoute{
		Prefix: prefix,
		NPaths: 1,
		Paths:  []fib_types_2110.FibPath{fibPath},
	}

	reply := &ip_2110.IPRouteAddDelReply{}

	if err := p.Channel.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to add vpp route, dst:", route.Dst, ", via:", route.Gw, ", dev:", route.Dev)
		return TimesOver
	}
	p.Log.Infoln("added vpp route, dst:", route.Dst, ", via:", route.Gw, ", dev:", route.Dev)
	return Success
}

func (p *Plugin) Pod_Add_Route(pod_name string, route Route_Info) ProcessResult {
	conn, ch := p.connect_to_pod_vpp(pod_name)
	defer conn.Disconnect()
	defer ch.Close()

	req := &ip_2009.IPRouteAddDel{
		// Multi path is always true
		IsMultipath: true,
		IsAdd:       true,
	}

	fibPath := fib_types_2009.FibPath{}

	if route.Gw.Ip != "" {
		fibPath.Nh = fib_types_2009.FibPathNh{
			Address:            ip_types_2009.AddressUnionIP4(ip_types_2009.IP4Address(route.Gw.parse_ipv4_address())),
			ClassifyTableIndex: ClassifyTableIndexUnset,
		}
		fibPath.Proto = fib_types_2009.FIB_API_PATH_NH_PROTO_IP4
	}

	prefix := ip_types_2009.Prefix{
		Address: ip_types_2009.Address{
			Af: ip_types_2009.ADDRESS_IP4,
			Un: ip_types_2009.AddressUnionIP4(route.Dst.parse_ipv4_address()),
		},
		Len: uint8(route.Dst.Mask),
	}

	if route.Dev != "" {
		fibPath.SwIfIndex = route.DevId
	}

	req.Route = ip_2009.IPRoute{
		Prefix: prefix,
		NPaths: 1,
		Paths:  []fib_types_2009.FibPath{fibPath},
	}

	reply := &ip_2009.IPRouteAddDelReply{}

	count := 0
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to add ip route for pod", pod_name, "err:", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to add ip route for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver
			}
		} else {
			p.Log.Infoln("added ip route for pod", pod_name)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) connect_to_pod_vpp(pod_name string) (*core.Connection, api.Channel) {
	conn, conev, err := govpp.AsyncConnect(pod_sock+pod_name+"/api.sock", core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		p.Log.Errorln("ERROR:", err, "pod:", pod_name)
	}

	select {
	case e := <-conev:
		if e.State != core.Connected {
			p.Log.Errorln("ERROR: connecting to pod-side VPP failed or interrupted:", e.Error, "pod:", pod_name)
		}
	}

	ch, err := conn.NewAPIChannel()
	if err != nil {
		p.Log.Errorln("error when connect to pod-side vpp", "pod:", pod_name)
	}
	return conn, ch
}
