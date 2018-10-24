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

package manager

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

var (
	tc = NewConfig()
	//testApp = "openw"
	testApp = "b4b1962d415d4d30ec71b28769fda585"
	tm *WalletManager
)

func init() {
	tc.IsTestnet = true
	tc.EnableBlockScan = true
	tc.SupportAssets = []string{
		//"BTC",
		"QTUM",
		//"LTC",
		//"ETH",
	}
	tm = NewWalletManager(tc)
	//tm.Init()
}

func TestWalletManager_CreateWallet(t *testing.T) {

	w := &openwallet.Wallet{Alias: "Apple", IsTrust: true, Password: "12345678"}
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
	wallet, err := tm.GetWalletInfo(testApp, "W3hxZRqw67PbBq5GFpULkaAJdKN9Mzasj5")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestWalletManager_GetWalletList(t *testing.T) {
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

	walletID := "W3hxZRqw67PbBq5GFpULkaAJdKN9Mzasj5"
	account := &openwallet.AssetsAccount{Alias: "Simon", WalletID: walletID, Required: 1, Symbol: "QTUM", IsTrust: true}
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

	walletID := "W3hxZRqw67PbBq5GFpULkaAJdKN9Mzasj5"
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

	walletID := "W3hxZRqw67PbBq5GFpULkaAJdKN9Mzasj5"
	//accountID := "KhJdnr4UJLdbeQcMvZgedyYykRVTdLaMLbsV2mx3GZiMva9Kfb"
	accountID := "26THHhacorJKJrF2RNCwkkNUrv16fnksdjsa7PxQXWry"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 10)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	time.Sleep(30 * time.Second)

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAddressList(t *testing.T) {
	walletID := "WEP6cD2YSV773QZw5UuSS5U74XKdw6oQE2"
	accountID := "26THHhacorJKJrF2RNCwkkNUrv16fnksdjsa7PxQXWry"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}
