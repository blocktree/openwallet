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
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
	"github.com/btcsuite/btcutil/hdkeychain"
)

var (
	//ID首字节的标识
	AccountIDVer = []byte{0x09}
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
	WalletID  string   `json:"walletID"`             //钱包ID
	Alias     string   `json:"alias"`                //别名
	AccountID string   `json:"accountID" storm:"id"` //账户ID，合成地址
	Index     uint64   `json:"index"`                //账户ID，索引位
	HDPath    string   `json:"hdPath"`               //衍生路径
	PublicKey string   `json:"publicKey"`            //主公钥
	OwnerKeys []string `json:"ownerKeys"`            //公钥数组，大于1为多签
	//Owners          map[string]AccountOwner //拥有者列表, 账户公钥: 拥有者
	ContractAddress string      `json:"contractAddress"` //多签合约地址
	Required        uint64      `json:"required"`        //必要签名数
	Symbol          string      `json:"symbol"`            //资产币种类别
	AddressIndex    int         `json:"addressCount"`
	Balance         string      `json:"balance"`
	core            interface{} //核心账户指针
}

func NewMultiSigAccount(wallets []*Wallet, required uint, creator *Wallet) (*AssetsAccount, error) {

	return nil, nil
}

//NewUserAccount 创建账户
func NewUserAccount() *AssetsAccount {
	account := &AssetsAccount{}
	return account
}

func (a *AssetsAccount) GetOwners() []AccountOwner {
	return nil
}

//computeKeyID 计算HDKey的KeyID
func (a *AssetsAccount) GetAccountID() string {

	if len(a.AccountID) > 0 {
		return a.AccountID
	}

	pub, err := owkeychain.OWDecode(a.PublicKey)
	if err != nil {
		return ""
	}

	//seed 通过hmac-sha256 两次 RIPEMD160 一次 得到keyID
	hash := crypto.Keccak256(pub.GetPublicKeyBytes())
	hash = crypto.Keccak256(hash)

	a.AccountID = owkeychain.Base58checkEncode(hash, AccountIDVer)

	return a.AccountID
}
