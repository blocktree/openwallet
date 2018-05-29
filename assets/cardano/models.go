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

import "github.com/tidwall/gjson"

//Wallet ada的钱包模型
type Wallet struct {
	WalletID       string `json:"walletID"`
	Name           string `json:"name"`
	Balance        string `json:"balance"`
	AccountsNumber uint64 `json:"accountsNumber"`
	Password       string `json:"password"`
	Mnemonic       string `json:"mnemonic"`
}

/*
//success 返回结果
{
	"Right": [{
		"cwId": "Ae2tdPwUPEZ2CVLXYWiatEb2aSwaR573k4NY581fMde9N2GiCqtL7h6ybhU",
		"cwMeta": {
			"cwName": "Personal Wallet 1",
			"cwAssurance": "CWANormal",
			"cwUnit": 0
		},
		"cwAccountsNumber": 3,
		"cwAmount": {
			"getCCoin": "3829106"
		},
		"cwHasPassphrase": false,
		"cwPassphraseLU": 1.517391540346574584e9
	}, ...]
}
*/

//NewWalletForV0 通过API/V0的结构实例化钱包
func NewWalletForV0(json gjson.Result) *Wallet {
	w := &Wallet{}
	//解析json
	w.WalletID = gjson.Get(json.Raw, "cwId").String()
	w.Name = gjson.Get(json.Raw, "cwMeta.cwName").String()
	w.Balance = gjson.Get(json.Raw, "cwAmount.getCCoin").String()
	w.AccountsNumber = gjson.Get(json.Raw, "cwAccountsNumber").Uint()

	return w
}

type Account struct {
	AcountID      string
	Name          string
	AddressNumber uint64
	Amount        string
}

/*
	//AcountInfo
	{
	  "caId": "string",
	  "caMeta": {
		"caName": "string"
	  },
	  "caAddresses": [

	  ],
	  "caAmount": {
		"getCCoin": "string"
	  }
	}

*/

func NewAccountV0(json gjson.Result) *Account {
	a := &Account{}
	//解析json
	a.AcountID = gjson.Get(json.Raw, "caId").String()
	a.Name = gjson.Get(json.Raw, "caMeta.caName").String()
	a.Amount = gjson.Get(json.Raw, "caAmount.getCCoin").String()

	address := gjson.Get(json.Raw, "caAddresses")
	if address.IsArray() {
		a.AddressNumber = uint64(len(address.Array()))
	}

	return a
}

type Address struct {
	Address string
	Amount  string
}

/*
	{
		"cadId": "string",
		"cadAmount": {
			"getCCoin": "string"
		},
		"cadIsUsed": true,
		"cadIsChange": true
	}
*/

func NewAddressV0(json gjson.Result) *Address {
	a := &Address{}
	//解析json
	a.Address = gjson.Get(json.Raw, "cadId").String()
	a.Amount = gjson.Get(json.Raw, "cadAmount.getCCoin").String()

	return a
}

type Transaction struct {
	TxID          string
	Amount        string
	Confirmations uint64
}

func NewTransactionV0(json gjson.Result) *Transaction {
	t := &Transaction{}
	//解析json
	t.TxID = gjson.Get(json.Raw, "ctId").String()
	t.Amount = gjson.Get(json.Raw, "ctAmount.getCCoin").String()
	t.Confirmations = gjson.Get(json.Raw, "ctConfirmations").Uint()

	return t
}
