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
	WORKPATH string
	CMD      [2][]string // Commands to run fullnode wallet ex: {{"/bin/sh", "mainnet"}, {"/bin/sh", "testnet"}}
	PORT     [][3]string // Which ports need to be mapped, ex: {{innerPort, mainNetPort, testNetPort}, ...}
	APIPORT  string      // Port of default fullnode API(within container), from PORT
	IMAGE    string      // Image that container run from
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
			CMD: [2][]string{{"/usr/bin/bitcoind", "-datadir=/openwallet/data", "-conf=/etc/bitcoin.conf"},
				{"/usr/bin/bitcoind", "-datadir=/openwallet/testdata", "-conf=/etc/bitcoin-test.conf"}},
			PORT:    [][3]string{{"18332/tcp", "10001", "20001"}},
			APIPORT: string("18332/tcp"), // Same within PORT
			IMAGE:   string("openwallet/btc:0.15.1"),
		},
		"bch": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/bitcoind", "-datadir=/openwallet/data", "-conf=/etc/bitcoin.conf"},
				{"/usr/bin/bitcoind", "-datadir=/openwallet/testdata", "-conf=/etc/bitcoin-test.conf"}},
			PORT:    [][3]string{{"18335/tcp", "10011", "20011"}},
			APIPORT: string("18335/tcp"),
			IMAGE:   string("openwallet/bch:0.17.1"),
		},
		"eth": &FullnodeContainerConfig{
			// CMD: [2][]string{{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/data", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"},
			// 	{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/testdata", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"}},
			CMD: [2][]string{{"/bin/bash", "-c", "/usr/sbin/geth.eth -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/data --port=30301 --rpcapi=eth,personal,net >> /openwallet/data/run.log 2>&1"},
				{"/bin/bash", "-c", "cp -rf /root/chain/* /openwallet/testdata/ && /usr/sbin/geth.eth --identity TestNode -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/testdata --port=30301 --rpcapi=eth,personal,net --nodiscover >> /openwallet/testdata/run.log 2>&1"}},
			PORT:    [][3]string{{"8545/tcp", "10021", "20021"}},
			APIPORT: string("8545/tcp"),
			IMAGE:   string("openwallet/eth:geth-1.7.3"),
		},
		"eos": &FullnodeContainerConfig{
			CMD:     [2][]string{{"/bin/bash", "-c", "while sleep 1; do date; done"}, {}},
			PORT:    [][3]string{{"3000/tcp", "10031", "20031"}, {"8888/tcp", "10032", "20032"}},
			APIPORT: string("3000/tcp"),
			IMAGE:   string("openwallet/eos:latest"),
		},
		"sc": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"},
				{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"}},
			PORT:    [][3]string{{"9980/tcp", "10041", "20041"}, {"9981/tcp", "10042", "20042"}},
			APIPORT: string("9980/tcp"),
			IMAGE:   string("openwallet/sc:1.3.3"),
		},
		"iota": &FullnodeContainerConfig{
			CMD:     [2][]string{{"/bin/bash", "-c", "while sleep 1; do date; done"}, {}},
			PORT:    [][3]string{{"18265/tcp", "10051", "20051"}},
			APIPORT: string("18265/tcp"),
			IMAGE:   string("openwallet/iota:latest"),
		},
		"bopo": &FullnodeContainerConfig{
			WORKPATH: "/usr/local/paicode",
			CMD:      [2][]string{{"/bin/bash", "-c", "cd /usr/local/paicode; ./gamepaicore --listen 0.0.0.0:7280 >> /openwallet/data/run.log 2>&1"}, {}},
			PORT:     [][3]string{{"7280/tcp", "10061", "20061"}},
			APIPORT:  string("7280/tcp"),
			IMAGE:    string("openwallet/bopo:latest"),
		},
		"hc": &FullnodeContainerConfig{
			WORKPATH: "/usr/local/paicode",
			CMD: [2][]string{
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/data --logdir=/openwallet/data --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 && /usr/local/hypercash/bin/hcwallet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"},
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/testdata --logdir=/openwallet/testdata --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 --testnet && /usr/local/hypercash/bin/hcwallet --testnet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"}},
			PORT:    [][3]string{{"12010/tcp", "12010", "22010"}, {"14009/tcp", "14009", "24009"}},
			APIPORT: string("12010/tcp"),
			IMAGE:   string("openwallet/hc:2.0.3dev"),
		},
	}
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
