package nebulasio

import (
	_ "github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

//ImportWatchOnlyAddress 导入观测地址
func (wm *WalletManager) ImportWatchOnlyAddress(address ...*openwallet.Address) error {
	return nil
}


//GetAddressWithBalance 获取多个地址余额，使用查账户和单地址
func (wm *WalletManager) GetAddressWithBalance(address ...*openwallet.Address) error {
	for _,addr := range address{
		//fmt.Printf("%v's balance=%v\n",addr.Address,addr.Balance)
		balance,err := wm.WalletClient.CallGetaccountstate(addr.Address,"balance")
		if err != nil{
			return err
		}

		NAS_decimal ,err := ConverWeiStringToNasDecimal(balance)
		if err != nil {
			return err
		}

		addr.Balance = NAS_decimal.String()
	}

	return nil
}


/*实现 openwallet/assets.go 中AssetsAdapter接口中的SymbolInfo方法*/
//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}
//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "Nebulasio"
}
//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}
//小数位精度
func (wm *WalletManager) Decimal() int32 {
	return 18
}

/*实现 openwallet/assets.go 中AssetsAdapter接口中的GetAddressDecode方法*/
//AddressDecode 地址解析器
func (wm *WalletManager) GetAddressDecode() openwallet.AddressDecoder {
	return wm.Decoder
}

/*实现 openwallet/assets.go 中AssetsAdapter接口中的GetTransactionDecoder方法*/
//TransactionDecoder 交易单解析器
func (wm *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return wm.TxDecoder
}

/*实现 openwallet/assets.go 中AssetsAdapter接口中的GetBlockScanner方法*/
//GetBlockScanner 获取区块链
func (wm *WalletManager) GetBlockScanner() openwallet.BlockScanner {

	//先加载是否有配置文件
	//err := wm.LoadConfig()
	//if err != nil {
	//	return nil
	//}
	return wm.Blockscanner
}


//实现 openwallet/assets.go 中AssetsAdapter接口中的GetSmartContractDecoder方法*/
func (wm *WalletManager) GetSmartContractDecoder() openwallet.SmartContractDecoder {
	return wm.ContractDecoder
}

