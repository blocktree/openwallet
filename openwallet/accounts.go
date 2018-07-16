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
	"github.com/btcsuite/btcutil/hdkeychain"
)

//解锁的密钥
type unlocked struct {
	Key   *hdkeychain.ExtendedKey
	abort chan struct{}
}

//UserAccount 用户的钱包账户组
type UserAccount struct {
	//用户
	*User
	//账户ID，相当与用于账户的扩展密钥的根地址
	AccountId string
	//钱包数组
	Wallets []*Wallet
	// 已解锁的钱包，集合（钱包地址, 钱包私钥）
	unlocked map[string]unlocked
}


//NewUserAccount 创建账户
func NewUserAccount(user User) *UserAccount {
	//TODO:实现从节点配置文件中读取私钥的存粗文件夹
	//cfg := gethConfig{Node: defaultNodeConfig()}
	//// Load config file.
	//if file := ctx.GlobalString(configFileFlag.Name); file != "" {
	//	if err := loadConfig(file, &cfg); err != nil {
	//		utils.Fatalf("%v", err)
	//	}
	//}
	//utils.SetNodeConfig(ctx, &cfg.Node)
	//scryptN, scryptP, keydir, err := cfg.Node.AccountConfig()
	//
	//if err != nil {
	//	utils.Fatalf("Failed to read configuration: %v", err)
	//}
	//
	//password := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))
	//

	//keystore.GenerateHDKey()
	account := &UserAccount{}
	return account
}

// init 初始化用户账户
func (u *UserAccount)init() {
	//1.加载用户所有钱包



}

func (u *UserAccount)loadKeystore() {




}
