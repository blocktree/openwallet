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

package openwallet

import (
	"fmt"
	"github.com/blocktree/openwallet/v2/concurrent"
	"sync"
	"time"

	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/timer"
)

// deprecated
// BlockScanAddressFunc 扫描地址是否存在算法
// @return 地址所属源标识，是否存在
type BlockScanAddressFunc func(address string) (string, bool)

// deprecated
// BlockScanTargetFunc 扫描对象是否存在算法
// @return 对象所属源标识，是否存在
type BlockScanTargetFunc func(target ScanTarget) (string, bool)

// BlockScanTargetFuncV2 扫描智能合约是否存在算法
// @return 对象所属源标识，是否存在
type BlockScanTargetFuncV2 func(target ScanTargetParam) ScanTargetResult

type ScanTarget struct {
	Address          string           //地址字符串
	PublicKey        string           //地址公钥
	Alias            string           //地址别名，可绑定用户别名
	Symbol           string           //币种类别
	BalanceModelType BalanceModelType //余额模型类别
}

// 扫描目标类型
const (
	ScanTargetTypeAccountAddress  = 0
	ScanTargetTypeAccountAlias    = 1
	ScanTargetTypeContractAddress = 2
	ScanTargetTypeContractAlias   = 3
	ScanTargetTypeAddressPubKey   = 4
	ScanTargetTypeAddressMemo     = 5
)

// ScanTargetParam 扫描目标参数
type ScanTargetParam struct {
	ScanTarget     string //地址字符串
	Symbol         string //主链类别
	ScanTargetType uint64 // 0: 账户地址，1：账户别名，2：合约地址，3：合约别名，4：地址公钥，5：地址备注
}

// ScanTargetResult 扫描结果
type ScanTargetResult struct {
	SourceKey string //关联键，地址关联的账户或ID等等
	Exist     bool   //是否存在

	//对象指针，对应ScanTargetType:
	//0：openwallet.Address
	//1：openwallet.Account
	//2: openwallet SmartContract
	//3: openwallet SmartContract
	//4: openwallet Address
	//5: openwallet Address
	TargetInfo interface{}
}

type BlockchainSyncStatus struct {
	NetworkBlockHeight uint64
	CurrentBlockHeight uint64
	Syncing            bool
}

// BlockScanner 区块扫描器
// 负责扫描新区块，给观察者推送订阅地址的新交易单。
type BlockScanner interface {

	//deprecated
	// SetBlockScanAddressFunc 设置区块扫描过程，查找地址过程方法
	SetBlockScanAddressFunc(scanAddressFunc BlockScanAddressFunc) error

	//deprecated
	// SetBlockScanTargetFunc 设置区块扫描过程，查找扫描对象过程方法
	SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

	//SetBlockScanTargetFuncV2 设置区块扫描过程，查找扫描对象过程方法
	//@required
	SetBlockScanTargetFuncV2(scanTargetFunc BlockScanTargetFuncV2) error

	//AddObserver 添加观测者
	//@optional
	AddObserver(obj BlockScanNotificationObject) error

	//RemoveObserver 移除观测者
	//@optional
	RemoveObserver(obj BlockScanNotificationObject) error

	//SetRescanBlockHeight 重置区块链扫描高度
	//@required
	SetRescanBlockHeight(height uint64) error

	//Run 运行
	//@optional
	Run() error

	//Stop 停止扫描
	//@optional
	Stop() error

	//Pause 暂停扫描
	//@optional
	Pause() error

	//Restart 继续扫描
	//@optional
	Restart() error

	//InitBlockScanner 初始化扫描器
	//@optional
	InitBlockScanner() error

	//CloseBlockScanner 关闭扫描器
	//@optional
	CloseBlockScanner() error

	//ScanBlock 扫描指定高度的区块
	//@required
	ScanBlock(height uint64) error

	//NewBlockNotify 新区块通知
	//@optional
	NewBlockNotify(header *BlockHeader) error

	//GetCurrentBlockHeight 获取当前区块高度
	//@required
	GetCurrentBlockHeader() (*BlockHeader, error)

	//GetGlobalMaxBlockHeight 获取区块链全网最大高度
	//@optional
	GetGlobalMaxBlockHeight() uint64

	//GetScannedBlockHeight 获取已扫区块高度
	//@required
	GetScannedBlockHeight() uint64

	//ExtractTransactionData 提取交易单数据
	//@required
	ExtractTransactionData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*TxExtractData, error)

	//GetBalanceByAddress 查询地址余额
	//@required
	GetBalanceByAddress(address ...string) ([]*Balance, error)

	//GetTransactionsByAddress 查询基于账户的交易记录，通过账户关系的地址
	//返回的交易记录以资产账户为集合的结果，转账数量以基于账户来计算
	//@optional
	GetTransactionsByAddress(offset, limit int, coin Coin, address ...string) ([]*TxExtractData, error)

	//SetBlockScanWalletDAI 设置区块扫描过程，上层提供一个钱包数据接口
	//@optional
	SetBlockScanWalletDAI(dai WalletDAI) error

	//SetBlockchainDAI 设置区块链数据访问接口，读取持久化的区块数据
	//@optional
	SetBlockchainDAI(dai BlockchainDAI) error

	//SupportBlockchainDAI 支持外部设置区块链数据访问接口
	//@optional
	SupportBlockchainDAI() bool

	//ExtractTransactionAndReceiptData 提取交易单及交易回执数据
	//@required
	ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFuncV2) (map[string][]*TxExtractData, map[string]*SmartContractReceipt, error)

	//GetBlockchainSyncStatus 获取当前区块链同步状态
	//@optional
	GetBlockchainSyncStatus() (*BlockchainSyncStatus, error)
}

// BlockScanNotificationObject 扫描被通知对象
type BlockScanNotificationObject interface {

	//BlockScanNotify 新区块扫描完成通知
	//@required
	BlockScanNotify(header *BlockHeader) error

	//BlockExtractDataNotify 区块提取结果通知
	//@required
	BlockExtractDataNotify(sourceKey string, data *TxExtractData) error

	//BlockExtractSmartContractDataNotify 区块提取智能合约交易结果通知
	//@param sourceKey: 为contractID
	//@param data: 合约交易回执
	//@required
	BlockExtractSmartContractDataNotify(sourceKey string, data *SmartContractReceipt) error
}

// TxExtractData 区块扫描后的交易单提取结果，每笔交易单
type TxExtractData struct {

	//消费记录，交易单输入部分
	TxInputs []*TxInput

	//充值记录，交易单输出部分
	TxOutputs []*TxOutPut

	//交易记录
	Transaction *Transaction
}

func NewBlockExtractData() *TxExtractData {
	data := TxExtractData{
		TxInputs:  make([]*TxInput, 0),
		TxOutputs: make([]*TxOutPut, 0),
	}
	return &data
}

const (
	periodOfTask = 5 * time.Second //定时任务执行隔间
)

// BlockScannerBase 区块链扫描器基本结构实现
type BlockScannerBase struct {
	AddressInScanning map[string]string                    //加入扫描的地址
	scanTask          *timer.TaskTimer                     //扫描定时器
	Mu                sync.RWMutex                         //读写锁
	Observers         map[BlockScanNotificationObject]bool //观察者
	Scanning          bool                                 //是否扫描中
	PeriodOfTask      time.Duration
	ScanAddressFunc   BlockScanAddressFunc  //区块扫描查询地址算法
	ScanTargetFunc    BlockScanTargetFunc   //区块扫描查询地址算法
	ScanTargetFuncV2  BlockScanTargetFuncV2 //区块扫描查询地址算法
	blockProducer     chan interface{}
	blockConsumer     chan interface{}
	isClose           bool //是否已关闭
	WalletDAI         WalletDAI
	BlockchainDAI     BlockchainDAI
}

// NewBTCBlockScanner 创建区块链扫描器
func NewBlockScannerBase() *BlockScannerBase {
	bs := BlockScannerBase{}
	bs.AddressInScanning = make(map[string]string)
	bs.Observers = make(map[BlockScanNotificationObject]bool)
	bs.PeriodOfTask = periodOfTask

	bs.InitBlockScanner()
	return &bs
}

// InitBlockScanner
func (bs *BlockScannerBase) InitBlockScanner() error {

	bs.blockProducer = make(chan interface{})
	bs.blockConsumer = make(chan interface{})

	go bs.newBlockNotifyConsume()
	go concurrent.ProducerToConsumerRuntime(bs.blockProducer, bs.blockConsumer)

	bs.isClose = false

	return nil
}

// deprecated
// SetBlockScanAddressFunc 设置区块扫描过程，查找地址过程方法
func (bs *BlockScannerBase) SetBlockScanAddressFunc(scanAddressFunc BlockScanAddressFunc) error {
	bs.ScanAddressFunc = scanAddressFunc
	return nil
}

// deprecated
// SetBlockScanTargetFunc 设置区块扫描过程，查找扫描对象过程方法
// @required
func (bs *BlockScannerBase) SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error {
	bs.ScanTargetFunc = scanTargetFunc

	//兼容已弃用的SetBlockScanAddressFunc
	scanAddressFunc := func(address string) (string, bool) {
		scanTarget := ScanTarget{
			Address:          address,
			BalanceModelType: BalanceModelTypeAddress,
		}
		return bs.ScanTargetFunc(scanTarget)
	}
	bs.SetBlockScanAddressFunc(scanAddressFunc)
	return nil
}

// SetBlockScanTargetFuncV2 设置区块扫描过程，查找扫描对象过程方法
// @required
func (bs *BlockScannerBase) SetBlockScanTargetFuncV2(scanTargetFuncV2 BlockScanTargetFuncV2) error {
	bs.ScanTargetFuncV2 = scanTargetFuncV2

	//兼容已弃用的SetBlockScanAddressFunc
	scanTargetFunc := func(scanTarget ScanTarget) (string, bool) {
		scanTargetParam := ScanTargetParam{Symbol: scanTarget.Symbol}
		if scanTarget.BalanceModelType == BalanceModelTypeAddress {
			scanTargetParam.ScanTarget = scanTarget.Address
			scanTargetParam.ScanTargetType = ScanTargetTypeAccountAddress
		} else {
			scanTargetParam.ScanTarget = scanTarget.Alias
			scanTargetParam.ScanTargetType = ScanTargetTypeAccountAlias
		}
		result := bs.ScanTargetFuncV2(scanTargetParam)
		return result.SourceKey, result.Exist
	}
	bs.SetBlockScanTargetFunc(scanTargetFunc)
	return nil
}

// AddObserver 添加观测者
func (bs *BlockScannerBase) AddObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()

	defer bs.Mu.Unlock()

	if obj == nil {
		return nil
	}
	if _, exist := bs.Observers[obj]; exist {
		//已存在，不重复订阅
		return nil
	}

	bs.Observers[obj] = true

	return nil
}

// RemoveObserver 移除观测者
func (bs *BlockScannerBase) RemoveObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()

	delete(bs.Observers, obj)

	return nil
}

// SetRescanBlockHeight 重置区块链扫描高度
func (bs *BlockScannerBase) SetRescanBlockHeight(height uint64) error {
	return nil
}

// SetTask
func (bs *BlockScannerBase) SetTask(task func()) {

	//if bs.scanTask == nil {
	//	//创建定时器
	//	task := timer.NewTask(bs.PeriodOfTask, task)
	//	bs.scanTask = task
	//}

	//运行中先关闭定时器
	if bs.scanTask != nil && bs.scanTask.Running() {
		bs.scanTask.Stop()
		bs.scanTask = nil
	}
	taskTimer := timer.NewTask(bs.PeriodOfTask, task)
	bs.scanTask = taskTimer
}

// Run 运行
func (bs *BlockScannerBase) Run() error {

	if bs.IsClose() {
		return fmt.Errorf("block scanner has been closed")
	}

	if bs.ScanAddressFunc == nil {
		return fmt.Errorf("BlockScanAddressFunc is not set up")
	}

	if bs.Scanning {
		log.Warn("block scanner is running... ")
		return nil
	}

	if bs.scanTask == nil {
		return fmt.Errorf("block scanner has not set scan task ")
	}
	bs.Scanning = true
	bs.scanTask.Start()
	return nil
}

// Stop 停止扫描
func (bs *BlockScannerBase) Stop() error {

	if bs.IsClose() {
		return fmt.Errorf("block scanner has been closed")
	}

	bs.scanTask.Stop()
	bs.Scanning = false
	return nil
}

// Pause 暂停扫描
func (bs *BlockScannerBase) Pause() error {

	if bs.IsClose() {
		return fmt.Errorf("block scanner has been closed")
	}

	bs.scanTask.Pause()
	bs.Scanning = false
	return nil
}

// Restart 继续扫描
func (bs *BlockScannerBase) Restart() error {

	if bs.IsClose() {
		return fmt.Errorf("block scanner has been closed")
	}

	bs.scanTask.Restart()
	bs.Scanning = true
	return nil
}

// IsClose 是否已经关闭
func (bs *BlockScannerBase) IsClose() bool {
	return bs.isClose
}

// ScanBlock 扫描指定高度区块
func (bs *BlockScannerBase) ScanBlock(height uint64) error {
	//扫描指定高度区块
	return fmt.Errorf("ScanBlock is not implemented")
}

// GetCurrentBlockHeight 获取当前区块高度
func (bs *BlockScannerBase) GetCurrentBlockHeader() (*BlockHeader, error) {
	return nil, fmt.Errorf("GetCurrentBlockHeader is not implemented")
}

// GetGlobalMaxBlockHeight 获取区块链全网最大高度
// @required
func (bs *BlockScannerBase) GetGlobalMaxBlockHeight() uint64 {
	return 0
}

// GetScannedBlockHeight 获取已扫区块高度
func (bs *BlockScannerBase) GetScannedBlockHeight() uint64 {
	return 0
}

func (bs *BlockScannerBase) ExtractTransactionData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*TxExtractData, error) {
	return nil, fmt.Errorf("ExtractTransactionData is not implemented")
}

// ExtractTransactionAndReceiptData 提取交易单及交易回执数据
// @required
func (bs *BlockScannerBase) ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFuncV2) (map[string][]*TxExtractData, map[string]*SmartContractReceipt, error) {
	return nil, nil, fmt.Errorf("ExtractTransactionAndReceiptData is not implemented")
}

// GetBalanceByAddress 查询地址余额
func (bs *BlockScannerBase) GetBalanceByAddress(address ...string) ([]*Balance, error) {
	return nil, fmt.Errorf("GetBalanceByAddress is not implemented")
}

//GetTokenBalanceByAddress 查询地址token余额列表
//func (bs *BlockScannerBase) GetTokenBalanceByAddress(address ...string) ([]*TokenBalance, error) {
//	return nil, nil
//}

// GetTransactionsByAddress 查询基于账户的交易记录，通过账户关系的地址
// 返回的交易记录以资产账户为集合的结果，转账数量以基于账户来计算
func (bs *BlockScannerBase) GetTransactionsByAddress(offset, limit int, coin Coin, address ...string) ([]*TxExtractData, error) {
	return nil, fmt.Errorf("GetTransactionsByAddress is not implemented")
}

// SetBlockScanWalletDAI 设置区块扫描过程，上层提供一个钱包数据接口
// @optional
func (bs *BlockScannerBase) SetBlockScanWalletDAI(dai WalletDAI) error {
	bs.WalletDAI = dai
	return nil
}

// SupportBlockchainDAI 支持外部设置区块链数据访问接口
// @optional
func (bs *BlockScannerBase) SupportBlockchainDAI() bool {
	return false
}

// SetBlockchainDAI 设置区块链数据访问接口，读取持久化的区块数据
// @optional
func (bs *BlockScannerBase) SetBlockchainDAI(dai BlockchainDAI) error {
	bs.BlockchainDAI = dai
	return nil
}

// NewBlockNotify 获得新区块后，发送到通知通道
func (bs *BlockScannerBase) NewBlockNotify(block *BlockHeader) error {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()
	if !bs.IsClose() {
		bs.blockProducer <- block
	}

	return nil
}

// CloseBlockScanner 关闭扫描器
func (bs *BlockScannerBase) CloseBlockScanner() error {

	//保证只关闭一次
	//bs.closeOnce.Do(func() {
	bs.Stop()

	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	bs.isClose = true
	close(bs.blockProducer)
	close(bs.blockConsumer)
	//})

	return nil
}

// newBlockNotifyConsume
func (bs *BlockScannerBase) newBlockNotifyConsume() {

	for {
		select {
		case obj, exist := <-bs.blockConsumer:
			if !exist {
				//log.Warning("newBlockNotifyConsume closed")
				return
			}
			header, ok := obj.(*BlockHeader)
			if ok {
				for o, _ := range bs.Observers {
					o.BlockScanNotify(header)
				}
			}
		}

	}
}

// GetBlockchainSyncStatus 获取当前区块链同步状态
// @optional
func (bs *BlockScannerBase) GetBlockchainSyncStatus() (*BlockchainSyncStatus, error) {
	return nil, fmt.Errorf("GetBlockchainSyncStatus is not implemented")
}
