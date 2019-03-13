/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet

import "github.com/tidwall/gjson"

type WalletConfig struct {
	/*
		| 参数名称 | 类型   | 是否可空 | 描述                                                           |
		|----------|--------|----------|----------------------------------------------------------------|
		| coin     | string | 否       | 币种                                                           |
		| walletID | string | 否       | 钱包ID                                                         |
		| surplus  | string | 否       | 剩余额，设置后，【余额—剩余额】低于第一笔提币金额则不提币(默认为0) |
		| fee      | string | 否       | 提币矿工费                                                     |
		| confirm  | int    | 否       | 确认次数(达到该确认次数后不再推送确认，默认30)                  |
	*/
	Key      string `storm:"id"`
	Coin     string `json:"coin"`
	WalletID string `json:"walletID" storm:"id"`
	Surplus  string `json:"surplus"`
	Fee      string `json:"fee"`
	Confirm  uint64 `json:"confirm"`
}

func NewWalletConfig(json gjson.Result) *WalletConfig {
	obj := &WalletConfig{}
	//解析json
	obj.Coin = gjson.Get(json.Raw, "coin").String()
	obj.WalletID = gjson.Get(json.Raw, "walletID").String()
	obj.Surplus = gjson.Get(json.Raw, "surplus").String()
	obj.Fee = gjson.Get(json.Raw, "fee").String()
	obj.Confirm = gjson.Get(json.Raw, "confirm").Uint()
	obj.Key = obj.Coin + "_" + obj.WalletID
	return obj
}
