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

import "testing"

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
	_, err := BackupWallet("/home/chbtc/openwallet/data/sc/backup/")

	if err != nil {
		t.Errorf("BackupWallet failed unexpected error: %v\n", err)
	} else {
		t.Logf("BackupWallet successfully\n")
	}
}

func TestUnlockWallet(t *testing.T) {
	err := UnlockWallet("1234567890")
	if err != nil {
		t.Errorf("UnlockWallet failed unexpected error: %v\n", err)
	} else {
		t.Logf("UnlockWallet successfully\n")
	}
}

func TestCreateNewWallet(t *testing.T) {
	password := "1234567890"
	seed, err := CreateNewWallet(password, true)
	if err != nil {
		t.Errorf("CreateNewWallet failed unexpected error: %v\n", err)
	} else {
		t.Logf("CreateNewWallet seed = %s\n", seed)
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

func TestCreateBatchAddress(t *testing.T) {
	CreateBatchAddress(1000)
}
