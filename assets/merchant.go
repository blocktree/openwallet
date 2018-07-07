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

package assets

//MerchantAssets 钱包与商户交互的资产接口
type MerchantAssets interface {

	//CreateMerchantWallet 创建钱包
	CreateMerchantWallet(alias string, password string) (*MerchantWallet, error)

	//GetMerchantWalletList 获取钱包列表
	GetMerchantWalletList() ([]*MerchantWallet, error)

	//ConfigMerchantWallet 钱包工具配置接口
	ConfigMerchantWallet(wallet *MerchantWallet) error

	//ImportMerchantAddress 导入地址
	ImportMerchantAddress(addresses []*MerchantAddress) error

	//CreateMerchantAddress 创建钱包地址
	CreateMerchantAddress(walletID string, count int) ([]*MerchantAddress, error)

	//GetMerchantAddressList 获取钱包地址
	GetMerchantAddressList(walletID string, offset uint64, limit uint64) ([]*MerchantAddress, error)

}

//MerchantWallet 标准商户钱包模型
type MerchantWallet struct {

	/*
		| 参数名称  | 类型   | 是否可空 | 描述     |
		|-----------|--------|----------|----------|
		| coin      | string | 否       | 币种     |
		| alias     | string | 否       | 钱包别名 |
		| walletID  | string | 否       | 钱包ID   |
		| publicKey | string | 是       | 钱包公钥 |
		| balance   | string | 否       | 余额     |
	*/

	Coin      string
	Alias     string
	WalletID  string
	PublicKey string
	balance   string

	//核心钱包，具体的区块链包的钱包模型
	Core interface{}
}

type MerchantAddress struct {

	/*
		| 参数名称 | 类型   | 是否可空 | 描述   |
		|----------|--------|----------|--------|
		| address  | string | 否       | 币种   |
		| walletID | string | 否       | 钱包ID |
		| balance  | string | 否       | 余额   |
		| isMemo   | bool   | 是       | 是否memo |
		| memo     | string | 否       | 备注     |
	*/

	Address  string
	WalletID string
	Balance  string
	IsMemo   bool
	Memo     string

	Core interface{}

}

// GetMerchantAssets 根据币种类型获取已注册的管理者
func GetMerchantAssets(symbol string) MerchantAssets {
	manager, ok := managers[symbol].(MerchantAssets)
	if !ok {
		return nil
	}
	return manager
}
