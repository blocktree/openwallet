/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package tron

import (
	"fmt"
	"github.com/shopspring/decimal"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/graarh/golang-socketio"
	"github.com/imroc/req"
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
	extractData map[string][]*openwallet.TxExtractData
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
			return
		}

		// 取上一个块作为初始
		block, err = bs.wm.GetBlockByNum(block.Height - 1)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get block by height;unexpected error:%v", err)
			return
		}

		currentHash = block.GetBlockHashID()
		currentHeight = block.GetHeight()
	}
	for {

		if !bs.Scanning {
			//区块扫描器已暂停，马上结束本次任务
			return
		}

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
		if currentHeight >= maxHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)

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
				bs.newBlockNotify(forkBlock, isFork)
			}

		} else {

			err = bs.BatchExtractTransaction(block.Height, block.Hash, block.Time, block.tx)
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
			bs.newBlockNotify(block, isFork)

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
	err := bs.BatchExtractTransaction(block.Height, block.Hash, block.Time, block.tx)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block, false)

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
		var blocktime int64
		trxs := make([]*Transaction, 0)
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
			blocktime = block.Time
			trxs = block.tx
		}
		err = bs.BatchExtractTransaction(height, hash, blocktime, trxs)
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
	header := block.Blockheader()
	header.Fork = isFork
	bs.NewBlockNotify(header)
}

//提取交易单
func (bs *TronBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, blockTime int64, trx *Transaction, scanAddressFunc openwallet.BlockScanAddressFunc) ExtractResult {
	var (
		success = true
		result  = ExtractResult{
			BlockHeight: blockHeight,
			TxID:        trx.TxID,
			extractData: make(map[string][]*openwallet.TxExtractData),
		}
	)

	//提出交易单明细
	for _, contractTRX := range trx.Contract {

		//bs.wm.Log.Std.Info("block scanner scanning tx: %+v", txid)
		//订阅地址为交易单中的发送者
		accountId, ok1 := scanAddressFunc(contractTRX.From)
		//订阅地址为交易单中的接收者
		accountId2, ok2 := scanAddressFunc(contractTRX.To)

		//相同账户
		if accountId == accountId2 && len(accountId) > 0 && len(accountId2) > 0 {
			contractTRX.SourceKey = accountId
			bs.InitTronExtractResult(contractTRX, &result, 0)
		} else {
			if ok1 {
				contractTRX.SourceKey = accountId
				bs.InitTronExtractResult(contractTRX, &result, 1)
			}

			if ok2 {
				contractTRX.SourceKey = accountId2
				bs.InitTronExtractResult(contractTRX, &result, 2)
			}
		}

		success = true

	}

	result.Success = success
	return result

}

//InitTronExtractResult operate = 0: 输入输出提取，1: 输入提取，2：输出提取
func (bs *TronBlockScanner) InitTronExtractResult(tx *Contract, result *ExtractResult, operate int64) {

	txExtractDataArray := result.extractData[tx.SourceKey]
	if txExtractDataArray == nil {
		txExtractDataArray = make([]*openwallet.TxExtractData, 0)
	}

	txExtractData := &openwallet.TxExtractData{}

	status := "1"
	reason := ""
	if tx.ContractRet != SUCCESS {
		status = "0"
		reason = tx.ContractRet
	}
	amount := decimal.Zero
	coin := openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	if len(tx.ContractAddress) > 0 {
		contractId := openwallet.GenContractID(bs.wm.Symbol(), tx.ContractAddress)
		coin.ContractID = contractId
		coin.IsContract = true
		coin.Contract = openwallet.SmartContract{
			ContractID: contractId,
			Address:    tx.ContractAddress,
			Symbol:     bs.wm.Symbol(),
			Protocol:   tx.Protocol,
		}
		amount = common.IntToDecimals(tx.Amount, 0)
	} else {
		amount = common.IntToDecimals(tx.Amount, bs.wm.Decimal())
	}

	transx := &openwallet.Transaction{
		Fees:        "0",
		Coin:        coin,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.TxID,
		Decimal:     bs.wm.Decimal(),
		Amount:      amount.String(),
		ConfirmTime: tx.BlockTime,
		From:        []string{tx.From + ":" + amount.String()},
		To:          []string{tx.To + ":" + amount.String()},
		Status:      status,
		Reason:      reason,
	}

	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID

	txExtractData.Transaction = transx
	if operate == 0 {
		bs.extractTxInput(tx, txExtractData)
		bs.extractTxOutput(tx, txExtractData)
	} else if operate == 1 {
		bs.extractTxInput(tx, txExtractData)
	} else if operate == 2 {
		bs.extractTxOutput(tx, txExtractData)
	}

	txExtractDataArray = append(txExtractDataArray, txExtractData)
	result.extractData[tx.SourceKey] = txExtractDataArray
}

//extractTxInput 提取交易单输入部分,无需手续费，所以只包含1个TxInput
func (bs *TronBlockScanner) extractTxInput(tx *Contract, txExtractData *openwallet.TxExtractData) {

	amount := decimal.Zero
	coin := openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	if len(tx.ContractAddress) > 0 {
		contractId := openwallet.GenContractID(bs.wm.Symbol(), tx.ContractAddress)
		coin.ContractID = contractId
		coin.IsContract = true
		coin.Contract = openwallet.SmartContract{
			ContractID: contractId,
			Address:    tx.ContractAddress,
			Symbol:     bs.wm.Symbol(),
			Protocol:   tx.Protocol,
		}
		amount = common.IntToDecimals(tx.Amount, 0)
	} else {
		amount = common.IntToDecimals(tx.Amount, bs.wm.Decimal())
	}

	//主网from交易转账信息，第一个TxInput
	txInput := &openwallet.TxInput{}
	txInput.Recharge.Sid = openwallet.GenTxInputSID(tx.TxID, bs.wm.Symbol(), "", uint64(0))
	txInput.Recharge.TxID = tx.TxID
	txInput.Recharge.Address = tx.From
	txInput.Recharge.Coin = coin
	txInput.Recharge.Amount = amount.String()
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Index = 0 //账户模型填0
	txInput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)
}

//extractTxOutput 提取交易单输入部分,只有一个TxOutPut
func (bs *TronBlockScanner) extractTxOutput(tx *Contract, txExtractData *openwallet.TxExtractData) {

	amount := decimal.Zero
	coin := openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: false,
	}
	if len(tx.ContractAddress) > 0 {
		contractId := openwallet.GenContractID(bs.wm.Symbol(), tx.ContractAddress)
		coin.ContractID = contractId
		coin.IsContract = true
		coin.Contract = openwallet.SmartContract{
			ContractID: contractId,
			Address:    tx.ContractAddress,
			Symbol:     bs.wm.Symbol(),
			Protocol:   tx.Protocol,
		}

		amount = common.IntToDecimals(tx.Amount, 0)
	} else {
		amount = common.IntToDecimals(tx.Amount, bs.wm.Decimal())
	}

	//主网to交易转账信息,只有一个TxOutPut
	txOutput := &openwallet.TxOutPut{}
	txOutput.Recharge.Sid = openwallet.GenTxOutPutSID(tx.TxID, bs.wm.Symbol(), "", uint64(0))
	txOutput.Recharge.TxID = tx.TxID
	txOutput.Recharge.Address = tx.To
	txOutput.Recharge.Coin = coin
	txOutput.Recharge.Amount = amount.String()
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Index = 0 //账户模型填0
	txOutput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
}

//发送通知
func (bs *TronBlockScanner) newExtractDataNotify(height uint64, extractData map[string][]*openwallet.TxExtractData) error {
	for o, _ := range bs.Observers {
		for key, array := range extractData {
			for _, data := range array {
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
	}
	return nil
}

func (wm *WalletManager) GetTransaction(txid string, blockhash string, blockheight uint64, blocktime int64) (*Transaction, error) {
	params := req.Param{"value": txid}
	r, err := wm.WalletClient.Call("/wallet/gettransactionbyid", params)
	if err != nil {
		return nil, err
	}
	tx := NewTransaction(r, blockhash, blockheight, blocktime, wm.Config.IsTestNet)
	return tx, nil
}

//func (wm *WalletManager) GetTRXTransaction(txid, blockHash string, blockHeight uint64) (*Transaction, error) {
//	params := req.Param{"value": txid}
//	r, err := wm.WalletClient.Call("/wallet/gettransactionbyid", params)
//	if err != nil {
//		return nil, err
//	}
//	tx := NewTransaction(r)
//	tx.BlockHeight = blockHeight
//	tx.BlockHash = blockHash
//	var txCore core.Transaction
//	var tc core.TriggerSmartContract
//	//err = proto.Unmarshal([]byte(r.Raw), &txCore)
//	err = jsonpb.UnmarshalString(r.Raw, &txCore)
//	//err = json.Unmarshal([]byte(r.Raw), &txCore)
//	if err != nil {
//		return nil, err
//	}
//	err = ptypes.UnmarshalAny(txCore.RawData.Contract[0].Parameter, &tc)
//	if err != nil {
//		return nil, err
//	}
//	tx.Core = &txCore
//	return tx, nil
//}

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
func (bs *TronBlockScanner) BatchExtractTransaction(blockHeight uint64, blockHash string, blockTime int64, txs []*Transaction) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(txs) //需要完成的总数
	)

	if len(txs) == 0 {
		return nil
	}

	bs.wm.Log.Std.Info("block scanner ready extract transactions total: %d ", len(txs))

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
	extractWork := func(eblockHeight uint64, eBlockHash string, eBlockTime int64, mTxs []*Transaction, eProducer chan ExtractResult) {
		for _, tx := range mTxs {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlockHeight uint64, mTx *Transaction, end chan struct{}, mProducer chan<- ExtractResult) {
				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, eBlockTime, mTx, bs.ScanAddressFunc)
				//释放
				<-end

			}(eblockHeight, tx, bs.extractingCH, eProducer)
		}
	}
	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(blockHeight, worker)

	//独立线程运行生产
	go extractWork(blockHeight, blockHash, blockTime, txs, producer)

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

	currentBlock, err := bs.wm.GetNowBlock()
	if err != nil {
		bs.wm.Log.Std.Info("get current block header error;unexpected error:%v", err)
		return nil, err
	}
	blockHeight = currentBlock.Height
	hash = currentBlock.Hash

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
		balance, err := bs.wm.getBalance(a)
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
