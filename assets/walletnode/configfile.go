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
// "github.com/shopspring/decimal"
// "path/filepath"
// "strings"
// "time"
)

var (
	Symbol string
)

// Node setup 节点配置
var (
	//是否测试网络
	isTestNet = true // default in TestNet
	// Docker 主节点API服务地址
	dockerAddr = ""
	// Docker 主节点API服务端口
	dockerPort = ""

	//    //RPC认证账户名
	//    rpcUser = ""
	//    //RPC认证账户密码
	//    rpcPassword = ""
	//    //钥匙备份路径
	//    keyDir = filepath.Join("data", strings.ToLower(Symbol), "key")
	//    //地址导出路径
	//    addressDir = filepath.Join("data", strings.ToLower(Symbol), "address")
	//    //配置文件路径
	//    configFilePath = filepath.Join("conf")
	//    //配置文件名
	//    configFileName = Symbol + ".ini"
	//    // 核心钱包是否只做监听
	//    CoreWalletWatchOnly = true
	//    //最大的输入数量
	//    maxTxInputs = 50
	//    //本地数据库文件路径
	//    dbPath = filepath.Join("data", strings.ToLower(Symbol), "db")
	//    //备份路径
	//    backupDir = filepath.Join("data", strings.ToLower(Symbol), "backup")
	//    //钱包服务API
	//    serverAPI = "http://127.0.0.1:10000"
	//    //钱包安装的路径
	//    nodeInstallPath = ""
	//    //钱包数据文件目录
	//    walletDataPath = ""
	//    //小数位长度
	//    coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
	//    //参与汇总的钱包
	//    // walletsInSum = make(map[string]*Wallet)	// ? 500
	//    //汇总阀值
	//    threshold decimal.Decimal = decimal.NewFromFloat(5)
	//    //汇总地址
	//    sumAddress = ""
	//    //汇总执行间隔时间
	//    cycleSeconds = time.Second * 10
	//默认配置内容
	defaultConfig = `# WMD
# start node command
startNodeCMD = ""
# stop node command
stopNodeCMD = ""
# node install path
nodeInstallPath = ""
# mainnet data path
mainNetDataPath = ""
# testnet data path
testNetDataPath = ""
# node api url
apiURL = ""
# RPC Authentication Username
rpcUser = ""
# RPC Authentication Password
rpcPassword = ""
# Is network test?
isTestNet = false
# the safe address that wallet send money to.
sumAddress = ""
# when wallet's balance is over this value, the wallet willl send money to [sumAddress]
threshold = ""

# docker master server addr
dockerAddr = "127.0.0.1"
# docker master server port
dockerPort = "2375"
`
)
