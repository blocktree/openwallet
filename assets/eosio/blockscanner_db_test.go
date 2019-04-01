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

package eosio

import (
	"testing"
)

//GetLocalNewBlock 获取本地记录的区块高度和hash
func TestGetLocalNewBlock(t *testing.T) {
	wm := testNewWalletManager()

	blockHeight, blockHash := wm.Blockscanner.GetLocalNewBlock()

	t.Log(blockHeight)
	t.Log(blockHash)
}

//GetLocalBlock 获取本地区块数据
func TestGetLocalBlock(t *testing.T) {

}

//SaveLocalNewBlock 记录区块高度和hash到本地
func TestSaveLocalNewBlock(t *testing.T) {

}

//SaveLocalBlock 记录本地新区块
func TestSaveLocalBlock(t *testing.T) {

}
