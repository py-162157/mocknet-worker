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

	"git.fd.io/govpp.git/core"

	fib_types_2110 "mocknet/binapi/vpp2110/fib_types"
	interfaces_2110 "mocknet/binapi/vpp2110/interface"
	interface_types_2110 "mocknet/binapi/vpp2110/interface_types"
	ip_2110 "mocknet/binapi/vpp2110/ip"
	ip_types_2110 "mocknet/binapi/vpp2110/ip_types"
	l2_2110 "mocknet/binapi/vpp2110/l2"
	memif_2110 "mocknet/binapi/vpp2110/memif"
	vxlan_2110 "mocknet/binapi/vpp2110/vxlan"

	ethernet_types_2009 "mocknet/binapi/vpp2009/ethernet_types"
	fib_types_2009 "mocknet/binapi/vpp2009/fib_types"
	interfaces_2009 "mocknet/binapi/vpp2009/interface"
	interface_types_2009 "mocknet/binapi/vpp2009/interface_types"
	ip_2009 "mocknet/binapi/vpp2009/ip"
	ip_neighbor_2009 "mocknet/binapi/vpp2009/ip_neighbor"
	ip_types_2009 "mocknet/binapi/vpp2009/ip_types"
	l2_2009 "mocknet/binapi/vpp2009/l2"
	memif_2009 "mocknet/binapi/vpp2009/memif"
	tap_2009 "mocknet/binapi/vpp2009/tapv2"
)

const (
	SOCKET_NAME         = "memif.sock"
	RETRY_TIME_INTERVAL = 1 * time.Second
	MAX_RETRY_TIMES     = 3
	KEEP_CONNECTION     = true
	MEMIF_TX_QX_QUEUES  = 1    // default = 1
	MEMIF_BUFFER_SIZE   = 0    // default = 0
	MEMIF_RING_SIZE     = 1024 // default = 1024
	TAP_NumRxQueues     = 1    // default = 1
	TAP_TX_QX_RINGSIZE  = 256  //default = 256
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
	conn         *core.Connection
}

type Deps struct {
	Log logging.PluginLogger
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	p.PluginName = "vpp"
	if KEEP_CONNECTION {
		p.conn = p.connect_to_main_vpp()
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
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when make channel to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to create memif socket, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to create memif socket")
				return TimesOver
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
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

	var conn *core.Connection
	if !KEEP_CONNECTION {
		conn = p.connect_to_main_vpp()
	} else {
		conn = p.conn
	}
	ch, err := conn.NewAPIChannel()

	if err != nil {
		p.Log.Errorln("error when connect to vpp")
		panic(err)
	}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to delete memif socket, id =", id)
		panic(err)
	}
	if !KEEP_CONNECTION {
		conn.Disconnect()
	}
	ch.Close()

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
	if MEMIF_TX_QX_QUEUES != 1 {
		req.TxQueues = MEMIF_TX_QX_QUEUES
		req.RxQueues = MEMIF_TX_QX_QUEUES
	}

	if MEMIF_RING_SIZE != 1024 {
		req.RingSize = MEMIF_RING_SIZE
	}

	if MEMIF_BUFFER_SIZE != 0 {
		req.BufferSize = MEMIF_BUFFER_SIZE
	}
	reply := &memif_2110.MemifCreateReply{}

	count := 0
	for {
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			if strings.Contains(err.Error(), "already exists") {
				p.Log.Warningln("the memif interface already exist, id:", id, "socket-id:", socket_id, "role:", role_string)
				return AlreadyExist, 0
			} else {
				p.Log.Warningln("failed to create memif interface, id:", id, "socket-id:", socket_id, "role:", role_string)
				p.Log.Warningln(err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to create memif interface")
					return TimesOver, 0
				}
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Infoln("created interface id:", id, "socket-id:", socket_id, "role:", role_string, "SwIfIndex", reply.SwIfIndex)
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

	var conn *core.Connection
	if !KEEP_CONNECTION {
		conn = p.connect_to_main_vpp()
	} else {
		conn = p.conn
	}
	ch, err := conn.NewAPIChannel()

	if err != nil {
		p.Log.Errorln("error when connect to vpp")
		panic(err)
	}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		if !KEEP_CONNECTION {
			conn.Disconnect()
		}
		ch.Close()
		if strings.Contains(err.Error(), "Invalid") {
			p.Log.Warningln("the memif interface to be deleted not exist")
			return NotExist
		} else {
			p.Log.Errorln("failed to delete memif interface, id =", id)
			return TimesOver
		}
	} else {
		if !KEEP_CONNECTION {
			conn.Disconnect()
		}
		ch.Close()
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

	req := &memif_2009.MemifCreate{
		Role:     role,
		ID:       id,
		SocketID: 0,
	}
	if MEMIF_TX_QX_QUEUES != 1 {
		req.TxQueues = MEMIF_TX_QX_QUEUES
		req.RxQueues = MEMIF_TX_QX_QUEUES
	}

	if MEMIF_RING_SIZE != 1024 {
		req.RingSize = MEMIF_RING_SIZE
	}

	if MEMIF_BUFFER_SIZE != 0 {
		req.BufferSize = MEMIF_BUFFER_SIZE
	}
	reply := &memif_2009.MemifCreateReply{}

	count := 0
	for {
		conn := p.connect_to_pod_vpp(pod_name)
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("failed to connect to pod", pod_name)
			conn.Disconnect()
			ch.Close()
			return TimesOver, 0
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			p.Log.Warningln("failed to create memif interface for pod", pod_name, ", retrying", err)
			if count >= MAX_RETRY_TIMES {
				conn.Disconnect()
				ch.Close()
				p.Log.Errorln("max retry times up to create memif interface for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver, 0
			}
		} else {
			conn.Disconnect()
			ch.Close()
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
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to set interface state up, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface state up")
				return TimesOver
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Infoln("set interface id", id, "state up")
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Set_interface_state_up(pod_name string, id uint32) ProcessResult {
	req := &interfaces_2009.SwInterfaceSetFlags{
		SwIfIndex: interface_types_2009.InterfaceIndex(id),
		Flags:     1,
	}
	reply := &interfaces_2009.SwInterfaceSetFlagsReply{}

	count := 0
	for {
		conn := p.connect_to_pod_vpp(pod_name)
		ch, err := conn.NewAPIChannel()

		if err != nil {
			conn.Disconnect()
			ch.Close()
			p.Log.Errorln("failed to connect to pod", pod_name)
			return TimesOver
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			conn.Disconnect()
			ch.Close()
			p.Log.Warningln("failed to set interface state up for pod", pod_name, ", retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface state up for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver
			}
		} else {
			conn.Disconnect()
			ch.Close()
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
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
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
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
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
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		// todo: unconfig of timesover
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to delete vxlan tunnel, retry, err:", err)
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
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
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to xconnect interfaces", "tx_id:", tx_id, "rx_id:", rx_id)
			p.Log.Warningln(err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to xconnect interfaces")
				return TimesOver
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Infoln("xconnect:", "tx_id:", tx_id, "rx_id:", rx_id)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Xconnect(pod_name string, tx_id uint32, rx_id uint32) ProcessResult {
	req := &l2_2009.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: interface_types_2009.InterfaceIndex(rx_id),
		TxSwIfIndex: interface_types_2009.InterfaceIndex(tx_id),
		Enable:      true,
	}
	reply := &l2_2009.SwInterfaceSetL2XconnectReply{}

	count := 0
	for {
		conn := p.connect_to_pod_vpp(pod_name)
		ch, err := conn.NewAPIChannel()

		if err != nil {
			conn.Disconnect()
			ch.Close()
			p.Log.Errorln("failed to connect to pod", pod_name)
			return TimesOver
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			conn.Disconnect()
			ch.Close()
			p.Log.Warningln("failed to xconnect pod-side interfaces for", pod_name, ", retrying, message is:", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to xconnect pod-side interfaces for pod", pod_name, ", imply pod has been restarted, skip for now")
				return TimesOver
			}
		} else {
			conn.Disconnect()
			ch.Close()
			p.Log.Infoln("xconnected for", pod_name, ": tx_id:", tx_id, "rx_id:", rx_id)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Bridge(pod_name string, ints_id []uint32, bridge_id uint32) ProcessResult {
	req := &l2_2009.BridgeDomainAddDel{
		BdID:    bridge_id,
		Learn:   true,
		IsAdd:   true,
		Flood:   true,
		UuFlood: true,
		Forward: true,
	}
	reply := &l2_2009.BridgeDomainAddDelReply{}

	conn := p.connect_to_pod_vpp(pod_name)
	ch, err := conn.NewAPIChannel()

	if err != nil {
		conn.Disconnect()
		ch.Close()
		p.Log.Errorln("failed to connect to pod", pod_name)
		return TimesOver
	}

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
			conn := p.connect_to_pod_vpp(pod_name)
			ch, err := conn.NewAPIChannel()

			if err != nil {
				p.Log.Errorln("failed to connect to pod", pod_name)
				return TimesOver
			}
			if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
				conn.Disconnect()
				ch.Close()
				p.Log.Warningln("failed to bridge interface ", id, "to domain", bridge_id, "in pod:", pod_name, "err:", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to bridge interface for pod", pod_name, ", imply pod has been restarted, skip for now")
					return TimesOver
				}
			} else {
				conn.Disconnect()
				ch.Close()
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
					Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(parse_ipv4_address(ip.Ip))),
				},
				Len: uint8(ip.Mask),
			},
		),
	}
	reply := &interfaces_2110.SwInterfaceAddDelAddressReply{}

	count := 0
	for {
		var conn *core.Connection
		if !KEEP_CONNECTION {
			conn = p.connect_to_main_vpp()
		} else {
			conn = p.conn
		}
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to set interface address, retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface address")
				return TimesOver
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Infoln("set interface address")
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Set_Interface_Ip(pod_name string, int_id uint32, ip IpNet) ProcessResult {
	req := &interfaces_2009.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types_2009.InterfaceIndex(int_id),
		IsAdd:     true,
		Prefix: ip_types_2009.AddressWithPrefix(
			ip_types_2009.Prefix{
				Address: ip_types_2009.Address{
					Af: 0,
					Un: ip_types_2009.AddressUnionIP4(ip_types_2009.IP4Address(parse_ipv4_address(ip.Ip))),
				},
				Len: uint8(ip.Mask),
			},
		),
	}
	reply := &interfaces_2009.SwInterfaceAddDelAddressReply{}

	count := 0
	for {
		conn := p.connect_to_pod_vpp(pod_name)
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("error when connect to pod vpp")
			panic(err)
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Warningln("failed to set interface address for pod", pod_name, ", retrying", err)
			if count >= MAX_RETRY_TIMES {
				p.Log.Errorln("max retry times up to set interface address for pod", pod_name, ",")
				return TimesOver
			}
		} else {
			if !KEEP_CONNECTION {
				conn.Disconnect()
			}
			ch.Close()
			p.Log.Infoln("set interface address for pod", pod_name)
			return Success
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

func (p *Plugin) Pod_Create_Tap(pod_name string) (ProcessResult, uint32) {
	req := &tap_2009.TapCreateV2{
		ID:  0,
		Tag: "test",
	}
	if TAP_TX_QX_RINGSIZE != 256 {
		req.RxRingSz = TAP_TX_QX_RINGSIZE
		req.TxRingSz = TAP_TX_QX_RINGSIZE
	}
	if TAP_NumRxQueues != 1 {
		req.NumRxQueues = TAP_NumRxQueues
	}
	reply := &tap_2009.TapCreateV2Reply{}

	count := 0
	for {
		conn := p.connect_to_pod_vpp(pod_name)
		ch, err := conn.NewAPIChannel()

		if err != nil {
			p.Log.Errorln("failed to connect to pod", pod_name)
			return TimesOver, 0
		}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			conn.Disconnect()
			ch.Close()
			if strings.Contains(err.Error(), "already exist") {
				p.Log.Warningln("tap interface already exist")
				return ProcessResult("AlreadyExist"), 0
			} else {
				p.Log.Warningln("failed to create pod tap interface, retrying", err)
				if count >= MAX_RETRY_TIMES {
					p.Log.Errorln("max retry times up to create tap interface for pod", pod_name)
					return ProcessResult("TimesOver"), 0
				}
			}
		} else {
			conn.Disconnect()
			ch.Close()
			p.Log.Infoln("created pod tap interface for pod", pod_name)
			return ProcessResult("Success"), uint32(reply.SwIfIndex)
		}
		time.Sleep(RETRY_TIME_INTERVAL)
		count += 1
	}
}

type Route_Info struct {
	Dst           IpNet
	Gw            IpNet
	Port          uint32
	Dev           string
	DevId         uint32
	DstName       string
	Next_hop_name string
	Local         bool
}

type IpNet struct {
	Ip   string
	Mask uint
}

func parse_ipv4_address(ipv4 string) [4]uint8 {
	ip_addr := strings.Split(ipv4, ".")
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
			Address:            ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(parse_ipv4_address(route.Gw.Ip))),
			ClassifyTableIndex: ClassifyTableIndexUnset,
		}
		fibPath.Proto = fib_types_2110.FIB_API_PATH_NH_PROTO_IP4
	}

	prefix := ip_types_2110.Prefix{
		Address: ip_types_2110.Address{
			Af: ip_types_2110.ADDRESS_IP4,
			Un: ip_types_2110.AddressUnionIP4(parse_ipv4_address(route.Dst.Ip)),
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

	var conn *core.Connection
	if !KEEP_CONNECTION {
		conn = p.connect_to_main_vpp()
	} else {
		conn = p.conn
	}
	ch, err := conn.NewAPIChannel()

	if err != nil {
		p.Log.Errorln("error when connect to vpp")
		panic(err)
	}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		p.Log.Errorln("failed to add vpp route, dst:", route.Dst, ", via:", route.Gw, ", dev:", route.Dev)
		return TimesOver
	}
	if !KEEP_CONNECTION {
		conn.Disconnect()
	}
	ch.Close()
	p.Log.Infoln("added vpp route, dst:", route.Dst, ", via:", route.Gw, ", dev:", route.Dev)
	return Success
}

func (p *Plugin) Pod_Add_Route(pod_name string, routes []Route_Info) ProcessResult {
	conn := p.connect_to_pod_vpp(pod_name)
	ch, err := conn.NewAPIChannel()

	if err != nil {
		p.Log.Errorln("failed to connect to pod", pod_name)
		return TimesOver
	}

	for _, route := range routes {
		req := &ip_2009.IPRouteAddDel{
			// Multi path is always true
			IsMultipath: true,
			IsAdd:       true,
		}

		fibPath := fib_types_2009.FibPath{}

		if route.Gw.Ip != "" {
			fibPath.Nh = fib_types_2009.FibPathNh{
				Address:            ip_types_2009.AddressUnionIP4(ip_types_2009.IP4Address(parse_ipv4_address(route.Gw.Ip))),
				ClassifyTableIndex: ClassifyTableIndexUnset,
			}
			fibPath.Proto = fib_types_2009.FIB_API_PATH_NH_PROTO_IP4
		}

		prefix := ip_types_2009.Prefix{
			Address: ip_types_2009.Address{
				Af: ip_types_2009.ADDRESS_IP4,
				Un: ip_types_2009.AddressUnionIP4(parse_ipv4_address(route.Dst.Ip)),
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

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			conn.Disconnect()
			ch.Close()
			p.Log.Warningln("failed to add ip route for pod", pod_name, "err:", err)
			return TimesOver
		}
	}

	conn.Disconnect()
	ch.Close()
	p.Log.Infoln("added ip route for pod", pod_name)
	return Success
}

type ARP struct {
	Int_id uint
	Ip     string
	Mac    string
}

func (p *Plugin) Pod_Set_Static_ARP(pod_name string, arps []ARP) ProcessResult {
	//p.Log.Infoln(arps)
	flag := true
	conn := p.connect_to_pod_vpp(pod_name)
	ch, err := conn.NewAPIChannel()
	if err != nil {
		conn.Disconnect()
		ch.Close()
		p.Log.Errorln("failed to connect to vpp of pod", pod_name)
		return TimesOver
	}
	for _, arp := range arps {
		req := &ip_neighbor_2009.IPNeighborAddDel{
			IsAdd: true,
			Neighbor: ip_neighbor_2009.IPNeighbor{
				SwIfIndex:  interface_types_2009.InterfaceIndex(arp.Int_id),
				MacAddress: ethernet_types_2009.MacAddress(p.string_to_MAC(arp.Mac)),
				IPAddress: ip_types_2009.Address{
					Af: ip_types_2009.ADDRESS_IP4,
					Un: ip_types_2009.AddressUnionIP4(parse_ipv4_address(arp.Ip)),
				},
			},
		}
		reply := &ip_neighbor_2009.IPNeighborAddDelReply{}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			conn.Disconnect()
			ch.Close()
			flag = false
			break
		}
		//p.Log.Infoln("added pod static arp for", pod_name, "ip =", arp.Ip, "intid =", arp.Int_id, "mac =", arp.Mac)
	}
	if flag {
		conn.Disconnect()
		ch.Close()
		p.Log.Infoln("added pod static arp for", pod_name)
		return Success
	} else {
		p.Log.Errorln("failed to add pod tap static arp for ", pod_name)
		return TimesOver
	}
}

func (p *Plugin) connect_to_pod_vpp(pod_name string) *core.Connection {
	conn, err := govpp.Connect(pod_sock + pod_name + "/api.sock")
	if err != nil {
		p.Log.Errorln("error when connect to pod-side vpp:", err, "pod:", pod_name)
	} else {
		//p.Log.Infoln("connected to pod", pod_name)
	}

	return conn
}

func (p *Plugin) connect_to_main_vpp() *core.Connection {
	var conn *core.Connection
	var err error
	for {
		conn, err = govpp.Connect(*sockAddr)
		if err != nil {
			if !strings.Contains(err.Error(), "temporarily") {
				p.Log.Warningln("Warning:", err)
			}
		} else {
			p.Log.Infoln("connected to main vpp")
			break
		}
		time.Sleep(time.Second)
	}

	return conn
}

func (p *Plugin) string_to_MAC(str string) [6]uint8 {
	mac_addr := [6]uint8{}
	strs := strings.Split(str, ":")
	for i, addr := range strs {
		high_byte := []byte(addr)[0]
		low_byte := []byte(addr)[1]

		total := 0
		if high_byte >= 48 && high_byte <= 57 {
			// ascii 0-9
			total += 16 * (int(high_byte) - 48)
		} else if high_byte >= 97 && high_byte <= 102 {
			// ascii a-f
			total += 16 * (int(high_byte) - 97 + 10)
		} else {
			p.Log.Println("error! the input mac addr is not Hexadecimal (16)")
		}

		if low_byte >= 48 && low_byte <= 57 {
			// ascii 0-9
			total += (int(low_byte) - 48)
		} else if low_byte >= 97 && low_byte <= 102 {
			// ascii a-f
			total += (int(low_byte) - 97 + 10)
		} else {
			p.Log.Println("error! the input mac addr is not Hexadecimal (16)")
		}
		mac_addr[i] = uint8(total)
	}
	return mac_addr
}
