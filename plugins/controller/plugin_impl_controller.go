package controller

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"mocknet/plugins/linux"
	"mocknet/plugins/vpp"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"
)

var (
	TIMEOUT              = time.Duration(3) * time.Second // 超时
	ETCD_ADDRESS         = "0.0.0.0:32379"
	vxlan_instance       = 0
	POD_VPP_CONFIG_FILE  = "/etc/vpp/vpp.conf"
	HOST_VPP_CONFIG_FILE = "/etc/vpp/startup.conf"
)

type Plugin struct {
	Deps

	PluginName       string
	K8sNamespace     string
	MasterEtcdClient *clientv3.Client
}

type Deps struct {
	Vpp             *vpp.Plugin
	Linux           *linux.Plugin
	MocknetTopology NetTopo
	PodInfos        map[string]podinfo
	Nodeinfos       map[string]Nodeinfo
	PodSet          map[string]map[string]string
	DirAssign       []string // really needed?
	DirPrefix       string
	ReadyCount      int
	// key: interface name (memifx/x)
	// value: interface id in host-side vpp
	InterfaceNameID map[string]uint32
	// key: interface id in host-side vpp
	// value: interface name (memifx/x)
	InterfaceIDName map[uint32]string
	Addresses       DepAdress
	// key: interface name (podname-interfaceid)
	// value: interface id in host-side vpp
	IntToVppId    map[string]uint32
	LocalHostName string

	Log logging.PluginLogger
}

type DepAdress struct {
	master_ip_address  string
	local_ip_address   string
	local_vtep_address string
	host_net_addr_24   string
}

type Nodeinfo struct {
	name   string
	nodeip string
	vtepip string
}

type podinfo struct {
	name      string
	namespace string
	podip     string
	hostip    string
	hostname  string
}

type NetTopo struct {
	pods  map[string]Pod
	links []Link
}

type Pod struct {
	name string
	// key: interface name in mininet,
	// value: interface id in vpp
	infs map[string]int
}

type Link struct {
	name    string
	pod1    string
	pod2    string
	pod1inf string
	pod2inf string
	vni     int
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	p.pod_vpp_core_bind()
	//p.host_vpp_core_bind()

	p.DirPrefix = "/var/run/mocknet/"
	p.create_work_directory(p.DirPrefix)

	p.MocknetTopology = NetTopo{
		pods:  make(map[string]Pod),
		links: make([]Link, 0),
	}
	p.PodSet = make(map[string]map[string]string)
	p.PodInfos = make(map[string]podinfo)
	p.Nodeinfos = make(map[string]Nodeinfo)
	p.DirAssign = make([]string, 0)
	p.ReadyCount = 0
	p.InterfaceIDName = make(map[uint32]string)
	p.InterfaceNameID = make(map[string]uint32)
	p.IntToVppId = make(map[string]uint32)

	// connect to master-etcd client
	config := clientv3.Config{
		Endpoints: []string{
			ETCD_ADDRESS,
		},
		DialTimeout: 10 * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		panic(err)
	} else {
		p.MasterEtcdClient = client
		p.Log.Infoln("successfully connected to master etcd")
	}

	if local_hostname, err := os.Hostname(); err != nil {
		panic(err)
	} else {
		p.LocalHostName = local_hostname
	}

	p.get_node_infos(p.LocalHostName)
	p.generate_etcd_config()

	go p.watch_topology(context.Background())
	go p.watch_pods_info(context.Background())
	go p.watch_transport_done(context.Background())

	return nil
}

func (p *Plugin) String() string {
	return "controller"
}

func (p *Plugin) Close() error {
	p.clear_socket()
	return nil
}

func (p *Plugin) create_work_directory(dir string) error {
	_, err := os.Stat(p.DirPrefix)
	if err == nil {
		p.Log.Infoln("work path '/var/run/mocknet/' already exist, no need to create again")
		return nil
	}
	if os.IsNotExist(err) {
		os.MkdirAll(p.DirPrefix, 0777)
	}
	return nil
}

func (p Plugin) generate_etcd_config() error {
	ip_cmd := `NODE_IP="$(ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|grep 192|awk '{print $2}'|tr -d "addr:"​)"

	if [ ! -d "/opt/etcd" ]; then
		mkdir /opt/etcd
		cd /opt/etcd
		touch etcd.conf
		chmod 777 etcd.conf
		echo "insecure-transport: true" >> etcd.conf
		echo "dial-timeout: 10000000000" >> etcd.conf
		echo "allow-delayed-start: true" >> etcd.conf
		echo "endpoints:" >> etcd.conf
		echo "  - "${NODE_IP}:32379"" >> etcd.conf
	fi`
	cmd := exec.Command("bash", "-c", ip_cmd)
	cmd.Run()

	p.Log.Infoln("successfully set etcd.conf!")

	return nil
}

func (p *Plugin) watch_topology(ctx context.Context) error {
	ctx1, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	var getResp *clientv3.GetResponse
	var err error
	getResp, err = p.MasterEtcdClient.Get(ctx1, "/mocknet/link/", clientv3.WithPrefix())

	if err != nil {
		return err
	}

	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/link/", clientv3.WithPrefix(), clientv3.WithRev(getResp.Header.Revision+1))

	for {
		select {
		case <-time.After(3 * time.Second):
			p.inform_finished("ParseTopologyInfo")
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					p.translate(*ev)
					p.topology_make()
				} else if ev.IsModify() {

				} else if ev.Type == 1 { // 1 present DELETE

				} else {
				}
			}
		}
	}
}

func (p *Plugin) inform_finished(event string) error {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	kv := clientv3.NewKV(p.MasterEtcdClient)
	kv.Put(context.Background(), "/mocknet/"+event+"/"+hostname, "done")
	return nil
}

func (p *Plugin) topology_make() error {
	for podname, infset := range p.PodSet {
		infs := make(map[string]int)
		for inf := range infset {
			// -1 represent unalloc
			infs[inf] = -1
		}
		pod := Pod{
			name: podname,
			infs: infs,
		}

		p.MocknetTopology.pods[podname] = pod
	}

	return nil
}

func (p *Plugin) watch_transport_done(ctx context.Context) error {
	ctx1 := context.Background()
	var getResp *clientv3.GetResponse
	var err error
	getResp, err = p.MasterEtcdClient.Get(ctx1, "/mocknet/topo/ready")

	if err != nil {
		return err
	}

	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/topo/ready", clientv3.WithRev(getResp.Header.Revision+1))

	for {
		select {
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				creation_count, err := strconv.Atoi(string(ev.Kv.Value))
				if err != nil {
					panic(err)
				}
				if p.ReadyCount < creation_count {
					for _, pod := range p.MocknetTopology.pods {
						if p.Is_Local(pod.name) {
							p.Log.Infoln("pod", pod.name, "is local, so start to config it")
							p.Log.Infoln("----------- configuring", pod.name, "-----------")
							// create memif socket
							p.assign_dir_to_pod(pod.name)
							id := p.get_socket_id(pod.name) + 1
							p.Log.Infoln("pod name:", pod.name)
							filedir := p.DirPrefix + pod.name
							p.Vpp.CreateSocket(uint32(id), filedir)

							// create memif interface
							pod.alloc_interface_id()

							for _, vpp_id := range pod.infs {
								// host-side
								global_id := p.Vpp.Create_Memif_Interface("master", uint32(vpp_id), uint32(id))
								p.Vpp.Set_interface_state_up(global_id)
								p.InterfaceIDName[global_id] = "memif" + strconv.Itoa(id) + "/" + strconv.Itoa(vpp_id)
								p.InterfaceNameID["memif"+strconv.Itoa(id)+"/"+strconv.Itoa(vpp_id)] = global_id
								p.IntToVppId[pod.name+"-"+strconv.Itoa(vpp_id)] = global_id

								// pod-side
								inf_name := "memif" + strconv.Itoa(vpp_id)
								// int_id is the id in a socket used to identify peer(slave or master)
								//inf_id := strconv.Itoa(vpp_id)
								//p.Create_Podside_Interface(pod.name, inf_name, inf_id)
								pod_int_id := p.Vpp.Pod_Create_Memif_Interface(pod.name, "slave", inf_name, uint32(vpp_id))
								p.Vpp.Pod_Set_interface_state_up(pod.name, pod_int_id)
							}

							// write pod-vpp-tap to pod-vpp-memif route
							if string([]byte(pod.name)[:1]) == "h" {
								// probe weather tap0 interface exist
								result, tap_id := p.Vpp.Pod_Create_Tap(pod.name)
								if result {
									p.Vpp.Pod_Set_interface_state_up(pod.name, uint32(tap_id))
									// recreate tap interface successfully, need to config it
									p.MasterEtcdClient.Put(context.Background(), "/mocknet/PodTapReCofiguration-"+pod.name, pod.name)
									p.wait_for_response("PodTapConfigFinished-" + pod.name)
								} else {
									tap_id = 1
									p.Vpp.Pod_Set_interface_state_up(pod.name, 1)
								}

								// write pod-vpp to pod pod-linux route
								p.Vpp.Pod_Add_Route(pod.name, vpp.Route_Info{
									Dst: vpp.IpNet{
										Ip:   p.PodInfos[pod.name].podip,
										Mask: 32,
									},
									Dev:   "tap0",
									DevId: uint32(tap_id),
								})
								p.Vpp.Pod_Xconnect(pod.name, 1, 2)
								p.Vpp.Pod_Xconnect(pod.name, 2, 1)

							} else if string([]byte(pod.name)[:1]) == "s" {
								// bridge interfaces in switch pod together
								// manually
								ints_id := make([]uint32, 0)
								for i := 0; i < len(p.MocknetTopology.pods[pod.name].infs); i++ {
									ints_id = append(ints_id, uint32(i+1))
								}
								p.Vpp.Pod_Bridge(pod.name, ints_id, 1)

							} else {
								p.Log.Errorln("the judgement character is ", string([]byte(pod.name)[:1]))
								panic("pod type judgement error!")
							}
							p.Log.Infoln("-----------", pod.name, "config finished -----------")
						}
					}

					// host-side vpp config
					p.Log.Infoln("-----------configuring host vpp -----------")
					for _, link := range p.MocknetTopology.links {
						if p.Is_Local(link.pod1) && p.Is_Local(link.pod2) {
							// for both locals, xconnect them together
							p.Vpp.XConnect(p.IntToVppId[link.pod1+"-"+link.pod1inf], p.IntToVppId[link.pod2+"-"+link.pod2inf])

						} else if p.Is_Local(link.pod1) && !p.Is_Local(link.pod2) {
							// for intf1 is local and intf2 not, create vxlan tunnel and xconnect intf1 with it
							src_address := p.Addresses.local_vtep_address
							dst_address := p.Nodeinfos[p.PodInfos[link.pod2].hostname].vtepip
							vni := uint32(link.vni)
							instance := vxlan_instance
							vxlan_id := p.Vpp.Create_Vxlan_Tunnel(src_address, dst_address, vni, uint32(instance))

							p.Vpp.XConnect(p.IntToVppId[link.pod1+"-"+link.pod1inf], vxlan_id)
							p.Vpp.XConnect(vxlan_id, p.IntToVppId[link.pod1+"-"+link.pod1inf])
							vxlan_instance += 1
						}
					}
					p.Log.Infoln("----------- host vpp config finished -----------")
					p.inform_finished("NetCreationFinished")
					// readycount + 1
					p.ReadyCount++
				}
			}
		}
	}
}

func (p *Plugin) watch_pods_info(ctx context.Context) error {
	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/pods", clientv3.WithPrefix())

	for {
		select {
		case <-time.After(3 * time.Second):
			p.inform_finished("ParsePodsInfo")
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					parse_result := parse_pod_info(*ev)
					p.PodInfos[parse_result.name] = parse_result
				} else if ev.IsModify() {
					parse_result := parse_pod_info(*ev)
					p.PodInfos[parse_result.name] = parse_result
				} else if ev.Type == 1 { // 1 present DELETE
					pod_name := strings.Split(string(ev.Kv.Key), "/")[2]
					delete(p.PodInfos, pod_name)
				} else {
				}
			}
		}
	}
}

func parse_pod_info(event clientv3.Event) podinfo {
	v := string(event.Kv.Value)
	split_info := strings.Split(v, ",")
	return podinfo{
		name:      strings.Split(split_info[0], ":")[1],
		namespace: strings.Split(split_info[1], ":")[1],
		podip:     strings.Split(split_info[2], ":")[1],
		hostip:    strings.Split(split_info[3], ":")[1],
		hostname:  strings.Split(split_info[4], ":")[1],
	}
}

func (p *Plugin) translate(event clientv3.Event) error {
	// teanslate etcd topology format
	v := string(event.Kv.Value)
	split_link := strings.Split(v, ",")
	node1 := strings.Split(split_link[0], ":")[1]
	node2 := strings.Split(split_link[1], ":")[1]
	node1_inf := strings.Split(split_link[2], ":")[1]
	node2_inf := strings.Split(split_link[3], ":")[1]
	vni, err := strconv.Atoi(strings.Split(split_link[4], ":")[1])
	if err != nil {
		panic(err)
	}

	if pod, ok := p.PodSet[node1]; ok {
		if _, ok := pod[node1_inf]; ok {

		} else {
			pod[node1_inf] = ""
		}
	} else {
		p.PodSet[node1] = make(map[string]string)
		p.PodSet[node1][node1_inf] = ""
	}

	if pod, ok := p.PodSet[node2]; ok {
		if _, ok := pod[node2_inf]; ok {

		} else {
			pod[node2_inf] = ""
		}
	} else {
		p.PodSet[node2] = make(map[string]string)
		p.PodSet[node2][node2_inf] = ""
	}

	// generate topology-links
	p.MocknetTopology.links = append(p.MocknetTopology.links,
		Link{
			name:    node1 + "-" + node2,
			pod1:    node1,
			pod2:    node2,
			pod1inf: node1_inf,
			pod2inf: node2_inf,
			vni:     vni,
		},
	)

	return nil
}

func (p *Plugin) Is_Local(name string) bool {
	if p.PodInfos[name].hostip == p.Addresses.local_ip_address {
		return true
	} else {
		return false
	}
}

func (p *Plugin) assign_dir_to_pod(name string) error {
	p.DirAssign = append(p.DirAssign, name)
	os.MkdirAll(p.DirPrefix+name, 0777)
	return nil
}

// presently the socket id qeual to index in p.DirAssign
func (p *Plugin) get_socket_id(name string) int {
	for index, pod_name := range p.DirAssign {
		if pod_name == name {
			return index
		}
	}
	// shouldn'd be panic if healthy
	panic("error: find the position of local pod name")
}

func (pod Pod) alloc_interface_id() error {
	for mininet_name, _ := range pod.infs {
		translate, err := strconv.Atoi(mininet_name)
		if err != nil {
			panic(err)
		}
		pod.infs[mininet_name] = translate
	}
	return nil
}

func (p *Plugin) clear_socket() error {
	dir, err := ioutil.ReadDir(p.DirPrefix)
	if err != nil {
		panic(err)
	}

	for _, d := range dir {
		os.RemoveAll(path.Join([]string{p.DirPrefix, d.Name()}...))
	}
	return nil
}

func (p *Plugin) wait_for_response(event string) error {
	p.Log.Infoln("waiting for event", event)
	for {
		temp, err := p.MasterEtcdClient.Get(context.Background(), "/mocknet/"+event, clientv3.WithPrefix())
		if err != nil {
			panic(err)
		}
		if len(temp.Kvs) != 0 {
			break
		}
	}
	return nil
}

func (p *Plugin) get_node_infos(local_hostname string) error {
	p.Log.Infoln("waiting for node data from master")
	p.wait_for_response("nodeinfo")
	p.Log.Infoln("----------- pre-configure for nodes -----------")
	resp, _ := p.MasterEtcdClient.Get(context.Background(), "/mocknet/nodeinfo", clientv3.WithPrefix())
	// to find local host infos
	for _, nodekv := range resp.Kvs {
		split_value := strings.Split(string(nodekv.Value), ",")
		name := strings.Split(split_value[0], ":")[1]
		nodeip := strings.Split(split_value[1], ":")[1]
		vtepip := strings.Split(split_value[2], ":")[1]
		p.Nodeinfos[name] = Nodeinfo{
			name:   name,
			nodeip: nodeip,
			vtepip: vtepip,
		}
		if name == local_hostname {
			p.Addresses.local_ip_address = nodeip
			p.Addresses.local_vtep_address = vtepip
			p.Addresses.host_net_addr_24 = get_ip_net_24(nodeip)

			p.Vpp.Create_Tap()
			p.Vpp.Set_interface_state_up(1)
			p.Vpp.Set_Interface_Ip(1, vpp.IpNet{
				Ip:   vtepip,
				Mask: 24,
			})
			p.Linux.Add_Route(linux.Route_Info{
				Dst: linux.IpNet{
					Ip:   vtepip,
					Mask: 32,
				},
				Dev: "tap0",
			})
			p.Vpp.Add_Route(vpp.Route_Info{
				Dst: vpp.IpNet{
					Ip:   nodeip,
					Mask: 32,
				},
				Dev:   "tap0",
				DevId: 1,
			})
		} else if name == "master" {
			p.Addresses.master_ip_address = nodeip
		}
	}

	for _, nodekv := range resp.Kvs {
		split_value := strings.Split(string(nodekv.Value), ",")
		name := strings.Split(split_value[0], ":")[1]
		nodeip := strings.Split(split_value[1], ":")[1]
		vtepip := strings.Split(split_value[2], ":")[1]
		if name != local_hostname && name != "master" {
			// route to other worker vtep interface
			p.Linux.Add_Route(linux.Route_Info{
				Dst: linux.IpNet{
					Ip:   vtepip,
					Mask: 32,
				},
				Gw: linux.IpNet{
					Ip: nodeip,
				},
				Dev: p.Linux.HostMainDevName,
			})
			p.Vpp.Add_Route(vpp.Route_Info{
				Dst: vpp.IpNet{
					Ip:   nodeip,
					Mask: 32,
				},
				Gw: vpp.IpNet{
					Ip:   p.Addresses.local_ip_address,
					Mask: 32,
				},
				Dev:   "tap0",
				DevId: 1,
			})
			p.Vpp.Add_Route(vpp.Route_Info{
				Dst: vpp.IpNet{
					Ip:   vtepip,
					Mask: 32,
				},
				Gw: vpp.IpNet{
					Ip:   p.Addresses.local_ip_address,
					Mask: 32,
				},
				Dev:   "tap0",
				DevId: 1,
			})
		}
	}

	// route from host-vpp to local host linux namespace
	p.Vpp.Add_Route(vpp.Route_Info{
		Dst: vpp.IpNet{
			Ip:   p.Addresses.host_net_addr_24,
			Mask: 24,
		},
		Dev:   "tap0",
		DevId: 1,
	})
	p.Log.Infoln("----------- pre-configure for nodes finished -----------")

	return nil
}

// inform other node of vxlan information.
func (p *Plugin) vxlan_info_operation(link Link) error {
	local_kv := clientv3.NewKV(p.MasterEtcdClient)
	resp, err := local_kv.Get(context.Background(), "/mocknet/vxlan/"+link.pod1)
	if err != nil {
		panic(err)
	}

	if len(resp.Kvs) == 0 {

	}

	return nil
}

func (p *Plugin) pod_vpp_core_bind() error {
	// automaticly bind vpp main thread and worker thread to cpu cores
	vppconf :=
		`
	unix {
		nodaemon
		cli-listen 0.0.0.0:5002
		cli-no-pager
		log /tmp/vpp.log
		full-coredump
	}
	plugins {
		plugin dpdk_plugin.so {
			disable
		}
	}
	api-trace {
		on
	}
	socksvr {
		socket-name /run/vpp/api.sock
	}
	statseg {
		socket-name /run/vpp/stats.sock
		per-node-counters on
	}

	cpu {
		main-core 2
		corelist-workers 3
	}
`
	_, err := os.Stat(POD_VPP_CONFIG_FILE)
	if err == nil {
		os.Remove(POD_VPP_CONFIG_FILE)
	}
	f, err := os.Create(POD_VPP_CONFIG_FILE)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err != nil {
		panic(err)
	} else {
		_, err = f.Write([]byte(vppconf))
		if err != nil {
			panic(err)
		} else {
			p.Log.Infoln("created pod vpp config file")
		}
	}
	return nil
}

func (p *Plugin) host_vpp_core_bind() error {
	vppconf :=
		`
unix {
	nodaemon
	log /var/log/vpp/vpp.log
	full-coredump
	cli-listen /run/vpp/cli.sock
	gid vpp
  }
  
  api-trace {
	on
  }
  
  api-segment {
	gid vpp
  }
  
  socksvr {
	default
  }
  
  cpu {
	skip-cores 1
	workers 1
	scheduler-policy fifo
	scheduler-priority 50
  }
`
	_, err := os.Stat(HOST_VPP_CONFIG_FILE)
	if err == nil {
		os.Remove(HOST_VPP_CONFIG_FILE)
	}
	f, err := os.Create(HOST_VPP_CONFIG_FILE)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err != nil {
		panic(err)
	} else {
		_, err = f.Write([]byte(vppconf))
		if err != nil {
			panic(err)
		} else {
			p.Log.Infoln("created host vpp config file")
		}
	}
	return nil
}

func get_ip_net_24(ip string) string {
	split_ip := strings.Split(ip, ".")
	return split_ip[0] + "." + split_ip[1] + "." + split_ip[2] + ".0"
}
