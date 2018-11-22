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
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"strings"
)

//钱包管理器组
var managers = make(map[string]interface{})

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
	if _, ok := managers[name]; ok {
		log.Error("assets: Register called twice for adapter ", name)
		return
	}

	managers[name] = manager

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
	manager, ok := managers[symbol]
	if !ok {
		return nil
	}
	return manager
}