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

import (
	"fmt"
	"github.com/tidwall/gjson"
)

//AddressDecoderV2
type AddressDecoderV2 interface {
	AddressDecoder

	// AddressDecode 地址解析
	AddressDecode(addr string, opts ...interface{}) ([]byte, error)
	// AddressEncode 地址编码
	AddressEncode(pub []byte, opts ...interface{}) (string, error)
	// CustomCreateAddress 自主实现创建账户地址
	CustomCreateAddress(account *AssetsAccount, newIndex uint64) (*Address, error)
	// SupportCustomCreateAddressFunction 支持创建地址实现
	SupportCustomCreateAddressFunction() bool
}

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
	AccountID string `json:"accountID" storm:"index"` //钱包ID
	Address   string `json:"address" storm:"id"`      //地址字符串
	PublicKey string `json:"publicKey"`               //地址公钥/赎回脚本
	Alias     string `json:"alias"`                   //地址别名，可绑定用户
	Tag       string `json:"tag"`                     //标签
	Index     uint64 `json:"index"`                   //账户ID，索引位
	HDPath    string `json:"hdPath"`                  //地址公钥根路径
	WatchOnly bool   `json:"watchOnly"`               //是否观察地址，true的时候，Index，RootPath，Alias都没有。
	Symbol    string `json:"symbol"`                  //币种类别
	Balance   string `json:"balance"`                 //余额
	IsMemo    bool   `json:"isMemo"`                  //是否备注
	Memo      string `json:"memo"`                    //备注
	//CreateAt    time.Time `json:"createdAt"`//创建时间
	CreatedTime int64  `json:"createdTime"` //创建时间
	IsChange    bool   `json:"isChange"`    //是否找零地址
	ExtParam    string `json:"extParam"`    //扩展参数，用于调用智能合约，json结构

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

// ImportAddress 待导入的地址记录
type ImportAddress struct {
	Address `storm:"inline"`
}

// AddressDecoderV2
type AddressDecoderV2Base struct {
}

//PrivateKeyToWIF 私钥转WIF
func (dec *AddressDecoderV2Base) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
	return "", fmt.Errorf("PrivateKeyToWIF not implement")
}

//PublicKeyToAddress 公钥转地址
func (dec *AddressDecoderV2Base) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {
	return "", fmt.Errorf("PublicKeyToAddress not implement")
}

//WIFToPrivateKey WIF转私钥
func (dec *AddressDecoderV2Base) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {
	return nil, fmt.Errorf("WIFToPrivateKey not implement")
}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (dec *AddressDecoderV2Base) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {
	return "", fmt.Errorf("RedeemScriptToAddress not implement")
}

// AddressDecode 地址解析
func (dec *AddressDecoderV2Base) AddressDecode(addr string, opts ...interface{}) ([]byte, error) {
	return nil, fmt.Errorf("AddressDecode not implement")
}

// AddressEncode 地址编码
func (dec *AddressDecoderV2Base) AddressEncode(pub []byte, opts ...interface{}) (string, error) {
	return "", fmt.Errorf("AddressEncode not implement")
}

// CustomCreateAddress 创建账户地址
func (dec *AddressDecoderV2Base) CustomCreateAddress(account *AssetsAccount, newIndex uint64) (*Address, error) {
	return nil, fmt.Errorf("CreateAddressByAccount not implement")
}

// SupportCustomCreateAddressFunction 支持创建地址实现
func (dec *AddressDecoderV2Base) SupportCustomCreateAddressFunction() bool {
	return false
}
