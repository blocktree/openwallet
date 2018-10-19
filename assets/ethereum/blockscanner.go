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
package ethereum

//block scan db内容:
//1. 扫描过的blockheader
//2. unscanned tx
//3. block height, block hash

import (
	"path/filepath"
	"strings"

	"github.com/blocktree/OpenWallet/crypto"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"

	//	"fmt"
	"errors"
	"strconv"

	//	"golang.org/x/text/currency"
	"encoding/base64"
	"fmt"
)

const (
	//BLOCK_CHAIN_BUCKET = "blockchain" //区块链数据集合
	//periodOfTask      = 5 * time.Second //定时任务执行隔间
	MAX_EXTRACTING_SIZE = 15 //并发的扫描线程数

	BLOCK_HASH_KEY   = "BlockHash"
	BLOCK_HEIGHT_KEY = "BlockHeight"
)

type ETHBLockScanner struct {
	*openwallet.BlockScannerBase
	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
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

//NewBTCBlockScanner 创建区块链扫描器
func NewETHBlockScanner(wm *WalletManager) *ETHBLockScanner {
	bs := ETHBLockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, MAX_EXTRACTING_SIZE)
	bs.wm = wm
	bs.IsScanMemPool = false
	bs.RescanLastBlockCount = 0

	//设置扫描任务
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

func (this *ETHBLockScanner) newBlockNotify(block *EthBlock, isFork bool) {
	for o, _ := range this.Observers {
		header := block.CreateOpenWalletBlockHeader()
		header.Fork = isFork
		header.Symbol = this.wm.SymbolID
		o.BlockScanNotify(header)
	}
}

func (this *ETHBLockScanner) ScanBlock(height uint64) error {
	curBlock, err := this.wm.WalletClient.ethGetBlockSpecByBlockNum(height, true)
	if err != nil {
		log.Errorf("ethGetBlockSpecByBlockNum failed, err = %v", err)
		return err
	}

	err = this.BatchExtractTransaction(curBlock.Transactions)
	if err != nil {
		log.Errorf("BatchExtractTransaction failed, err = %v", err)
		return err
	}

	go this.newBlockNotify(curBlock, false)

	return nil
}

func (this *ETHBLockScanner) ScanTxMemPool() error {
	log.Infof("block scanner start to scan mempool.")

	txs, err := this.GetTxPoolPendingTxs()
	if err != nil {
		log.Errorf("get txpool pending txs failed, err=%v", err)
		return err
	}

	err = this.BatchExtractTransaction(txs)
	if err != nil {
		log.Errorf("batch extract transactions failed, err=%v", err)
		return err
	}
	return nil
}

func (this *ETHBLockScanner) RescanFailedTransactions() error {
	unscannedTxs, err := this.wm.GetAllUnscannedTransactions()
	if err != nil {
		log.Errorf("GetAllUnscannedTransactions failed. err=%v", err)
		return err
	}

	txs, err := this.wm.RecoverUnscannedTransactions(unscannedTxs)
	if err != nil {
		log.Errorf("recover transactions from unscanned result failed. err=%v", err)
		return err
	}

	err = this.BatchExtractTransaction(txs)
	if err != nil {
		log.Errorf("batch extract transactions failed, err=%v", err)
		return err
	}

	err = this.wm.DeleteUnscannedTransactions(unscannedTxs)
	if err != nil {
		log.Errorf("batch extract transactions failed, err=%v", err)
		return err
	}
	return nil
}

func (this *ETHBLockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := this.GetCurrentBlockHeader()
	if err != nil {
		log.Errorf("block scanner can not get new block height; unexpected error: %v", err)
		return
	}

	curBlockHeight := blockHeader.Height
	curBlockHash := blockHeader.Hash
	var previousHeight uint64 = 0
	for {
		maxBlockHeight, err := this.wm.WalletClient.ethGetBlockNumber()
		if err != nil {
			log.Errorf("get max height of eth failed, err=%v", err)
			break
		}

		log.Info("current block height:", "0x"+strconv.FormatUint(curBlockHeight, 16), " maxBlockHeight:", "0x"+strconv.FormatUint(maxBlockHeight, 16))
		if curBlockHeight == maxBlockHeight {
			log.Infof("block scanner has done with scan. current height:%v", "0x"+strconv.FormatUint(maxBlockHeight, 16))
			break
		}

		//扫描下一个区块
		curBlockHeight += 1
		log.Infof("block scanner try to scan block No.%v", curBlockHeight)

		curBlock, err := this.wm.WalletClient.ethGetBlockSpecByBlockNum(curBlockHeight, true)
		if err != nil {
			log.Errorf("ethGetBlockSpecByBlockNum failed, err = %v", err)
			break
		}

		isFork := false

		if curBlock.PreviousHash != curBlockHash {
			previousHeight = curBlockHeight - 1
			log.Infof("block has been fork on height: %v.", "0x"+strconv.FormatUint(curBlockHeight, 16))
			log.Infof("block height: %v local hash = %v ", "0x"+strconv.FormatUint(previousHeight, 16), curBlockHash)
			log.Infof("block height: %v mainnet hash = %v ", "0x"+strconv.FormatUint(previousHeight, 16), curBlock.PreviousHash)

			log.Infof("delete recharge records on block height: %v.", "0x"+strconv.FormatUint(previousHeight, 16))

			//本地数据库并不存储交易
			//err = this.DeleteTransactionsByHeight(previousHeight)
			//if err != nil {
			//	log.Errorf("DeleteTransactionsByHeight failed, height=%v, err=%v", "0x"+strconv.FormatUint(previousHeight, 16), err)
			//	break
			//}

			err = this.wm.DeleteUnscannedTransactionByHeight(previousHeight)
			if err != nil {
				log.Errorf("DeleteUnscannedTransaction failed, height=%v, err=%v", "0x"+strconv.FormatUint(previousHeight, 16), err)
				break
			}

			curBlockHeight = previousHeight - 1 //倒退2个区块重新扫描

			curBlock, err = this.wm.RecoverBlockHeader(curBlockHeight)
			if err != nil && err != storm.ErrNotFound {
				log.Errorf("RecoverBlockHeader failed, block number=%v, err=%v", "0x"+strconv.FormatUint(curBlockHeight, 16), err)
				break
			} else if err == storm.ErrNotFound {
				curBlock, err = this.wm.WalletClient.ethGetBlockSpecByBlockNum(curBlockHeight, false)
				if err != nil {
					log.Errorf("ethGetBlockSpecByBlockNum  failed, block number=%v, err=%v", "0x"+strconv.FormatUint(curBlockHeight, 16), err)
					break
				}
			}
			curBlockHash = curBlock.BlockHash
			log.Infof("rescan block on height:%v, hash:%v.", curBlockHeight, curBlockHash)

			err = this.wm.SaveLocalBlockScanned(curBlock.blockHeight, curBlock.BlockHash)
			if err != nil {
				log.Errorf("save local block unscaned failed, err=%v", err)
				break
			}

			isFork = true
		} else {
			err = this.BatchExtractTransaction(curBlock.Transactions)
			if err != nil {
				log.Errorf("block scanner can not extractRechargeRecords; unexpected error: %v", err)
				break
			}

			err = this.wm.SaveBlockHeader2(curBlock)
			if err != nil {
				log.Errorf("SaveBlockHeader2 failed")
			}

			isFork = false
		}

		curBlockHeight = curBlock.blockHeight
		curBlockHash = curBlock.BlockHash
		go this.newBlockNotify(curBlock, isFork)
	}

	if this.IsScanMemPool {
		this.ScanTxMemPool()
	}

	this.RescanFailedTransactions()
}

//GetSourceKeyByAddress 获取地址对应的数据源标识
func (this *ETHBLockScanner) GetSourceKeyByAddress(address string) (string, bool) {
	this.Mu.RLock()
	defer this.Mu.RUnlock()

	sourceKey, ok := this.AddressInScanning[address]
	return sourceKey, ok
}

//newExtractDataNotify 发送通知
func (this *ETHBLockScanner) newExtractDataNotify(height uint64, tx *BlockTransaction, extractData map[string]*openwallet.TxExtractData) error {

	for o, _ := range this.Observers {
		for key, data := range extractData {
			err := o.BlockExtractDataNotify(key, data)
			if err != nil {
				//记录未扫区块
				//unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
				//err = this.SaveUnscanRecord(unscanRecord)
				reason := fmt.Sprintf("BlockExtractDataNotify account[%v] failed, err = %v", key, err)
				log.Errorf(reason)
				err = this.wm.SaveUnscannedTransaction(tx, reason)
				if err != nil {
					log.Errorf("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
					return err
				}
			}
		}
	}

	return nil
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (this *ETHBLockScanner) BatchExtractTransaction(txs []BlockTransaction) error {
	for i := range txs {
		extractResult, err := this.TransactionScanning(&txs[i])
		if err != nil {
			log.Errorf("transaction scanning failed, err=%v", err)
			return err
		}

		if extractResult != nil {
			err := this.newExtractDataNotify(txs[i].BlockHeight, &txs[i], extractResult.extractData)
			if err != nil {
				log.Errorf("newExtractDataNotify failed, err=%v", err)
				return err
			}
		}
	}
	return nil
}

func (this *ETHBLockScanner) GetTxPoolPendingTxs() ([]BlockTransaction, error) {
	txpoolContent, err := this.wm.WalletClient.ethGetTxPoolContent()
	if err != nil {
		log.Errorf("get txpool content failed, err=%v", err)
		return nil, err
	}

	var txs []BlockTransaction
	for from, txsets := range txpoolContent.Pending {
		if this.IsExistAddress(strings.ToLower(from)) {
			for nonce, _ := range txsets {
				txs = append(txs, txsets[nonce])
			}
		} else {
			for nonce, _ := range txsets {
				if this.IsExistAddress(strings.ToLower(txsets[nonce].To)) {
					txs = append(txs, txsets[nonce])
				}

			}
		}
	}
	return txs, nil
}

func (this *ETHBLockScanner) InitEthTokenExtractResult(tx *BlockTransaction, tokenEvent *TransferEvent, result *ExtractResult, isFromAccount bool) {
	txExtractData := &openwallet.TxExtractData{}
	ContractId := base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("{%v}_{%v}", this.wm.Symbol(), tx.To))))
	coin := openwallet.Coin{
		Symbol:     this.wm.Symbol(),
		IsContract: true,
		ContractID: ContractId,
		Contract: openwallet.SmartContract{
			ContractID: ContractId,
			Address:    tx.To,
			Symbol:     this.wm.Symbol(),
		},
	}

	transx := &openwallet.Transaction{
		//From: tx.From,
		//To:   tx.To,
		Fees:        tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin:        coin,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.Hash,
		Decimal:     18,
	}
	//contractId :=
	//base64.StdEncoding.EncodeToString([]byte(auth))
	//crypto.SHA256([]byte(s))
	//base64(sha256({symbol}_{address}))

	transx.From = append(transx.From, tx.From)
	transx.To = append(transx.To, tx.To)
	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID
	txExtractData.Transaction = transx

	var sourceKey string
	if isFromAccount {
		sourceKey = tokenEvent.FromSourceKey
	} else {
		sourceKey = tokenEvent.ToSourceKey
	}

	txInput := &openwallet.TxInput{
		SourceTxID: tx.Hash,
	}
	txInput.Recharge.Address = tokenEvent.TokenFrom
	txInput.Recharge.Amount = tokenEvent.Value
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Coin = coin
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)
	txOutput := &openwallet.TxOutPut{}
	txOutput.Recharge.Address = tokenEvent.TokenTo
	txOutput.Recharge.Amount = tokenEvent.Value
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Coin = coin
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
	result.extractData[sourceKey] = txExtractData
}

func (this *ETHBLockScanner) InitEthExtractResult(tx *BlockTransaction, result *ExtractResult, isFromAccount bool) {
	txExtractData := &openwallet.TxExtractData{}
	transx := &openwallet.Transaction{
		//From: tx.From,
		//To:   tx.To,
		Fees: tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin: openwallet.Coin{
			Symbol:     this.wm.Symbol(),
			IsContract: false,
		},
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		TxID:        tx.Hash,
		Decimal:     18,
	}
	transx.From = append(transx.From, tx.From)
	transx.To = append(transx.To, tx.To)
	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID
	txExtractData.Transaction = transx
	if isFromAccount {
		result.extractData[tx.FromSourceKey] = txExtractData
	} else {
		result.extractData[tx.ToSourceKey] = txExtractData
	}
}

func (this *WalletManager) GetErc20TokenEvent(transactionID string) (*TransferEvent, error) {
	receipt, err := this.WalletClient.ethGetTransactionReceipt(transactionID)
	if err != nil {
		log.Errorf("get transaction receipt failed, err=%v", err)
		return nil, err
	}

	transEvent := receipt.ParseTransferEvent()
	if transEvent == nil {
		return nil, nil
	}

	return transEvent, nil
}

func (this *ETHBLockScanner) GetErc20TokenEvent(tx *BlockTransaction) (*TransferEvent, error) {
	//非合约交易或未打包交易跳过
	//obj, _ := json.MarshalIndent(tx, "", " ")
	//log.Debugf("tx:%v", string(obj))
	if tx.Data == "0x" || tx.Data == "" || tx.BlockHeight == 0 || tx.BlockHash == "" {
		return nil, nil
	}

	return this.wm.GetErc20TokenEvent(tx.Hash)
}

func (this *ETHBLockScanner) TransactionScanning(tx *BlockTransaction) (*ExtractResult, error) {
	//txToNotify := make(map[string][]BlockTransaction)

	blockHeight, err := ConvertToUint64(tx.BlockNumber, 16)
	if err != nil {
		log.Errorf("convert block number from string to uint64 failed, err=%v", err)
		return nil, err
	}

	tx.BlockHeight = blockHeight
	var result = ExtractResult{
		BlockHeight: blockHeight,
		TxID:        tx.BlockHash,
		extractData: make(map[string]*openwallet.TxExtractData),
	}

	tokenEvent, err := this.GetErc20TokenEvent(tx)
	if err != nil {
		log.Errorf("GetErc20TokenEvent failed, err=%v", err)
		return nil, err
	}
	log.Debugf("get token Event:%v", tokenEvent)

	// 普通交易, from地址在监听地址中
	if this.IsExistAddress(tx.From) && tokenEvent == nil {
		log.Debugf("tx.from found in transaction [%v] .", tx.Hash)
		if sourceKey, exist := this.GetSourceKeyByAddress(tx.From); exist {
			tx.FromSourceKey = sourceKey
			this.InitEthExtractResult(tx, &result, true)
		} else {
			return nil, errors.New("tx.from unexpected error.")
		}
	} else if tokenEvent != nil && this.IsExistAddress(tokenEvent.TokenFrom) {
		//erc20 token交易, from地址在监听地址中
		if sourceKey, exist := this.GetSourceKeyByAddress(tokenEvent.TokenFrom); exist {
			tokenEvent.FromSourceKey = sourceKey
			this.InitEthTokenExtractResult(tx, tokenEvent, &result, true)
		}

	} else {
		log.Debugf("tx.from[%v] not found in scanning address.", tx.From)
	}

	if this.IsExistAddress(tx.To) && tokenEvent == nil {
		log.Debugf("tx.to found in transaction [%v].", tx.Hash)
		if sourceKey, exist := this.GetSourceKeyByAddress(tx.To); exist {
			if _, exist = result.extractData[sourceKey]; !exist {
				tx.ToSourceKey = sourceKey
				this.InitEthExtractResult(tx, &result, false)
			}

		} else if tokenEvent != nil && this.IsExistAddress(tokenEvent.TokenTo) {
			if sourceKey, exist := this.GetSourceKeyByAddress(tokenEvent.TokenTo); exist {
				if _, exist = result.extractData[sourceKey]; !exist {
					tokenEvent.ToSourceKey = sourceKey
					this.InitEthTokenExtractResult(tx, tokenEvent, &result, false)
				}
			}

		} else {
			return nil, errors.New("tx.to unexpected error.")
		}

	} else if len(result.extractData) == 0 {
		log.Debugf("tx.to[%v] not found in scanning address.", tx.To)
		return nil, nil
	}

	return &result, nil
}

//GetLocalNewBlock 获取本地记录的区块高度和hash
func (this *WalletManager) GetLocalNewBlock() (uint64, string, error) {

	var (
		blockHeight uint64 = 0
		blockHash          = ""
	)

	//获取本地区块高度
	filePath := filepath.Join(this.Config.DbPath, this.Config.BlockchainFile)
	db, err := storm.Open(filePath)
	if err != nil {
		log.Errorf("open %v failed, err=%v", filePath, err)
		return 0, "", err
	}
	defer db.Close()

	err = db.Get(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil && err != storm.ErrNotFound {
		log.Errorf("get local block height failed, err = %v", err)
		return 0, "", err
	}
	err = db.Get(BLOCK_CHAIN_BUCKET, BLOCK_HASH_KEY, &blockHash)
	if err != nil && err != storm.ErrNotFound {
		log.Errorf("get local block hash failed, err = %v", err)
		return 0, "", err
	}

	return blockHeight, blockHash, nil
}

//GetCurrentBlockHeader 获取当前区块高度
func (this *ETHBLockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash, err = this.wm.GetLocalNewBlock()
	if err != nil {
		log.Errorf("get local new block failed, err=%v", err)
		return nil, err
	}

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeight, err = this.wm.WalletClient.ethGetBlockNumber()
		if err != nil {
			log.Errorf("ethGetBlockNumber failed, err=%v", err)
			return nil, err
		}

		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		block, err := this.wm.WalletClient.ethGetBlockSpecByBlockNum(blockHeight, false)
		if err != nil {
			log.Errorf("get block spec by block number failed, err=%v", err)
			return nil, err
		}
		hash = block.BlockHash
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}
