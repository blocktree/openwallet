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

package cardano

import (
	"github.com/blocktree/OpenWallet/common"
	"testing"
)

func TestGetWalletInfo(t *testing.T) {
	tests := []struct {
		wid string
		tag string
	}{
		{
			wid: "",
			tag: "all",
		},
		{
			wid: "124556",
			tag: "id no exist",
		},
		{
			wid: "Ae2tdPwUPEZ2CVLXYWiatEb2aSwaR573k4NY581fMde9N2GiCqtL7h6ybhU",
			tag: "id exist",
		},
	}

	for i, test := range tests {

		wallets, err := GetWalletInfo(test.wid)
		if err != nil {
			t.Errorf("GetWalletInfo[%d] failed unexpected error: %v", i, err)
		}
		for j, w := range wallets {
			t.Logf("GetWalletInfo[%d] wallet[%d]  = %v", i, j, w)
		}

	}
}

func TestCreateNewWallet(t *testing.T) {
	//密钥32
	password := common.NewString("12345678").SHA256()
	words := genMnemonic()
	//annual only pact client fatigue choice arrive achieve country indoor engage coil spatial engine among paper dawn tackle bonus task lock pepper deny eye
	result := callCreateWalletAPI("chance wallet", words, password)
	t.Logf("%v\n", result)
}

func TestGetAccountInfo(t *testing.T) {

	tests := []struct {
		aid string
		tag string
	}{
		//{
		//	aid: "",
		//	tag: "all",
		//},
		//{
		//	aid: "124556",
		//	tag: "id no exist",
		//},
		{
			aid: "Ae2tdPwUPEYynni3SmWmLnDkgMkyXFraSz2MJqhgqFiAWHXtacmE9Dpk23z",
			tag: "search by wallet id",
		},
	}

	for i, test := range tests {

		accounts, err := GetAccountInfo(test.aid)
		if err != nil {
			t.Errorf("GetAccountInfo[%d] failed unexpected error: %v", i, err)
		}
		for j, a := range accounts {
			t.Logf("GetAccountInfo[%d] acount[%d]  = %v", i, j, a)
		}

	}
}

func TestCreateAccount(t *testing.T) {

	tests := []struct {
		wid      string
		name     string
		password string
		tag      string
	}{
		{
			wid:      "Ae2tdPwUPEZKrS6hL1E9XgKSet8ydFYHQkV3gmEG8RGUwh5XKugoUFEk7Lx",
			name:     "chance2",
			password: common.NewString("12345678").SHA256(),
			tag:      "normal",
		},
		{
			wid:      "Ae2tdPwUPEZKrS6hL1E9XgKSet8ydFYHQkV3gmEG8RGUwh5XKugoUFEk7Lx",
			name:     "zhiquan911",
			password: "",
			tag:      "Passphrase doesn't match",
		},
	}

	for i, test := range tests {

		err := CreateNewAccount(test.name, test.wid, test.password)
		if err != nil {
			t.Fatalf("CreateAccount[%d] failed unexpected error: %v", i, err)
		}
		t.Logf("CreateAccount[%d] address = %s", i, "")
	}

}

func TestCreateBatchAddress(t *testing.T) {
	//	Ae2tdPwUPEZKrS6hL1E9XgKSet8ydFYHQkV3gmEG8RGUwh5XKugoUFEk7Lx@2727649476

	var (
		//results = make(chan []*Address, 0)
	)

	password := common.NewString("12345678").SHA256()

	tests := []struct {
		aid   string
		count uint
	}{
		{
			aid:   "Ae2tdPwUPEYynni3SmWmLnDkgMkyXFraSz2MJqhgqFiAWHXtacmE9Dpk23z@2147483648",
			count: 10,
		},
		{
			aid:   "",
			count: 1,
		},
	}

	for i, test := range tests {

		addrs, err := CreateBatchAddress(test.aid, password, test.count)
		if err != nil {
			t.Fatalf("CreateBatchAddress[%d] failed unexpected error: %v", i, err)
		}
		for j, a := range addrs {
			t.Logf("CreateBatchAddress[%d] NewAddress[%d]  = %s", i, j, a.Address)
		}

	}
}

func TestWriteSomething(t *testing.T) {
	WriteSomething()
}

