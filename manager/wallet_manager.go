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
	"github.com/blocktree/OpenWallet/hdkeystore"
	"fmt"
	"github.com/asdine/storm"
)

func (wm *WalletManager) CreateWallet(appID string, wallet *openwallet.Wallet) (*openwallet.Wallet, *hdkeystore.HDKey, error) {

	var (
		key *hdkeystore.HDKey
	)
	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, nil, err
	}

	//托管密钥
	wallet.AppID = appID
	if wallet.IsTrust {

		if len(wallet.Password) == 0 {
			return nil, nil, fmt.Errorf("password is empty")
		}

		//生成keystore
		key, filePath, err := hdkeystore.StoreHDKey(wm.cfg.keyDir, wallet.Alias, wallet.Password, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
		if err != nil {
			return nil, nil, err
		}

		wallet.KeyFile = filePath
		wallet.WalletID = key.KeyID
		wallet.RootPath = key.RootPath
	}

	if len(wallet.WalletID) == 0 {
		return nil, nil, fmt.Errorf("walletID is empty")
	}

	//数据路径
	wallet.DBFile = db.FileName

	//保存钱包到本地应用数据库
	err = db.Save(wallet)
	if err != nil {
		return nil, nil, err
	}

	return wallet, key, nil
}

func (wm *WalletManager) GetWalletInfo(appID string, walletID string) (*openwallet.Wallet, error) {

	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	var wallet openwallet.Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, fmt.Errorf("can not find wallet: %s, unexpected error: %v", walletID, err)
	}

	return &wallet, nil
}


func (wm *WalletManager) GetWalletList(appID string, offset, limit int) ([]*openwallet.Wallet, error) {

	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	var wallets []*openwallet.Wallet
	err = db.All(&wallets, storm.Limit(limit), storm.Skip(offset))
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

