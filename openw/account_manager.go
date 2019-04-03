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
	"strings"
	"time"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

// CreateAssetsAccount
func (wm *WalletManager) CreateAssetsAccount(appID, walletID, password string, account *openwallet.AssetsAccount, otherOwnerKeys []string) (*openwallet.AssetsAccount, *openwallet.Address, error) {

	var (
		wallet *openwallet.Wallet
	)

	if len(account.Alias) == 0 {
		return nil, nil, fmt.Errorf("account alias is empty")
	}

	if len(account.Symbol) == 0 {
		return nil, nil, fmt.Errorf("account symbol is empty")
	}

	if account.Required == 0 {
		account.Required = 1
	}

	symbolInfo, err := GetSymbolInfo(account.Symbol)
	if err != nil {
		return nil, nil, err
	}

	wrapper, err := wm.newWalletWrapper(appID, walletID)
	if err == nil {
		wallet = wrapper.GetWallet()
	}

	if account.IsTrust {

		if wallet == nil {
			return nil, nil, fmt.Errorf("wallet not exist")
		}

		log.Debugf("wallet[%v] is trusted", wallet.WalletID)
		//使用私钥创建子账户
		key, err := wrapper.HDKey(password)
		if err != nil {
			return nil, nil, err
		}

		newAccIndex := wallet.AccountIndex + 1

		// root/n' , 使用强化方案
		account.HDPath = fmt.Sprintf("%s/%d'", wallet.RootPath, newAccIndex)

		childKey, err := key.DerivedKeyWithPath(account.HDPath, symbolInfo.CurveType())
		if err != nil {
			return nil, nil, err
		}
		account.PublicKey = childKey.GetPublicKey().OWEncode()
		account.Index = uint64(newAccIndex)
		account.AccountID = account.GetAccountID()

		wallet.AccountIndex = newAccIndex
	} else {

		if wallet == nil {

			//非托管的，创建资产账户的钱包
			wallet, _, err = wm.CreateWallet(appID, &openwallet.Wallet{
				Alias:    "imported",
				WalletID: walletID,
				IsTrust:  false,
			})
			if err != nil {
				return nil, nil, err
			}
		}

		wallet.AccountIndex = int(account.Index)
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
		return nil, nil, fmt.Errorf("account publicKey is empty")
	}

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, nil, err
	}

	defer tx.Rollback()

	err = tx.Save(wallet)
	if err != nil {

		return nil, nil, err
	}

	err = tx.Save(account)
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	log.Debug("new account create success:", account.AccountID)

	addresses, err := wm.CreateAddress(appID, walletID, account.GetAccountID(), 1)
	if err != nil {
		log.Debug("new address create failed, unexpected error:", err)
	}

	var addr *openwallet.Address
	if len(addresses) > 0 {
		addr = addresses[0]
		account.AddressIndex++
	}

	return account, addr, nil
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
//func (wm *WalletManager) RefreshAssetsAccountBalance(appID, accountID string) error {
//
//	wrapper, err := wm.newWalletWrapper(appID, "")
//	if err != nil {
//		return err
//	}
//
//	account, err := wrapper.GetAssetsAccountInfo(accountID)
//	if err != nil {
//		return err
//	}
//
//	assetsMgr, err := GetAssetsAdapter(account.Symbol)
//	if err != nil {
//		return err
//	}
//
//	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID)
//	if err != nil {
//		return err
//	}
//
//	err = assetsMgr.GetAddressWithBalance(addresses...)
//	if err != nil {
//		return err
//	}
//
//	balanceDel := decimal.New(0, 0)
//
//	//打开数据库
//	db, err := wrapper.OpenStormDB()
//	if err != nil {
//		return err
//	}
//	defer wrapper.CloseDB()
//
//	tx, err := db.Begin(true)
//	if err != nil {
//		return err
//	}
//
//	defer tx.Rollback()
//
//	//批量插入到本地数据库
//	//设置utxo的钱包账户
//	for _, address := range addresses {
//
//		amount, _ := decimal.NewFromString(address.Balance)
//		balanceDel = balanceDel.Add(amount)
//		//log.Debug("address:", address.Address, "amount:", amount)
//		address.Balance = amount.StringFixed(assetsMgr.Decimal())
//		err = tx.Save(address)
//		if err != nil {
//			return err
//		}
//	}
//
//	account.Balance = balanceDel.StringFixed(assetsMgr.Decimal())
//
//	err = tx.Save(account)
//	if err != nil {
//		return err
//	}
//
//	return tx.Commit()
//}

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

	assetsMgr, err := GetAssetsAdapter(account.Symbol)
	if err != nil {
		return nil, err
	}

	addrs, err := wrapper.CreateAddress(accountID, count, assetsMgr.GetAddressDecode(), false, false)
	if err != nil {
		return nil, err
	}

	log.Debug("addrs:", addrs)
	//导入地址到核心钱包中
	//go func() {
	//
	//	inErr := assetsMgr.ImportWatchOnlyAddress(addrs...)
	//	if err != nil {
	//		log.Error("import watch only address failed， unexpected err:", inErr)
	//	}
	//}()

	//导入新地址到区块扫描器
	for _, address := range addrs {
		key := wm.encodeSourceKey(appID, address.AccountID)
		wm.AddAddressForBlockScan(address.Address, key)
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

// GetAddress 获取单个地址信息
func (wm *WalletManager) GetAddress(appID, walletID, accountID, address string) (*openwallet.Address, error) {

	wrapper, err := wm.newWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	addr, err := wrapper.GetAddress(address)
	if err != nil {
		return nil, err
	}

	//var addrs []*openwallet.Address
	//err = db.Find("AccountID", accountID, &addrs, storm.Limit(limit), storm.Skip(offset))
	//if err != nil {
	//	return nil, err
	//}

	return addr, nil
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

	//assetsMgr, err := GetAssetsAdapter(account.Symbol)
	//if err != nil {
	//	return err
	//}

	//导入地址到核心钱包中
	//go assetsMgr.ImportWatchOnlyAddress(addresses...)

	//导入新地址到区块扫描器
	//scanner := assetsMgr.GetBlockScanner()

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

	//importAddrs := make([]openwallet.ImportAddress, 0)

	for _, a := range addresses {
		a.WatchOnly = true //观察地址
		a.Symbol = strings.ToUpper(account.Symbol)
		a.AccountID = account.AccountID
		a.CreatedTime = createdAt.Unix()
		err = tx.Save(a)
		if err != nil {
			return err
		}

		key := wm.encodeSourceKey(appID, a.AccountID)
		wm.AddAddressForBlockScan(a.Address, key)

		//记录要导入到核心钱包的地址
		imported := openwallet.ImportAddress{
			Address: *a,
		}

		err = tx.Save(&imported)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Debug("import addresses success count:", len(addresses))

	return nil
}

// importNewAddressToCoreWallet 导入新地址到核心钱包
//func (wm *WalletManager) importNewAddressToCoreWallet() {
//
//	//log.Notice("import new address to core wallet start")
//
//	var (
//		importAddressMap map[string][]*openwallet.Address
//		limit            = 50
//	)
//
//	//加载已存在所有app
//	appIDs, err := wm.loadAllAppIDs()
//	if err != nil {
//		return
//	}
//
//	//处理所有App待导入地址，导入到核心钱包
//LoopApp:
//	for _, appID := range appIDs {
//
//		importAddressMap = make(map[string][]*openwallet.Address)
//
//		wrapper, err := wm.newWalletWrapper(appID, "")
//		if err != nil {
//			continue LoopApp
//		}
//
//		//获取应用所有待导入的地址
//		addresses, err := wrapper.GetImportAddressList(0, limit)
//		if err != nil {
//			continue LoopApp
//		}
//
//		for _, a := range addresses {
//			imported, ok := importAddressMap[a.Symbol]
//			if !ok {
//				imported = make([]*openwallet.Address, 0)
//			}
//
//			imported = append(imported, &a.Address)
//
//			importAddressMap[a.Symbol] = imported
//		}
//
//		db, err := wm.OpenDB(appID)
//		if err != nil {
//			continue LoopApp
//		}
//
//		tx, err := db.Begin(true)
//		if err != nil {
//			continue LoopApp
//		}
//
//		defer tx.Rollback()
//
//		for symbol, importeds := range importAddressMap {
//
//			assetsMgr, err := GetAssetsAdapter(symbol)
//			if err != nil {
//				continue LoopApp
//			}
//
//			//导入地址到核心钱包中
//			err = assetsMgr.ImportWatchOnlyAddress(importeds...)
//			if err != nil {
//				log.Error("ImportWatchOnlyAddress failed unexpected error:", err)
//				continue LoopApp
//			}
//
//		}
//
//		//导入地址成功删除记录
//		for _, a := range addresses {
//			err = tx.DeleteStruct(a)
//			if err != nil {
//				log.Error("delete import address failed, unexpected error:", err)
//				continue
//			}
//		}
//
//		err = tx.Commit()
//		if err != nil {
//			continue
//		}
//
//		log.Debug("import", "App:", appID, "addresses to core wallet success count:", len(addresses))
//
//	}
//
//	//log.Notice("import new address to core wallet end")
//
//	return
//}
