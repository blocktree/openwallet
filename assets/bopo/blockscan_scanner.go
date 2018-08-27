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
	"encoding/base64"
	// "errors"
	"fmt"
	"log"
	// "path/filepath"
	// "sync"
	//"time"
	// "github.com/asdine/storm"
	// "github.com/blocktree/OpenWallet/crypto"
	//"github.com/blocktree/OpenWallet/log"
	//"github.com/blocktree/OpenWallet/timer"
	// "github.com/tidwall/gjson"
)

//scanning 扫描
func (bs *FabricBlockScanner) scanBlock() {

	//currentHeight, err := GetBlockHeight()
	blockinfo, err := bs.wm.GetBlockChainInfo()
	if err != nil {
		// log.Printf("block scanner can not get new block height; unexpected error: %v\n", err)
		return
	}
	currentHeight := blockinfo.Blocks

	//for height := uint64(330000); height <= currentHeight; height++ { //Foreach Blocks
	for height := uint64(372321); height <= currentHeight; height++ { //Foreach Blocks

		// Load Block Info
		block, err := bs.wm.GetBlockContent(height)
		if err != nil {
			log.Printf("Get block [%d] faild: %v\n", height, err)
		}

		fmt.Printf("Height=[%d/%d]Len(TXs)=[%d]\tStateHash[%s]PreHash[%s]\n", height, currentHeight, len(block.Transactions), block.Statehash, block.Previousblockhash)

		for i, v := range block.Transactions { // Foreach all transactions

			fmt.Printf("\tNo.[%2d]\tType=[%s]\tChaincodeID[%s]", i, v.Type, v.ChaincodeID)

			if payloadSpec, err := bs.wm.GetBlockPayload(base64.StdEncoding.EncodeToString(v.Payload)); err != nil {
				log.Printf("Decode TX [%d] Payload faild: %v\n", height, err)
			} else {
				//fmt.Println(payloadSpec)
				fmt.Printf("\tFrom[%s]to[%s]with[%d Pai]", payloadSpec.From, payloadSpec.To, payloadSpec.Amount)
				if payloadSpec.From == "5ZaPXfJaLNrGnXuyXunFE4xKxakEzgTIZQ" {
					fmt.Println("simonluo")
					if payloadSpec.To == "5ZFVVP47Rf5j-k7LoiRcNozlc8dynbPYng" {
						fmt.Println("xcluo")
					}
				}
			}
			fmt.Printf("\n")
		}
	}

	// for {

	// 	//获取最大高度
	// 	maxHeight, err := bs.wm.GetBlockHeight()
	// 	if err != nil {
	// 		//下一个高度找不到会报异常
	// 		log.Printf("block scanner can not get rpc-server block height; unexpected error: %v\n", err)
	// 		break
	// 	}

	// 	//是否已到最新高度
	// 	if currentHeight == maxHeight {
	// 		log.Printf("block scanner has scanned full chain data. Current height: %d\n", maxHeight)
	// 		break
	// 	}

	// 	//继续扫描下一个区块
	// 	currentHeight = currentHeight + 1

	// 	log.Printf("block scanner scanning height: %d ...\n", currentHeight)

	// 	hash, err := bs.wm.GetBlockHash(currentHeight)
	// 	if err != nil {
	// 		//下一个高度找不到会报异常
	// 		log.Printf("block scanner can not get new block hash; unexpected error: %v\n", err)
	// 		break
	// 	}

	// 	block, err := bs.wm.GetBlock(hash)
	// 	if err != nil {
	// 		log.Printf("block scanner can not get new block data; unexpected error: %v\n", err)

	// 		//记录未扫区块
	// 		unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
	// 		bs.SaveUnscanRecord(unscanRecord)
	// 		log.Printf("block height: %d extract failed.\n", currentHeight)
	// 		continue
	// 	}

	// 	//判断hash是否上一区块的hash
	// 	if currentHash != block.Previousblockhash {

	// 		log.Printf("block has been fork on height: %d.\n", currentHeight)
	// 		log.Printf("block height: %d local hash = %s \n", currentHeight-1, currentHash)
	// 		log.Printf("block height: %d mainnet hash = %s \n", currentHeight-1, block.Previousblockhash)

	// 		log.Printf("delete recharge records on block height: %d.\n", currentHeight-1)

	// 		//删除上一区块链的所有充值记录
	// 		bs.DeleteRechargesByHeight(currentHeight - 1)
	// 		//删除上一区块链的未扫记录
	// 		bs.wm.DeleteUnscanRecord(currentHeight - 1)
	// 		currentHeight = currentHeight - 2 //倒退2个区块重新扫描
	// 		if currentHeight <= 0 {
	// 			currentHeight = 1
	// 		}

	// 		localBlock, err := bs.wm.GetLocalBlock(currentHeight)
	// 		if err != nil {
	// 			log.Printf("block scanner can not get local block; unexpected error: %v\n", err)
	// 			break
	// 		}

	// 		//重置当前区块的hash
	// 		currentHash = localBlock.Hash

	// 		log.Printf("rescan block on height: %d, hash: %s .\n", currentHeight, currentHash)

	// 		//重新记录一个新扫描起点
	// 		bs.wm.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)
	// 	} else {

	// 		err = bs.BatchExtractTransaction(block.Height, block.tx)
	// 		if err != nil {
	// 			log.Printf("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
	// 		}

	// 		//重置当前区块的hash
	// 		currentHash = hash

	// 		//保存本地新高度
	// 		bs.wm.SaveLocalNewBlock(currentHeight, currentHash)
	// 		bs.wm.SaveLocalBlock(block)

	// 		//通知新区块给观测者，异步处理
	// 		go bs.newBlockNotify(block)
	// 	}
	// }

	// //扫描交易内存池
	// bs.ScanTxMemPool()

	// //重扫失败区块
	// bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *FabricBlockScanner) ScanBlock(height uint64) error {

	log.Printf("block scanner scanning height: %d ...\n", height)

	// hash, err := bs.wm.GetBlockHash(height)
	// if err != nil {
	// 	//下一个高度找不到会报异常
	// 	log.Printf("block scanner can not get new block hash; unexpected error: %v\n", err)
	// 	return err
	// }

	block, err := bs.wm.GetBlockContent(height)
	if err != nil {
		log.Printf("block scanner can not get new block data; unexpected error: %v\n", err)

		// //记录未扫区块
		// unscanRecord := NewUnscanRecord(height, "", err.Error())
		// bs.SaveUnscanRecord(unscanRecord)
		// log.Printf("block height: %d extract failed.\n", height)
		return err
	}

	// err = bs.BatchExtractTransaction(block.Height, block.tx)
	// if err != nil {
	// 	log.Printf("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
	// }

	//保存区块
	bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block)

	return nil
}

// ----------------------------------------------------------------------
//newBlockNotify 获得新区块后，通知给观测者
func (bs *FabricBlockScanner) newBlockNotify(block *Block) {
	for o, _ := range bs.observers {
		_ = o
		// o.BlockScanNotify(block.Block{})
	}
}

// //BatchExtractTransaction 批量提取交易单
// //bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
// func (bs *FabricBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []string) error {

// 	var (
// 		quit       = make(chan struct{})
// 		done       = 0 //完成标记
// 		failed     = 0
// 		shouldDone = len(txs) //需要完成的总数
// 	)

// 	if len(txs) == 0 {
// 		return errors.New("BatchExtractTransaction block is nil.")
// 	}

// 	//生产通道
// 	producer := make(chan ExtractResult)
// 	defer close(producer)

// 	//消费通道
// 	worker := make(chan ExtractResult)
// 	defer close(worker)

// 	//保存工作
// 	saveWork := func(height uint64, result chan ExtractResult) {
// 		//回收创建的地址
// 		for gets := range result {

// 			//saveResult := SaveResult{}
// 			//saveResult.TxID = gets.TxID
// 			//saveResult.BlockHeight = height

// 			if gets.Success {
// 				saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
// 				if saveErr != nil {
// 					//log.Std.Error("SaveTxToWalletDB unexpected error: %v", saveErr)
// 					//saveResult.Success = false
// 					failed++ //标记保存失败数
// 				} else {
// 					//saveResult.Success = true
// 				}
// 			} else {
// 				//记录未扫区块
// 				unscanRecord := NewUnscanRecord(height, gets.TxID, gets.Reason)
// 				bs.SaveUnscanRecord(unscanRecord)
// 				log.Std.Info("block height: %d extract failed.", height)
// 				//saveResult.Success = false
// 				failed++ //标记保存失败数
// 			}

// 			//累计完成的线程数
// 			done++
// 			if done == shouldDone {
// 				//log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
// 				close(quit) //关闭通道，等于给通道传入nil
// 			}
// 		}
// 	}

// 	//提取工作
// 	extractWork := func(eblockHeight uint64, eBlockHash string, mTxs []string, eProducer chan ExtractResult) {
// 		for _, txid := range mTxs {
// 			bs.extractingCH <- struct{}{}
// 			//shouldDone++
// 			go func(mBlockHeight uint64, mTxid string, end chan struct{}, mProducer chan<- ExtractResult) {

// 				//导出提出的交易
// 				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTxid)
// 				//释放
// 				<-end

// 			}(eblockHeight, txid, bs.extractingCH, eProducer)
// 		}
// 	}

// 	/*	开启导出的线程	*/

// 	//独立线程运行消费
// 	go saveWork(blockHeight, worker)

// 	//独立线程运行生产
// 	go extractWork(blockHeight, blockHash, txs, producer)

// 	//以下使用生产消费模式
// 	bs.extractRuntime(producer, worker, quit)

// 	if failed > 0 {
// 		return fmt.Errorf("SaveTxToWalletDB failed")
// 	} else {
// 		return nil
// 	}
// 	//return nil
// }
