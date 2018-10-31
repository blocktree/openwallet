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

package nebulasio

import (
	"fmt"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
	"testing"
)

var wm *WalletManager

func init() {
	wm = NewWalletManager()
	wm.InitConfigFlow()
	wm.Config.ServerAPI = "http://127.0.0.1:8685"
	//wm.Config.ServerAPI = "https://mainnet.nebulas.io"
	wm.WalletClient = NewClient(wm.Config.ServerAPI, true)
}

func TestWalletManager_InitConfigFlow(t *testing.T) {
	wm.InitConfigFlow()
}



func TestApi(t *testing.T) {
//	ret := wm.WalletClient.CallGetHeader()
//	t.Logf(string(ret))

//	ret := wm.WalletClient.CallGetbalance("n1SWYV6hXxqV6pDTbULH4qvJXmxfdkqYRzs")
//	t.Logf(string(ret))

//	balance ,err := wm.WalletClient.CallGetaccountstate("n1SWYV6hXxqV6pDTbULH4qvJXmxfdkqYRzs","balance")
//	if err != nil {
//
//	}
//	fmt.Printf("balance=%s\n",balance)

//	GasPrice := wm.WalletClient.CallGetGasPrice()
//	fmt.Printf("GasPrice=%s\n",GasPrice)

	chain_id_result,_ := wm.WalletClient.CallGetnebstate("chain_id")

	chain_id := uint32(chain_id_result.Uint())
	fmt.Printf("chain_id=%v\n",chain_id)

//	wm.WalletClient.CallTestJson()
}

func TestCreateNewWallet(t *testing.T) {
	w, keyfile, err := wm.CreateNewWallet("wjq3", "123456789")
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
	fmt.Printf("00wm.WalletClient.BaseURL=%s\n",wm.WalletClient.BaseURL)
	err := wm.GetWalletList()  //此处会加载配置文件，配置文件中最好填
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
	fpath, addrs, err := wm.CreateBatchAddress("WMf6HjiKiXWoWm6ZigH5EmdeQVMCKKpLov", "123456789", 5)
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
	w, err := wm.GetWalletByID("WMf6HjiKiXWoWm6ZigH5EmdeQVMCKKpLov")
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
		b ,_ := wm.WalletClient.CallGetaccountstate(a.Address,"balance")
		t.Logf("%s    %s", a.Address, string(b))
	}
}

//查询指定钱包余额
func TestWalletManager_getBanlance(t *testing.T) {
	//w, err := wm.GetWalletByID("WEY5DDuXbvHrBUa5UBKmVpwLCwP69bieeB")
	w, err := wm.GetWalletByID("WMf6HjiKiXWoWm6ZigH5EmdeQVMCKKpLov")
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

//查询所有钱包余额
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

	var test_walletID string = "WMf6HjiKiXWoWm6ZigH5EmdeQVMCKKpLov"
	w, err := wm.GetWalletByID(test_walletID)
	if err != nil {
		t.Error("get wallet by id error")
		return
	}

	keystore, _ := w.HDKey("123456789")

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
		if a.Address == "n1RC2cksMwFr8ivFVYDydeUSWe4h9aWNc44" {
			sender = a

			break
		}
	}

	//n1Gi9RnhcwNVEAAaQywYjdxTBLyW4cWYArM
	//PrivateKey32字节:[110 177 159 126 62 241 11 234 124 91 38 126 149 178 174 6 61 198 198 144 28 230 70 160 138 7 184 194 154 210 33 41]
	//PublicKey33字节：已压缩,验签时需要先解压
	key, err := wm.getKeys(keystore, sender)
	if err != nil {
		t.Error("get key error")
	}

	dst := "n1KtPWggi7B9fokhdXgjPKpoyNggshjUPUM"

	fmt.Printf("key.Nonce=%v\n",key.Nonce)
	txid ,err := wm.Transfer( key,key.Address, dst,Gaslimit,"10000000000")
	if err != nil{
		t.Logf("Transfer Fail!\n",)
	}else{
		t.Logf("Transfer Success! txid=%s\n",txid)

		err := NotenonceInDB(key.Address , db)
		if err != nil {
			t.Error("NotenonceInDB error")
		}
	}

}


func TestModifySaveDB(t *testing.T) {

	wallet, err := wm.GetWalletByID("W4RWqnfcAmeBgUjoTD4TMNKA4bRg5nU5qK")
	if err != nil {
		fmt.Printf("get address failed0, err=%v\n", err)
	}

//	obj :=  wallet.GetAddress("n1UQcu9bLEogyRDd9QERmiULvWq3n2c4tvm")
//	fmt.Printf("before_obj.ExtParam=%v\n",obj.ExtParam)

	//GetAddress
	db, err := wallet.OpenDB()
	if err != nil {
		fmt.Printf("get address failed1, err=%v\n", err)

	}
	defer db.Close()

	var obj openwallet.Address
	err = db.One("Address", "n1PhfKZ4XNy8WfWgaCEjsUN33GrMVVHeLZy", &obj)
	if err != nil {
		fmt.Printf("get address failed2, err=%v\n", err)
	}

	//modifyAddress
	fmt.Printf("before_obj.ExtParam=%v\n",obj.ExtParam)
	obj.ExtParam = "9999"
	fmt.Printf("after_obj.ExtParam=%v\n",obj.ExtParam)

	//saveAddress
	tx, err := db.Begin(true)
	if err != nil {
		fmt.Printf("get address failed3, err=%v\n", err)
	}
	defer tx.Rollback()

	err = tx.Save(&obj)
	if err != nil {

		fmt.Printf("get address failed4, err=%v\n", err)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("get address failed5, err=%v\n", err)
	}

}

func TestCreateRawTransaction(t *testing.T) {

	raw_tx,err := wm.CreateRawTransaction("n1HEEtUecE5CQ3wCeHWvKVacsS2S5GAPCCu",
		"n1Prn7ZbZtd5CTN8Yrj4K9c3gD4u8tjFQzX",Gaslimit,"1000000","200000000000000",1)  //此处会加载配置文件，配置文件中最好填
	if err != nil {
		t.Error("get wallet list error")
		t.Logf(err.Error())
	}

	fmt.Printf("raw_tx=%+v\n",raw_tx)
}


func TestDecimal(t *testing.T) {

	//func (d Decimal) Coefficient() *big.Int
	//func (d Decimal) Div(d2 Decimal) Decimal

	balance_decimal := decimal.RequireFromString("18123343455453221")
	var coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000000000000000)

	balance_int := balance_decimal.IntPart()
	fmt.Printf("balance_int=%v\n",balance_int)

	balance_string := balance_decimal.String()
	fmt.Printf("balance_string=%v\n",balance_string)

	Div_balance := balance_decimal.Div(coinDecimal)
	fmt.Printf("Div_balance=%v\n",Div_balance)

	div_str := Div_balance.String()
	fmt.Printf("div_str=%s\n",div_str)

	balance_leave := decimal.RequireFromString("10000000000000")
	balance_safe := balance_decimal.Sub(balance_leave)
	fmt.Printf("balance_safe_str=%s\n",balance_safe.String())


}