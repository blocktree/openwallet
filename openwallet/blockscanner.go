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

type BlockScanner interface {

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

	//SetRescanBlockHeight 重置区块链扫描高度
	SetRescanBlockHeight(height uint64) error

	//Run 运行
	Run()

	//Stop 停止扫描
	Stop()

	//Pause 暂停扫描
	Pause()

	//Restart 继续扫描
	Restart()

	//ScanBlock 扫描指定高度的区块
	ScanBlock(height uint64) error

	//rescanFailedRecord 重扫失败记录
	RescanFailedRecord()

	//GetCurrentBlockHeight 获取当前区块高度
	GetCurrentBlockHeader() (*BlockHeader, error)

	//IsExistAddress 指定地址是否已登记扫描
	IsExistAddress(address string) bool

	//IsExistWallet 指定账户的钱包是否已登记扫描
	IsExistWallet(accountID string) bool
}

//BlockScanNotificationObject 扫描被通知对象
type BlockScanNotificationObject interface {

	//BlockScanNotify 新区块扫描完成通知
	BlockScanNotify(header *BlockHeader)
}

type ExtractTxResult interface {
	//是否成功
	Success() bool
	//获取提出到的交易记录
	GetRecharges() []*Recharge
}
