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
	"github.com/blocktree/OpenWallet/openwallet"
)

//blockScanNotify 区块扫描结果通知
func (wm *WalletManager) BlockScanNotify(header *openwallet.BlockHeader) error {
	//推送数据
	for o, _ := range wm.observers {
		o.BlockScanNotify(header)
	}
	return nil
}

//BlockExtractDataNotify 区块提取结果通知
func (wm *WalletManager) BlockExtractDataNotify(sourceKey string, data *openwallet.BlockExtractData) error {
	//保存提取出来的数据

	wrapper, err := wm.newWalletWrapper(sourceKey)
	if err != nil {
		return err
	}

	txWrapper := openwallet.NewTransactionWrapper(wrapper)
	err = txWrapper.SaveBlockExtractData(data)
	if err != nil {
		return err
	}

	for o, _ := range wm.observers {
		o.BlockExtractDataNotify(sourceKey, data)
	}

	//TODO:定时删除过时的记录，保证数据库不会无限增加

	return nil
}
