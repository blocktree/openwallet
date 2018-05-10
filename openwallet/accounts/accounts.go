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

package accounts

import (
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/blocktree/OpenWallet/openwallet"
)

//解锁的密钥
type unlocked struct {
	Key   *hdkeychain.ExtendedKey
	abort chan struct{}
}

//UserAccount 用户的钱包账户组
type UserAccount struct {
	//用户
	*openwallet.User
	//储存的账户钥匙
	store *keystore.HDKeystore
	//钱包数组
	Wallets []*openwallet.Wallet
	// 已解锁的钱包，集合（钱包地址, 钱包私钥）
	unlocked map[string]unlocked
}


// NewUserAccount 创建账户
//func NewUserAccount(userKey, keydir string) *UserAccount {
//	keydir, _ = filepath.Abs(keydir)
//	acount := &UserAccount{
//		UserKey: userKey,
//		store: &keystore.HDKeystore{
//			userKey,
//			keydir,
//			keystore.StandardScryptN,
//			keystore.StandardScryptP,
//		},
//	}
//
//	return acount
//}

// init 初始化用户账户
func (u *UserAccount)init() {
	//1.加载用户所有钱包



}

func (u *UserAccount)loadKeystore() {




}
