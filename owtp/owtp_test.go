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
	"testing"
	"time"
)

var (
	hostURL = "127.0.0.1:9432"
	mqURL = "192.168.30.160:5672"
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

func hello(ctx *Context) {
	ctx.Resp = Response{
		Status: 0,
		Msg:    "success",
		Result: map[string]interface{}{
			"hello": "hello world",
		},
	}
}

func createHost() *OWTPNode {

	//主机
	host := RandomOWTPNode()
	host.Listen(":9432")

	host.HandleFunc("hello", hello)

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


func TestOtherMQConnectNode(t *testing.T){
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	err := nodeA.Connect(mqURL, "dasda",MQ)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}
	nodeA.Call("dasda", "hello", nil, true, func(resp Response) {
		hello := resp.JsonData().Get("hello").String()
		fmt.Printf("nodeA call hello, result: %s\n", hello)
	})
}

func TestMQConnectNode(t *testing.T) {

	host := createHost()

	time.Sleep(5 * time.Second)

	//客户端
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	err := nodeA.Connect(mqURL, "dasda",MQ)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeB := RandomOWTPNode()
	nodeB.HandleFunc("getInfo", getInfo)
	err = nodeB.Connect(mqURL, host.NodeID(),MQ)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeC := RandomOWTPNode()
	nodeC.HandleFunc("getWallegetInfotInfo", getInfo)
	err = nodeC.Connect(mqURL, host.NodeID(),MQ)
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

	host := createHost()
	//
	//time.Sleep(5 * time.Second)

	//客户端
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	err := nodeA.Connect(hostURL, host.NodeID(),Websocket)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeB := RandomOWTPNode()
	nodeB.HandleFunc("getInfo", getInfo)
	err = nodeB.Connect(hostURL, host.NodeID(),Websocket)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	nodeC := RandomOWTPNode()
	nodeC.HandleFunc("getWallegetInfotInfo", getInfo)
	err = nodeC.Connect(hostURL, host.NodeID(),Websocket)
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

	nodeA.Call(nodeA.NodeID(), "getInfo", nil, true, func(resp Response) {
		result := resp.JsonData()
		fmt.Printf("host call nodeA, result: %s\n", result)
	})

	t.Logf("node close \n")

	time.Sleep(3 * time.Second)

	nodeA.ClosePeer(host.NodeID())

	time.Sleep(5 * time.Second)

	host.Close()

	t.Logf("stop running \n")

	time.Sleep(5 * time.Second)

	t.Logf("end testing \n")

}

func TestConcurrentConnect(t *testing.T) {

	host := createHost()

	time.Sleep(5 * time.Second)

	for i := 0; i < 100; i++ {
		go func(h *OWTPNode) {

			//客户端
			node := createClient()
			err := node.Connect(hostURL, host.NodeID(),Websocket)
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
