package stc2345

import (
	"fmt"

	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/openwallet"
	owcrypt "github.com/blocktree/go-OWCrypt"
)

const (
	Symbol       = "STC2345"
	MasterKey    = "STC2345 seed"
	TIME_POSTFIX = "20060102150405"
	CurveType    = owcrypt.ECC_CURVE_SECP256K1
)

type WalletManager struct {
	ethereum.WalletManager
}

func makeStc2345DefaultConfig() string {
	/*conf := &ethereum.WalletConfig{}
	conf.SymbolID = SymbolID
	conf.MasterKey = MasterKey
	conf.CurveType = CurveType
	conf.RootDir = rootDir
	//钥匙备份路径
	conf.KeyDir = filepath.Join(rootDir, "stc2345", "key")
	//地址导出路径
	conf.AddressDir = filepath.Join(rootDir, "stc2345", "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(SymbolID), "blockchain")
	//配置文件路径
	conf.ConfigFilePath = configFilePath //filepath.Join(rootDir, "stc2345", "conf") //filepath.Join("conf")
	//配置文件名
	conf.ConfigFileName = "stc2345.ini"
	//区块链数据文件
	conf.BlockchainFile = "blockchain.db"
	//是否测试网络
	conf.IsTestNet = true

	//本地数据库文件路径
	conf.DbPath = filepath.Join(rootDir, "stc2345", "db")
	//备份路径
	conf.BackupDir = filepath.Join(rootDir, "stc2345", "backup")
	//钱包服务API
	conf.ServerAPI = "https://jfx.alliancechain.net"

	conf.Threshold = big.NewInt(5) //decimal.NewFromFloat(5)
	conf.ThreaholdStr = "5"
	//汇总地址
	conf.SumAddress = ""
	//汇总执行间隔时间
	conf.CycleSeconds = 10
	//本地维护nonce
	conf.LocalNonce = true
	conf.ChainID = 922337203685*/
	defaultConfigStr := `
SymbolID = "ETH"
MasterKey = "Ethereum seed"
CurveType = %v
RootDir = "data"
#key file path
KeyDir = "data/stc2345/key"
#key export path
AddressDir = "data/stc2345/address"
#config file path
ConfigFilePath = "conf"
#config file name
ConfigFileName = "stc2345.ini"
#block chain db name
BlockchainFile = "blockchain.db"
#check if it's test net
IsTestNet = true
#db file path
DbPath = "data/stc2345/db" 
#wallet backup path
BackupDir = "data/stc2345/backup" 
#wallet api url
ServerAPI = "https://jfx.alliancechain.net"
#wallet summary threshold
Threshold = 5 
#summary address
SumAddress = ""
#summary time interval
CycleSeconds = 10
#eth node default key store path	
EthereumKeyPath = ""
#whether check txpool to find nonce
LocalNonce = true
#block chain ID
ChainID = 922337203685
`
	return fmt.Sprintf(defaultConfigStr, CurveType)
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	//wm.RootDir = rootDir
	//configPath := filepath.Join(rootDir, "stc2345", "conf")
	wm.RootPath = "data"
	wm.ConfigPath = "conf"
	wm.SymbolID = Symbol

	wm.DefaultConfig = makeStc2345DefaultConfig()
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = ethereum.NewETHBlockScanner(&wm.WalletManager)
	wm.Decoder = &ethereum.AddressDecoder{}
	wm.TxDecoder = ethereum.NewTransactionDecoder(&wm.WalletManager)
	//	g_manager = &wm
	return &wm
}
