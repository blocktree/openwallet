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
package nebulasio

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

const (
	byHeight int = iota //0
	byHash
)

const (
	blockchainBucket = "blockchain" //区块链数据集合
	//periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 15 //并发的扫描线程数

	//RPCServerCore     = 1 //RPC服务，bitcoin核心钱包
	//RPCServerExplorer = 2 //RPC服务，insight-API
)

//NASBlockScanner nebulasio的区块链扫描器
type NASBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
	//	RPCServer            int
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	extractData map[string]*openwallet.TxExtractData

	//Recharges   []*openwallet.Recharge
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

//NewNASBlockScanner 创建区块链扫描器
func NewNASBlockScanner(wm *WalletManager) *NASBlockScanner {
	bs := NASBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = false //Nas不扫内存池
	bs.RescanLastBlockCount = 0

	//设置扫描任务
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *NASBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height <= 0 {
		return errors.New("block height to rescan must greater than 0.")
	}

	hash, err := bs.wm.GetBlockHashByHeight(height)
	if err != nil {
		return errors.New("The height Don't Get the hash From Blockchain!")
	}

	bs.wm.SaveLocalNewBlock(height, hash)

	return nil
}

//ScanBlockTask 扫描任务
func (bs *NASBlockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := bs.GetCurrentBlockHeader()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block height; unexpected error: %v", err)
		return
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		//获取最大高度
		maxHeight, err := bs.wm.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			bs.wm.Log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}

		//是否已到最新高度
		if currentHeight == maxHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height: %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)

		hash, err := bs.wm.GetBlockHashByHeight(currentHeight)
		if err != nil {
			//下一个高度找不到会报异常
			bs.wm.Log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
			break
		}

		block, err := bs.wm.GetBlockByHash(hash)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			bs.wm.Log.Std.Info("block height: %d extract failed.", currentHeight)
			continue
		}

		isFork := false

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {

			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)

			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录

			//查询本地分叉的区块
			forkBlock, _ := bs.wm.GetLocalBlock(currentHeight - 1)

			bs.wm.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.wm.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)

				//查找core钱包的RPC
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)

				prevHash, err := bs.wm.GetBlockHashByHeight(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

				localBlock, err = bs.wm.GetBlockByHash(prevHash)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.wm.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)

			isFork = true

			if forkBlock != nil {

				//通知分叉区块给观测者，异步处理
				go bs.newBlockNotify(forkBlock, isFork)
			}

		} else {
			err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
			if err != nil {
				bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.wm.SaveLocalNewBlock(currentHeight, currentHash)
			bs.wm.SaveLocalBlock(block)

			isFork = false

			//通知新区块给观测者，异步处理
			go bs.newBlockNotify(block, isFork)
		}

	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.ScanBlock(i)
	}

	if bs.IsScanMemPool {
		//NAS暂不扫描交易内存池
	}

	//重扫失败区块
	bs.RescanFailedRecord()
}

//ScanBlock 扫描指定高度区块
func (bs *NASBlockScanner) ScanBlock(height uint64) error {

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", height)

	hash, err := bs.wm.GetBlockHashByHeight(height)
	if err != nil {
		//下一个高度找不到会报异常
		bs.wm.Log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
		return err
	}

	block, err := bs.wm.GetBlockByHash(hash)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return err
	}

	err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block, false)

	return nil
}

/*
//ScanTxMemPool 扫描交易内存池
func (bs *BTCBlockScanner) ScanTxMemPool() {

	bs.wm.Log.Std.Info("block scanner scanning mempool ...")

	//提取未确认的交易单
	txIDsInMemPool, err := bs.wm.GetTxIDsInMemPool()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get mempool data; unexpected error: %v", err)
	}

	err = bs.BatchExtractTransaction(0, "", txIDsInMemPool)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

}
*/
//rescanFailedRecord 重扫失败记录
func (bs *NASBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.wm.GetUnscanRecords()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get rescan data; unexpected error: %v", err)
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

		if height == 0 {
			continue
		}

		var hash string

		bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", height)

		if len(txs) == 0 {

			hash, err := bs.wm.GetBlockHashByHeight(height)
			if err != nil {
				//下一个高度找不到会报异常
				bs.wm.Log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
				continue
			}

			block, err := bs.wm.GetBlockByHash(hash)
			if err != nil {
				bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
				continue
			}

			txs = block.tx
		}

		err = bs.BatchExtractTransaction(height, hash, txs)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.wm.DeleteUnscanRecord(height)
	}

	//删除未没有找到交易记录的重扫记录
	bs.wm.DeleteUnscanRecordNotFindTX()
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *NASBlockScanner) newBlockNotify(block *Block, isFork bool) {
	for o, _ := range bs.Observers {
		header := block.BlockHeader()
		header.Fork = isFork
		o.BlockScanNotify(block.BlockHeader())
	}
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *NASBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []string) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(txs) //需要完成的总数
	)

	if len(txs) == 0 {
		return nil
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
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "Scan failed!")
				bs.SaveUnscanRecord(unscanRecord)
				bs.wm.Log.Std.Info("block height: %d extract failed.", height)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
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
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTxid, bs.ScanAddressFunc)
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
func (bs *NASBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {

		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			//bs.wm.Log.Std.Info("block scanner have been scanned!")
			return
		case activeWorker <- activeValue:
			//wm.Log.Std.Info("Get %d", len(activeValue))
			values = values[1:]
		}
	}

	return

}

//判断地址是否为合约地址
func CheckIsContract(Toaddr string) bool {
	addr, _ := addressEncoder.Base58Decode(Toaddr, addressEncoder.NewBase58Alphabet("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"))
	if addr[1] == Contract_prefix[1] {
		return true
	}
	//	fmt.Printf("addr=%x,addr(1)=%x,addr[1]=%v\n",addr,addr[0:2],addr[0:2])
	return false
}

//extractTxOutput 提取交易单输入部分,只有一个TxOutPut
func (bs *NASBlockScanner) extractTxOutput(tx *NasTransaction, txExtractData *openwallet.TxExtractData) {

	//主网to交易转账信息,只有一个TxOutPut
	txOutput := &openwallet.TxOutPut{
		ExtParam: "", //扩展参数，用于记录utxo的解锁字段，账户模型中为空
	}
	txOutput.Recharge.Sid = openwallet.GenTxOutPutSID(tx.Hash, bs.wm.Symbol(), "", uint64(0))
	txOutput.Recharge.TxID = tx.Hash
	txOutput.Recharge.Address = tx.To
	txOutput.Recharge.Coin = openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	txOutput.Recharge.Amount = tx.Value.Div(coinDecimal).String()
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Index = 0 //账户模型填0
	txOutput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
}

//extractTxInput 提取交易单输入部分,包含两个TxInput
func (bs *NASBlockScanner) extractTxInput(tx *NasTransaction, txExtractData *openwallet.TxExtractData) {

	//主网from交易转账信息，第一个TxInput
	txInput := &openwallet.TxInput{
		SourceTxID: "", //utxo模型上的上一个交易输入源
	}
	txInput.Recharge.Sid = openwallet.GenTxInputSID(tx.Hash, bs.wm.Symbol(), "", uint64(0))
	txInput.Recharge.TxID = tx.Hash
	txInput.Recharge.Address = tx.From
	txInput.Recharge.Coin = openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	txInput.Recharge.Amount = tx.Value.Div(coinDecimal).String()
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Index = 0 //账户模型填0
	txInput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)

	//主网from交易转账手续费信息，第二个TxInput
	txInputfees := &openwallet.TxInput{
		SourceTxID: "", //utxo模型上的上一个交易输入源
	}
	txInputfees.Recharge.Sid = openwallet.GenTxInputSID(tx.Hash, bs.wm.Symbol(), "", uint64(1))
	txInputfees.Recharge.TxID = tx.Hash
	txInputfees.Recharge.Address = tx.From
	txInputfees.Recharge.Coin = openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	txInputfees.Recharge.Amount = decimal.RequireFromString(tx.Gas_used).Div(coinDecimal).String()
	txInputfees.Recharge.BlockHash = tx.BlockHash
	txInputfees.Recharge.BlockHeight = tx.BlockHeight
	txInputfees.Recharge.Index = 0 //账户模型填0
	txInputfees.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInputfees)
}

func (bs *NASBlockScanner) InitNasExtractResult(tx *NasTransaction, result *ExtractResult, isFromAccount bool) {
	amount := tx.Value.Div(coinDecimal).StringFixed(bs.wm.Decimal())
	txExtractData := &openwallet.TxExtractData{}
	transx := &openwallet.Transaction{
		Fees: decimal.RequireFromString(tx.Gas_used).Mul(decimal.RequireFromString(tx.Gas_price)).Div(coinDecimal).StringFixed(bs.wm.Decimal()),
		Coin: openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		},
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.Hash,
		Decimal:     18,
		Amount:      amount,
		ConfirmTime: int64(tx.BlockTime),
	}
	submitTime, _ := strconv.ParseInt(tx.Timestamp, 10, 64)
	transx.SubmitTime = submitTime
	transx.From = append(transx.From, tx.From+":"+amount)
	transx.To = append(transx.To, tx.To+":"+amount)
	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID
	txExtractData.Transaction = transx

	if isFromAccount {
		bs.extractTxInput(tx, txExtractData)
		result.extractData[tx.FromAccountId] = txExtractData
	} else {
		bs.extractTxOutput(tx, txExtractData)
		result.extractData[tx.ToAccountId] = txExtractData
	}
}

/*func (bs *NASBlockScanner) InitNasExtractResult(tx *NasTransaction, result *ExtractResult,isFromAccount bool) {
	txExtractData := &openwallet.TxExtractData{}
	transx := &openwallet.Transaction{
		Fees: decimal.RequireFromString(tx.Gas_used).Div(coinDecimal).String() ,
		Coin: openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		},
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.Hash,
		Decimal:     18,
		Amount:		 tx.Value.Div(coinDecimal).String(),
		ConfirmTime:  int64(tx.BlockTime),
	}

	submitTime ,_ := strconv.ParseInt(tx.Timestamp,10,64)
	transx.SubmitTime = submitTime
	transx.From = append(transx.From, tx.From)
	transx.To = append(transx.To, tx.To)
	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID
	txExtractData.Transaction = transx
	if isFromAccount {
		//input
		txInput := &openwallet.TxInput{
			SourceTxID:  "",  //utxo模型上的上一个交易输入源
			SourceIndex: "",	//utxo模型上的上一个交易输入源
		}
		txInput.Recharge.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, 0, tx.From))))
		txInput.Recharge.TxID = tx.Hash
		txInput.Recharge.Address = tx.From
		txInput.Recharge.Coin = openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		}
		txInput.Recharge.Amount = tx.Value.Div(coinDecimal).String()
		txInput.Recharge.BlockHash = tx.BlockHash
		txInput.Recharge.BlockHeight = tx.BlockHeight
		txInput.Recharge.Index = 0 //账户模型填0
		txInput.Recharge.CreateAt = time.Now().Unix()
		txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)

		result.extractData[tx.FromAccountId] = txExtractData
	} else {
		//output
		txOutput := &openwallet.TxOutPut{
			ExtParam: "", //扩展参数，用于记录utxo的解锁字段，账户模型中为空
		}
		txOutput.Recharge.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, 0, tx.From))))
		txOutput.Recharge.TxID = tx.Hash
		txOutput.Recharge.Address = tx.To
		txOutput.Recharge.Coin = openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		}
		txOutput.Recharge.Amount = tx.Value.Div(coinDecimal).String()
		txOutput.Recharge.BlockHash = tx.BlockHash
		txOutput.Recharge.BlockHeight = tx.BlockHeight
		txOutput.Recharge.Index = 0 //账户模型填0
		txOutput.Recharge.CreateAt = time.Now().Unix()
		txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)

		result.extractData[tx.ToAccountId] = txExtractData
	}
}*/

//ExtractTransaction 提取交易单
func (bs *NASBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, txid string, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {

	var (
		//transactions = make([]*openwallet.Recharge, 0)
		success = true
		result  = ExtractResult{
			BlockHeight: blockHeight,
			TxID:        txid,
			extractData: make(map[string]*openwallet.TxExtractData),
		}
	)
	//bs.wm.Log.Std.Debug("block scanner scanning tx: %s ...", txid)

	tx_nas, err := bs.wm.GetTransaction(txid, blockHeight)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
		//记录哪个区块哪个交易单没有完成扫描
		success = false
		//return nil, failedTx, nil
	} else {
		//bs.wm.Log.Std.Info("block scanner scanning tx: %+v", txid)
		//订阅地址为交易单中的发送者
		if _, ok := scanAddressFunc(tx_nas.From); ok {
			bs.wm.Log.Std.Info("tx.from found in transaction [%v] .", tx_nas.Hash)
			if accountId, exist := scanAddressFunc(tx_nas.From); exist {
				tx_nas.FromAccountId = accountId
				bs.InitNasExtractResult(tx_nas, &result, true)
			} else {
				bs.wm.Log.Std.Info("tx.from unexpected error.")
			}
		} else {
			//bs.wm.Log.Std.Info("tx.from[%v] not found in scanning address.", tx_nas.From)
		}

		//订阅地址为交易单中的接收者
		if _, ok2 := scanAddressFunc(tx_nas.To); ok2 {
			bs.wm.Log.Std.Info("tx.to found in transaction [%v].", tx_nas.Hash)
			if accountId, exist := scanAddressFunc(tx_nas.To); exist {
				if _, exist = result.extractData[accountId]; !exist {
					tx_nas.ToAccountId = accountId
					bs.InitNasExtractResult(tx_nas, &result, false)
				}

			} else {
				bs.wm.Log.Std.Info("tx.to unexpected error.")
			}

		} else if len(result.extractData) == 0 {
			//bs.wm.Log.Std.Info("tx.to[%v] not found in scanning address.", tx_nas.To)
		}
		success = true
	}

	result.Success = success

	return result
}

/*
//ExtractTxInput 提取交易单输入部分
func (bs *BTCBlockScanner) ExtractTxInput(blockHeight uint64, blockHash string, trx *Transaction, result *ExtractResult) ([]string, decimal.Decimal) {

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
		sourceKey, ok := bs.GetSourceKeyByAddress(addr)
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
			if blockHeight > 0 {
				//在哪个区块高度时消费
				input.BlockHeight = blockHeight
				input.BlockHash = blockHash
				//outPut.Confirm = confirmations
			}

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
func (bs *BTCBlockScanner) ExtractTxOutput(blockHeight uint64, blockHash string, trx *Transaction, result *ExtractResult) ([]string, decimal.Decimal) {

	var (
		to          = make([]string, 0)
		totalAmount = decimal.Zero
	)

	confirmations := trx.Confirmations
	vout := trx.Vouts
	txid := trx.TxID
	//bs.wm.Log.Debug("vout:", vout.Array())
	createAt := time.Now().Unix()
	for _, output := range vout {

		amount := output.Value
		n := output.N
		addr := output.Addr
		sourceKey, ok := bs.GetSourceKeyByAddress(addr)
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
			if blockHeight > 0 {
				outPut.BlockHeight = blockHeight
				outPut.BlockHash = blockHash
				outPut.Confirm = int64(confirmations)
			}

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
*/
//newExtractDataNotify 发送通知
func (bs *NASBlockScanner) newExtractDataNotify(height uint64, extractData map[string]*openwallet.TxExtractData) error {

	for o, _ := range bs.Observers {
		for key, data := range extractData {
			err := o.BlockExtractDataNotify(key, data)
			if err != nil {
				bs.wm.Log.Error("BlockExtractDataNotify unexpected error:", err)
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
				err = bs.SaveUnscanRecord(unscanRecord)
				if err != nil {
					bs.wm.Log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
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
//					bs.wm.Log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
//				}
//
//			} else {
//				bs.wm.Log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
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
func (bs *NASBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

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

		hash, err = bs.wm.GetBlockHashByHeight(blockHeight)
		if err != nil {
			return nil, err
		}
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *NASBlockScanner) GetScannedBlockHeight() uint64 {
	localHeight, _ := bs.wm.GetLocalNewBlock()
	return localHeight
}

//ExtractTransactionData 扫描一笔交易
func (bs *NASBlockScanner) ExtractTransactionData(txid string, scanAddressFunc openwallet.BlockScanAddressFunc) (map[string][]*openwallet.TxExtractData, error) {

	result := bs.ExtractTransaction(0, "", txid, scanAddressFunc)
	if !result.Success {
		return nil, fmt.Errorf("extract transaction failed")
	}

	extData := make(map[string][]*openwallet.TxExtractData)
	for key, data := range result.extractData {
		txs := extData[key]
		if txs == nil {
			txs = make([]*openwallet.TxExtractData, 0)
		}
		txs = append(txs, data)
		extData[key] = txs
	}
	return extData, nil
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
func (bs *NASBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	if record.BlockHeight == 0 {
		bs.wm.Log.Warn("unconfirmed transaction do not rescan")
		return nil
	}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//GetSourceKeyByAddress 获取地址对应的数据源标识
//func (bs *NASBlockScanner) GetSourceKeyByAddress(address string) (string, bool) {
//	bs.Mu.RLock()
//	defer bs.Mu.RUnlock()
//
//	sourceKey, ok := bs.AddressInScanning[address]
//	return sourceKey, ok
//}

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
func (wm *WalletManager) SaveLocalNewBlock(blockHeight uint64, blockHash string) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return errors.New(fmt.Sprintf("Open dbPath BlockchainFile Fail!"))
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)

	return nil
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

//GetTransaction 获取交易单
func (wm *WalletManager) GetTransaction(txid string, height uint64) (*NasTransaction, error) {
	Transaction_result, err := wm.WalletClient.CallGetTransactionReceipt(txid)
	if err != nil {
		return nil, err
	}

	height_string := strconv.FormatUint(height, 10)
	block, err := wm.GetBlockByHeight(height_string)
	if err != nil {
		return nil, err
	}

	return newNasTransaction(Transaction_result, block), nil
}

/*
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
*/
/*
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

/*
	return result, nil

}
*/
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
func (bs *NASBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	addrsBalance := make([]*openwallet.Balance, 0)

	for _, a := range address {
		balance, err := bs.wm.BsGetBalanceByAddress(a)
		if err != nil {
			return nil, err
		}
		addrsBalance = append(addrsBalance, balance)
	}

	return addrsBalance, nil
}

//实现BlockScanNotificationObject interface下的方法
type subscriber struct {
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriber) BlockScanNotify(header *openwallet.BlockHeader) error {
	//fmt.Printf("header:%+v\n", header)
	return nil
}

//BlockExtractDataNotify 区块提取结果通知
func (sub *subscriber) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	fmt.Printf("account:%+v\n", sourceKey)

	for _, TxInput := range data.TxInputs {
		fmt.Printf("TxInput=%+v\n", TxInput)
	}
	for _, TxOutput := range data.TxOutputs {
		fmt.Printf("TxOutput=%+v\n", TxOutput)
	}
	//	fmt.Printf("data.TxOutputs=%+v\n", data.TxOutputs)
	fmt.Printf("data.Transaction=%+v\n", data.Transaction)
	return nil
}

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (uint64, error) {

	result, err := wm.WalletClient.CallGetnebstate("height")
	if err != nil {
		return 0, err
	}

	height, _ := strconv.ParseUint(result.String(), 10, 64)
	return height, nil
}

//GetBlockHashByHeight 根据区块高度获得区块hash
func (wm *WalletManager) GetBlockHashByHeight(height uint64) (string, error) {

	height_string := strconv.FormatUint(height, 10)
	result, err := wm.WalletClient.CallgetBlockByHeightOrHash(height_string, byHeight)
	if err != nil {
		fmt.Printf("err=%v\n", err)
		return "", err
	}

	hash := gjson.Get(result.String(), "hash")

	return hash.String(), nil
}

//GetBlockByHeight 根据区块高度获得区块信息
func (wm *WalletManager) GetBlockByHeight(height string) (*Block, error) {

	result, err := wm.WalletClient.CallgetBlockByHeightOrHash(height, byHeight)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("block=%v\n",NewBlock(result))
	return NewBlock(result), nil
}

//GetBlockByHash 根据区块hash获得区块信息
func (wm *WalletManager) GetBlockByHash(hash string) (*Block, error) {

	result, err := wm.WalletClient.CallgetBlockByHeightOrHash(hash, byHash)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("block_String=%v\n",result.String())
	return NewBlock(result), nil
}
