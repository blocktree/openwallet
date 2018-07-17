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

//AccountOwner 账户拥有者接口
type AccountOwner interface {

}

//AssetsAccount 千张包资产账户
type AssetsAccount struct {
	//钱包
	Wallet *Wallet
	//别名
	Alias string
	//账户ID
	AccountId string
	//账户ID，索引位
	Index uint
	//账户公钥根路径
	RootPath string
	//类型类型： 单签，多签
	Type WalletType
	//公钥数组
	PublicKeys []Bytes
	//拥有者列表, 钱包账户ID: 拥有者
	Owners map[string]AccountOwner
	//创建者的账户ID
	Creator string
	//多签合约地址
	ContractAddress string
	//必要签名数
	Required uint
	//资产币种类别
	Assets *Assets

	//核心账户指针
	core interface{}
}



func NewMultiSigAccount(wallets []*Wallet, required uint, creator *Wallet) (*AssetsAccount, error) {

	return nil, nil
}

//NewUserAccount 创建账户
func NewUserAccount(user User) *AssetsAccount {
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
	account := &AssetsAccount{}
	return account
}

// init 初始化用户账户
func (u *AssetsAccount)init() {
	//1.加载用户所有钱包



}

func (u *AssetsAccount)loadKeystore() {




}
