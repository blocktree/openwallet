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
package ethereum

import (
	"io/ioutil"
	"math/big"
	"path/filepath"
	"time"

	//	"github.com/astaxie/beego/config"
	"encoding/json"

	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/log"
	owcrypt "github.com/blocktree/go-OWCrypt"
)

const (
	//	BLOCK_CHAIN_DB     = "blockchain.db"
	BLOCK_CHAIN_BUCKET = "blockchain"
	ERC20TOKEN_DB      = "erc20Token.db"
)

const (
	Symbol       = "ETH"
	MasterKey    = "Ethereum seed"
	TIME_POSTFIX = "20060102150405"
	CurveType    = owcrypt.ECC_CURVE_SECP256K1

//	CHAIN_ID     = 922337203685 //12
)

type WalletConfig struct {

	//币种
	Symbol    string
	MasterKey string
	RootDir   string
	//RPC认证账户名
	//RpcUser string
	//RPC认证账户密码
	//RpcPassword string
	//证书目录
	//CertsDir string
	//钥匙备份路径
	KeyDir string
	//地址导出路径
	AddressDir string
	//配置文件路径
	ConfigFilePath string
	//配置文件名
	ConfigFileName string
	//rpc证书
	//CertFileName string
	//区块链数据文件
	BlockchainFile string
	//是否测试网络
	IsTestNet bool
	// 核心钱包是否只做监听
	//CoreWalletWatchOnly bool
	//最大的输入数量
	//MaxTxInputs int
	//本地数据库文件路径
	DbPath string
	//备份路径
	BackupDir string
	//钱包服务API
	ServerAPI string
	//钱包安装的路径
	//NodeInstallPath string
	//钱包数据文件目录
	//WalletDataPath string
	//汇总阀值
	ThreaholdStr string
	Threshold    *big.Int `json:"-"`
	//汇总地址
	SumAddress string
	//汇总执行间隔时间
	CycleSeconds uint64 //time.Duration
	//默认配置内容
	//	DefaultConfig string
	//曲线类型
	CurveType uint32
	//小数位长度
	//	CoinDecimal decimal.Decimal `json:"-"`
	EthereumKeyPath string
	//是否完全依靠本地维护nonce
	LocalNonce bool
	ChainID    uint64
}

func (this *WalletConfig) LoadConfig2() (*WalletConfig, error) {
	return this.LoadConfig(this.ConfigFilePath, this.ConfigFileName, nil)
}

func makeEthDefaultConfig(rootDir string) *WalletConfig {
	conf := &WalletConfig{}
	conf.Symbol = Symbol
	conf.MasterKey = MasterKey
	conf.CurveType = CurveType
	conf.RootDir = rootDir
	//钥匙备份路径
	conf.KeyDir = filepath.Join(rootDir, "eth", "key")
	//地址导出路径
	conf.AddressDir = filepath.Join(rootDir, "eth", "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	conf.ConfigFilePath = filepath.Join(rootDir, "eth", "conf") //filepath.Join("conf")
	//配置文件名
	conf.ConfigFileName = "eth.json"
	//区块链数据文件
	conf.BlockchainFile = "blockchain.db"
	//是否测试网络
	conf.IsTestNet = true

	//本地数据库文件路径
	conf.DbPath = filepath.Join(rootDir, "eth", "db")
	//备份路径
	conf.BackupDir = filepath.Join(rootDir, "eth", "backup")
	//钱包服务API
	conf.ServerAPI = "http://127.0.0.1:8545"

	conf.Threshold = big.NewInt(5) //decimal.NewFromFloat(5)
	conf.ThreaholdStr = "5"
	//汇总地址
	conf.SumAddress = ""
	//汇总执行间隔时间
	conf.CycleSeconds = 10
	//	this.ChainId = 12
	conf.EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	//每次都向节点查询nonce
	conf.LocalNonce = false
	//区块链ID
	conf.ChainID = 12
	return conf
}

func (this *WalletConfig) LoadConfig(configFilePath string, configFileName string, defaultConf *WalletConfig) (*WalletConfig, error) {
	absFile := filepath.Join(configFilePath, configFileName)
	dat, err := ioutil.ReadFile(absFile)
	if err != nil && defaultConf == nil {
		log.Error("read config file[", configFilePath, "/", configFileName, "] failed, err=", err)
		return nil, err
	} else if err != nil {
		log.Info("cannot find the config file[", configFilePath, "/", configFileName, "] , set config to default. ")
		*this = *defaultConf

		//创建目录
		file.MkdirAll(this.RootDir)
		file.MkdirAll(configFilePath)
		file.MkdirAll(this.DbPath)
		file.MkdirAll(this.BackupDir)
		file.MkdirAll(this.KeyDir)

		configStr, _ := json.MarshalIndent(this, "", " ")
		err = ioutil.WriteFile(absFile, configStr, 0644)
		if err != nil {
			log.Error("write to config file failed, err=", err)
			return nil, err
		}

		return this, nil
	}

	err = json.Unmarshal(dat, this)
	if err != nil {
		log.Error("decode config fail from json failed, err=", err)
		return nil, err
	}

	this.Threshold, err = ConvertEthStringToWei(this.ThreaholdStr)
	if err != nil {
		log.Error("convert threshold to big.int failed, err=%v", err)
		return nil, err
	}

	configStr, _ := json.MarshalIndent(this, "", " ")
	log.Debug("load ", this.Symbol, " config :", string(configStr))
	return nil, nil
}

/*func NewConfig(rootDir string, symbol string, masterKey string) *WalletConfig {
	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.MasterKey = masterKey
	c.CurveType = CurveType

	//RPC认证账户名
	//c.RpcUser = ""
	//RPC认证账户密码
	//c.RpcPassword = ""
	//证书目录
	//c.CertsDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "certs")
	//钥匙备份路径
	c.KeyDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "key")
	//地址导出路径
	c.AddressDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.ConfigFilePath = filepath.Join("conf")
	//配置文件名
	c.ConfigFileName = c.Symbol + ".ini"
	//rpc证书
	//c.CertFileName = "rpc.cert"
	//区块链数据文件
	c.BlockchainFile = "blockchain.db"
	//是否测试网络
	c.IsTestNet = true
	// 核心钱包是否只做监听
	//c.CoreWalletWatchOnly = true
	//最大的输入数量
	//c.MaxTxInputs = 50
	//本地数据库文件路径
	c.DbPath = filepath.Join(rootDir, strings.ToLower(c.Symbol), "db")
	//备份路径
	c.BackupDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "backup")
	//钱包服务API
	c.ServerAPI = "http://127.0.0.1:8545"
	//钱包安装的路径
	//c.NodeInstallPath = ""
	//钱包数据文件目录
	//c.WalletDataPath = ""
	//汇总阀值
	c.Threshold = decimal.NewFromFloat(5)
	//汇总地址
	c.SumAddress = ""
	//汇总执行间隔时间
	c.CycleSeconds = time.Second * 10
	//小数位长度
	c.CoinDecimal = decimal.NewFromFloat(100000000)

	//默认配置内容
	c.DefaultConfig = `
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
# RPC api url
serverAPI = ""
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

	//创建目录
	file.MkdirAll(c.DbPath)
	file.MkdirAll(c.BackupDir)
	file.MkdirAll(c.KeyDir)

	return &c
}*/

var (
	//	dataDir = filepath.Join("data", strings.ToLower(Symbol))
	//钥匙备份路径
	//	KeyDir = filepath.Join(dataDir, "key")
	//地址导出路径
	//	AddressDir = filepath.Join(dataDir, "address")
	//配置文件路径
	//ConfigFilePath = filepath.Join("conf")
	//配置文件名
	//	ConfigFileName = Symbol + ".ini"
	//备份路径
	//	BackupDir = filepath.Join(dataDir, "backup")
	//钱包数据文件目录
	//	walletDataPath = ""
	//本地数据库文件路径
	//DbPath = filepath.Join(dataDir, "db")
	//钱包安装的路径
	//nodeInstallPath = ""
	//参与汇总的钱包
	//	walletsInSum = make(map[string]*Wallet)
	//钱包服务API
	//serverAPI = "http://127.0.0.1:8545"
	//汇总阀值 ???--这个要设置成什么
	//threshold = big.NewInt(0) //decimal.Decimal = decimal.NewFromFloat(5)
	//汇总地址
	//	sumAddress = "0x2a63b2203955b84fefe52baca3881b3614991b34"
	//是否测试网络
	//	isTestNet = true
	//网络号
	//	chainID int64 = 12
	//ethereum key存放路径
	//	EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	//钱包下keystore默认密码
	//DefaultPasswordForEthKey = "yjHFlngdBDl12"
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

func loadConfig() error {
	return nil
}

/*func loadConfig_() error {
	var c config.Configer
	var err error

	//读取配置
	absFile := filepath.Join(ConfigFilePath, ConfigFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	threshold, _ = threshold.SetString(c.String("threshold"), 10) //decimal.NewFromString(c.String("threshold"))
	sumAddress = c.String("sumAddress")
	isTestNet, _ = c.Bool("isTestNet")
	//	if isTestNet {
	//		walletDataPath = c.String("testNetDataPath")
	//	} else {
	//		walletDataPath = c.String("mainNetDataPath")
	//	}

	client = &Client{
		BaseURL: serverAPI,
		Debug:   false,
	}
	return nil
}*/

/*func newConfigFile(
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

	//filepath.Join()

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
	file.MkdirAll(ConfigFilePath)
	absFile := filepath.Join(ConfigFilePath, ConfigFileName)
	err = c.SaveConfigFile(absFile)
	if err != nil {
		return nil, "", err
	}

	return c, absFile, nil
}*/

//initConfig 初始化配置文件
/*func initConfig() {

	//读取配置
	absFile := filepath.Join(ConfigFilePath, ConfigFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(ConfigFilePath)
		file.WriteFile(absFile, []byte(defaultConfig), false)
	}

}*/

/*func printConfig() error {
	initConfig()
	//读取配置
	absFile := filepath.Join(ConfigFilePath, ConfigFileName)
	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")
	return nil

}*/
