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
	"data": {
		"createdAt": "2032-07-26T13:46:01.035803",
		"syncState": {
			"tag": "synced",
			"data": null
			},
		"balance": 41984918983627330,
		"hasSpendingPassword": false,
		"assuranceLevel": "normal",
		"name": "My wallet",
		"id": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg",
		"spendingPasswordLastUpdate": "2029-04-05T12:13:13.241896"
	},
	"status": "success",
	"meta": {
	"pagination": {
		"totalPages": 0,
		"page": 1,
		"perPage": 10,
		"totalEntries": 0
		}
	}
}
*/

//NewWalletForV1 通过API/V1的结构实例化钱包
func NewWalletForV1(json gjson.Result) *Wallet {
	w := &Wallet{}
	//解析json
	w.WalletID = gjson.Get(json.Raw, "id").String()
	w.Name = gjson.Get(json.Raw, "name").String()
	w.Balance = gjson.Get(json.Raw, "balance").String()

	return w
}

type Account struct {
	Index         int64
	Name          string
	AddressNumber uint64
	Amount        string
	WId           string
	AddressId     string
	Used          bool
	ChangeAddress bool
}

/*
	//AcountInfo
	{
		"amount": 11091176260625604,
		"addresses": [
		{
			"used": true,
			"changeAddress": true,
			"id": "Ae2tdPwUPEZ3hmyBxBGfSLZHzETDof8afvg1vFogbukKJS75xChNyvwiuR6"
		}
		],
		"name": "My account",
		"walletId": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg",
		"index": 1
	}

*/

func NewAccountV1(json gjson.Result) *Account {
	a := &Account{}
	//解析json
	a.Index = gjson.Get(json.Raw, "index").Int()
	a.Name = gjson.Get(json.Raw, "name").String()
	a.Amount = gjson.Get(json.Raw, "amount").String()
	a.WId = gjson.Get(json.Raw, "walletId").String()
	a.AddressId = gjson.Get(json.Raw, "addresses.id").String()
	a.Used = gjson.Get(json.Raw, "addresses.used").Bool()
	a.ChangeAddress = gjson.Get(json.Raw, "addresses.changeAddress").Bool()

	address := gjson.Get(json.Raw, "addresses")
	if address.IsArray() {
		a.AddressNumber = uint64(len(address.Array()))
	}

	return a
}

type Address struct {
	Address string
	Used          bool
	ChangeAddress bool
}

/*
{
	"data": {
		"used": false,
		"changeAddress": true,
		"id": "YQdiVKZwx4SzRjJNqf6TNvRa8PYARGWRhPYHuJzfxAy5yydanhUhuuhE3j9LTuV5FBytAfN"
	},
	"status": "success",
	"meta": {
	"pagination": {
		"totalPages": 0,
		"page": 1,
		"perPage": 10,
		"totalEntries": 0
		}
	}
}
*/

func NewAddressV1(json gjson.Result) *Address {
	a := &Address{}
	//解析json
	a.Address = gjson.Get(json.Raw, "id").String()
	a.Used = gjson.Get(json.Raw, "used").Bool()
	a.ChangeAddress = gjson.Get(json.Raw, "changeAddress").Bool()

	return a
}

/*
tx
{
  "data": {
    "creationTime": "1977-09-26T13:35:28.425218",
    "status": {
      "tag": "persisted",
      "data": {}
    },
    "amount": 28311699119270856,
    "inputs": [
      {
        "amount": 13019322832706804,
        "address": "EqGAuA8vHnPAaHvxA8iajiUpsuoabdaXjEuYZRtonABww7hszPPS7ia8QJYUzbpnSnCrMhgJVvrniXvyBrLzsRgB8B73Uyei1zoizF32YvUE2CyvQijGdy5"
      }
    ],
    "direction": "outgoing",
    "outputs": [
      {
        "amount": 14984536053034712,
        "address": "2cWKMJemoBakXCwSJnQirfBC6DAgQwER6dXHtnY9LYQ6p4T1wREnzMmi83upacCB3KfT9"
      },
      {
        "amount": 13290197696443540,
        "address": "2657WMsDfac6VJfxAHRezMNbBhvYzrDVf5F2W7s7LGAgtp6m9B5Df8EQkwktLsEee"
      },
      {
        "amount": 21095182977599616,
        "address": "Un4ZpTvXtoLTFKHDvR7cPV94gefp41djrSZeUtE9rDE5zEKbAzTZppoqydgw2jfFXvdRbTMW4KCVLatj1YycK11YtmgRF2QT9Qewj3HT8rdMJjaa"
      }
    ],
    "confirmations": 4,
    "id": "848d6fadc04dcd9af1bc462df5938ecfbe810c5ecb50971db1cf7ae224bb5955",
    "type": "foreign"
  },
  "status": "success",
  "meta": {
    "pagination": {
      "totalPages": 0,
      "page": 1,
      "perPage": 10,
      "totalEntries": 0
    }
  }
}
*/

type Transaction struct {
	TxID          string
	Amount        string
	Confirmations uint64
}

func NewTransactionV1(json gjson.Result) *Transaction {
	t := &Transaction{}
	//解析json
	t.TxID = gjson.Get(json.Raw, "id").String()
	t.Amount = gjson.Get(json.Raw, "amount").String()
	t.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()

	return t
}
