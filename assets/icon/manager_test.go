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

package icon

import (
	"testing"
	"github.com/blocktree/OpenWallet/openwallet"
	"time"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strconv"
	"github.com/shopspring/decimal"
)

var wm *WalletManager

func init() {
	wm = NewWalletManager()
	wm.InitConfigFlow()
	wm.Config.ServerAPI = "https://ctz.solidwallet.io/api/v3"
	wm.WalletClient = NewClient(wm.Config.ServerAPI, false)
}

func TestWalletManager_InitConfigFlow(t *testing.T) {
	wm.InitConfigFlow()
}

func TestCreateNewWallet(t *testing.T) {
	w, keyfile, err := wm.CreateNewWallet("11", "1234qwer")
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
	fpath, addrs, err := wm.CreateBatchAddress("WGBtjfnG5qVucwAtTdYaaQSG9LGBMxsXuW", "1234qwer", 5)
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
	w, err := wm.GetWalletByID("WGBtjfnG5qVucwAtTdYaaQSG9LGBMxsXuW")
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
		b, _ := wm.WalletClient.Call_icx_getBalance(a.Address)
		t.Logf("%s    %s", a.Address, string(b))
	}
}

func TestWalletManager_getBanlance(t *testing.T) {
	//w, err := wm.GetWalletByID("WEY5DDuXbvHrBUa5UBKmVpwLCwP69bieeB")
	w, err := wm.GetWalletByID("WGBtjfnG5qVucwAtTdYaaQSG9LGBMxsXuW")
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
	w, err := wm.GetWalletByID("WGBtjfnG5qVucwAtTdYaaQSG9LGBMxsXuW")
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
		if a.Address == "hx2006f91de4cd0b9ce74cb00a06e66eaeb44c70b1" {
			sender = a
			break
		}
	}

	key, err := wm.getKeys(keystore, sender)
	t.Log(key)
	if err != nil {
		t.Error("get key error")
	}
}

func TestWalletManager_CalculateTxHash(t *testing.T) {
	from := "hxb12addba58c934ff924aa87ee65d06ee20f89eb8"
	to := "hxb12addba58c934ff924aa87ee65d06ee20f89eb8"
	value := 0.1

	_, hash := wm.CalculateTxHash(from, to, value, 1000000, 10)
	t.Log(hexutil.Encode(hash[:]))

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	t.Log(timestamp)
}

func TestWalletManager_getTranx(t *testing.T) {
	txhash := "0xfda563191b7f97c7193f08eec38bf830987a5f3d9e6d550c586077cf473be215"

	ret, _ := wm.WalletClient.Call_icx_getTransactionByHash(txhash)

	t.Log(ret)
}

func TestWalletManager_Transfer(t *testing.T) {
	w, err := wm.GetWalletByID("WGBtjfnG5qVucwAtTdYaaQSG9LGBMxsXuW")
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
	for _, a := range addrs {
		if a.Address == "hx2006f91de4cd0b9ce74cb00a06e66eaeb44c70b1" {
			sender = a
			break
		}
	}

	key, err := wm.getKeys(keystore, sender)
	t.Log(key.PrivateKey)
	t.Log(len(key.PrivateKey))

	from := "hx2006f91de4cd0b9ce74cb00a06e66eaeb44c70b1"
	to := "hxb12addba58c934ff924aa87ee65d06ee20f89eb8"
	value := 0.02

	ret, err := wm.Transfer(key.PrivateKey, from, to, value, 100000, 100)
	if err != nil {
		t.Error(err)
	}
	t.Log(ret)
}

func Test_timeFormat(t *testing.T) {
	mic := time.Now().UnixNano() / 1000
	t.Log(mic)

	tm := "0x" + strconv.FormatInt(time.Now().UnixNano() / 1000, 16)
	t.Log(tm)

	hex := "d19c2ff547173c4"
	f, _ := strconv.ParseInt(hex, 16, 64)
	s := decimal.New(f, 0).Div(coinDecimal)
	t.Log(s)

	bigint, _ := hexutil.DecodeBig("0x2c68af0bb140000")
	b := decimal.NewFromBigInt(bigint, 0).Div(coinDecimal)
	t.Log(b)

	a := decimal.NewFromFloat(0.0001)
	c, _ := a.Float64()
	t.Log(c)

}


 func TestClient_Call_icx_sendTransaction(t *testing.T) {
	 req := map[string]interface{}{
	 	"version": "0x3",
	 	"from": "hx2006f91de4cd0b9ce74cb00a06e66eaeb44c70b1",
	 	"stepLimit": "0xf4240",
	 	"timestamp": "0x5796c8f55c75e",
	 	"nid": "0x1",
	 	"to": "hxb12addba58c934ff924aa87ee65d06ee20f89eb8",
	 	"value": "0x8f0d180",
	 	"nonce": "0x64",
	 	"signature": "q2TurzvpijxzXJlsWCGP6JbNm6MFu7XnDorD4jCbDgtywtg4tZ/IdBzhmDc4vBFQsGrUiyGvjWZIxfwYl5pCTAE=",
	 }

	 ret, err := wm.WalletClient.Call_icx_sendTransaction(req)
	 if err != nil {
	 	t.Log(err)
	 	return
	 }
	 t.Log(ret)
 }

