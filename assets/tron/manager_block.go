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
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"github.com/tronprotocol/grpc-gateway/api"
	"github.com/tronprotocol/grpc-gateway/core"

	"github.com/blocktree/OpenWallet/log"
)

// Done
// Function：Query the latest block
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getnowblock
// Parameters：None
// Return value：Latest block on full node
func (wm *WalletManager) GetNowBlock() (block *core.Block, err error) {

	r, err := wm.WalletClient.Call("/wallet/getnowblock", nil)
	if err != nil {
		return nil, err
	}

	block = &core.Block{}
	if err := gjson.Unmarshal(r, block); err != nil {
		log.Error(err)
		return nil, err
	}

	timestamp := block.GetBlockHeader().GetRawData().GetTimestamp() // Unit: ms
	currstamp := time.Now().UnixNano() / (1000 * 1000)              // Unit: ms

	if timestamp < currstamp-(5*1000) {
		log.Warningf(fmt.Sprintf("Get block timestamp: %d [%+v]", timestamp, time.Unix(timestamp/1000, 0)))
	}
	if timestamp < currstamp-(1*60*60*1000) {
		log.Error(fmt.Sprintf("Get block timestamp: %d [%+v]", timestamp, time.Unix(timestamp/1000, 0)))
		// log.Error(fmt.Sprintf("Current d timestamp: %d [%+v]", currstamp, time.Unix(currstamp/1000, 0)))
		log.Error(fmt.Sprintf("Now block height: %d", block.GetBlockHeader().GetRawData().GetNumber()))
		return nil, errors.New("GetNowBlock returns with unsynced!")
	}

	return block, nil
}

// Done
// Function：Query block by height
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbynum -d ‘
// 		{“num” : 100}’
// Parameters：
// 	Num is the height of the block
// Return value：specified Block object
func (wm *WalletManager) GetBlockByNum(num uint64) (block *core.Block, error error) {

	request := req.Param{"num": num}
	r, err := wm.WalletClient.Call("/wallet/getblockbynum", request)
	if err != nil {
		return nil, err
	}

	block = &core.Block{}
	if err := gjson.Unmarshal(r, block); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return block, nil
}

// Writing! Always return none?
// Function：Query block by ID
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbyid -d ‘
// 		{“value”: “0000000000038809c59ee8409a3b6c051e369ef1096603c7ee723c16e2376c73”}’
// Parameters：Block ID.
// Return value：Block Object
func (wm *WalletManager) GetBlockByID(blockID string) (block *core.Block, err error) {

	// request := req.Param{"value": base64.StdEncoding.EncodeToString([]byte(blockID))}
	request := req.Param{"value": blockID}
	r, err := wm.WalletClient.Call("/wallet/getblockbyid", request)
	if err != nil {
		return nil, err
	}

	block = &core.Block{}
	if err := gjson.Unmarshal(r, block); err != nil {
		return nil, err
	}

	fmt.Println("XDM = ", block)
	return block, nil
}

// Done
// Function：Query a range of blocks by block height
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbylimitnext -d ‘
// 		{“startNum”: 1, “endNum”: 2}’
// Parameters：
// 	startNum：Starting block height, including this block
// 	endNum：Ending block height, excluding that block
// Return value：A list of Block Objects
func (wm *WalletManager) GetBlockByLimitNext(startNum, endNum uint64) (blocks *api.BlockList, err error) {

	params := req.Param{
		"startNum": startNum,
		"endNum":   endNum,
	}

	r, err := wm.WalletClient.Call("/wallet/getblockbylimitnext", params)
	if err != nil {
		return nil, err
	}

	blocks = &api.BlockList{}
	if err := gjson.Unmarshal(r, blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}

// Done
// Function：Query the latest blocks
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbylatestnum -d ‘
// 		{“num”: 5}’
// Parameters：The number of blocks to query
// Return value：A list of Block Objects
func (wm *WalletManager) GetBlockByLatestNum(num uint64) (blocks *api.BlockList, err error) {

	if num >= 1000 {
		return nil, errors.New("Too large with parameter num to search!")
	}

	params := req.Param{
		"num": num,
	}
	r, err := wm.WalletClient.Call("/wallet/getblockbylatestnum", params)
	if err != nil {
		return nil, err
	}

	blocks = &api.BlockList{}
	if err := gjson.Unmarshal(r, blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}

// ----------------------------------- Functions -----------------------------------------------
func printBlock(block *core.Block) {
	if block == nil {
		fmt.Println("Block == nil")
	}

	fmt.Println("\nvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	fmt.Println("Transactions:")
	if block.Transactions != nil {
		for i, tx := range block.Transactions {
			// sign0 := ""
			if tx.Signature != nil {
				// sign0 = hex.EncodeToString(tx.Signature[0])
			}
			fmt.Printf("\t tx%2d: <BlockBytes=%v, BlockNum=%v, BlockHash=%+v, Contact=%v, Signature_len=%v>\n", i+1, tx.RawData.RefBlockBytes, tx.RawData.RefBlockNum, hex.EncodeToString(tx.RawData.RefBlockHash), tx.RawData.Contract, len(tx.Signature))
			// fmt.Printf("\t tx%2d=<%v> \n", i+1, tx)
		}
	}

	fmt.Println("Block Header:")
	if block.BlockHeader != nil {
		fmt.Printf("\tRawData: <Number=%v, timestamp=%v, ParentHash=%v> \n", block.BlockHeader.RawData.Number, time.Unix(block.GetBlockHeader().GetRawData().GetTimestamp()/1000, 0), hex.EncodeToString(block.BlockHeader.RawData.ParentHash))
		// fmt.Printf("\t <%+v> \n", block.BlockHeader)
	}

	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
}
