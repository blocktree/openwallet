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

package bopo

import (
	// "fmt"
	"github.com/tidwall/gjson"
	// "github.com/pkg/errors"
	// "log"
)

//GetBlockChainInfo 获取钱包区块链信息
func GetBlockChainInfo() (*BlockchainInfo, error) {
	var blockchain *BlockchainInfo

	if r, err := client.Call("synclog", "GET", nil); err != nil {
		return nil, err
	} else {
		// Bopo Return: data={"syncLog":{"timestamp":"2018-08-03T06:02:34.533332344Z","currentBlocksHeight":262311}}
		syncLog := gjson.GetBytes(r, "syncLog").Map()

		blockchain = &BlockchainInfo{Blocks: syncLog["currentBlocksHeight"].Uint()}
	}

	return blockchain, nil
}
