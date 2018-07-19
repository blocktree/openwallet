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
	WalletID   string   `json:"walletID"`              //钱包ID
	Alias      string   `json:"alias"`                 //别名
	AccountID  string   `json:"accountID"  storm:"id"` //账户ID，合成地址
	Index      uint64   `json:"index"`                 //账户ID，索引位
	HDPath     string   `json:"hdPath"`                //衍生路径
	PublicKeys []string `json:"publicKeys"`            //公钥数组，大于1为多签
	//Owners          map[string]AccountOwner //拥有者列表, 账户公钥: 拥有者
	ContractAddress string      `json:"contractAddress"` //多签合约地址
	Required        uint64      `json:"required"`        //必要签名数
	Symbol          string      `json:"coin"`            //资产币种类别
	AddressCount    uint64      `json:"addressCount"`
	Balance         string      `json:"balance"`
	core            interface{} //核心账户指针
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
func (a *AssetsAccount) init() {
	//1.加载用户所有钱包

}

func (a *AssetsAccount) loadKeystore() {

}

func (a *AssetsAccount) GetOwners() []AccountOwner {
	return nil
}
