/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package cardano

import (
	"fmt"
	"github.com/blocktree/openwallet/common/file"
	"path/filepath"
	"github.com/blocktree/go-owcrypt"
	"github.com/shopspring/decimal"
	"time"
	"strings"
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
	Symbol    = "ADA"
	MasterKey = "ada seed"
	CurveType = owcrypt.ECC_CURVE_SECP256K1
)

type WalletConfig struct {
	//币种
	Symbol    string
	MasterKey string

	keyDir string
	//地址导出路径
	addressDir string
	//证书路径
	CertsDir string
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
	//转账最小额度
	MinSendAmount decimal.Decimal
	//最小矿工费
	MinFees decimal.Decimal
}

func NewConfig(symbol string, masterKey string) *WalletConfig {
	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.MasterKey = masterKey
	c.CurveType = CurveType
	//钥匙备份路径
	c.keyDir = filepath.Join("data", strings.ToLower(c.Symbol), "key")
	//证书路径
	c.CertsDir = filepath.Join("data", strings.ToLower(c.Symbol), "certs")
	//地址导出路径
	c.addressDir = filepath.Join("data", strings.ToLower(c.Symbol), "address")
	//区块链数据
	//blockchainDir = filepath.Join("data", strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = c.Symbol + ".ini"
	//是否测试网络
	c.IsTestNet = false
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.Symbol), "db")
	//备份路径
	c.backupDir = filepath.Join("data", strings.ToLower(c.Symbol), "backup")
	//钱包服务API
	c.ServerAPI = ""
	//钱包安装的路径
	c.NodeInstallPath = ""
	//钱包数据文件目录
	c.WalletDataPath = ""
	//汇总阀值
	c.Threshold = decimal.NewFromFloat(10000).Mul(coinDecimal)
	//汇总地址
	c.SumAddress = ""
	//最小转账额度
	c.MinSendAmount = decimal.NewFromFloat(100).Mul(coinDecimal)
	//最小手续费
	c.MinFees = decimal.NewFromFloat(0.3).Mul(coinDecimal)
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
apiUrl = ""
# transaction min send amount, 
minSendAmount = ""
# the safe address that wallet send money to.
sumAddress = ""
# when wallet's balance is over this value, the wallet willl send money to [sumAddress]
#unit is
threshold = ""
# summary task timer cycle time, sample: 1h, 1h1m , 2m, 30s, 3m20s etc...
cycleSeconds = ""
#min fees
minFees = ""
#wallet data path
walletDataPath = ""
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
