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

package bitcoin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"path/filepath"
	"strings"
	"time"
)

const (
	blockchainBucket = "blockchain" //区块链数据集合
	//periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20 //并发的扫描线程数

	RPCServerCore     = 1 //RPC服务，bitcoin核心钱包
	RPCServerExplorer = 2 //RPC服务，insight-API
)

//BTCBlockScanner bitcoin的区块链扫描器
type BTCBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
	RPCServer            int
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

//NewBTCBlockScanner 创建区块链扫描器
func NewBTCBlockScanner(wm *WalletManager) *BTCBlockScanner {
	bs := BTCBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0
	bs.RPCServer = RPCServerCore

	//设置扫描任务
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BTCBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return errors.New("block height to rescan must greater than 0.")
	}

	hash, err := bs.wm.GetBlockHash(height)
	if err != nil {
		return err
	}

	bs.wm.SaveLocalNewBlock(height, hash)

	return nil
}

//ScanBlockTask 扫描任务
func (bs *BTCBlockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := bs.GetCurrentBlockHeader()
	if err != nil {
		log.Std.Info("block scanner can not get new block height; unexpected error: %v", err)
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		//获取最大高度
		maxHeight, err := bs.wm.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}

		//是否已到最新高度
		if currentHeight == maxHeight {
			log.Std.Info("block scanner has scanned full chain data. Current height: %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		log.Std.Info("block scanner scanning height: %d ...", currentHeight)

		hash, err := bs.wm.GetBlockHash(currentHeight)
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
			break
		}

		block, err := bs.wm.GetBlock(hash)
		if err != nil {
			log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			log.Std.Info("block height: %d extract failed.", currentHeight)
			continue
		}

		isFork := false

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {

			log.Std.Info("block has been fork on height: %d.", currentHeight)
			log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)

			log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.wm.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.wm.GetLocalBlock(currentHeight)
			if err != nil {
				log.Std.Error("block scanner can not get local block; unexpected error: %v", err)

				//查找core钱包的RPC
				log.Info("block scanner prev block height:", currentHeight)

				prevHash, err := bs.wm.GetBlockHash(currentHeight)
				if err != nil {
					log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

				localBlock, err = bs.wm.GetBlock(prevHash)
				if err != nil {
					log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.wm.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)

			isFork = true

		} else {

			err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
			if err != nil {
				log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.wm.SaveLocalNewBlock(currentHeight, currentHash)
			bs.wm.SaveLocalBlock(block)

			isFork = false
		}

		//通知新区块给观测者，异步处理
		go bs.newBlockNotify(block, isFork)
	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.ScanBlock(i)
	}

	if bs.IsScanMemPool {
		//扫描交易内存池
		bs.ScanTxMemPool()
	}

	//重扫失败区块
	bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *BTCBlockScanner) ScanBlock(height uint64) error {

	log.Std.Info("block scanner scanning height: %d ...", height)

	hash, err := bs.wm.GetBlockHash(height)
	if err != nil {
		//下一个高度找不到会报异常
		log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
		return err
	}

	block, err := bs.wm.GetBlock(hash)
	if err != nil {
		log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		log.Std.Info("block height: %d extract failed.", height)
		return err
	}

	err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
	if err != nil {
		log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block, false)

	return nil
}

//ScanTxMemPool 扫描交易内存池
func (bs *BTCBlockScanner) ScanTxMemPool() {

	log.Std.Info("block scanner scanning mempool ...")

	//提取未确认的交易单
	txIDsInMemPool, err := bs.wm.GetTxIDsInMemPool()
	if err != nil {
		log.Std.Info("block scanner can not get mempool data; unexpected error: %v", err)
	}

	err = bs.BatchExtractTransaction(0, "", txIDsInMemPool)
	if err != nil {
		log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

}

//rescanFailedRecord 重扫失败记录
func (bs *BTCBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.wm.GetUnscanRecords()
	if err != nil {
		log.Std.Info("block scanner can not get rescan data; unexpected error: %v", err)
	}

	//组合成批处理
	for _, r := range list {

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

		log.Std.Info("block scanner rescanning height: %d ...", height)

		if len(txs) == 0 {

			hash, err := bs.wm.GetBlockHash(height)
			if err != nil {
				//下一个高度找不到会报异常
				log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
				continue
			}

			block, err := bs.wm.GetBlock(hash)
			if err != nil {
				log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
				continue
			}

			txs = block.tx
		}

		err = bs.BatchExtractTransaction(height, hash, txs)
		if err != nil {
			log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.wm.DeleteUnscanRecord(height)
	}

	//删除未没有找到交易记录的重扫记录
	bs.wm.DeleteUnscanRecordNotFindTX()
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *BTCBlockScanner) newBlockNotify(block *Block, isFork bool) {
	for o, _ := range bs.Observers {
		header := block.BlockHeader()
		header.Fork = isFork
		o.BlockScanNotify(block.BlockHeader())
	}
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *BTCBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []string) error {

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

				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				//saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
				if notifyErr != nil {
					failed++ //标记保存失败数
					log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "")
				bs.SaveUnscanRecord(unscanRecord)
				log.Std.Info("block height: %d extract failed.", height)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				//log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblockHeight uint64, eBlockHash string, mTxs []string, eProducer chan ExtractResult) {
		for _, txid := range mTxs {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlockHeight uint64, mTxid string, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTxid, bs.GetSourceKeyByAddress)
				//释放
				<-end

			}(eblockHeight, txid, bs.extractingCH, eProducer)
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
		return fmt.Errorf("block scanner saveWork failed")
	} else {
		return nil
	}

	//return nil
}

//extractRuntime 提取运行时
func (bs *BTCBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

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
			return
		default:

			//当数据队列有数据时，释放顶部，传输给消费者
			if len(values) > 0 {
				worker <- values[0]
				values = values[1:]
			}
		}
	}

	return

}

//ExtractTransaction 提取交易单
func (bs *BTCBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, txid string, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {

	var (
		result = ExtractResult{
			BlockHeight: blockHeight,
			TxID:        txid,
			extractData: make(map[string]*openwallet.TxExtractData),
		}
	)

	//log.Std.Debug("block scanner scanning tx: %s ...", txid)
	trx, err := bs.wm.GetTransaction(txid)

	if err != nil {
		log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
		result.Success = false
		return result
	}

	//优先使用传入的高度
	if blockHeight > 0 && trx.BlockHeight == 0 {
		trx.BlockHeight = blockHeight
		trx.BlockHash = blockHash
	}
	result = bs.extractTransaction(trx, scanAddressFunc)
	return result

}

//ExtractTransactionData 提取交易单
func (bs *BTCBlockScanner) extractTransaction(trx *Transaction, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {

	var (
		success = true
		result  = ExtractResult{
			BlockHeight: trx.BlockHeight,
			TxID:        trx.TxID,
			extractData: make(map[string]*openwallet.TxExtractData),
		}
	)

	if trx == nil {
		//记录哪个区块哪个交易单没有完成扫描
		success = false
	} else {

		vin := trx.Vins
		blocktime := trx.Blocktime
		for _, input := range vin {

			if len(input.Coinbase) > 0 {
				//coinbase skip
				success = true
				break
			}

			//如果input中没有地址和数量，需要查上一笔交易的output提取
			if len(input.Addr) == 0 && len(input.Value) == 0 {

				intxid := input.TxID
				vout := input.Vout

				preTx, err := bs.wm.GetTransaction(intxid)
				if err != nil {
					success = false
					break
				} else {
					preVouts := preTx.Vouts
					if len(preVouts) > int(vout) {
						preOut := preVouts[vout]
						input.Addr = preOut.Addr
						input.Value = preOut.Value
						//vinout = append(vinout, output[vout])
						success = true

						//log.Debug("GetTxOut:", output[vout])

					}
				}

			}

		}

		if success {

			//提取出账部分记录
			from, totalSpent := bs.extractTxInput(trx, &result, scanAddressFunc)
			log.Debug("from:", from, "totalSpent:", totalSpent)

			//提取入账部分记录
			to, totalReceived := bs.extractTxOutput(trx, &result, scanAddressFunc)
			log.Debug("to:", to, "totalReceived:", totalReceived)

			for _, extractData := range result.extractData {
				tx := &openwallet.Transaction{
					From: from,
					To:   to,
					Fees: totalSpent.Sub(totalReceived).StringFixed(8),
					Coin: openwallet.Coin{
						Symbol:     bs.wm.Symbol(),
						IsContract: false,
					},
					BlockHash:   trx.BlockHash,
					BlockHeight: trx.BlockHeight,
					TxID:        trx.TxID,
					Decimal:     8,
					ConfirmTime: blocktime,
				}
				wxID := openwallet.GenTransactionWxID(tx)
				tx.WxID = wxID
				extractData.Transaction = tx

				//log.Debug("Transaction:", extractData.Transaction)
			}

		}

		success = true

	}
	result.Success = success

	return result

}

//ExtractTxInput 提取交易单输入部分
func (bs *BTCBlockScanner) extractTxInput(trx *Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanAddressFunc) ([]string, decimal.Decimal) {

	//vin := trx.Get("vin")

	var (
		from        = make([]string, 0)
		totalAmount = decimal.Zero
	)

	createAt := time.Now().Unix()
	for i, output := range trx.Vins {

		//in := vin[i]

		txid := output.TxID
		vout := output.Vout
		//
		//output, err := bs.wm.GetTxOut(txid, vout)
		//if err != nil {
		//	return err
		//}

		amount := output.Value
		addr := output.Addr
		sourceKey, ok := scanAddressFunc(addr)
		if ok {
			input := openwallet.TxInput{}
			input.SourceTxID = txid
			input.SourceIndex = vout
			input.TxID = result.TxID
			input.Address = addr
			//transaction.AccountID = a.AccountID
			input.Amount = amount
			input.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			input.Index = output.N
			input.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, i, addr))))
			input.CreateAt = createAt
			//在哪个区块高度时消费
			input.BlockHeight = trx.BlockHeight
			input.BlockHash = trx.BlockHash

			//transactions = append(transactions, &transaction)

			ed := result.extractData[sourceKey]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[sourceKey] = ed
			}

			ed.TxInputs = append(ed.TxInputs, &input)

		}

		from = append(from, addr)
		dAmount, _ := decimal.NewFromString(amount)
		totalAmount = totalAmount.Add(dAmount)

	}
	return from, totalAmount
}

//ExtractTxInput 提取交易单输入部分
func (bs *BTCBlockScanner) extractTxOutput(trx *Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanAddressFunc) ([]string, decimal.Decimal) {

	var (
		to          = make([]string, 0)
		totalAmount = decimal.Zero
	)

	confirmations := trx.Confirmations
	vout := trx.Vouts
	txid := trx.TxID
	//log.Debug("vout:", vout.Array())
	createAt := time.Now().Unix()
	for _, output := range vout {

		amount := output.Value
		n := output.N
		addr := output.Addr
		sourceKey, ok := scanAddressFunc(addr)
		if ok {

			//a := wallet.GetAddress(addr)
			//if a == nil {
			//	continue
			//}

			outPut := openwallet.TxOutPut{}
			outPut.TxID = txid
			outPut.Address = addr
			//transaction.AccountID = a.AccountID
			outPut.Amount = amount
			outPut.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			outPut.Index = n
			outPut.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("output_%s_%d_%s", txid, n, addr))))

			//保存utxo到扩展字段
			outPut.ExtParam = output.ScriptPubKey
			outPut.CreateAt = createAt
			outPut.BlockHeight = trx.BlockHeight
			outPut.BlockHash = trx.BlockHash
			outPut.Confirm = int64(confirmations)

			//transactions = append(transactions, &transaction)

			ed := result.extractData[sourceKey]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[sourceKey] = ed
			}

			ed.TxOutputs = append(ed.TxOutputs, &outPut)

		}

		to = append(to, addr)
		dAmount, _ := decimal.NewFromString(amount)
		totalAmount = totalAmount.Add(dAmount)

	}

	return to, totalAmount
}

//newExtractDataNotify 发送通知
func (bs *BTCBlockScanner) newExtractDataNotify(height uint64, extractData map[string]*openwallet.TxExtractData) error {

	for o, _ := range bs.Observers {
		for key, data := range extractData {
			err := o.BlockExtractDataNotify(key, data)
			if err != nil {
				log.Error("BlockExtractDataNotify unexpected error:", err)
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
				err = bs.SaveUnscanRecord(unscanRecord)
				if err != nil {
					log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
				}

			}
		}
	}

	return nil
}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func (wm *WalletManager) DeleteUnscanRecordNotFindTX() error {

	//删除找不到交易单
	reason := "[-5]No information available about transaction"

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.All(&list)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	for _, r := range list {
		if strings.HasPrefix(r.Reason, reason) {
			tx.DeleteStruct(r)
		}
	}
	return tx.Commit()
}

//SaveRechargeToWalletDB 保存交易单内的充值记录到钱包数据库
//func (bs *BTCBlockScanner) SaveRechargeToWalletDB(height uint64, list []*openwallet.Recharge) error {
//
//	for _, r := range list {
//
//		//accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
//		wallet, ok := bs.GetWalletByAddress(r.Address)
//		if ok {
//
//			//a := wallet.GetAddress(r.Address)
//			//if a == nil {
//			//	continue
//			//}
//			//
//			//r.AccountID = a.AccountID
//
//			err := wallet.SaveUnreceivedRecharge(r)
//			//如果blockHash没有值，添加到重扫，避免遗留
//			if err != nil || len(r.BlockHash) == 0 {
//
//				//记录未扫区块
//				unscanRecord := NewUnscanRecord(height, r.TxID, "save to wallet failed.")
//				err = bs.SaveUnscanRecord(unscanRecord)
//				if err != nil {
//					log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
//				}
//
//			} else {
//				log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
//			}
//		} else {
//			return errors.New("address in wallet is not found")
//		}
//
//	}
//
//	return nil
//}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *BTCBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash = bs.wm.GetLocalNewBlock()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeight, err = bs.wm.GetBlockHeight()
		if err != nil {

			return nil, err
		}

		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		hash, err = bs.wm.GetBlockHash(blockHeight)
		if err != nil {
			return nil, err
		}
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *BTCBlockScanner) GetScannedBlockHeight() uint64 {
	localHeight, _ := bs.wm.GetLocalNewBlock()
	return localHeight
}

func (bs *BTCBlockScanner) ExtractTransactionData(txid string) (map[string]*openwallet.TxExtractData, error) {
	result := bs.ExtractTransaction(0, "", txid, bs.GetSourceKeyByAddress)
	if !result.Success {
		return nil, fmt.Errorf("extract transaction failed")
	}
	return result.extractData, nil
}

//DropRechargeRecords 清楚钱包的全部充值记录
//func (bs *BTCBlockScanner) DropRechargeRecords(accountID string) error {
//	bs.mu.RLock()
//	defer bs.mu.RUnlock()
//
//	wallet, ok := bs.walletInScanning[accountID]
//	if !ok {
//		errMsg := fmt.Sprintf("accountID: %s wallet is not found", accountID)
//		return errors.New(errMsg)
//	}
//
//	return wallet.DropRecharge()
//}

//DeleteRechargesByHeight 删除某区块高度的充值记录
//func (bs *BTCBlockScanner) DeleteRechargesByHeight(height uint64) error {
//
//	bs.mu.RLock()
//	defer bs.mu.RUnlock()
//
//	for _, wallet := range bs.walletInScanning {
//
//		list, err := wallet.GetRecharges(false, height)
//		if err != nil {
//			return err
//		}
//
//		db, err := wallet.OpenDB()
//		if err != nil {
//			return err
//		}
//
//		tx, err := db.Begin(true)
//		if err != nil {
//			return err
//		}
//
//		for _, r := range list {
//			err = db.DeleteStruct(&r)
//			if err != nil {
//				return err
//			}
//		}
//
//		tx.Commit()
//
//		db.Close()
//	}
//
//	return nil
//}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//if record.BlockHeight == 0 {
	//	return errors.New("unconfirmed transaction do not rescan")
	//}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//GetSourceKeyByAddress 获取地址对应的数据源标识
func (bs *BTCBlockScanner) GetSourceKeyByAddress(address string) (string, bool) {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()

	sourceKey, ok := bs.AddressInScanning[address]
	return sourceKey, ok
}

//GetWalletByAddress 获取地址对应的钱包
//func (bs *BTCBlockScanner) GetWalletByAddress(address string) (*openwallet.Wallet, bool) {
//	bs.mu.RLock()
//	defer bs.mu.RUnlock()
//
//	account, ok := bs.addressInScanning[address]
//	if ok {
//		wallet, ok := bs.walletInScanning[account]
//		return wallet, ok
//
//	} else {
//		return nil, false
//	}
//}

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (uint64, error) {

	if wm.Blockscanner.RPCServer == RPCServerExplorer {
		return wm.getBlockHeightByExplorer()
	} else {
		return wm.getBlockHeightByCore()
	}
}

//getBlockHeightByCore 获取区块链高度
func (wm *WalletManager) getBlockHeightByCore() (uint64, error) {

	result, err := wm.WalletClient.Call("getblockcount", nil)
	if err != nil {
		return 0, err
	}

	return result.Uint(), nil
}

//GetLocalNewBlock 获取本地记录的区块高度和hash
func (wm *WalletManager) GetLocalNewBlock() (uint64, string) {

	var (
		blockHeight uint64 = 0
		blockHash   string = ""
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return 0, ""
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalNewBlock 记录区块高度和hash到本地
func (wm *WalletManager) SaveLocalNewBlock(blockHeight uint64, blockHash string) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)
}

//SaveLocalBlock 记录本地新区块
func (wm *WalletManager) SaveLocalBlock(block *Block) {

	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Save(block)
}

//GetBlockHash 根据区块高度获得区块hash
func (wm *WalletManager) GetBlockHash(height uint64) (string, error) {

	if wm.Blockscanner.RPCServer == RPCServerExplorer {
		return wm.getBlockHashByExplorer(height)
	} else {
		return wm.getBlockHashByCore(height)
	}
}

//getBlockHashByCore 根据区块高度获得区块hash
func (wm *WalletManager) getBlockHashByCore(height uint64) (string, error) {

	request := []interface{}{
		height,
	}

	result, err := wm.WalletClient.Call("getblockhash", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

//GetLocalBlock 获取本地区块数据
func (wm *WalletManager) GetLocalBlock(height uint64) (*Block, error) {

	var (
		block Block
	)

	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.One("Height", height, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

//GetBlock 获取区块数据
func (wm *WalletManager) GetBlock(hash string) (*Block, error) {

	if wm.Blockscanner.RPCServer == RPCServerExplorer {
		return wm.getBlockByExplorer(hash)
	} else {
		return wm.getBlockByCore(hash)
	}
}

//getBlockByCore 获取区块数据
func (wm *WalletManager) getBlockByCore(hash string) (*Block, error) {

	request := []interface{}{
		hash,
	}

	result, err := wm.WalletClient.Call("getblock", request)
	if err != nil {
		return nil, err
	}

	return NewBlock(result), nil
}

//GetTxIDsInMemPool 获取待处理的交易池中的交易单IDs
func (wm *WalletManager) GetTxIDsInMemPool() ([]string, error) {

	var (
		txids = make([]string, 0)
	)

	result, err := wm.WalletClient.Call("getrawmempool", nil)
	if err != nil {
		return nil, err
	}

	if !result.IsArray() {
		return nil, errors.New("no query record")
	}

	for _, txid := range result.Array() {
		txids = append(txids, txid.String())
	}

	return txids, nil
}

//GetTransaction 获取交易单
func (wm *WalletManager) GetTransaction(txid string) (*Transaction, error) {

	//request := []interface{}{
	//	txid,
	//	true,
	//}
	//
	//result, err := wm.WalletClient.Call("getrawtransaction", request)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return result, nil

	if wm.Blockscanner.RPCServer == RPCServerExplorer {
		return wm.getTransactionByExplorer(txid)
	} else {
		return wm.getTransactionByCore(txid)
	}

}

//getTransactionByCore 获取交易单
func (wm *WalletManager) getTransactionByCore(txid string) (*Transaction, error) {

	request := []interface{}{
		txid,
		true,
	}

	result, err := wm.WalletClient.Call("getrawtransaction", request)
	if err != nil {
		return nil, err
	}

	return newTxByCore(result), nil
}

//GetTxOut 获取交易单输出信息，用于追溯交易单输入源头
func (wm *WalletManager) GetTxOut(txid string, vout uint64) (*gjson.Result, error) {

	request := []interface{}{
		txid,
		vout,
	}

	result, err := wm.WalletClient.Call("gettxout", request)
	if err != nil {
		return nil, err
	}

	/*
		{
			"bestblock": "0000000000012164c0fb1f7ac13462211aaaa83856073bf94faf2ea9c6ea193a",
			"confirmations": 8,
			"value": 0.64467249,
			"scriptPubKey": {
				"asm": "OP_DUP OP_HASH160 dbb494b649a48b22bfd6383dca1712cc401cddde OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914dbb494b649a48b22bfd6383dca1712cc401cddde88ac",
				"reqSigs": 1,
				"type": "pubkeyhash",
				"addresses": ["n1Yec3dmXEW4f8B5iJa5EsspNQ4Ar6K3Ek"]
			},
			"coinbase": false

		}
	*/

	return result, nil

}

//获取未扫记录
func (wm *WalletManager) GetUnscanRecords() ([]*UnscanRecord, error) {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (wm *WalletManager) DeleteUnscanRecord(height uint64) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		db.DeleteStruct(r)
	}

	return nil
}

//GetAssetsAccountBalanceByAddress 查询账户相关地址的交易记录
func (bs *BTCBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	if bs.RPCServer != RPCServerExplorer {
		return nil, nil
	}

	addrsBalance := make([]*openwallet.Balance, 0)

	for _, a := range address {
		balance, err := bs.wm.getBalanceByExplorer(a)
		if err != nil {
			return nil, err
		}

		addrsBalance = append(addrsBalance, balance)
	}

	return addrsBalance, nil
}

//GetTokenBalanceByAddress 查询地址token余额列表
func (bs *BTCBlockScanner) GetTokenBalanceByAddress(address ...string) ([]*openwallet.TokenBalance, error) {
	return nil, nil
}

//GetAssetsAccountTransactionsByAddress 查询账户相关地址的交易记录
func (bs *BTCBlockScanner) GetTransactionsByAddress(offset, limit int, coin openwallet.Coin, address ...string) ([]*openwallet.TxExtractData, error) {

	var (
		array = make([]*openwallet.TxExtractData, 0)
	)

	trxs, err := bs.wm.getMultiAddrTransactionsByExplorer(offset, limit, address...)
	if err != nil {
		return nil, err
	}

	key := "account"

	//提取账户相关的交易单
	var scanAddressFunc openwallet.BlockScanAddressFunc = func(findAddr string) (string, bool) {
		for _, a := range address {
			if findAddr == a {
				return key, true
			}
		}
		return "", false
	}

	//要检查一下tx.BlockHeight是否有值

	for _, tx := range trxs {
		result := bs.extractTransaction(tx, scanAddressFunc)
		data := result.extractData
		txExtract := data[key]
		if txExtract != nil {
			array = append(array, txExtract)
		}
	}

	return array, nil
}
