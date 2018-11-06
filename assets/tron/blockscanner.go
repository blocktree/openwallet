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
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/graarh/golang-socketio"
)

const (
	blockchainBucket = "blockchain" //区块链数据集合
	// periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20 //并发的扫描线程数

	RPCServerCore     = 0 //RPC服务，核心钱包
	RPCServerExplorer = 1 //RPC服务，insight-API
)

//TronBlockScanner tron的区块链扫描器
type TronBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64             //当前区块高度
	extractingCH         chan struct{}      //扫描工作令牌
	wm                   *WalletManager     //钱包管理者
	RescanLastBlockCount uint64             //重扫上N个区块数量
	socketIO             *gosocketio.Client //socketIO客户端
	// IsScanMemPool        bool               //是否扫描交易池
}

//NewTronBlockScanner 创建区块链扫描器
func NewTronBlockScanner(wm *WalletManager) *TronBlockScanner {
	bs := TronBlockScanner{BlockScannerBase: openwallet.NewBlockScannerBase()}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	// bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0

	// bs.walletInScanning = make(map[string]*openwallet.Wallet)
	// bs.addressInScanning = make(map[string]string)
	// bs.observers = make(map[openwallet.BlockScanNotificationObject]bool)

	//设置扫描任务
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	extractData map[string]*openwallet.TxExtractData
	TxID        string
	BlockHeight uint64
	Success     bool
}

//SaveResult 保存结果
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// ---------------------------------------- Interface -----------------------------------------
//SetRescanBlockHeight 重置区块链扫描高度
func (bs *TronBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return errors.New("block height to rescan must greater than 0.")
	}

	block, err := bs.wm.GetBlockByNum(height)
	if err != nil {
		return err
	}
	hash := block.Hash
	bs.SaveLocalNewBlock(height, hash)

	return nil
}

// ---------------------------------------- Scanner -------------------------------------------
//ScanBlockTask 扫描任务
func (bs *TronBlockScanner) ScanBlockTask() {

	var (
		currentHeight uint64
		currentHash   string
	)

	//获取本地区块高度
	currentHeight, currentHash = bs.GetLocalNewBlock()

	//如果本地没有记录，查询接口的高度
	if currentHeight == 0 {
		log.Std.Info("No records found in local, get now block as the local!")

		block, err := bs.wm.GetNowBlock()
		if err != nil {
			log.Std.Error("ScanBlockTask Faild: %+v", err)
		}

		// 取上一个块作为初始
		block, err = bs.wm.GetBlockByNum(block.Height - 1)
		if err != nil {
			log.Std.Error("ScanBlockTask Faild: %+v", err)
		}

		currentHash = block.GetBlockHashID()
		currentHeight = block.GetHeight()
	}
	log.Std.Info("Local block height: %v", currentHeight)

	i := 0
	for {
		log.Std.Info("\n ------------------------------------ Foreach Start")

		time.Sleep(time.Second * (5 * time.Duration(i*i)))
		i += 1

		block, err := bs.wm.GetNowBlock()
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}
		hash := block.Hash
		maxHeight := block.Height

		log.Std.Info("Get now block: height=%v, hash=%v", maxHeight, hash)

		//是否已到最新高度
		if currentHeight == maxHeight {
			log.Std.Info("Break: block scanner has scanned full chain data. Current height %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		log.Std.Info("Block scanner scanning next height: %d ...", currentHeight)

		block, err = bs.wm.GetBlockByNum(currentHeight)
		if err != nil {
			log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			log.Std.Info("block height: %d extract failed.", currentHeight)
			continue
		}
		currentHash = block.GetBlockHashID()

		isFork := false

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {
			// if true {
			log.Std.Info("\n\t 分叉？")
			log.Std.Info("\tblock has been fork on height: %d.", currentHeight)
			log.Std.Info("\tblock height: %d local hash = %s ", currentHeight-1, currentHash)
			// log.Std.Info("\tblock height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)

			log.Std.Info("\tdelete recharge records on block height: %d.", currentHeight-1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				log.Std.Error("block scanner can not get local block; unexpected error: %v", err)

				//查找core钱包的RPC
				log.Info("\tblock scanner prev block height:", currentHeight)

				localBlock, err = bs.wm.GetBlockByNum(currentHeight)
				if err != nil {
					log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}
			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			log.Std.Info("\trescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)

			isFork = true

		} else {

			// err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
			// if err != nil {
			// 	log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			// }

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.SaveLocalNewBlock(currentHeight, currentHash)
			bs.SaveLocalBlock(block)

			isFork = false
		}

		//通知新区块给观测者，异步处理
		go bs.newBlockNotify(block, isFork)
	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.ScanBlock(i)
	}

	// if bs.IsScanMemPool {
	// 	//扫描交易内存池
	// 	bs.ScanTxMemPool()
	// }

	// //重扫失败区块
	// bs.RescanFailedRecord()

}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *TronBlockScanner) newBlockNotify(block *Block, isFork bool) {
	// for o, _ := range bs.Observers {
	// 	header := block.BlockHeader()
	// 	header.Fork = isFork
	// 	o.BlockScanNotify(block.BlockHeader())
	// }
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *TronBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []string) error {

	// var (
	// 	quit       = make(chan struct{})
	// 	done       = 0 //完成标记
	// 	failed     = 0
	// 	shouldDone = len(txs) //需要完成的总数
	// )

	// if len(txs) == 0 {
	// 	return errors.New("BatchExtractTransaction block is nil.")
	// }

	// //生产通道
	// producer := make(chan ExtractResult)
	// defer close(producer)

	// //消费通道
	// worker := make(chan ExtractResult)
	// defer close(worker)

	// //保存工作
	// saveWork := func(height uint64, result chan ExtractResult) {
	// 	//回收创建的地址
	// 	for gets := range result {

	// 		if gets.Success {

	// 			// notifyErr := bs.newExtractDataNotify(height, gets.extractData)
	// 			// //saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
	// 			// if notifyErr != nil {
	// 			// 	failed++ //标记保存失败数
	// 			// 	log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
	// 			// }
	// 		} else {
	// 			//记录未扫区块
	// 			unscanRecord := NewUnscanRecord(height, "", "")
	// 			bs.SaveUnscanRecord(unscanRecord)
	// 			log.Std.Info("block height: %d extract failed.", height)
	// 			failed++ //标记保存失败数
	// 		}
	// 		//累计完成的线程数
	// 		done++
	// 		if done == shouldDone {
	// 			//log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
	// 			close(quit) //关闭通道，等于给通道传入nil
	// 		}
	// 	}
	// }

	// //提取工作
	// extractWork := func(eblockHeight uint64, eBlockHash string, mTxs []string, eProducer chan ExtractResult) {
	// 	for _, txid := range mTxs {
	// 		bs.extractingCH <- struct{}{}
	// 		//shouldDone++
	// 		go func(mBlockHeight uint64, mTxid string, end chan struct{}, mProducer chan<- ExtractResult) {

	// 			//导出提出的交易
	// 			mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTxid, bs.GetSourceKeyByAddress)
	// 			//释放
	// 			<-end

	// 		}(eblockHeight, txid, bs.extractingCH, eProducer)
	// 	}
	// }

	return nil
}

// //ExtractTransaction 提取交易单
// func (bs *TronBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, txid string, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {

// 	var (
// 		result = ExtractResult{
// 			BlockHeight: blockHeight,
// 			TxID:        txid,
// 			extractData: make(map[string]*openwallet.TxExtractData),
// 		}
// 	)

// 	//log.Std.Debug("block scanner scanning tx: %s ...", txid)
// 	trx, isSuccess, err := bs.wm.GetTransactionByID(txid)

// 	if err != nil || isSuccess != true {
// 		log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
// 		result.Success = false
// 		return result
// 	}

// 	// //优先使用传入的高度
// 	// if blockHeight > 0 && trx.BlockHeight == 0 {
// 	// 	// trx.BlockHeight = blockHeight
// 	// 	// trx.BlockHash = blockHash
// 	// }
// 	// bs.extractTransaction(trx, &result, scanAddressFunc)
// 	return result

// }

// //ExtractTransactionData 提取交易单
// func (bs *TronBlockScanner) extractTransaction(trx *Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanAddressFunc) {

// 	var (
// 		success = true
// 	)

// 	if trx == nil {
// 		//记录哪个区块哪个交易单没有完成扫描
// 		success = false
// 	} else {

// 		vin := trx.Vins
// 		blocktime := trx.Blocktime

// 		//检查交易单输入信息是否完整，不完整查上一笔交易单的输出填充数据
// 		for _, input := range vin {

// 			if len(input.Coinbase) > 0 {
// 				//coinbase skip
// 				success = true
// 				break
// 			}

// 			//如果input中没有地址，需要查上一笔交易的output提取
// 			if len(input.Addr) == 0 {

// 				intxid := input.TxID
// 				vout := input.Vout

// 				preTx, isSuccess, err := bs.wm.GetTransactionByID(intxid)
// 				if err != nil || isSuccess != true {
// 					success = false
// 					break
// 				} else {
// 					preVouts := preTx.Vouts
// 					if len(preVouts) > int(vout) {
// 						preOut := preVouts[vout]
// 						input.Addr = preOut.Addr
// 						input.Value = preOut.Value
// 						//vinout = append(vinout, output[vout])
// 						success = true

// 						//log.Debug("GetTxOut:", output[vout])

// 					}
// 				}

// 			}

// 		}

// 		if success {

// 			//提取出账部分记录
// 			from, totalSpent := bs.extractTxInput(trx, result, scanAddressFunc)
// 			//log.Debug("from:", from, "totalSpent:", totalSpent)

// 			//提取入账部分记录
// 			to, totalReceived := bs.extractTxOutput(trx, result, scanAddressFunc)
// 			//log.Debug("to:", to, "totalReceived:", totalReceived)

// 			for _, extractData := range result.extractData {
// 				tx := &openwallet.Transaction{
// 					From: from,
// 					To:   to,
// 					Fees: totalSpent.Sub(totalReceived).StringFixed(8),
// 					Coin: openwallet.Coin{
// 						Symbol:     bs.wm.Symbol(),
// 						IsContract: false,
// 					},
// 					BlockHash:   trx.BlockHash,
// 					BlockHeight: trx.BlockHeight,
// 					TxID:        trx.TxID,
// 					Decimal:     8,
// 					ConfirmTime: blocktime,
// 				}
// 				wxID := openwallet.GenTransactionWxID(tx)
// 				tx.WxID = wxID
// 				extractData.Transaction = tx

// 				//log.Debug("Transaction:", extractData.Transaction)
// 			}

// 		}

// 		success = true

// 	}
// 	result.Success = success
// }
