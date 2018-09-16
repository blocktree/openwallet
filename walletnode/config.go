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
	"errors"
	"log"
	"strings"
)

const (
	RPCUser       = "walletUser"         //RPC默认的认证账户名
	RPCPassword   = "walletPassword2017" //RPC默认的认证账户密码
	RPCDockerPort = "9360/tcp"           //Docker中默认的RPC端口

	// DockerAllowed = "127.0.0.1" // ?500 For productive environment
	DockerAllowed = "0.0.0.0"

	MainNetDataPath = "/data" //容器中目录，实则在物理机："/openwallet/<Symbol>/data"
	TestNetDataPath = "/data" //容器中目录，实则在物理机："/openwallet/<Symbol>/testdata"

	MountSrcPrefix = "/openwallet/data" // The prefix to mounted source directory
	MountDstDir    = "/data"            // Which directory will be mounted in container
)

// Node setup 节点配置
type WalletnodeConfig struct {

	//钱包全节点服务
	walletnodePrefix          string // container prefix for name to run same within one server
	walletnodeServerType      string // "service"/"localdocker"/"remotedocker"
	walletnodeServerAddr      string // type:remotedocker required
	walletnodeServerPort      string // type:remotedocker required
	walletnodeStartNodeCMD    string // type:local required (from old: startNodeCMD)
	walletnodeStopNodeCMD     string // type:local required (from old: stopNodeCMD)
	walletnodeMainNetDataPath string
	walletnodeTestNetDataPath string
	walletnodeIsEncrypted     string // true/false
	// walletnodeServerSocket string "/var/run/docker.sock" // type:localdocker required
	// walletnodePubAPIs      string ""                     // walletnode returns API to rpc client, etc.

	isTestNet   string // 是否测试网络，default in TestNet
	RPCUser     string // RPC认证账户名
	RPCPassword string // RPC认证账户密码
	WalletURL   string // Fullnode API URL

	//------------------------------------------------------------------------------
	//默认配置内容
	defaultConfig string
}

func (wc WalletnodeConfig) getDataDir() (string, error) {
	if wc.isTestNet == "true" {
		if wc.walletnodeTestNetDataPath == "" {
			return "", errors.New("TestNetDataPath not config!")
		} else {
			return wc.walletnodeTestNetDataPath, nil
		}
	} else {
		if wc.walletnodeMainNetDataPath == "" {
			return "", errors.New("")
		} else {
			return wc.walletnodeMainNetDataPath, nil
		}
	}
}

type FullnodeContainerConfig struct {
	// WORKPATH    string
	CMD     [2][]string // Commands to run fullnode wallet ex: {{"/bin/sh", "mainnet"}, {"/bin/sh", "testnet"}}
	PORT    [][3]string // Which ports need to be mapped, ex: {{innerPort, mainNetPort, testNetPort}, ...}
	RPCPORT string      // Port of default fullnode API(within container), from PORT
	IMAGE   string      // Image that container run from
	ENCRYPT []string    // Encrypt wallet fullnode as an option
}

func (nc *FullnodeContainerConfig) isEncrypted() bool {
	if nc.ENCRYPT != nil {
		return true
	} else {
		return false
	}
}

func GetFullnodeConfig(symbol string) *FullnodeContainerConfig {
	if v, exist := FullnodeContainerConfigs[strings.ToLower(symbol)]; exist {
		return v
	} else {
		return nil
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	WNConfig = &WalletnodeConfig{

		//钱包全节点服务
		walletnodePrefix:          "",
		walletnodeServerType:      "",
		walletnodeServerAddr:      "",
		walletnodeServerPort:      "",
		walletnodeStartNodeCMD:    "",
		walletnodeStopNodeCMD:     "",
		walletnodeMainNetDataPath: MainNetDataPath,
		walletnodeTestNetDataPath: TestNetDataPath,
		// walletnodeServerSocket string "/var/run/docker.sock" // type:localdocker required
		// walletnodePubAPIs      string ""                     // walletnode returns API to rpc client, etc.

		isTestNet:   "",
		RPCUser:     "walletUser",
		RPCPassword: "walletPassword2017",
		WalletURL:   "",

		//------------------------------------------------------------------------------
		//默认配置内容
		defaultConfig: `
# Is network test?
isTestNet = true
# node api url
WalletURL = ""
# RPC Authentication Username
rpcUser = "walletUser"
# RPC Authentication Password
rpcPassword = "walletPassword2017"
		
[walletnode]
# walletnode server type: local/docker
servertype = "docker"
# remote docker master server addr
serveraddr = "192.168.2.194"
# remote docker master server port
serverport = "2375"
# local docker master server socket if serveraddr=="" and servertype=="docker"
serversocket = "/var/run/docker.socket"

# prefix for container name
prefix = "openw_",
# wallet fullnode is crypted?
isEnCrypted = ""

# start node command if servertype==local
startNodeCMD = "",
# stop node command if servertype==local
stopNodeCMD = "",

# mainnet data path
mainNetDataPath = "/data"
# testnet data path
testNetDataPath = "/data"
`,
	}

	FullnodeContainerConfigs = map[string]*FullnodeContainerConfig{
		"btc": &FullnodeContainerConfig{
			PORT:    [][3]string{{"18332/tcp", "10001", "20001"}},
			RPCPORT: string("18332/tcp"),
			IMAGE:   string("openw/btc:v0.15.1"),
			ENCRYPT: []string{"bitcoin-cli", "-datadir=/data", "-conf=/etc/litecoin.conf", "encryptwallet 1234qwer"},
		},
		"eth": &FullnodeContainerConfig{
			// CMD: [2][]string{{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/data", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"},
			// 	{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/testdata", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"}},
			// CMD: [2][]string{{"/bin/bash", "-c", "/usr/sbin/geth.eth -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/data --port=30301 --rpcapi=eth,personal,net >> /openwallet/data/run.log 2>&1"},
			// 	{"/bin/bash", "-c", "cp -rf /root/chain/* /openwallet/testdata/ && /usr/sbin/geth.eth --identity TestNode -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/testdata --port=30301 --rpcapi=eth,personal,net --nodiscover >> /openwallet/testdata/run.log 2>&1"}},
			PORT:    [][3]string{{"8545/tcp", "10002", "20002"}},
			RPCPORT: string("8545/tcp"),
			IMAGE:   string("openw/eth:geth-v1.8.15"),
		},
		"eos": &FullnodeContainerConfig{ // Writing
			PORT:    [][3]string{{"8888/tcp", "10003", "20003"}},
			RPCPORT: string("8888/tcp"),
			IMAGE:   string("openw/eos:v1.2.5"),
		},
		// "iota": &FullnodeContainerConfig{
		// 	CMD:     [2][]string{{"/bin/bash", "-c", "while sleep 1; do date; done"}, {}},
		// 	PORT:    [][3]string{{"18265/tcp", "10004", "20004"}},
		// 	RPCPORT: string("18265/tcp"),
		// 	IMAGE:   string("openwallet/iota:latest"),
		// },

		"bch": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/bitcoind", "-datadir=/openwallet/data", "-conf=/etc/bitcoin.conf"},
				{"/usr/bin/bitcoind", "-datadir=/openwallet/testdata", "-conf=/etc/bitcoin-test.conf"}},
			PORT:    [][3]string{{"18335/tcp", "10011", "20011"}},
			RPCPORT: string("18335/tcp"),
			IMAGE:   string("openwallet/bch:0.17.1"),
		},
		"bopo": &FullnodeContainerConfig{
			CMD:     [2][]string{{}, {}},
			PORT:    [][3]string{{"9360/tcp", "10021", "20021"}},
			RPCPORT: string("9360/tcp"),
			IMAGE:   string("openw/bopo:latest"),
		},
		"qtum": &FullnodeContainerConfig{
			CMD:     [2][]string{},
			PORT:    [][3]string{{"9360/tcp", "10031", "20031"}},
			RPCPORT: string("9360/tcp"),
			IMAGE:   string("openw/qtum:0.15.3"),
		},
		"sc": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"},
				{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"}},
			PORT:    [][3]string{{"9980/tcp", "10041", "20041"}, {"9981/tcp", "10042", "20042"}},
			RPCPORT: string("9980/tcp"),
			IMAGE:   string("openwallet/sc:1.3.3"),
		},
		"hc": &FullnodeContainerConfig{
			CMD: [2][]string{
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/data --logdir=/openwallet/data --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 && /usr/local/hypercash/bin/hcwallet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"},
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/testdata --logdir=/openwallet/testdata --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 --testnet && /usr/local/hypercash/bin/hcwallet --testnet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"}},
			PORT:    [][3]string{{"12010/tcp", "10051", "20051"}, {"14009/tcp", "14009", "24009"}},
			RPCPORT: string("12010/tcp"),
			IMAGE:   string("openwallet/hc:2.0.3dev"),
		},
		"ltc": &FullnodeContainerConfig{ // litecoin
			CMD:     [2][]string{},
			PORT:    [][3]string{{"9360/tcp", "10061", "20061"}},
			RPCPORT: string("9360/tcp"),
			IMAGE:   string("openw/litecoin:v0.16.0"),
			ENCRYPT: []string{"litecoind", "-datadir=/data", "-conf=/etc/litecoin.conf", "encryptwallet 1234qwer"},
		},
	}
}
