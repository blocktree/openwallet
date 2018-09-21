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
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
)

// CreateAssetsAccount
func (wm *WalletManager) CreateAssetsAccount(appID, walletID, password string, account *openwallet.AssetsAccount, otherOwnerKeys []string) (*openwallet.AssetsAccount, error) {

	var (
		wallet *openwallet.Wallet
	)

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

	if account.IsTrust {

		wrapper, err := wm.newWalletWrapper(appID, walletID)
		if err != nil {
			return nil, err
		}

		wallet = wrapper.GetWallet()

		log.Debugf("wallet[%v] is trusted", wallet.WalletID)
		//使用私钥创建子账户
		key, err := wrapper.HDKey(password)
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

		account.PublicKey = childKey.GetPublicKey().OWEncode()
		account.Index = uint64(newAccIndex)
		account.AccountID = account.GetAccountID()

		wallet.AccountIndex = newAccIndex
	} else {

		//非托管的，创建资产账户的钱包
		wallet, _, err = wm.CreateWallet(appID, &openwallet.Wallet{
			Alias:    "imported",
			WalletID: walletID,
			IsTrust:  false,
		})
		if err != nil {
			return nil, err
		}

	}

	account.AddressIndex = -1

	//组合拥有者
	account.OwnerKeys = []string{
		account.PublicKey,
	}

	for _, otherKey := range otherOwnerKeys {
		if len(otherKey) > 0 {
			account.OwnerKeys = append(account.OwnerKeys, otherKey)
		}
	}

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

	wrapper, err := wm.newWalletWrapper(appID, "")
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

	wrapper, err := wm.newWalletWrapper(appID, "")
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
	if err != nil {
		return err
	}

	err = assetsMgr.GetAddressWithBalance(address...)
	if err != nil {
		return err
	}

	balanceDel := decimal.New(0, 0)

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return err
	}
	defer wrapper.CloseDB()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, address := range address {

		amount, _ := decimal.NewFromString(address.Balance)
		balanceDel = balanceDel.Add(amount)

		err = tx.Save(address)
		if err != nil {
			return err
		}
	}

	account.Balance = balanceDel.StringFixed(assetsMgr.Decimal())

	err = tx.Save(account)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAssetsAccountList
func (wm *WalletManager) GetAssetsAccountList(appID, walletID string, offset, limit int) ([]*openwallet.AssetsAccount, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
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

	wrapper, err := wm.newWalletWrapper(appID, "")
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

	addrs, err := wrapper.CreateAddress(accountID, count, assetsMgr.GetAddressDecode(), false, wm.cfg.IsTestnet)
	if err != nil {
		return nil, err
	}

	log.Debug("addrs:", addrs)
	//导入地址到核心钱包中
	err = assetsMgr.ImportWatchOnlyAddress(addrs...)
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

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	addrs, err := wrapper.GetAddressList(offset, limit, "AccountID", accountID, "WatchOnly", watchOnly)
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

	//wrapper, err := wm.newWalletWrapper(appID, "")
	//if err != nil {
	//	return err
	//}

	//导入观测地址
	//err = wrapper.ImportWatchOnlyAddress(addresses...)
	//if err != nil {
	//	return err
	//}

	assetsMgr, err := GetAssetsManager(account.Symbol)
	if err != nil {
		return err
	}

	//导入地址到核心钱包中
	assetsMgr.ImportWatchOnlyAddress(addresses...)

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
		a.Symbol = strings.ToUpper(account.Symbol)
		a.AccountID = account.AccountID
		a.CreatedAt = createdAt.Unix()
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
