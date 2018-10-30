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

package qtum

import (
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"fmt"
	"encoding/hex"
	"strconv"
	"github.com/shopspring/decimal"
	"github.com/blocktree/OpenWallet/openwallet"
	"math/big"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/common"
)

func (wm *WalletManager) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {

	var tokenBalanceList []*openwallet.TokenBalance

 	for i:=0; i<len(address); i++ {
		QRC20Utox, err := wm.GetQRC20UnspentByAddress(contract.Address, address[i])
		if err != nil {
			log.Errorf("get address[%v] QRC20 token balance failed, err=%v", address[i], err)
		}

		sotashiUnspent, _ := strconv.ParseInt(QRC20Utox.Output,16,64)
		sotashiUnspentDecimal, _ := decimal.NewFromString(common.NewString(sotashiUnspent).String())
		balanceAll := sotashiUnspentDecimal.Div(coinDecimal)

		balanceConfirmed := balanceAll
		//		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
		balanceUnconfirmed := big.NewInt(0)
		//balanceUnconfirmed.Sub(balanceAll, balanceConfirmed)

		balance := &openwallet.TokenBalance{
			Contract: &contract,
			Balance: &openwallet.Balance{
				Address:          address[i],
				Symbol:           contract.Symbol,
				Balance:          balanceAll.String(),
				ConfirmBalance:   balanceConfirmed.String(),
				UnconfirmBalance: balanceUnconfirmed.String(),
			},
		}

		tokenBalanceList = append(tokenBalanceList, balance)
	}

	return tokenBalanceList, nil
}

func AddressTo32bytesArg(address string) ([]byte, error) {

	addressToHash160, _ := addressEncoder.AddressDecode(address, addressEncoder.QTUM_testnetAddressP2PKH)
	//fmt.Printf("addressToHash160: %s\n",hex.EncodeToString(addressToHash160))

	to32bytesArg := append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, addressToHash160[:]...)
	//fmt.Printf("to32bytesArg: %s\n",hex.EncodeToString(to32bytesArg))

	return to32bytesArg, nil
}

// GetQRC20Balance 获取qrc20余额
func (wm *WalletManager) GetQRC20Balance(token openwallet.SmartContract, address string) (decimal.Decimal, error) {
	if wm.config.RPCServerType == RPCServerExplorer {
		return wm.getAddressTokenBalanceByExplorer(token, address)
	} else {
		return wm.getAddressTokenBalanceByCore(token, address)
	}
}

//getAddressTokenBalanceByCore 通过合约地址查询用户地址的余额
func (wm *WalletManager) getAddressTokenBalanceByCore(token openwallet.SmartContract, address string) (decimal.Decimal, error) {

	unspent, err := wm.GetQRC20UnspentByAddress(token.Address, address)
	if err != nil {
		return decimal.New(0, 0), err
	}

	dec := decimal.New(1, int32(token.Decimals))
	sotashiUnspent, _ := strconv.ParseInt(unspent.Output,16,64)
	balance, _ := decimal.NewFromString(common.NewString(sotashiUnspent).String())
	balance = balance.Div(dec)
	return balance, nil

}

func (wm *WalletManager)GetQRC20UnspentByAddress(contractAddress, address string) (*QRC20Unspent,error) {

	to32bytesArg, err := AddressTo32bytesArg(address)
	if err != nil {
		return nil, err
	}

	combineString := hex.EncodeToString(append([]byte{0x70, 0xa0, 0x82, 0x31}, to32bytesArg[:]...))
	//fmt.Printf("combineString: %s\n",combineString)

	request := []interface{}{
		contractAddress,
		combineString,
	}

	result, err := wm.walletClient.Call("callcontract", request)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("Callcontract result: %s\n", result.String())

	QRC20Utox := NewQRC20Unspent(result)

	return QRC20Utox, nil
}

func AmountTo32bytesArg(amount int64) (string, error) {

	hexAmount := strconv.FormatInt(amount, 16)

	defaultLen := 64
	addLen := defaultLen - len(hexAmount)
	var bytesArg string

	for i := 0; i<addLen; i++ {
		bytesArg = bytesArg + "0"
	}

	bytesArg = bytesArg + hexAmount

	return bytesArg, nil
}

func (wm *WalletManager)QRC20Transfer(contractAddress string, from string, to string, gasPrice string, amount decimal.Decimal, gasLimit int64) (string, error){

	amountDecimal := amount.Mul(coinDecimal)
	sotashiAmount := amountDecimal.IntPart()

	amountToArg, err := AmountTo32bytesArg(sotashiAmount)
	if err != nil {
		return "", err
	}

	addressToArg, err := AddressTo32bytesArg(to)
	if err != nil {
		return "", err
	}

	combineString := hex.EncodeToString(append([]byte{0xa9, 0x05, 0x9c, 0xbb}, addressToArg[:]...))

	dataHex := combineString + amountToArg
	fmt.Printf("dataHex: %s\n",dataHex)

	request := []interface{}{
		contractAddress,
		dataHex,
		0,
		gasLimit,
		gasPrice,
		from,
	}

	result, err := wm.walletClient.Call("sendtocontract", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}
