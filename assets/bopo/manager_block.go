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
	//"errors"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"

	//"github.com/hyperledger/fabric/core/util"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos"
	"golang.org/x/crypto/sha3"
)

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (uint64, error) {

	if d, err := wm.fullnodeClient.Call("synclog", "GET", nil); err != nil {
		return 0, err
	} else {
		/*
			d={"syncLog":{
				"timestamp":"2018-08-03T06:02:34.533332344Z",
				"currentBlocksHeight":262311}}
		*/
		syncLog := gjson.GetBytes(d, "syncLog").Map()

		return syncLog["currentBlocksHeight"].Uint(), nil
	}
}

/*
GetBlockChainInfo 获取钱包区块链信息

Return:
	&{Chain: Blocks:377379 Headers:0 Bestblockhash: Difficulty: Mediantime:0 Verificationprogress: Chainwork: Pruned:false}
*/
func (wm *WalletManager) GetBlockChainInfo() (*BlockchainInfo, error) {

	var blockchain *BlockchainInfo

	if d, err := wm.fullnodeClient.Call("synclog", "GET", nil); err != nil {
		return nil, err
	} else {
		/*
			d={"syncLog":{"timestamp":"2018-08-03T06:02:34.533332344Z", "currentBlocksHeight":262311}}
		*/
		syncLog := gjson.GetBytes(d, "syncLog").Map()

		blockchain = &BlockchainInfo{Blocks: syncLog["currentBlocksHeight"].Uint()}
	}

	return blockchain, nil
}

//GetBlockContent 获取钱包区块链内容
func (wm *WalletManager) GetBlockContent(height uint64) (block *Block, err error) {

	d, err := wm.fullnodeClient.Call(fmt.Sprintf("blocks/%d", height), "GET", nil)
	if err != nil {
		return nil, err
	}

	block = &Block{}
	if err := gjson.Unmarshal(d, block); err != nil {
		return nil, err
	}

	hash, err := GenBlockHash(d)
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	return block, nil
}

//GetBlockTxPayload 获取区块中交易详情
func (wm *WalletManager) GetBlockPayload(payload string) (payloadSpec *PayloadSpec, err error) {

	d, err := wm.fullnodeClient.Call("blocks/parsetxpayload", "POST", req.Param{"payload": string(payload)})
	if err != nil {
		if err.Error() != "Invalid payload" {
			log.Println(err)
		}
		return nil, err
	}

	payloadSpec = &PayloadSpec{}
	if err := gjson.Unmarshal(d, payloadSpec); err != nil {
		log.Println(err)
		return nil, err
	}

	return payloadSpec, err
}

//GetBlockHash 根据区块高度获得区块hash
func (wm *WalletManager) GetBlockHash(height uint64) (hash string, err error) {

	d, err := wm.fullnodeClient.Call(fmt.Sprintf("blocks/%d", height), "GET", nil)
	if err != nil {
		return "", err
	}

	hash, err = GenBlockHash(d)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func GenBlockHash(blockData []byte) (hash string, err error) {

	block := &protos.Block{}
	if err := gjson.Unmarshal(blockData, block); err != nil {
		return "", err
	}

	blockBytes, err := block.Bytes()
	if err != nil {
		return "", err
	}

	blockCopy, err := protos.UnmarshallBlock(blockBytes)
	if err != nil {
		return "", err
	}
	blockCopy.NonHashData = nil

	// Hash the block
	data, err := proto.Marshal(blockCopy)
	if err != nil {
		return "", err
	}

	hashBytes := make([]byte, 64)
	sha3.ShakeSum256(hashBytes, data)

	hash = base64.StdEncoding.EncodeToString(hashBytes)
	return hash, nil
}

/*
	func DecodeTxPayload(payload []byte) (chaincodeSpec *PayloadSpec, err error) {
		chaincodeSpec = &PayloadSpec{}

		// Protobuf decoding, to generate protos.ChaincodeInvocationSpec
		cpp := &protos.ChaincodeInvocationSpec{}
		if err := proto.Unmarshal(payload, cpp); err != nil {
			log.Println(err)
			return chaincodeSpec, err
		}
		chaincodeSpec.Spec = cpp

		args := cpp.ChaincodeSpec.CtorMsg.Args
		cpp.ChaincodeSpec.CtorMsg.Args = nil

		// Get TX Detail - 1: Fund type
		chaincodeSpec.Input.Type = string(args[0]) // example: "USER_FUND"

		// Get TX Detail - 2: TX Header
		userTxHeader := &pb.UserTxHeader{}
		if d, err := base64.StdEncoding.DecodeString(string(args[1])); err != nil {
		} else {
			if err := proto.Unmarshal(d, userTxHeader); err != nil {
			}
		}
		chaincodeSpec.Input.UserTxHeader = userTxHeader

		// Get TX Detail - 3: TX Fund
		fund := &pb.Fund{}
		if d, err := base64.StdEncoding.DecodeString(string(args[2])); err != nil {
		} else {
			if err := proto.Unmarshal(d, fund); err != nil {
			}
		}
		chaincodeSpec.Input.Fund = fund

		// Get TX Detail - 4: TX Signature
		signature := &pb.Signature{}
		if d, err := base64.StdEncoding.DecodeString(string(args[3])); err != nil {
		} else {
			if err := proto.Unmarshal(d, signature); err != nil {
			}
		}
		chaincodeSpec.Input.Signature = signature

		return chaincodeSpec, nil
	}
*/
