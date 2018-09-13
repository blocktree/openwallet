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
	"github.com/asdine/storm/q"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
	"time"
	"encoding/hex"
)

type WalletDBFile WrapperSourceFile

type WalletKeyFile string

// WalletWrapper 钱包包装器，扩展钱包功能
type WalletWrapper struct {
	*AppWrapper
	wallet  *Wallet //需要包装的钱包
	keyFile string  //钱包密钥文件路径

}

func NewWalletWrapper(args ...interface{}) *WalletWrapper {

	wrapper := NewAppWrapper(args...)

	walletWrapper := WalletWrapper{AppWrapper: wrapper}

	for _, arg := range args {
		switch obj := arg.(type) {
		case *Wallet:
			walletWrapper.wallet = obj
		case WalletDBFile:
			walletWrapper.sourceFile = string(obj)
		case WalletKeyFile:
			walletWrapper.keyFile = string(obj)
		case *AppWrapper:
			walletWrapper.AppWrapper = obj
		}
	}

	return &walletWrapper
}

//GetWallet 获取钱包
func (wrapper *WalletWrapper) GetWallet() *Wallet {

	return wrapper.wallet
}

//GetWalletByID 通过钱包ID获取
func (wrapper *WalletWrapper) GetWalletByID(walletID string) (*Wallet, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var wallet Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
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

	if wrapper.wallet != nil {
		query = append(query, q.Eq("WalletID", wrapper.wallet.WalletID))
	}

	if len(cols)%2 != 0 {
		return nil, fmt.Errorf("condition param is not pair")
	}

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

//GetAssetsAccountByAddress 通过地址获取资产账户对象
func (wrapper *WalletWrapper) GetAssetsAccountByAddress(address string) (*AssetsAccount, error) {
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

	var account AssetsAccount
	err = db.One("AccountID", obj.AccountID, &account)
	if err != nil {
		return nil, fmt.Errorf("can not find account by address: %s", address)
	}

	return &account, nil
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
func (wrapper *WalletWrapper) GetAddressList(offset, limit int, cols ...interface{}) ([]*Address, error) {
	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var addrs []*Address

	query := make([]q.Matcher, 0)

	if len(cols)%2 != 0 {
		return nil, fmt.Errorf("condition param is not pair")
	}

	for i := 0; i < len(cols); i = i + 2 {
		field := common.NewString(cols[i])
		val := cols[i+1]
		query = append(query, q.Eq(field.String(), val))
	}

	if limit > 0 {

		err = db.Select(q.And(
			query...,
		)).Limit(limit).Skip(offset).Find(&addrs)

	} else {

		err = db.Select(q.And(
			query...,
		)).Skip(offset).Find(&addrs)

	}

	if err != nil {
		return nil, fmt.Errorf("can not find addresses")
	}

	return addrs, nil
}

// CreateAddress 创建地址
//@param accountID	指定账户
//@param count		创建数量
//@param decoder	地址解释器
//@param isChange	是否找零地址
//@param isTestNet	是否测试网
func (wrapper *WalletWrapper) CreateAddress(accountID string, count uint64, decoder AddressDecoder, isChange bool, isTestNet bool) ([]*Address, error) {

	var (
		newKeys = make([][]byte, 0)
		address string
		addrs   = make([]*Address, 0)
		publicKey string
	)

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	changeIndex := uint32(common.BoolToUInt(isChange))

	for i := uint64(0); i < count; i++ {

		address = ""

		publicKey = ""

		newIndex := account.AddressIndex + 1

		derivedPath := fmt.Sprintf("%s/%d/%d", account.HDPath, changeIndex, newIndex)

		//通过多个拥有者公钥生成地址
		for _, pub := range account.OwnerKeys {

			pubkey, err := owkeychain.OWDecode(pub)
			if err != nil {
				return nil, err
			}

			start, err := pubkey.GenPublicChild(changeIndex)
			newKey, err := start.GenPublicChild(uint32(newIndex))
			newKeys = append(newKeys, newKey.GetPublicKeyBytes())
		}

		if len(account.OwnerKeys) > 1 {
			address, err = decoder.RedeemScriptToAddress(newKeys, account.Required, isTestNet)
			if err != nil {
				return nil, err
			}
			publicKey = ""
		} else {
			address, err = decoder.PublicKeyToAddress(newKeys[0], isTestNet)
			if err != nil {
				return nil, err
			}
			publicKey = hex.EncodeToString(newKeys[0])
		}

		addr := &Address{
			Address:   address,
			AccountID: accountID,
			HDPath:    derivedPath,
			CreatedAt: time.Now(),
			Symbol:    account.Symbol,
			Index:     uint64(newIndex),
			WatchOnly: false,
			IsChange:  isChange,
			PublicKey: publicKey,
		}

		account.AddressIndex = newIndex

		err = tx.Save(account)
		if err != nil {

			return nil, err
		}

		err = tx.Save(addr)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, addr)

	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

//SaveAssetsAccount 更新账户信息
func (wrapper *WalletWrapper) SaveAssetsAccount(account *AssetsAccount) error {
	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return err
	}
	defer wrapper.CloseDB()

	return db.Save(account)
}