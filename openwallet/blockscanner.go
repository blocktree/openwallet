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

package openwallet

import (
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/timer"
	"sync"
	"time"
)

// BlockScanAddressFunc 扫描地址是否存在算法
// @return 地址所属源标识，是否存在
type BlockScanAddressFunc func(address string) (string, bool)

// BlockScanner 区块扫描器
// 负责扫描新区块，给观察者推送订阅地址的新交易单。
type BlockScanner interface {

	//AddWallet 添加扫描地址
	//@param address 地址
	//@param sourceKey 数据源标识，可以是地址所属的应用钱包的唯一标识，资产账户唯一标识
	AddAddress(address, sourceKey string) error

	//AddObserver 添加观测者
	AddObserver(obj BlockScanNotificationObject) error

	//RemoveObserver 移除观测者
	RemoveObserver(obj BlockScanNotificationObject) error

	//Clear 清理订阅扫描的内容
	Clear() error

	//SetRescanBlockHeight 重置区块链扫描高度
	SetRescanBlockHeight(height uint64) error

	//Run 运行
	Run() error

	//Stop 停止扫描
	Stop() error

	//Pause 暂停扫描
	Pause() error

	//Restart 继续扫描
	Restart() error

	//ScanBlock 扫描指定高度的区块
	ScanBlock(height uint64) error

	//GetCurrentBlockHeight 获取当前区块高度
	GetCurrentBlockHeader() (*BlockHeader, error)

	//IsExistAddress 指定地址是否已登记扫描
	IsExistAddress(address string) bool

	//GetScannedBlockHeight 获取已扫区块高度
	GetScannedBlockHeight() uint64

	//ExtractTransactionData 提取交易单数据
	ExtractTransactionData(txid string) (map[string]*TxExtractData, error)

	//GetBalanceByAddress 查询地址余额
	GetBalanceByAddress(address ...string) ([]*Balance, error)

	//GetTransactionsByAddress 查询基于账户的交易记录，通过账户关系的地址
	//返回的交易记录以资产账户为集合的结果，转账数量以基于账户来计算
	GetTransactionsByAddress(offset, limit int, coin Coin, address ...string) ([]*TxExtractData, error)
}

//BlockScanNotificationObject 扫描被通知对象
type BlockScanNotificationObject interface {

	//BlockScanNotify 新区块扫描完成通知
	BlockScanNotify(header *BlockHeader) error

	//BlockExtractDataNotify 区块提取结果通知
	BlockExtractDataNotify(sourceKey string, data *TxExtractData) error
}

//TxExtractData 区块扫描后的交易单提取结果，每笔交易单
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

//BlockScannerBase 区块链扫描器基本结构实现
type BlockScannerBase struct {
	AddressInScanning map[string]string                    //加入扫描的地址
	scanTask          *timer.TaskTimer                     //扫描定时器
	Mu                sync.RWMutex                         //读写锁
	Observers         map[BlockScanNotificationObject]bool //观察者
	Scanning          bool                                 //是否扫描中
	PeriodOfTask      time.Duration
}

//NewBTCBlockScanner 创建区块链扫描器
func NewBlockScannerBase() *BlockScannerBase {
	bs := BlockScannerBase{}
	bs.AddressInScanning = make(map[string]string)
	//bs.WalletInScanning = make(map[string]*WalletWrapper)
	bs.Observers = make(map[BlockScanNotificationObject]bool)
	bs.PeriodOfTask = periodOfTask
	return &bs
}

//AddAddress 添加订阅地址
func (bs *BlockScannerBase) AddAddress(address, sourceKey string) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	bs.AddressInScanning[address] = sourceKey
	return nil
	//if _, exist := bs.WalletInScanning[sourceKey]; exist {
	//	return
	//}
	//bs.WalletInScanning[sourceKey] = wrapper
}

//IsExistAddress 指定地址是否已登记扫描
func (bs *BlockScannerBase) IsExistAddress(address string) bool {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()

	_, exist := bs.AddressInScanning[address]
	return exist
}

//IsExistWallet 指定账户的钱包是否已登记扫描
//func (bs *BlockScannerBase) IsExistWallet(accountID string) bool {
//	bs.Mu.RLock()
//	defer bs.Mu.RUnlock()
//
//	_, exist := bs.WalletInScanning[accountID]
//	return exist
//}

//AddObserver 添加观测者
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

//RemoveObserver 移除观测者
func (bs *BlockScannerBase) RemoveObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()

	delete(bs.Observers, obj)

	return nil
}

//Clear 清理订阅扫描的内容
func (bs *BlockScannerBase) Clear() error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	//bs.WalletInScanning = nil
	bs.AddressInScanning = nil
	bs.AddressInScanning = make(map[string]string)
	//bs.WalletInScanning = make(map[string]*WalletWrapper)
	return nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BlockScannerBase) SetRescanBlockHeight(height uint64) error {
	return nil
}

//SetTask
func (bs *BlockScannerBase) SetTask(task func()) {

	if bs.scanTask == nil {
		//创建定时器
		task := timer.NewTask(bs.PeriodOfTask, task)
		bs.scanTask = task
	}

}

//Run 运行
func (bs *BlockScannerBase) Run() error {

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

//Stop 停止扫描
func (bs *BlockScannerBase) Stop() error {
	bs.scanTask.Stop()
	bs.Scanning = false
	return nil
}

//Pause 暂停扫描
func (bs *BlockScannerBase) Pause() error {
	bs.scanTask.Pause()
	return nil
}

//Restart 继续扫描
func (bs *BlockScannerBase) Restart() error {
	bs.scanTask.Restart()
	return nil
}

//scanning 扫描
//func (bs *BlockScannerBase) ScanTask() {
//	//执行扫描任务
//}

//ScanBlock 扫描指定高度区块
func (bs *BlockScannerBase) ScanBlock(height uint64) error {
	//扫描指定高度区块
	return nil
}

//GetCurrentBlockHeight 获取当前区块高度
func (bs *BlockScannerBase) GetCurrentBlockHeader() (*BlockHeader, error) {
	return nil, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *BlockScannerBase) GetScannedBlockHeight() uint64 {
	return 0
}

func (bs *BlockScannerBase) ExtractTransactionData(txid string) (map[string]*TxExtractData, error) {
	return nil, nil
}

//GetBalanceByAddress 查询地址余额
func (bs *BlockScannerBase) GetBalanceByAddress(address ...string) ([]*Balance, error) {
	return nil, nil
}

//GetTokenBalanceByAddress 查询地址token余额列表
//func (bs *BlockScannerBase) GetTokenBalanceByAddress(address ...string) ([]*TokenBalance, error) {
//	return nil, nil
//}

//GetTransactionsByAddress 查询基于账户的交易记录，通过账户关系的地址
//返回的交易记录以资产账户为集合的结果，转账数量以基于账户来计算
func (bs *BlockScannerBase) GetTransactionsByAddress(offset, limit int, coin Coin, address ...string) ([]*TxExtractData, error) {
	return nil, nil
}
