package etcd

import (
	"context"
	"strconv"
	"strings"
	"time"

	"mocknet/plugins/vpp"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"
)

type Plugin struct {
	Deps

	PluginName   string
	K8sNamespace string
	EtcdClient   *clientv3.Client
}

type Deps struct {
	PluginInitFinished bool
	LocalHostName      string

	Log logging.PluginLogger
}

const (
	TIMEOUT = time.Duration(10) * time.Second // 超时
)

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	if client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"0.0.0.0:32379"},
		DialTimeout: 5 * time.Second,
	}); err != nil {
		p.Log.Errorln(err)
		panic(err)
	} else {
		p.EtcdClient = client
		p.Log.Infoln("successfully connected to master etcd!")
	}

	p.PluginInitFinished = true

	return nil
}

func (p *Plugin) String() string {
	return "etcd"
}

func (p *Plugin) Close() error {
	return nil
}

func (p *Plugin) Inform_finished(event string) error {
	kv := clientv3.NewKV(p.EtcdClient)
	//p.Log.Infoln("local host name =", p.LocalHostName)
	_, err := kv.Put(context.Background(), "/mocknet/"+event+"/"+p.LocalHostName, "done")
	//p.Log.Infoln("key =", "/mocknet/"+event+"/"+p.LocalHostName)
	if err != nil {
		panic(err)
	}
	p.Log.Infoln("informed the master that finished event", event)
	return nil
}

// wait for all workers finishing some work
func (p *Plugin) Wait_for_order(order string) error {
	p.Log.Infoln("waiting for order", order)
	kvs := clientv3.NewKV(p.EtcdClient)
	for {
		order_resp, err := kvs.Get(context.Background(), "/mocknet/order/"+order)
		if err != nil {
			panic(err)
		}
		if len(order_resp.Kvs) == 1 {
			break
		}
	}
	return nil
}

/*func (p *Plugin) Watch(event string, withprefix bool) error {
	var getResp *clientv3.GetResponse
	var err error
	if withprefix {
		getResp, err = p.EtcdClient.Get(context.Background(), event, clientv3.WithPrefix())
	} else {
		getResp, err = p.EtcdClient.Get(context.Background(), event)
	}

	if err != nil {
		return err
	}



	return nil
}*/

func (p *Plugin) Put(key string, value string) error {
	kv := clientv3.NewKV(p.EtcdClient)
	kv.Put(context.Background(), key, value)
	return nil
}

func (p *Plugin) Get(key string, With_Prefix bool) *clientv3.GetResponse {
	kv := clientv3.NewKV(p.EtcdClient)
	var resp *clientv3.GetResponse
	var err error
	if With_Prefix {
		resp, err = kv.Get(context.Background(), key, clientv3.WithPrefix())
	} else {
		resp, err = kv.Get(context.Background(), key)
	}
	if err != nil {
		panic(err)
	}
	return resp
}

func (p *Plugin) Upload_Route(routes_map map[string][]vpp.Route_Info) {
	for src, routes := range routes_map {
		for _, route := range routes {
			key := "/mocknet/routes/" + src + "-" + route.DstName
			value := "Src:" + src + ","
			value += "DstIp:" + route.Dst.Ip + ","
			value += "DstMask:" + strconv.Itoa(int(route.Dst.Mask)) + ","
			value += "GwIp:" + route.Gw.Ip + ","
			value += "GwMask:" + strconv.Itoa(int(route.Gw.Mask)) + ","
			value += "Port:" + strconv.Itoa(int(route.Port)) + ","
			value += "Dev:" + route.Dev + ","
			value += "Devid:" + strconv.Itoa(int(route.DevId)) + ","
			value += "DstName:" + route.DstName + ","
			value += "NextHopName:" + route.Next_hop_name + ","
			value += "LocalHostName:" + p.LocalHostName

			p.Put(key, value)
		}
	}

	p.Inform_finished("RouteUpload")
}

func (p *Plugin) Upload_Speed(speed_map map[string]float64) {
	for pod, speed := range speed_map {
		key := "/mocknet/speed/" + pod + "/" + p.LocalHostName
		value := "pod:" + pod + "," + "speed:" + strconv.FormatFloat(speed, 'E', -1, 64)
		p.Put(key, value)
	}
}

func (p *Plugin) Parse_Speed() map[string]float64 {
	speed_map := make(map[string]float64)
	resp := p.Get("/mocknet/speed/", true)
	for _, kv := range resp.Kvs {
		value := kv.Value
		split_value := strings.Split(string(value), ",")
		podname := strings.Split(split_value[0], ":")[1]
		speed_string := strings.Split(split_value[1], ":")[1]
		speed, err := strconv.ParseFloat(speed_string, 64)
		if err != nil {
			panic(err)
		}
		if _, ok := speed_map[podname]; !ok {
			speed_map[podname] = speed
		} else {
			speed_map[podname] += speed
		}
	}

	return speed_map
}

func (p *Plugin) Parse_Routes() map[string][]vpp.Route_Info {
	routes_map := make(map[string][]vpp.Route_Info)
	resp := p.Get("/mocknet/routes/", true)
	for _, kv := range resp.Kvs {
		//fmt.Println(string(kv.Value))
		value := kv.Value
		split_value := strings.Split(string(value), ",")
		src := strings.Split(split_value[0], ":")[1]
		dstip := strings.Split(split_value[1], ":")[1]
		dstmask_string := strings.Split(split_value[2], ":")[1]
		dstmask, _ := strconv.Atoi(dstmask_string)
		gwip := strings.Split(split_value[3], ":")[1]
		gwmask_string := strings.Split(split_value[4], ":")[1]
		gwmask, _ := strconv.Atoi(gwmask_string)
		port_string := strings.Split(split_value[5], ":")[1]
		port, _ := strconv.Atoi(port_string)
		dev := strings.Split(split_value[6], ":")[1]
		devid_string := strings.Split(split_value[7], ":")[1]
		devid, _ := strconv.Atoi(devid_string)
		dstname := strings.Split(split_value[8], ":")[1]
		nexthop := strings.Split(split_value[9], ":")[1]
		hostname := strings.Split(split_value[10], ":")[1]
		var local bool
		if hostname == p.LocalHostName {
			local = true
		} else {
			local = false
		}
		if _, ok := routes_map[src]; !ok {
			routes_map[src] = make([]vpp.Route_Info, 0)
		}
		routes_map[src] = append(routes_map[src], vpp.Route_Info{
			Dst: vpp.IpNet{
				Ip:   dstip,
				Mask: uint(dstmask),
			},
			Gw: vpp.IpNet{
				Ip:   gwip,
				Mask: uint(gwmask),
			},
			Port:          uint32(port),
			Dev:           dev,
			DevId:         uint32(devid),
			DstName:       dstname,
			Next_hop_name: nexthop,
			Local:         local,
		})
	}

	return routes_map
}
