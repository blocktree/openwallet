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

package hypercash

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/tidwall/gjson"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	blockchainBucket  = "blockchain"    //区块链数据集合
	periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20              //并发的扫描线程数
)

//BTCBlockScanner bitcoin的区块链扫描器
type BTCBlockScanner struct {
	addressInScanning    map[string]string                               //加入扫描的地址
	walletInScanning     map[string]*openwallet.Wallet                   //加入扫描的钱包
	CurrentBlockHeight   uint64                                          //当前区块高度
	scanTask             *timer.TaskTimer                                //扫描定时器
	extractingCH         chan struct{}                                   //扫描工作令牌
	mu                   sync.RWMutex                                    //读写锁
	observers            map[openwallet.BlockScanNotificationObject]bool //观察者
	scanning             bool                                            //是否扫描中
	wm                   *WalletManager                                  //钱包管理者
	IsScanMemPool        bool                                            //是否扫描交易池
	RescanLastBlockCount uint64                                          //重扫上N个区块数量
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	Recharges   []*openwallet.Recharge
	TxID        string
	BlockHeight uint64
	Success     bool
	Reason      string
}

//SaveResult 保存结果
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

//NewBTCBlockScanner 创建区块链扫描器
func NewBTCBlockScanner(wm *WalletManager) *BTCBlockScanner {
	bs := BTCBlockScanner{}
	bs.addressInScanning = make(map[string]string)
	bs.walletInScanning = make(map[string]*openwallet.Wallet)
	bs.observers = make(map[openwallet.BlockScanNotificationObject]bool)
	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = false
	bs.RescanLastBlockCount = 10
	return &bs
}

//AddAddress 添加订阅地址
func (bs *BTCBlockScanner) AddAddress(address, accountID string, wallet *openwallet.Wallet) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.addressInScanning[address] = accountID

	if _, exist := bs.walletInScanning[accountID]; exist {
		return
	}
	bs.walletInScanning[accountID] = wallet
}

//AddWallet 添加扫描钱包
func (bs *BTCBlockScanner) AddWallet(accountID string, wallet *openwallet.Wallet) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if _, exist := bs.walletInScanning[accountID]; exist {
		//已存在，不重复订阅
		return
	}

	bs.walletInScanning[accountID] = wallet

	//删除充值记录
	//wallet.DropRecharge()

	//导入钱包该账户的所有地址
	addrs := wallet.GetAddressesByAccount(accountID)
	if addrs == nil {
		return
	}

	log.Std.Info("block scanner load wallet [%s] existing addresses: %d ", accountID, len(addrs))

	for _, address := range addrs {
		bs.addressInScanning[address.Address] = accountID
	}

}

//IsExistAddress 指定地址是否已登记扫描
func (bs *BTCBlockScanner) IsExistAddress(address string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.addressInScanning[address]
	return exist
}

//IsExistWallet 指定账户的钱包是否已登记扫描
func (bs *BTCBlockScanner) IsExistWallet(accountID string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.walletInScanning[accountID]
	return exist
}

//AddObserver 添加观测者
func (bs *BTCBlockScanner) AddObserver(obj openwallet.BlockScanNotificationObject) {
	bs.mu.Lock()

	defer bs.mu.Unlock()

	if obj == nil {
		return
	}
	if _, exist := bs.observers[obj]; exist {
		//已存在，不重复订阅
		return
	}

	bs.observers[obj] = true
}

//RemoveObserver 移除观测者
func (bs *BTCBlockScanner) RemoveObserver(obj openwallet.BlockScanNotificationObject) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.observers, obj)
}

//Clear 清理订阅扫描的内容
func (bs *BTCBlockScanner) Clear() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.walletInScanning = nil
	bs.addressInScanning = nil
	bs.addressInScanning = make(map[string]string)
	bs.walletInScanning = make(map[string]*openwallet.Wallet)
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

//Run 运行
func (bs *BTCBlockScanner) Run() {

	if bs.scanning {
		return
	}

	if bs.scanTask == nil {
		//创建定时器
		task := timer.NewTask(periodOfTask, bs.scanBlock)
		bs.scanTask = task
	}
	bs.scanning = true
	bs.scanTask.Start()
}

//Stop 停止扫描
func (bs *BTCBlockScanner) Stop() {
	bs.scanTask.Stop()
	bs.scanning = false
}

//Pause 暂停扫描
func (bs *BTCBlockScanner) Pause() {
	bs.scanTask.Pause()
}

//Restart 继续扫描
func (bs *BTCBlockScanner) Restart() {
	bs.scanTask.Restart()
}

//scanning 扫描
func (bs *BTCBlockScanner) scanBlock() {

	//获取本地区块高度
	blockHeader, err := bs.GetCurrentBlockHeader()
	if err != nil {
		log.Std.Error("block scanner can not get new block height; unexpected error: %v", err)
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		//获取最大高度
		maxHeight, err := bs.wm.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			log.Std.Error("block scanner can not get rpc-server block height; unexpected error: %v", err)
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
			log.Std.Error("block scanner can not get new block hash; unexpected error: %v", err)
			break
		}

		block, err := bs.wm.GetBlock(hash)
		if err != nil {
			log.Std.Error("block scanner can not get new block data; unexpected error: %v", err)

			//记录未扫区块
			unscanRecord := NewUnscanRecord(currentHeight, "", err.Error())
			bs.SaveUnscanRecord(unscanRecord)
			log.Std.Info("block height: %d extract failed.", currentHeight)
			continue
		}

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {

			log.Std.Info("block has been fork on height: %d.", currentHeight)
			log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)

			log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//删除上一区块链的所有充值记录
			bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.wm.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.wm.GetLocalBlock(currentHeight)
			if err != nil {
				log.Std.Error("block scanner can not get local block; unexpected error: %v", err)
				break
			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.wm.SaveLocalNewBlock(localBlock.Height, localBlock.Hash)
		} else {

			err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
			if err != nil {
				log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.wm.SaveLocalNewBlock(currentHeight, currentHash)
			bs.wm.SaveLocalBlock(block)

			//通知新区块给观测者，异步处理
			go bs.newBlockNotify(block)
		}
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

	//重扫未确认的记录
	//bs.RescanUnconfirmRechargeRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *BTCBlockScanner) ScanBlock(height uint64) error {

	log.Std.Info("block scanner scanning height: %d ...", height)

	hash, err := bs.wm.GetBlockHash(height)
	if err != nil {
		//下一个高度找不到会报异常
		log.Std.Error("block scanner can not get new block hash; unexpected error: %v", err)
		return err
	}

	block, err := bs.wm.GetBlock(hash)
	if err != nil {
		log.Std.Error("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		log.Std.Info("block height: %d extract failed.", height)
		return err
	}

	err = bs.BatchExtractTransaction(block.Height, block.Hash, block.tx)
	if err != nil {
		log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	//通知新区块给观测者，异步处理
	go bs.newBlockNotify(block)

	return nil
}

//ScanTxMemPool 扫描交易内存池
func (bs *BTCBlockScanner) ScanTxMemPool() {

	log.Std.Info("block scanner scanning mempool ...")

	//提取未确认的交易单
	txIDsInMemPool, err := bs.wm.GetTxIDsInMemPool()
	if err != nil {
		log.Std.Error("block scanner can not get mempool data; unexpected error: %v", err)
	}

	err = bs.BatchExtractTransaction(0, "", txIDsInMemPool)
	if err != nil {
		log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

}

//rescanFailedRecord 重扫失败记录
func (bs *BTCBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.wm.GetUnscanRecords()
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

		log.Std.Info("block scanner rescanning height: %d ...", height)

		if len(txs) == 0 {

			hash, err := bs.wm.GetBlockHash(height)
			if err != nil {
				//下一个高度找不到会报异常
				log.Std.Error("block scanner can not get new block hash; unexpected error: %v", err)
				continue
			}

			block, err := bs.wm.GetBlock(hash)
			if err != nil {
				log.Std.Error("block scanner can not get new block data; unexpected error: %v", err)
				continue
			}

			txs = block.tx
		}

		err = bs.BatchExtractTransaction(height, hash, txs)
		if err != nil {
			log.Std.Error("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.wm.DeleteUnscanRecord(height)
	}

	//删除未没有找到交易记录的重扫记录
	bs.wm.DeleteUnscanRecordNotFindTX()
}

//RescanUnconfirmRechargeRecord
func (bs *BTCBlockScanner) RescanUnconfirmRechargeRecord() {

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

//newBlockNotify 获得新区块后，通知给观测者
func (bs *BTCBlockScanner) newBlockNotify(block *Block) {
	for o, _ := range bs.observers {
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

			//saveResult := SaveResult{}
			//saveResult.TxID = gets.TxID
			//saveResult.BlockHeight = height

			if gets.Success {
				saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
				if saveErr != nil {
					//log.Std.Error("SaveTxToWalletDB unexpected error: %v", saveErr)
					//saveResult.Success = false
					failed++ //标记保存失败数
				} else {
					//saveResult.Success = true
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, gets.TxID, gets.Reason)
				bs.SaveUnscanRecord(unscanRecord)
				log.Std.Info("block height: %d extract failed.", height)
				//saveResult.Success = false
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
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTxid)
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
		return fmt.Errorf("SaveTxToWalletDB failed")
	} else {
		return nil
	}
	//return nil
}

//extractRuntime 提取运行时
func (bs *BTCBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) error {

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

	return nil

}

//ExtractTransaction 提取交易单
func (bs *BTCBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, txid string) ExtractResult {

	var (
		transactions = make([]*openwallet.Recharge, 0)
		success      = false
		resaon       = ""
	)

	trx, err := bs.wm.GetTransaction(txid)
	if err != nil {
		log.Std.Error("block scanner can not extract transaction data; unexpected error: %v", err)
		//记录哪个区块哪个交易单没有完成扫描
		success = false
		resaon = err.Error()
		//return nil, failedTx, nil
	} else {

		realblockHash := trx.Get("blockhash").String()
		realBlockHeight := trx.Get("blockheight").Uint()
		confirmations := trx.Get("confirmations").Int()
		vout := trx.Get("vout")
		createAt := time.Now()
		for _, output := range vout.Array() {

			amount := output.Get("value").String()
			n := output.Get("n").Uint()
			addresses := output.Get("scriptPubKey.addresses").Array()
			if len(addresses) == 1 {
				addr := addresses[0].String()
				wallet, ok := bs.GetWalletByAddress(addr)
				if ok {

					a := wallet.GetAddress(addr)
					if a == nil {
						continue
					}

					log.Info("find tx for address:", a.Address, "txid:", txid, "block height:", realBlockHeight, "blockhash:", realblockHash)
					transaction := openwallet.Recharge{}
					transaction.TxID = txid
					transaction.Address = addr
					transaction.AccountID = a.AccountID
					transaction.Symbol = Symbol
					transaction.Index = n
					transaction.Amount = amount
					transaction.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("%s_%d_%s", txid, n, addr))))
					transaction.CreateAt = createAt.Unix()

					if realBlockHeight > 0 && len(realblockHash) > 0 {
						transaction.BlockHeight = realBlockHeight
						transaction.BlockHash = realblockHash
						transaction.Confirm = confirmations
					}

					//有高度记录高度信息
					if blockHeight > 0 && len(blockHash) > 0 {
						transaction.BlockHeight = blockHeight
						transaction.BlockHash = blockHash
						transaction.Confirm = confirmations
					}

					transactions = append(transactions, &transaction)

				}
			}

		}

		success = true

	}

	result := ExtractResult{
		BlockHeight: blockHeight,
		TxID:        txid,
		Recharges:   transactions,
		Success:     success,
		Reason:      resaon,
	}

	return result

}

//SaveRechargeToWalletDB 保存交易单内的充值记录到钱包数据库
func (bs *BTCBlockScanner) SaveRechargeToWalletDB(height uint64, list []*openwallet.Recharge) error {

	var (
		saveSuccess = true
	)

	for _, r := range list {

		//accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
		wallet, ok := bs.GetWalletByAddress(r.Address)
		if ok {

			//a := wallet.GetAddress(r.Address)
			//if a == nil {
			//	continue
			//}
			//
			//r.AccountID = a.AccountID
			reason := ""
			err := wallet.SaveUnreceivedRecharge(r)
			//如果blockHash没有值，添加到重扫，避免遗留
			if err != nil {
				saveSuccess = false
				//记录未扫区块
				reason = err.Error()
				log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
				unscanRecord := NewUnscanRecord(height, r.TxID, reason)

				err = bs.SaveUnscanRecord(unscanRecord)
				if err != nil {
					log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
				}

			} else {
				log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
			}

			//if err != nil || len(r.BlockHash) == 0 {
			//	saveSuccess = false
			//	//记录未扫区块
			//	if err != nil {
			//		reason = err.Error()
			//		log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
			//	}
			//	unscanRecord := NewUnscanRecord(height, r.TxID, reason)
			//	err = bs.SaveUnscanRecord(unscanRecord)
			//	if err != nil {
			//		log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
			//	}
			//
			//} else {
			//	log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
			//}
		} else {
			log.Error("address:", r.Address, "in wallet is not found, txid:", r.TxID)
			return errors.New("address in wallet is not found")
		}

	}

	if !saveSuccess {
		return errors.New("have unscan record")
	}

	return nil
}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *BTCBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash = bs.wm.GetLocalNewBlock()

	//如果本地没有记录，查询接口的高度
	if blockHeight <= 0 {
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

//DropRechargeRecords 清楚钱包的全部充值记录
func (bs *BTCBlockScanner) DropRechargeRecords(accountID string) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	wallet, ok := bs.walletInScanning[accountID]
	if !ok {
		errMsg := fmt.Sprintf("accountID: %s wallet is not found", accountID)
		return errors.New(errMsg)
	}

	return wallet.DropRecharge()
}

//DeleteRechargesByHeight 删除某区块高度的充值记录
func (bs *BTCBlockScanner) DeleteRechargesByHeight(height uint64) error {

	bs.mu.RLock()
	defer bs.mu.RUnlock()

	for _, wallet := range bs.walletInScanning {

		list, err := wallet.GetRecharges(false, height)
		if err != nil {
			return err
		}

		db, err := wallet.OpenDB()
		if err != nil {
			return err
		}

		tx, err := db.Begin(true)
		if err != nil {
			return err
		}

		for _, r := range list {
			err = db.DeleteStruct(&r)
			if err != nil {
				return err
			}
		}

		tx.Commit()

		db.Close()
	}

	return nil
}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//if record.BlockHeight == 0 {
	//	return errors.New("unconfirmed transaction do not rescan")
	//}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//GetWalletByAddress 获取地址对应的钱包
func (bs *BTCBlockScanner) GetWalletByAddress(address string) (*openwallet.Wallet, bool) {
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

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (uint64, error) {

	result, err := wm.hcdClient.Call("getblockcount", nil)
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
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
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
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)
}

//SaveLocalBlock 记录本地新区块
func (wm *WalletManager) SaveLocalBlock(block *Block) {

	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Save(block)
}

//SaveTransaction 记录高度到本地
func (wm *WalletManager) SaveTransaction(blockHeight uint64) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

//GetBlockHash 根据区块高度获得区块hash
func (wm *WalletManager) GetBlockHash(height uint64) (string, error) {

	request := []interface{}{
		height,
	}

	result, err := wm.hcdClient.Call("getblockhash", request)
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

	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
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

	request := []interface{}{
		hash,
	}

	result, err := wm.hcdClient.Call("getblock", request)
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

	result, err := wm.hcdClient.Call("getrawmempool", nil)
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
func (wm *WalletManager) GetTransaction(txid string) (*gjson.Result, error) {

	request := []interface{}{
		txid,
		1,
	}

	result, err := wm.hcdClient.Call("getrawtransaction", request)
	if err != nil {
		return nil, err
	}

	return result, nil

}

//获取未扫记录
func (wm *WalletManager) GetUnscanRecords() ([]*UnscanRecord, error) {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
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
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	for _, r := range list {
		tx.DeleteStruct(r)
	}

	return tx.Commit()
}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func (wm *WalletManager) DeleteUnscanRecordNotFindTX() error {

	//删除找不到交易单
	reason := "[-5]No information available about transaction"

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
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

//DeleteUnscanRecordByTxID 删除未扫记录
func (wm *WalletManager) DeleteUnscanRecordByTxID(height uint64, txid string) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	db.Select(q.And(
		q.Eq("TxID", txid),
		q.Eq("BlockHeight", height),
	)).Find(&list)
	//err = db.Find("TxID", txid, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		db.DeleteStruct(r)
	}

	return nil
}

//func (wm *WalletManager) ClearUnscanRecord() error {
//	//检查交易单是否完整，删除记录
//	for _, txid := range txs {
//		//获取本地区块高度
//		db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
//		if err != nil {
//			return err
//		}
//		defer db.Close()
//
//		var list []*UnscanRecord
//		err = db.Find("TxID", txid, &list)
//		if err != nil {
//			return err
//		}
//
//		for _, r := range list {
//			db.DeleteStruct(r)
//		}
//	}
//}
