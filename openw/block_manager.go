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

package openw

import (
	"fmt"
)

//AddAddressForBlockScan 添加订阅地址
func (wm *WalletManager) AddAddressForBlockScan(address, sourceKey string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.AddressInScanning[address] = sourceKey
	return nil
//if _, exist := wm.WalletInScanning[sourceKey]; exist {
//	return
//}
//wm.WalletInScanning[sourceKey] = wrapper
}

//IsExistAddressForBlockScan 指定地址是否已登记扫描
func (wm *WalletManager) IsExistAddressForBlockScan(address string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	_, exist := wm.AddressInScanning[address]
	return exist
}

//ClearAddressForBlockScan 清理订阅扫描的内容
func (wm *WalletManager) ClearAddressForBlockScan() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	//wm.WalletInScanning = nil
	wm.AddressInScanning = nil
	wm.AddressInScanning = make(map[string]string)
	//wm.WalletInScanning = make(map[string]*WalletWrapper)
	return nil
}

//GetSourceKeyByAddressForBlockScan 获取地址对应的数据源标识
func (wm *WalletManager) GetSourceKeyByAddressForBlockScan(address string) (string, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	sourceKey, ok := wm.AddressInScanning[address]
	return sourceKey, ok
}

func (wm *WalletManager) RescanBlockHeight(symbol string, startHeight uint64, endHeight uint64) error {

	assetsMgr, err := GetAssetsAdapter(symbol)
	if err != nil {
		return err
	}

	scanner := assetsMgr.GetBlockScanner()

	if scanner == nil {
		return fmt.Errorf("%s is not support block scan", symbol)
	}

	if startHeight <= endHeight {
		for i := startHeight; i <= endHeight; i++ {
			err := scanner.ScanBlock(i)
			if err != nil {
				continue
			}
		}
	} else {
		return fmt.Errorf("start block height: %d is greater than end block height: %d", startHeight, endHeight)
	}

	return nil
}

//SetRescanBlockHeight 重置区块高度起扫描
func (wm *WalletManager) SetRescanBlockHeight(symbol string, height uint64) error {

	assetsMgr, err := GetAssetsAdapter(symbol)
	if err != nil {
		return err
	}

	scanner := assetsMgr.GetBlockScanner()

	if scanner == nil {
		return fmt.Errorf("%s is not support block scan", symbol)
	}

	err = scanner.SetRescanBlockHeight(height)
	if err != nil {
		return err
	}

	return nil
}

//GetNewBlockHeight 获取区块高度，（最新高度，已扫描高度）
func (wm *WalletManager) GetNewBlockHeight(symbol string) (uint64, uint64, error) {

	assetsMgr, err := GetAssetsAdapter(symbol)
	if err != nil {
		return 0, 0, err
	}

	scanner := assetsMgr.GetBlockScanner()

	if scanner == nil {
		return 0, 0, fmt.Errorf("%s is not support block scan", symbol)
	}

	header, err := scanner.GetCurrentBlockHeader()
	if err != nil {
		return 0, 0, err
	}

	scannedHeight := scanner.GetScannedBlockHeight()

	return header.Height, scannedHeight, nil
}
