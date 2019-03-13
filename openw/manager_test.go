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

package openw

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

var (
	testApp = "b4b1962d415d4d30ec71b28769fda585"
)

func testInitWalletManager() *WalletManager {

	tc := NewConfig()

	tc.IsTestnet = true
	tc.EnableBlockScan = false
	tc.SupportAssets = []string{
		"BTC",
		"QTUM",
		"LTC",
		"ETH",
		"NAS",
		"TRX",
		"BCH",
		//"ONT",
		"VSYS",
	}
	return NewWalletManager(tc)
	//tm.Init()
}

func TestWalletManager_CreateWallet(t *testing.T) {
	tm := testInitWalletManager()
	w := &openwallet.Wallet{Alias: "HELLO KITTY", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("wallet:", nw)
	log.Info("key:", key)

}

func TestWalletManager_ConcurrentCreateWallet(t *testing.T) {

	//w := &Wallet{Alias: "bitbank", IsTrust: true, Password: "12345678"}
	//_, _, err := tm.CreateWallet(defaultAppName, w)
	//if err != nil {
	//	log.Error(err)
	//	return
	//}

	tm := testInitWalletManager()

	var wg sync.WaitGroup
	timestamp := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				wid := fmt.Sprintf("w_%d_%d_%d", timestamp, id, j)
				w := &openwallet.Wallet{WalletID: wid, Alias: "bitbank", IsTrust: false, Password: "12345678"}
				_, _, err := tm.CreateWallet(testApp, w)
				if err != nil {
					log.Error("wallet[", id, "-", j, "] unexpected error:", err)
					continue
				}
				//log.Info("wallet[", id, "] :", nw)
				//log.Info("key:", key)
			}

		}(i)

	}

	wg.Wait()

	tm.CloseDB(testApp)
}

func TestWalletManager_GetWalletInfo(t *testing.T) {

	tm := testInitWalletManager()

	wallet, err := tm.GetWalletInfo(testApp, "VzQTLspxvbXSmfRGcN6LJVB8otYhJwAGWc")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestWalletManager_GetWalletList(t *testing.T) {

	tm := testInitWalletManager()

	list, err := tm.GetWalletList(testApp, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("wallet[", i, "] :", w)
	}
	log.Info("wallet count:", len(list))

	tm.CloseDB(testApp)
}

func TestWalletManager_CreateAssetsAccount(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	account := &openwallet.AssetsAccount{Alias: "MainnetQTUM", WalletID: walletID, Required: 1, Symbol: "QTUM", IsTrust: true}
	account, address, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)
	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAssetsAccountList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	list, err := tm.GetAssetsAccountList(testApp, walletID, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("account[", i, "] :", w)
	}
	log.Info("account count:", len(list))

	tm.CloseDB(testApp)

}

func TestWalletManager_CreateAddress(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "HyKAYbaLKXXa1U8YNsseP78YHGqB4vzSzJkKp8x4A7CC"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 5)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAddressList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "WMTUzB3LWaSKNKEQw9Sn73FjkEoYGHEp4B"
	accountID := "2by6wzbzw7cnWkxiA31xMHpFmE99bqL3BnjkUJnJtEN6"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w.Address)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}
