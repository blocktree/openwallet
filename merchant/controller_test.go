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

package merchant

import (
	"github.com/blocktree/openwallet/owtp"
	"github.com/imroc/req"
	"log"
	"path/filepath"
	"testing"
	"time"
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
	//endRunning = make(chan bool, 1)
	)

	m, err := NewMerchantNode(nodeConfig)
	if err != nil {
		t.Errorf("GetChargeAddressVersion failed unexpected error: %v", err)
	}

	inputs := []Subscription{
		Subscription{Type: 1, Coin: "BTC", WalletID: "WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu", Version: 1},
		Subscription{Type: 2, Coin: "btm", WalletID: "AD044IDNF42", Version: 1},
		Subscription{Type: 2, Coin: "BTC", WalletID: "WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu", Version: 1},
	}

	ctx := generateCTX("subscribe", inputs)

	m.subscribe(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

	//<- endRunning
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

	inputs := map[string]interface{}{
		"coin":         "btc",
		"alias":        "YOU Mac",
		"passwordType": 0,
		"password":     "1234qwer",
	}

	ctx := generateCTX("createWallet", inputs)

	m.createWallet(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

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

	inputs := map[string]interface{}{
		"coin":     "btc",
		"walletID": "WAAVruvecxJTNxcuMBdZc6QRq5WQtMiXKe",
		"count":    100,
		"password": "1234qwer",
	}

	ctx := generateCTX("createAddress", inputs)

	m.createAddress(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

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

	inputs := map[string]interface{}{
		"coin":     "btc",
		"walletID": "WAAVruvecxJTNxcuMBdZc6QRq5WQtMiXKe",
		"offset":   0,
		"limit":    1000,
	}

	ctx := generateCTX("getAddressList", inputs)

	m.getAddressList(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

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

	inputs := map[string]interface{}{
		"coin": "btc",
	}

	ctx := generateCTX("getWalletList", inputs)

	m.getWalletList(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

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

	inputs := map[string]interface{}{
		"withdraws": []interface{}{
			map[string]interface{}{
				"address":  "mffc7puT4ZrCQGuAZAUuzjRWPKZKKYT4qG",
				"amount":   "0.0001",
				"coin":     "BTC",
				"isMemo":   0,
				"sid":      "BTC@WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu@null",
				"walletID": "WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu",
				"password": "1234qwer",
			},
		},
	}

	ctx := generateCTX("submitTransaction", inputs)

	m.submitTransaction(ctx)

	t.Logf("reponse: %v\n", ctx.Resp)

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

func TestSubSlice(t *testing.T) {
	pageSize := 10
	recharges := []string{
		//"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
		//"21", "22", "23", "24", "25", "26", "27", "28", "29", "30",
		//"31", "32", "33", "34", "35", "36", "37", "38", "39", "40",
		//"41", "42", "43", "44", "45", "46", "47", "48", "49", "50",
		"51", "52", "53", "54", "55", "56", "57",
	}

	for {

		var subRecharges []string

		if len(recharges) == 0 {
			break
		}

		if len(recharges) <= pageSize {
			subRecharges = recharges
			recharges = recharges[:0]
		} else {
			subRecharges = recharges[:pageSize]
			recharges = recharges[pageSize:]
		}

		log.Print(subRecharges)

	}

	log.Println("Game over")

}
