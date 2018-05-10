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

//AssetsInferface 是一个给钱包调用资产的抽象接口
type AssetsInferface interface {
	//Deposit 返回钱包对该资产的充值地址
	Deposit() []byte
	//Transfer 转账amount数量，to目标地址
	Transfer(amount uint, to []byte) Transaction
	//GetBalance 获取资产
	GetBalance() uint
	//资产ABI
	ContractABI() string
	//资产的合约地址
	ContractAddress() []byte
	//资产名字
	Name() string
	//Init 初始化资产控制器
	Init(w *Wallet, app interface{})
	//DeployMultiSigWallet 部署多重签名钱包
	DeployMultiSigWallet(by Wallet) []byte
	//Boardcast 广播交易
	Boardcast(tx Transaction) []byte
	//MethodMapping() map[string]func()
}

//AssetsController 资产控制器基类
type AssetsController struct {

	//钱包
	Wallet *Wallet
	//主体控制器
	AppController interface{}
}

//Init 初始化资产控制器
func (a *AssetsController) Init(w *Wallet, app interface{}){

}

//DeployMultiSigWallet 部署多重签名钱包
func (a *AssetsController) DeployMultiSigWallet(by Wallet) []byte{
	return []byte{}
}


//Deposit 返回钱包对该资产的充值地址
func (a *AssetsController) Deposit() []byte {
	return []byte{}
}

//Transfer 转账amount数量，to目标地址
func (a *AssetsController) Transfer(amount uint, to []byte) Transaction{
	return Transaction{}
}

//GetBalance 获取资产
func (a *AssetsController) GetBalance() uint{
	return 0
}

//资产ABI
func (a *AssetsController) ContractABI() string{
	return ""
}

//资产的合约地址
func (a *AssetsController) ContractAddress() []byte{
	return []byte{}
}

//资产名字
func (a *AssetsController) Name() string{
	return ""
}

//Boardcast 广播交易
func (a *AssetsController) Boardcast(tx Transaction) []byte {
	return []byte{}
}

