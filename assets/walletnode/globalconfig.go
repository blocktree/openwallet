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
	"log"
	s "strings"
	"time"
)

var (
	Symbol                   string                              // Current fullnode wallet's symbol
	FullnodeContainerConfigs map[string]*FullnodeContainerConfig // All configs of fullnode wallet for Docker
	// FullnodeContainerPath    *FullnodeContainerPathConfig        // All paths of fullnode wallet on Docker
)

// Node setup 节点配置
var (
	//钱包服务服务地址和端口
	serverAddr = "localhost"
	serverPort = "2375"
	//是否测试网络
	isTestNet = "true" // default in TestNet
	//RPC认证账户名
	rpcUser = "wallet"
	//RPC认证账户密码
	rpcPassword = "walletPassword2017"
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
	mainNetDataPath = "/data"
	testNetDataPath = "/data/testdata"

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
isTestNet = true
# RPC Authentication Username
rpcUser = "wallet"
# RPC Authentication Password
rpcPassword = "walletPassword2017"
# mainnet data path
mainNetDataPath = "/data"
# testnet data path
testNetDataPath = "/data/testdata"
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

type FullnodeContainerPathConfig struct {
	EXEC_PATH string
	DATA_PATH string
}

type FullnodeContainerConfig struct {
	CMD     [2][]string // Commands to run fullnode wallet ex: {{"/bin/sh", "-c", "main"}, {"/bin/sh", "-c", "testnet"}}
	PORT    [][2]string // Which ports need to be mapped, ex: {{hostport, innerport}, ...}
	APIPORT string      // Port of default fullnode API(within container), from PORT
	IMAGE   string      // Image that container run from
}

func (p *FullnodeContainerPathConfig) init(symbol string) error {
	p.EXEC_PATH = s.Replace(p.EXEC_PATH, "<Symbol>", symbol, 1)
	p.DATA_PATH = s.Replace(p.DATA_PATH, "<Symbol>", symbol, 1)
	return nil
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	FullnodeContainerConfigs = map[string]*FullnodeContainerConfig{
		"btc": &FullnodeContainerConfig{
			CMD: [2][]string{{"/exec/bitcoin/bin/bitcoind", "-datadir=/data/", "-conf=/exec/bitcoin.conf"},
				{"/exec/bitcoin/bin/bitcoind", "-datadir=/data/testdata/", "-conf=/exec/bitcoin-test.conf"}},
			PORT:    [][2]string{{"18332/tcp", "10001"}},
			APIPORT: string("18332/tcp"), // Same within PORT
			IMAGE:   string("openwallet/btc:0.15.1"),
		},
		"bch": &FullnodeContainerConfig{
			CMD: [2][]string{{"/exec/bitcoin/bin/bitcoind", "-datadir=/data/", "-conf=/data/bitcoin.conf"},
				{}},
			PORT:    [][2]string{{"18332/tcp", "10011"}},
			APIPORT: string("18332/tcp"),
			IMAGE:   string("openwallet/bch:latest"),
		},
		"eth": &FullnodeContainerConfig{
			CMD: [2][]string{{"/exec/bitcoin/bin/bitcoind", "-datadir=/data/", "-conf=/data/bitcoin.conf"},
				{}},
			PORT:    [][2]string{{"18332/tcp", "10021"}},
			APIPORT: string("18332/tcp"),
			IMAGE:   string("openwallet/eth:latest"),
		},
		"eos": &FullnodeContainerConfig{
			CMD: [2][]string{{"/exec/bitcoin/bin/bitcoind", "-datadir=/data/", "-conf=/data/bitcoin.conf"},
				{}},
			PORT:    [][2]string{{"3000/tcp", "10400"}, {"8888/tcp", "10480"}},
			APIPORT: string("3000/tcp"),
			IMAGE:   string("openwallet/eos:latest"),
		},
	}
	// FullnodeContainerPath = &FullnodeContainerPathConfig{
	// 	EXEC_PATH: "/openwallet/exec/<Symbol>/",
	// 	DATA_PATH: "/openwallet/data/<Symbol>/",
	// }
}

// Private function, generate container name by <Symbol> and <isTestNet>
func _GetCName(symbol string) (string, error) {
	// Load global config
	err := loadConfig(symbol)
	if err != nil {
		return "", err
	}

	// Within testnet, use "<Symbol>_testnet" as container name
	if isTestNet == "true" {
		return s.ToLower(symbol) + "_testnet", nil
	} else {
		return s.ToLower(symbol), nil
	}
}
