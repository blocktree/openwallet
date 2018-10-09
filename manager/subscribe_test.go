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

package manager

import (
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"testing"
)

type subscriber struct {
}

func init() {
	//tm.Init()
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriber) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Debug("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriber) BlockTxExtractDataNotify(account *openwallet.AssetsAccount, data *openwallet.TxExtractData) error {
	log.Debug("account:", account)
	log.Debug("data:", data)
	return nil
}

func TestSubscribe(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	sub := subscriber{}
	tm.AddObserver(&sub)
	//tm.SetRescanBlockHeight("QTUM", 236098)
	log.Debug("SupportAssets:", tm.cfg.SupportAssets)
	<-endRunning
}

