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
	"testing"
	"time"
)

func TestGenerateRangeNum(t *testing.T) {
	for i := 0;i<1000 ;i++  {
		num := GenerateRangeNum(0, 1023)
		t.Logf("num [%d]= %d", i, num)
	}
}

func TestConnectNode(t *testing.T) {

	//var (
	//	endRunning = make(chan bool, 1)
	//)

	//cert = Certificate{}

	//testUrl := "ws://192.168.30.28:8084/websocket"
	testUrl := "ws://127.0.0.1:9094"
	pid := "merchant"

	//主机
	host := DefaultNode
	host.Listen(":9094")

	time.Sleep(5 * time.Second)

	//客户端
	certA, _ := NewCertificate(RandomPrivateKey(),"")
	nodeA := NewOWTPNode(certA, 0, 0)

	err := nodeA.Connect(testUrl, pid)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(2 * time.Second)

	certB, _ := NewCertificate(RandomPrivateKey(),"")
	nodeB := NewOWTPNode(certB, 0, 0)

	err = nodeB.Connect(testUrl, pid)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(2 * time.Second)

	certC, _ := NewCertificate(RandomPrivateKey(),"")
	nodeC := NewOWTPNode(certC, 0, 0)

	err = nodeC.Connect(testUrl, pid)
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Logf("node close \n")

	nodeA.ClosePeer(pid)

	time.Sleep(5 * time.Second)

	host.Close()

	t.Logf("stop running \n")

	time.Sleep(10 * time.Second)

	t.Logf("end testing \n")

	//{
	//	"r": 2,
	//	"m": "subscribe",
	//	"n": 1,
	//	"t": 1528520843,
	//	"d": {
	//"status": 1000,
	//"msg": "success",
	//"result": null
	//},
	//	"s": "Qwse=="
	//}
	//

	//node.HandleFunc("getWalletInfo", getWalletInfo)
	//
	//err = node.Call("merchant", "subscribe", nil, false, func(resp Response) {
	//
	//})
	//
	//if err != nil {
	//	t.Errorf("Call failed unexpected error: %v", err)
	//	return
	//}

}

func TestListener(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	node := DefaultNode
	node.Listen(":9094")

	<- endRunning
	t.Logf("end listening \n")
}

func getWalletInfo(ctx *Context) {
	ctx.Resp = Response{
		Status: 0,
		Msg: "success",
		Result: map[string]interface{} {
			"symbols": []interface{} {
				map[string]interface{} {
					"coin": "btc",
					"walletID": "mywll",
				},
				map[string]interface{} {
					"coin": "btm",
					"walletID": "mkk",
				},
			},
		},
	}
}