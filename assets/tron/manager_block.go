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
	"log"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"

	"github.com/tronprotocol/grpc-gateway/api"
	"github.com/tronprotocol/grpc-gateway/core"
)

// Done
// Function：Query the latest block
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getnowblock
// Parameters：None
// Return value：Latest block on full node
func (wm *WalletManager) GetNowBlock() (block *core.Block, err error) {

	r, err := wm.WalletClient.Call2("/wallet/getnowblock", nil)
	if err != nil {
		return nil, err
	}

	block = &core.Block{}
	if err := gjson.Unmarshal(r, block); err != nil {
		log.Println(err)
		return nil, err
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
	r, err := wm.WalletClient.Call2("/wallet/getblockbynum", request)
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

// Done
// Function：Query block by ID
// 	demo: curl -X POST http://127.0.0.1:8090/wallet/getblockbyid -d ‘
// 		{“value”: “0000000000038809c59ee8409a3b6c051e369ef1096603c7ee723c16e2376c73”}’
// Parameters：Block ID.
// Return value：Block Object
func (wm *WalletManager) GetBlockByID(blockID string) (block *core.Block, err error) {

	request := req.Param{"blockID": blockID}
	r, err := wm.WalletClient.Call2("/wallet/getblockbyid", request)
	if err != nil {
		return nil, err
	}

	block = &core.Block{}
	if err := gjson.Unmarshal(r, block); err != nil {
		return nil, err
	}

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

	r, err := wm.WalletClient.Call2("/wallet/getblockbylimitnext", params)
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
	r, err := wm.WalletClient.Call2("/wallet/getblockbylatestnum", params)
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

	fmt.Println("\n------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("Block Header:")
	if block.BlockHeader != nil {
		fmt.Printf("\tRawData: <Number=%v, ParentHash=%v, TxTrieRoot=%v> \n", block.BlockHeader.RawData.Number, hex.EncodeToString(block.BlockHeader.RawData.ParentHash), hex.EncodeToString(block.BlockHeader.RawData.TxTrieRoot))
		// fmt.Printf("\t <%+v> \n", block.BlockHeader)
	}

	fmt.Println("Transactions:")
	if block.Transactions != nil {
		for i, tx := range block.Transactions {
			fmt.Printf("\t tx%2d: <BlockBytes=%v, BlockNum=%v, BlockHash=%+v, Contact=%v, Signature_0=%v>\n", i+1, tx.RawData.RefBlockBytes, tx.RawData.RefBlockNum, hex.EncodeToString(tx.RawData.RefBlockHash), tx.RawData.Contract, hex.EncodeToString(tx.Signature[0]))
			// fmt.Printf("\t tx%2d=<%v> \n", i+1, tx)
		}
	}
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------")
}
