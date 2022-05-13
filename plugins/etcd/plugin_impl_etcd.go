package etcd

import (
	"context"
	"time"

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
	p.Log.Infoln("put a data to etcd")
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
