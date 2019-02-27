package virtualeconomy

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
)

type AddrBalance struct {
	Address string
	Balance *big.Int
	index   int
}

func convertFlostStringToBigInt(amount string) (*big.Int, error) {
	vDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert from string to decimal failed, err=", err)
		return nil, err
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < 9; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}
	d, _ := decimal.NewFromString(decimalInt.String())
	vDecimal = vDecimal.Mul(d)
	rst := new(big.Int)
	if _, valid := rst.SetString(vDecimal.String(), 10); !valid {
		log.Error("conver to big.int failed")
		return nil, errors.New("conver to big.int failed")
	}
	return rst, nil
}

func convertBigIntToFloatDecimal(amount string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert string to deciaml failed, err=", err)
		return d, err
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < 9; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}

	w, _ := decimal.NewFromString(decimalInt.String())
	d = d.Div(w)
	return d, nil
}

func convertIntStringToBigInt(amount string) (*big.Int, error) {
	vInt64, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		log.Error("convert from string to int failed, err=", err)
		return nil, err
	}

	return big.NewInt(vInt64), nil
}

type ContractDecoder struct {
	wm *WalletManager
}

//NewContractDecoder 智能合约解析器
func NewContractDecoder(wm *WalletManager) *ContractDecoder {
	decoder := ContractDecoder{}
	decoder.wm = wm
	return &decoder
}

func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {

	// var tokenBalanceList []*openwallet.TokenBalance

	// for i := 0; i < len(address); i++ {
	// 	tokenBalance := openwallet.TokenBalance{
	// 		Contract: &contract,
	// 	}
	// 	if contract.ContractID == ontologyTransaction.ONTContractAddress {
	// 		balance, err := decoder.wm.RPCClient.getONTBalance(address[i])
	// 		if err != nil {
	// 			log.Error("Get ONT balance of address [%v] failed with error : [%v]", address[i], err)
	// 			return nil, err
	// 		}
	// 		tokenBalance.Balance = &openwallet.Balance{
	// 			Address:          address[i],
	// 			Symbol:           contract.Symbol,
	// 			Balance:          balance.ONTBalance.String(),
	// 			ConfirmBalance:   balance.ONTBalance.String(),
	// 			UnconfirmBalance: "0",
	// 		}
	// 	} else if contract.ContractID == ontologyTransaction.ONGContractAddress {
	// 		balance, err := decoder.wm.RPCClient.getONGBalance(address[i])
	// 		if err != nil {
	// 			log.Error("Get ONG balance of address [%v] failed with error : [%v]", address[i], err)
	// 			return nil, err
	// 		}
	// 		ong, err := convertBigIntToFloatDecimal(balance.ONGBalance.String())
	// 		if err != nil {
	// 			log.Error("Get ONG balance of address [%v] failed with error : [%v]", address[i], err)
	// 		}
	// 		ongUnbound, err := convertBigIntToFloatDecimal(balance.ONGUnbound.String())
	// 		if err != nil {
	// 			log.Error("Get ONG balance of address [%v] failed with error : [%v]", address[i], err)
	// 		}

	// 		tokenBalance.Balance = &openwallet.Balance{
	// 			Address:          address[i],
	// 			Symbol:           contract.Symbol,
	// 			Balance:          "0",
	// 			ConfirmBalance:   ong.String(),
	// 			UnconfirmBalance: ongUnbound.String(),
	// 		}
	// 	} else {
	// 		// other contract
	// 	}
	// 	tokenBalanceList = append(tokenBalanceList, &tokenBalance)
	// }

	// return tokenBalanceList, nil
	return nil, nil
}
