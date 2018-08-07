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

package openwallet

import "github.com/tidwall/gjson"

type Transaction struct {

	/*

		| -> coin     | string    | 否       | 币名                              |
		| -> walletID | string    | 否       | 每次返回的代提币条目数（1-50       |
		| -> sid      | string    | 否       | 安全id（防止重复提交）              |
		| -> isMemo   | int       | 否       | 1为memo，0为address                |
		| -> address  | string    | 否       | 地址（账号）                        |
		| -> amount   | string    | 否       | 提笔金额                          |
		| -> memo     | string    | 是       | 备注                              |

		| 参数名称    | 类型   | 是否可空 | 描述                                                 |
		|-------------|--------|----------|------------------------------------------------------|
		| txid        | string | 否       | 唯一交易单号                                         |
		| from        | string | 否       | 发送地址                                             |
		| to          | string | 否       | 接收地址                                             |
		| amount      | string | 否       | 充值金额                                             |
		| confirm     | int    | 否       | 确认数                                               |
		| blockHash   | string | 否       | 区块哈希                                             |
		| blockHeight | string | 否       | 区块链高度（可以是组合高度例如ada 53_11323，poch_slot） |
		| isMemo      | int    | 否       | 1为memo，0为address                                   |
		| memo        | string | 否       | 备注                                                 |

	*/

	TxID        string `json:"txid"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
}

type Recharge struct {
	Sid         string `json:"sid"  storm:"id"`
	TxID        string `json:"txid"`
	AccountID   string `json:"accountID"`
	Address     string `json:"address"`
	Symbol      string `json:"symbol"`
	Amount      string `json:"amount"`
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight" storm:"index"`
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
	Index       uint64 `json:"index"`
	Received    bool
}

type Withdraw struct {
	Symbol   string `json:"coin"`
	WalletID string `json:"walletID"`
	Sid      string `json:"sid"  storm:"id"`
	IsMemo   bool   `json:"isMemo"`
	Address  string `json:"address"`
	Amount   string `json:"amount"`
	Memo     string `json:"memo"`
	Password string `json:"password"`
}

//NewWithdraw 创建提现单
func NewWithdraw(json gjson.Result) *Withdraw {
	w := &Withdraw{}
	//解析json
	w.Symbol = gjson.Get(json.Raw, "coin").String()
	w.WalletID = gjson.Get(json.Raw, "walletID").String()
	w.Sid = gjson.Get(json.Raw, "sid").String()
	w.IsMemo = gjson.Get(json.Raw, "isMemo").Bool()
	w.Address = gjson.Get(json.Raw, "address").String()
	w.Amount = gjson.Get(json.Raw, "amount").String()
	w.Memo = gjson.Get(json.Raw, "memo").String()
	w.Password = gjson.Get(json.Raw, "password").String()
	return w
}
