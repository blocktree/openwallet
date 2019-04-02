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

package openw

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/assets/bitcoin"
	"github.com/blocktree/openwallet/assets/bitcoincash"
	"github.com/blocktree/openwallet/assets/eosio"
	"github.com/blocktree/openwallet/assets/ethereum"
	"github.com/blocktree/openwallet/assets/litecoin"
	"github.com/blocktree/openwallet/assets/nebulasio"
	"github.com/blocktree/openwallet/assets/ontology"
	"github.com/blocktree/openwallet/assets/qtum"
	"github.com/blocktree/openwallet/assets/tron"
	"github.com/blocktree/openwallet/assets/truechain"
	"github.com/blocktree/openwallet/assets/virtualeconomy"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"strings"
)

//type AssetsManager interface {
//	openwallet.AssetsAdapter
//
//	//ImportWatchOnlyAddress 导入观测地址
//	ImportWatchOnlyAddress(address ...*openwallet.Address) error
//
//	//GetAddressWithBalance 获取多个地址余额，使用查账户和单地址
//	GetAddressWithBalance(address ...*openwallet.Address) error
//}

//注册钱包管理工具
func initAssetAdapter() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	RegAssets(ethereum.Symbol, ethereum.NewWalletManager())
	RegAssets(bitcoin.Symbol, bitcoin.NewWalletManager())
	RegAssets(litecoin.Symbol, litecoin.NewWalletManager())
	RegAssets(qtum.Symbol, qtum.NewWalletManager())
	RegAssets(nebulasio.Symbol, nebulasio.NewWalletManager())
	RegAssets(bitcoincash.Symbol, bitcoincash.NewWalletManager())
	RegAssets(ontology.Symbol, ontology.NewWalletManager())
	RegAssets(tron.Symbol, tron.NewWalletManager())
	RegAssets(virtualeconomy.Symbol, virtualeconomy.NewWalletManager())
	RegAssets(eosio.Symbol, eosio.NewWalletManager())
	RegAssets(truechain.Symbol, truechain.NewWalletManager())
}


//区块链适配器注册组
var assetsAdapterManagers = make(map[string]interface{})

// RegAssets 注册资产
// @param name 资产别名
// @param manager 资产适配器或管理器
// @param config 加载配置
// 资产适配器实现了openwallet.AssetsConfig，可以传入配置接口完成预加载配置
// usage:
// RegAssets(cardano.Symbol, &cardano.WalletManager{}, c)
// RegAssets(bytom.Symbol, &bytom.WalletManager{}, c)
func RegAssets(name string, manager interface{}, config ...config.Configer) {
	name = strings.ToUpper(name)
	if manager == nil {
		panic("assets: Register adapter is nil")
	}
	if _, ok := assetsAdapterManagers[name]; ok {
		log.Error("assets: Register called twice for adapter ", name)
		return
	}

	assetsAdapterManagers[name] = manager

	//如果有配置则加载所有配置
	if ac, ok := manager.(openwallet.AssetsConfig); ok && config != nil {
		for _, c := range config {
			ac.LoadAssetsConfig(c)
		}
	}
}

// GetAssets 根据币种类型获取已注册的管理者
func GetAssets(symbol string) interface{} {
	symbol = strings.ToUpper(symbol)
	manager, ok := assetsAdapterManagers[symbol]
	if !ok {
		return nil
	}
	return manager
}

// GetSymbolInfo 获取资产的币种信息
func GetSymbolInfo(symbol string) (openwallet.SymbolInfo, error) {
	adapter := GetAssets(symbol)
	if adapter == nil {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	manager, ok := adapter.(openwallet.SymbolInfo)
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	return manager, nil
}

// GetAssetsAdapter 获取资产控制器
func GetAssetsAdapter(symbol string) (openwallet.AssetsAdapter, error) {

	adapter := GetAssets(symbol)
	if adapter == nil {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	manager, ok := adapter.(openwallet.AssetsAdapter)
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	return manager, nil
}
