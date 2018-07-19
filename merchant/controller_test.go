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

package merchant

import (
	"github.com/blocktree/OpenWallet/owtp"
	"testing"
	"time"
	"path/filepath"
	"github.com/imroc/req"
	"log"
)

func init() {
	nodeConfig = NodeConfig{
		NodeKey:         "",
		PublicKey:       "dajosidjaiosjdioajsdioajsdiowhefi",
		PrivateKey:      "",
		MerchantNodeURL: "ws://192.168.30.4:8084/websocket",
		NodeID:          1,
		CacheFile:       filepath.Join(merchantDir, cacheFile),
	}

}

func generateCTX(method string, inputs interface{}) *owtp.Context {
	nonce := uint64(time.Now().Unix())
	ctx := owtp.NewContext(owtp.WSRequest, nonce, method, inputs)
	return ctx
}

func TestSubscribe(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	inputs := []Subscription {
		Subscription{Type: 2, Coin:"btc",WalletID:"21212",Version:222},
		Subscription{Type: 2, Coin:"btm",WalletID:"21212",Version:222},
		Subscription{Type: 2, Coin:"ltc",WalletID:"21212",Version:222},
	}

	ctx := generateCTX("subscribe", inputs)

	m.subscribe(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	<- endRunning
}

func TestCreateWallet(t *testing.T) {

	var (
		//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("CreateWallet failed unexpected error: %v", err)
		return
	}

	inputs := map[string] interface{} {
		"coin": "btc",
		"alias": "YOU Mac",
		"passwordType": 0,
		"password": "1234qwer",
	}

	ctx := generateCTX("createWallet", inputs)

	m.createWallet(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	//<- endRunning

}


func TestCreateAddress(t *testing.T) {

	var (
	//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("CreateWallet failed unexpected error: %v", err)
		return
	}

	inputs := map[string] interface{} {
		"coin": "btc",
		"walletID": "WAAVruvecxJTNxcuMBdZc6QRq5WQtMiXKe",
		"count": 100,
		"password": "1234qwer",
	}

	ctx := generateCTX("createAddress", inputs)

	m.createAddress(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	//<- endRunning

}


func TestGetAddressList(t *testing.T) {

	var (
	//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("CreateWallet failed unexpected error: %v", err)
		return
	}

	inputs := map[string] interface{} {
		"coin": "btc",
		"walletID": "WAAVruvecxJTNxcuMBdZc6QRq5WQtMiXKe",
		"offset": 0,
		"limit": 1000,
	}

	ctx := generateCTX("getAddressList", inputs)

	m.getAddressList(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	//<- endRunning

}


func TestGetWalletList(t *testing.T) {

	var (
	//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("CreateWallet failed unexpected error: %v", err)
		return
	}

	inputs := map[string] interface{} {
		"coin": "btc",
	}

	ctx := generateCTX("getWalletList", inputs)

	m.getWalletList(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	//<- endRunning

}


func TestSubmitTransaction(t *testing.T) {

	var (
	//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("CreateWallet failed unexpected error: %v", err)
		return
	}

	inputs := map[string] interface{} {
		"withdraws": []interface{} {
			map[string]interface{} {
				"address":"mffc7puT4ZrCQGuAZAUuzjRWPKZKKYT4qG",
				"amount":"0.0001",
				"coin":"BTC",
				"isMemo":0,
				"sid":"BTC@WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu@null",
				"walletID":"WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu",
				"password": "1234qwer",
			},
		},
	}

	ctx := generateCTX("submitTransaction", inputs)

	m.submitTransaction(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)

	//<- endRunning

}

func TestHTTP(t *testing.T) {
	r, err := req.Get("http://192.168.2.193:10050/chains/main/blocks/")
	if err != nil {
		log.Printf("unexpected error: %v", err)
		return
	}
	log.Printf("%+v\n", r)
}