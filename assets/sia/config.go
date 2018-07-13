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

package sia

import (
	"path/filepath"
	"strings"
	"github.com/astaxie/beego/config"
	"encoding/json"
	"github.com/blocktree/OpenWallet/common/file"
	"fmt"
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
	Symbol    = "SC"
	MasterKey = "Siacoin seed"
)

var (
	dataDir = filepath.Join("data", strings.ToLower(Symbol))
	//钥匙备份路径
	keyDir = filepath.Join(dataDir, "key")
	//地址导出路径
	addressDir = filepath.Join(dataDir, "address")
	//配置文件路径
	configFilePath = filepath.Join("conf")
	//配置文件名
	configFileName = Symbol + ".ini"
	//接口授权密码
	rpcPassword = "123"
	//备份路径
	backupDir = filepath.Join(dataDir, "backup")
	//钱包数据文件目录
	walletDataPath = "C:/Users/Administrator/AppData/Roaming/Sia-UI/sia/wallet"
	//walletDataPath = ""
	//本地数据库文件路径
	dbPath = filepath.Join(dataDir, "db")
	//钱包安装的路径
	nodeInstallPath = ""
	//参与汇总的钱包
	walletsInSum = make(map[string]*Wallet)
	//默认配置内容
	defaultConfig = `
# start node command
startNodeCMD = ""
# stop node command
stopNodeCMD = ""
# node install path
nodeInstallPath = ""
# when wallet's balance is over this value, the wallet will send money to [sumAddress]
threshold = ""
# node api url
apiURL = ""
# the safe address that wallet send money to.
sumAddress = ""
# Auth password
rpcPassword = ""
# wallet data path for backup
walletDataPath = ""
`
)

//isExistConfigFile 检查配置文件是否存在
func isExistConfigFile() bool {
	_, err := config.NewConfig("ini",
		filepath.Join(configFilePath, configFileName))
	if err != nil {
		return false
	}
	return true
}

//newConfigFile 创建配置文件
func newConfigFile(
	apiURL, walletPath, sumAddress,
	threshold string) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":     apiURL,
		"walletPath": walletPath,
		"sumAddress": sumAddress,
		"threshold":  threshold,
	}

	filepath.Join()

	bytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, "", err
	}

	//实例化配置
	c, err := config.NewConfigData("ini", bytes)
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
