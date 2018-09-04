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

	//AddressDecode 地址解析器
	AddressDecode() openwallet.AddressDecoder

	//TransactionDecoder 交易单解析器
	TransactionDecoder() openwallet.TransactionDecoder
}



// GetAssetsController 获取资产控制器
func GetAssetsManager(symbol string) (AssetsManager, error) {
	manager, ok := assets.Managers[strings.ToLower(symbol)].(AssetsManager)
	if !ok {
		return nil, fmt.Errorf("assets: %s is not support", symbol)
	}
	return manager, nil
}