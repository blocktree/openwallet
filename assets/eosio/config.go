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

package eosio

import (
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/common/file"
	"github.com/shopspring/decimal"
	"path/filepath"
	"strings"
)

const (
	//币种
	Symbol    = "EOS"
	CurveType = owcrypt.ECC_CURVE_SECP256K1

	//默认配置内容
	defaultConfig = `

# RPC api url
serverAPI = ""
# RPC Authentication Username
rpcUser = ""
# RPC Authentication Password
rpcPassword = ""

`
)

type WalletConfig struct {

	//币种
	Symbol string
	//RPC认证账户名
	RpcUser string
	//RPC认证账户密码
	RpcPassword string
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
	//区块链数据文件
	BlockchainFile string
	//最大的输入数量
	MaxTxInputs int
	//本地数据库文件路径
	dbPath string
	//钱包服务API
	ServerAPI string
	//默认配置内容
	DefaultConfig string
	//曲线类型
	CurveType uint32
	//小数位长度
	CoinDecimal decimal.Decimal
	//链ID
	ChainID uint64
}

func NewConfig(symbol string) *WalletConfig {

	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.CurveType = CurveType

	//RPC认证账户名
	c.RpcUser = ""
	//RPC认证账户密码
	c.RpcPassword = ""
	//区块链数据
	//blockchainDir = filepath.Join("data", strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = c.Symbol + ".ini"
	//区块链数据文件
	c.BlockchainFile = "blockchain.db"
	//最大的输入数量
	c.MaxTxInputs = 50
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.Symbol), "db")
	//钱包服务API
	c.ServerAPI = ""
	//小数位长度
	c.CoinDecimal = decimal.NewFromFloat(100000000)

	//创建目录
	file.MkdirAll(c.dbPath)

	return &c
}
