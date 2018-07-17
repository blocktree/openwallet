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

package walletnode

import (
	"github.com/shopspring/decimal"
	// "path/filepath"
	// s "strings"
	"time"
)

const (
	DATAPATH = "/storage/openwallet/"
)

var (
	Symbol       string
	FullnodeAddr string
	FullnodePort string
)

// Node setup 节点配置
var (
	//钱包服务服务地址和端口
	serverAddr = "localhost"
	serverPort = "2375"
	//是否测试网络
	isTestNet = "true" // default in TestNet
	//RPC认证账户名
	rpcUser = ""
	//RPC认证账户密码
	rpcPassword = ""
	//钥匙备份路径
	// keyDir = filepath.Join("data", strings.ToLower(Symbol), "key")
	//汇总地址
	sumAddress = ""
	keyDir     = ""
	//地址导出路径
	// addressDir = filepath.Join("data", strings.ToLower(Symbol), "address")
	addressDir = ""
	//配置文件路径
	// configFilePath = filepath.Join("conf")
	configFilePath = ""

	// data path
	mainNetDataPath = ""
	testNetDataPath = ""

	// Fullnode API URL
	apiURL = ""

	//配置文件名
	configFileName = Symbol + ".ini"
	// 核心钱包是否只做监听
	CoreWalletWatchOnly = true
	//最大的输入数量
	maxTxInputs = 50
	//本地数据库文件路径
	//dbPath = filepath.Join("data", strings.ToLower(Symbol), "db")
	dbPath = ""
	//备份路径
	// backupDir = filepath.Join("data", strings.ToLower(Symbol), "backup")
	backupDir = ""
	//钱包数据文件目录
	walletDataPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
	//参与汇总的钱包
	// walletsInSum = make(map[string]*Wallet)	// ? 500
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(5)
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
	//默认配置内容
	defaultConfig = `
# start node command
startNodeCMD = ""
# stop node command
stopNodeCMD = ""
# node install path
nodeInstallPath = ""
# node api url
apiURL = ""
# Is network test?
isTestNet = false
# RPC Authentication Username
rpcUser = ""
# RPC Authentication Password
rpcPassword = ""
# mainnet data path
mainNetDataPath = ""
# testnet data path
testNetDataPath = ""
# the safe address that wallet send money to.
sumAddress = ""
# when wallet's balance is over this value, the wallet willl send money to [sumAddress]
threshold = ""

# docker master server addr
serverAddr = "localhost"
# docker master server port
serverPort = ""
`
)
