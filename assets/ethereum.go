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

import "github.com/blocktree/OpenWallet/openwallet"

//EthereumAssets Ethereum资产控制器实现
type EthereumAssets struct {
	openwallet.AssetsController
}

//Deposit 返回钱包对该资产的充值地址
func (a *EthereumAssets) Deposit() []byte {
	return []byte("hello world")
}