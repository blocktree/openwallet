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
	"log"
	"github.com/tidwall/gjson"
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
			wid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj",
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
	result := callCreateWalletAPI("test wallet", words, password)
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
		{
			aid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj",
			tag: "id no exist",
		},
		//{
		//	aid: "Ae2tdPwUPEYynni3SmWmLnDkgMkyXFraSz2MJqhgqFiAWHXtacmE9Dpk23z",
		//	tag: "search by wallet id",
		//},
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

		a, err := CreateNewAccount(test.name, test.wid, test.password)
		if err != nil {
			t.Fatalf("CreateAccount[%d] failed unexpected error: %v", i, err)
		}
		t.Logf("CreateAccount[%d] AcountID = %s", i, a.AcountID)
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
			count: 101,
		},
		//{
		//	aid:   "",
		//	count: 1,
		//},
	}

	for i, test := range tests {

		_, _, err := CreateBatchAddress(test.aid, password, test.count)
		if err != nil {
			t.Fatalf("CreateBatchAddress[%d] failed unexpected error: %v", i, err)
		}
		//for j, a := range addrs {
		//	t.Logf("CreateBatchAddress[%d] NewAddress[%d]  = %s", i, j, a.Address)
		//}

	}
}

func TestGetAddressInfo(t *testing.T) {

	tests := []struct {
		aid string
		tag string
	}{
		//{
		//	aid: "",
		//	tag: "all",
		//},
		//{
		//	aid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@999",
		//	tag: "id no exist",
		//},
		{
			aid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@2147483648",
			tag: "search by wallet id",
		},
	}

	for i, test := range tests {

		addrs, err := GetAddressInfo(test.aid)
		if err != nil {
			t.Errorf("GetAddressInfo[%d] failed unexpected error: %v", i, err)
		}
		for j, a := range addrs {
			t.Logf("GetAddressInfo[%d] address[%d]  = %v", i, j, a)
		}

	}
}

func TestSendTx(t *testing.T) {
	tests := []struct {
		aid    string
		to     string
		amount uint64
		tag    string
	}{
		//{
		//	aid: "",
		//	tag: "all",
		//},
		//{
		//	aid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@999",
		//	tag: "id no exist",
		//},
		{
			aid:    "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@2147483648",
			to:     "DdzFFzCqrhtAFVdoEx9jtTUhovkRnuMNUiCnuv8pQ2WYxhDVXAAhYJCVuydGSa1qJeSCYH8x6ZV8QjCaTiAHr9QReMaMRVtUQk9fMNTW",
			amount: 2500000,
			tag:    "sendTx by acount id",
		},
	}

	//密钥32
	password := common.NewString("12345678").SHA256()

	for i, test := range tests {

		tx, err := SendTx(test.aid, test.to, test.amount, password)
		if err != nil {
			t.Errorf("SendTx[%d] failed unexpected error: %v", i, err)
		} else {
			t.Logf("SendTx[%d] tx = %v", i, tx)
		}


	}
}

func TestSummaryFollow(t *testing.T) {
	SummaryFollow()
}

func TestEstimateFees(t *testing.T) {

	tests := []struct {
		aid    string
		to     string
		amount uint64
		tag    string
	}{
		//{
		//	aid: "",
		//	tag: "all",
		//},
		//{
		//	aid: "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@999",
		//	tag: "id no exist",
		//},
		{
			aid:    "Ae2tdPwUPEYyDotaZtiqRJMeu55Ukt9CHm3XMoWcxipoRC9QSxgU9KPjEMj@2147483648",
			to:     "DdzFFzCqrhtAFVdoEx9jtTUhovkRnuMNUiCnuv8pQ2WYxhDVXAAhYJCVuydGSa1qJeSCYH8x6ZV8QjCaTiAHr9QReMaMRVtUQk9fMNTW",
			amount: 5134348,
			tag:    "sendTx by acount id",
		},
	}

	for _, test := range tests {

		result := callEstimateFeesAPI(test.aid, test.to, test.amount)
		err := isError(result)
		if err != nil {
			log.Printf("%v\n", err)
		}
		fees := gjson.GetBytes(result, "Right.getCCoin")
		log.Printf("fees %s\n", fees.String())
	}


}