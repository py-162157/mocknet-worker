package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	config := clientv3.Config{
		Endpoints:   []string{"0.0.0.0:32379"},
		DialTimeout: 10 * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("successfully create new etcd client!")
	}
	defer client.Close()

	kv := clientv3.NewKV(client)

	ctx, _ := context.WithTimeout(context.TODO(), 5*time.Second)

	getResp, err := kv.Get(ctx, "", clientv3.WithPrefix())
	//_, err = kv.Put(ctx, "/vnf-agent/mocknet-pod-h1/config/vpp/v2/interfaces/memif3", "{\"name\":\"memif3\",\"type\":\"MEMIF\",\"enabled\":true,\"memif\":{\"id\":1,\"socket_filename\":\"/run/vpp/memif.sock\"}}")
	if err != nil {
		panic(err)
	}
	for _, kv := range getResp.Kvs {
		fmt.Println("key:", string(kv.Key))
		fmt.Println("value:", string(kv.Value))
	}

	//kvs := clientv3.NewKV(client)
	/*getresp, err := kvs.Get(context.Background(), "", clientv3.WithPrefix())
	if err != nil {
		panic("error when get all key-values")
	}
	for _, key := range getresp.Kvs {
		_, err = kvs.Delete(context.Background(), string(key.Key))
		if err != nil {
			panic("error when clear all key-value")
		}
	}
	fmt.Println("clear all info finished")*/

	/*getResp, err := kvs.Get(context.Background(), "", clientv3.WithPrefix())

	if err != nil {
		panic("error when clear all key-value")
	}
	for _, kv := range getResp.Kvs {
		fmt.Println("key:", string(kv.Key))
		fmt.Println("value:", string(kv.Value))
	}*/
	/*_, err = kvs.Put(context.Background(), "/vnf-agent/mocknet-pod-h1/config/vpp/v2/interfaces/memif3", "{\"name\":\"memif3\",\"type\":\"MEMIF\",\"enabled\":true,\"memif\":{\"id\":1,\"socket_filename\":\"/run/vpp/memif.sock\"}}")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("successfully put a value to etcd")
	}*/
}
