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

package tron

import (
	// "encoding/base64"
	// "errors"
	// "path/filepath"

	"sync"
	"time"

	// "github.com/asdine/storm"
	// "github.com/blocktree/OpenWallet/crypto"

	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/timer"
	// "github.com/tidwall/gjson"
)

const (
	blockchainBucket  = "blockchain"    //区块链数据集合
	periodOfTask      = 5 * time.Second //定时任务执行隔间
	maxExtractingSize = 20              //并发的扫描线程数
)

//TronBlockScanner tron的区块链扫描器
type TronBlockScanner struct {
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

//NewTronBlockScanner 创建区块链扫描器
func NewTronBlockScanner(wm *WalletManager) *TronBlockScanner {
	bs := TronBlockScanner{}
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
