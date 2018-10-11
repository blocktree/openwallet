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
package nebulasio

import (
	"fmt"
	"github.com/tidwall/gjson"
)

const (
	byHeight int = iota //0
	byHash
)

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (string, error) {

	result, err := wm.WalletClient.CallGetnebstate("height")
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

//GetBlockHashByHeight 根据区块高度获得区块hash
func (wm *WalletManager) GetBlockHashByHeight(height string) (string, error) {


	result, err := wm.WalletClient.CallgetBlockByHeightOrHash(height,byHeight)
	if err != nil {
		return "", err
	}

	hash := gjson.Get(result.String(),"hash")

	return hash.String(), nil
}

//GetBlockByHeight 根据区块高度获得区块信息
func (wm *WalletManager) GetBlockByHeight(height string) (string, error) {


	block, err := wm.WalletClient.CallgetBlockByHeightOrHash(height,byHeight)
	if err != nil {
		return "", err
	}

	fmt.Printf("block.String()=%v\n",block.String())

	return block.String(), nil
}

//GetBlockByHash 根据区块hash获得区块信息
func (wm *WalletManager) GetBlockByHash(hash string) (string, error) {

	block, err := wm.WalletClient.CallgetBlockByHeightOrHash(hash,byHash)
	if err != nil {
		return "", err
	}

	fmt.Printf("block.String()=%v\n",block.String())
	return block.String(), nil
}