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

package eosio

import (
	"github.com/blocktree/openwallet/openwallet"
)

const (
	blockchainBucket = "blockchain" // blockchain dataset
	//periodOfTask      = 5 * time.Second // task interval
	maxExtractingSize = 10 // thread count
)

//EOSBlockScanner EOS block scanner
type EOSBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

//ExtractResult extract result
type ExtractResult struct {
	extractData map[string]*openwallet.TxExtractData
	TxID        string
	BlockHeight uint64
	Success     bool
}

//SaveResult result
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// NewEOSBlockScanner create a block scanner
func NewEOSBlockScanner(wm *WalletManager) *EOSBlockScanner {
	bs := EOSBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0

	// set task
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

// GetCurrentBlockHeader current local block header
// func (bs *EOSBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {
// 	var (
// 		blockHeight uint64 = 0
// 		hash        string
// 		err         error
// 	)

// 	blockHeight, hash = bs.GetLocalNewBlock()
// }

// ScanBlockTask scan block task
func (bs *EOSBlockScanner) ScanBlockTask() {

	var (
		currentHeight uint32
		currentHash   string
	)

	// get local block header
	currentHeight, currentHash = bs.GetLocalNewBlock()
	
	if currentHeight == 0 {
		bs.wm.Log.Std.Info("No records found in local, get current block as the local!")

		// get head block
		infoResp, err := bs.wm.Api.GetInfo()
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get info;unexpected error:%v", err)
			return
		}

		currentHash = infoResp.HeadBlockID.String()
		currentHeight = infoResp.HeadBlockNum
	}

	for {
		if !bs.Scanning {
			// stop scan
			return
		}

		infoResp, err := bs.wm.Api.GetInfo()
		if err != nil {
			bs.wm.Log.Errorf("get max height of eth failed, err=%v", err)
			break
		}

		maxBlockHeight := infoResp.HeadBlockNum

		bs.wm.Log.Info("current block height:", currentHeight, currentHash, " maxBlockHeight:", maxBlockHeight)
		// if currentHeight == maxBlockHeight {
		// 	bs.wm.Log.Infof("block scanner has done with scan. current height:%v", maxBlockHeight)
		// 	break
		// }

	}
}
