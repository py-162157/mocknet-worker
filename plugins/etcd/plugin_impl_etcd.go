package etcd

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	//"mocknet/plugins/server/rpctest"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.ligato.io/cn-infra/v2/logging"
)

var (
	TIMEOUT             = time.Duration(3) * time.Second // 超时
	MASTER_ETCD_ADDRESS = []string{"192.168.122.100:22379"}
	LOCAL_ETCD_ADDRESS  = []string{"192.168.122.101:22379"}
)

// Listener 对外通知
/*type Listener interface {
	Set([]byte, []byte)
	Create([]byte, []byte)
	Modify([]byte, []byte)
	Delete([]byte)
}*/

type EtcdWatcher struct {
	cli *clientv3.Client // etcd client
	wg  sync.WaitGroup
	//listener     Listener
	mu           sync.Mutex
	closeHandler map[string]func()
	log          logging.PluginLogger
	// local etcd client
	localclient *clientv3.Client
	// on or off
	watch_switch string
}

type Plugin struct {
	Deps

	PluginName       string
	K8sNamespace     string
	LocalEtcdClient  *clientv3.Client
	MasterEtcdClient *clientv3.Client
	MasterWatcher    *EtcdWatcher
}

type Deps struct {
	Log logging.PluginLogger
}

func (p *Plugin) Init() error {
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	go func() {
		// prepare etcd binary file, make sure the relative position of etcd.sh
		cmd := exec.Command("bash", "./../scripts/etcd.sh") // 以当前命令行路径为准，当前路径为contiv-agent,目前不需要
		output, err := cmd.StdoutPipe()
		if err != nil {
			p.Log.Errorln(err)
			panic(err)
		}

		if err := cmd.Start(); err != nil {
			p.Log.Errorln(err)
			panic(err)
		} else {
			p.Log.Infoln("successfully setup and start etcd!")
		}

		reader := bufio.NewReader(output)

		var contentArray = make([]string, 0, 5)
		var index int
		contentArray = contentArray[0:0]

		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			p.Log.Infoln(line)
			index++
			contentArray = append(contentArray, line)
		}
	}()

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

	// create master-etcd watcher
	if watcher, err := p.NewEtcdWatcher(MASTER_ETCD_ADDRESS); err != nil {
		p.Log.Errorln(err)
		panic(err)
	} else {
		p.MasterWatcher = watcher
		p.Log.Infoln("successfully create new master-etcd watcher!")
	}

	p.MasterWatcher.watch_switch = "on"

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

	p.sync_etcd_data()

	// begin to sync master etcd data to local's
	go p.MasterWatcher.AddWatch("", true)

	return nil
}

func (p *Plugin) String() string {
	return "etcd"
}

// mininet make a same default network (pod's name and interface's) if given
// same topology, which make the key-value in etcd unchanged. To avoid it, clear
// topology key in etcd when plugin closed.
func (p *Plugin) Close() error {
	p.MasterWatcher.watch_switch = "off"
	p.MasterWatcher.wg.Done()
	p.Log.Infoln("master watcher has been stopped")
	kvs := clientv3.NewKV(p.LocalEtcdClient)
	_, err := kvs.Delete(context.Background(), "/", clientv3.WithPrefix())
	if err != nil {
		panic("error when clear all value")
	}
	p.Log.Infoln("clear all key-value finished")
	return nil
}

func (p *Plugin) clear_topology() error {
	kvs := clientv3.NewKV(p.MasterEtcdClient)
	getresp, err := kvs.Get(context.Background(), "/mocknet/link/", clientv3.WithPrefix())
	if err != nil {
		panic("error when get topology key-values")
	}
	for _, topokey := range getresp.Kvs {
		_, err = kvs.Delete(context.Background(), string(topokey.Key))
		if err != nil {
			panic("error when clear topology")
		}
	}
	p.Log.Infoln("clear topology info finished")
	return nil
}

func (p *Plugin) clear_pod_info() error {
	kvs := clientv3.NewKV(p.MasterEtcdClient)
	getresp, err := kvs.Get(context.Background(), "/mocknet/pods/", clientv3.WithPrefix())
	if err != nil {
		panic("error when get pod info key-values")
	}
	for _, topokey := range getresp.Kvs {
		_, err = kvs.Delete(context.Background(), string(topokey.Key))
		if err != nil {
			panic("error when clear pod info")
		}
	}
	p.Log.Infoln("clear pod info finished")
	return nil
}

func (p *Plugin) reset_transport_done() error {
	kvs := clientv3.NewKV(p.MasterEtcdClient)
	_, err := kvs.Put(context.Background(), "/mocknet/topo/ready", "0")
	if err != nil {
		panic("error when reset transport-done flag")
	}
	p.Log.Infoln("transport-done flag reset finished")
	return nil
}

// completely sync master-etcd data to local etcd
func (p *Plugin) sync_etcd_data() error {
	masterkv := clientv3.NewKV(p.MasterEtcdClient)
	ctx := context.Background()
	getResp, err := masterkv.Get(ctx, "/", clientv3.WithPrefix())

	if err != nil {
		p.Log.Errorln(err)
	} else {
		p.Log.Infoln("successfully get all keys from master-etcd!")
	}

	localkv := clientv3.NewKV(p.LocalEtcdClient)
	txn := localkv.Txn(ctx)
	ops := []clientv3.Op{}

	for _, KV := range getResp.Kvs {
		ops = append(ops, clientv3.OpPut(string(KV.Key), string(KV.Value)))
	}

	txn.Then(ops...)

	_, err = txn.Commit()
	if err != nil {
		p.Log.Errorln(err)
		panic(err)
	} else {
		p.Log.Infoln("successfully commit data to local etcd")
	}

	return nil
}

func (p *Plugin) NewEtcdWatcher(servers []string) (*EtcdWatcher, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   servers,
		DialTimeout: TIMEOUT,
	})
	if err != nil {
		return nil, err
	}

	localclient, err := clientv3.New(clientv3.Config{
		Endpoints:   LOCAL_ETCD_ADDRESS,
		DialTimeout: TIMEOUT,
	})

	ew := &EtcdWatcher{
		cli:          cli,
		closeHandler: make(map[string]func()),
		log:          p.Log,
		localclient:  localclient,
	}

	return ew, nil
}

func (w *EtcdWatcher) AddWatch(key string, prefix bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.closeHandler[key]; ok {
		return false
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.closeHandler[key] = cancel

	w.wg.Add(1)
	go w.watch(ctx, key, prefix)

	return true
}

// RemoveWatch 删除监视
func (w *EtcdWatcher) RemoveWatch(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	cancel, ok := w.closeHandler[key]
	if !ok {
		return false
	}
	cancel()
	delete(w.closeHandler, key)

	return true
}

func (w *EtcdWatcher) CloseWatcher(wait bool) {
	w.ClearWatch()

	if wait {
		w.wg.Wait()
	}

	w.cli.Close()
	w.cli = nil
}

func (w *EtcdWatcher) ClearWatch() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for k := range w.closeHandler {
		w.closeHandler[k]()
	}
	w.closeHandler = make(map[string]func())
}

func (w *EtcdWatcher) watch(ctx context.Context, key string, prefix bool) error {
	defer w.wg.Done()

	ctx1, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	var getResp *clientv3.GetResponse
	var err error
	if prefix {
		getResp, err = w.cli.Get(ctx1, key, clientv3.WithPrefix())
	} else {
		getResp, err = w.cli.Get(ctx1, key)
	}
	if err != nil {
		return err
	}

	localkvc := clientv3.NewKV(w.localclient)

	var watchChan clientv3.WatchChan
	if prefix {
		watchChan = w.cli.Watch(context.Background(), key, clientv3.WithPrefix(), clientv3.WithRev(getResp.Header.Revision+1))
	} else {
		watchChan = w.cli.Watch(context.Background(), key, clientv3.WithRev(getResp.Header.Revision+1))
	}
	for {
		if w.watch_switch == "off" {
			break
		}
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
					ctx, cancelFunc := context.WithTimeout(context.TODO(), time.Duration(2)*time.Second)
					_, err := localkvc.Put(ctx, string(ev.Kv.Key), string(ev.Kv.Value))
					cancelFunc()
					if err != nil {
						w.log.Errorln(err)
					} else {
						//w.log.Infoln("put a new contiv-etcd key-value pair to local etcd")
						//w.log.Infoln("key: ", string(ev.Kv.Key), ", value: ", string(ev.Kv.Value))
						//w.log.Infoln("")
					}
					//listener.Create(ev.Kv.Key, ev.Kv.Value)
				} else if ev.IsModify() {
					ctx, cancelFunc := context.WithTimeout(context.TODO(), time.Duration(2)*time.Second)
					_, err := localkvc.Put(ctx, string(ev.Kv.Key), string(ev.Kv.Value))
					cancelFunc()
					if err != nil {
						w.log.Errorln(err)
					} else {
						//if string(ev.Kv.Key) != "/vnf-agent/contiv-ksr/status/gauges" {
						//	w.log.Infoln("modified a contiv-etcd key-value pair to local etcd")
						//	w.log.Infoln("key: ", string(ev.Kv.Key), ", value: ", string(ev.Kv.Value))
						//	w.log.Infoln("")
						//}
					}
					//listener.Modify(ev.Kv.Key, ev.Kv.Value)
				} else if ev.Type == 1 { // 1 present DELETE
					ctx, cancelFunc := context.WithTimeout(context.TODO(), time.Duration(2)*time.Second)
					_, err := localkvc.Delete(ctx, string(ev.Kv.Key))
					cancelFunc()
					if err != nil {
						w.log.Errorln(err)
					} else {
						//w.log.Infoln("delete a contiv-etcd key-value pair to local etcd")
						//w.log.Infoln("")
					}
					//listener.Delete(ev.Kv.Key)
				} else {
				}
			}
		}
	}
	w.log.Infoln("stop watching master etcd")
	return nil
}
