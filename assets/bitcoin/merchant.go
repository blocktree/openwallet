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
func (w *WalletManager) CreateMerchantWallet(wallet *openwallet.Wallet) error {

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

	wallet.DBFile = coreWallet.DBFile()
	wallet.KeyFile = keyFile

	return nil
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
func (w *WalletManager) ImportMerchantAddress(wallet *openwallet.Wallet, addresses []*openwallet.Address) error {

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
		a.WalletID = wallet.WalletID
		log.Printf("import %s address: %s", a.Symbol, a.Address)
		tx.Save(a)
	}

	tx.Commit()

	return nil
}

//CreateMerchantAddress 创建钱包地址
func (w *WalletManager) CreateMerchantAddress(wallet *openwallet.Wallet, count int) ([]*openwallet.Address, error) {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return nil, errors.New("The wallet node is not config!")
	}

	_, newAddrs, err := CreateBatchAddress(wallet.Alias, wallet.Password, uint64(count))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//GetMerchantAddressList 获取钱包地址
func (w *WalletManager) GetMerchantAddressList(wallet *openwallet.Wallet, offset uint64, limit uint64) ([]*openwallet.Address, error) {
	return nil, nil
}

//SubmitTransaction 提交转账申请
func (w *WalletManager) SubmitTransactions(wallet *openwallet.Wallet, withdraws []*openwallet.Withdraw) (*openwallet.Transaction, error) {
	coreWallet := readWallet(wallet.KeyFile)
	addrs := make([]string, 0)
	amounts := make([]decimal.Decimal, 0)
	for _, with := range withdraws {
		amount, _ := decimal.NewFromString(with.Amount)
		addrs = append(addrs, with.Address)
		amounts = append(amounts, amount)

	}

	txID, err := SendBatchTransaction(coreWallet.WalletID, addrs, amounts, wallet.Password)
	if err != nil {
		return nil, err
	}
	t := openwallet.Transaction{TxID: txID}

	return &t, nil
}
