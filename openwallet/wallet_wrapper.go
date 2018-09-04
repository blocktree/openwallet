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

package openwallet

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/blocktree/OpenWallet/common"
	"github.com/coreos/bbolt"
	"sync"
	"time"
)

type WalletDBFile string

type WalletKeyFile string

// WalletWrapper 钱包包装器，扩展钱包功能
// 基于OpenWallet钱包体系模型，专门处理钱包的持久化问题，关系数据查询
type WalletWrapper struct {
	wallet       *Wallet      //需要包装的钱包
	walletDB     *StormDB     //存储钱包相关数据的数据库，目前使用boltdb作为持久方案
	mu           sync.RWMutex //锁
	isExternalDB bool         //是否外部加载的数据库，非内部打开，内部打开需要关闭
	dbFile       string       //钱包数据库文件路径，用于内部打开
	keyFile      string       //钱包密钥文件路径
}

func NewWalletWrapper(wallet *Wallet, args ...interface{}) (*WalletWrapper, error) {

	if wallet == nil {
		return nil, fmt.Errorf("wallet is nil")
	}

	wrapper := WalletWrapper{wallet: wallet}

	for _, arg := range args {
		switch obj := arg.(type) {
		case *StormDB:
			if obj != nil {
				if !obj.Opened {
					return nil, fmt.Errorf("wallet db is close")
				}

				wrapper.isExternalDB = true
				wrapper.walletDB = obj
			}
		case WalletDBFile:
			wrapper.dbFile = string(obj)
		case WalletKeyFile:
			wrapper.keyFile = string(obj)
		}
	}

	return &wrapper, nil
}

//OpenStormDB 打开数据库
func (wrapper *WalletWrapper) OpenStormDB() (*StormDB, error) {

	var (
		db  *StormDB
		err error
	)

	if wrapper.walletDB != nil && wrapper.walletDB.Opened {
		return wrapper.walletDB, nil
	}

	//保证数据库文件并发下不被同时打开
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	//解锁进入后，再次确认是否已经存在
	if wrapper.walletDB != nil && wrapper.walletDB.Opened {
		return wrapper.walletDB, nil
	}

	db, err = OpenStormDB(
		wrapper.wallet.DBFile,
		storm.BoltOptions(0600, &bolt.Options{Timeout: 3 * time.Second}),
	)

	if err != nil {
		return nil, err
	}

	wrapper.isExternalDB = false
	wrapper.walletDB = db

	return db, nil
}

//SetStormDB 设置钱包的应用数据库
func (wrapper *WalletWrapper) SetExternalDB(db *StormDB) error {

	//关闭之前的数据库
	wrapper.CloseDB()

	//保证数据库文件并发下不被同时打开
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	wrapper.walletDB = db
	wrapper.isExternalDB = true

	return nil
}

//CloseDB 关闭数据库
func (wrapper *WalletWrapper) CloseDB() {
	// 如果是外部引入的数据库不进行关闭，因为这样会外部无法再操作同一个数据库实力
	if wrapper.isExternalDB == false {
		if wrapper.walletDB != nil && wrapper.walletDB.Opened {
			wrapper.walletDB.Close()
		}
	}
}

//GetWallet 获取钱包
func (wrapper *WalletWrapper) GetWallet() *Wallet {
	return wrapper.wallet
}

// GetAssetsAccountInfo 获取指定账户
func (wrapper *WalletWrapper) GetAssetsAccountInfo(accountID string) (*AssetsAccount, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var account AssetsAccount
	err = db.One("AccountID", accountID, &account)
	if err != nil {
		return nil, fmt.Errorf("can not find account: %s", accountID)
	}

	return &account, nil
}

//GetAssetsAccounts 获取某种区块链的全部资产账户
func (wrapper *WalletWrapper) GetAssetsAccountList(offset, limit int, cols ...interface{}) ([]*AssetsAccount, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var accounts []*AssetsAccount

	query := make([]q.Matcher, 0)

	query = append(query, q.Eq("WalletID", wrapper.wallet.WalletID))

	for i := 0; i < len(cols); i = i + 2 {
		field := common.NewString(cols[i])
		val := cols[i+1]
		query = append(query, q.Eq(field.String(), val))
	}

	if limit > 0 {

		err = db.Select(q.And(
			query...,
		)).Limit(limit).Skip(offset).Find(&accounts)

	} else {

		err = db.Select(q.And(
			query...,
		)).Skip(offset).Find(&accounts)

	}

	if err != nil {
		return nil, fmt.Errorf("can not find accounts")
	}

	return accounts, nil
}

//GetAddress 通过地址字符串获取地址对象
func (wrapper *WalletWrapper) GetAddress(address string) (*Address, error) {
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var obj Address
	err = db.One("Address", address, &obj)

	if err != nil {
		return nil, fmt.Errorf("can not find address")
	}

	return &obj, nil
}

// GetAddresses 获取资产账户地址列表
func (wrapper *WalletWrapper) GetAddressList(accountID string, offset, limit int, watchOnly bool) ([]*Address, error) {
	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var addrs []*Address

	if limit > 0 {

		err = db.Select(q.And(
			q.Eq("AccountID", accountID),
			q.Eq("WatchOnly", watchOnly),
		)).Limit(limit).Skip(offset).Find(&addrs)

		//err = db.Find("AccountID", walletID, &addresses, storm.Limit(limit), storm.Skip(offset))
	} else {

		err = db.Select(q.And(
			q.Eq("AccountID", accountID),
			q.Eq("WatchOnly", watchOnly),
		)).Skip(offset).Find(&addrs)

	}

	if err != nil {
		return nil, fmt.Errorf("can not find addresses")
	}

	return addrs, nil
}
