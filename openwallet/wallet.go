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
	"errors"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

type Wallet struct {
	//资产地址表
	AssetsAddress map[string]string
	//类型类型： 单签，多签
	Type WalletType
	//公钥
	PublicKeys []Bytes
	//拥有者列表, 公钥hex与用户key映射
	Owners map[string]string
	//创建者
	Creator *User
	//私链合约地址
	ContractAddress string
	//必要签名数
	Required uint
	//OpenWallet的统一地址
	OpenwAddress string
}

//NewWallet 创建钱包
func NewWallet(publickeys []Bytes, users []*User, required uint, creator *User) (*Wallet, error) {

	var (
		owner    = make(map[string]string, 0)
		firstApp *App
	)

	//检查公钥和用户数量是否相等
	if len(publickeys) != len(users) {
		return nil, errors.New("PublicKey count is not equal Users count")
	}

	for i, pk := range publickeys {

		if i == 0 {
			firstApp = users[i].App
		}

		//检查公钥长度是否正确

		//检查用户key是否为空
		if len(users[i].UserKey) == 0 {
			return nil, errors.New("User's userkey is empty")
		}

		//检查用户的应用方是否一致
		if firstApp != nil && firstApp != users[i].App {
			return nil, errors.New("User's Application is not unified")
		}

		pkHex := common.Bytes2Hex(pk)
		owner[pkHex] = users[i].UserKey
	}

	w := &Wallet{
		PublicKeys: publickeys,
		Required:  required,
		Owners:    owner,
	}

	return w, nil
}

//NewHDWallet 创建HD钱包
func NewHDWallet(users []*User, derivedPath string, required uint, creator *User) (*Wallet, error) {

	var (
		publicKeys []Bytes = make([]Bytes, 0)
	)

	//通过用户根公钥和衍生路径，生成确定的子公钥

	return NewWallet(publicKeys, users, required, creator)
}

//GetUserByPublicKey 通过公钥获取用户
func (w *Wallet) GetUserByPublicKey(publickey PublicKey) *User {

	pkHex := common.Bytes2Hex(publickey)

	user := &User{
		UserKey: w.Owners[pkHex],
	}

	return user
}

// Deposit 充值
func (w *Wallet) Deposit(assets string) []byte {

	c := w.FindAssets(assets)

	return c.Deposit()
}

//FindAssets 寻找资产
func (w *Wallet) FindAssets(assets string) AssetsInferface {

	var (
		runAssets  reflect.Type
		findAssets bool
		assetsInfo *AssetsInfo
	)

	assetsInfo, findAssets = OpenWalletApp.Handlers.FindAssetsInfo(assets)

	if !findAssets {
		return nil
	}

	runAssets = assetsInfo.controllerType

	vc := reflect.New(runAssets)
	execController, ok := vc.Interface().(AssetsInferface)
	if !ok {
		return nil
	}

	execController.Init(w, vc.Interface())

	return execController
}
