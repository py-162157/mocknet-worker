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
	"os"
	"time"

	//"os"
	"strconv"
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"

	//"git.fd.io/govpp.git/binapi/interfaces"
	//"git.fd.io/govpp.git/binapi/memif"

	//"git.fd.io/govpp.git/binapi/vxlan"
	"git.fd.io/govpp.git/core"
	l2_2009 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/l2"
	fib_types_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/fib_types"
	interfaces_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	interface_types_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	ip_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip"
	ip_types_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	memif_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/memif"
	vxlan_2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vxlan"
)

var (
	sockAddr = flag.String("sock", "/run/vpp/api.sock", "Path to VPP binary API socket file")
	//sockAddr = flag.String("sock", "/var/run/mocknet/h1/api.sock", "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting simple client example")

	// connect to VPP asynchronously
	conn, conev, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
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
	select {
	case e := <-conev:
		if e.State != core.Connected {
			log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
		}
	}

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
	//CreateMemifInterface(ch)
	//interface_state(ch)
	//create_vxlan_tunnel(ch)
	//get_interfaces(ch)
	//xconnect(ch)
	//vxlan(ch)
	//Set_Interface_Ip(ch)
	Add_Route(ch)

	if err := ch.CheckCompatiblity(vxlan_2106.AllMessages()...); err != nil {
		fmt.Println("error!", err)
	} else {
		fmt.Println("interfaces compatiblity check passed")
	}

	/*if err := ch.CheckCompatiblity(memif.AllMessages()...); err != nil {
		fmt.Println("error!", err)
	} else {
		fmt.Println("memif compatiblity check passed")
	}*/

	if len(Errors) > 0 {
		fmt.Printf("finished with %d errors\n", len(Errors))
		os.Exit(1)
	} else {
		fmt.Println("finished successfully")
	}
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

	req := &memif_2106.MemifSocketFilenameAddDel{
		IsAdd:          true,
		SocketID:       5,
		SocketFilename: "/home/ubuntu/test.sock",
	}
	reply := &memif_2106.MemifSocketFilenameAddDelReply{}

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
	req := &vxlan_2106.VxlanAddDelTunnel{
		IsAdd:    true,
		Instance: 5,
		SrcAddress: ip_types_2106.Address{
			Af: 0,
			Un: ip_types_2106.AddressUnionIP4(ip_types_2106.IP4Address{
				1, 1, 1, 1,
			}),
		},
		DstAddress: ip_types_2106.Address{
			Af: 0,
			Un: ip_types_2106.AddressUnionIP4(ip_types_2106.IP4Address{
				1, 1, 1, 2,
			}),
		},
		Vni:            5,
		DecapNextIndex: 1,
	}
	reply := &vxlan_2106.VxlanAddDelTunnelReply{}
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

	req := &interfaces_2106.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types_2106.InterfaceIndex(int_id),
		IsAdd:     true,
		Prefix: ip_types_2106.AddressWithPrefix(
			ip_types_2106.Prefix{
				Address: ip_types_2106.Address{
					Af: 0,
					Un: ip_types_2106.AddressUnionIP4(ip_types_2106.IP4Address{
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
	reply := &interfaces_2106.SwInterfaceAddDelAddressReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to set interface address")
		panic(err)
	}

	fmt.Println("successfully set interface address")

	return nil
}

func Add_Route(ch api.Channel) error {
	req := &ip_2106.IPRouteAddDel{
		IsAdd:       true,
		IsMultipath: true,
		Route: ip_2106.IPRoute{
			Prefix: ip_types_2106.Prefix{
				Address: ip_types_2106.Address{
					Af: ip_types_2106.ADDRESS_IP4,
					Un: ip_types_2106.AddressUnionIP4(ip_types_2106.IP4Address(
						[4]uint8{5, 5, 5, 5},
					)),
				},
				Len: 24,
			},
			NPaths: 1,
			Paths: []fib_types_2106.FibPath{
				{
					SwIfIndex: 2,
					Proto:     fib_types_2106.FIB_API_PATH_NH_PROTO_IP4,
					Nh: fib_types_2106.FibPathNh{
						Address: ip_types_2106.AddressUnionIP4(ip_types_2106.IP4Address(
							[4]uint8{192, 168, 122, 101},
						)),
					},
				},
			},
		},
	}
	reply := &ip_2106.IPRouteAddDelReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("failed to add route")
		panic(err)
	}

	fmt.Println("successfully added route")

	return nil
}
