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
	"errors"
	"fmt"
	"time"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

//scanning 扫描
func (bs *FabricBlockScanner) scanBlock() {

	blockHeader, err := bs.GetCurrentBlockHeader()
	if err != nil {
		log.Std.Error("block scanner can not get new block height; unexpected error: %v", err)
	}
	currentHeight, currentHash := blockHeader.Height, blockHeader.Hash
	log.Std.Info("Start -> [height=%d] [hash=%s]\n", currentHeight, currentHash)

	for {

		//获取最大高度
		maxHeight, err := bs.wm.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v\n", err)
			break
		}

		//是否已到最新高度
		if currentHeight == maxHeight {
			log.Std.Info("block scanner has scanned full chain data. Current height: %d\n", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1
		log.Std.Info("Block scanner scanning height from: %d ...\n", currentHeight)

		hash, err := bs.wm.GetBlockHash(currentHeight)
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Info("block scanner can not get new block hash; unexpected error: %v\n", err)
			break
		}

		block, err := bs.wm.GetBlockContent(currentHeight)
		if err != nil {
			log.Std.Info("block scanner can not get new block data; unexpected error: %v\n", err)

			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, currentHash, err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			log.Std.Info("block height: %d extract failed.\n", currentHeight)
			continue
		}

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {

			log.Std.Info("block has been fork on height: %d.", currentHeight)
			log.Std.Info("block height: %d local hash = %s", currentHeight-1, currentHash)
			log.Std.Info("block height: %d mainnet hash = %s", currentHeight-1, block.Previousblockhash)

			log.Std.Info("delete recharge records on block height: %d.\n", currentHeight-1)

			//删除上一区块链的所有充值记录
			bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				log.Std.Info("block scanner can not get local block; unexpected error: %v\n", err)
				break
			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			log.Std.Info("rescan block on height: %d, hash: %s .\n", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)
		} else {

			err = bs.BatchExtractTransaction(block.Height, "", block.Transactions)
			if err != nil {
				log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
			}

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.SaveLocalNewBlock(currentHeight, currentHash)
			bs.SaveLocalBlock(block)

			//通知新区块给观测者，异步处理
			go bs.newBlockNotify(block)
		}
	}

	// //扫描交易内存池
	// bs.ScanTxMemPool()		// Fabric not support

	//重扫失败区块
	bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *FabricBlockScanner) ScanBlock(height uint64) error {

	log.Std.Info("[Start] block scanner scanning height: %d ...", height)

	block, err := bs.wm.GetBlockContent(height)
	if err != nil {
		log.Std.Info("block scanner can not get new block data; unexpected error: %v\n", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		log.Std.Info("block height: %d extract failed.\n", height)
		return err
	}

	err = bs.BatchExtractTransaction(block.Height, block.Hash, block.Transactions)
	if err != nil {
		log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
	}

	//保存区块
	// bs.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block)

	return nil
}

// ----------------------------------------------------------------------
//BatchExtractTransaction 批量提取交易单
//Fabric block can be set any size to include Txs, and that default value is less than 20 about BOPO net.
func (bs *FabricBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []*BlockTX) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(txs) //需要完成的总数
	)

	if len(txs) == 0 {
		return errors.New("BatchExtractTransaction block is nil.")
	}

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				// log.Std.Info("chan -> height:", height, " rechanges:", gets.Recharges)
				saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
				if saveErr != nil {
					log.Std.Error("SaveRechargeToWalletDB unexpected error: %v", saveErr)
					//saveResult.Success = false
					failed++ //标记保存失败数
				} else {
					//saveResult.Success = true
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, gets.TxID, gets.Reason)
				bs.SaveUnscanRecord(unscanRecord)
				log.Std.Info("\t Failed! block height: %d extract failed.", height)
				//saveResult.Success = false
				failed++ //标记保存失败数
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				log.Std.Info("\tdone = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblockHeight uint64, eBlockHash string, mTxs []*BlockTX, eProducer chan ExtractResult) {
		for _, mTx := range mTxs {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlockHeight uint64, mTx *BlockTX, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTx)
				//释放
				<-end

			}(eblockHeight, mTx, bs.extractingCH, eProducer)
		}
	}

	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(blockHeight, worker)

	//独立线程运行生产
	go extractWork(blockHeight, blockHash, txs, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return errors.New("SaveTxToWalletDB failed")
	}

	return nil
}

//ExtractTransaction 提取交易单
func (bs *FabricBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, mTx *BlockTX) ExtractResult {

	var (
		recharges = make([]*openwallet.Recharge, 0)
		success   = false
		resaon    = ""
		accountID = ""
	)

	payloadSpec, err := bs.wm.GetBlockPayload(base64.StdEncoding.EncodeToString(mTx.Payload))
	if err != nil {
		log.Std.Error("block scanner can not extract transaction data; unexpected error: %v", err)
		//记录哪个区块哪个交易单没有完成扫描
		success = false
		resaon = err.Error()
		//return nil, failedTx, nil
	} else {
		addr := payloadSpec.From //payloadSpec.From, payloadSpec.To,
		// log.Info("Find TX addr:", addr, "txid:", mTx.Txid, "blockheight:", blockHeight, "blockhash:", blockHash)

		if wallet, ok := bs.GetWalletByAddress(addr); ok {
			a := wallet.GetAddress(addr)
			if a == nil {
				// ?500
				accountID = a.AccountID
			}
		}

		recharge := openwallet.Recharge{
			TxID:      mTx.Txid,
			Address:   addr,
			AccountID: accountID,
			Symbol:    Symbol,
			Index:     0,
			Amount:    string(payloadSpec.Amount),
			Sid:       base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("%s_%d_%s", mTx.Txid, 0, addr)))),
			CreateAt:  time.Now().Unix(),
		}
		recharges = append(recharges, &recharge)
		success = true
	}

	result := ExtractResult{
		BlockHeight: blockHeight,
		TxID:        mTx.Txid,
		Recharges:   recharges,
		Success:     success,
		Reason:      resaon,
	}

	return result

}

//extractRuntime 提取运行时
func (bs *FabricBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) error {

	var (
		values = make([]ExtractResult, 0)
	)

	for {
		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			//log.Std.Info("block scanner have been scanned!")
			return nil
		default:

			//当数据队列有数据时，释放顶部，传输给消费者
			if len(values) > 0 {
				worker <- values[0]
				values = values[1:]
			}
		}
	}

	// return nil
}

// -----------------------------------------------------------------------
//newBlockNotify 获得新区块后，通知给观测者
func (bs *FabricBlockScanner) newBlockNotify(block *Block) {
	for o, _ := range bs.observers {
		_ = o
		// o.BlockScanNotify(block)
	}
}

//rescanFailedRecord 重扫失败记录
func (bs *FabricBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.GetUnscanRecords()
	if err != nil {
		log.Std.Error("block scanner can not get rescan data; unexpected error: %v", err)
	}

	//组合成批处理
	for _, r := range list {

		//先删除重扫次数超过最大数的记录，一般这种记录可能已经不存在交易池了

		if _, exist := blockMap[r.BlockHeight]; !exist {
			blockMap[r.BlockHeight] = make([]string, 0)
		}

		if len(r.TxID) > 0 {
			arr := blockMap[r.BlockHeight]
			arr = append(arr, r.TxID)

			blockMap[r.BlockHeight] = arr
		}
	}

	for height, txs := range blockMap {

		var hash string
		var txss []*BlockTX

		log.Std.Info("block scanner rescanning height: %d ...", height)

		if len(txs) == 0 {

			// hash, err := bs.wm.GetBlockHash(height)
			// if err != nil {
			// 	//下一个高度找不到会报异常
			// 	log.Std.Error("block scanner can not get new block hash; unexpected error: %v", err)
			// 	continue
			// }

			block, err := bs.wm.GetBlockContent(height)
			if err != nil {
				log.Std.Error("block scanner can not get new block data; unexpected error: %v", err)
				continue
			}
			hash = block.Hash

			txss = block.Transactions
		}

		err = bs.BatchExtractTransaction(height, hash, txss)
		if err != nil {
			log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.DeleteUnscanRecord(height)
	}

	//删除未没有找到交易记录的重扫记录
	bs.DeleteUnscanRecordNotFindTX()
}

//GetWalletByAddress 获取地址对应的钱包
func (bs *FabricBlockScanner) GetWalletByAddress(address string) (*openwallet.Wallet, bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	account, ok := bs.addressInScanning[address]
	if ok {
		wallet, ok := bs.walletInScanning[account]
		return wallet, ok
	} else {
		return nil, false
	}
}

/*
// Fabric not support
//ScanTxMemPool 扫描交易内存池
func (bs *FabricBlockScanner) ScanTxMemPool() {

	log.Std.Info("block scanner scanning mempool ...")

	//提取未确认的交易单
	txIDsInMemPool, err := bs.GetTxIDsInMemPool()
	if err != nil {
		log.Std.Error("block scanner can not get mempool data; unexpected error: %v", err)
	}

	err = bs.BatchExtractTransaction(0, "", txIDsInMemPool)
	if err != nil {
		log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}
}


//Fabric not support
//RescanUnconfirmRechargeRecord
func (bs *FabricBlockScanner) RescanUnconfirmRechargeRecord() {

	bs.mu.RLock()
	defer bs.mu.RUnlock()

	var (
		txs = make([]string, 0)
	)

	currentTime := time.Now()
	//30分钟过期
	m30, _ := time.ParseDuration("-30m")

	d3, _ := time.ParseDuration("-24h")

	//计算过期时间
	expiredTime := currentTime.Add(m30)

	//计算清理时间
	clearTime := currentTime.Add(d3)

	for _, wallet := range bs.walletInScanning {

		records, err := wallet.GetUnconfrimRecharges(expiredTime.Unix())
		if err != nil {
			return
		}
		//重扫未确认记录
		for _, r := range records {
			//删除过期的
			if r.CreateAt <= clearTime.Unix() {
				r.Delete = true
				wallet.SaveUnreceivedRecharge(r)
			} else {
				txs = append(txs, r.TxID)
			}
		}

		err = bs.BatchExtractTransaction(0, "", txs)
		if err != nil {
			log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}
	}
}
*/
