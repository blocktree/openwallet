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

package hypercash

import (
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/shopspring/decimal"
	"path/filepath"
	"strings"
	"time"
)

/*
	工具可以读取各个币种钱包的默认的配置资料，
	币种钱包的配置资料放在conf/{symbol}.conf，例如：ADA.conf, BTC.conf，ETH.conf。
	执行wmd wallet -s <symbol> 命令会先检查是否存在该币种钱包的配置文件。
	没有：执行ConfigFlow，配置文件初始化。
	有：执行常规命令。
	使用者还可以通过wmd config -s 进行修改配置文件。
	或执行wmd config flow 重新进行一次配置初始化流程。

*/

const (
	//币种
	Symbol    = "HC"
	MasterKey = "hypercash seed"
)

type WalletConfig struct {

	//币种
	symbol    string
	masterKey string

	//RPC认证账户名
	rpcUser string
	//RPC认证账户密码
	rpcPassword string
	//证书目录
	certsDir string
	//钥匙备份路径
	keyDir string
	//地址导出路径
	addressDir string
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
	//rpc证书
	certFileName string
	//区块链数据文件
	blockchainFile string
	//是否测试网络
	isTestNet bool
	// 核心钱包是否只做监听
	CoreWalletWatchOnly bool
	//最大的输入数量
	maxTxInputs int
	//本地数据库文件路径
	dbPath string
	//备份路径
	backupDir string
	//链服务API
	chainAPI string
	//钱包服务API
	walletAPI string
	//钱包安装的路径
	nodeInstallPath string
	//钱包数据文件目录
	walletDataPath string
	//汇总阀值
	threshold decimal.Decimal
	//汇总地址
	sumAddress string
	//汇总执行间隔时间
	cycleSeconds time.Duration
	//默认配置内容
	defaultConfig string
}

func NewConfig() *WalletConfig {

	c := WalletConfig{}

	//币种
	c.symbol = Symbol
	c.masterKey = MasterKey

	//RPC认证账户名
	c.rpcUser = ""
	//RPC认证账户密码
	c.rpcPassword = ""
	//证书目录
	c.certsDir = filepath.Join("data", strings.ToLower(c.symbol), "certs")
	//钥匙备份路径
	c.keyDir = filepath.Join("data", strings.ToLower(c.symbol), "key")
	//地址导出路径
	c.addressDir = filepath.Join("data", strings.ToLower(c.symbol), "address")
	//区块链数据
	//blockchainDir = filepath.Join("data", strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = c.symbol + ".ini"
	//rpc证书
	c.certFileName = "rpc.cert"
	//区块链数据文件
	c.blockchainFile = "blockchain.db"
	//是否测试网络
	c.isTestNet = true
	// 核心钱包是否只做监听
	c.CoreWalletWatchOnly = true
	//最大的输入数量
	c.maxTxInputs = 50
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.symbol), "db")
	//备份路径
	c.backupDir = filepath.Join("data", strings.ToLower(c.symbol), "backup")
	//链服务API
	c.chainAPI = "http://127.0.0.1:10000"
	//钱包服务API
	c.walletAPI = "http://127.0.0.1:10000"
	//钱包安装的路径
	c.nodeInstallPath = ""
	//钱包数据文件目录
	c.walletDataPath = ""
	//汇总阀值
	c.threshold = decimal.NewFromFloat(5)
	//汇总地址
	c.sumAddress = ""
	//汇总执行间隔时间
	c.cycleSeconds = time.Second * 10

	//默认配置内容
	c.defaultConfig = `
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
# hcd api url
chainAPI = ""
# hcwallet api url
walletAPI = ""
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
`

	return &c
}

//printConfig Print config information
func (wc *WalletConfig) printConfig() error {

	wc.initConfig()
	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	fmt.Printf("-----------------------------------------------------------\n")

	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}

//initConfig 初始化配置文件
func (wc *WalletConfig) initConfig() {

	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(wc.configFilePath)
		file.WriteFile(absFile, []byte(wc.defaultConfig), false)
	}

}
