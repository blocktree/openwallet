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
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"

	//	"github.com/astaxie/beego/config"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/keystore"
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
	//ThreaholdStr string
	Threshold *big.Int `json:"-"`
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

func makeEthDefaultConfig(ConfigFilePath string) string {
	/*conf := &WalletConfig{}
	conf.SymbolID = SymbolID
	conf.MasterKey = MasterKey
	conf.CurveType = CurveType
	conf.RootDir = rootDir
	//钥匙备份路径
	conf.KeyDir = filepath.Join(rootDir, "eth", "key")
	//地址导出路径
	conf.AddressDir = filepath.Join(rootDir, "eth", "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(SymbolID), "blockchain")
	//配置文件路径
	conf.ConfigFilePath = ConfigFilePath //filepath.Join(rootDir, "eth", "conf") //filepath.Join("conf")
	//配置文件名
	conf.ConfigFileName = "eth.ini"
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

	//汇总地址
	conf.SumAddress = ""
	//汇总执行间隔时间
	conf.CycleSeconds = 10
	//
	conf.EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	//每次都向节点查询nonce
	conf.LocalNonce = false
	//区块链ID
	conf.ChainID = 12*/

	defaultConfigStr := `
SymbolID = "ETH"
MasterKey = "Ethereum seed"
CurveType = %v
#key file path
KeyDir = "data/eth/key"
#key export path
AddressDir = "data/eth/address"
#config file name
ConfigFileName = "eth.ini"
#block chain db name
BlockchainFile = "blockchain.db"
#check if it's test net
IsTestNet = true
#db file path
DbPath = "data/eth/db" 
#wallet backup path
BackupDir = "data/eth/backup" 
#wallet api url
ServerAPI = "http://127.0.0.1:8545"
#wallet summary threshold
Threshold = 5 
#summary address
SumAddress = ""
#summary time interval
CycleSeconds = 10
#eth node default key store path	
EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
#whether check txpool to find nonce
LocalNonce = false
#block chain ID
ChainID = 12
`
	return fmt.Sprintf(defaultConfigStr, CurveType)
}

func (this *WalletConfig) LoadConfig(configFilePath string, configFileName string) (*WalletConfig, error) {
	absFile := filepath.Join(configFilePath, configFileName)
	fmt.Println("config path:", absFile)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil, errors.New("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	this.Symbol = c.String("SymbolID")     //SymbolID
	this.MasterKey = c.String("MasterKey") //MasterKey
	curveType, err := c.Int64("CurveType") //CurveType
	if err != nil {
		log.Error("curve type failed, err=", err)
		return nil, err
	}

	this.CurveType = uint32(curveType) //CurveType
	this.RootDir = c.String("RootDir") //rootDir
	//钥匙备份路径
	this.KeyDir = c.String("KeyDir") //filepath.Join(rootDir, "eth", "key")
	//地址导出路径
	this.AddressDir = c.String("AddressDir") //filepath.Join(rootDir, "eth", "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(SymbolID), "blockchain")
	//配置文件路径
	this.ConfigFilePath = c.String("ConfigFilePath") //ConfigFilePath //filepath.Join(rootDir, "eth", "conf") //filepath.Join("conf")
	//配置文件名
	this.ConfigFileName = c.String("ConfigFileName") //"eth.ini"
	//区块链数据文件
	this.BlockchainFile = c.String("BlockchainFile") //"blockchain.db"
	//是否测试网络
	isTestNet, err := c.Bool("isTestNet")
	if err != nil {
		log.Error("isTestNet error, err=", err)
		return nil, err
	}
	this.IsTestNet = isTestNet //true

	//本地数据库文件路径
	this.DbPath = c.String("DbPath") //filepath.Join(rootDir, "eth", "db")
	//备份路径
	this.BackupDir = c.String("BackupDir") //filepath.Join(rootDir, "eth", "backup")
	//钱包服务API
	this.ServerAPI = c.String("ServerAPI") //"http://127.0.0.1:8545"

	threshold, err := c.Int64("Threshold")
	if err != nil {
		log.Error("Threshold error, err=", err)
		return nil, err
	}
	this.Threshold = big.NewInt(threshold) //decimal.NewFromFloat(5)
	//this.ThreaholdStr = "5"
	//汇总地址
	this.SumAddress = c.String("SumAddress") //""
	//汇总执行间隔时间
	cycleSeconds, err := c.Int64("CycleSeconds")
	if err != nil {
		log.Error("CycleSeconds error, err=", err)
		return nil, err
	}
	this.CycleSeconds = uint64(cycleSeconds) //c.Int64("CycleSeconds")
	//	this.ChainId = 12
	this.EthereumKeyPath = c.String("EthereumKeyPath") //"/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	//每次都向节点查询nonce
	localnonce, err := c.Bool("LocalNonce")
	if err != nil {
		log.Error("LocalNonce error, err=", err)
		return nil, err
	}
	this.LocalNonce = localnonce //c.Bool("LocalNonce")
	//区块链ID
	chainId, err := c.Int64("ChainID")
	if err != nil {
		log.Error("ChainID error, err=", err)
		return nil, err
	}
	this.ChainID = uint64(chainId) //c.Int64("ChainID") //12
	return this, nil
}

func (this *WalletManager) NewConfig(rootDir string, masterKey string) *WalletConfig {
	c := WalletConfig{}

	//币种
	c.Symbol = this.SymbolID
	c.MasterKey = masterKey
	c.CurveType = CurveType

	//RPC认证账户名
	//c.RpcUser = ""
	//RPC认证账户密码
	//c.RpcPassword = ""
	//证书目录
	//c.CertsDir = filepath.Join(rootDir, strings.ToLower(c.SymbolID), "certs")
	//钥匙备份路径
	c.RootDir = rootDir
	c.KeyDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "key")
	//地址导出路径
	c.AddressDir = filepath.Join(rootDir, strings.ToLower(c.Symbol), "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(SymbolID), "blockchain")
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
	c.Threshold = big.NewInt(5) //decimal.NewFromFloat(5)
	//汇总地址
	c.SumAddress = ""
	//汇总执行间隔时间
	c.CycleSeconds = 10
	//小数位长度
	//c.CoinDecimal = decimal.NewFromFloat(100000000)
	c.EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	this.Config = &c
	//创建目录
	file.MkdirAll(c.DbPath)
	file.MkdirAll(c.BackupDir)
	file.MkdirAll(c.KeyDir)

	return &c
}

func (this *WalletManager) loadConfig() error {
	if this.Config == nil {
		this.Config = &WalletConfig{}
	}
	log.Debug("symbol:", this.SymbolID+".ini")
	_, err := this.Config.LoadConfig(this.ConfigPath, this.SymbolID+".ini")
	if err != nil {
		log.Error(err)
		return err
	}
	this.StorageOld = keystore.NewHDKeystore(this.Config.KeyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	storage := hdkeystore.NewHDKeystore(this.Config.KeyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	this.Storage = storage
	client := &Client{BaseURL: this.Config.ServerAPI, Debug: false}
	this.WalletClient = client
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

func (this *WalletManager) PrintConfig() error {
	this.InitConfig()
	//读取配置
	absFile := filepath.Join(this.ConfigPath, this.SymbolID+".ini")
	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")
	return nil

}

//initConfig 初始化配置文件
func (this *WalletManager) InitConfig() {

	//读取配置
	absFile := filepath.Join(this.ConfigPath, this.SymbolID+".ini")
	if !file.Exists(absFile) {
		file.MkdirAll(this.ConfigPath)
		file.WriteFile(absFile, []byte(this.DefaultConfig), false)
	}
}

//初始化配置流程
func (this *WalletManager) InitConfigFlow() error {

	this.InitConfig()
	file := filepath.Join(this.ConfigPath, this.SymbolID+".ini")
	fmt.Printf("You can run 'vim %s' to edit wallet's Config.\n", file)
	return nil
}

//查看配置信息
func (wm *WalletManager) ShowConfig() error {
	return wm.PrintConfig()
}
