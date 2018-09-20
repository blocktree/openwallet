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

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/asdine/storm"

	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/timer"
)

const (
	maxExtractingSize = 20              //并发的扫描线程数
	periodOfTask      = 5 * time.Second //定时任务执行隔间
)

type ETHBlockScanner struct {
	*openwallet.BlockScannerBase
	addressInScanning  map[string]string
	walletInScanning   map[string]*Wallet
	currentBlockHeight uint64
	scanTask           *timer.TaskTimer
	extractingCH       chan struct{}
	mu                 sync.RWMutex
	observers          map[openwallet.BlockScanNotificationObject]bool
	scanning           bool
	wmanager           *WalletManager
}

func NewETHBlockScanner(wm *WalletManager) *ETHBlockScanner {
	bs := &ETHBlockScanner{}
	bs.addressInScanning = make(map[string]string)
	bs.walletInScanning = make(map[string]*Wallet)
	bs.observers = make(map[openwallet.BlockScanNotificationObject]bool)
	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	return bs
}

func (this *ETHBlockScanner) AddAddress(address string, accountID string,
	wallet *Wallet) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.addressInScanning[strings.ToLower(address)] = accountID

	if _, exist := this.walletInScanning[accountID]; exist {
		return
	}
	this.walletInScanning[accountID] = wallet
}

func (this *ETHBlockScanner) AddWallet(accountId string, wallet *Wallet) error {
	addrs, err := GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("get addresses by wallet[%v] failed, err = %v", accountId, err)
		return err
	}
	this.addWallet(accountId, addrs, wallet)
	return nil
}

func (this *ETHBlockScanner) RetrieveWallet(accountID string) *Wallet {
	this.mu.RLock()
	defer this.mu.RUnlock()
	var w *Wallet
	w = this.walletInScanning[accountID]
	return w
}

func (this *ETHBlockScanner) addWallet(accountID string, addrs []*Address,
	wallet *Wallet) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if _, exist := this.walletInScanning[accountID]; exist {
		openwLogger.Log.Infof("account[%v] already exist wallets of eth scanner", accountID)
		return
	}

	this.walletInScanning[accountID] = wallet

	for _, address := range addrs {
		this.addressInScanning[strings.ToLower(address.Address)] = accountID
	}
}

func (this *ETHBlockScanner) IsExistAddress(address string) bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, exist := this.addressInScanning[strings.ToLower(address)]
	return exist
}

func (this *ETHBlockScanner) IsExistWallet(accountId string) bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, exist := this.walletInScanning[accountId]
	return exist
}

func (this *ETHBlockScanner) AddObserver(obj openwallet.BlockScanNotificationObject) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if _, exist := this.observers[obj]; exist {
		return
	}
	this.observers[obj] = true
}

func (this *ETHBlockScanner) RemoveObserver(obj openwallet.BlockScanNotificationObject) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.observers, obj)
}

func (this *ETHBlockScanner) Clear() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.walletInScanning = make(map[string]*Wallet)
	this.addressInScanning = make(map[string]string)
}

func (this *ETHBlockScanner) Run() {
	if this.scanning {
		return
	}

	if this.scanTask == nil {
		task := timer.NewTask(periodOfTask, this.ScanBlock)
		this.scanTask = task
	}

	this.scanning = true
	this.scanTask.Start()
}

func (this *ETHBlockScanner) GetWalletByAddress(address string) (*Wallet, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	account, ok := this.addressInScanning[strings.ToLower(address)]
	if ok {
		wallet, ok := this.walletInScanning[account]
		return wallet, ok

	} else {
		return nil, false
	}
}

func (this *ETHBlockScanner) RescanFailedTransactions() (map[string][]BlockTransaction, error) {
	txs, err := this.wmanager.GetAllUnscannedTransactions()
	if err != nil {
		openwLogger.Log.Errorf("GetAllUnscannedTransactions failed. err=%v", err)
		return nil, err
	}
	return this.TransactionScanning(txs)
}

func (this *ETHBlockScanner) TransactionScanning(transactions []BlockTransaction) (map[string][]BlockTransaction, error) {
	txToNotify := make(map[string][]BlockTransaction)
	for _, tx := range transactions {
		var fromWalletId string
		var toWalletId string
		if this.IsExistAddress(tx.From) {
			openwLogger.Log.Debugf("found a transaction [%v] whose from account belong to scanning wallet.", tx.Hash)
			if w, exist := this.GetWalletByAddress(tx.From); !exist {
				panic(22)
			} else {
				fromWalletId = w.WalletID
				if _, exist = txToNotify[fromWalletId]; !exist {
					txArray := make([]BlockTransaction, 0)
					txToNotify[fromWalletId] = txArray
				}
				txToNotify[fromWalletId] = append(txToNotify[fromWalletId], tx)
			}

		} else {
			openwLogger.Log.Debugf("tx.from[%v] not found in scanning address.", tx.From)
		}

		if this.IsExistAddress(tx.To) {
			openwLogger.Log.Debugf("found a transaction [%v] whose to account belong to scanning wallet.", tx.Hash)
			if w, exist := this.GetWalletByAddress(tx.To); !exist {
				panic(22)
			} else {
				toWalletId = w.WalletID
				if fromWalletId == toWalletId {
					continue
				}
				if _, exist = txToNotify[toWalletId]; !exist {
					txArray := make([]BlockTransaction, 0)
					txToNotify[toWalletId] = txArray
				}
				txToNotify[toWalletId] = append(txToNotify[toWalletId], tx)
			}

		} else {
			openwLogger.Log.Debugf("tx.to[%v] not found in scanning address.", tx.To)
		}
	}

	for walletId, _ := range txToNotify {
		wallet := this.RetrieveWallet(walletId)
		err := wallet.SaveTransactions(txToNotify[walletId])
		if err != nil {
			openwLogger.Log.Errorf("save wallet[%v] transaction failed, err=%v", walletId, err)
			return txToNotify, err
		}
	}

	if len(transactions) > 0 {
		openwLogger.Log.Debugf("transactions in block:%v", transactions)
		openwLogger.Log.Debugf("txToNotify:%v", txToNotify)
	}
	return txToNotify, nil
}

func (this *ETHBlockScanner) ExtractTransactions(block *EthBlock) (map[string][]BlockTransaction, error) {
	return this.TransactionScanning(block.Transactions)
}

func (this *ETHBlockScanner) DeleteTransactionsByHeight(height *big.Int) error {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, w := range this.walletInScanning {
		err := w.DeleteTransactionByHeight(height)
		if err != nil {
			openwLogger.Log.Errorf("delete wallet[%v] transaction for block height[%v] failed, err= %v", w.WalletID, height, err)
			return err
		}
	}
	return nil
}

func (this *ETHBlockScanner) ScanBlock() {
	curBlock, err := this.GetCurrentBlock()
	if err != nil {
		openwLogger.Log.Errorf("get block height from db failed, err=%v", err)
		return
	}

	openwLogger.Log.Infof("get current block:%v", curBlock)
	previousHeight := big.NewInt(0)
	for {
		curBlockHeight := curBlock.blockHeight
		curBlockHash := curBlock.BlockHash

		maxBlockHeight, err := ethGetBlockNumber()
		if err != nil {
			openwLogger.Log.Errorf("block scanner cannot get block height through RPC, err=%v", err)
			break
		}

		fmt.Println("current block height:", "0x"+curBlockHeight.Text(16), " maxBlockHeight:", "0x"+maxBlockHeight.Text(16))
		if curBlockHeight.Cmp(maxBlockHeight) == 0 {
			openwLogger.Log.Infof("block scanner has done with scan. current height:%v", "0x"+maxBlockHeight.Text(16))
			break
		}

		//扫描下一个区块
		curBlockHeight = curBlockHeight.Add(curBlockHeight, big.NewInt(1))

		openwLogger.Log.Infof("block scanner try to scan block No.%v", curBlockHeight)

		curBlock, err = ethGetBlockSpecByBlockNum(curBlockHeight, true)
		if err != nil {
			openwLogger.Log.Errorf("ethGetBlockSpecByBlockNum failed, err = %v", err)
			break
		}

		if curBlock.PreviousHash != curBlockHash {
			previousHeight = previousHeight.Sub(curBlockHeight, big.NewInt(1))
			openwLogger.Log.Infof("block has been fork on height: %v.", "0x"+curBlockHeight.Text(16))
			openwLogger.Log.Infof("block height: %v local hash = %v ", "0x"+previousHeight.Text(16), curBlockHash)
			openwLogger.Log.Infof("block height: %v mainnet hash = %v ", "0x"+previousHeight.Text(16), curBlock.PreviousHash)

			openwLogger.Log.Infof("delete recharge records on block height: %v.", "0x"+previousHeight.Text(16))
			err = this.DeleteTransactionsByHeight(previousHeight)
			if err != nil {
				openwLogger.Log.Errorf("DeleteTransactionsByHeight failed, height=%v, err=%v", "0x"+previousHeight.Text(16), err)
				break
			}

			err = this.wmanager.DeleteUnscannedTransaction(previousHeight)
			if err != nil {
				openwLogger.Log.Errorf("DeleteUnscannedTransaction failed, height=%v, err=%v", "0x"+previousHeight.Text(16), err)
				break
			}

			curBlockHeight.Sub(previousHeight, big.NewInt(1))

			curBlock, err = this.wmanager.RecoverBlockHeader(curBlockHeight)
			if err != nil {
				openwLogger.Log.Errorf("RecoverBlockHeader failed, block number=%v, err=%v", "0x"+curBlockHeight.Text(16), err)
				break
			}

			openwLogger.Log.Infof("rescan block on height: %v, hash: %v .", curBlock.BlockNumber, curBlock.BlockHash)
			err = this.wmanager.UpdateLocalBlockHeight(curBlockHeight)
			if err != nil {
				openwLogger.Log.Errorf("update local block height failed, err=%v", err)
				break
			}

		} else {
			txToNotify, err := this.ExtractTransactions(curBlock)
			if err != nil {
				openwLogger.Log.Errorf("extract transactions failed, err=%v", err)
				break
			}

			err = this.wmanager.SaveBlockHeader2(curBlock)
			if err != nil {
				openwLogger.Log.Errorf("save block failed, err=%v", err)
				break
			}
			this.Notify(txToNotify)
		}
	}

}

func (this *ETHBlockScanner) Notify(txs interface{}) {

}

func (this *ETHBlockScanner) GetCurrentBlock() (*EthBlock, error) {
	blockHeight, err := this.wmanager.GetLocalBlockHeight()
	var block *EthBlock
	if err != nil && err != storm.ErrNotFound {
		openwLogger.Log.Errorf("GetLocalBlockHeight faield, err=%v", err)
		return nil, err
	} else if err == storm.ErrNotFound {
		blockHeight, err = ethGetBlockNumber()
		if err != nil {
			openwLogger.Log.Errorf("get block height failed, err=%v", err)
			return nil, err
		}
		blockHeight = blockHeight.Sub(blockHeight, big.NewInt(1))
		block, err = ethGetBlockSpecByBlockNum(blockHeight, false)
		if err != nil {
			openwLogger.Log.Errorf("ethGetBlockSpecByBlockNum failed, err=%v", err)
			return nil, err
		}
	} else {
		block, err = this.wmanager.RecoverBlockHeader(blockHeight)
		if err != nil {
			openwLogger.Log.Errorf("RecoverBlockHeader failed, block number=%v, err=%v", blockHeight, err)
			return nil, err
		}
	}

	return block, nil
}

func (this *ETHBlockScanner) SetLocalBlock(blockheight string) error {

	block, err := ethGetBlockSpecByBlockNum2(blockheight, false)
	if err != nil {
		openwLogger.Log.Errorf("get spec of block [%v] failed, err=%v", blockheight, err)
		return err
	}

	err = this.wmanager.SaveBlockHeader(block)
	if err != nil {
		openwLogger.Log.Errorf("save block header, err = %v", err)
		return err
	}

	err = this.wmanager.UpdateLocalBlockHeight(block.blockHeight)
	if err != nil {
		openwLogger.Log.Errorf("update local block height failed, err=%v", err)
		return err
	}
	return nil
}

func PrepareForBlockScanTest(fromAddr []string, passwords []string) (string, error) {

	beforeBlockNum, err := ethGetBlockNumber()
	if err != nil {
		openwLogger.Log.Errorf("get block number failed, err=%v", err)
		return "", err
	}

	openwLogger.Log.Debugf("get block number[%v] before transactions made.", "0x"+beforeBlockNum.Text(16))

	accounts, err := ethGetAccounts()
	if err != nil {
		openwLogger.Log.Errorf("get accounts failed, err=%v", err)
		return "", err
	}

	value, err := toHexBigIntForEtherTrans("0x1", 16, TRNAS_AMOUNT_UNIT_ETHER)
	if err != nil {
		openwLogger.Log.Errorf("toHexBigIntForEtherTrans failed, err = %v", err)
		return "", err
	}

	txs := make([]string, 0, 20)
	for i, from := range fromAddr {
		for _, to := range accounts {
			if from != to {
				tx, err := SendTransactionToAddr(makeSimpleTransactiomnPara2(from, to, value, passwords[i]))
				if err != nil {
					openwLogger.Log.Errorf("send transaction from [%v] to [%v] failed, err=%v", from, to, err)
					return "", err
				}
				//openwLogger.Log.Infof("done with transaction %v", tx)
				txs = append(txs, tx)
			}
		}
	}

	for {
		pendingNum, _, err := ethGetTxpoolStatus()
		if err != nil {
			openwLogger.Log.Errorf("get txpool statusl failed, err=%v", err)
			break
		}
		if pendingNum != 0 {
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}

	walletmanage := &WalletManager{}
	err = walletmanage.UpdateLocalBlockHeight(beforeBlockNum)
	if err != nil {
		openwLogger.Log.Errorf("update local block height failed, err = %v", err)
		return "", err
	}

	blockNum, err := walletmanage.GetLocalBlockHeight()
	if err != nil {
		openwLogger.Log.Errorf("get current block number failed, err= %v", err)
		return "", err
	}

	openwLogger.Log.Debugf("transactions [%v] have been sent. ", txs)
	//openwLogger.Log.Debugf("current local block number is %v", "0x"+blockNum.Text(16))
	block, err := ethGetBlockSpecByBlockNum(blockNum, false)
	if err != nil {
		openwLogger.Log.Errorf("get spec of block [%v] failed, err=%v", "0x"+blockNum.Text(16), err)
		return "", err
	}

	err = walletmanage.SaveBlockHeader(block)
	if err != nil {
		openwLogger.Log.Errorf("save block header, err = %v", err)
		return "", err
	}
	//DumpBlockScanDb()

	return block.BlockNumber, nil
}
