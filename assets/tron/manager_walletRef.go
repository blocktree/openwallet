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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/bndr/gotabulate"
	"github.com/btcsuite/btcutil/hdkeychain"
)

//GetWallets 获取钱包列表
func (wm *WalletManager) GetWallets() ([]*openwallet.Wallet, error) {

	wallets, err := openwallet.GetWalletsByKeyDir(wm.Config.keyDir)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		w.DBFile = filepath.Join(wm.Config.dbPath, w.FileName()+".db")
	}

	return wallets, nil

}

//GetWalletInfo 获取钱包列表
func (wm *WalletManager) GetWalletInfo(walletID string) (*openwallet.Wallet, error) {

	wallets, err := wm.GetWallets()
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist")
}

//CreateNewWallet 创建钱包
func (wm *WalletManager) CreateNewWallet(name, password string) (*openwallet.Wallet, string, error) {

	var (
		err     error
		wallets []*openwallet.Wallet
	)

	//检查钱包名是否存在
	wallets, err = wm.GetWallets()
	for _, w := range wallets {
		if w.Alias == name {
			return nil, "", errors.New("The wallet's alias is duplicated")
		}
	}

	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil, "", err
	}

	extSeed, err := hdkeystore.GetExtendSeed(seed, wm.Config.MasterKey)
	if err != nil {
		return nil, "", err
	}

	key, keyFile, err := hdkeystore.StoreHDKeyWithSeed(wm.Config.keyDir, name, password, extSeed, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	if err != nil {
		return nil, "", err
	}

	file.MkdirAll(wm.Config.dbPath)
	file.MkdirAll(wm.Config.keyDir)

	w := &openwallet.Wallet{
		WalletID: key.KeyID,
		Alias:    key.Alias,
		KeyFile:  keyFile,
		DBFile:   filepath.Join(wm.Config.dbPath, key.FileName()+".db"),
	}

	w.SaveToDB()

	return w, keyFile, nil
}

// -------------------------------------------------- Functions ---------------------------------------------
//打印钱包列表
func (wm *WalletManager) printWalletList(list []*openwallet.Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		a := w.SingleAssetsAccount(wm.Config.Symbol)
		// a.Balance = wm.GetWalletBalance(a.AccountID)  ?500
		tableInfo = append(tableInfo, []interface{}{
			i, a.WalletID, a.Alias, a.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}
