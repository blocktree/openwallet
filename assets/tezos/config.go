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

package tezos

import (
	"github.com/blocktree/go-OWCrypt"
	"github.com/shopspring/decimal"
	"time"
	"path/filepath"
	"strings"
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
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
	Symbol = "XTZ"
	MasterKey = "Tezos seed"
	CurveType = owcrypt.ECC_CURVE_ED25519
)

/*
var (
	dataDir = filepath.Join("data", strings.ToLower(Symbol))
	//钥匙备份路径
	keyDir = filepath.Join(dataDir, "key")
	//地址导出路径
	addressDir = filepath.Join(dataDir, "address")
	//配置文件路径
	configFilePath = filepath.Join("conf")
	//配置文件名
	configFileName = Symbol + ".json"
	//备份路径
	backupDir = filepath.Join(dataDir, "backup")
	//本地数据库文件路径
	dbPath = filepath.Join(dataDir, "db")
	walletsInSum = make(map[string]*openwallet.Wallet)
)

//isExistConfigFile 检查配置文件是否存在
func isExistConfigFile() bool {
	_, err := config.NewConfig("json",
		filepath.Join(configFilePath, configFileName))
	if err != nil {
		return false
	}
	return true
}

//newConfigFile 创建配置文件
func newConfigFile(
	apiURL, sumAddress,
	threshold, minSendAmount, minFees, gasLimit, storageLimit string) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":         apiURL,
		"sumAddress":     sumAddress,
		"threshold":      threshold,
		"minSendAmount":  minSendAmount,
		"minFees":        minFees,
		"gasLimit":       gasLimit,
		"storageLimit":   storageLimit,
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

func printConfig() error {
	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	apiURL := c.String("apiURL")
	threshold := c.String("threshold")
	minSendAmount := c.String("minSendAmount")
	minFees := c.String("minFees")
	sumAddress := c.String("sumAddress")
	gasLimit := c.String("gasLimit")
	storageLimit := c.String("storageLimit")

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("Node API url: %s\n", apiURL)
	fmt.Printf("Summary address: %s\n", sumAddress)
	fmt.Printf("Summary threshold: %s\n", threshold)
	fmt.Printf("Minimum transfer amount: %s\n", minSendAmount)
	fmt.Printf("Transfer fees: %s\n", minFees)
	fmt.Printf("Gas limit: %s\n", gasLimit)
	fmt.Printf("Storage limit: %s\n", storageLimit)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}
*/
type WalletConfig struct {

	//币种
	Symbol    string
	MasterKey string

	keyDir string
	//地址导出路径
	addressDir string
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
	//是否测试网络
	IsTestNet bool
	//本地数据库文件路径
	dbPath string
	//备份路径
	backupDir string
	//钱包服务API
	ServerAPI string
	//gas limit & storage limit
	GasLimit decimal.Decimal
	StorageLimit decimal.Decimal
	//最小矿工费
	MinFee decimal.Decimal
	//钱包安装的路径
	NodeInstallPath string
	//钱包数据文件目录
	WalletDataPath string
	//汇总阀值
	Threshold decimal.Decimal
	//汇总地址
	SumAddress string
	//汇总执行间隔时间
	CycleSeconds time.Duration
	//默认配置内容
	DefaultConfig string
	//曲线类型
	CurveType uint32
}

func NewConfig(symbol string, masterKey string) *WalletConfig {

	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.MasterKey = masterKey
	c.CurveType = CurveType
	//钥匙备份路径
	c.keyDir = filepath.Join("data", strings.ToLower(c.Symbol), "key")
	//地址导出路径
	c.addressDir = filepath.Join("data", strings.ToLower(c.Symbol), "address")
	//区块链数据
	//blockchainDir = filepath.Join("data", strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = c.Symbol + ".ini"
	//是否测试网络
	c.IsTestNet = true
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.Symbol), "db")
	//备份路径
	c.backupDir = filepath.Join("data", strings.ToLower(c.Symbol), "backup")
	//钱包服务API
	c.ServerAPI = "http://192.168.2.192:10053"
	//gas limit & storage limit
	c.GasLimit = decimal.NewFromFloat(100)        //0.0001 XTZ
	c.StorageLimit = decimal.NewFromFloat(100)    //0.0001 XTZ
	//最小矿工费
	c.MinFee = decimal.NewFromFloat(100)
	//钱包安装的路径
	c.NodeInstallPath = ""
	//钱包数据文件目录
	c.WalletDataPath = ""
	//汇总阀值
	c.Threshold = decimal.NewFromFloat(5)
	//汇总地址
	c.SumAddress = ""
	//汇总执行间隔时间
	c.CycleSeconds = time.Second * 10

	//默认配置内容
	c.DefaultConfig = `
# start node command
startNodeCMD = ""
# stop node command
stopNodeCMD = ""
# node install path
nodeInstallPath = ""
# node api url
apiUrl = "https://rpc.tezrpc.me"
# min fees 0.0001 XTZ
minFee = "100"
# gas limit 0.0001 XTZ
gasLimit = "100"
# storage limit 0.0001 XTZ
storageLimit = "100"
# the safe address that wallet send money to.
sumAddress = "tz1WwFXj4QLxZVRgPTZYy67GKT1u7tMmoTyK"
# when wallet's balance is over this value, the wallet willl send money to [sumAddress]  1XTZ
threshold = "1000000"
`

	return &c
}

//printConfig Print config information
func (wc *WalletConfig) PrintConfig() error {

	wc.InitConfig()
	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)

	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}

//initConfig 初始化配置文件
func (wc *WalletConfig) InitConfig() {

	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(wc.configFilePath)
		file.WriteFile(absFile, []byte(wc.DefaultConfig), false)
	}

}
