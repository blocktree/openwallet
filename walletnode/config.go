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

const (
	RPCUser       = "wallet"             //RPC默认的认证账户名
	RPCPassword   = "walletPassword2017" //RPC默认的认证账户密码
	RPCDockerPort = "9360/tcp"           //Docker中默认的RPC端口

	MainNetDataPath = "/data" //容器中目录，实则在物理机："/openwallet/<Symbol>/"
	TestNetDataPath = "/data" //容器中目录，实则在物理机："/openwallet/<Symbol>/testdata"
)

// Node setup 节点配置
type WalletnodeConfig struct {

	//钱包全节点服务
	walletnodePrefix       string "openw_"               // container prefix for name to run same within one server
	walletnodeServerType   string "localdocker"          // "service"/"localdocker"/"remotedocker"
	walletnodeServerAddr   string "192.168.2.194"        // type:remotedocker required
	walletnodeServerPort   string "2375"                 // type:remotedocker required
	walletnodeServerSocket string "/var/run/docker.sock" // type:localdocker required
	// walletnodeStartNodeCMD string ""                     // type:local required (from old: startNodeCMD)
	// walletnodeStopNodeCMD  string ""                     // type:local required (from old: stopNodeCMD)
	// walletnodePubAPIs      string ""                     // walletnode returns API to rpc client, etc.

	TestNet     string "true"               // 是否测试网络，default in TestNet
	RPCUser     string "wallet"             //RPC认证账户名
	RPCPassword string "walletPassword2017" //RPC认证账户密码
	RPCAddr     string
	RPCPort     string

	// data path
	mainNetDataPath string "/openwallet/data"
	testNetDataPath string "/openwallet/testdata"

	// Fullnode API URL
	apiURL string ""

	//------------------------------------------------------------------------------
	//默认配置内容
	defaultConfig string `

# node api url
apiURL = ""
# Is network test?
isTestNet = true
# RPC Authentication Username
rpcUser = "wallet"
# RPC Authentication Password
rpcPassword = "walletPassword2017"
# mainnet data path
mainNetDataPath = "/openwallet/data"
# testnet data path
testNetDataPath = "/openwallet/testdata"
# the safe address that wallet send money to.
sumAddress = ""
# when wallet's balance is over this value, the wallet willl send money to [sumAddress]
threshold =

[walletnode]
# walletnode server type: service/localdocker/remotedocker
walletnodeservertype = "remotedocker"
# remote docker master server addr
walletnodeserveraddr = "192.168.2.194:2375"
# local docker master server socket
walletnodeserversocket = "/var/run/docker.socket"
`
}

type FullnodeContainerConfig struct {
	WORKPATH string
	CMD      [2][]string // Commands to run fullnode wallet ex: {{"/bin/sh", "mainnet"}, {"/bin/sh", "testnet"}}
	PORT     [][3]string // Which ports need to be mapped, ex: {{innerPort, mainNetPort, testNetPort}, ...}
	RPCPORT  string      // Port of default fullnode API(within container), from PORT
	IMAGE    string      // Image that container run from
}

func init() {
	// log.SetFlags(log.Lshortfile | log.LstdFlags)

	FullnodeContainerConfigs = map[string]*FullnodeContainerConfig{
		"btc": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/local/bitcoin/bin/bitcoind", "-datadir=/openwallet/data", "-conf=/usr/local/bitcoin/etc/bitcoin.conf"},
				{"/usr/bin/bitcoind", "-datadir=/openwallet/testdata", "-conf=/etc/bitcoin-test.conf"}},
			PORT:    [][3]string{{"18332/tcp", "10001", "20001"}},
			RPCPORT: string("18332/tcp"), // Same within PORT
			IMAGE:   string("openwallet/btc:0.15.1"),
		},
		"bch": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/bitcoind", "-datadir=/openwallet/data", "-conf=/etc/bitcoin.conf"},
				{"/usr/bin/bitcoind", "-datadir=/openwallet/testdata", "-conf=/etc/bitcoin-test.conf"}},
			PORT:    [][3]string{{"18335/tcp", "10011", "20011"}},
			RPCPORT: string("18335/tcp"),
			IMAGE:   string("openwallet/bch:0.17.1"),
		},
		"eth": &FullnodeContainerConfig{
			// CMD: [2][]string{{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/data", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"},
			// 	{"/usr/bin/parity", "--port=30307", "--datadir=/openwallet/testdata", "--cache-size=4096", "--min-peers=25", "--max-peers=50", "--jsonrpc-interface=0.0.0.0", "--jsonrpc-port=18332"}},
			CMD: [2][]string{{"/bin/bash", "-c", "/usr/sbin/geth.eth -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/data --port=30301 --rpcapi=eth,personal,net >> /openwallet/data/run.log 2>&1"},
				{"/bin/bash", "-c", "cp -rf /root/chain/* /openwallet/testdata/ && /usr/sbin/geth.eth --identity TestNode -rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=/openwallet/testdata --port=30301 --rpcapi=eth,personal,net --nodiscover >> /openwallet/testdata/run.log 2>&1"}},
			PORT:    [][3]string{{"8545/tcp", "10021", "20021"}},
			RPCPORT: string("8545/tcp"),
			IMAGE:   string("openwallet/eth:geth-1.7.3"),
		},
		// "eos": &FullnodeContainerConfig{ // Writing
		// 	CMD:     [2][]string{{"/bin/bash", "-c", "while sleep 1; do date; done"}, {}},
		// 	PORT:    [][3]string{{"3000/tcp", "10031", "20031"}, {"8888/tcp", "10032", "20032"}},
		// 	RPCPORT: string("3000/tcp"),
		// 	IMAGE:   string("openwallet/eos:latest"),
		// },
		"sc": &FullnodeContainerConfig{
			CMD: [2][]string{{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"},
				{"/usr/bin/siad", "-M gctwrh", "--api-addr=0.0.0.0:9980", "--authenticate-api", "--disable-api-security"}},
			PORT:    [][3]string{{"9980/tcp", "10041", "20041"}, {"9981/tcp", "10042", "20042"}},
			RPCPORT: string("9980/tcp"),
			IMAGE:   string("openwallet/sc:1.3.3"),
		},
		// "iota": &FullnodeContainerConfig{
		// 	CMD:     [2][]string{{"/bin/bash", "-c", "while sleep 1; do date; done"}, {}},
		// 	PORT:    [][3]string{{"18265/tcp", "10051", "20051"}},
		// 	RPCPORT: string("18265/tcp"),
		// 	IMAGE:   string("openwallet/iota:latest"),
		// },
		"bopo": &FullnodeContainerConfig{
			WORKPATH: "/usr/local/paicode",
			CMD:      [2][]string{{"/bin/bash", "-c", "cd /usr/local/paicode; ./gamepaicore --listen 0.0.0.0:7280 >> /openwallet/data/run.log 2>&1"}, {}},
			PORT:     [][3]string{{"7280/tcp", "17280", "27280"}},
			RPCPORT:  string("7280/tcp"),
			IMAGE:    string("openwallet/bopo:latest"),
		},
		"hc": &FullnodeContainerConfig{
			WORKPATH: "/usr/local/paicode",
			CMD: [2][]string{
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/data --logdir=/openwallet/data --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 && /usr/local/hypercash/bin/hcwallet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"},
				{"/bin/bash", "-c", "/usr/local/hypercash/bin/hcd --datadir=/openwallet/testdata --logdir=/openwallet/testdata --rpcuser=wallet --rpcpass=walletPassword2017 --txindex --rpclisten=0.0.0.0:14009 --testnet && /usr/local/hypercash/bin/hcwallet --testnet --rpcconnect=127.0.0.1:14009 --username=wallet --password=walletPassword2017 --rpclisten=0.0.0.0:12010"}},
			PORT:    [][3]string{{"12010/tcp", "12010", "22010"}, {"14009/tcp", "14009", "24009"}},
			RPCPORT: string("12010/tcp"),
			IMAGE:   string("openwallet/hc:2.0.3dev"),
		},
		"qtum": &FullnodeContainerConfig{
			WORKPATH: "/data",
			CMD: [2][]string{
				{"qtumd", "--datadir=/data", "--logdir=/data", "--rpcuser=wallet", "--rpcpass=walletPassword2017", "--rpclisten=127.0.0.1:14009"},
				{"qtumd", "--datadir=/data", "--logdir=/data", "--rpcuser=wallet", "--rpcpass=walletPassword2017", "--rpclisten=127.0.0.1:14009", "--testnet"}},
			PORT:    [][3]string{{"9360/tcp", "12010", "22010"}, {"14009/tcp", "14009", "24009"}},
			RPCPORT: string("9360/tcp"),
			IMAGE:   string("openw/qtum:0.15.3"),
		},
		"ltc": &FullnodeContainerConfig{ // litecoin
			WORKPATH: "/data",
			CMD:      [2][]string{},
			PORT:     [][3]string{{"9360/tcp", "10004", "20004"}},
			RPCPORT:  string("9360/tcp"),
			IMAGE:    string("openw/litecoin:0.16.0"),
		},
	}
}
