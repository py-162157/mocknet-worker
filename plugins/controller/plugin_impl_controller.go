package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"mocknet/plugins/linux"
	"mocknet/plugins/vpp"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"
)

const (
	TIMEOUT              = time.Duration(3) * time.Second // 超时
	ETCD_ADDRESS         = "0.0.0.0:32379"
	POD_VPP_CONFIG_FILE  = "/etc/vpp/vpp.conf"
	HOST_VPP_CONFIG_FILE = "/etc/vpp/startup.conf"
	RETRY_TIMES          = 3
)

type Plugin struct {
	Deps

	PluginName       string
	K8sNamespace     string
	MasterEtcdClient *clientv3.Client
	CloseSignal      bool
}

type Deps struct {
	Vpp             *vpp.Plugin
	Linux           *linux.Plugin
	MocknetTopology NetTopo
	PodInfos        PodInfosSync
	NodeInfos       NodeInfosSync
	PodSet          map[string]map[string]string
	DirAssign       []string // really needed?
	DirPrefix       string
	ReadyCount      int
	InterfaceNameID map[string]uint32 // key: interface name (memifx/x), value: interface id in host-side vpp
	InterfaceIDName map[uint32]string // key: interface id in host-side vpp, value: interface name (memifx/x)
	Addresses       DepAdress
	IntToVppId      IntToVppIdSync // key: interface name (podname-interfaceid), value: interface id in host-side vpp
	LocalHostName   string
	PodConfig       map[string]*sync.RWMutex // only the routine that got the pod's lock can config it
	LocalPods       []string
	PodType         map[string]string // for fat-tree bin test, mark host as sender or receiver
	PodPair         map[string]string // pod pair for test (sender-receiver)
	TopologyType    string            // fat-tree or other
	PodIntToVppId   PodIntToVppIdSync // for pod, key: interface name (podname-interfaceid), value: interface id in host-side vpp

	Log logging.PluginLogger
}

type IntToVppIdSync struct {
	Lock *sync.Mutex
	List map[string]uint32
}
type PodIntToVppIdSync struct {
	Lock *sync.Mutex
	List map[string]uint32
}

type PodInfosSync struct {
	Lock *sync.Mutex
	List map[string]podinfo
}

type NodeInfosSync struct {
	Lock *sync.Mutex
	List map[string]Nodeinfo
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
	name         string
	namespace    string
	podip        string
	hostip       string
	hostname     string
	restartcount int
	containerid  string
}

type NetTopo struct {
	pods  map[string]*Pod
	links []Link
}

type Pod struct {
	name string
	// key: interface name in mininet,
	// value: interface id in vpp
	infs  map[string]int
	links []Link
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
		pods:  make(map[string]*Pod),
		links: make([]Link, 0),
	}
	p.PodSet = make(map[string]map[string]string)
	p.PodInfos = PodInfosSync{
		Lock: &sync.Mutex{},
		List: make(map[string]podinfo),
	}
	p.NodeInfos = NodeInfosSync{
		Lock: &sync.Mutex{},
		List: make(map[string]Nodeinfo),
	}
	p.PodIntToVppId = PodIntToVppIdSync{
		Lock: &sync.Mutex{},
		List: make(map[string]uint32),
	}
	p.DirAssign = make([]string, 0)
	p.ReadyCount = 0
	p.InterfaceIDName = make(map[uint32]string)
	p.InterfaceNameID = make(map[string]uint32)
	p.IntToVppId = IntToVppIdSync{
		Lock: &sync.Mutex{},
		List: make(map[string]uint32),
	}
	p.PodConfig = make(map[string]*sync.RWMutex)
	p.LocalPods = make([]string, 0)
	p.PodType = make(map[string]string)
	p.PodPair = make(map[string]string)
	p.TopologyType = "other"

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
		p.Log.Infoln("connected to master etcd")
	}

	if local_hostname, err := os.Hostname(); err != nil {
		panic(err)
	} else {
		p.LocalHostName = local_hostname
	}

	go p.get_node_infos(p.LocalHostName)
	go p.watch_assignment(context.Background())
	//p.generate_etcd_config()

	go p.watch_topology(context.Background())
	go p.watch_PodType(context.Background())
	go p.watch_PodPair(context.Background())
	go p.watch_receiver_ready(context.Background())
	go p.watch_test_command(context.Background())
	go p.watch_close_signal(context.Background())
	go p.watch_topo_type(context.Background())

	for {
		resp, err := p.MasterEtcdClient.Get(context.Background(), "/mocknet/ParseTopologyInfo/"+p.LocalHostName)
		if err != nil {
			p.Log.Errorln("failed to probe ParseTopologyInfo event")
		} else {
			if len(resp.Kvs) != 0 {
				if string(resp.Kvs[0].Value) == "done" {
					break
				}
			}
		}
	}

	go p.watch_pods_creation(context.Background())

	return nil
}

func (p *Plugin) String() string {
	return "controller"
}

func (p *Plugin) Close() error {
	p.clear_directory()
	return nil
}

func (p *Plugin) generate_fattree_routes() map[string][]vpp.Route_Info {
	fat_tree_routes := make(map[string][]vpp.Route_Info)

	host_numbers := 0
	for podname, _ := range p.PodSet {
		if strings.Contains(podname, "h") {
			host_numbers++
		}
	}

	k := int(math.Pow(float64(4*host_numbers), -3))
	switch_numbers := k*k/4 + k
	p.Log.Infoln("k =", k)

	cores := make([]string, k)
	aggregations := make([]string, k*k/2)
	accesses := make([]string, k*k/2)
	hosts := make([]string, k*k*k/4)

	for podname, _ := range p.PodSet {
		if !strings.Contains(podname, "h") {
			// switchs
			s, err := strconv.Atoi(strings.Split(podname, "s")[1])
			if err != nil {
				panic(err)
			}
			if s <= switch_numbers/2 {
				// left half
				if s <= k/2 {
					// core
					cores[s-1] = podname
				} else if s%2 != (k/2)%2 {
					// aggregation
					aggregations[(4*s-k)/(2*k)-1] = podname
				} else {
					// access
					accesses[(2*s-k)/k-1] = podname
				}
			} else {
				// right half
				s -= switch_numbers / 2

				if s <= k/2 {
					// core
					cores[s+k/2-1] = podname
				} else if s%2 != (k/2)%2 {
					// aggregation
					aggregations[(4*s-k)/(2*k)+k*k/4-1] = podname
				} else {
					// access
					accesses[(2*s-k)/k+k*k/4-1] = podname
				}
			}

		} else {
			// hosts
			s, err := strconv.Atoi(strings.Split(podname, "s")[1])
			if err != nil {
				panic(err)
			}
			h, err := strconv.Atoi(strings.Split(strings.Split(podname, "s")[0], "h")[1])
			if err != nil {
				panic(err)
			}
			if s <= switch_numbers/2 {
				hosts[s-k+h-1] = podname
			} else {
				hosts[s-k+k*k*k/8+h-1] = podname
			}
		}
	}

	p.PodInfos.Lock.Lock()
	p.PodIntToVppId.Lock.Lock()

	// core switch routes
	for _, core_name := range cores {
		fat_tree_routes[core_name] = make([]vpp.Route_Info, 0)
		for host_index, host_name := range hosts {
			port := host_index / k
			fat_tree_routes[core_name] = append(fat_tree_routes[core_name], vpp.Route_Info{
				Dst: vpp.IpNet{
					Ip:   p.PodInfos.List[host_name].podip,
					Mask: 16,
				},
				Dev:   "memif0/" + strconv.Itoa(port),
				DevId: p.PodIntToVppId.List[core_name+"-"+strconv.Itoa(port)],
			})
		}
	}

	// aggregation switch routes
	for aggregation_index, aggregation_name := range aggregations {
		fat_tree_routes[aggregation_name] = make([]vpp.Route_Info, 0)
		for host_index, host_name := range hosts {
			if host_index/(k*k/4) == aggregation_index/(k/2) {
				// if aggregation switch and host are in same pod, down forward the package
				// find out which access switch to forward
				pod_inner_host_id := host_index % (k * k / 4)
				port := pod_inner_host_id / (k / 2)
				fat_tree_routes[aggregation_name] = append(fat_tree_routes[aggregation_name], vpp.Route_Info{
					Dst: vpp.IpNet{
						Ip:   p.PodInfos.List[host_name].podip,
						Mask: 16,
					},
					Dev:   "memif0/" + strconv.Itoa(port),
					DevId: p.PodIntToVppId.List[aggregation_name+"-"+strconv.Itoa(port)],
				})
			} else {
				// else, up forward teh package
				// find out which core switch to forward
				port := host_index/(k*k/2) + k/2
				fat_tree_routes[aggregation_name] = append(fat_tree_routes[aggregation_name], vpp.Route_Info{
					Dst: vpp.IpNet{
						Ip:   p.PodInfos.List[host_name].podip,
						Mask: 16,
					},
					Dev:   "memif0/" + strconv.Itoa(port),
					DevId: p.PodIntToVppId.List[aggregation_name+"-"+strconv.Itoa(port)],
				})
			}
		}
	}

	// access switch routes
	for access_index, access_name := range accesses {
		fat_tree_routes[access_name] = make([]vpp.Route_Info, 0)
		for host_index, host_name := range hosts {
			if host_index/(k/2) == access_index {
				// it access switch and host are linkd, down forward the package
				// find out which port to forward
				access_inner_host_id := host_index % (k / 2)
				port := access_inner_host_id / (k / 2)
				fat_tree_routes[access_name] = append(fat_tree_routes[access_name], vpp.Route_Info{
					Dst: vpp.IpNet{
						Ip:   p.PodInfos.List[host_name].podip,
						Mask: 16,
					},
					Dev:   "memif0/" + strconv.Itoa(port),
					DevId: p.PodIntToVppId.List[access_name+"-"+strconv.Itoa(port)],
				})
			} else {
				// else, up forward teh package
				// find out which aggregation switch to forward
				port := host_index/(k*k/2) + k/2
				fat_tree_routes[access_name] = append(fat_tree_routes[access_name], vpp.Route_Info{
					Dst: vpp.IpNet{
						Ip:   p.PodInfos.List[host_name].podip,
						Mask: 16,
					},
					Dev:   "memif0/" + strconv.Itoa(port),
					DevId: p.PodIntToVppId.List[access_name+"-"+strconv.Itoa(port)],
				})
			}
		}
	}

	p.PodInfos.Lock.Unlock()
	p.PodIntToVppId.Lock.Unlock()

	return fat_tree_routes
}

func (p *Plugin) watch_close_signal(ctx context.Context) error {
	for {
		resp, err := p.MasterEtcdClient.Get(context.Background(), "/mocknet/Close")
		if err != nil {
			p.Log.Errorln("failed to probe Close signal")
		} else {
			if len(resp.Kvs) != 0 {
				if string(resp.Kvs[0].Value) == "true" {
					p.Log.Infoln("received a close signal")
					p.Linux.RestartVpp()
					p.CloseSignal = true
					p.inform_finished("Close")
					break
				}
			}
		}
	}
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

func (p *Plugin) watch_assignment(ctx context.Context) error {
	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/assignment/", clientv3.WithPrefix())

	receive_count := 0
	for {
		select {
		case <-time.After(TIMEOUT):
			if receive_count > 0 {
				p.create_directory()
				p.inform_finished("DirectoryCreationFinished")
				receive_count = 0
			}
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			receive_count += 1
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					p.parse_assignment(*ev)
				}
			}
		}
	}
}

func (p *Plugin) create_directory() {
	for _, podname := range p.LocalPods {
		os.MkdirAll(p.DirPrefix+podname, 0777)
	}
}

func (p *Plugin) watch_PodType(ctx context.Context) error {
	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/podtype/", clientv3.WithPrefix())

	receive_count := 0
	for {
		select {
		case <-time.After(TIMEOUT):
			if receive_count > 0 {
				p.inform_finished("PodTypeReceiveFinished")
				receive_count = 0
			}
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			receive_count += 1
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					p.parse_podtype(*ev)
				}
			}
		}
	}
}

func (p *Plugin) parse_podtype(event clientv3.Event) {
	// event.Kv.Value: key----/mocknet/podtype/h1s1, value----sender
	parse_result := strings.Split(string(event.Kv.Value), ",")
	name_string := parse_result[0]
	kind_string := parse_result[1]
	name := strings.Split(name_string, ":")[1]
	kind := strings.Split(kind_string, ":")[1]
	p.PodType[name] = kind
}

func (p *Plugin) watch_PodPair(ctx context.Context) error {
	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/podpair/", clientv3.WithPrefix())

	receive_count := 0
	for {
		select {
		case <-time.After(TIMEOUT):
			if receive_count > 0 {
				p.inform_finished("PodPairReceiveFinished")
				receive_count = 0
			}

		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			receive_count += 1
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					p.parse_podpair(*ev)
				}
			}
		}
	}
}

func (p *Plugin) parse_podpair(event clientv3.Event) {
	// event.Kv.Value: key----/mocknet/podpair/h1s1-h2s1, value----h1s1-h2s1
	parse_result := strings.Split(string(event.Kv.Value), "-")
	pod1 := parse_result[0]
	pod2 := parse_result[1]
	p.PodPair[pod1] = pod2
}

func (p *Plugin) watch_topo_type(ctx context.Context) error {
	watchChan := p.MasterEtcdClient.Watch(context.Background(), "/mocknet/topo/type")
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
					if string(ev.Kv.Value) == "fat-tree" {
						p.TopologyType = "fat-tree"
						p.inform_finished("TopologyType")
						return nil
					}
				}
			}
		}
	}
}

func (p *Plugin) watch_receiver_ready(ctx context.Context) error {
	for {
		resp, err := p.MasterEtcdClient.Get(context.Background(), "mocknet/command/GetReceiverReady")
		if err != nil {
			p.Log.Errorln("failed to probe GetReceiverReady command")
		} else {
			if len(resp.Kvs) != 0 {
				if string(resp.Kvs[0].Value) == "true" {
					p.get_receiver_ready()
					break
				}
			}
		}
	}
	return nil
}

func (p *Plugin) get_receiver_ready() {
	flag := true
	for _, podname := range p.LocalPods {
		if p.PodType[podname] == "receiver" {
			containerid := p.PodInfos.List[podname].containerid
			result := p.Linux.Set_Receiver(containerid, podname)
			if result != linux.Success {
				// cann't set iperf3 server for all local receiver
				flag = false
			}
		}
	}
	if flag {
		p.inform_finished("ReceiverReady")
	}
}

func (p *Plugin) watch_test_command(ctx context.Context) error {
	for {
		resp, err := p.MasterEtcdClient.Get(context.Background(), "mocknet/command/FullTest")
		if err != nil {
			p.Log.Errorln("failed to probe FullTest command")
		} else {
			if len(resp.Kvs) != 0 {
				if string(resp.Kvs[0].Value) == "true" {
					p.fulltest()
					break
				}
			}
		}
	}
	return nil
}

func (p *Plugin) fulltest() {
	//flag := true
	for _, podname := range p.LocalPods {
		if p.PodType[podname] == "sender" {
			containerid := p.PodInfos.List[podname].containerid
			paird_pod := p.PodPair[podname]
			p.PodInfos.Lock.Lock()
			dst_ip := p.PodInfos.List[paird_pod].podip
			p.Linux.Set_Sender(containerid, podname, dst_ip)
		}
	}
}

func (p *Plugin) parse_assignment(event clientv3.Event) {
	parse_result := strings.Split(string(event.Kv.Value), ",")
	if strings.Split(parse_result[1], ":")[1] == p.LocalHostName {
		p.LocalPods = append(p.LocalPods, strings.Split(parse_result[0], ":")[1])
	}
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

	receive_count := 0
	for {
		select {
		case <-time.After(TIMEOUT):
			if receive_count > 0 {
				//p.Log.Infoln("p.MocknetTopology.pods =", p.MocknetTopology.pods)
				//p.Log.Infoln("p.MocknetTopology.links =", p.MocknetTopology.links)
				p.inform_finished("ParseTopologyInfo")
				routes := p.generate_fattree_routes()
				fmt.Print(routes)
				receive_count = 0
			}
		case <-ctx.Done():
			return nil
		case resp := <-watchChan:
			receive_count += 1
			err := resp.Err()
			if err != nil {
				return err
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					//p.Log.Infoln("receive a link infomation, key =", string(ev.Kv.Key), ", value =", string(ev.Kv.Value))
					p.translate(*ev)
					p.topology_make()
				}
			}
		}
	}
}

func (p *Plugin) inform_finished(event string) error {
	kv := clientv3.NewKV(p.MasterEtcdClient)
	kv.Put(context.Background(), "/mocknet/"+event+"/"+p.LocalHostName, "done")
	p.Log.Infoln("informed the master that finished event", event)
	return nil
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

		p.MocknetTopology.pods[podname] = &pod
	}

	for _, link := range p.MocknetTopology.links {
		if pod, ok := p.MocknetTopology.pods[link.pod1]; ok {
			pod.links = append(pod.links, link)
		} else {
			panic("pod not exist in '.MocknetTopology.Pods'")
		}
	}

	return nil
}

func (p *Plugin) watch_pods_creation(ctx context.Context) error {
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
				if err != nil {
					panic(err)
				}
				if ev.IsCreate() {
					parse_result := parse_pod_info(*ev)
					p.PodConfig[parse_result.name] = &sync.RWMutex{}
					p.Log.Infoln("receive new create pod message for", parse_result.name)
					p.PodInfos.Lock.Lock()
					//p.Log.Infoln(parse_result.name, "has got the lock of PodInfos")
					p.PodInfos.List[parse_result.name] = parse_result
					p.PodInfos.Lock.Unlock()
					//p.Log.Infoln(parse_result.name, "has released the lock of PodInfos")
					p.Set_Create_Pod(parse_result)
				} else if ev.IsModify() {
					parse_result := parse_pod_info(*ev)
					p.Log.Infoln("receive restart pod message for", parse_result.name)
					p.PodInfos.Lock.Lock()
					//p.Log.Infoln(parse_result.name, "has got the lock of PodInfos")
					p.PodInfos.List[parse_result.name] = parse_result
					p.PodInfos.Lock.Unlock()
					//p.Log.Infoln(parse_result.name, "has released the lock of PodInfos")
					if err != nil {
						panic("restart count is invalid")
					}

					// to be convenient, assume that if value changes, mean restart of the pod
					p.Set_Recreate_Pod(parse_result)
				}
			}
		}
	}
}

func (p *Plugin) Set_Create_Pod(parse_result podinfo) error {
	p.assign_dir_to_pod(parse_result.name)
	//p.Log.Infoln("parse result =", parse_result)
	//p.Log.Infoln("pod =", pod)
	if p.Is_Local(parse_result.name) {
		p.Log.Infoln("pod", parse_result.name, "is local")
		pod := p.MocknetTopology.pods[parse_result.name]
		p.Log.Infoln(pod.name, "is waitting for the lock of PodConfig")
		p.PodConfig[pod.name].Lock()
		//p.Log.Infoln(pod.name, "has got the lock of PodConfig")
		//p.Log.Infoln("pod", pod.name, "is local, so start to config it")
		if p.Completed_Config(*pod) {
			p.Set_Connection(*pod)
		}
	}

	return nil
}

func (p *Plugin) Set_Recreate_Pod(parse_result podinfo) error {
	//p.assign_dir_to_pod(parse_result.name)
	pod := p.MocknetTopology.pods[parse_result.name]
	if p.Is_Local(pod.name) {
		p.Log.Warningln("received a restart signal fo pod", parse_result.name)
		p.Log.Infoln("start to reconfig pod", parse_result.name)
		p.PodConfig[pod.name].Lock()
		//p.Log.Infoln(parse_result.name, "has got the lock of PodConfig")
		p.Clear_Completed_Config(*pod)
		if p.Completed_Config(*pod) {
			p.Set_Connection(*pod)
		}
	}

	return nil
}

func (p *Plugin) Completed_Config(pod Pod) bool {
	// create memif socket
	id := p.get_socket_id(pod.name) + 1
	filedir := p.DirPrefix + pod.name
	p.Vpp.CreateSocket(uint32(id), filedir)

	// create memif interface
	pod.alloc_interface_id()

	var pod_memif_id uint32
	var vppresult vpp.ProcessResult
	var linuxresult linux.ProcessResult

	for _, vpp_id := range pod.infs {
		// host-side
		if result, global_id := p.Vpp.Create_Memif_Interface("master", uint32(vpp_id), uint32(id)); result == vpp.Success {
			p.Vpp.Set_interface_state_up(global_id)
			p.InterfaceIDName[global_id] = "memif" + strconv.Itoa(id) + "/" + strconv.Itoa(vpp_id)
			p.InterfaceNameID["memif"+strconv.Itoa(id)+"/"+strconv.Itoa(vpp_id)] = global_id
			p.IntToVppId.List[pod.name+"-"+strconv.Itoa(vpp_id)] = global_id
		} else {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}

		// pod-side
		inf_name := "memif" + strconv.Itoa(vpp_id)
		// int_id is the id in a socket used to identify peer(slave or master)
		vppresult, pod_memif_id = p.Vpp.Pod_Create_Memif_Interface(pod.name, "slave", inf_name, uint32(vpp_id))
		if vppresult == vpp.Success {
			p.PodIntToVppId.Lock.Lock()
			p.PodIntToVppId.List[pod.name+"-"+strconv.Itoa(vpp_id)] = pod_memif_id
			p.PodIntToVppId.Lock.Unlock()
			p.Vpp.Pod_Set_interface_state_up(pod.name, pod_memif_id)
		} else {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}
	}

	// write pod-vpp-tap to pod-vpp-memif route
	if strings.Contains(pod.name, "h") {
		p.PodInfos.Lock.Lock()
		containerid := p.PodInfos.List[pod.name].containerid
		podip := p.PodInfos.List[pod.name].podip
		p.PodInfos.Lock.Unlock()

		var tap_id uint32
		if vppresult, tap_id = p.Vpp.Pod_Create_Tap(pod.name); vppresult != vpp.Success {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}
		if vppresult = p.Vpp.Pod_Set_interface_state_up(pod.name, tap_id); vppresult != vpp.Success {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}
		if linuxresult = p.Linux.Pod_Add_Route(containerid, pod.name); linuxresult != linux.Success {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}
		if linuxresult = p.Linux.Pod_Set_Ip(containerid, pod.name, podip); linuxresult != linux.Success {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}

		// write pod-vpp to pod pod-linux route
		if vppresult = p.Vpp.Pod_Add_Route(pod.name, vpp.Route_Info{
			Dst: vpp.IpNet{
				Ip:   p.PodInfos.List[pod.name].podip,
				Mask: 32,
			},
			Dev:   "tap0",
			DevId: uint32(1),
		}); vppresult == vpp.Success {
			p.Vpp.Pod_Xconnect(pod.name, pod_memif_id, tap_id)
			p.Vpp.Pod_Xconnect(pod.name, tap_id, pod_memif_id)
		} else {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}

	} else if strings.Contains(pod.name, "s") {
		// bridge interfaces in switch pod together
		ints_id := make([]uint32, 0)
		for i := 0; i < len(p.MocknetTopology.pods[pod.name].infs); i++ {
			ints_id = append(ints_id, uint32(i+1))
		}
		if vppresult = p.Vpp.Pod_Bridge(pod.name, ints_id, 1); vppresult != vpp.Success {
			p.Log.Warningln("config for pod", pod.name, "has been interrupted")
			//p.Log.Infoln(pod.name, "is going to release the lock of PodConfig")
			p.PodConfig[pod.name].Unlock()
			//p.Log.Infoln(pod.name, "has released the lock of PodConfig")
			return false
		}
	} else {
		p.Log.Errorln("the judgement name is ", pod.name)
		panic("pod type judgement error!")
	}
	return true
}

func (p *Plugin) Clear_Completed_Config(pod Pod) error {
	/*id := p.get_socket_id(pod.name) + 1
	p.Vpp.DeleteSocket(uint32(id))*/

	p.Log.Infoln("deleting host-side memif interface for pod", pod.name)
	for _, vpp_id := range pod.infs {
		memif_int_id := p.IntToVppId.List[pod.name+"-"+strconv.Itoa(vpp_id)]
		p.Vpp.Delete_Memif_Interface(memif_int_id)
	}

	// no need to delete vxlan tunnel for now
	/*for int_name, vpp_id := range p.IntToVppId.List {
		if strings.Split(int_name, "-")[1] == pod.name {
			p.Vpp.Delete_Vxlan_Tunnel(vpp_id)
		}
	}*/

	return nil
}

type Vxlan_Count_Sync struct {
	lock  *sync.RWMutex
	count int
}

func (p *Plugin) Set_Connection(pod Pod) error {
	vxlan_total := 0
	for _, link := range pod.links {
		if p.Is_Local(link.pod1) && !p.Is_Local(link.pod2) {
			vxlan_total += 1
		}
	}

	vxlan_count := Vxlan_Count_Sync{
		lock:  &sync.RWMutex{},
		count: 0,
	}
	// host-side vpp config
	for _, link := range pod.links {
		go p.Connect(link, &vxlan_count, vxlan_total)
	}
	return nil
}

func (p *Plugin) Connect(link Link, vxlan_count *Vxlan_Count_Sync, vxlan_total int) error {
	var dst, src string
	var vni uint32
	var Is_Local1, Is_Local2 bool

	try_count := 0
	for {
		hit_flag := false
		// it means that dst pod's info hasn't been received, wait until it does
		p.PodInfos.Lock.Lock()
		p.NodeInfos.Lock.Lock()
		//p.Log.Infoln(link.pod1, "has got the lock of PodInfos")
		//p.Log.Infoln(link.pod1, "has got the lock of NodeInfos")

		pod2host := p.PodInfos.List[link.pod2].hostname
		//p.Log.Infoln(link.pod1, ":    pod2host =", pod2host)
		if pod2host != "" {
			dst = p.NodeInfos.List[pod2host].vtepip
			Is_Local1 = p.Is_Local(link.pod1)
			Is_Local2 = p.Is_Local(link.pod2)
			if try_count >= 1 {
				p.Log.Infoln(link.pod1+"'s", "information request of pod", link.pod2, "has been met")
			}
			hit_flag = true
		}

		p.PodInfos.Lock.Unlock()
		p.NodeInfos.Lock.Unlock()
		//p.Log.Infoln(link.pod1, "has released the lock of PodInfos")
		//p.Log.Infoln(link.pod1, "has released the lock of NodeInfos")

		if hit_flag {
			break
		}

		// only print once
		if try_count == 0 {
			p.Log.Infoln("pod", link.pod1, "is waitting for info of pod", link.pod2)
		}
		time.Sleep(3 * time.Second)
		try_count += 1
	}

	if Is_Local1 && Is_Local2 {
		// for both locals, xconnect them together
		go p.direct_xconnect(link)
	} else if Is_Local1 && !Is_Local2 {
		// for intf1 is local and intf2 not, create vxlan tunnel and xconnect intf1 with it
		src = p.Addresses.local_vtep_address
		vni = uint32(link.vni)
		vxlan_name := "vxlan" + strconv.Itoa(int(vni))
		var vxlan_id uint32

		if result, id := p.Vpp.Create_Vxlan_Tunnel(src, dst, vni, vni); result != vpp.TimesOver {
			if result == vpp.Success {
				vxlan_id = id
				p.IntToVppId.Lock.Lock()
				//p.Log.Infoln(link.pod1, "has got the lock of PodInfos")
				p.IntToVppId.List[vxlan_name] = vxlan_id
				p.IntToVppId.Lock.Unlock()
				//p.Log.Infoln(link.pod1, "has released the lock of PodInfos")
			} else if result == vpp.AlreadyExist {
				p.IntToVppId.Lock.Lock()
				//p.Log.Infoln(link.pod1, "has got the lock of PodInfos")
				vxlan_id = p.IntToVppId.List[vxlan_name]
				p.IntToVppId.Lock.Unlock()
				//p.Log.Infoln(link.pod1, "has released the lock of PodInfos")
			}
			p.IntToVppId.Lock.Lock()
			p.Vpp.XConnect(p.IntToVppId.List[link.pod1+"-"+link.pod1inf], vxlan_id)
			p.Vpp.XConnect(vxlan_id, p.IntToVppId.List[link.pod1+"-"+link.pod1inf])
			p.IntToVppId.Lock.Unlock()
		}
		vxlan_count.lock.Lock()
		vxlan_count.count += 1
		p.Log.Infoln("now vxlan_count =", vxlan_count.count, "total =", vxlan_total)
		if vxlan_count.count == vxlan_total {
			//p.Log.Infoln(link.pod1, "is going to release the lock of PodConfig")
			p.PodConfig[link.pod1].Unlock()
			//p.Log.Infoln(link.pod1, "has released the lock of PodConfig")
			p.Log.Infoln("---------------- config for", link.pod1, "finished ----------------")
		}
		vxlan_count.lock.Unlock()
	}

	return nil
}

func (p *Plugin) direct_xconnect(link Link) error {
	for {
		p.IntToVppId.Lock.Lock()
		intf2_id, ok := p.IntToVppId.List[link.pod2+"-"+link.pod2inf]
		p.IntToVppId.Lock.Unlock()
		// waitting for the pod2's memif interface to be created
		if ok {
			p.Vpp.XConnect(p.IntToVppId.List[link.pod1+"-"+link.pod1inf], intf2_id)
			break
		}
		time.Sleep(time.Second)
	}
	return nil
}

func parse_pod_info(event clientv3.Event) podinfo {
	v := string(event.Kv.Value)
	split_info := strings.Split(v, ",")
	restart_count, err := strconv.Atoi(strings.Split(split_info[5], ":")[1])
	if err != nil {
		panic("restart count is invalid")
	}
	//fmt.Println("name=", strings.Split(split_info[0], ":")[1], ", dockerid=", strings.Split(split_info[6], ":")[1])
	return podinfo{
		name:         strings.Split(split_info[0], ":")[1],
		namespace:    strings.Split(split_info[1], ":")[1],
		podip:        strings.Split(split_info[2], ":")[1],
		hostip:       strings.Split(split_info[3], ":")[1],
		hostname:     strings.Split(split_info[4], ":")[1],
		restartcount: restart_count,
		containerid:  strings.Split(split_info[6], ":")[1],
	}
}

func (p *Plugin) Is_Local(name string) bool {
	if p.PodInfos.List[name].hostip == p.Addresses.local_ip_address {
		return true
	} else {
		return false
	}
}

func (p *Plugin) assign_dir_to_pod(name string) error {
	p.DirAssign = append(p.DirAssign, name)
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

func (p *Plugin) clear_directory() error {
	dir, err := ioutil.ReadDir(p.DirPrefix)
	if err != nil {
		panic(err)
	}

	for _, pod := range dir {
		pod_dir, err := ioutil.ReadDir(p.DirPrefix + pod.Name())
		if err != nil {
			panic(err)
		}
		for _, d := range pod_dir {
			os.RemoveAll(path.Join([]string{p.DirPrefix, pod.Name(), d.Name()}...))
		}
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
	p.Log.Infoln("----------- configure for nodes -----------")
	resp, _ := p.MasterEtcdClient.Get(context.Background(), "/mocknet/nodeinfo", clientv3.WithPrefix())
	// to find local host infos
	for _, nodekv := range resp.Kvs {
		split_value := strings.Split(string(nodekv.Value), ",")
		name := strings.Split(split_value[0], ":")[1]
		nodeip := strings.Split(split_value[1], ":")[1]
		vtepip := strings.Split(split_value[2], ":")[1]
		p.NodeInfos.List[name] = Nodeinfo{
			name:   name,
			nodeip: nodeip,
			vtepip: vtepip,
		}
		if name == local_hostname {
			p.Addresses.local_ip_address = nodeip
			p.Addresses.local_vtep_address = vtepip
			p.Addresses.host_net_addr_24 = get_ip_net_24(nodeip)

			p.Vpp.Set_interface_state_up(1)
			p.Vpp.Set_Interface_Ip(1, vpp.IpNet{
				Ip:   vtepip,
				Mask: 24,
			})
		} else if name == "master" {
			p.Addresses.master_ip_address = nodeip
		}
	}
	for _, nodekv := range resp.Kvs {
		split_value := strings.Split(string(nodekv.Value), ",")
		name := strings.Split(split_value[0], ":")[1]
		vtepip := strings.Split(split_value[2], ":")[1]
		if name != local_hostname && name != "master" {
			p.Vpp.Add_Route(vpp.Route_Info{
				Dst: vpp.IpNet{
					Ip:   vtepip,
					Mask: 32,
				},
				Dev:   "temp_vetp_name",
				DevId: 1,
			})
		}
	}
	p.Log.Infoln("----------- configure for nodes finished -----------")

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

func get_ip_net_24(ip string) string {
	split_ip := strings.Split(ip, ".")
	return split_ip[0] + "." + split_ip[1] + "." + split_ip[2] + ".0"
}
