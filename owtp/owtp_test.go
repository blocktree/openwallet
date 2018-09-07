/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package owtp

import (
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"testing"
	"time"
)

var (
	hostURL     = "127.0.0.1:9432"
	transferURL = "127.0.0.1:9431"
	mqURL       = "192.168.30.160:5672"
	hostNodeID  = "AR7ZxNbPJeQS7iqvzqEPCq5koTJQvnggNhWR7SSD6LCS"
	hostkey     = "3JYgidyyjhqbTsGzduK9rkM2JaYht4gzRWyhUdCAH1vf"
)

func init() {
	Debug = true
}

func getInfo(ctx *Context) {
	ctx.Resp = Response{
		Status: 0,
		Msg:    "success",
		Result: map[string]interface{}{
			"symbols": []interface{}{
				map[string]interface{}{
					"coin":     "btc",
					"walletID": "mywll",
				},
				map[string]interface{}{
					"coin":     "btm",
					"walletID": "mkk",
				},
			},
		},
	}
}

func (node *OWTPNode) hello(ctx *Context) {

	log.Info("Call host Hello")

	ctx.Resp = Response{
		Status: 0,
		Msg:    "success",
		Result: map[string]interface{}{
			"hello": "hello world",
		},
	}

}

func (node *OWTPNode) transferHello(ctx *Context) {
	//ctx.Resp = Response{
	//	Status: 0,
	//	Msg:    "success",
	//	Result: map[string]interface{}{
	//		"hello": "hello world",
	//	},
	//}

	log.Info("Call transfer Hello")

	//转发主机
	node.Call(hostNodeID, ctx.Method, ctx.Params(), true, func(resp Response) {
		ctx.Resp = resp
	})

}

func createHost() *OWTPNode {

	cert, err := NewCertificate(hostkey, "")
	if err != nil {
		return nil
	}

	//主机
	host := NewOWTPNode(cert, 0, 0)
	host.Listen(":9432")

	host.HandleFunc("hello", host.hello)

	return host
}

func createClient() *OWTPNode {
	//客户端
	node := RandomOWTPNode()
	node.HandleFunc("getInfo", getInfo)
	return node
}

func TestGenerateRangeNum(t *testing.T) {
	for i := 0; i < 1000; i++ {
		num := GenerateRangeNum(0, 1023)
		t.Logf("num [%d]= %d", i, num)
	}
}

func TestOtherMQConnectNode(t *testing.T) {
	config := make(map[string]string)
	config["address"] = mqURL
	config["connectType"] = MQ
	config["exchange"] = "DEFAULT_EXCHANGE"
	config["queueName"] = "DEFAULT_QUEUE"
	config["receiveQueueName"] = "DEFAULT_QUEUE"
	config["account"] = "admin"
	config["password"] = "admin"
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	err := nodeA.Connect("dasda", config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}
	time.Sleep(3 * time.Second)
	nodeA.Call("dasda", "hello", nil, true, func(resp Response) {
		hello := resp.JsonData().Get("hello").String()
		fmt.Printf("nodeA call hello, result: %s\n", hello)
	})
}

func TestMQConnectNode(t *testing.T) {

	host := createHost()

	time.Sleep(5 * time.Second)

	config := make(map[string]string)
	config["address"] = hostURL
	config["connectType"] = MQ

	//客户端
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	err := nodeA.Connect("dasda", config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeB := RandomOWTPNode()
	nodeB.HandleFunc("getInfo", getInfo)
	err = nodeB.Connect("dasda", config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeC := RandomOWTPNode()
	nodeC.HandleFunc("getWallegetInfotInfo", getInfo)
	err = nodeC.Connect("dasda", config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(3 * time.Second)

	//调用方法
	nodeA.Call(host.NodeID(), "hello", nil, true, func(resp Response) {
		hello := resp.JsonData().Get("hello").String()
		fmt.Printf("nodeA call hello, result: %s\n", hello)
	})

	time.Sleep(3 * time.Second)

	//host.Call(nodeA.NodeID(), "getInfo", nil, true, func(resp Response) {
	//	result := resp.JsonData()
	//	fmt.Printf("host call nodeA, result: %s\n", result)
	//})

	t.Logf("node close \n")

	time.Sleep(3 * time.Second)

	nodeA.ClosePeer(host.NodeID())

	time.Sleep(5 * time.Second)

	host.Close()

	t.Logf("stop running \n")

	time.Sleep(5 * time.Second)

	t.Logf("end testing \n")

}

func TestConnectNode(t *testing.T) {

	//A,B连接host，host连接C
	//A,B请求经host转发给C，C处理业务返回结果

	//主机
	host := createHost()
	//host.Listen(":9432")
	//host.HandleFunc("hello", host.hello)

	//中转
	transfer := RandomOWTPNode()
	transfer.Listen(":9431")
	transfer.HandleFunc("hello", transfer.transferHello)

	//客户端
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	nodeB := RandomOWTPNode()
	nodeB.HandleFunc("getInfo", getInfo)

	time.Sleep(5 * time.Second)

	transferConfig := make(map[string]string)
	transferConfig["address"] = hostURL
	transferConfig["connectType"] = Websocket

	//中转连接主机
	err := transfer.Connect(host.NodeID(), transferConfig)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	config := make(map[string]string)
	config["address"] = transferURL
	config["connectType"] = Websocket

	//A连接中转
	err = nodeA.Connect(transfer.NodeID(), config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	//B连接中转
	err = nodeB.Connect(transfer.NodeID(), config)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(3 * time.Second)

	//调用方法
	nodeA.Call(transfer.NodeID(), "hello", nil, true, func(resp Response) {
		hello := resp.JsonData().Get("hello").String()
		fmt.Printf("nodeA call transfer, result: %s\n", hello)
	})

	time.Sleep(3 * time.Second)

	nodeB.Call(transfer.NodeID(), "hello", nil, true, func(resp Response) {
		hello := resp.JsonData().Get("hello").String()
		fmt.Printf("nodeB call transfer, result: %s\n", hello)
	})

	t.Logf("node close \n")

	time.Sleep(2 * time.Second)

	nodeA.ClosePeer(transfer.NodeID())

	time.Sleep(4 * time.Second)

	host.Close()

	t.Logf("stop running \n")

	time.Sleep(5 * time.Second)

	t.Logf("end testing \n")

}

func TestConcurrentConnect(t *testing.T) {

	host := createHost()

	time.Sleep(5 * time.Second)

	config := make(map[string]string)
	config["address"] = hostURL
	config["connectType"] = Websocket

	for i := 0; i < 100; i++ {
		go func(h *OWTPNode) {

			//客户端
			node := createClient()
			err := node.Connect(host.NodeID(), config)
			if err != nil {
				t.Errorf("Connect failed unexpected error: %v", err)
				return
			}

			time.Sleep(3 * time.Second)

			node.ClosePeer(host.NodeID())

		}(host)
	}

	time.Sleep(30 * time.Second)

	host.Close()
}
