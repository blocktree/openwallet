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

package bopo

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/common/file"
	"github.com/shopspring/decimal"
)

const (
	//币种
	Symbol    = "BOPO"
	MasterKey = "BopoPlat"
)

//var (
//	//秘钥存取
//	// storage *keystore.HDKeystore
//	// 节点客户端
//	client *Client
//)

// Node setup 节点配置
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

	//小数位长度
	coinDecimal decimal.Decimal
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
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
	c.keyDir = filepath.Join("data", strings.ToLower(Symbol), "key")
	//地址导出路径
	c.addressDir = filepath.Join("data", strings.ToLower(Symbol), "address")
	//区块链数据文件
	c.blockchainFile = "blockchain.db"
	//是否测试网络
	c.isTestNet = true
	// 核心钱包是否只做监听
	c.CoreWalletWatchOnly = true
	//最大的输入数量
	c.maxTxInputs = 50
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(Symbol), "db")
	//备份路径
	c.backupDir = filepath.Join("data", strings.ToLower(Symbol), "backup")
	//链服务API
	// c.chainAPI = "http://127.0.0.1:10000"
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

	//小数位长度
	c.coinDecimal = decimal.NewFromFloat(100000000)
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = Symbol + ".ini"
	//默认配置内容
	c.defaultConfig = `
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
`

	return &c
}

// //newConfigFile 创建配置文件
// func newConfigFile(
// 	apiURL, walletPath, sumAddress string,
// 	threshold string, isTestNet bool) (config.Configer, string, error) {

// 	//	生成配置
// 	configMap := map[string]interface{}{
// 		"apiURL":     apiURL,
// 		"walletPath": walletPath,
// 		"sumAddress": sumAddress,
// 		"threshold":  threshold,
// 		"isTestNet":  isTestNet,
// 	}

// 	filepath.Join()

// 	bytes, err := json.Marshal(configMap)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	//实例化配置
// 	c, err := config.NewConfigData("json", bytes)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	//写入配置到文件
// 	file.MkdirAll(c.configFilePath)
// 	absFile := filepath.Join(configFilePath, configFileName)
// 	err = c.SaveConfigFile(absFile)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	return c, absFile, nil
// }

//initConfig 初始化配置文件
func (wc *WalletConfig) initConfig() {

	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(wc.configFilePath)
		file.WriteFile(absFile, []byte(wc.defaultConfig), false)
	}

}

//printConfig Print config information
func (wc *WalletConfig) printConfig() error {

	wc.initConfig()

	wc.initConfig()
	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	fmt.Printf("-----------------------------------------------------------\n")

	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil
}

// // loadConfig 读取配置
// func loadConfig() error {

// 	var c config.Configer

// 	//读取配置
// 	absFile := filepath.Join(configFilePath, configFileName)
// 	c, err := config.NewConfig("ini", absFile)
// 	if err != nil {
// 		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
// 	}

// 	serverAPI = c.String("apiURL")
// 	threshold, _ = decimal.NewFromString(c.String("threshold"))
// 	sumAddress = c.String("sumAddress")
// 	rpcUser = c.String("rpcUser")
// 	rpcPassword = c.String("rpcPassword")
// 	nodeInstallPath = c.String("nodeInstallPath")
// 	isTestNet, _ = c.Bool("isTestNet")
// 	if isTestNet {
// 		walletDataPath = c.String("testNetDataPath")
// 	} else {
// 		walletDataPath = c.String("mainNetDataPath")
// 	}

// 	// token := basicAuth(rpcUser, rpcPassword)

// 	client = &Client{
// 		BaseURL: serverAPI,
// 		Debug:   false,
// 		// AccessToken: token,
// 	}
// 	return nil
// }
