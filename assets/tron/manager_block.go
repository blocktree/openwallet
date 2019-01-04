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

package tron

import (
	"errors"
	"fmt"
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/imroc/req"
)

func (wm *WalletManager) GetCurrentBlock() (block *Block, err error) {
	r, err := wm.WalletClient.Call("/wallet/getnowblock", nil)
	if err != nil {
		return nil, err
	}
	block = NewBlock(r)
	if block.GetBlockHashID() == "" || block.GetHeight() <= 0 {
		return nil, errors.New("GetNowBlock failed: No found <block>")
	}
	return block, nil
}

// GetNowBlock Done!
// Function：Query the latest block
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getnowblock
// Parameters：None
// Return value：Latest block on full node
func (wm *WalletManager) GetNowBlock() (block *Block, err error) {

	r, err := wm.WalletClient.Call("/wallet/getnowblock", nil)
	if err != nil {
		return nil, err
	}

	block = NewBlock(r)
	if block.GetBlockHashID() == "" || block.GetHeight() <= 0 {
		return nil, errors.New("GetNowBlock failed: No found <block>")
	}

	// Check for TX
	currstamp := time.Now().UnixNano() / (1000 * 1000) // Unit: ms
	timestamp := int64(block.Time)
	if timestamp < currstamp-(5*1000) {
		log.Warningf(fmt.Sprintf("Get block timestamp: %d [%+v]", timestamp, time.Unix(timestamp/1000, 0)))
		log.Warningf(fmt.Sprintf("Current d timestamp: %d [%+v]", currstamp, time.Unix(currstamp/1000, 0)))
		log.Warningf("Diff seconds: %ds ", (currstamp-timestamp)/1000)
	}
	if timestamp < currstamp-(5*60*1000) {
		log.Error(fmt.Sprintf("Get block timestamp: %d [%+v]", timestamp, time.Unix(timestamp/1000, 0)))
		// log.Error(fmt.Sprintf("Current d timestamp: %d [%+v]", currstamp, time.Unix(currstamp/1000, 0)))
		log.Error(fmt.Sprintf("Now block height: %d", block.GetHeight()))
		return nil, errors.New("GetNowBlock returns with unsynced")
	}

	return block, nil
}

// GetBlockByNum Done!
// Function：Query block by height
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbynum -d ‘
// 		{“num” : 100}’
// Parameters：
// 	Num is the height of the block
// Return value：specified Block object
func (wm *WalletManager) GetBlockByNum(num uint64) (block *Block, error error) {

	r, err := wm.WalletClient.Call("/wallet/getblockbynum", req.Param{"num": num})
	if err != nil {
		return nil, err
	}

	block = NewBlock(r)
	if block.GetBlockHashID() == "" || block.GetHeight() <= 0 {
		return nil, errors.New("GetBlockByNum failed: No found <block>")
	}

	return block, nil
}

// GetBlockByID Done!
// Function：Query block by ID
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbyid -d ‘
// 		{“value”: “0000000000038809c59ee8409a3b6c051e369ef1096603c7ee723c16e2376c73”}’
// Parameters：Block ID.
// Return value：Block Object
func (wm *WalletManager) GetBlockByID(blockID string) (block *Block, err error) {

	r, err := wm.WalletClient.Call("/wallet/getblockbyid", req.Param{"value": blockID})
	if err != nil {
		return nil, err
	}

	block = NewBlock(r)
	if block.GetBlockHashID() == "" || block.GetHeight() <= 0 {
		return nil, errors.New("GetBlockByID failed: No found <block>")
	}

	return block, nil
}

// GetBlockByLimitNext Done!
// Function：Query a range of blocks by block height
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbylimitnext -d ‘
// 		{“startNum”: 1, “endNum”: 2}’
// Parameters：
// 	startNum：Starting block height, including this block
// 	endNum：Ending block height, excluding that block
// Return value：A list of Block Objects
func (wm *WalletManager) GetBlockByLimitNext(startNum, endNum uint64) (blocks []*Block, err error) {

	params := req.Param{
		"startNum": startNum,
		"endNum":   endNum,
	}

	r, err := wm.WalletClient.Call("/wallet/getblockbylimitnext", params)
	if err != nil {
		return nil, err
	}

	blocks = []*Block{}
	for _, raw := range r.Get("block").Array() {
		b := NewBlock(&raw)
		blocks = append(blocks, b)
	}

	return blocks, nil
}

// GetBlockByLatestNum Done!
// Function：Query the latest blocks
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbylatestnum -d ‘
// 		{“num”: 5}’
// Parameters：The number of blocks to query
// Return value：A list of Block Objects
func (wm *WalletManager) GetBlockByLatestNum(num uint64) (blocks []*Block, err error) {

	if num >= 1000 {
		return nil, errors.New("Too large with parameter num to search")
	}

	r, err := wm.WalletClient.Call("/wallet/getblockbylatestnum", req.Param{"num": num})
	if err != nil {
		return nil, err
	}

	// blocks = &api.BlockList{}
	// if err := gjson.Unmarshal(r, blocks); err != nil {
	// 	return nil, err
	// }
	blocks = []*Block{}
	for _, raw := range r.Get("block").Array() {
		b := NewBlock(&raw)
		blocks = append(blocks, b)
	}

	return blocks, nil
}

// ----------------------------------- Functions -----------------------------------------------
func printBlock(block *Block) {
	if block == nil {
		fmt.Println("Block == nil")
	}

	fmt.Println("\nvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	fmt.Println("Transactions:")
	for i, tx := range block.GetTransactions() {
		fmt.Printf("\t tx%2d: IsSuccess=%v \t<BlockBytes=%v, BlockNum=%v, BlockHash=%+v, Contact=%v, Signature_len=%v>\n", i+1, tx.IsSuccess, "", tx.BlockHeight, tx.BlockHash, "", "")
	}

	fmt.Println("Block Header:")
	if block != nil {
		fmt.Printf("\tRawData: <Number=%v, timestamp=%v, ParentHash=%v> \n", block.GetHeight(), time.Unix(int64(block.Time)/1000, 0), block.Previousblockhash)
	}

	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
}
