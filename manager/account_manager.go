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
	"fmt"
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"strings"
	"time"
)

// CreateAssetsAccount
func (wm *WalletManager) CreateAssetsAccount(appID, walletID string, account *openwallet.AssetsAccount, otherOwnerKeys []string) (*openwallet.AssetsAccount, error) {

	wallet, err := wm.GetWalletInfo(appID, walletID)
	if err != nil {
		return nil, err
	}

	if len(account.Alias) == 0 {
		return nil, fmt.Errorf("account alias is empty")
	}

	if len(account.Symbol) == 0 {
		return nil, fmt.Errorf("account symbol is empty")
	}

	if account.Required == 0 {
		account.Required = 1
	}

	symbolInfo, err := assets.GetSymbolInfo(account.Symbol)
	if err != nil {
		return nil, err
	}

	if wallet.IsTrust {

		//使用私钥创建子账户
		key, err := wallet.HDKey()
		if err != nil {
			return nil, err
		}

		newAccIndex := wallet.AccountIndex + 1

		// root/n' , 使用强化方案
		account.HDPath = fmt.Sprintf("%s/%d'", wallet.RootPath, newAccIndex)

		childKey, err := key.DerivedKeyWithPath(account.HDPath, symbolInfo.CurveType())
		if err != nil {
			return nil, err
		}

		account.PublicKey = childKey.OWEncode()
		account.Index = uint64(newAccIndex)
		account.AccountID = account.GetAccountID()

		wallet.AccountIndex = newAccIndex
	}

	account.AddressIndex = -1

	//组合拥有者
	account.OwnerKeys = []string{
		account.PublicKey,
	}

	account.OwnerKeys = append(account.OwnerKeys, otherOwnerKeys...)

	if len(account.PublicKey) == 0 {
		return nil, fmt.Errorf("account publicKey is empty")
	}

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = tx.Save(wallet)
	if err != nil {

		return nil, err
	}

	err = tx.Save(account)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.Debug("new account create success:", account.AccountID)

	_, err = wm.CreateAddress(appID, walletID, account.GetAccountID(), 1)
	if err != nil {
		log.Debug("new address create failed, unexpected error:", err)
	}

	return account, nil
}

// GetAssetsAccountInfo
func (wm *WalletManager) GetAssetsAccountInfo(appID, walletID, accountID string) (*openwallet.AssetsAccount, error) {

	wrapper, err := wm.newWalletWrapper(appID)
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	return account, nil
}

//RefreshAssetsAccountBalance 刷新资产账户余额
func (wm *WalletManager) RefreshAssetsAccountBalance(appID, accountID string) error {

	wrapper, err := wm.newWalletWrapper(appID)
	if err != nil {
		return err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return err
	}

	address, err := wrapper.GetAddressList(0, -1, "AccountID", accountID)

	searchAddrs := make([]string, 0)
	for _, address := range address {
		searchAddrs = append(searchAddrs, address.Address)
	}

	balance, err := assetsMgr.CountBalanceByAddresses(searchAddrs...)
	if err != nil {
		return err
	}

	account.Balance = balance

	return wrapper.SaveAssetsAccount(account)
}

// GetAssetsAccountList
func (wm *WalletManager) GetAssetsAccountList(appID, walletID string, offset, limit int) ([]*openwallet.AssetsAccount, error) {

	wrapper, err := wm.newWalletWrapper(appID)
	if err != nil {
		return nil, err
	}

	accounts, err := wrapper.GetAssetsAccountList(offset, limit, "WalletID", walletID)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// CreateAddress
func (wm *WalletManager) CreateAddress(appID, walletID string, accountID string, count uint64) ([]*openwallet.Address, error) {

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	wrapper, err := wm.newWalletWrapper(appID)
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return nil, err
	}

	addrs, err := wrapper.CreateAddress(accountID, count, assetsMgr.GetAddressDecode(), false, wm.cfg.isTestnet)
	if err != nil {
		return nil, err
	}

	//导入新地址到区块扫描器
	scanner := assetsMgr.GetBlockScanner()

	if scanner == nil {
		log.Warn(account.Symbol, "is not support block scan")
	} else {
		for _, address := range addrs {
			scanner.AddAddress(address.Address, appID)
		}
	}

	log.Debug("new addresses create success:", addrs)

	return addrs, nil
}

// GetAddressList
func (wm *WalletManager) GetAddressList(appID, walletID, accountID string, offset, limit int, watchOnly bool) ([]*openwallet.Address, error) {

	wrapper, err := wm.newWalletWrapper(appID)
	if err != nil {
		return nil, err
	}

	addrs, err := wrapper.GetAddressList(offset, limit, "AccountID", accountID, "WatchOnly", false)
	if err != nil {
		return nil, err
	}

	//var addrs []*openwallet.Address
	//err = db.Find("AccountID", accountID, &addrs, storm.Limit(limit), storm.Skip(offset))
	//if err != nil {
	//	return nil, err
	//}

	return addrs, nil
}

// ImportWatchOnlyAddress
func (wm *WalletManager) ImportWatchOnlyAddress(appID, walletID, accountID string, addresses []*openwallet.Address) error {

	account, err := wm.GetAssetsAccountInfo(appID, walletID, accountID)
	if err != nil {
		return err
	}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return err
	}

	//导入新地址到区块扫描器
	scanner := assetsMgr.GetBlockScanner()

	createdAt := time.Now()

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, a := range addresses {
		a.WatchOnly = true //观察地址
		a.Symbol = strings.ToLower(account.Symbol)
		a.AccountID = account.AccountID
		a.CreatedAt = createdAt
		err = tx.Save(a)
		if err != nil {
			return err
		}

		if scanner == nil {
			log.Warn(account.Symbol, "is not support block scan")
		} else {
			//导入新地址到区块扫描器
			scanner := assetsMgr.GetBlockScanner()
			scanner.AddAddress(a.Address, appID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Debug("import addresses success count:", len(addresses))

	return nil
}
