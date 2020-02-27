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
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"sync"
	"testing"
)

var (
	multiConnectHost       *OWTPNode
	multiConnectHTTPURL    = "127.0.0.1:8422"
	multiConnectWSURL      = "127.0.0.1:8423"
	multiConnectPrv        = "FSomdQBZYzgu9YYuuSr3qXd8sP1sgQyk4rhLFo6gyi32"
	multiConnectHostNodeID = "54dZTdotBmE9geGJmJcj7Qzm6fzNrEUJ2NcDwZYp2QEp"
)

func TestMultiConnectHostRun(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	cert, _ := NewCertificate(multiConnectPrv)
	multiConnectHost = NewNode(NodeConfig{Cert: cert})
	//httpHost.SetPeerstore(globalSessions)
	fmt.Printf("nodeID = %s \n", multiConnectHost.NodeID())

	//config["enableSignature"] = "1"
	multiConnectHost.HandleFunc("getInfo", getInfo)
	multiConnectHost.HandlePrepareFunc(func(ctx *Context) {
		log.Notice("prepare")
		log.Infof("peer[%+v] config", ctx.Peer.ConnectConfig())
	})
	multiConnectHost.HandleFinishFunc(func(ctx *Context) {
		log.Notice("finish")
	})

	//listen HTTP
	multiConnectHost.Listen(
		ConnectConfig{
			Address:     multiConnectHTTPURL,
			ConnectType: HTTP,
		})

	//listen websocket
	multiConnectHost.Listen(
		ConnectConfig{
			Address:     multiConnectWSURL,
			ConnectType: Websocket,
		})

	multiConnectHost.SetOpenHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("peer[%s] connected", peer.ID)
		log.Infof("peer[%+v] config", peer.Config)
		//n.ClosePeer(peer.ID)
	})

	<-endRunning
}

func TestMultiConnectClientCall(t *testing.T) {
	var wait sync.WaitGroup

	for i := 0; i < 2; i++ {
		wait.Add(1)

		go func(k int) {
			var connectType, addr string
			if k%2 == 0 {
				connectType = HTTP
				addr = multiConnectHTTPURL
			} else {
				connectType = Websocket
				addr = multiConnectWSURL
			}

			client := RandomOWTPNode()
			client.HandleFunc("getInfo", getInfo)
			_, err := client.Connect(multiConnectHostNodeID, ConnectConfig{
				Address:     addr,
				ConnectType: connectType,
			})

			if err != nil {
				log.Errorf("Connect unexcepted error: %v", err)
				return
			}

			err = client.KeyAgreement(multiConnectHostNodeID, "aes")
			if err != nil {
				log.Errorf("KeyAgreement unexcepted error: %v", err)
				return
			}

			params := map[string]interface{}{
				"name": "chance",
				"age":  18,
			}

			err = client.Call(multiConnectHostNodeID, "getInfo", params, true, func(resp Response) {

				result := resp.JsonData()
				symbols := result.Get("symbols")
				fmt.Printf("symbols: %v\n", symbols)
			})

			if err != nil {
				log.Errorf("unexcepted error: %v", err)
				return
			}

			wait.Done()
		}(i)
	}
	wait.Wait()
}
