package ethereum

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	//	"github.com/astaxie/beego/config"
	"encoding/json"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/shopspring/decimal"
)

const (
	Symbol    = "ETH"
	MasterKey = "Ethereum seed"
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
	//备份路径
	backupDir = filepath.Join(dataDir, "backup")
	//钱包数据文件目录
	walletDataPath = ""
	//本地数据库文件路径
	dbPath = filepath.Join(dataDir, "db")
	//钱包安装的路径
	nodeInstallPath = ""
	//参与汇总的钱包
	walletsInSum = make(map[string]*Wallet)
	//钱包服务API
	serverAPI = "http://127.0.0.1:8545"
	//汇总阀值 ???--这个要设置成什么
	threshold decimal.Decimal = decimal.NewFromFloat(5)
	//汇总地址
	sumAddress = ""
	//是否测试网络
	isTestNet = true
	//ethereum key存放路径
	EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	//钱包下keystore默认密码
	DefaultPasswordForEthKey = "yjHFlngdBDl12"
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

func loadConfig_() error {
	var c config.Configer
	var err error

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	sumAddress = c.String("sumAddress")
	isTestNet, _ = c.Bool("isTestNet")
	if isTestNet {
		walletDataPath = c.String("testNetDataPath")
	} else {
		walletDataPath = c.String("mainNetDataPath")
	}

	client = &Client{
		BaseURL: serverAPI,
		Debug:   false,
	}
	return nil
}

func newConfigFile(
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
	file.MkdirAll(configFilePath)
	absFile := filepath.Join(configFilePath, configFileName)
	err = c.SaveConfigFile(absFile)
	if err != nil {
		return nil, "", err
	}

	return c, absFile, nil
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

func printConfig() error {
	initConfig()
	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")
	return nil

}
