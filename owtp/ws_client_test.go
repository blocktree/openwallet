/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package owtp

import (
	"encoding/json"
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/session"
	"testing"
	"time"
)

var (
	testUrl = "ws://192.168.30.4:8083/websocket?a=dajosidjaiosjdioajsdioajsdiowhefi&t=1529669837&n=2&s=adisjdiasjdioajsdiojasidjioasjdojasd"
)

func init() {

}

func TestDial(t *testing.T) {

	client, err := Dial("hello", testUrl, nil, nil, 1024, 1024)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	defer client.close()

	client.openPipe()
}

func TestEncodeDataPacket(t *testing.T) {

	//封装数据包
	packet := DataPacket{
		Method:    "subscribe",
		Req:       WSRequest,
		Nonce:     1,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"sss": "sdfsdf",
		},
	}

	respBytes, err := json.Marshal(packet)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	t.Logf("Send: %s \n", string(respBytes))

}

var (
	wsHost           *OWTPNode
	wsURL            = ":8423"
	wsHostPrv        = "FSomdQBZYzgu9YYuuSr3qXd8sP1sgQyk4rhLFo6gyi32"
	wsHostNodeID     = "54dZTdotBmE9geGJmJcj7Qzm6fzNrEUJ2NcDwZYp2QEp"
	wsGlobalSessions *SessionManager
)

func testSetupWSGlobalSession() {
	wsGlobalSessions, _ = NewSessionManager("memory", &session.ManagerConfig{
		Gclifetime: 10,
	})
	go wsGlobalSessions.GC()
}

func init() {
	testSetupWSGlobalSession()
}

func TestWSHostRun(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	cert, _ := NewCertificate(wsHostPrv, "aes")
	wsHost = NewOWTPNode(cert, 0, 0)
	wsHost.SetPeerstore(wsGlobalSessions)
	fmt.Printf("nodeID = %s \n", wsHost.NodeID())
	config := ConnectConfig{}
	config.Address = wsURL
	config.ConnectType = Websocket
	config.EnableSignature = true
	wsHost.HandleFunc("getInfo", getInfo)
	wsHost.HandlePrepareFunc(func(ctx *Context) {
		log.Notice("remoteAddress:", ctx.RemoteAddress)
		log.Notice("prepare")
		//ctx.ResponseStopRun(nil, StatusSuccess, "success")
	})
	wsHost.HandleFinishFunc(func(ctx *Context) {
		username := ctx.GetSession("username")
		log.Notice("username:", username)
		log.Notice("finish")

		params := map[string]interface{}{
			"name": "jjo",
			"age":  22,
		}

		err := wsHost.Call(ctx.PID, "getInfo", params, true, func(resp Response) {

			result := resp.JsonData()
			symbols := result.Get("symbols")
			fmt.Printf("symbols: %v\n", symbols)
		})

		if err != nil {
			t.Errorf("unexcepted error: %v", err)
			return
		}
	})

	wsHost.SetOpenHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("peer[%s] connected", peer.ID)
		log.Infof("peer[%+v] config", peer.Config)
		//n.Call(peer.ID, "hello", nil, true, func(resp Response) {
		//	log.Infof("resp: %v", resp)
		//})
		//n.ClosePeer(peer.ID)
	})

	//断开连接
	wsHost.SetCloseHandler(func(n *OWTPNode, peer PeerInfo) {
		
	})

	wsHost.Listen(config)

	<-endRunning
}

func TestWSClientCall(t *testing.T) {

	config := ConnectConfig{}
	config.Address = wsURL
	config.ConnectType = Websocket
	config.EnableSignature = true
	wsClient := RandomOWTPNode()
	wsClient.SetPeerstore(wsGlobalSessions)
	//_, pub := httpClient.Certificate().KeyPair()
	//log.Info("pub:", pub)
	wsClient.HandleFunc("getInfo", getInfo)
	err := wsClient.Connect(wsHostNodeID, config)
	if err != nil {
		t.Errorf("Connect unexcepted error: %v", err)
		return
	}

	wsClient.SetOpenHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("client: peer[%s] connected", peer.ID)
		log.Infof("client: peer[%+v] config", peer.Config)
		//n.ClosePeer(peer.ID)
	})

	//断开连接
	wsClient.SetCloseHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("client: peer[%s] disconnected", peer.ID)
	})

	err = wsClient.KeyAgreement(httpHostNodeID, "aes")
	if err != nil {
		t.Errorf("KeyAgreement unexcepted error: %v", err)
		return
	}

	params := map[string]interface{}{
		"name": "chance",
		"age":  18,
	}
	//err = wsClient.Connect(wsHostNodeID, config)
	err = wsClient.Call(wsHostNodeID, "getInfo", params, true, func(resp Response) {

		result := resp.JsonData()
		symbols := result.Get("symbols")
		fmt.Printf("symbols: %v\n", symbols)
	})

	if err != nil {
		t.Errorf("unexcepted error: %v", err)
		return
	}

}


func TestWSClientConnectAndCall(t *testing.T) {

	config := ConnectConfig{}
	config.Address = wsURL
	config.ConnectType = Websocket
	config.EnableSignature = true
	wsClient := RandomOWTPNode()
	wsClient.SetPeerstore(wsGlobalSessions)
	wsClient.HandleFunc("getInfo", getInfo)

	wsClient.SetOpenHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("client: peer[%s] connected", peer.ID)
		log.Infof("client: peer[%+v] config", peer.Config)
		//n.ClosePeer(peer.ID)
	})

	//断开连接
	wsClient.SetCloseHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("client: peer[%s] disconnected", peer.ID)
	})

	params := map[string]interface{}{
		"name": "chance",
		"age":  18,
	}

	for i := 0;i<100;i++ {
		//err = wsClient.Connect(wsHostNodeID, config)
		err := wsClient.ConnectAndCall(wsHostNodeID, config, "getInfo", params, false, func(resp Response) {

			result := resp.JsonData()
			symbols := result.Get("symbols")
			fmt.Printf("symbols: %v\n", symbols)
		})

		if err != nil {
			t.Errorf("unexcepted error: %v", err)
			return
		}
		time.Sleep(1 * time.Second)
	}
}
