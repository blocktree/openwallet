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

package manager

import (
	"encoding/json"
	"github.com/tidwall/gjson"
)

type Subscription struct {

	//ID       int    // primary key
	Type     int64  `json:"type"`
	Symbol   string `json:"symbol"`
	WalletID string `json:"walletID"`
	AppID    int64  `json:"appID"`
}

func NewSubscription(res gjson.Result) *Subscription {
	var s Subscription
	err := json.Unmarshal([]byte(res.Raw), &s)
	if err != nil {
		return nil
	}
	return &s
}


func (wm *WalletManager) Subscribe(subscription []*Subscription) error {
	//TODO:待实现
	return nil
}
