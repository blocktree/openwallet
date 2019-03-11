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
	Debug = false
}

func getInfo(ctx *Context) {
	log.Info("param:", ctx.Params())
	ctx.SetSession("username", "kiiik")

	ctx.Resp = Response{
		Status: StatusSuccess,
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

func (node *OWTPNode) subscribe(ctx *Context) {
	ctx.Resp = Response{
		Status: 0,
		Msg:    "success",
		Result: map[string]interface{}{
			"subscribe": "subscribe",
		},
	}
	log.Info("Call transfer subscribe")
}

func createHost() *OWTPNode {

	cert, err := NewCertificate(hostkey, "")
	if err != nil {
		return nil
	}

	//主机
	host := NewOWTPNode(cert, 0, 0)

	config := ConnectConfig{
		Address:     "127.0.0.1:9432",
		ConnectType: Websocket,
	}

	host.Listen(config)
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
	config := ConnectConfig{}
	config.Address = mqURL
	config.ConnectType = MQ
	config.Exchange = "DEFAULT_EXCHANGE"
	config.WriteQueueName = "OW_RPC_GO"
	config.ReadQueueName = "OW_RPC_JAVA"
	config.Account = "admin"
	config.Password = "admin"
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

	config := ConnectConfig{}
	config.Address = hostURL
	config.ConnectType = MQ

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

func TestHttp(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)
	transfer := RandomOWTPNode()
	config := ConnectConfig{}
	config.Address = "127.0.0.1:8422"
	config.ConnectType = HTTP
	transfer.HandleFunc("getInfo", getInfo)
	transfer.Listen(config)

	//httpTestClient := NewHTTPT("http://"+config["address"], "", true)
	//httpTestClient.Call("getInfo", nil)
	<-endRunning
}

func TestMQtNode(t *testing.T) {
	transfer := RandomOWTPNode()
	config := ConnectConfig{}
	config.Address = ":94321"
	config.ConnectType = MQ
	transfer.Listen(config)
	transfer.HandleFunc("createWallet", transfer.subscribe)
	transferConfig := ConnectConfig{}
	transferConfig.Address = mqURL
	transferConfig.ConnectType = MQ
	transferConfig.Exchange = "DEFAULT_EXCHANGE"
	transferConfig.WriteQueueName = "OW_RPC_GO"
	transferConfig.ReadQueueName = "OW_RPC_JAVA"
	transferConfig.Account = "admin"
	transferConfig.Password = "admin"

	//中转连接主机
	err := transfer.Connect("dasda", transferConfig)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}
	transfer.Run()
	time.Sleep(100 * time.Second)
}

func TestConnectNode(t *testing.T) {

	//A,B连接transfer，transfer连接host
	//A,B请求经transfer转发给host，host处理业务返回结果

	cert, err := NewCertificate(hostkey, "")
	if err != nil {
		return
	}

	//主机
	host := NewOWTPNode(cert, 0, 0)
	config := ConnectConfig{}
	config.Address = ":9432"
	config.ConnectType = Websocket
	host.Listen(config)
	host.HandleFunc("hello", host.hello)

	//中转
	transfer := RandomOWTPNode()
	config1 := ConnectConfig{}
	config1.Address = ":9431"
	config1.ConnectType = Websocket
	transfer.Listen(config1)
	transfer.HandleFunc("hello", transfer.transferHello)

	//客户端
	nodeA := RandomOWTPNode()
	nodeA.HandleFunc("getInfo", getInfo)
	nodeB := RandomOWTPNode()
	nodeB.HandleFunc("getInfo", getInfo)

	time.Sleep(5 * time.Second)

	//transferConfig := make(map[string]string)
	//transferConfig["address"] = mqURL
	//transferConfig["connectType"] = MQ

	transferConfig := ConnectConfig{}
	transferConfig.Address = mqURL
	transferConfig.ConnectType = MQ
	transferConfig.Exchange = "DEFAULT_EXCHANGE"
	transferConfig.WriteQueueName = "DEFAULT_QUEUE"
	transferConfig.ReadQueueName = "DEFAULT_QUEUE"
	transferConfig.Account = "admin"
	transferConfig.Password = "admin"

	//中转连接主机
	err = transfer.Connect("dasda", transferConfig)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	config2 := ConnectConfig{}
	config2.Address = transferURL
	config2.ConnectType = Websocket

	//A连接中转
	err = nodeA.Connect(transfer.NodeID(), config2)
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

	config := ConnectConfig{}
	config.Address = hostURL
	config.ConnectType = Websocket

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

// 测试订阅服务
func TestSubscribeConnectNode(t *testing.T) {

	var (
		subscribeNid string
		connectTypes = []string{HTTP, Websocket}
	)

	for _, connectType := range connectTypes {

		log.Infof("%s connectType test", connectType)

		//服务端
		hostConnectConfig := ConnectConfig{
			Address:     "127.0.0.1:9432",
			ConnectType: connectType,
		}
		host := RandomOWTPNode()
		log.Infof("host NodeID: %s", host.NodeID())
		//开启监听，用于处理服务
		host.Listen(hostConnectConfig)
		//处理订阅业务
		host.HandleFunc("subscribe", func(ctx *Context) {
			subscribeNid = ctx.Params().Get("nodeID").String()
			addr := ctx.Params().Get("address").String()
			connectType := ctx.Params().Get("connectType").String()
			log.Infof("host connecting client: %s", addr)
			errConnect := host.Connect(subscribeNid, ConnectConfig{
				Address:     addr,
				ConnectType: connectType,
			})
			if errConnect != nil {
				t.Errorf("host connect client failed unexpected error: %v", errConnect)
				return
			}

			ctx.Response(nil, StatusSuccess, "subscribe success")

		})

		//客户端
		clientConnectConfig := ConnectConfig{
			Address:     "127.0.0.1:9433",
			ConnectType: connectType,
		}
		client := RandomOWTPNode()
		log.Infof("client NodeID: %s", client.NodeID())
		//开启监听，用于处理回调服务
		client.Listen(clientConnectConfig)
		//订阅数据处理
		client.HandleFunc("handlesSubscription", func(ctx *Context) {
			username := ctx.Params().Get("username").String()
			log.Info("username:", username)
			ctx.Response(nil, StatusSuccess, "handlesSubscription success")
		})

		//连接主机
		log.Info("client connecting host")
		err := client.Connect(host.NodeID(), hostConnectConfig)
		if err != nil {
			t.Errorf("client connect failed unexpected error: %v", err)
			return
		}

		log.Info("client calling host [subscribe]")
		//调用订阅方法
		err = client.Call(
			host.NodeID(),
			"subscribe",
			map[string]interface{}{
				"nodeID":      "clientsub",
				"address":     "127.0.0.1:9433",
				"connectType": connectType,
			},
			true, func(resp Response) {
				log.Info("resp:", resp)
			})
		if err != nil {
			t.Errorf("client call subscribe failed unexpected error: %v", err)
			return
		}

		//主动回调订阅数据
		err = host.Call(subscribeNid,
			"handlesSubscription",
			map[string]interface{}{
				"username": "john",
			},
			true, func(resp Response) {
				log.Info("resp:", resp)
			})
		if err != nil {
			t.Errorf("host call handlesSubscription failed unexpected error: %v", err)
			return
		}

		//主动回调订阅数据
		err = host.Call(subscribeNid,
			"handlesSubscription",
			map[string]interface{}{
				"username": "rocky",
			},
			true, func(resp Response) {
				log.Info("resp:", resp)
			})
		if err != nil {
			t.Errorf("host call handlesSubscription failed unexpected error: %v", err)
			return
		}

		log.Info("stop running")

		//关闭连接
		host.CloseListener(connectType)
		client.Close()

	}
}
