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

package mqnode

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
)

//NodeConfig 节点配置
type NodeConfig struct {
	MerchantNodeID  string
	LocalPublicKey  string
	LocalPrivateKey string
	CacheFile       string
	MerchantNodeURL string
	ConnectType     string
	Exchange       string
	QueueName       string
	ReceiveQueueName       string
	Account       string
	Password       string
}

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


func  Subscribe(subscription []*Subscription) error {
	return nil
}


type AddressVersion struct {
	Key      string `storm:"id"`
	Coin     string `json:"coin"`
	WalletID string `json:"walletID"`
	Version  uint64 `json:"version"`
	Total    uint64 `json:"total"`
}

func NewAddressVersion(json gjson.Result) *AddressVersion {
	obj := &AddressVersion{}
	//解析json
	obj.Coin = gjson.Get(json.Raw, "coin").String()
	obj.WalletID = gjson.Get(json.Raw, "walletID").String()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.Total = gjson.Get(json.Raw, "total").Uint()

	key := fmt.Sprintf("%s_%s", obj.Coin, obj.WalletID)
	obj.Key = key

	return obj
}

// MerchantDataStruct 商户数据结构协议
type MerchantDataStruct interface {
	// ToJSON 实现转为JSON
	// extra 附加条件，默认不传入
	EncodeMerchantJSON(extra ...map[string]interface{}) (json map[string]interface{})
}
