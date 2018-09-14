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
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
	"strings"
	"fmt"
)

type AssetsManager interface {

	assets.SymbolInfo

	//GetAddressDecode 地址解析器
	GetAddressDecode() openwallet.AddressDecoder

	//GetTransactionDecoder 交易单解析器
	GetTransactionDecoder() openwallet.TransactionDecoder

	//GetBlockScanner 获取区块链
	GetBlockScanner() openwallet.BlockScanner

	//CountBalanceByAddresses 计算地址的余额，用于计算账户的所有地址余额
	CountBalanceByAddresses(address ...string) (balance string, err error)

	//GetAssetsAccountTransactions(wrapper *openwallet.WalletWrapper, account *openwallet.AssetsAccount) (txs []*openwallet.Transaction, err error)
}



// GetAssetsController 获取资产控制器
func GetAssetsManager(symbol string) (AssetsManager, error) {
	manager, ok := assets.Managers[strings.ToLower(symbol)].(AssetsManager)
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}
	return manager, nil
}