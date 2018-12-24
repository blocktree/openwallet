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

package openw

import (
	"fmt"

	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/assets/bitcoin"
	"github.com/blocktree/OpenWallet/assets/bitcoincash"
	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/assets/litecoin"
	"github.com/blocktree/OpenWallet/assets/nebulasio"
	"github.com/blocktree/OpenWallet/assets/ontology"
	"github.com/blocktree/OpenWallet/assets/qtum"
	"github.com/blocktree/OpenWallet/assets/tron"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
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
	assets.RegAssets(ethereum.Symbol, ethereum.NewWalletManager())
	assets.RegAssets(bitcoin.Symbol, bitcoin.NewWalletManager())
	assets.RegAssets(litecoin.Symbol, litecoin.NewWalletManager())
	assets.RegAssets(qtum.Symbol, qtum.NewWalletManager())
	assets.RegAssets(nebulasio.Symbol, nebulasio.NewWalletManager())
	assets.RegAssets(bitcoincash.Symbol, bitcoincash.NewWalletManager())
	assets.RegAssets(ontology.Symbol, ontology.NewWalletManager())
	assets.RegAssets(tron.Symbol, tron.NewWalletManager())
}

// GetSymbolInfo 获取资产的币种信息
func GetSymbolInfo(symbol string) (openwallet.SymbolInfo, error) {
	adapter := assets.GetAssets(symbol)
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

	adapter := assets.GetAssets(symbol)
	if adapter == nil {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	manager, ok := adapter.(openwallet.AssetsAdapter)
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}

	return manager, nil
}
