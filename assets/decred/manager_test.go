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

package decred

import (
	"github.com/codeskyblue/go-sh"
	"github.com/shopspring/decimal"
	"math"
	"testing"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/asdine/storm/q"
	"github.com/asdine/storm"
	"path/filepath"
)

var (
	tw *WalletManager
)

func init() {

	tw = NewWalletManager()

	tw.config.walletAPI = "https://192.168.2.193:10066"
	tw.config.chainAPI = "https://192.168.2.193:10065"
	tw.config.rpcUser = "walletdcr"
	tw.config.rpcPassword = "walletdcr"
	token := basicAuth(tw.config.rpcUser, tw.config.rpcPassword)
	tw.walletClient = NewClient(tw.config.walletAPI, token, true)
	tw.dcrdClient = NewClient(tw.config.chainAPI, token, false)
}

func TestWalletManager_InitConfigFlow(t *testing.T) {
	err := tw.InitConfigFlow()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestCreateNewWallet(t *testing.T) {
	_, _, err := tw.CreateNewWallet("jj", "jinxin")
	if err != nil {
		t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		return
	}
}


func TestGetAddressesByAccount(t *testing.T) {
	addresses, err := tw.GetAddressesByAccount("default")
	if err != nil {
		t.Errorf("GetAddressesByAccount failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		if a == "TsS18cTdw81WnH8z2JQ89WoiGs6ReWXiPwJ" {
			t.Logf("specify address[%d] = %s\n", i, a)
		}
		//t.Logf("GetAddressesByAccount address[%d] = %s\n", i, a)
	}
}

func TestCreateBatchAddress(t *testing.T) {
	_, _, err := tw.CreateBatchAddress("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW", "jinxin", 10)
	if err != nil {
		t.Errorf("CreateBatchAddress failed unexpected error: %v\n", err)
		return
	}
}

func TestUnlockWallet(t *testing.T) {
	err := tw.UnlockWallet("jinxin", 1)
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestGetWalletBalance(t *testing.T) {

	tests := []struct {
		name string
		tag  string
	}{
		{
			name: "WFPCC2GUN6MSZ4V6E2b7fHEiY5jtWdQuRD",
			tag:  "first",
		},
		{
			name: "W3K2C9q4tM4PDiRQfwz3FbsZcH2AMfpqH6",
			tag:  "second",
		},
		{
			name: "WDMKSZgiYM3gQjzEKLSqCoyV65LMWBbW7Gs",
			tag:  "all",
		},
		{
			name: "W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW",
			tag:  "testnet",
		},
	}

	for i, test := range tests {
		balance := tw.GetWalletBalance(test.name)
		t.Logf("GetWalletBalance[%d] %s balance = %s \n", i, test.name, balance)
	}

}

func TestGetWallets(t *testing.T) {
	wallets, err := tw.GetWallets()
	if err != nil {
		t.Errorf("GetWallets failed unexpected error: %v\n", err)
		return
	}

	for i, w := range wallets {
		t.Logf("GetWallets wallet[%d] = %v", i, w)
	}
}


func TestGetWalletList(t *testing.T) {
	err := tw.GetWalletList()
	if err != nil {
		t.Errorf("GetWalleList failed unexpected error: %v\n", err)
		return
	}
}

//func TestCreateBatchPrivateKey(t *testing.T) {
//
//	w, err := tw.GetWalletInfo("WKD6QUMLyv93qBBdnURokKCrQKHeTQYeVu")
//	if err != nil {
//		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//
//	key, err := w.HDKey("123")
//	if err != nil {
//		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//
//	wifs, err := tw.CreateBatchPrivateKey(key, 5)
//	if err != nil {
//		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//
//	for i, wif := range wifs {
//		t.Logf("CreateBatchPrivateKey[%d] wif = %v \n", i, wif)
//	}
//
//}
//
//func TestWalletManager_ImportPrivateKey(t *testing.T) {
//	tw.UnlockWallet("123",5)
//	err := tw.ImportPrivateKey("PtWVHGkven8UKMgrjbrnNDvBBQ6a4NN3GxtMefXNNhiEiDK1umBq2","imported")
//	if err != nil {
//		t.Errorf("ImportPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//}

func TestWalletManager_CreateNewAddress(t *testing.T) {
	//address, err := tw.CreateNewAddress("")
	//if err != nil {
	//	t.Errorf("CreateNewAddress failed unexpected error: %v\n", err)
	//	return
	//}
	//t.Logf("address: %s \n", address)
}

//func TestImportMulti(t *testing.T) {
//
//	addresses := []string{
//		"1CoRcQGjPEyWmB1ZyG6CEDN3SaMsaD3ERa",
//		"1ESGCsXkNr3h5wvWScdCpVHu2GP3KJtCdV",
//	}
//
//	keys := []string{
//		"L5k8VYSvuZxC5FCczGVC8MmnKKix3Mcs6t185eUJVKTzZb1f6bsX",
//		"L3RVDjPVBSc7DD4WtmzbHkAHJW4kDbyXbw4vBppZ4DRtPt5u8Naf",
//	}
//
//	UnlockWallet("1234qwer", 120)
//	failed, err := ImportMulti(addresses, keys, "Zhiquan Test")
//	if err != nil {
//		t.Errorf("ImportMulti failed unexpected error: %v\n", err)
//	} else {
//		t.Errorf("ImportMulti result: %v\n", failed)
//	}
//}

func TestBackupWallet(t *testing.T) {

	tw.config.walletDataPath = "/Users/maizhiquan/Library/Application Support/hcGUI/wallets/mainnet/zhiquan911/mainnet/"

	backupFile, err := tw.BackupWallet("WBJH3u4QCFYcGTisDBiZvssrkG8YJAcmhS")
	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("BackupWallet filePath: %v\n", backupFile)
	}
}

//func TestBackupWalletData(t *testing.T) {
//	tw.config.walletDataPath = "/home/www/btc/testdata/testnet3/"
//	tmpWalletDat := fmt.Sprintf("tmp-walllet-%d.dat", time.Now().Unix())
//	backupFile := filepath.Join(tw.config.walletDataPath, tmpWalletDat)
//	err := tw.BackupWalletData(backupFile)
//	if err != nil {
//		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
//	} else {
//		t.Errorf("BackupWallet filePath: %v\n", backupFile)
//	}
//}

func TestGOSH(t *testing.T) {
	//text, err := sh.Command("go", "env").Output()
	//text, err := sh.Command("wmd", "version").Output()
	text, err := sh.Command("wmd", "config", "see", "-s", "btm").Output()
	if err != nil {
		t.Errorf("GOSH failed unexpected error: %v\n", err)
	} else {
		t.Errorf("GOSH output: %v\n", string(text))
	}
}

func TestGetBlockChainInfo(t *testing.T) {
	b, err := tw.GetBlockChainInfo()
	if err != nil {
		t.Errorf("GetBlockChainInfo failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetBlockChainInfo info: %v\n", b)
	}
}

func TestListUnspent(t *testing.T) {
	utxos, err := tw.ListUnspent(1)
	if err != nil {
		t.Errorf("ListUnspent failed unexpected error: %v\n", err)
		return
	}

	for _, u := range utxos {
		t.Logf("ListUnspent: %v\n", u)
	}
}

func TestGetAddressesFromLocalDB(t *testing.T) {
	//wallet, _ := tw.GetWalletInfo("W3K2C9q4tM4PDiRQfwz3FbsZcH2AMfpqH6")
	addresses, err := tw.GetAddressesFromLocalDB("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW", false,0, -1)
	if err != nil {
		t.Errorf("GetAddressesFromLocalDB failed unexpected error: %v\n", err)
		return
	}
	//db, _ := wallet.OpenDB()
	//defer db.Close()
	for i, a := range addresses {
		t.Logf("GetAddressesFromLocalDB address[%d] = %v\n", i, a)
		//a.WatchOnly = false
		//db.Save(a)
	}
}

//GetAddressesFromLocalDB 从本地数据库
func TestGetAddressesFromLocalDBPath(t *testing.T) {


	db, err := storm.Open(filepath.Join(tw.config.dbPath, "hccharge.db"))
	if err != nil {
		return
	}
	defer db.Close()

	var addresses []*openwallet.Address
	//err = db.Find("WalletID", walletID, &addresses)
	err = db.Select(q.And(
		q.Eq("Address", "TseMaJhgyH4aZuc5wQMeMbznJkRcoKWmFdo"),
		q.Eq("AccountID", "hccharge"),
		q.Eq("WatchOnly", true),
	)).Skip(0).Find(&addresses)

	for i, a := range addresses {
		t.Logf("GetAddressesFromLocalDB address[%d] = %v\n", i, a)
		//a.WatchOnly = false
	}

}

func TestWalletManager_RescanCorewallet(t *testing.T) {
	err := tw.RescanCorewallet(7000)
	if err != nil {
		t.Errorf("RescanCorewallet failed unexpected error: %v\n", err)
		return
	}

	t.Logf("RescanCorewallet successfully.\n")
}

func TestRebuildWalletUnspent(t *testing.T) {

	err := tw.RebuildWalletUnspent("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW")
	if err != nil {
		t.Errorf("RebuildWalletUnspent failed unexpected error: %v\n", err)
		return
	}

	t.Logf("RebuildWalletUnspent successfully.\n")
}

func TestGetBalance(t *testing.T) {

	result, err := tw.walletClient.Call("getbalance", nil)
	if err != nil {
		t.Errorf("getbalance failed unexpected error: %v\n", err)
		return
	}

	t.Logf("getbalance: %s \n", result.String())

}

func TestDumpprivkey(t *testing.T) {
	err := tw.UnlockWallet("jinxin", 10)
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
		return
	}

	request := []interface{}{
		"Dsnz4MPERzcPjVWoA6iyGVJKmGEh1ZeBkKi",
	}

	result, err := tw.walletClient.Call("dumpprivkey", request)
	if err != nil {
		t.Errorf("dumpprivkey failed unexpected error: %v\n", err)
		return
	}

	t.Logf("privkey: %s \n", result.String())
}

func TestGetreive(t *testing.T) {
	err := tw.UnlockWallet("jinxin", 10)
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
		return
	}

	request := []interface{}{
		"Dsnz4MPERzcPjVWoA6iyGVJKmGEh1ZeBkKi",
		2,
	}

	result, err := tw.walletClient.Call("getreceivedbyaddress", request)
	if err != nil {
		t.Errorf("getreceivedbyaddress failed unexpected error: %v\n", err)
		return
	}

	t.Logf("getreceivedbyaddress: %s \n", result.String())
}

func TestListUnspentFromLocalDB(t *testing.T) {
	utxos, err := tw.ListUnspentFromLocalDB("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW")
	if err != nil {
		t.Errorf("ListUnspentFromLocalDB failed unexpected error: %v\n", err)
		return
	}
	t.Logf("ListUnspentFromLocalDB totalCount = %d\n", len(utxos))
	total := decimal.New(0, 0)
	for _, u := range utxos {
		amount, _ := decimal.NewFromString(u.Amount)
		total = total.Add(amount)
		t.Logf("ListUnspentFromLocalDB %v: %s = %s\n", u.HDAddress, u.AccountID, u.Amount)
	}
	t.Logf("ListUnspentFromLocalDB total = %s\n", total.StringFixed(8))
}

func TestBuildTransaction(t *testing.T) {
	walletID := "W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW"
	utxos, err := tw.ListUnspentFromLocalDB(walletID)
	if err != nil {
		t.Errorf("BuildTransaction failed unexpected error: %v\n", err)
		return
	}

	txRaw, _, err := tw.BuildTransaction(utxos, []string{"TsiTCM9KqDPTJLt6iVBV2FCtPKzAgAtZmQG"}, "TsjkXU58hAxA8w24tZZyjdPLHVSTMeeesd6", []decimal.Decimal{decimal.NewFromFloat(0.2)}, decimal.NewFromFloat(0.001))
	if err != nil {
		t.Errorf("BuildTransaction failed unexpected error: %v\n", err)
		return
	}

	t.Logf("BuildTransaction txRaw = %s\n", txRaw)

	//hex, err := SignRawTransaction(txRaw, walletID, "1234qwer", utxos)
	//if err != nil {
	//	t.Errorf("BuildTransaction failed unexpected error: %v\n", err)
	//	return
	//}
	//
	//t.Logf("BuildTransaction signHex = %s\n", hex)
}

func TestEstimateFee(t *testing.T) {
	feeRate, err := tw.EstimateFeeRate()
	if err != nil {
		t.Errorf("EstimateFeeRate failed unexpected error: %v\n", err)
		return
	}

	t.Logf("EstimateFee feeRate = %s\n", feeRate.StringFixed(8))
	fees, _ := tw.EstimateFee(1, 2, feeRate)
	t.Logf("EstimateFee fees = %s\n", fees.StringFixed(8))
}

func TestSendTransaction(t *testing.T) {

	sends := []string{
		"Dsnz4MPERzcPjVWoA6iyGVJKmGEh1ZeBkKi",
	}

	tw.RebuildWalletUnspent("W7ANJPR91CjFhA648fSWaPAPWWWUGY4ffQ")

	for _, to := range sends {

		txIDs, err := tw.SendTransaction("W7ANJPR91CjFhA648fSWaPAPWWWUGY4ffQ", to, decimal.NewFromFloat(0.8), "jinxin", true)

		if err != nil {
			t.Errorf("SendTransaction failed unexpected error: %v\n", err)
			return
		}

		t.Logf("SendTransaction txid = %v\n", txIDs)

	}

}

func TestSendBatchTransaction(t *testing.T) {

	sends := []string{
		"TsS18cTdw81WnH8z2JQ89WoiGs6ReWXiPwJ",
		"TsfZzvj4scfJhi69QwsYAyYCKp5hnmpnpxk",
		"TshP8pvSba5aZfEk2rq8wmjpeRqsrb5KJNU",
		//"n1t8xJxkHuXsnaCD4hxPZrJRGYi6yQ83uC",
	}

	amounts := []decimal.Decimal{
		decimal.NewFromFloat(03.3),
		decimal.NewFromFloat(2.03),
		decimal.NewFromFloat(6.04),
	}

	tw.RebuildWalletUnspent("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW")

	txID, err := tw.SendBatchTransaction("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW", sends, amounts, "jinxin")

	if err != nil {
		t.Errorf("TestSendBatchTransaction failed unexpected error: %v\n", err)
		return
	}

	t.Logf("SendTransaction txid = %v\n", txID)

}

func TestMath(t *testing.T) {
	piece := int64(math.Ceil(float64(67) / float64(30)))

	t.Logf("ceil = %d", piece)
}

func TestPrintConfig(t *testing.T) {
	tw.config.printConfig()
}

func TestRestoreWallet(t *testing.T) {
	keyFile := "/myspace/workplace/go-workspace/projects/bin/data/btc/key/MacOS-W9JyC464XAZEJgdiAZxUXbPpsZZ2JeAujV.key"
	dbFile := "/myspace/workplace/go-workspace/projects/bin/data/btc/db/MacOS-W9JyC464XAZEJgdiAZxUXbPpsZZ2JeAujV.db"
	datFile := "/myspace/workplace/go-workspace/projects/bin/testdatfile/wallet.dat"
	tw.loadConfig()
	err := tw.RestoreWallet(keyFile, dbFile, datFile, "1234qwer")
	if err != nil {
		t.Errorf("RestoreWallet failed unexpected error: %v\n", err)
	}

}

func TestWalletManager_CreateNewChangeAddress(t *testing.T) {
	address, err := tw.CreateNewChangeAddress("W1xmuWQv2qCubnjS22zaVTyJ3XZ5JyNjnW")
	if err != nil {
		t.Errorf("CreateNewChangeAddress failed unexpected error: %v\n", err)
		return
	}
	t.Logf("address: %v", address)
}

func TestStopNode(t *testing.T) {
	err := tw.stopNode()
	if err != nil {
		t.Errorf("stopNode failed unexpected error: %v\n", err)
	}
}

func TestStartNode(t *testing.T) {
	err := tw.startNode()
	if err != nil {
		t.Errorf("startNode failed unexpected error: %v\n", err)
	}
}