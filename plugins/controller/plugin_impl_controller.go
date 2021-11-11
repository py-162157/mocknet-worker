package controller

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	//"mocknet/plugins/server/rpctest"

	"mocknet/plugins/vpp"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"
)

var (
	TIMEOUT             = time.Duration(3) * time.Second // 超时
	MASTER_IP_ADDRESS   = "192.168.122.100"
	LOCAL_IP_ADDRESS    = "192.168.122.101"
	ETCD_PORT           = "22379"
	MASTER_ETCD_ADDRESS = []string{MASTER_IP_ADDRESS + ":" + ETCD_PORT}
	LOCAL_ETCD_ADDRESS  = []string{LOCAL_IP_ADDRESS + ":" + ETCD_PORT}
	vxlan_instance      = 0
)

type Plugin struct {
	Deps

	PluginName       string
	K8sNamespace     string
	MasterEtcdClient *clientv3.Client
	LocalEtcdClient  *clientv3.Client
}

type Deps struct {
	Vpp             *vpp.Plugin
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
	IntToVppId map[string]uint32

	Log logging.PluginLogger
}

type DepAdress struct {
	master_ip_address   string
	local_ip_address    string
	local_vtep_address  string
	master_etcd_address []string
	local_etcd_address  []string
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
		Endpoints:   MASTER_ETCD_ADDRESS,
		DialTimeout: 10 * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		panic(err)
	} else {
		p.MasterEtcdClient = client
		p.Log.Infoln("successfully create new etcd client!")
	}

	// connect with local etcd
	if client, err := clientv3.New(clientv3.Config{
		Endpoints:   LOCAL_ETCD_ADDRESS,
		DialTimeout: 5 * time.Second,
	}); err != nil {
		p.Log.Errorln(err)
		panic(err)
	} else {
		p.LocalEtcdClient = client
		p.Log.Infoln("successfully connected to local etcd!")
	}

	local_hostname, err := os.Hostname()
	p.get_node_infos(local_hostname)

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
		panic(err)
	}
	if os.IsNotExist(err) {
		os.MkdirAll(p.DirPrefix, 0777)
	}
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
					p.inform_finished("ParseTopologyInfo")
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
	kv.Put(context.Background(), "/mocknet/event/"+event+"/"+hostname, "done")
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
	ctx1, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
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

					// pod-side config
					for _, pod := range p.MocknetTopology.pods {
						if p.Is_Local(pod.name) {
							p.Log.Infoln("pod", pod.name, "is local")
							// create memif socket
							p.assign_dir_to_pod(pod.name)
							id := p.get_socket_id(pod.name) + 1
							p.Log.Infoln("pod name:", pod.name)
							filedir := p.DirPrefix + pod.name
							p.Vpp.CreateSocket(uint32(id), filedir)

							// create memif interface
							pod.alloc_interface_id()

							p.Log.Infoln("pod.infs =", pod.infs)

							for _, vpp_id := range pod.infs {
								// host-side
								global_id := p.Vpp.CreateMemifInterface("master", uint32(vpp_id), uint32(id))
								p.Vpp.Set_interface_state_up(global_id)
								p.InterfaceIDName[global_id] = "memif" + strconv.Itoa(id) + "/" + strconv.Itoa(vpp_id)
								p.InterfaceNameID["memif"+strconv.Itoa(id)+"/"+strconv.Itoa(vpp_id)] = global_id
								p.IntToVppId[pod.name+"-"+strconv.Itoa(vpp_id)] = global_id

								// pod-side
								inf_name := "memif" + strconv.Itoa(vpp_id)
								// int_id is the id in a socket used to identify peer(slave or master)
								inf_id := strconv.Itoa(vpp_id)
								p.Create_Podside_Interface(pod.name, inf_name, inf_id)
							}

							// write pod-vpp to pod pod-linux route
							kv := clientv3.NewKV(p.MasterEtcdClient)
							ip_address := p.PodInfos[pod.name].podip
							key := "/vnf-agent/mocknet-pod-" + pod.name + "/config/vpp/v2/route/vrf/0/dst/" + ip_address + "/32/gw/" + ip_address
							value := "{\"dst_network\":\"" + ip_address + "/32\",\"next_hop_addr\":\"" + ip_address + "\",\"outgoing_interface\":\"tap0\"}"
							kv.Put(context.Background(), key, value)

							// write pod-vpp-tap to pod-vpp-memif route
							if string([]byte(pod.name)[:1]) == "h" {

								/*
									// currently, all host have only 1 interface and named "memif0/0"
									net_address := "10.1.0.0"
									key = "/vnf-agent/mocknet-pod-" + pod.name + "/config/vpp/v2/route/if/memif0/0/vrf/0/dst/" + net_address + "/16/gw/" + net_address
									value = "{\"dst_network\":\"" + net_address + "/16\",\"next_hop_addr\":\"" + net_address + "\",\"outgoing_interface\":\"memif0/0\"}"
									kv.Put(context.Background(), key, value)
								*/

								// or l2xconnect them together
								/*
									key = "/vnf-agent/mocknet-pod-" + pod.name + "/config/vpp/l2/v2/xconnect/tap0"
									value = "{\"transmit_interface\":\"memif0/0\", \"receive_interface\":\"tap0\"}"
									kv.Put(context.Background(), key, value)

									key = "/vnf-agent/mocknet-pod-" + pod.name + "/config/vpp/l2/v2/xconnect/memif0/0"
									value = "{\"transmit_interface\":\"tap0\", \"receive_interface\":\"memif0/0\"}"
									kv.Put(context.Background(), key, value)
								*/

								// config l2xconnect by govpp
								p.Vpp.Pod_Xconnect(pod.name, 1, 2)
								p.Vpp.Pod_Xconnect(pod.name, 2, 1)

							} else if string([]byte(pod.name)[:1]) == "s" {

								// currently, set switch route as single topology in mininet in case of test
								// in switch pod-vpp, interface name is memif0/ + id(same with id in host-vpp)
								/*for _, link := range p.MocknetTopology.links {
									// choose switch pod's link
									if link.pod1 == pod.name {
										dst_ip_address := p.PodInfos[link.pod2].podip
										int_id := pod.infs[link.pod1inf]
										int_name := "memif0/" + strconv.Itoa(int_id)
										key = "/vnf-agent/mocknet-pod-" + pod.name + "/config/vpp/v2/route/vrf/0/dst/" + dst_ip_address + "/32/gw/" + dst_ip_address
										value = "{\"dst_network\":\"" + dst_ip_address + "\",\"next_hop_addr\":\"" + dst_ip_address + "\",\"outgoing_interface\":\"" + int_name + "\"}"
										kv.Put(context.Background(), key, value)
									}
								}*/

								// bridge interfaces in switch pod together
								// manually
								ints_id := make([]uint32, 0)
								for i := 0; i < len(p.MocknetTopology.pods[pod.name].infs); i++ {
									ints_id = append(ints_id, uint32(i+2))
								}
								p.Vpp.Pod_Bridge(pod.name, ints_id, 1)

							} else {
								p.Log.Errorln("the judgement character is ", string([]byte(pod.name)[:1]))
								panic("pod type judgement error!")
							}
						}
					}

					for _, link := range p.MocknetTopology.links {
						if p.Is_Local(link.pod1) && p.Is_Local(link.pod2) {
							p.Vpp.XConnect(p.IntToVppId[link.pod1+"-"+link.pod1inf], p.IntToVppId[link.pod2+"-"+link.pod2inf])
						} else if p.Is_Local(link.pod1) && !p.Is_Local(link.pod2) {
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

					// host-side vpp config
					// create vxlan tunnel and xconnect it with memif interface
					/*for memif_id, _ := range p.InterfaceIDName {
						// take memif interface id as tunnel vni
						vxlan_id := p.Vpp.Create_Vxlan_Tunnel(VXLAN_SRC_ADDRESS, VXLAN_DST_ADDRESS, memif_id, memif_id)
						p.Vpp.XConnect(vxlan_id, memif_id)
						p.Vpp.XConnect(memif_id, vxlan_id)
					}*/
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
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					if string(ev.Kv.Value) == "done" {
						p.inform_finished("ParsePodsInfo")
					} else {
						parse_result := parse_pod_info(*ev)
						p.PodInfos[parse_result.name] = parse_result
					}
				} else if ev.IsModify() {
					if string(ev.Kv.Value) == "done" {
						p.inform_finished("ParsePodsInfo")
					} else {
						parse_result := parse_pod_info(*ev)
						p.PodInfos[parse_result.name] = parse_result
					}
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
	if p.PodInfos[name].hostip == LOCAL_IP_ADDRESS {
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

func (p *Plugin) Create_Podside_Interface(podname string, name string, id string) error {
	key := "/vnf-agent/" + "mocknet-pod-" + podname + "/config/vpp/v2/interfaces/memif" + id
	value := "{\"name\":\"" + name + "\",\"type\":\"MEMIF\",\"enabled\":true,\"memif\":{\"id\":" + id + ",\"socket_filename\":\"/run/vpp/memif.sock\"}}"
	kv := clientv3.NewKV(p.LocalEtcdClient)

	_, err := kv.Put(context.Background(), key, value)
	if err != nil {
		panic(err)
	} else {
		p.Log.Infoln("successfully commit pod-side config to etcd")
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

func (p *Plugin) get_node_infos(local_hostname string) error {
	p.Log.Infoln("waiting for data from master")
	kv := clientv3.NewKV(p.LocalEtcdClient)
	var resp *clientv3.GetResponse
	for {
		temp, err := kv.Get(context.Background(), "/mocknet/nodeinfo/", clientv3.WithPrefix())
		if err != nil {
			panic(err)
		}
		if len(temp.Kvs) != 0 {
			resp = temp
			break
		}
	}
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
			p.Addresses.local_etcd_address = []string{nodeip + ":" + ETCD_PORT}
			p.Vpp.Set_interface_state_up(1)
			p.Log.Infoln(vtepip)
			p.Vpp.Set_Interface_Ip(1, vtepip, 24)
		} else if name == "master" {
			p.Addresses.master_ip_address = nodeip
			p.Addresses.master_etcd_address = []string{nodeip + ":" + ETCD_PORT}
		}

	}

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
