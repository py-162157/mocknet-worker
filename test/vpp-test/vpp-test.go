// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// simple-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	//"os"
	"strconv"
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"

	//"git.fd.io/govpp.git/binapi/interfaces"
	//"git.fd.io/govpp.git/binapi/memif"

	//"git.fd.io/govpp.git/binapi/vxlan"
	arp_2009 "mocknet/binapi/vpp2009/arp"
	ip_types_2009 "mocknet/binapi/vpp2009/ip_types"
	l2_2009 "mocknet/binapi/vpp2009/l2"
	tap_2009 "mocknet/binapi/vpp2009/tapv2"
	arp_2110 "mocknet/binapi/vpp2110/arp"
	fib_types_2110 "mocknet/binapi/vpp2110/fib_types"
	interfaces_2110 "mocknet/binapi/vpp2110/interface"
	interface_types_2110 "mocknet/binapi/vpp2110/interface_types"
	ip_2110 "mocknet/binapi/vpp2110/ip"
	ip_types_2110 "mocknet/binapi/vpp2110/ip_types"
	memif_2110 "mocknet/binapi/vpp2110/memif"
	vxlan_2110 "mocknet/binapi/vpp2110/vxlan"
)

var (
	//sockAddr = flag.String("sock", "/run/vpp/api.sock", "Path to VPP binary API socket file")
	sockAddr = flag.String("sock", "/var/run/mocknet/s2/api.sock", "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	// connect to VPP asynchronously
	conn, err := govpp.Connect(*sockAddr)
	//vppclient := conn.VppClient
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	/*msgid, err := vppclient.GetMsgID("create_loopback_instance_reply", "5383d31f")
	if err != nil {
		log.Fatalln("ERROR:", err)
	} else {
		fmt.Println("the msg id is", msgid)
	}*/
	defer conn.Disconnect()

	// wait for Connected event

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch.Close()

	//vppVersion(ch)

	//createLoopback(ch)
	//createLoopback(ch)
	//interfaceDump(ch)

	//addIPAddress(ch)
	//ipAddressDump(ch)

	//interfaceNotifications(ch)
	//CreateSocket(ch)
	arp_test_2009(ch)
	//CreateMemifInterface(ch)
	//interface_state(ch)
	//create_vxlan_tunnel(ch)
	//get_interfaces(ch)
	//xconnect(ch)
	//vxlan(ch)
	//Set_Interface_Ip(ch)
	//Add_Route(ch)
	//Create_Tap(ch)
	//Delete_Vxlan_Tunnel(ch)
	//Delete_Memif_Interface(ch)

	/*if err := ch.CheckCompatiblity(memif_2110.AllMessages()...); err != nil {
		fmt.Println("error!", err)
	} else {
		fmt.Println("interfaces compatiblity check passed")
	}*/

	/*if err := ch.CheckCompatiblity(memif.AllMessages()...); err != nil {
		fmt.Println("error!", err)
	} else {
		fmt.Println("memif compatiblity check passed")
	}*/

	/*if len(Errors) > 0 {
		fmt.Printf("finished with %d errors\n", len(Errors))
		os.Exit(1)
	} else {
		fmt.Println("finished successfully")
	}*/
}

var Errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	Errors = append(Errors, err)
}

func cleanString(str string) string {
	return strings.Split(str, "\x00")[0]
}

func CreateSocket(ch api.Channel) {
	fmt.Println("Creating memif socket")

	req := &memif_2110.MemifSocketFilenameAddDel{
		IsAdd:          true,
		SocketID:       5,
		SocketFilename: "/home/test.sock",
	}
	reply := &memif_2110.MemifSocketFilenameAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println(err, "error creating memif socket")
		return
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Printf("memif socket index: %v\n", reply.Retval)
	fmt.Println("OK")

}

/*func CreateMemifInterface(ch api.Channel) {
	fmt.Println("Creating memif inerface")

	req := &memif.MemifCreate{
		Role:     1,
		ID:       1,
		SocketID: 0,
	}
	reply := &memif.MemifCreateReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println(err, "error creating memif interface")
		return
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Printf("memif interface index: %v\n", reply.Retval)
	fmt.Println("OK")

}*/

/*func interface_state(ch api.Channel) {
	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex: 0,
		Flags:     1,
	}
	reply := &interfaces.SwInterfaceSetFlagsReply{}
	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println(err, "error creating memif interface")
	}
	fmt.Printf("reply: %+v\n", reply)
}*/

/*func get_interfaces(ch api.Channel) {
	req := &interfaces.SwInterfaceDetails{}
	reply := &interfaces.SwInterfaceGetTableReply{}
	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println(err, "error get interface table")
	}
	fmt.Printf("reply: %+v\n", reply)

}*/

/*func create_vxlan_tunnel(ch api.Channel) {
	req := &vxlan.VxlanAddDelTunnel{
		IsAdd:    true,
		Instance: 10,
		SrcAddress: vxlan.Address{Af: 0, Un: vxlan.AddressUnionIP4(vxlan.IP4Address{
			10, 3, 0, 1,
		})},
		DstAddress: vxlan.Address{Af: 0, Un: vxlan.AddressUnionIP4(vxlan.IP4Address{
			10, 3, 0, 2,
		})},
		Vni: 4,
	}
	reply := &vxlan.VxlanAddDelTunnelReply{}
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			fmt.Println("failed to create vxlan tunnel, retry")
			time.Sleep(500000000)
			fmt.Println(err)
		} else {
			fmt.Println("successfully created vxlan tunnel")
			break
		}
	}
}*/

func pod_xconnect(ch api.Channel) {
	req := &l2_2009.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: 0,
		TxSwIfIndex: 1,
		Enable:      true,
	}
	reply := &l2_2009.SwInterfaceSetL2XconnectReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to create vxlan tunnel, retry")
		time.Sleep(500000000)
		fmt.Println(err)
	} else {
		fmt.Println("successfully created vxlan tunnel")
	}
}

func vxlan(ch api.Channel) {
	req := &vxlan_2110.VxlanAddDelTunnel{
		IsAdd:    true,
		Instance: 5,
		SrcAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				1, 1, 1, 1,
			}),
		},
		DstAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				1, 1, 1, 2,
			}),
		},
		Vni:            5,
		DecapNextIndex: 1,
	}
	reply := &vxlan_2110.VxlanAddDelTunnelReply{}
	for {
		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			fmt.Println("failed to create vxlan tunnel, retry")
			time.Sleep(500000000)
			fmt.Println(err)
		} else {
			break
		}
	}
}

func Set_Interface_Ip(ch api.Channel) error {
	ip := "10.2.0.1"
	int_id := 1
	ip_addr := strings.Split(ip, ".")
	ip_addr_slice := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		conv, err := strconv.Atoi(ip_addr[i])
		if err != nil {
			panic("error parsing dst ip address string to int")
		}
		ip_addr_slice[i] = uint8(conv)
	}

	req := &interfaces_2110.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types_2110.InterfaceIndex(int_id),
		IsAdd:     true,
		Prefix: ip_types_2110.AddressWithPrefix(
			ip_types_2110.Prefix{
				Address: ip_types_2110.Address{
					Af: 0,
					Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
						ip_addr_slice[0],
						ip_addr_slice[1],
						ip_addr_slice[2],
						ip_addr_slice[3],
					}),
				},
				Len: 24,
			},
		),
	}
	reply := &interfaces_2110.SwInterfaceAddDelAddressReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to set interface address")
		panic(err)
	}

	fmt.Println("successfully set interface address")

	return nil
}

func Add_Route(ch api.Channel) error {
	req := &ip_2110.IPRouteAddDel{
		IsAdd:       true,
		IsMultipath: true,
		Route: ip_2110.IPRoute{
			Prefix: ip_types_2110.Prefix{
				Address: ip_types_2110.Address{
					Af: ip_types_2110.ADDRESS_IP4,
					Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(
						[4]uint8{192, 168, 122, 102},
					)),
				},
				Len: 32,
			},
			NPaths: 1,
			Paths: []fib_types_2110.FibPath{
				{
					SwIfIndex: 1,
					Proto:     fib_types_2110.FIB_API_PATH_NH_PROTO_IP4,
					Nh: fib_types_2110.FibPathNh{
						Address: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address(
							[4]uint8{192, 168, 122, 102},
						)),
					},
				},
			},
		},
	}
	reply := &ip_2110.IPRouteAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to add route")
		panic(err)
	}

	fmt.Println("successfully added route")

	return nil
}

func Create_Tap(ch api.Channel) error {
	req := &tap_2009.TapCreateV2{
		ID: 0,
	}
	reply := &tap_2009.TapCreateV2Reply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to create host tap interface")
	}

	fmt.Println("successfully created host tap interface")

	return nil
}

func Delete_Vxlan_Tunnel(ch api.Channel) error {
	req := &vxlan_2110.VxlanAddDelTunnel{
		IsAdd:          false,
		McastSwIfIndex: interface_types_2110.InterfaceIndex(14),
		SrcAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				10, 1, 1, 1,
			}),
		},
		DstAddress: ip_types_2110.Address{
			Af: 0,
			Un: ip_types_2110.AddressUnionIP4(ip_types_2110.IP4Address{
				10, 1, 1, 2,
			}),
		},
	}
	reply := &vxlan_2110.VxlanAddDelTunnelReply{}
	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to delete vxlan tunnel, retry")
		fmt.Println(err)
	} else {
		fmt.Println("deleted vxlan tunnel:", "id:", 14)
	}
	return nil
}

func Delete_Memif_Interface(ch api.Channel) error {
	req := &memif_2110.MemifDelete{
		SwIfIndex: interface_types_2110.InterfaceIndex(100),
	}
	reply := &memif_2110.MemifDeleteReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to delete memif interface, id =", 100)
		panic(err)
	}

	fmt.Println("deleted interface id:", 100)
	return nil
}

func arp_test_2110(ch api.Channel) {
	req := &arp_2110.ProxyArpAddDel{
		IsAdd: true,
		Proxy: arp_2110.ProxyArp{
			TableID: 0,
			Low:     ip_types_2110.IP4Address{10, 1, 0, 1},
			Hi:      ip_types_2110.IP4Address{10, 2, 0, 1},
		},
	}

	reply := &arp_2110.ProxyArpAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to set arp proxy")
		panic(err)
	}

	fmt.Println("set arp proxy")
}

func arp_test_2009(ch api.Channel) {
	req := &arp_2009.ProxyArpAddDel{
		IsAdd: true,
		Proxy: arp_2009.ProxyArp{
			TableID: 0,
			Low:     ip_types_2009.IP4Address{10, 1, 0, 1},
			Hi:      ip_types_2009.IP4Address{10, 2, 0, 1},
		},
	}

	reply := &arp_2009.ProxyArpAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to set arp proxy")
		panic(err)
	}

	fmt.Println("set arp proxy")
}
