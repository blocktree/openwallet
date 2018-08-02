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

package hypercash

import (
	"fmt"
	"github.com/HcashOrg/hcrpcclient"
	"github.com/codeskyblue/go-sh"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"path/filepath"
	"testing"
	"time"
)

var (
	tw *WalletManager
)

func init() {

	tw = NewWalletManager()

	tw.config.serverAPI = "https://192.168.2.193:10128"
	tw.config.rpcUser = "wallethcd"
	tw.config.rpcPassword = "wallethcdpw"
	token := basicAuth(tw.config.rpcUser, tw.config.rpcPassword)
	tw.client = NewClient(tw.config.serverAPI, token, true)
}

func TestWalletManager_GetBlockHeight(t *testing.T) {

	height, err := tw.GetBlockHeight()
	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHeight height = %d \n", height)

}

func TestCreateNewWallet(t *testing.T) {
	_, _, err := tw.CreateNewWallet("ZBG2", "123456")
	if err != nil {
		t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestGetAddressesByAccount(t *testing.T) {
	addresses, err := tw.GetAddressesByAccount("WKboeMWpu9cRFp3smkciS6qVY8TECRTxzJ")
	if err != nil {
		t.Errorf("GetAddressesByAccount failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesByAccount address[%d] = %s\n", i, a)
	}
}

func TestCreateBatchAddress(t *testing.T) {
	_, _, err := tw.CreateBatchAddress("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", "1234qwer", 100)
	if err != nil {
		t.Errorf("CreateBatchAddress failed unexpected error: %v\n", err)
		return
	}
}

func TestUnlockWallet(t *testing.T) {
	err := tw.UnlockWallet("1234qwer", 1)
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
			name: "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ",
			tag:  "first",
		},
		{
			name: "Wallet Test",
			tag:  "second",
		},
		{
			name: "*",
			tag:  "all",
		},
		{
			name: "llllll",
			tag:  "account not exist",
		},
	}

	for i, test := range tests {
		balance := tw.GetWalletBalance(test.name)
		t.Logf("GetWalletBalance[%d] %s balance = %s \n", i, test.name, balance)
	}

}

func TestGetWalleList(t *testing.T) {
	wallets, err := tw.GetWalletList()
	if err != nil {
		t.Errorf("GetWalleList failed unexpected error: %v\n", err)
		return
	}

	for i, w := range wallets {
		t.Logf("GetWalleList wallet[%d] = %v", i, w)
	}
}

func TestCreateBatchPrivateKey(t *testing.T) {

	w, err := tw.GetWalletInfo("WKD6QUMLyv93qBBdnURokKCrQKHeTQYeVu")
	if err != nil {
		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
		return
	}

	key, err := w.HDKey("123")
	if err != nil {
		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
		return
	}

	wifs, err := tw.CreateBatchPrivateKey(key, 5)
	if err != nil {
		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
		return
	}

	for i, wif := range wifs {
		t.Logf("CreateBatchPrivateKey[%d] wif = %v \n", i, wif)
	}

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

	backupFile, err := tw.BackupWallet("W9JyC464XAZEJgdiAZxUXbPpsZZ2JeAujV")
	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("BackupWallet filePath: %v\n", backupFile)
	}
}

func TestBackupWalletData(t *testing.T) {
	tw.config.walletDataPath = "/home/www/btc/testdata/testnet3/"
	tmpWalletDat := fmt.Sprintf("tmp-walllet-%d.dat", time.Now().Unix())
	backupFile := filepath.Join(tw.config.walletDataPath, tmpWalletDat)
	err := tw.BackupWalletData(backupFile)
	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("BackupWallet filePath: %v\n", backupFile)
	}
}

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
		t.Errorf("GetBlockChainInfo info: %v\n", b)
	}
}

func TestListUnspent(t *testing.T) {
	utxos, err := tw.ListUnspent(0)
	if err != nil {
		t.Errorf("ListUnspent failed unexpected error: %v\n", err)
		return
	}

	for _, u := range utxos {
		t.Logf("ListUnspent %s: %s = %s\n", u.Address, u.AccountID, u.Amount)
	}
}

func TestGetAddressesFromLocalDB(t *testing.T) {
	addresses, err := tw.GetAddressesFromLocalDB("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", 0, -1)
	if err != nil {
		t.Errorf("GetAddressesFromLocalDB failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesFromLocalDB address[%d] = %v\n", i, a)
	}
}

func TestRebuildWalletUnspent(t *testing.T) {

	err := tw.RebuildWalletUnspent("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")
	if err != nil {
		t.Errorf("RebuildWalletUnspent failed unexpected error: %v\n", err)
		return
	}

	t.Logf("RebuildWalletUnspent successfully.\n")
}

func TestListUnspentFromLocalDB(t *testing.T) {
	utxos, err := tw.ListUnspentFromLocalDB("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")
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
	walletID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	utxos, err := tw.ListUnspentFromLocalDB(walletID)
	if err != nil {
		t.Errorf("BuildTransaction failed unexpected error: %v\n", err)
		return
	}

	txRaw, _, err := tw.BuildTransaction(utxos, []string{"mrThNMQ6bMf1YNPjBj9jYXmYYzw1Rt8GFU"}, "n33cHpEc9qAvECM9pFgabZ6ktJimLSeWdy", []decimal.Decimal{decimal.NewFromFloat(0.2)}, decimal.NewFromFloat(0.00002))
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
	feeRate, _ := tw.EstimateFeeRate()
	t.Logf("EstimateFee feeRate = %s\n", feeRate.StringFixed(8))
	fees, _ := tw.EstimateFee(10, 2, feeRate)
	t.Logf("EstimateFee fees = %s\n", fees.StringFixed(8))
}

func TestSendTransaction(t *testing.T) {

	sends := []string{
		"mpkUFiXonEZriywHUhig6PTDQXKzT6S5in",
	}

	tw.RebuildWalletUnspent("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")

	for _, to := range sends {

		txIDs, err := tw.SendTransaction("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", to, decimal.NewFromFloat(0.02), "1234qwer", false)

		if err != nil {
			t.Errorf("SendTransaction failed unexpected error: %v\n", err)
			return
		}

		t.Logf("SendTransaction txid = %v\n", txIDs)

	}

}

func TestSendBatchTransaction(t *testing.T) {

	sends := []string{
		"mfYksPvrRS9Xb28uVUiQPJTnc92TBEP1P6",
		//"mfXVvSn76et4GcNsyphRKxbVwZ6BaexYLG",
		//"miqpBeCQnYraAV73TeTrCtDsFK5ebKU7P9",
		//"n1t8xJxkHuXsnaCD4hxPZrJRGYi6yQ83uC",
	}

	amounts := []decimal.Decimal{
		decimal.NewFromFloat(0.3),
		//decimal.NewFromFloat(0.03),
		//decimal.NewFromFloat(0.04),
	}

	tw.RebuildWalletUnspent("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")

	txID, err := tw.SendBatchTransaction("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", sends, amounts, "1234qwer")

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

func TestGetBlockHeight(t *testing.T) {
	// Load the certificate for the TLS connection which is automatically
	// generated by hcd when it starts the RPC server and doesn't already
	// have one.
	//hcdHomeDir := hcutil.AppDataDir("hcd", false)
	//certs, err := ioutil.ReadFile(filepath.Join(hcdHomeDir, "rpc.cert"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Create a new RPC client using websockets.  Since this example is
	//// not long-lived, the connection will be closed as soon as the program
	//// exits.
	//connCfg := &hcrpcclient.ConnConfig{
	//	Host:         "localhost:9678",
	//	Endpoint:     "ws",
	//	User:         "USER",
	//	Pass:         "PASSWORD",
	//	Certificates: certs,
	//}

	connCfg := &hcrpcclient.ConnConfig{
		Host:       "192.168.2.193:10127",
		Endpoint:   "ws",
		User:       "wallethcd",
		Pass:       "wallethcdpw",
		DisableTLS: true,
	}

	client, err := hcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Query the RPC server for the current block count and display it.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)
}
