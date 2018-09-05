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
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/timer"
	"sync"
	"time"
)

const (
	periodOfTask = 5 * time.Second //定时任务执行隔间
)

//BlockScannerWrapper 区块链扫描器包装
type BlockScannerWrapper struct {
	addressInScanning map[string]string                    //加入扫描的地址
	walletInScanning  map[string]*WalletWrapper            //加入扫描的钱包
	scanTask          *timer.TaskTimer                     //扫描定时器
	mu                sync.RWMutex                         //读写锁
	observers         map[BlockScanNotificationObject]bool //观察者
	scanning          bool                                 //是否扫描中
	PeriodOfTask      time.Duration
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	Recharges   []*Recharge
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
func NewBlockScannerWrapper() *BlockScannerWrapper {
	bs := BlockScannerWrapper{}
	bs.addressInScanning = make(map[string]string)
	bs.walletInScanning = make(map[string]*WalletWrapper)
	bs.observers = make(map[BlockScanNotificationObject]bool)
	bs.PeriodOfTask = periodOfTask
	return &bs
}

//AddAddress 添加订阅地址
func (bs *BlockScannerWrapper) AddAddress(address, sourceKey string, wrapper *WalletWrapper) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.addressInScanning[address] = sourceKey

	if _, exist := bs.walletInScanning[sourceKey]; exist {
		return
	}
	bs.walletInScanning[sourceKey] = wrapper
}

//AddWallet 添加扫描钱包
func (bs *BlockScannerWrapper) AddWallet(sourceKey string, wrapper *WalletWrapper) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if _, exist := bs.walletInScanning[sourceKey]; exist {
		//已存在，不重复订阅
		return
	}

	bs.walletInScanning[sourceKey] = wrapper

	//删除充值记录
	//wallet.DropRecharge()

	//导入钱包该账户的所有地址
	addrs, err := wrapper.GetAddressList(0, -1)
	if err != nil {
		return
	}

	log.Std.Info("block scanner load wallet [%s] existing addresses: %d ", sourceKey, len(addrs))

	for _, address := range addrs {
		bs.addressInScanning[address.Address] = sourceKey
	}

}

//IsExistAddress 指定地址是否已登记扫描
func (bs *BlockScannerWrapper) IsExistAddress(address string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.addressInScanning[address]
	return exist
}

//IsExistWallet 指定账户的钱包是否已登记扫描
func (bs *BlockScannerWrapper) IsExistWallet(accountID string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.walletInScanning[accountID]
	return exist
}

//AddObserver 添加观测者
func (bs *BlockScannerWrapper) AddObserver(obj BlockScanNotificationObject) {
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
func (bs *BlockScannerWrapper) RemoveObserver(obj BlockScanNotificationObject) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.observers, obj)
}

//Clear 清理订阅扫描的内容
func (bs *BlockScannerWrapper) Clear() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.walletInScanning = nil
	bs.addressInScanning = nil
	bs.addressInScanning = make(map[string]string)
	bs.walletInScanning = make(map[string]*WalletWrapper)
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BlockScannerWrapper) SetRescanBlockHeight(height uint64) error {
	return nil
}

//Run 运行
func (bs *BlockScannerWrapper) Run() {

	if bs.scanning {
		return
	}

	if bs.scanTask == nil {
		//创建定时器
		task := timer.NewTask(bs.PeriodOfTask, bs.ScanTask)
		bs.scanTask = task
	}
	bs.scanning = true
	bs.scanTask.Start()
}

//Stop 停止扫描
func (bs *BlockScannerWrapper) Stop() {
	bs.scanTask.Stop()
	bs.scanning = false
}

//Pause 暂停扫描
func (bs *BlockScannerWrapper) Pause() {
	bs.scanTask.Pause()
}

//Restart 继续扫描
func (bs *BlockScannerWrapper) Restart() {
	bs.scanTask.Restart()
}

//scanning 扫描
func (bs *BlockScannerWrapper) ScanTask() {
	//执行扫描任务
}

//ScanBlock 扫描指定高度区块
func (bs *BlockScannerWrapper) ScanBlock(height uint64) error {
	//扫描指定高度区块
	return nil
}

//GetWalletByAddress 获取地址对应的钱包
func (bs *BlockScannerWrapper) GetWalletWrapperByAddress(address string) (*WalletWrapper, bool) {
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
