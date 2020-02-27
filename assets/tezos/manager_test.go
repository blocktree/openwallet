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

package tezos

import (
	"testing"
	"github.com/blocktree/openwallet/openwallet"
)

var wm *WalletManager

func init() {
	wm = NewWalletManager()
	wm.InitConfigFlow()
	wm.Config.ServerAPI = ""
	wm.WalletClient = NewClient(wm.Config.ServerAPI, false)
}

func TestWalletManager_InitConfigFlow(t *testing.T) {
	wm.InitConfigFlow()
}

func TestApi(t *testing.T) {
	ret := wm.WalletClient.CallGetHeader()
	t.Logf(string(ret))

	ret = wm.WalletClient.CallGetbalance("tz1Neor2KRu3zp5FdMox98sxYLvFqtUs4fCJ")
	t.Logf(string(ret))
}

func TestCreateNewWallet(t *testing.T) {
	w, keyfile, err := wm.CreateNewWallet("11", "jinxin")
	if err != nil {
		t.Error("create new wallet fail")
		return
	}

	t.Logf(w.WalletID)
	t.Logf(keyfile)

	ret, err := wm.GetWalletByID(w.WalletID)
	if err != nil {
		t.Error("get wallet by id err")
		t.Logf(err.Error())
		return
	}

	t.Logf(ret.Alias)
}

func TestLoadConfig(t *testing.T) {
	err := wm.LoadConfig()
	if err != nil {
		t.Error("load config error")
		t.Logf(err.Error())
	}
}

func TestWalletManager_GetWallets(t *testing.T) {
	err := wm.GetWalletList()
	if err != nil {
		t.Error("get wallet list error")
		t.Logf(err.Error())
	}
}

func TestWalletConfig_PrintConfig(t *testing.T) {
	err := wm.Config.PrintConfig()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestWalletManager_CreateBatchAddress(t *testing.T) {
	var addrs []*openwallet.Address
	fpath, addrs, err := wm.CreateBatchAddress("W7nQCzsbEBtSS7U4Vgz82s4RDnwoS6vYqd", "jinxin", 1)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf(fpath)
	for _, a := range addrs {
		t.Logf(a.Address)
	}
}

func TestGetAddreses(t *testing.T) {
	w, err := wm.GetWalletByID("VyXcihm3vcXxv7nnBsFxq7TRNYcdBmFoPW")
	if err != nil {
		t.Error("get wallet by id error")
		return
	}

	db, err := w.OpenDB()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	for _, a := range addrs {
		b := wm.WalletClient.CallGetbalance(a.Address)
		t.Logf("%s    %s", a.Address, string(b))
	}
}

func TestWalletManager_getBanlance(t *testing.T) {
	//w, err := wm.GetWalletByID("WEY5DDuXbvHrBUa5UBKmVpwLCwP69bieeB")
	w, err := wm.GetWalletByID("VyXcihm3vcXxv7nnBsFxq7TRNYcdBmFoPW")
	if err != nil {
		t.Error("get wallet by id error")
		return
	}

	var addrs []*openwallet.Address
	balance, addrs , _ := wm.getWalletBalance(w)
	t.Log(balance)

	for _, a := range(addrs) {
		if a.Balance != "0" {
			t.Logf("addresss: %s  balance:%s", a.Address, a.Balance)
		}
	}
}

func TestWalletManager_getWalletBalance(t *testing.T) {
	//查询所有钱包信息
	wallets, err := wm.GetWallets()
	if err != nil {
		t.Logf("The node did not create any wallet!\n")
	}
	addrs := wm.printWalletList(wallets, true)

	for _, addr := range(addrs) {
		for _, a := range(addr) {
			if a.Balance != "0" {
				t.Logf("addresss: %s  balance:%s", a.Address, a.Balance)
			}
		}
	}
}

func TestWalletManager_TransferFlow(t *testing.T) {
	w, err := wm.GetWalletByID("WGp7QW8jG3BF9CCrQvBq54zpS6huYasiAC")
	if err != nil {
		t.Error("get wallet by id error")
		return
	}

	keystore, _ := w.HDKey("1234qwer")

	db, err := w.OpenDB()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	var sender *openwallet.Address
	//key, _ := wm.getKeys(keystore, addrs[0])
	for _, a := range addrs {
		if a.Address == "tz1bQce59RgUkuV9WfvTvkZGiLLt8aedqvzd" {
			sender = a
			break
		}
	}

	key, err := wm.getKeys(keystore, sender)
	if err != nil {
		t.Error("get key error")
	}

	dst := "tz1Neor2KRu3zp5FdMox98sxYLvFqtUs4fCJ"

	inj, pre := wm.Transfer(*key, dst, "100", "100", "100", "801")
	t.Logf("inj: %s\n, pre: %s\n", inj, pre)
}

func Test_Sign(t *testing.T) {

	str := "cb81d878a257e9f7bb864a33700f0c411c142d2f4eaa423338112e37152a1879080000acf886e911fa1118ffd5e7394b740ddbaa81b67700acb60cc801c801000000e3f24bd5bca29cb50422b59a3a779698c0b14e4400"
	w, err := wm.GetWalletByID("WGp7QW8jG3BF9CCrQvBq54zpS6huYasiAC")
	if err != nil {
		t.Error("get wallet by id error")
		return
	}

	keystore, _ := w.HDKey("1234qwer")

	db, err := w.OpenDB()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	var sender *openwallet.Address
	//key, _ := wm.getKeys(keystore, addrs[0])
	for _, a := range addrs {
		if a.Address == "tz1bQce59RgUkuV9WfvTvkZGiLLt8aedqvzd" {
			sender = a
			break
		}
	}

	key, err := wm.getKeys(keystore, sender)
	if err != nil {
		t.Error("get key error")
	}

	sign, _, _ := wm.signTransaction(str, key.PrivateKey, watermark["generic"])
	t.Log(sign)
	}

/*
	manager_test.go:121: tz1NGveHiqsNEjUUGqkUHRfp69PkGyNaq8Ax
	manager_test.go:121: tz1QtdKMHA8vq6EfeU9YMvsNiyNomErX4Hv7
	manager_test.go:121: tz1RFB11LxGgJj6QxS7wjk1zPrqE4krTbjx1
	manager_test.go:121: tz1UUH6ixu3AdzjMMjEPSrC9ErsH4VMEvDTB
	manager_test.go:121: tz1UjEGWubFhpBm4kHGqKNAmf9tfq4nKri64
	manager_test.go:121: tz1XSuiDu5Fevzwv3CG26dxmsinejRsUviC2
	manager_test.go:121: tz1XgSKaXyhx1tB3HrhScJP3cJcZA2LXC9fF
	manager_test.go:121: tz1Y2yZvK2iMMxVcaNm9h5AhdC6uKvRc2JQf
	manager_test.go:121: tz1YtYZS6KKWJeRjZCCCY4FYu7qyS8hEEz4Z
	manager_test.go:121: tz1hKLUu55xmfKCizGj5X5w8akxKJrUwnVU3
*/
