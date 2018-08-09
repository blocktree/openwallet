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
	"github.com/shopspring/decimal"
	"strings"
	"log"
	"github.com/pkg/errors"
)

//CreateMerchantWallet 创建钱包
func (wm *WalletManager) CreateMerchantWallet(wallet *openwallet.Wallet) error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return errors.New("The wallet node is not config!")
	}

	coreWallet, keyFile, err := CreateNewWallet(wallet.Alias, wallet.Password)
	if err != nil {
		return err
	}

	//创建钱包资源文件夹
	file.MkdirAll(dbPath)

	//使用BTC的ID
	wallet.WalletID = coreWallet.WalletID
	wallet.DBFile = coreWallet.DBFile()
	wallet.KeyFile = keyFile

	return nil
}

func (wm *WalletManager) CreateMerchantAssetsAccount(wallet *openwallet.Wallet) (*openwallet.AssetsAccount, error) {

	return nil, nil
}

//GetMerchantWalletList 获取钱包列表
func (wm *WalletManager) GetMerchantWalletList() ([]*openwallet.Wallet, error) {

	return nil, nil
}

func (wm *WalletManager) GetMerchantAssetsAccountList(wallet *openwallet.Wallet) ([]*openwallet.AssetsAccount, error) {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config!")
	}

	balance := GetWalletBalance(wallet.WalletID)
	account := wallet.SingleAssetsAccount(Symbol)
	account.Balance = balance
	return []*openwallet.AssetsAccount{account}, nil
}

//ConfigMerchantWallet 钱包工具配置接口
func (wm *WalletManager) ConfigMerchantWallet(wallet *openwallet.Wallet) error {

	return nil
}

//ImportMerchantAddress 导入地址
func (wm *WalletManager) ImportMerchantAddress(wallet *openwallet.Wallet, account *openwallet.AssetsAccount, addresses []*openwallet.Address) error {

	//写入到数据库中
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range addresses {
		a.WatchOnly = true //观察地址
		a.Symbol = strings.ToLower(Symbol)
		a.AccountID = account.AccountID
		log.Printf("import %s address: %s", a.Symbol, a.Address)
		tx.Save(a)

		wm.blockscanner.AddAddress(a.Address, wallet.WalletID, wallet)
	}

	tx.Commit()

	return nil
}

//CreateMerchantAddress 创建钱包地址
func (wm *WalletManager) CreateMerchantAddress(wallet *openwallet.Wallet, account *openwallet.AssetsAccount, count uint64) ([]*openwallet.Address, error) {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config!")
	}
	log.Printf("wallet: %s create %d address...\n", wallet.WalletID, count)
	_, newAddrs, err := CreateBatchAddress(wallet.WalletID, wallet.Password, count)
	if err != nil {
		return nil, err
	}

	return newAddrs, nil
}

//GetMerchantAddressList 获取钱包地址
func (wm *WalletManager) GetMerchantAddressList(wallet *openwallet.Wallet, account *openwallet.AssetsAccount, offset uint64, limit uint64) ([]*openwallet.Address, error) {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config!")
	}
	return GetAddressesFromLocalDB(wallet.WalletID, int(offset), int(limit))
	//return nil, nil
}

//SubmitTransaction 提交转账申请
func (wm *WalletManager) SubmitTransactions(wallet *openwallet.Wallet, account *openwallet.AssetsAccount, withdraws []*openwallet.Withdraw, surplus string) (*openwallet.Transaction, error) {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config!")
	}

	//coreWallet := readWallet(wallet.KeyFile)
	addrs := make([]string, 0)
	amounts := make([]decimal.Decimal, 0)
	totalSend := decimal.New(0, 0)
	for _, with := range withdraws {

		if len(with.Address) == 0 {
			continue
		}

		amount, _ := decimal.NewFromString(with.Amount)
		addrs = append(addrs, with.Address)
		amounts = append(amounts, amount)

	}

	balance := GetWalletBalance(wallet.WalletID)
	b, _ := decimal.NewFromString(balance)
	sp, _ := decimal.NewFromString(surplus)
	//【余额—剩余额】低于提币数量，不允许提
	if b.Sub(sp).LessThan(totalSend) {
		return nil, errors.New("the wallet balance is lower than surplus. ")
	}

	//重新加载utxo
	//RebuildWalletUnspent(wallet.WalletID)

	txID, err := SendBatchTransaction(wallet.WalletID, addrs, amounts, wallet.Password)
	if err != nil {
		return nil, err
	}
	t := openwallet.Transaction{TxID: txID}

	return &t, nil
}

//AddMerchantObserverForBlockScan 添加区块链观察者，当扫描出新区块时进行通知
func (wm *WalletManager) AddMerchantObserverForBlockScan(obj openwallet.BlockScanNotificationObject, wallet *openwallet.Wallet) error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return errors.New("The wallet node is not config!")
	}

	wm.blockscanner.AddObserver(obj)
	wm.blockscanner.AddWallet(wallet.WalletID, wallet)

	wm.blockscanner.Run()
	return nil
}

//RemoveMerchantObserverForBlockScan 移除区块链扫描的观测者
func (wm *WalletManager) RemoveMerchantObserverForBlockScan(obj openwallet.BlockScanNotificationObject) {
	wm.blockscanner.RemoveObserver(obj)
	if len(wm.blockscanner.observers) == 0 {
		wm.blockscanner.Stop()
		wm.blockscanner.Clear()
	}
}

//GetBlockchainInfo 获取区块链信息
func (wm *WalletManager) GetBlockchainInfo() (*openwallet.Blockchain, error) {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config! ")
	}

	height, err := GetBlockHeight()
	if err != nil {
		return nil, err
	}

	localHeight, _ := GetLocalNewBlock()
	if err != nil {
		return nil, err
	}

	info := openwallet.Blockchain{
		Blocks: height,
		ScanHeight: localHeight,
	}

	return &info, nil
}

//GetMerchantAssetsAccount 获取账户资产
func (wm *WalletManager) GetMerchantWalletBalance(walletID string) (string, error) {
	return "0", nil
}

//GetMerchantAssetsAccount 获取地址资产
func (wm *WalletManager) GetMerchantAddressBalance(walletID, address string) (string, error) {
	return "0", nil
}

//SetMerchantRescanBlockHeight 商户重置区块链扫描高度
func (wm *WalletManager) SetMerchantRescanBlockHeight(height uint64) error {
	return wm.blockscanner.SetRescanBlockHeight(height)
}