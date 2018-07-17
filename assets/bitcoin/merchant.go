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

package bitcoin

import (
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/openwallet"
	"path/filepath"
)

//CreateMerchantWallet 创建钱包
func (w *WalletManager) CreateMerchantWallet(alias string, password string) (*openwallet.Wallet, error) {

	wallet, keyFile, err := CreateNewWallet(alias, password)
	if err != nil {
		return nil, err
	}

	//创建钱包资源文件夹
	walletDataFolder := filepath.Join(dbPath, wallet.DBFile())
	file.MkdirAll(walletDataFolder)

	owWallet := openwallet.Wallet{
		Alias:   alias,
		RootPub: wallet.RootPub,
		DBFile:  walletDataFolder,
		KeyFile: keyFile,
	}
	return &owWallet, nil
}

//GetMerchantWalletList 获取钱包列表
func (w *WalletManager) GetMerchantWalletList() ([]*openwallet.Wallet, error) {

	return nil, nil
}

//ConfigMerchantWallet 钱包工具配置接口
func (w *WalletManager) ConfigMerchantWallet(wallet *openwallet.Wallet) error {

	return nil
}

//ImportMerchantAddress 导入地址
func (w *WalletManager) ImportMerchantAddress(addresses []*openwallet.Address) error {

	//写入到数据库中

	return nil
}

//CreateMerchantAddress 创建钱包地址
func (w *WalletManager) CreateMerchantAddress(walletID string, count int) ([]*openwallet.Address, error) {
	return nil, nil
}

//GetMerchantAddressList 获取钱包地址
func (w *WalletManager) GetMerchantAddressList(walletID string, offset uint64, limit uint64) ([]*openwallet.Address, error) {
	return nil, nil
}
