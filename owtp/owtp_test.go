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

import "testing"

func TestGenerateRangeNum(t *testing.T) {
	for i := 0;i<1000 ;i++  {
		num := GenerateRangeNum(0, 1023)
		t.Logf("num [%d]= %d", i, num)
	}
}

func TestConnectNode(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	testUrl := "ws://192.168.30.4:8083/websocket?a=dajosidjaiosjdioajsdioajsdiowhefi&t=1529669837&n=4&s=adisjdiasjdioajsdiojasidjioasjdojasd"
	node := NewOWTPNode(1, testUrl, "")
	err := node.Connect()
	if err != nil {
		t.Errorf("Connect failed unexpected error: %v", err)
		return
	}

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

	node.HandleFunc("getWalletInfo", getWalletInfo)

	err = node.Call("subscribe", nil, func(resp Response) {

	}, false)

	if err != nil {
		t.Errorf("Call failed unexpected error: %v", err)
		return
	}

	<- endRunning
	t.Logf("end connect \n")

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