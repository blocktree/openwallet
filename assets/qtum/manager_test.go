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

package qtum

import (
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/codeskyblue/go-sh"
	"github.com/shopspring/decimal"
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

	tw.config.serverAPI = "http://120.78.220.105:3889"
	tw.config.rpcUser = "test"
	tw.config.rpcPassword = "test1234"
	token := basicAuth(tw.config.rpcUser, tw.config.rpcPassword)
	tw.walletClient = NewClient(tw.config.serverAPI, token, true)
}

func TestImportPrivKey(t *testing.T) {

	tests := []struct {
		seed []byte
		name string
		tag  string
	}{
		{
			seed: tw.generateSeed(),
			name: "Chance",
			tag:  "first",
		},
		{
			seed: tw.generateSeed(),
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
		err = tw.UnlockWallet("1234qwer", 120)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
		}

		//导入私钥
		err = tw.ImportPrivKey(wif.String(), test.name)
		if err != nil {
			t.Errorf("ImportPrivKey[%d] failed unexpected error: %v\n", i, err)
		} else {
			t.Logf("ImportPrivKey[%d] success \n", i)
		}
	}
}

func TestGetCoreWalletinfo(t *testing.T) {
	tw.GetCoreWalletinfo()
}

func TestKeyPoolRefill(t *testing.T) {

	//解锁钱包
	err := tw.UnlockWallet("1234qwer", 120)
	if err != nil {
		t.Errorf("KeyPoolRefill failed unexpected error: %v\n", err)
	}

	err = tw.KeyPoolRefill(10000)
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
		{
			account: "Chance",
			tag:     "normal",
		},
	}

	for i, test := range tests {

		a, err := tw.CreateReceiverAddress(test.account)
		if err != nil {
			t.Errorf("CreateReceiverAddress[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("CreateReceiverAddress[%d] address = %v", i, a)
		}

	}

}

func TestGetAddressesByAccount(t *testing.T) {
	addresses, err := tw.GetAddressesByAccount("WJjFgnZucp86LR3s18AbjxT3ju9csXduff")
	if err != nil {
		t.Errorf("GetAddressesByAccount failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesByAccount address[%d] = %s\n", i, a)
	}
}

func TestCreateBatchAddress(t *testing.T) {
	_, _, err := tw.CreateBatchAddress("WJjFgnZucp86LR3s18AbjxT3ju9csXduff", "1234qwer", 10)
	if err != nil {
		t.Errorf("CreateBatchAddress failed unexpected error: %v\n", err)
		return
	}
}

func TestGenerateQtumAddress(t *testing.T){
	address,err :=tw.GenQtumAddress()
	if err != nil {
		t.Errorf("GenQtumAddress failed unexpected error: %v\n", err)
		return
	}else {
		t.Logf("Generate Qtum address successfully, address is: %s",address)
	}
}

func TestEncryptWallet(t *testing.T) {
	err := tw.EncryptWallet("11111111")
	if err != nil {
		t.Errorf("EncryptWallet failed unexpected error: %v\n", err)
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

func TestCreateNewWallet(t *testing.T) {
	_, _, err := tw.CreateNewWallet("sam2", "1234qwer")
	if err != nil {
		t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestGetWalletKeys(t *testing.T) {
	wallets, err := tw.GetWallets()
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
			name: "WJjFgnZucp86LR3s18AbjxT3ju9csXduff",
			tag:  "first",
		},
		{
			name: "QTUM02",
			tag:  "second",
		},
		{
			name: "W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW",
			tag:  "all",
		},
		{
			name: "ZB",
			tag:  "account not exist",
		},
	}

	for i, test := range tests {
		balance := tw.GetWalletBalance(test.name)
		t.Logf("GetWalletBalance[%d] %s balance = %s \n", i, test.name, balance)
	}

}

//有问题the extended key at this index is invalid
func TestCreateNewPrivateKey(t *testing.T) {

	tests := []struct {
		name     string
		password string
		tag      string
	}{
		{
			name:     "WJjFgnZucp86LR3s18AbjxT3ju9csXduff",
			password: "1234qwer",
			tag:      "wallet not exist",
		},
		{
			name:     "QTUM02",
			password: "1234qwer",
			tag:      "normal",
		},
		{
			name:     "W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW",
			password: "1234qwer",
			tag:      "wrong password",
		},
	}

	for i, test := range tests {
		w, err := tw.GetWalletInfo(test.name)
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
		wif, a, err := tw.CreateNewPrivateKey(key, uint64(timestamp), 0)
		if err != nil {
			t.Errorf("CreateNewPrivateKey[%d] failed unexpected error: %v\n", i, err)
			continue
		}

		t.Logf("CreateNewPrivateKey[%d] wif = %v \n", i, wif)
		t.Logf("CreateNewPrivateKey[%d] address = %v \n", i, a)
	}
}

func TestGetWalleInfo(t *testing.T) {
	w, err := tw.GetWalletInfo("W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW")
	if err != nil {
		t.Errorf("GetWalletInfo failed unexpected error: %v\n", err)
		return
	}

	t.Logf("GetWalletInfo wallet = %v \n", w)
}

//func TestCreateBatchPrivateKey(t *testing.T) {
//
//	w, err := tw.GetWalletInfo("Zhiquan Test")
//	if err != nil {
//		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//
//	key, err := w.HDKey("1234qwer")
//	if err != nil {
//		t.Errorf("CreateBatchPrivateKey failed unexpected error: %v\n", err)
//		return
//	}
//
//	wifs, err := tw.CreateBatchPrivateKey(key, 10000)
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

	backupFile, err := tw.BackupWallet("W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW")
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

func TestDumpWallet(t *testing.T) {
	tw.UnlockWallet("1234qwer", 120)
	file := filepath.Join(".", "openwallet", "")
	err := tw.DumpWallet(file)
	if err != nil {
		t.Errorf("DumpWallet failed unexpected error: %v\n", err)
	} else {
		t.Errorf("DumpWallet filePath: %v\n", file)
	}
}

func TestGOSH(t *testing.T) {
	//text, err := sh.Command("go", "env").Output()
	//text, err := sh.Command("wmd", "version").Output()
	text, err := sh.Command("wmd", "config", "see", "-s", "qtum").Output()
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

//还未有币，不能测试
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
	addresses, err := tw.GetAddressesFromLocalDB("WJjFgnZucp86LR3s18AbjxT3ju9csXduff", 0, -1)
	if err != nil {
		t.Errorf("GetAddressesFromLocalDB failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesFromLocalDB address[%d] = %v\n", i, a)
	}
}

//还未有币，不能测试
func TestRebuildWalletUnspent(t *testing.T) {

	err := tw.RebuildWalletUnspent("WJjFgnZucp86LR3s18AbjxT3ju9csXduff")
	if err != nil {
		t.Errorf("RebuildWalletUnspent failed unexpected error: %v\n", err)
		return
	}

	t.Logf("RebuildWalletUnspent successfully.\n")
}

//还未有币，不能测试
func TestListUnspentFromLocalDB(t *testing.T) {
	utxos, err := tw.ListUnspentFromLocalDB("WJjFgnZucp86LR3s18AbjxT3ju9csXduff")
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

//还未有币，不能测试
func TestBuildTransaction(t *testing.T) {
	walletID := "WJjFgnZucp86LR3s18AbjxT3ju9csXduff"
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

//还未有币，不能测试
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

//还未有币，不能测试
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

func TestGetNetworkInfo(t *testing.T) {
	tw.GetNetworkInfo()
}

func TestPrintConfig(t *testing.T) {
	tw.config.printConfig()
}

//密码错误？
func TestRestoreWallet(t *testing.T) {
	keyFile := "QTUM03-W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW.key"
	dbFile := "QTUM03-W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW.db"
	datFile := "/data/qtum/backup/QTUM03-W2JgPVMS2jEQZ7yUkfHEa4D1ST4NccLCAW-20180821160056/wallet.dat"
	tw.loadConfig()
	err := tw.RestoreWallet(keyFile, dbFile, datFile, "1234qwer")
	if err != nil {
		t.Errorf("RestoreWallet failed unexpected error: %v\n", err)
	}

}
