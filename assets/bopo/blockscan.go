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

package bopo

import (
	// "encoding/base64"
	// "errors"
	// "path/filepath"

	"sync"
	"time"

	// "github.com/asdine/storm"
	// "github.com/blocktree/openwallet/crypto"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/openwallet/timer"
	"github.com/pkg/errors"
	// "github.com/tidwall/gjson"
)

const (
	blockchainBucket  = "blockchain"    //区块链数据集合
	periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20              //并发的扫描线程数
)

//FabricBlockScanner bitcoin的区块链扫描器
type FabricBlockScanner struct {
	walletInScanning  map[string]*openwallet.Wallet                   //加入扫描的钱包
	addressInScanning map[string]string                               //加入扫描的地址
	observers         map[openwallet.BlockScanNotificationObject]bool //观察者

	scanning           bool             //是否扫描中
	CurrentBlockHeight uint64           //当前区块高度
	extractingCH       chan struct{}    //扫描工作令牌
	scanTask           *timer.TaskTimer //扫描定时器
	mu                 sync.RWMutex     //读写锁
	wm                 *WalletManager   //钱包管理者
}

//NewFabricBlockScanner 创建区块链扫描器
func NewFabricBlockScanner(wm *WalletManager) *FabricBlockScanner {
	bs := FabricBlockScanner{}
	bs.walletInScanning = make(map[string]*openwallet.Wallet)
	bs.addressInScanning = make(map[string]string)
	bs.observers = make(map[openwallet.BlockScanNotificationObject]bool)
	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	return &bs
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

/*
	Implements to Interface "openwallet.BlockScanner"

	Prepare for Scanner:
		//AddAddress 添加扫描地址，账户ID，其钱包指针
		AddAddress(address, accountID string, wallet *Wallet)

		//AddWallet 添加扫描账户及其钱包指针
		AddWallet(accountID string, wallet *Wallet)

		//AddObserver 添加观测者
		AddObserver(obj BlockScanNotificationObject)

		//RemoveObserver 移除观测者
		RemoveObserver(obj BlockScanNotificationObject)

		//Clear 清理订阅扫描的内容
		Clear()
*/
//AddAddress 增加预期扫描的地址
func (bs *FabricBlockScanner) AddAddress(address, accountID string, wallet *openwallet.Wallet) {

	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.addressInScanning[address] = accountID

	if _, exist := bs.walletInScanning[accountID]; exist {
		return
	}

	bs.walletInScanning[accountID] = wallet
}

//AddWallet 添加扫描钱包
func (bs *FabricBlockScanner) AddWallet(accountID string, wallet *openwallet.Wallet) {

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
		log.Std.Debug("Not found any address by account!")
		return
	}

	log.Std.Info("block scanner load wallet [%s] existing addresses: %d ", accountID, len(addrs))

	for _, address := range addrs {
		bs.addressInScanning[address.Address] = accountID
	}

}

//AddObserver 添加观测者
func (bs *FabricBlockScanner) AddObserver(obj openwallet.BlockScanNotificationObject) {

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
func (bs *FabricBlockScanner) RemoveObserver(obj openwallet.BlockScanNotificationObject) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.observers, obj)
}

//Clear 清理订阅扫描的内容
func (bs *FabricBlockScanner) Clear() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.addressInScanning = nil
	bs.addressInScanning = make(map[string]string)

	bs.walletInScanning = nil
	bs.walletInScanning = make(map[string]*openwallet.Wallet)
}

/*
	Implements to Interface "openwallet.BlockScanner"

	Actions of Scanner's operations:
		Run()		//Run 运行
		Stop()		//Stop 停止扫描
		Pause()		//Pause 暂停扫描
		Restart()	//Restart 继续扫描

		//ScanBlock 扫描指定高度的区块
		ScanBlock(height uint64) error

		//rescanFailedRecord 重扫失败记录
		RescanFailedRecord()
*/
//Run 运行
func (bs *FabricBlockScanner) Run() {

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
func (bs *FabricBlockScanner) Stop() {
	bs.scanTask.Stop()
	bs.scanning = false
}

//Pause 暂停扫描
func (bs *FabricBlockScanner) Pause() {
	bs.scanTask.Pause()
}

//Restart 继续扫描
func (bs *FabricBlockScanner) Restart() {
	bs.scanTask.Restart()
}

/*
	Implements to Interface "openwallet.BlockScanner"

	Extra methods to make more controls:
		//SetRescanBlockHeight 重置区块链扫描高度
		SetRescanBlockHeight(height uint64) error

		//GetCurrentBlockHeight 获取当前区块高度
		GetCurrentBlockHeader() (*BlockHeader, error)

		//IsExistAddress 指定地址是否已登记扫描
		IsExistAddress(address string) bool

		//IsExistWallet 指定账户的钱包是否已登记扫描
		IsExistWallet(accountID string) bool
*/
//SetRescanBlockHeight 重置区块链扫描高度
func (bs *FabricBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return errors.New("block height to rescan must greater than 0.")
	}

	// hash, err := bs.wm.GetBlockHash(height)
	// if err != nil {
	// 	return err
	// }
	hash := "" // Fabric can get Block by Height without Hash

	bs.SaveLocalNewBlock(height, hash)

	return nil
}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *FabricBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash = bs.GetLocalNewBlock()

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

//IsExistAddress 指定地址是否已登记扫描
func (bs *FabricBlockScanner) IsExistAddress(address string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.addressInScanning[address]
	return exist
}

//IsExistWallet 指定账户的钱包是否已登记扫描
func (bs *FabricBlockScanner) IsExistWallet(accountID string) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	_, exist := bs.walletInScanning[accountID]
	return exist
}
