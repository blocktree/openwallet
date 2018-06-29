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

package merchant

import (
	"github.com/tidwall/gjson"
	"encoding/json"
)

type Subscription struct {

	/*


		#### 订阅内容 `Subscription`

		| 参数名称 | 类型   | 是否可空 | 描述                                                                   |
		|----------|--------|----------|------------------------------------------------------------------------|
		| type     | int    | 是       | 订阅类型，1：钱包余额，2：充值记录，3：汇总日志                              |
		| coin     | string | 是       | 订阅的币种钱包类型                                                     |
		| walletID | string | 否       | 钱包账户id，由商户定义，可与钱包主机里的钱包账户关联，也可与订阅地址关联 |
		| version  | int    | 否       | 地址版本号，订阅类型为2时，需要钱包管理工具主动下载订阅的地址列表        |

	*/
	ID       int    // primary key
	Type     int64  `json:"type"`
	Coin     string `json:"coin"`
	WalletID string `json:"walletID"`
	Version  int64  `json:"version"`
}

func NewSubscription(res gjson.Result) *Subscription {
	var s Subscription
	err := json.Unmarshal([]byte(res.Raw), &s)
	if err != nil {
		return nil
	}
	return &s
}

type Address struct {

}