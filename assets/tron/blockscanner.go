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
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/imroc/req"
	"github.com/shopspring/decimal"
)

const (
	blockchainBucket = "blockchain" //区块链数据集合
	// periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20 //并发的扫描线程数

	//RPCServerCore RPC服务，核心钱包
	RPCServerCore = 0
	//RPCServerExplorer RPC服务，insight-API
	RPCServerExplorer = 1
)

var (
	transferContract      = "TransferContract"
	transferAssetContract = "TransferAssetContract"
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

//NewTronBlockScanner 创建区块链扫描器
func NewTronBlockScanner(wm *WalletManager) *TronBlockScanner {
	bs := TronBlockScanner{BlockScannerBase: openwallet.NewBlockScannerBase()}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	//bs.IsScanMemPool = false //Tron不扫内存池
	bs.RescanLastBlockCount = 0

	// bs.walletInScanning = make(map[string]*openwallet.Wallet)
	// bs.addressInScanning = make(map[string]string)
	// bs.observers = make(map[openwallet.BlockScanNotificationObject]bool)

	//设置扫描任务
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *TronBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		bs.wm.Log.Std.Info("block height to rescan must greater than 0")
		return fmt.Errorf("block height to rescan must greater than 0")
	}
	block, err := bs.wm.GetBlockByNum(height)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get block by height;unexpected error:%v", err)
		return err
	}
	hash := block.Hash
	bs.SaveLocalNewBlock(height, hash)

	return nil
}

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
		bs.wm.Log.Std.Info("No records found in local, get current block as the local!")

		block, err := bs.wm.GetNowBlock()
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get current block;unexpected error:%v", err)
		}

		// 取上一个块作为初始
		block, err = bs.wm.GetBlockByNum(block.Height - 1)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get block by height;unexpected error:%v", err)
		}

		currentHash = block.GetBlockHashID()
		currentHeight = block.GetHeight()
	}
	for {
		//获取最大高度
		maxHeightBlock, err := bs.wm.GetNowBlock()
		if err != nil {
			//下一个高度找不到会报异常
			bs.wm.Log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}
		//maxHeightBlockHash := maxHeightBlock.Hash
		maxHeight := maxHeightBlock.Height
		//是否已到最新高度
		if currentHeight == maxHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		block, err := bs.wm.GetBlockByNum(currentHeight)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			continue
		}
		hash := block.GetBlockHashID()
		isFork := false

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {
			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)
			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)
			//查询本地分叉的区块
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)
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
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)
				//查找core钱包的RPC
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)
				localBlock, err = bs.wm.GetBlockByNum(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}
			}
			//重置当前区块的hash
			currentHash = localBlock.Hash

			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)

			isFork = true
			if forkBlock != nil {
				//通知分叉区块给观测者，异步处理
				go bs.newBlockNotify(forkBlock, isFork)
			}

		} else {

			txHash := make([]string, len(block.tx))
			for i, _ := range block.tx {
				txHash[i] = block.tx[i].TxID
			}
			err = bs.BatchExtractTransaction(block.Height, block.Hash, txHash)
			if err != nil {
				//bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}
			//重置当前区块的hash
			currentHash = hash
			//保存本地新高度
			bs.SaveLocalNewBlock(currentHeight, currentHash)
			bs.SaveLocalBlock(block)
			isFork = false
			//通知新区块给观测者，异步处理
			go bs.newBlockNotify(block, isFork)

		}
	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.ScanBlock(i)
	}

	//重扫失败区块
	bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *TronBlockScanner) ScanBlock(height uint64) error {

	block, err := bs.wm.GetBlockByNum(height)
	if err != nil {
		//下一个高度找不到会报异常
		bs.wm.Log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
		return err
	}
	block, err = bs.wm.GetBlockByID(block.Hash)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		//log.Std.Info("block height: %d extract failed.", height)
		return err
	}

	bs.scanBlock(block)

	return nil
}

func (bs *TronBlockScanner) scanBlock(block *Block) error {

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", block.Height)
	txHash := make([]string, len(block.tx))
	for i, _ := range block.tx {
		txHash[i] = block.tx[i].TxID
	}
	err := bs.BatchExtractTransaction(block.Height, block.Hash, txHash)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block, false)

	return nil
}

//rescanFailedRecord 重扫失败记录
func (bs *TronBlockScanner) RescanFailedRecord() {

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

		var hash string
		bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", height)
		if len(txs) == 0 {
			block, err := bs.wm.GetBlockByNum(height)
			if err != nil {
				//下一个高度找不到会报异常
				bs.wm.Log.Std.Info("block scanner can not get new block hash; unexpected error: %v", err)
				continue
			}

			block, err = bs.wm.GetBlockByID(block.Hash)
			if err != nil {
				bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
				continue
			}
			txs = make([]string, len(block.tx))
			for i, _ := range block.tx {
				txs[i] = block.tx[i].TxID
			}
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

//newBlockNotify 获得新区块后，通知给观测者
func (bs *TronBlockScanner) newBlockNotify(block *Block, isFork bool) {
	for o, _ := range bs.Observers {
		header := block.Blockheader()
		header.Fork = isFork
		o.BlockScanNotify(block.Blockheader())
	}
}

//提取交易单
func (bs *TronBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, txid string, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {
	var (
		success = true
		result  = ExtractResult{
			BlockHeight: blockHeight,
			TxID:        txid,
			extractData: make(map[string]*openwallet.TxExtractData),
		}
	)
	trx, err := bs.wm.GetTransaction(txid, blockHeight)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extract transaction data,unexpected error:%v", err)
		success = false
		return result
	}
	switch trx.Type {
	case transferContract: //TRX
		//bs.wm.Log.Std.Info("block scanner scanning tx: %+v", txid)
		//订阅地址为交易单中的发送者
		_, ok1 := scanAddressFunc(trx.From)
		if ok1 {
			bs.wm.Log.Std.Info("tx.from found in transaction [%v] .", trx.TxID)
			if accountId, exist := scanAddressFunc(trx.From); exist {
				trx.FromAccountId = accountId
				bs.InitTronExtractResult(trx, &result, true)
			} else {
				bs.wm.Log.Std.Info("tx.from unexpected error.")
			}
		} else {
			//bs.wm.Log.Std.Info("tx.from[%v] not found in scanning address.", trx.From)
		}
		//订阅地址为交易单中的接收者
		_, ok2 := scanAddressFunc(trx.To)
		if ok2 {
			bs.wm.Log.Std.Info("tx.to found in transaction [%v].", trx.TxID)
			if accountId, exist := scanAddressFunc(trx.To); exist {
				if _, exist = result.extractData[accountId]; !exist {
					trx.ToAccountId = accountId
					bs.InitTronExtractResult(trx, &result, false)
				}
			} else {
				bs.wm.Log.Std.Info("tx.to unexpected error.")
			}
		} else if len(result.extractData) == 0 {
			//bs.wm.Log.Std.Info("tx.to[%v] not found in scanning address.", tx_nas.To)
		}
		success = true
	case transferAssetContract: //other asset,TODO
		success = false
	default:
		success = false
	}
	result.Success = success
	return result

}

func (bs *TronBlockScanner) InitTronExtractResult(tx *Transaction, result *ExtractResult, isFromAccount bool) {
	value := decimal.RequireFromString(tx.Amount)
	amount := value.Div(coinDecimal).StringFixed(bs.wm.Decimal())
	txExtractData := &openwallet.TxExtractData{}
	transx := &openwallet.Transaction{
		Fees: "0",
		Coin: openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		},
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.TxID,
		Decimal:     6,
		Amount:      amount,
		ConfirmTime: int64(tx.Blocktime),
	}
	//transx.SubmitTime = int64(tx.Blocktime)
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

//extractTxInput 提取交易单输入部分,无需手续费，所以只包含1个TxInput
func (bs *TronBlockScanner) extractTxInput(tx *Transaction, txExtractData *openwallet.TxExtractData) {
	value := decimal.RequireFromString(tx.Amount)
	amount := value.Div(coinDecimal).StringFixed(bs.wm.Decimal())
	//主网from交易转账信息，第一个TxInput
	txInput := &openwallet.TxInput{
		SourceTxID: "", //utxo模型上的上一个交易输入源
	}
	txInput.Recharge.Sid = openwallet.GenTxInputSID(tx.TxID, bs.wm.Symbol(), "", uint64(0))
	txInput.Recharge.TxID = tx.TxID
	txInput.Recharge.Address = tx.From
	txInput.Recharge.Coin = openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	txInput.Recharge.Amount = amount
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Index = 0 //账户模型填0
	txInput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)

	/*
		//主网from交易转账手续费信息，第二个TxInput
		txInputfees := &openwallet.TxInput{
			SourceTxID: "", //utxo模型上的上一个交易输入源
		}
		txInputfees.Recharge.Sid = openwallet.GenTxInputSID(tx.TxID, bs.wm.Symbol(), "", uint64(1))
		txInputfees.Recharge.TxID = tx.TxID
		txInputfees.Recharge.Address = tx.From
		txInputfees.Recharge.Coin = openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		}
		txInputfees.Recharge.Amount = "0"
		txInputfees.Recharge.BlockHash = tx.BlockHash
		txInputfees.Recharge.BlockHeight = tx.BlockHeight
		txInputfees.Recharge.Index = 0 //账户模型填0
		txInputfees.Recharge.CreateAt = time.Now().Unix()
		txExtractData.TxInputs = append(txExtractData.TxInputs, txInputfees)
	*/
}

//extractTxOutput 提取交易单输入部分,只有一个TxOutPut
func (bs *TronBlockScanner) extractTxOutput(tx *Transaction, txExtractData *openwallet.TxExtractData) {
	value := decimal.RequireFromString(tx.Amount)
	amount := value.Div(coinDecimal).StringFixed(bs.wm.Decimal())
	//主网to交易转账信息,只有一个TxOutPut
	txOutput := &openwallet.TxOutPut{
		ExtParam: "", //扩展参数，用于记录utxo的解锁字段，账户模型中为空
	}
	txOutput.Recharge.Sid = openwallet.GenTxOutPutSID(tx.TxID, bs.wm.Symbol(), "", uint64(0))
	txOutput.Recharge.TxID = tx.TxID
	txOutput.Recharge.Address = tx.To
	txOutput.Recharge.Coin = openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	txOutput.Recharge.Amount = amount /*tx.Value.Div(coinDecimal).String()*/
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Index = 0 //账户模型填0
	txOutput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
}

//发送通知
func (bs *TronBlockScanner) newExtractDataNotify(height uint64, extractData map[string]*openwallet.TxExtractData) error {
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

func (wm *WalletManager) GetTransaction(txid string, height uint64) (*Transaction, error) {
	params := req.Param{"value": txid}
	r, err := wm.WalletClient.Call("/wallet/gettransactionbyid", params)
	if err != nil {
		return nil, err
	}
	tx := NewTransaction(r)
	tx.BlockHeight = height
	block, err := wm.GetBlockByNum(height)
	if err != nil {
		return nil, err
	}
	tx.BlockHash = block.Hash
	return tx, nil
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

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *TronBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, txs []string) error {

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
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "")
				bs.SaveUnscanRecord(unscanRecord)
				//log.Std.Info("block height: %d extract failed.", height)
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
func (bs *TronBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

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
			//log.Std.Info("block scanner have been scanned!")
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}
	//return
}

/*
 //SaveTxToWalletDB 保存交易记录到钱包数据库
 func (bs *TronBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

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
*/

func (bs *TronBlockScanner) GetGlobalMaxBlockHeight() uint64 {

	currentBlock, err := bs.wm.GetCurrentBlock()
	if err != nil {
		bs.wm.Log.Std.Info("get global max block height error;unexpected error:%v", err)
		return 0
	}
	blockHeight := currentBlock.Height
	return blockHeight
}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *TronBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
	)
	blockHeight, hash = bs.GetLocalNewBlock()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		currentBlock, err := bs.wm.GetNowBlock()
		if err != nil {
			bs.wm.Log.Std.Info("get current block header error;unexpected error:%v", err)
			return nil, err
		}
		blockHeight = currentBlock.Height
		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		block, err := bs.wm.GetBlockByNum(blockHeight)
		if err != nil {
			bs.wm.Log.Std.Info("get global max block height error;unexpected error:%v", err)
			return nil, err
		}
		hash = block.Hash
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *TronBlockScanner) GetScannedBlockHeight() uint64 {
	localHeight, _ := bs.GetLocalNewBlock()
	return localHeight
}

//GetAssetsAccountBalanceByAddress 查询账户相关地址的交易记录
func (bs *TronBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

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

//GetSourceKeyByAddress 获取地址对应的数据源标识
func (bs *TronBlockScanner) GetSourceKeyByAddress(address string) (string, bool) {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()
	sourceKey, ok := bs.AddressInScanning[address]
	return sourceKey, ok
}

//Run 运行
func (bs *TronBlockScanner) Run() error {
	bs.BlockScannerBase.Run()
	return nil
}

////Stop 停止扫描
func (bs *TronBlockScanner) Stop() error {
	bs.BlockScannerBase.Stop()
	return nil
}

//Restart 继续扫描
func (bs *TronBlockScanner) Restart() error {
	bs.BlockScannerBase.Restart()
	return nil
}

/******************* 使用insight socket.io 监听区块 *******************/

//setupSocketIO 配置socketIO监听新区块
func (bs *TronBlockScanner) setupSocketIO() error {

	bs.wm.Log.Std.Info("block scanner use socketIO to listen new data")
	var (
		room = "inv"
	)

	if bs.socketIO == nil {

		apiUrl, err := url.Parse(bs.wm.Config.ServerAPI)
		if err != nil {
			return err
		}
		domain := apiUrl.Hostname()
		port := common.NewString(apiUrl.Port()).Int()
		c, err := gosocketio.Dial(
			gosocketio.GetUrl(domain, port, false),
			transport.GetDefaultWebsocketTransport())
		if err != nil {
			return err
		}
		bs.socketIO = c
	}

	err := bs.socketIO.On("tx", func(h *gosocketio.Channel, args interface{}) {
		txMap, ok := args.(map[string]interface{})
		if ok {
			txid := txMap["txid"].(string)
			errInner := bs.BatchExtractTransaction(0, "", []string{txid})
			if errInner != nil {
				bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", errInner)
			}
		}

	})
	if err != nil {
		return err
	}

	/*
		 err = bs.socketIO.On("block", func(h *gosocketio.Channel, args interface{}) {
			 log.Info("block scanner socketIO get new block received: ", args)
			 hash, ok := args.(string)
			 if ok {

				 block, errInner := bs.wm.GetBlock(hash)
				 if errInner != nil {
					 log.Std.Info("block scanner can not get new block data; unexpected error: %v", errInner)
				 }

				 errInner = bs.scanBlock(block)
				 if errInner != nil {
					 log.Std.Info("block scanner can not block: %d; unexpected error: %v", block.Height, errInner)
				 }
			 }

		 })
		 if err != nil {
			 return err
		 }
	*/

	err = bs.socketIO.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		bs.wm.Log.Std.Info("block scanner socketIO disconnected")
	})
	if err != nil {
		return err
	}

	err = bs.socketIO.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		bs.wm.Log.Std.Info("block scanner socketIO connected")
		h.Emit("subscribe", room)
	})
	if err != nil {
		return err
	}

	return nil
}
