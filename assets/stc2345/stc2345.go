package stc2345

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/blocktree/OpenWallet/log"
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

func makeStc2345DefaultConfig(rootDir string) *ethereum.WalletConfig {
	conf := &ethereum.WalletConfig{}
	conf.Symbol = Symbol
	conf.MasterKey = MasterKey
	conf.CurveType = CurveType
	conf.RootDir = rootDir
	//钥匙备份路径
	conf.KeyDir = filepath.Join(rootDir, "stc2345", "key")
	//地址导出路径
	conf.AddressDir = filepath.Join(rootDir, "stc2345", "address")
	//区块链数据
	//blockchainDir = filepath.Join(rootDir, strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	conf.ConfigFilePath = filepath.Join(rootDir, "stc2345", "conf") //filepath.Join("conf")
	//配置文件名
	conf.ConfigFileName = "stc2345.json"
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
	conf.ChainID = 922337203685
	//conf.EthereumKeyPath = "/Users/peter/workspace/bitcoin/wallet/src/github.com/ethereum/go-ethereum/chain/keystore"
	return conf
}

func NewWalletManager(rootDir string) *WalletManager {
	wm := WalletManager{}
	//wm.RootDir = rootDir

	configPath := filepath.Join(rootDir, "stc2345", "conf")
	wm.Config = &ethereum.WalletConfig{}
	_, err := wm.Config.LoadConfig(configPath, "stc2345.json", makeStc2345DefaultConfig(rootDir))
	if err != nil {
		log.Error("wm.Config.LoadConfig failed, err=", err)
		os.Exit(-1)
	}
	storage := hdkeystore.NewHDKeystore(wm.Config.KeyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = ethereum.NewETHBlockScanner(&wm.WalletManager)
	wm.Decoder = &ethereum.AddressDecoder{}
	wm.TxDecoder = ethereum.NewTransactionDecoder(&wm.WalletManager)

	wm.StorageOld = keystore.NewHDKeystore(wm.Config.KeyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	wm.WalletInSumOld = make(map[string]*ethereum.Wallet)

	client := &ethereum.Client{BaseURL: wm.Config.ServerAPI, Debug: false}
	wm.WalletClient = client
	//	g_manager = &wm
	return &wm
}
