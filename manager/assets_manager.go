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
	"fmt"
	"strings"

	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
)

type AssetsManager interface {
	assets.SymbolInfo

	//GetAddressDecode 地址解析器
	GetAddressDecode() openwallet.AddressDecoder

	//GetTransactionDecoder 交易单解析器
	GetTransactionDecoder() openwallet.TransactionDecoder

	//GetBlockScanner 获取区块链
	GetBlockScanner() openwallet.BlockScanner

	//ImportWatchOnlyAddress 导入观测地址
	ImportWatchOnlyAddress(address ...*openwallet.Address) error

	//GetAddressWithBalance 获取多个地址余额，使用查账户和单地址
	GetAddressWithBalance(address ...*openwallet.Address) error

	//GetAssetsAccountTransactions(wrapper *openwallet.WalletWrapper, account *openwallet.AssetsAccount) (txs []*openwallet.Transaction, err error)
}

// GetAssetsController 获取资产控制器
func GetAssetsManager(symbol string) (AssetsManager, error) {
	manager, ok := assets.Managers[strings.ToLower(symbol)]
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}
	return manager.(AssetsManager), nil
}
