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

package sia

import (
	"testing"
	"log"
	"github.com/bytom/common"
	"github.com/NebulousLabs/entropy-mnemonics"
	"strings"
)

func init() {
	//serverAPI = "http://192.168.2.224:10056"
	serverAPI = "http://192.168.2.193:10051"
	client = &Client{
		BaseURL: serverAPI,
		Debug:   true,
		Auth:    "123",
	}
}

func TestGetWalletInfo(t *testing.T) {

	wallets, err := GetWalletInfo()
	if err != nil {
		t.Errorf("GetWalletInfo failed unexpected error: %v", err)
	} else {
		for i, w := range wallets {
			t.Logf("GetWalletInfo wallet[%d] = %v", i, w)
		}
	}
}

func TestBackupWallet(t *testing.T) {

	backupFile, err := BackupWallet()
	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Logf("BackupWallet filePath: %v\n", backupFile)
	}
}

func TestUnlockWallet(t *testing.T) {
	err := UnlockWallet("1234567890abc")
	t.Logf("Unlock information: %v\n", err)
}

//慎用新建钱包，会替换旧的钱包（要先备份旧钱包）
func TestCreateNewWallet(t *testing.T) {
	password := "1234567890abc"
	wallets, err := GetWalletInfo()
	if !wallets[0].Encrypted || err != nil {
		seed, err := CreateNewWallet(password,false)
		if err != nil {
			t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		} else {
			t.Logf("CreateNewWallet seed = %s\n", seed)
		}
	}else{
		seed, err := CreateNewWallet(password,true)
		if err != nil {
			t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
		} else {
			t.Logf("CreateNewWallet seed = %s\n", seed)
		}
	}
}

func TestGetAddressInfo(t *testing.T) {

	addrs, err := GetAddressInfo()
	if err != nil {
		t.Errorf("GetAddressInfo failed unexpected error: %v", err)
		return
	}
	for j, a := range addrs {
		t.Logf("GetAddressInfo address[%d]  = %v", j, a)
	}
}

func TestGetConsensus(t *testing.T) {
	GetConsensus()
}

//先生成一个地址测试一下钱包有没有问题，实际开发是应该减去一个数量
func TestCreateBatchAddress(t *testing.T) {
	_, err := CreateAddress()
	if err != nil {
		t.Errorf("CreateBatchAddress failed unexpected error: %v", err)
		return
	}
	CreateBatchAddress(10)
}

func TestCreateAddress(t *testing.T) {

	address, err := CreateAddress()
	if err != nil {
		t.Errorf("CreateAddress failed unexpected error: %v", err)
		return
	}
	t.Logf("CreateAddress address:[%s]", address)
}

func TestSendTransaction(t *testing.T) {
	txID, err := SendTransaction("1000000000000000000000000", "70e848d92b8d729052d2d614446df07fed787d022a989d6106a5549816680f6d85aee6044f86")
	if err != nil {
		t.Errorf("SendTransaction failed unexpected error: %v", err)
		return
	}
	t.Logf("SendTransaction success, Transaction ID:%s\n",txID)
}

func TestSummaryWallets(t *testing.T) {
	SummaryWallets()
}

func TestRestoreWallet(t *testing.T) {

	dbFile := "/myspace/workplace/go-workspace/projects/bin/data/sc/backup/wallet-backup-20180713122219/wallet.db"
	//dbFile := "C:/Users/Administrator/AppData/Roaming/Sia-UI/wallet_company/wallet.db"
	//loadConfig()
	walletDataPath = "/Users/maizhiquan/Library/Application Support/Sia-UI/sia/wallet"
	err := RestoreWallet(dbFile)
	if err != nil {
		t.Errorf("RestoreWallet failed unexpected error: %v\n", err)
	}
}

func TestMnemonicToSeed(t *testing.T) {
	m := "cajun rarest unusual swept dummy stellar sneeze onward kiosk racetrack ostrich nostril gutter recipe sickness emotion palace pests speedy axes cycling waffle bicycle western yearbook lexicon odds onslaught after"
	log.Printf("Original = %s", m)
	seed, err := mnemonics.FromString(m, mnemonics.English)
	if err != nil {
		t.Errorf("MnemonicToSeed failed unexpected error: %v\n", err)
		return
	}
	log.Printf("seed = %s", common.Bytes2Hex(seed))
	new, err := mnemonics.ToPhrase(seed, mnemonics.English)
	if err != nil {
		t.Errorf("MnemonicToSeed failed unexpected error: %v\n", err)
		return
	}
	log.Printf("new = %s", new.String())
	isEqualed := strings.EqualFold(new.String(), m)
	log.Printf("equal: %v", isEqualed)
}