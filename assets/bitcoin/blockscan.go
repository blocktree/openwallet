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
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"log"
	"path/filepath"
	"sync"
	"time"
)

/*
	步骤：
	1.添加需要扫块的钱包，及传入初始高度，-1为本地高度。
	2.获取已扫描的本地高度。
	3.获取高度+1的区块hash，通过区块链hash获取区块链数据，获取mempool数据。
	4.判断区块链的父区块hash是否与本地上一区块hash一致。
	5.解析新区块链的交易单数组。
	6.遍历交易单结构，检查每个output地址是否存在钱包的地址表中
	7.检查地址是否合法，存在地址表，生成充值记录。
	8.定时程推送充值记录到钱包的充值通道。先检查交易hash是否存在区块中。
	9.接口返回确认，标记充值记录已确认。
*/

const (
	//区块链数据集合
	blockchainBucket = "blockchain"
	//定时任务执行隔间
	periodOfTask      = 5 * time.Second
	maxExtractingSize = 1
)

type BTCBlockScanner struct {
	addressInScanning  map[string]string //加入扫描的钱包资产账户
	CurrentBlockHeight uint64            //当前区块高度
	isScanning         bool
	scanTask           *timer.TaskTimer
	extractingCH       chan struct{}
	mu                 sync.RWMutex
}

type ExtractResult struct {
	Recharges   []*openwallet.Recharge
	TxID        string
	BlockHeight uint64
	Success     bool
}

//exportTRXS 导出交易单回调函数
//@param  txs 每扫完区块链，与地址相关的交易到
type exportTRXS func(txs []*openwallet.Transaction) []string

//NewBTCBlockScanner 创建区块链扫描器
func NewBTCBlockScanner() *BTCBlockScanner {
	bs := BTCBlockScanner{}
	bs.addressInScanning = make(map[string]string)
	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	return &bs
}

//AddAddress 添加订阅地址
func (bs *BTCBlockScanner) AddAddress(address string, accountID string) {
	bs.mu.Lock()
	bs.addressInScanning[address] = accountID
	bs.mu.Unlock()
}

//Run 运行
func (bs *BTCBlockScanner) Run() {

	if bs.scanTask == nil {
		//创建定时器
		task := timer.NewTask(periodOfTask, bs.scanning)
		bs.scanTask = task
	}

	bs.scanTask.Start()
}

//Stop 停止扫描
func (bs *BTCBlockScanner) Stop() {
	bs.scanTask.Stop()
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
func (bs *BTCBlockScanner) scanning() {

	//获取本地区块高度
	currentHeight, currentHash, err := bs.GetCurrentBlockHeight()
	if err != nil {
		log.Printf("block scanner can not get new block height; unexpected error: %v\n", err)
	}

	for {

		//if currentHeight > 1355044 {
		//	return
		//}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		log.Printf("block scanner scanning height: %d ...\n", currentHeight)

		hash, err := GetBlockHash(currentHeight)
		if err != nil {
			//下一个高度找不到会报异常
			log.Printf("block scanner can not get new block hash; unexpected error: %v\n", err)
			return
		}

		block, err := GetBlock(hash)
		if err != nil {
			log.Printf("block scanner can not get new block data; unexpected error: %v\n", err)

			//记录未扫区块
			unscanRecord := UnscanRecords{
				BlockHeight: currentHeight,
				TxID:        "",
				Reason:      err.Error(),
			}
			bs.SaveUnscanRecord(&unscanRecord)

			continue
		}

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {
			//TODO:回滚之前的区块，直到分叉起始处，并删除分叉区块链的相关充值记录
			//删除上一区块链的所有充值记录
		}

		err = bs.BatchExtractTransaction(block.Height, block.Tx)
		if err != nil {
			log.Printf("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
		}
		//for _, txid := range block.Tx {
		//	//bs.extractingCH <- true
		//	bs.ExtractRechargeRecords(currentHeight, txid)
		//	//err = bs.ExtractRechargeRecords(currentHeight, txid)
		//	//if err != nil {
		//	//	log.Printf("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
		//	//}
		//
		//}

		//保存本地新高度
		SaveLocalNewBlock(currentHeight, hash)
		SaveLocalBlock(block)

	}

	txIDsInMemPool, err := GetTxIDsInMemPool()
	if err != nil {
		log.Printf("block scanner can not get mempool data; unexpected error: %v\n", err)
	}

	err = bs.BatchExtractTransaction(0, txIDsInMemPool)
	if err != nil {
		log.Printf("block scanner can not extractRechargeRecords; unexpected error: %v\n", err)
	}

}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *BTCBlockScanner) BatchExtractTransaction(blockHeight uint64, txs []string) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
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
				saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
				if saveErr != nil {
					log.Printf("SaveTxToWalletDB unexpected error: %v", saveErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := UnscanRecords{
					BlockHeight: height,
					TxID:        "",
					Reason:      "",
				}
				bs.SaveUnscanRecord(&unscanRecord)
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				//log.Printf("done = %d, shouldDone = %d \n", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(mbs *BTCBlockScanner, eblockHeight uint64, mTxs []string, eProducer chan ExtractResult) {
		for _, txid := range mTxs {
			mbs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlockHeight uint64, mTxid string, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- mbs.ExtractTransaction(mBlockHeight, mTxid)
				//释放
				<-end

			}(eblockHeight, txid, mbs.extractingCH, eProducer)
		}
	}

	/*	开启导出的线程	*/

	//独立线程运行
	go saveWork(blockHeight, worker)

	//独立线程运行
	go extractWork(bs, blockHeight, txs, producer)

	values := make([]ExtractResult, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult

		//当数据队列有数据时，释放顶部，激活消费
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
			//log.Printf("completed %d", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Printf("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			log.Printf("block have been scanned!")
			return nil
		}
	}

	return nil
}

//ExtractTransaction 提取交易单
func (bs *BTCBlockScanner) ExtractTransaction(blockHeight uint64, txid string) ExtractResult {

	var (
		transactions = make([]*openwallet.Recharge, 0)
		success      = false
	)

	trx, err := GetTransaction(txid)
	if err != nil {
		log.Printf("block scanner can not extract transaction data; unexpected error: %v\n", err)
		//记录哪个区块哪个交易单没有完成扫描
		success = false
		//return nil, failedTx, nil
	} else {

		blockhash := trx.Get("blockhash").String()
		confirmations := trx.Get("confirmations").Int()
		vout := trx.Get("vout")

		for _, output := range vout.Array() {

			amount := output.Get("value").String()
			n := output.Get("n").Uint()
			addresses := output.Get("scriptPubKey.addresses").Array()
			if len(addresses) > 0 {
				addr := addresses[0].String()
				accountID, ok := bs.addressInScanning[addr]

				if ok {

					transaction := openwallet.Recharge{}
					transaction.TxID = txid
					transaction.Address = addr
					transaction.AccountID = accountID
					transaction.Confirm = confirmations
					transaction.BlockHash = blockhash
					transaction.Amount = amount
					transaction.BlockHeight = blockHeight
					transaction.Symbol = Symbol
					transaction.Index = n
					transaction.Sid = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%s_%d_%s", txid, n, addr))))

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
	}

	return result

}

func (bs *BTCBlockScanner) SaveRechargeToWalletDB(height uint64, list []*openwallet.Recharge) error {

	for _, r := range list {

		//accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
		accountID, ok := bs.addressInScanning[r.Address]
		if ok {
			r.AccountID = accountID
			wallet, err := GetWallet(accountID)
			if err != nil {

				//记录未扫区块
				unscanRecord := UnscanRecords{
					BlockHeight: height,
					TxID:        "",
					Reason:      err.Error(),
				}
				bs.SaveUnscanRecord(&unscanRecord)

				return err
			}

			err = wallet.SaveRecharge(r)
			if err != nil {
				//保存为未扫记录

				//记录未扫区块
				unscanRecord := UnscanRecords{
					BlockHeight: height,
					TxID:        r.TxID,
					Reason:      err.Error(),
				}
				bs.SaveUnscanRecord(&unscanRecord)
			}

		} else {
			return errors.New("address in wallet is not found")
		}

	}

	return nil
}

//GetCurrentBlockHeight 获取当前区块高度
func (bs *BTCBlockScanner) GetCurrentBlockHeight() (uint64, string, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash = GetLocalNewBlock()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeight, err = GetBlockHeight()
		if err != nil {
			return 0, "", err
		}

		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		hash, err = GetBlockHash(blockHeight)
		if err != nil {
			return 0, "", err
		}
	}

	return blockHeight, hash, nil
}

//extractRechargeRecords 从交易单中提取充值记录
func (bs *BTCBlockScanner) ExtractRechargeRecords(blockHeight uint64, txid string) error {

	var (
		transaction openwallet.Recharge
		saved       bool = true
		err         error
		trx         *gjson.Result
	)

	trx, err = GetTransaction(txid)
	if err != nil {
		log.Printf("block scanner can not extract transaction data; unexpected error: %v\n", err)
		//记录哪个区块哪个交易单没有完成扫描
		saved = false
	} else {

		blockhash := trx.Get("blockhash").String()
		confirmations := trx.Get("confirmations").Int()
		vout := trx.Get("vout")

		for _, output := range vout.Array() {

			amount := output.Get("value").String()
			n := output.Get("n").Uint()
			addresses := output.Get("scriptPubKey.addresses").Array()
			if len(addresses) > 0 {
				addr := addresses[0].String()
				//_, ok := bs.addressInScanning[addr]
				//
				//if ok {

				transaction = openwallet.Recharge{}
				transaction.TxID = txid
				transaction.Address = addr
				transaction.Confirm = confirmations
				transaction.BlockHash = blockhash
				transaction.Amount = amount
				transaction.BlockHeight = blockHeight
				transaction.Symbol = Symbol
				transaction.Index = n
				transaction.Sid = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%s_%d_%s", txid, n, addr))))
				//写入数据库地址相关的钱包数据库
				err = bs.SaveTxToWalletDB(&transaction)
				if err != nil {
					log.Printf("SaveTxToWalletDB unexpected error: %v", err)
					saved = false
				} else {
					saved = true
				}

				//}
			}

		}

	}

	//保存不成功加入到重扫表中
	if !saved {
		unscanRecord := UnscanRecords{
			BlockHeight: blockHeight,
			TxID:        txid,
			Reason:      err.Error(),
		}
		bs.SaveUnscanRecord(&unscanRecord)
	} else {
		err = nil
	}

	//<-ch
	//log.Printf("txid: %s extract finished\n", txid)

	return err
}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveTxToWalletDB(tx *openwallet.Recharge) error {

	if tx == nil {
		return errors.New("the transaction to save is nil")
	}
	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	//accountID, ok := bs.addressInScanning[tx.Address]
	//if ok {
	tx.AccountID = accountID
	wallet, err := GetWallet(accountID)
	if err != nil {
		return err
	}

	return wallet.SaveRecharge(tx)
	//} else {
	//	return errors.New("address in wallet is not found")
	//}

}

//DropRechargeRecords 清楚钱包的全部充值记录
func (bs *BTCBlockScanner) DropRechargeRecords(accountID string) error {
	wallet, err := GetWalletInfo(accountID)
	if err != nil {
		return err
	}

	return wallet.DropRecharge()
}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveUnscanRecord(record *UnscanRecords) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//GetBlockHeight 获取区块链高度
func GetBlockHeight() (uint64, error) {

	result, err := client.Call("getblockcount", nil)
	if err != nil {
		return 0, err
	}

	return result.Uint(), nil
}

//GetLocalNewBlock 获取本地记录的区块高度和hash
func GetLocalNewBlock() (uint64, string) {

	var (
		blockHeight uint64 = 0
		blockHash   string = ""
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return 0, ""
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalNewBlock 记录区块高度和hash到本地
func SaveLocalNewBlock(blockHeight uint64, blockHash string) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)
}

//SaveLocalBlock 记录本地新区块
func SaveLocalBlock(block *Block) {

	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Save(block)
}

//SaveTransaction 记录高度到本地
func SaveTransaction(blockHeight uint64) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

//GetBlockHash 根据区块高度获得区块hash
func GetBlockHash(height uint64) (string, error) {

	request := []interface{}{
		height,
	}

	result, err := client.Call("getblockhash", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

//GetLocalBlock 获取本地区块数据
func GetLocalBlock(hash string) (*Block, error) {

	var (
		block *Block
	)

	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Find("Hash", hash, &block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

//GetBlock 获取区块数据
func GetBlock(hash string) (*Block, error) {

	request := []interface{}{
		hash,
	}

	result, err := client.Call("getblock", request)
	if err != nil {
		return nil, err
	}

	return NewBlock(result), nil
}

//GetTxIDsInMemPool 获取待处理的交易池中的交易单IDs
func GetTxIDsInMemPool() ([]string, error) {

	var (
		txids = make([]string, 0)
	)

	result, err := client.Call("getrawmempool", nil)
	if err != nil {
		return nil, err
	}

	for _, txid := range result.Array() {
		txids = append(txids, txid.String())
	}

	return txids, nil
}

//GetTransaction 获取交易单
func GetTransaction(txid string) (*gjson.Result, error) {

	request := []interface{}{
		txid,
		true,
	}

	result, err := client.Call("getrawtransaction", request)
	if err != nil {
		return nil, err
	}

	return result, nil

}
