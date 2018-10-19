package ethereum

import (
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

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
	err := this.loadConfig()
	if err != nil {
		log.Errorf("load config failed, err=%v", err)
		return nil
	}

	return this.Blockscanner
}

//ImportWatchOnlyAddress 导入观测地址
func (this *WalletManager) ImportWatchOnlyAddress(address ...*openwallet.Address) error {
	return nil
}

//GetAddressWithBalance 获取多个地址余额，使用查账户和单地址
func (this *WalletManager) GetAddressWithBalance(addresses ...*openwallet.Address) error {
	for _, addr := range addresses {
		log.Debugf("wallet[%v] address[%v]:", addr.AccountID, addr.Address)
		amount, err := this.WalletClient.GetAddrBalance("0x" + addr.Address)
		if err != nil {
			log.Error("get address[", addr.Address, "] balance failed, err=", err)
			return err
		}

		dm, err := ConverWeiStringToEthDecimal(amount.String())
		if err != nil {
			log.Error("ConverWeiStringToEthDecimal amount[", amount.String(), "] failed, err=", err)
			return err
		}

		addr.Balance = dm.String()
	}

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

//SymbolID 币种标识
func (this *WalletManager) Symbol() string {
	return this.Config.Symbol
}

//小数位精度
func (this *WalletManager) Decimal() int32 {
	return 18
}
