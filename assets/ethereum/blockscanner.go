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
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"

	//	"fmt"
	"errors"

	//	"golang.org/x/text/currency"

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
	extractData map[string][]*openwallet.TxExtractData

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

//SetRescanBlockHeight 重置区块链扫描高度
func (this *ETHBLockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return errors.New("block height to rescan must greater than 0.")
	}

	block, err := this.wm.WalletClient.ethGetBlockSpecByBlockNum(height, false)
	if err != nil {
		log.Errorf("get block spec by block number[%v] failed, err=%v", height, err)
		return err
	}

	err = this.wm.SaveLocalBlockScanned(height, block.BlockHash)
	if err != nil {
		log.Errorf("save local block scanned failed, err=%v", err)
		return err
	}

	return nil
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

		log.Info("current block height:", curBlockHeight, " maxBlockHeight:", maxBlockHeight)
		if curBlockHeight == maxBlockHeight {
			log.Infof("block scanner has done with scan. current height:%v", maxBlockHeight)
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
			log.Infof("block has been fork on height: %v.", curBlockHeight)
			log.Infof("block height: %v local hash = %v ", previousHeight, curBlockHash)
			log.Infof("block height: %v mainnet hash = %v ", previousHeight, curBlock.PreviousHash)

			log.Infof("delete recharge records on block height: %v.", previousHeight)

			//本地数据库并不存储交易
			//err = this.DeleteTransactionsByHeight(previousHeight)
			//if err != nil {
			//	log.Errorf("DeleteTransactionsByHeight failed, height=%v, err=%v", "0x"+strconv.FormatUint(previousHeight, 16), err)
			//	break
			//}

			err = this.wm.DeleteUnscannedTransactionByHeight(previousHeight)
			if err != nil {
				log.Errorf("DeleteUnscannedTransaction failed, height=%v, err=%v", previousHeight, err)
				break
			}

			curBlockHeight = previousHeight - 1 //倒退2个区块重新扫描

			curBlock, err = this.wm.RecoverBlockHeader(curBlockHeight)
			if err != nil && err != storm.ErrNotFound {
				log.Errorf("RecoverBlockHeader failed, block number=%v, err=%v", curBlockHeight, err)
				break
			} else if err == storm.ErrNotFound {
				curBlock, err = this.wm.WalletClient.ethGetBlockSpecByBlockNum(curBlockHeight, false)
				if err != nil {
					log.Errorf("ethGetBlockSpecByBlockNum  failed, block number=%v, err=%v", curBlockHeight, err)
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

//newExtractDataNotify 发送通知
func (this *ETHBLockScanner) newExtractDataNotify(height uint64, tx *BlockTransaction, extractDataList map[string][]*openwallet.TxExtractData) error {

	for o, _ := range this.Observers {
		for key, extractData := range extractDataList {
			for _, data := range extractData {
				log.Debugf("before notify, data.tx.Amount:%v", data.Transaction.Amount)
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
				log.Debugf("data.tx.Amount:%v", data.Transaction.Amount)
			}
		}
	}

	return nil
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (this *ETHBLockScanner) BatchExtractTransaction(txs []BlockTransaction) error {
	for i := range txs {
		txs[i].filterFunc = this.ScanAddressFunc
		extractResult, err := this.TransactionScanning(&txs[i])
		if err != nil {
			log.Errorf("transaction  failed, err=%v", err)
			return err
		}

		if extractResult.extractData != nil {
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
		if _, ok := this.ScanAddressFunc(strings.ToLower(from)); ok {
			for nonce, _ := range txsets {
				txs = append(txs, txsets[nonce])
			}
		} else {
			for nonce, _ := range txsets {
				if _, ok2 := this.ScanAddressFunc(strings.ToLower(txsets[nonce].To)); ok2 {
					txs = append(txs, txsets[nonce])
				}

			}
		}
	}
	return txs, nil
}

// func (this *ETHBLockScanner) InitEthTokenExtractResult(tx *BlockTransaction, tokenEvent *TransferEvent, result *ExtractResult, isFromAccount bool) {
// 	txExtractData := &openwallet.TxExtractData{}
// 	ContractId := base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("{%v}_{%v}", this.wm.Symbol(), tx.To))))
// 	coin := openwallet.Coin{
// 		Symbol:     this.wm.Symbol(),
// 		IsContract: true,
// 		ContractID: ContractId,
// 		Contract: openwallet.SmartContract{
// 			ContractID: ContractId,
// 			Address:    tx.To,
// 			Symbol:     this.wm.Symbol(),
// 		},
// 	}

// 	gasPriceStr := ""
// 	gasPrice, _ := ConvertToBigInt(tx.GasPrice, 16)
// 	gas, _ := ConvertToBigInt(tx.Gas, 16)
// 	feeInteger := big.NewInt(0)
// 	feeInteger.Mul(gasPrice, gas)
// 	fee, err := ConverWeiStringToEthDecimal(feeInteger.String())
// 	if err != nil {
// 		log.Errorf("convert to eth decimal failed, err=%v", err)
// 		gasPriceStr = tx.GasPrice
// 	} else {
// 		gasPriceStr = fee.String()
// 	}

// 	transx := &openwallet.Transaction{
// 		//From: tx.From,
// 		//To:   tx.To,
// 		Fees:        gasPriceStr, //totalSpent.Sub(totalReceived).StringFixed(8),
// 		Coin:        coin,
// 		BlockHash:   tx.BlockHash,
// 		BlockHeight: tx.BlockHeight,
// 		TxID:        tx.Hash,
// 		Decimal:     18,
// 	}
// 	//contractId :=
// 	//base64.StdEncoding.EncodeToString([]byte(auth))
// 	//crypto.SHA256([]byte(s))
// 	//base64(sha256({symbol}_{address}))

// 	transx.From = append(transx.From, tx.From)
// 	transx.To = append(transx.To, tx.To)
// 	wxID := openwallet.GenTransactionWxID(transx)
// 	transx.WxID = wxID
// 	txExtractData.Transaction = transx

// 	var sourceKey string
// 	if isFromAccount {
// 		sourceKey = tokenEvent.FromSourceKey
// 	} else {
// 		sourceKey = tokenEvent.ToSourceKey
// 	}

// 	txInput := &openwallet.TxInput{
// 		SourceTxID: tx.Hash,
// 	}
// 	txInput.Recharge.Address = tokenEvent.TokenFrom
// 	txInput.Recharge.Amount = tokenEvent.Value
// 	txInput.Recharge.BlockHash = tx.BlockHash
// 	txInput.Recharge.BlockHeight = tx.BlockHeight
// 	txInput.Recharge.Coin = coin
// 	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)
// 	txOutput := &openwallet.TxOutPut{}
// 	txOutput.Recharge.Address = tokenEvent.TokenTo
// 	txOutput.Recharge.Amount = tokenEvent.Value
// 	txOutput.Recharge.BlockHash = tx.BlockHash
// 	txOutput.Recharge.BlockHeight = tx.BlockHeight
// 	txOutput.Recharge.Coin = coin
// 	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
// 	result.extractData[sourceKey] = txExtractData
// }

// func (this *ETHBLockScanner) InitEthExtractResult(tx *BlockTransaction, result *ExtractResult, isFromAccount bool) {
// 	txExtractData := &openwallet.TxExtractData{}

// 	gasPriceStr := ""
// 	gasPrice, _ := ConvertToBigInt(tx.GasPrice, 16)
// 	gas, _ := ConvertToBigInt(tx.Gas, 16)
// 	feeInteger := big.NewInt(0)
// 	feeInteger.Mul(gasPrice, gas)
// 	fee, err := ConverWeiStringToEthDecimal(feeInteger.String())
// 	if err != nil {
// 		log.Errorf("convert to eth decimal failed, err=%v", err)
// 		gasPriceStr = tx.GasPrice
// 	} else {
// 		gasPriceStr = fee.String()
// 	}

// 	amount, _ := ConvertToBigInt(tx.Value, 16)
// 	amountVal, _ := ConverWeiStringToEthDecimal(amount.String())
// 	log.Debugf("tx.Value:%v amount:%v amountVal:%v ", tx.Value, amount.String(), amountVal.String())
// 	transx := &openwallet.Transaction{
// 		//From: tx.From,
// 		//To:   tx.To,
// 		Fees: gasPriceStr, //tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
// 		Coin: openwallet.Coin{
// 			Symbol:     this.wm.Symbol(),
// 			IsContract: false,
// 		},
// 		BlockHash:   tx.BlockHash,
// 		BlockHeight: tx.BlockHeight,
// 		TxID:        tx.Hash,
// 		Decimal:     18,
// 		Amount:      amountVal.String(),
// 	}
// 	log.Debugf("transx.Amount:%v", transx.Amount)
// 	transx.From = append(transx.From, tx.From)
// 	transx.To = append(transx.To, tx.To)
// 	wxID := openwallet.GenTransactionWxID(transx)
// 	transx.WxID = wxID
// 	txExtractData.Transaction = transx
// 	if isFromAccount {
// 		result.extractData[tx.FromSourceKey] = txExtractData
// 	} else {
// 		result.extractData[tx.ToSourceKey] = txExtractData
// 	}
// }

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

func (this *ETHBLockScanner) UpdateTxByReceipt(tx *BlockTransaction) (*TransferEvent, error) {
	//过滤掉未打包交易
	if tx.BlockHeight == 0 || tx.BlockHash == "" {
		return nil, nil
	}

	receipt, err := this.wm.WalletClient.ethGetTransactionReceipt(tx.Hash)
	if err != nil {
		log.Errorf("get transaction receipt failed, err=%v", err)
		return nil, err
	}

	tx.Gas = receipt.GasUsed
	// transEvent := receipt.ParseTransferEvent()
	// if transEvent == nil {
	// 	return nil, nil
	// }
	return receipt.ParseTransferEvent(), nil
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

func (this *ETHBLockScanner) MakeToExtractData(tx *BlockTransaction, tokenEvent *TransferEvent) (string, []*openwallet.TxExtractData, error) {
	if tokenEvent == nil {
		return this.MakeSimpleToExtractData(tx)
	}
	return this.MakeTokenToExtractData(tx, tokenEvent)
}

func (this *ETHBLockScanner) MakeSimpleToExtractData(tx *BlockTransaction) (string, []*openwallet.TxExtractData, error) {
	var sourceKey string
	var exist bool
	var extractDataList []*openwallet.TxExtractData
	if sourceKey, exist = tx.filterFunc(tx.To); !exist { //this.GetSourceKeyByAddress(tx.To)
		return "", extractDataList, nil
	}

	feeprice, err := tx.GetTxFeeEthString()
	if err != nil {
		log.Errorf("calc tx fee in eth failed, err=%v", err)
		return "", extractDataList, err
	}

	amountVal, err := tx.GetAmountEthString()
	if err != nil {
		log.Errorf("calc amount to eth decimal failed, err=%v", err)
		return "", extractDataList, err
	}

	nowUnix := time.Now().Unix()

	balanceTxOut := openwallet.TxOutPut{
		Recharge: openwallet.Recharge{
			Sid:      openwallet.GenTxOutPutSID(tx.Hash, this.wm.Symbol(), "", 0), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tx.To)))),
			CreateAt: nowUnix,
			TxID:     tx.Hash,
			Address:  tx.To,
			Coin: openwallet.Coin{
				Symbol:     this.wm.Symbol(),
				IsContract: false,
			},
			Amount:      amountVal,
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}

	from := []string{
		tx.From,
	}
	to := []string{
		tx.To,
	}

	transx := &openwallet.Transaction{
		WxID:        openwallet.GenTransactionWxID2(tx.Hash, this.wm.Symbol(), ""),
		TxID:        tx.Hash,
		From:        from,
		To:          to,
		Decimal:     18,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		Fees:        feeprice,
		Coin: openwallet.Coin{
			Symbol:     this.wm.Symbol(),
			IsContract: false,
		},
		SubmitTime:  nowUnix,
		ConfirmTime: nowUnix,
	}

	txExtractData := &openwallet.TxExtractData{}
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, &balanceTxOut)
	txExtractData.Transaction = transx

	extractDataList = append(extractDataList, txExtractData)

	return sourceKey, extractDataList, nil
}

func (this *ETHBLockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {
	type addressBalance struct {
		Address string
		Index   uint64
		Balance *openwallet.Balance
	}

	threadControl := make(chan int, 20)
	defer close(threadControl)
	resultChan := make(chan *addressBalance, 1024)
	defer close(resultChan)
	done := make(chan int, 1)
	count := len(address)
	resultBalance := make([]*openwallet.Balance, count)
	resultSaveFailed := false
	//save result
	go func() {
		for i := 0; i < count; i++ {
			addr := <-resultChan
			if addr.Balance != nil {
				resultBalance[addr.Index] = addr.Balance
			} else {
				resultSaveFailed = true
			}
		}
		done <- 1
	}()

	query := func(addr *addressBalance) {
		threadControl <- 1
		defer func() {
			resultChan <- addr
			<-threadControl
		}()

		balanceConfirmed, err := this.wm.WalletClient.GetAddrBalance2(appendOxToAddress(addr.Address), "latest")
		if err != nil {
			log.Error("get address[", addr.Address, "] balance failed, err=", err)
			return
		}

		balanceAll, err := this.wm.WalletClient.GetAddrBalance2(appendOxToAddress(addr.Address), "pending")
		if err != nil {
			log.Errorf("get address[%v] erc20 token balance failed, err=%v", address, err)
			return
		}

		//		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
		balanceUnconfirmed := big.NewInt(0)
		balanceUnconfirmed.Sub(balanceAll, balanceConfirmed)

		balance := &openwallet.Balance{
			Symbol:  this.wm.Symbol(),
			Address: addr.Address,
		}
		confirmed, err := ConverWeiStringToEthDecimal(balanceConfirmed.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal confirmed balance failed, err=%v", err)
			return
		}
		all, err := ConverWeiStringToEthDecimal(balanceAll.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal all balance failed, err=%v", err)
			return
		}

		unconfirmed, err := ConverWeiStringToEthDecimal(balanceUnconfirmed.String())
		if err != nil {
			log.Errorf("ConverWeiStringToEthDecimal unconfirmed balance failed, err=%v", err)
			return
		}

		balance.Balance = all.String()
		balance.UnconfirmBalance = unconfirmed.String()
		balance.ConfirmBalance = confirmed.String()
		addr.Balance = balance
	}

	for i, _ := range address {
		addrbl := &addressBalance{
			Address: address[i],
			Index:   uint64(i),
		}
		go query(addrbl)
	}

	<-done
	if resultSaveFailed {
		return nil, errors.New("get balance of addresses failed.")
	}
	return resultBalance, nil
}

func (this *ETHBLockScanner) MakeTokenToExtractData(tx *BlockTransaction, tokenEvent *TransferEvent) (string, []*openwallet.TxExtractData, error) {
	var sourceKey string
	var exist bool
	var extractDataList []*openwallet.TxExtractData
	if sourceKey, exist = tx.filterFunc(tokenEvent.TokenTo); !exist { //this.GetSourceKeyByAddress(tokenEvent.TokenTo)
		return "", extractDataList, nil
	}
	// fee, err := tx.GetTxFeeEthString()
	// if err != nil {
	// 	log.Errorf("calc tx fee in eth failed, err=%v", err)
	// 	return "", extractDataList, err
	// }

	contractId := openwallet.GenContractID(this.wm.Symbol(), tx.To) //base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("{%v}_{%v}", this.wm.Symbol(), tx.To))))
	nowUnix := time.Now().Unix()

	coin := openwallet.Coin{
		Symbol:     this.wm.Symbol(),
		IsContract: true,
		ContractID: contractId,
		Contract: openwallet.SmartContract{
			ContractID: contractId,
			Address:    tx.To,
			Symbol:     this.wm.Symbol(),
		},
	}

	tokenValue, err := ConvertToBigInt(tokenEvent.Value, 16)
	if err != nil {
		log.Errorf("convert token value to big.int failed, err=%v", err)
		return "", extractDataList, err
	}

	tokenBalanceTxOutput := openwallet.TxOutPut{
		Recharge: openwallet.Recharge{
			Sid:         openwallet.GenTxOutPutSID(tx.Hash, this.wm.Symbol(), contractId, 0), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tokenEvent.TokenTo)))),
			CreateAt:    nowUnix,
			TxID:        tx.Hash,
			Address:     tokenEvent.TokenTo,
			Coin:        coin,
			Amount:      tokenValue.String(),
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}
	from := []string{
		tokenEvent.TokenFrom,
	}
	to := []string{
		tokenEvent.TokenTo,
	}

	tokentransx := &openwallet.Transaction{
		WxID: openwallet.GenTransactionWxID2(tx.Hash, this.wm.Symbol(), contractId),
		TxID: tx.Hash,
		From: from,
		To:   to,
		//		Decimal:     18,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		Fees:        "0", //tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin:        coin,
		SubmitTime:  nowUnix,
		ConfirmTime: nowUnix,
	}

	tokenTransExtractData := &openwallet.TxExtractData{}
	tokenTransExtractData.Transaction = tokentransx
	tokenTransExtractData.TxOutputs = append(tokenTransExtractData.TxOutputs, &tokenBalanceTxOutput)
	extractDataList = append(extractDataList, tokenTransExtractData)

	return sourceKey, extractDataList, nil
}

func (this *ETHBLockScanner) MakeFromExtractData(tx *BlockTransaction, tokenEvent *TransferEvent) (string, []*openwallet.TxExtractData, error) {
	if tokenEvent == nil {
		return this.MakeSimpleTxFromExtractData(tx)
	}
	return this.MakeTokenTxFromExtractData(tx, tokenEvent)
}

func (this *ETHBLockScanner) MakeSimpleTxFromExtractData(tx *BlockTransaction) (string, []*openwallet.TxExtractData, error) {
	var sourceKey string
	var exist bool
	var extractDataList []*openwallet.TxExtractData
	if sourceKey, exist = tx.filterFunc(tx.From); !exist { //this.GetSourceKeyByAddress(tx.From)
		return "", extractDataList, nil
	}

	feeprice, err := tx.GetTxFeeEthString()
	if err != nil {
		log.Errorf("calc tx fee in eth failed, err=%v", err)
		return "", extractDataList, err
	}

	amountVal, err := tx.GetAmountEthString()
	if err != nil {
		log.Errorf("calc amount to eth decimal failed, err=%v", err)
		return "", extractDataList, err
	}

	nowUnix := time.Now().Unix()

	deductTxInput := openwallet.TxInput{
		Recharge: openwallet.Recharge{
			Sid:      openwallet.GenTxInputSID(tx.Hash, this.wm.Symbol(), "", 0), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tx.From)))),
			CreateAt: nowUnix,
			TxID:     tx.Hash,
			Address:  tx.From,
			Coin: openwallet.Coin{
				Symbol:     this.wm.Symbol(),
				IsContract: false,
			},
			Amount:      amountVal,
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}

	feeTxInput := openwallet.TxInput{
		Recharge: openwallet.Recharge{
			Sid:      openwallet.GenTxInputSID(tx.Hash, this.wm.Symbol(), "", 1), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tx.From)))),
			CreateAt: nowUnix,
			TxID:     tx.Hash,
			Address:  tx.From,
			Coin: openwallet.Coin{
				Symbol:     this.wm.Symbol(),
				IsContract: false,
			},
			Amount:      feeprice,
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}

	from := []string{
		tx.From,
	}
	to := []string{
		tx.To,
	}

	transx := &openwallet.Transaction{
		WxID:        openwallet.GenTransactionWxID2(tx.Hash, this.wm.Symbol(), ""),
		TxID:        tx.Hash,
		From:        from,
		To:          to,
		Decimal:     18,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		Fees:        feeprice, //tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin: openwallet.Coin{
			Symbol:     this.wm.Symbol(),
			IsContract: false,
		},
		SubmitTime:  nowUnix,
		ConfirmTime: nowUnix,
	}

	txExtractData := &openwallet.TxExtractData{}
	txExtractData.TxInputs = append(txExtractData.TxInputs, &deductTxInput)
	txExtractData.TxInputs = append(txExtractData.TxInputs, &feeTxInput)
	txExtractData.Transaction = transx

	extractDataList = append(extractDataList, txExtractData)

	return sourceKey, extractDataList, nil
}

func (this *ETHBLockScanner) MakeTokenTxFromExtractData(tx *BlockTransaction, tokenEvent *TransferEvent) (string, []*openwallet.TxExtractData, error) {
	var sourceKey string
	var exist bool
	var extractDataList []*openwallet.TxExtractData
	if sourceKey, exist = tx.filterFunc(tokenEvent.TokenFrom); !exist { //this.GetSourceKeyByAddress(tokenEvent.TokenFrom)
		return "", extractDataList, nil
	}

	contractId := openwallet.GenContractID(this.wm.Symbol(), tx.To) //base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("{%v}_{%v}", this.wm.Symbol(), tx.To))))
	nowUnix := time.Now().Unix()

	coin := openwallet.Coin{
		Symbol:     this.wm.Symbol(),
		IsContract: true,
		ContractID: contractId,
		Contract: openwallet.SmartContract{
			ContractID: contractId,
			Address:    tx.To,
			Symbol:     this.wm.Symbol(),
		},
	}

	feeprice, err := tx.GetTxFeeEthString()
	if err != nil {
		log.Errorf("calc tx fee in eth failed, err=%v", err)
		return "", extractDataList, err
	}

	tokenValue, err := ConvertToBigInt(tokenEvent.Value, 16)
	if err != nil {
		log.Errorf("convert token value to big.int failed, err=%v", err)
		return "", extractDataList, err
	}

	deductTxInput := openwallet.TxInput{
		Recharge: openwallet.Recharge{
			Sid:         openwallet.GenTxInputSID(tx.Hash, this.wm.Symbol(), contractId, 0), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tx.From)))),
			CreateAt:    nowUnix,
			TxID:        tx.Hash,
			Address:     tx.From,
			Coin:        coin,
			Amount:      tokenValue.String(),
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}
	from := []string{
		tokenEvent.TokenFrom,
	}
	to := []string{
		tokenEvent.TokenTo,
	}

	tokentransx := &openwallet.Transaction{
		WxID: openwallet.GenTransactionWxID2(tx.Hash, this.wm.Symbol(), contractId),
		TxID: tx.Hash,
		From: from,
		To:   to,
		//		Decimal:     18,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		Fees:        "0", //tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin:        coin,
		SubmitTime:  nowUnix,
		ConfirmTime: nowUnix,
	}

	tokenTransExtractData := &openwallet.TxExtractData{}
	tokenTransExtractData.Transaction = tokentransx
	tokenTransExtractData.TxInputs = append(tokenTransExtractData.TxInputs, &deductTxInput)

	feeTxInput := openwallet.TxInput{
		Recharge: openwallet.Recharge{
			Sid:      openwallet.GenTxInputSID(tx.Hash, this.wm.Symbol(), "", 0), //base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", tx.Hash, 0, tokenEvent.TokenFrom)))),
			CreateAt: nowUnix,
			TxID:     tx.Hash,
			Address:  tx.From,
			Coin: openwallet.Coin{
				Symbol:     this.wm.Symbol(),
				IsContract: false,
			},
			Amount:      feeprice,
			BlockHash:   tx.BlockHash,
			BlockHeight: tx.BlockHeight,
		},
	}
	from = []string{
		tx.From,
	}
	to = []string{
		tx.To,
	}

	feeTx := &openwallet.Transaction{
		WxID:        openwallet.GenTransactionWxID2(tx.Hash, this.wm.Symbol(), contractId),
		TxID:        tx.Hash,
		From:        from,
		To:          to,
		Decimal:     18,
		BlockHash:   tx.BlockHash,
		BlockHeight: tx.BlockHeight,
		Fees:        feeprice, //tx.GasPrice, //totalSpent.Sub(totalReceived).StringFixed(8),
		Coin:        coin,
		SubmitTime:  nowUnix,
		ConfirmTime: nowUnix,
	}

	feeExtractData := &openwallet.TxExtractData{}
	feeExtractData.Transaction = feeTx
	feeExtractData.TxInputs = append(feeExtractData.TxInputs, &feeTxInput)

	extractDataList = append(extractDataList, tokenTransExtractData)
	extractDataList = append(extractDataList, feeExtractData)

	return sourceKey, extractDataList, nil
}

//ExtractTransactionData 扫描一笔交易
func (this *ETHBLockScanner) ExtractTransactionData(txid string, scanAddressFunc openwallet.BlockScanAddressFunc) (map[string][]*openwallet.TxExtractData, error) {
	//result := bs.ExtractTransaction(0, "", txid, scanAddressFunc)
	tx, err := this.wm.WalletClient.EthGetTransactionByHash(txid)
	if err != nil {
		log.Errorf("get transaction by has failed, err=%v", err)
		return nil, fmt.Errorf("get transaction by has failed, err=%v", err)
	}
	tx.filterFunc = scanAddressFunc
	result, err := this.TransactionScanning(tx)
	if err != nil {
		log.Errorf("scan transaction[%v] failed, err=%v", txid, err)
		return nil, fmt.Errorf("scan transaction[%v] failed, err=%v", txid, err)
	}
	return result.extractData, nil
}

func (this *ETHBLockScanner) TransactionScanning(tx *BlockTransaction) (*ExtractResult, error) {
	//txToNotify := make(map[string][]BlockTransaction)
	if tx.BlockNumber == "" {
		return &ExtractResult{
			BlockHeight: 0,
			TxID:        "",
			extractData: make(map[string][]*openwallet.TxExtractData),
			Success:     true,
		}, nil
	}

	blockHeight, err := ConvertToUint64(tx.BlockNumber, 16)
	if err != nil {
		log.Errorf("convert block number from string to uint64 failed, err=%v", err)
		return nil, err
	}

	tx.BlockHeight = blockHeight
	var result = ExtractResult{
		BlockHeight: blockHeight,
		TxID:        tx.BlockHash,
		extractData: make(map[string][]*openwallet.TxExtractData),
		Success:     true,
	}

	tokenEvent, err := this.UpdateTxByReceipt(tx)
	if err != nil {
		log.Errorf("UpdateTxByReceipt failed, err=%v", err)
		return nil, err
	}
	//log.Debugf("get token Event:%v", tokenEvent)

	FromSourceKey, fromExtractDataList, err := this.MakeFromExtractData(tx, tokenEvent)
	if err != nil {
		log.Errorf("MakeFromExtractData failed, err=%v", err)
		return nil, err
	}

	ToSourceKey, toExtractDataList, err := this.MakeToExtractData(tx, tokenEvent)
	if err != nil {
		log.Errorf("MakeToExtractData failed, err=%v", err)
		return nil, err
	}

	if FromSourceKey == ToSourceKey && FromSourceKey != "" {
		for i, _ := range fromExtractDataList {
			for j, _ := range toExtractDataList {
				if fromExtractDataList[i].Transaction.To[0] == toExtractDataList[j].Transaction.To[0] {
					fromExtractDataList[i].TxOutputs = toExtractDataList[j].TxOutputs
				}
			}
		}

		result.extractData[FromSourceKey] = fromExtractDataList
	} else if FromSourceKey != "" && ToSourceKey != "" {
		result.extractData[FromSourceKey] = fromExtractDataList
		result.extractData[ToSourceKey] = toExtractDataList
	} else if FromSourceKey != "" {
		result.extractData[FromSourceKey] = fromExtractDataList
	} else if ToSourceKey != "" {
		result.extractData[ToSourceKey] = toExtractDataList
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
