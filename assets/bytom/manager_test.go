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

package bytom

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestCreateNewWallet(t *testing.T) {
	tests := []struct {
		alias    string
		password string
		tag      string
	}{
		{
			alias:    "hello",
			password: "1234567",
			tag:      "normal",
		},
		{
			alias:    "hello",
			password: "",
			tag:      "no password",
		},
		{
			alias:    "",
			password: "",
			tag:      "no alias password",
		},
	}

	for i, test := range tests {

		w, err := CreateNewWallet(test.alias, test.password)
		if err != nil {
			t.Errorf("CreateNewWallet[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("CreateNewWallet[%d] wallet = %v", i, w)
		}

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

func TestCreateNormalAccount(t *testing.T) {

	tests := []struct {
		pubkey string
		alias  string
		tag    string
	}{
		{
			pubkey: "3462b88883acc1e430dd5a6327dc8c4aca0d6d0ce087d43d98b84216e51e0885cc6e35ed92eefc244e587ae51a87fe082a0facea0e641038c80940d502520b50",
			alias:  "john",
			tag:    "normal",
		},
		{
			pubkey: "3462b88883acc1e430dd5a6327dc8c4aca0d6d0ce087d43d98b84216e51e0885cc6e35ed92eefc244e587ae51a87fe082a0facea0e641038c80940d502520b50",
			alias:  "john",
			tag:    "same alias",
		},
		{
			pubkey: "3462b88883acc1e430dd5a6327dc8c4aca0d6d0ce087d43d98b84216e51e0885cc6e35ed92eefc244e587ae51a87fe082a0facea0e641038c80940d502520b50",
			alias:  "",
			tag:    "no alias",
		},
		{
			pubkey: "",
			alias:  "",
			tag:    "no pubkey",
		},
	}

	for i, test := range tests {

		a, err := CreateNormalAccount(test.pubkey, test.alias)
		if err != nil {
			t.Errorf("CreateNormalAccount[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("CreateNormalAccount[%d] account = %v", i, a)
		}
		//3462b88883acc1e430dd5a6327dc8c4aca0d6d0ce087d43d98b84216e51e0885cc6e35ed92eefc244e587ae51a87fe082a0facea0e641038c80940d502520b50
	}

}

func TestGetAccountInfo(t *testing.T) {
	accounts, err := GetAccountInfo()
	if err != nil {
		t.Errorf("GetAccountInfo failed unexpected error: %v", err)
	} else {
		for i, a := range accounts {
			t.Logf("GetAccountInfo account[%d] = %v", i, a)
		}

	}
}

func TestCreateReceiverAddress(t *testing.T) {

	tests := []struct {
		accountAlias string
		accountID    string
		tag          string
	}{
		{
			accountAlias: "john",
			accountID:    "0E6MHCTMG0A04",
			tag:          "normal",
		},
		{
			accountAlias: "john",
			accountID:    "0E6MHCTMG0A02",
			tag:          "wrong id",
		},
		{
			accountAlias: "john222",
			accountID:    "0E6MHCTMG0A04",
			tag:          "wrong alias",
		},
		{
			accountAlias: "",
			accountID:    "",
			tag:          "empty",
		},
	}

	for i, test := range tests {

		a, err := CreateReceiverAddress(test.accountAlias, test.accountID)
		if err != nil {
			t.Errorf("CreateReceiverAddress[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("CreateReceiverAddress[%d] address = %v", i, a)
		}
		//3462b88883acc1e430dd5a6327dc8c4aca0d6d0ce087d43d98b84216e51e0885cc6e35ed92eefc244e587ae51a87fe082a0facea0e641038c80940d502520b50
	}

}

func TestGetAddressInfo(t *testing.T) {
	addresses, err := GetAddressInfo("john", "0E6MHCTMG0A04")
	if err != nil {
		t.Errorf("GetAddressInfo failed unexpected error: %v", err)
	} else {
		for i, a := range addresses {
			t.Logf("GetAddressInfo address[%d] = %v", i, a)
		}

	}
}

func TestCreateBatchAddress(t *testing.T) {
	CreateBatchAddress("john", "0E6MHCTMG0A04", 100000)
}

func TestGetAccountBalance(t *testing.T) {
	accounts, err := GetAccountBalance("", assetsID_btm)
	if err != nil {
		t.Errorf("GetAccountBalance failed unexpected error: %v", err)
	} else {
		for i, a := range accounts {
			t.Logf("GetAccountBalance account[%d] = %v", i, a)
		}

	}
}

func TestBuildTransaction(t *testing.T) {

	amount := decimal.NewFromFloat(0.1).Mul(coinDecimal)

	tx, err := BuildTransaction("0E6MV60100A08",
		"bm1qjf6v463sj3w04zyk8vq0aefjrzug0jtwz62mz0",
		assetsID_btm, uint64(amount.IntPart()), 0)
	if err != nil {
		t.Errorf("BuildTransaction failed unexpected error: %v", err)
	} else {
		t.Logf("BuildTransaction tx = %v", tx)
	}
}

func TestSendTransaction(t *testing.T) {

	amount := decimal.NewFromFloat(1).Mul(coinDecimal)

	//建立交易单
	txID, err := SendTransaction("0E6MV60100A08",
		"bm1qjf6v463sj3w04zyk8vq0aefjrzug0jtwz62mz0",
		assetsID_btm, uint64(amount.IntPart()), "1234567", true)
	if err != nil {
		t.Errorf("SendTransaction failed unexpected error: %v", err)
		return
	}

	t.Logf("SendTransaction txid = %v", txID)
}

func TestGetTransactions(t *testing.T) {
	tx, err := GetTransactions("0E6MV60100A08")
	if err != nil {
		t.Errorf("GetTransactions failed unexpected error: %v", err)
	} else {
		t.Logf("GetTransactions tx = %v", tx)
	}
}

func TestSummaryWallets(t *testing.T) {
	a := &AccountBalance{}
	a.AccountID = "0E6MV60100A08"
	a.Password = "1234567"

	sumAddress = "bm1qjf6v463sj3w04zyk8vq0aefjrzug0jtwz62mz0"
	//添加汇总钱包的账户
	AddWalletInSummary("0E6MV60100A08", a)

	SummaryWallets()
}
