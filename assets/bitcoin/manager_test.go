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
	"testing"
	"time"
	"path/filepath"
	"github.com/codeskyblue/go-sh"
)

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
	addresses, err := GetAddressesByAccount("John Test")
	if err != nil {
		t.Errorf("GetAddressesByAccount failed unexpected error: %v\n", err)
		return
	}

	for i, a := range addresses {
		t.Logf("GetAddressesByAccount address[%d] = %s\n", i, a)
	}
}

func TestCreateBatchAddress(t *testing.T) {
	CreateBatchAddress("John Test", "1234qwer", 10000)
}

func TestEncryptWallet(t *testing.T) {
	err := EncryptWallet("11111111")
	if err != nil {
		t.Errorf("EncryptWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestUnlockWallet(t *testing.T) {
	err := UnlockWallet("1234qwer",1)
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
		return
	}
}

func TestCreateNewWallet(t *testing.T) {
	err := CreateNewWallet("John Test", "1234qwer")
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
			name: "Chance",
			tag:  "first",
		},
		{
			name: "Zhiquan Test",
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
		name string
		password string
		tag  string
	}{
		{
			name: "Chance",
			password: "1234qwer",
			tag:  "wallet not exist",
		},
		{
			name: "Zhiquan Test",
			password: "1234qwer",
			tag:  "normal",
		},
		{
			name: "Zhiquan Test",
			password: "121212121212",
			tag:  "wrong password",
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

func TestImportMulti(t *testing.T) {

	addresses := []string{
		"1CoRcQGjPEyWmB1ZyG6CEDN3SaMsaD3ERa",
		"1ESGCsXkNr3h5wvWScdCpVHu2GP3KJtCdV",
	}

	keys := []string{
		"L5k8VYSvuZxC5FCczGVC8MmnKKix3Mcs6t185eUJVKTzZb1f6bsX",
		"L3RVDjPVBSc7DD4WtmzbHkAHJW4kDbyXbw4vBppZ4DRtPt5u8Naf",
	}

	UnlockWallet("1234qwer", 120)
	failed, err := ImportMulti(addresses, keys, "Zhiquan Test")
	if err != nil {
		t.Errorf("ImportMulti failed unexpected error: %v\n", err)
	} else {
		t.Errorf("ImportMulti result: %v\n", failed)
	}
}

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