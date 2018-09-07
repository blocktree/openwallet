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

package bitcoincash

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/shopspring/decimal"
	"path/filepath"
	"strings"
	"time"
)

const (
	// 币
	Symbol    = "BCH"
	MasterKey = "BitcoinCash Seed"
)

// Node setup 节点配置
var (
	//RPC认证账户名
	rpcUser = ""
	//RPC认证账户密码
	rpcPassword = ""
	//钥匙备份路径
	keyDir = filepath.Join("data", strings.ToLower(Symbol), "key")
	//地址导出路径
	addressDir = filepath.Join("data", strings.ToLower(Symbol), "address")
	//配置文件路径
	configFilePath = filepath.Join("conf")
	//配置文件名
	configFileName = Symbol + ".ini"
	//是否测试网络
	isTestNet = true
	// 核心钱包是否只做监听
	CoreWalletWatchOnly = true
	//最大的输入数量
	maxTxInputs = 50
	//本地数据库文件路径
	dbPath = filepath.Join("data", strings.ToLower(Symbol), "db")
	//备份路径
	backupDir = filepath.Join("data", strings.ToLower(Symbol), "backup")
	//钱包服务API
	serverAPI = "http://127.0.0.1:10000"
	//钱包安装的路径
	nodeInstallPath = ""
	//钱包数据文件目录
	walletDataPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
	//参与汇总的钱包
	walletsInSum = make(map[string]*Wallet)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(5)
	//汇总地址
	sumAddress = ""
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
# summary task timer cycle time, sample: 1m , 30s, 3m20s etc
cycleSeconds = ""
`
)

func setWalletConfig(sumAddress string, threshold string) {

}

//newConfigFile 创建配置文件
func newConfigFile(
	apiURL, walletPath, sumAddress string,
	threshold string, isTestNet bool) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":     apiURL,
		"walletPath": walletPath,
		"sumAddress": sumAddress,
		"threshold":  threshold,
		"isTestNet":  isTestNet,
	}

	filepath.Join()

	bytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, "", err
	}

	//实例化配置
	c, err := config.NewConfigData("json", bytes)
	if err != nil {
		return nil, "", err
	}

	//写入配置到文件
	file.MkdirAll(configFilePath)
	absFile := filepath.Join(configFilePath, configFileName)
	err = c.SaveConfigFile(absFile)
	if err != nil {
		return nil, "", err
	}

	return c, absFile, nil
}

//printConfig Print config information
func printConfig() error {

	initConfig()
	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	//apiURL := c.String("apiURL")
	//walletPath := c.String("walletPath")
	//threshold := c.String("threshold")
	//minSendAmount := c.String("minSendAmount")
	//minFees := c.String("minFees")
	//sumAddress := c.String("sumAddress")
	//isTestNet, _ := c.Bool("isTestNet")

	fmt.Printf("-----------------------------------------------------------\n")
	//fmt.Printf("Wallet API URL: %s\n", apiURL)
	//fmt.Printf("Wallet Data FilePath: %s\n", walletPath)
	//fmt.Printf("Summary Address: %s\n", sumAddress)
	//fmt.Printf("Summary Threshold: %s\n", threshold)
	//fmt.Printf("Min Send Amount: %s\n", minSendAmount)
	//fmt.Printf("Transfer Fees: %s\n", minFees)
	//if isTestNet {
	//	fmt.Printf("Network: TestNet\n")
	//} else {
	//	fmt.Printf("Network: MainNet\n")
	//}
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil
}

//initConfig 初始化配置文件
func initConfig() {

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(configFilePath)
		file.WriteFile(absFile, []byte(defaultConfig), false)
	}

}
