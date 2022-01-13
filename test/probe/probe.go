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
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"

	interface_2009 "mocknet/binapi/vpp2009/interface"
)

var (
	sockAddr = flag.String("sock", "/run/vpp/api.sock", "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	// connect to VPP asynchronously
	conn, err := govpp.Connect(*sockAddr)
	//vppclient := conn.VppClient
	if err != nil {
		log.Fatalln("ERROR:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
		os.Exit(1)
	}
	defer ch.Close()

	CreateSocket(ch)
}

func cleanString(str string) string {
	return strings.Split(str, "\x00")[0]
}

func CreateSocket(ch api.Channel) {
	req := &interface_2009.SwInterfaceGetTable{
		SwIfIndex: 0,
	}
	reply := &interface_2009.SwInterfaceGetTableReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	} else {
		fmt.Println("probe successed")
		os.Exit(0)
	}
}
