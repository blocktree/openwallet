package ontology

import (
	"github.com/shopspring/decimal"
	"time"
	"path/filepath"
	"strings"
	"github.com/blocktree/OpenWallet/common/file"
	"fmt"
)

const (
	//币种
	Symbol    = "ONT"
	MasterKey = "Ontology seed"
	//CurveType = owcrypt.ECC_CURVE_SECP256K1
)

type WalletConfig struct {

	//币种
	Symbol    string
	MasterKey string

	//RPC认证账户名
	RpcUser string
	//RPC认证账户密码
	RpcPassword string
	//证书目录
	CertsDir string
	//钥匙备份路径
	keyDir string
	//地址导出路径
	addressDir string
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
	//rpc证书
	CertFileName string
	//区块链数据文件
	BlockchainFile string
	//是否测试网络
	IsTestNet bool
	// 核心钱包是否只做监听
	CoreWalletWatchOnly bool
	//最大的输入数量
	MaxTxInputs int
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
	//小数位长度
	CoinDecimal decimal.Decimal
	//核心钱包密码，配置有值用于自动解锁钱包
	WalletPassword string
	//后台数据源类型
	RPCServerType int
	//s是否支持隔离验证
	SupportSegWit bool
}

func NewConfig(symbol string, masterKey string) *WalletConfig {

	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.MasterKey = masterKey
	//c.CurveType = CurveType

	//RPC认证账户名
	c.RpcUser = ""
	//RPC认证账户密码
	c.RpcPassword = ""
	//证书目录
	c.CertsDir = filepath.Join("data", strings.ToLower(c.Symbol), "certs")
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
	//rpc证书
	c.CertFileName = "rpc.cert"
	//区块链数据文件
	c.BlockchainFile = "blockchain.db"
	//是否测试网络
	c.IsTestNet = true
	// 核心钱包是否只做监听
	c.CoreWalletWatchOnly = true
	//最大的输入数量
	c.MaxTxInputs = 50
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.Symbol), "db")
	//备份路径
	c.backupDir = filepath.Join("data", strings.ToLower(c.Symbol), "backup")
	//钱包服务API
	c.ServerAPI = "http://127.0.0.1:10000"
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
	//小数位长度
	c.CoinDecimal = decimal.NewFromFloat(100000000)
	//核心钱包密码，配置有值用于自动解锁钱包
	c.WalletPassword = ""
	//后台数据源类型
	c.RPCServerType = 0
	//支持隔离见证
	c.SupportSegWit = true

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
# RPC Server Type，0: CoreWallet RPC; 1: Explorer API
rpcServerType = 0
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
# walletPassword use to unlock bitcoin core wallet
walletPassword = ""
`

	//创建目录
	file.MkdirAll(c.dbPath)
	file.MkdirAll(c.backupDir)
	file.MkdirAll(c.keyDir)

	return &c
}

//printConfig Print config information
func (wc *WalletConfig) PrintConfig() error {

	wc.InitConfig()
	//读取配置
	absFile := filepath.Join(wc.configFilePath, wc.configFileName)
	//apiURL := c.String("apiURL")
	//walletPath := c.String("walletPath")
	//threshold := c.String("threshold")
	//minSendAmount := c.String("minSendAmount")
	//minFees := c.String("minFees")
	//sumAddress := c.String("sumAddress")
	//isTestNet, _ := c.Bool("isTestNet")

	fmt.Printf("-----------------------------------------------------------\n")
	//fmt.Printf("Wallet API URL: %s\n", apiURL)
	//fmt.Printf("Wallet Data FilePath: %s\n", walletPath)
	//fmt.Printf("Summary Address: %s\n", sumAddress)
	//fmt.Printf("Summary Threshold: %s\n", threshold)
	//fmt.Printf("Min Send Amount: %s\n", minSendAmount)
	//fmt.Printf("Transfer Fees: %s\n", minFees)
	//if isTestNet {
	//	fmt.Printf("Network: TestNet\n")
	//} else {
	//	fmt.Printf("Network: MainNet\n")
	//}
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