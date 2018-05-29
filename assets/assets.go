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

package assets

import (
	"github.com/blocktree/OpenWallet/assets/cardano"
	"strings"
	"github.com/blocktree/OpenWallet/assets/bytom"
)

//WalletManager 钱包管理器
type WalletManager interface {
	//初始化配置流程
	InitConfigFlow() error
	//查看配置信息
	ShowConfig() error
	//创建钱包流程
	CreateWalletFlow() error
	//创建地址流程
	CreateAddressFlow() error
	//汇总钱包流程
	SummaryFollow() error
	//备份钱包流程
	BackupWalletFlow() error
	//查看钱包列表，显示信息
	GetWalletList() error
	//发送交易
	TransferFlow() error
}

//钱包管理器组
var managers = make(map[string]WalletManager)

// RegWMD makes a WalletManager available by the name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func RegWMD(name string, manager WalletManager) {
	if manager == nil {
		panic("WalletManager: Register adapter is nil")
	}
	if _, ok := managers[name]; ok {
		panic("WalletManager: Register called twice for adapter " + name)
	}
	managers[name] = manager
}


// GetWMD 根据币种类型获取已注册的管理者
func GetWMD(symbol string) WalletManager {
	manager, ok := managers[symbol]
	if !ok {
		return nil
	}
	return manager
}

//注册钱包管理工具
func init() {
	//注册钱包管理工具
	RegWMD(strings.ToLower(cardano.Symbol), &cardano.WalletManager{})
	RegWMD(strings.ToLower(bytom.Symbol), &bytom.WalletManager{})
}