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

package bitcoin

import (
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/codeskyblue/go-sh"
	"github.com/shopspring/decimal"
	"path/filepath"
	"testing"
	"time"
	"math"
)

func init() {

	serverAPI = "http://192.168.2.192:10000"
	rpcUser = "wallet"
	rpcPassword = "walletPassword2017"
	token := basicAuth(rpcUser, rpcPassword)
	client = &Client{
		BaseURL:     serverAPI,
		AccessToken: token,
		Debug:       false,
	}
}

func TestImportPrivKey(t *testing.T) {

	tests := []struct {
		seed []byte
		name string
		tag  string
	}{
		{
			seed: generateSeed(),
			name: "Chance",
			tag:  "first",
		},
		{
			seed: generateSeed(),
			name: "Chance",
			tag:  "second",
		},
	}

	for i, test := range tests {
		key, err := keystore.NewHDKey(test.seed, test.name, "m/44'/88'")
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		privateKey, err := key.MasterKey.ECPrivKey()
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		publicKey, err := key.MasterKey.ECPubKey()
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		wif, err := btcutil.NewWIF(privateKey, &chaincfg.MainNetParams, true)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		t.Logf("Privatekey wif[%d] = %s\n", i, wif.String())

		address, err := btcutil.NewAddressPubKey(publicKey.SerializeCompressed(), &chaincfg.MainNetParams)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		t.Logf("Privatekey address[%d] = %s\n", i, address.EncodeAddress())

		//解锁钱包
		err = UnlockWallet("1234qwer", 120)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
		}

		//导入私钥
		err = ImportPrivKey(wif.String(), test.name)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
		} else {
			t.Logf("ImportPrivKey[%d] success \n", i)
		}
	}

}

func TestGetCoreWalletinfo(t *testing.T) {
	GetCoreWalletinfo()
}

func TestKeyPoolRefill(t *testing.T) {

	//解锁钱包
	err := UnlockWallet("1234qwer", 120)
	if err != nil {
		t.Errorf("KeyPoolRefill failed unexpected error: %v\n", err)
	}

	err = KeyPoolRefill(10000)
	if err != nil {
		t.Errorf("KeyPoolRefill failed unexpected error: %v\n", err)
	}
}

func TestCreateReceiverAddress(t *testing.T) {

	tests := []struct {
		account string
		tag     string
	}{
		{
			account: "john",
			tag:     "normal",
		},
		//{
		//	account: "Chance",
		//	tag:     "normal",
		//},
	}

	for i, test := range tests {

		a, err := CreateReceiverAddress(test.account)
		if err != nil {
			t.Errorf("CreateReceiverAddress[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("CreateReceiverAddress[%d] address = %v", i, a)
		}

	}

}

func TestGetAddressesByAccount(t *testing.T) {
	addresses, err := GetAddressesByAccount("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")
	if err != nil {
		t.Errorf("GetAddressesByAccount failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesByAccount address[%d] = %s\n", i, a)
	}
}

func TestCreateBatchAddress(t *testing.T) {
	CreateBatchAddress("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", "1234qwer", 100)
}

func TestEncryptWallet(t *testing.T) {
	err := EncryptWallet("11111111")
	if err != nil {
		t.Errorf("EncryptWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestUnlockWallet(t *testing.T) {
	err := UnlockWallet("1234qwer", 1)
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestCreateNewWallet(t *testing.T) {
	err := CreateNewWallet("Rocky", "1234qwer")
	if err != nil {
		t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestGetWalletKeys(t *testing.T) {
	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		t.Errorf("GetWalletKeys failed unexpected error: %v\n", err)
		return
	}

	for i, w := range wallets {
		t.Logf("GetWalletKeys wallet[%d] = %v", i, w)
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
		balance, err := GetWalletBalance(test.name)
		if err != nil {
			t.Errorf("GetWalletBalance[%d] failed unexpected error: %v\n", i, err)
		} else {
			t.Logf("GetWalletBalance[%d] %s balance = %s \n", i, test.name, balance)
		}
	}

}

func TestGetWalleList(t *testing.T) {
	wallets, err := GetWalleList()
	if err != nil {
		t.Errorf("GetWalleList failed unexpected error: %v\n", err)
		return
	}

	for i, w := range wallets {
		t.Logf("GetWalleList wallet[%d] = %v", i, w)
	}
}

func TestCreateNewPrivateKey(t *testing.T) {

	tests := []struct {
		name     string
		password string
		tag      string
	}{
		{
			name:     "Chance",
			password: "1234qwer",
			tag:      "wallet not exist",
		},
		{
			name:     "Zhiquan Test",
			password: "1234qwer",
			tag:      "normal",
		},
		{
			name:     "Zhiquan Test",
			password: "121212121212",
			tag:      "wrong password",
		},
	}

	for i, test := range tests {
		w, err := GetWalletInfo(test.name)
		if err != nil {
			t.Errorf("CreateNewPrivateKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		key, err := w.HDKey(test.password)
		if err != nil {
			t.Errorf("CreateNewPrivateKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		timestamp := time.Now().Unix()
		t.Logf("CreateNewPrivateKey[%d] timestamp = %v \n", i, timestamp)
		wif, a, err := CreateNewPrivateKey(key, uint64(timestamp), 0)
		if err != nil {
			t.Errorf("CreateNewPrivateKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		t.Logf("CreateNewPrivateKey[%d] wif = %v \n", i, wif)
		t.Logf("CreateNewPrivateKey[%d] address = %v \n", i, a)
	}
}

func TestGetWalleInfo(t *testing.T) {
	w, err := GetWalletInfo("Zhiquan Test")
	if err != nil {
		t.Errorf("GetWalletInfo failed unexpected error: %v\n", err)
		return
	}

	t.Logf("GetWalletInfo wallet = %v \n", w)
}

func TestCreateBatchPrivateKey(t *testing.T) {

	w, err := GetWalletInfo("Zhiquan Test")
	if err != nil {
		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
		return
	}

	key, err := w.HDKey("1234qwer")
	if err != nil {
		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
		return
	}

	wifs, err := CreateBatchPrivateKey(key, 10000)
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
	UnlockWallet("1234qwer", 120)
	//backupFile := "/home/chbtc/openwallet/backup.dat"
	backupFile := filepath.Join(".", "backup.dat")
	err := BackupWallet(backupFile)
	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("BackupWallet filePath: %v\n", backupFile)
	}
}

func TestDumpWallet(t *testing.T) {
	UnlockWallet("1234qwer", 120)
	file := filepath.Join(".", "openwallet", "")
	err := DumpWallet(file)
	if err != nil {
		t.Errorf("DumpWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("DumpWallet filePath: %v\n", file)
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
	b, err := GetBlockChainInfo()
	if err != nil {
		t.Errorf("GetBlockChainInfo failed unexpected error: %v\n", err)
	} else {
		t.Errorf("GetBlockChainInfo info: %v\n", b)
	}
}

func TestListUnspent(t *testing.T) {
	utxos, err := ListUnspent(1)
	if err != nil {
		t.Errorf("ListUnspent failed unexpected error: %v\n", err)
		return
	}

	for _, u := range utxos {
		t.Logf("ListUnspent %s: %s = %s\n", u.Address, u.Account, u.Amount)
	}
}

func TestGetAddressesFromLocalDB(t *testing.T) {
	addresses, err := GetAddressesFromLocalDB("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")
	if err != nil {
		t.Errorf("GetAddressesFromLocalDB failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesFromLocalDB address[%d] = %s\n", i, a)
	}
}

func TestRebuildWalletUnspent(t *testing.T) {

	err := RebuildWalletUnspent()
	if err != nil {
		t.Errorf("RebuildWalletUnspent failed unexpected error: %v\n", err)
		return
	}

	t.Logf("RebuildWalletUnspent successfully.\n")
}

func TestListUnspentFromLocalDB(t *testing.T) {
	utxos, err := ListUnspentFromLocalDB("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ")
	if err != nil {
		t.Errorf("ListUnspentFromLocalDB failed unexpected error: %v\n", err)
		return
	}

	for _, u := range utxos {
		t.Logf("ListUnspentFromLocalDB %v: %s = %s\n", u.HDAddress, u.Account, u.Amount)
	}
}

func TestBuildTransaction(t *testing.T) {
	walletID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	utxos, err := ListUnspentFromLocalDB(walletID)
	if err != nil {
		t.Errorf("BuildTransaction failed unexpected error: %v\n", err)
		return
	}

	txRaw, _, err := BuildTransaction(utxos, "mrThNMQ6bMf1YNPjBj9jYXmYYzw1Rt8GFU", "n33cHpEc9qAvECM9pFgabZ6ktJimLSeWdy", decimal.NewFromFloat(0.2), decimal.NewFromFloat(0.00002))
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
	feeRate, _ := EstimateFeeRate()
	fees, _ := EstimateFee(10, 2, feeRate)
	t.Logf("EstimateFee fees = %s\n", fees.StringFixed(8))
}

func TestSendTransaction(t *testing.T) {

	sends := []string{
		"mzRMyCYNDTC91g7VTryFgabYBg2GDRTDfE",
		"mwJ5Jrj8tPq9D2wE4h8mVkGgAMGbQ4GCo2",
		"mmewNfURua2iCHCg87m6A1VP2awZ34GS9x",
		"n4ExgJ9dHQ6bYjkh6KPChonX99L4KJqYUq",
		"mj9nRF11ZCDN8wBAc6DtRxxBw9ehHWJJtG",
		"n2GxGUsL78JgdYe4un8g2aH1PowkgfSLXB",
		"mxL2SXgVPdyjVPkpG42Fo5Vph3UTQiH34o",
		"mqD9ouMpTnjiWBD1vrNVasBwNfKoekfck5",
		"n3nPyV6JrxXQrP6TYmxJCQzUzcyJAN3NL8",
		"mzJeoPF1r8QozHJYc2JsohchpJWgUdUhmh",
		"n4aU52rgZWEbYFekrUjJPg5oTfcXC4Ct69",
		"mt3kHYS2piWaJhRKzA8jfTrymzw7eDkGys",
		"n4GvMSbFLvPcuSSopRvgDY5ysr7Ui4XGfk",
		"mkDW74UpsJxsJk6qPyV1avWCJTRPVYYBqT",
		"mw1DWAVCyLN2FzbCTHZ679d6TuTRxBBd63",
		"mqmZx9Apw81p7kvGHnemxesKJTTVez2VYT",
		"mk42ZbczUTHBvm3RrzitZgnBXRYDMyX6JY",
		"mysdRzsFyJGFSWJwFt8evcds2GN5CEv3Np",
		"mrmVvPmVpmdvnyPpqgnWLsstV2PEmBZX94",
		"mzBqw9Hph42kTTn9vPhSCNwUahfmqPMkAt",
		"msZ5WN6FLsBAFDG1hTeRzNjZeMSgx15GdV",
		"mvNqtLTSKkshqfaJPkoPVgZ6gvuuxgqTEr",
		"mkPFETnfrEtTCyjjs5ZU5ZCo4VCJJK1Jqy",
	}

	for _, to := range sends {

		RebuildWalletUnspent()

		txIDs, err := SendTransaction("W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ", to, decimal.NewFromFloat(0.02), "1234qwer", false)

		if err != nil {
			t.Errorf("SendTransaction failed unexpected error: %v\n", err)
			return
		}

		t.Logf("SendTransaction txid = %v\n", txIDs)

	}


}

func TestMath(t *testing.T) {
	piece := int64(math.Ceil(float64(67) / float64(30)))

	t.Logf("ceil = %d", piece)
}