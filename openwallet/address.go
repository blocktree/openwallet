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

import (
	"github.com/tidwall/gjson"
	"time"
)

type AddressDecoder interface {

	//PrivateKeyToWIF 私钥转WIF
	PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error)
	//PublicKeyToAddress 公钥转地址
	PublicKeyToAddress(pub []byte, isTestnet bool) (string, error)
	//WIFToPrivateKey WIF转私钥
	WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error)
	//RedeemScriptToAddress 多重签名赎回脚本转地址
	RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error)
}

//Address OpenWallet地址
type Address struct {
	AccountID string    `json:"accountID" storm:"index"` //钱包ID
	Address   string    `json:"address" storm:"id"`      //地址字符串
	PublicKey string    `json:"publicKey"`               //地址公钥/赎回脚本
	Alias     string    `json:"alias"`                   //地址别名，可绑定用户
	Tag       string    `json:"tag"`                     //标签
	Index     uint64    `json:"index"`                   //账户ID，索引位
	HDPath    string    `json:"hdPath"`                  //地址公钥根路径
	WatchOnly bool      `json:"watchOnly"`               //是否观察地址，true的时候，Index，RootPath，Alias都没有。
	Symbol    string    `json:"symbol"`                  //币种类别
	Balance   string    `json:"balance"`                 //余额
	IsMemo    bool      `json:"isMemo"`                  //是否备注
	Memo      string    `json:"memo"`                    //备注
	CreatedAt time.Time `json:"createdAt"`               //创建时间
	IsChange  bool      `json:"isChange"`                //是否找零地址

	//核心地址指针
	Core interface{}
}

func NewAddress(json gjson.Result) *Address {
	obj := &Address{}
	//解析json
	obj.AccountID = gjson.Get(json.Raw, "accountID").String()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.Alias = gjson.Get(json.Raw, "alias").String()
	obj.IsMemo = gjson.Get(json.Raw, "isMemo").Bool()
	obj.Memo = gjson.Get(json.Raw, "memo").String()
	obj.Tag = gjson.Get(json.Raw, "tag").String()
	obj.Index = gjson.Get(json.Raw, "index").Uint()
	obj.HDPath = gjson.Get(json.Raw, "hdPath").String()
	obj.Symbol = gjson.Get(json.Raw, "coin").String()
	obj.Balance = gjson.Get(json.Raw, "balance").String()

	return obj
}
