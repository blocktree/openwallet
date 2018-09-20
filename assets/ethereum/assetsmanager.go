package ethereum

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
)

//loadConfig 读取配置
func (this *WalletManager) LoadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置

	//fmt.Println("config file path:", this.Config.configFilePath)
	//fmt.Println("config file name:", this.Config.configFileName)
	absFile := filepath.Join(this.Config.configFilePath, this.Config.configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	this.Config.ServerAPI = c.String("serverAPI")
	this.Config.Threshold, _ = decimal.NewFromString(c.String("threshold"))
	this.Config.SumAddress = c.String("sumAddress")
	this.Config.RpcUser = c.String("rpcUser")
	this.Config.RpcPassword = c.String("rpcPassword")
	this.Config.NodeInstallPath = c.String("nodeInstallPath")
	this.Config.IsTestNet, _ = c.Bool("isTestNet")
	if this.Config.IsTestNet {
		this.Config.WalletDataPath = c.String("testNetDataPath")
	} else {
		this.Config.WalletDataPath = c.String("mainNetDataPath")
	}

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}

	this.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)

	//token := BasicAuth(wm.Config.RpcUser, wm.Config.RpcPassword)

	//wm.WalletClient = NewClient(wm.Config.ServerAPI, token, false)

	return nil
}

//GetAddressDecode 地址解析器
func (this *WalletManager) GetAddressDecode() openwallet.AddressDecoder {
	return this.Decoder
}

//GetTransactionDecoder 交易单解析器
func (this *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return this.TxDecoder
}

//GetBlockScanner 获取区块链
func (this *WalletManager) GetBlockScanner() openwallet.BlockScanner {
	//先加载是否有配置文件
	err := this.LoadConfig()
	if err != nil {
		return nil
	}

	return nil //this.Blockscanner.
}

//ImportWatchOnlyAddress 导入观测地址
func (this *WalletManager) ImportWatchOnlyAddress(address ...*openwallet.Address) error {
	return nil
}

//GetAddressWithBalance 获取多个地址余额，使用查账户和单地址
func (this *WalletManager) GetAddressWithBalance(addresses ...*openwallet.Address) error {
	return nil
}

//CurveType 曲线类型
func (this *WalletManager) CurveType() uint32 {
	return this.Config.CurveType
}

//FullName 币种全名
func (this *WalletManager) FullName() string {
	return "Ethereum"
}

//Symbol 币种标识
func (this *WalletManager) Symbol() string {
	return this.Config.Symbol
}

//小数位精度
func (this *WalletManager) Decimal() int32 {
	return 18
}
