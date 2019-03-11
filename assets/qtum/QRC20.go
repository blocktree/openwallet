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
	"encoding/hex"
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

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

	var tokenBalanceList []*openwallet.TokenBalance

 	for i:=0; i<len(address); i++ {
		unspent, _ := decoder.wm.GetQRC20Balance(contract, address[i], decoder.wm.config.isTestNet)
		//if err != nil {
		//	log.Errorf("get address[%v] QRC20 token balance failed, err=%v", address[i], err)
		//}

		balanceConfirmed := unspent
		//		log.Debugf("got balanceAll of [%v] :%v", address, balanceAll)
		balanceUnconfirmed := big.NewInt(0)
		//balanceUnconfirmed.Sub(balanceAll, balanceConfirmed)

		balance := &openwallet.TokenBalance{
			Contract: &contract,
			Balance: &openwallet.Balance{
				Address:          address[i],
				Symbol:           contract.Symbol,
				Balance:          unspent.String(),
				ConfirmBalance:   balanceConfirmed.String(),
				UnconfirmBalance: balanceUnconfirmed.String(),
			},
		}

		tokenBalanceList = append(tokenBalanceList, balance)
	}

	return tokenBalanceList, nil
}

func AddressTo32bytesArg(address string, isTestNet bool) ([]byte, error) {

	var addressToHash160 []byte
	if isTestNet {
		addressToHash160, _ = addressEncoder.AddressDecode(address, addressEncoder.QTUM_testnetAddressP2PKH)
	}else {
		addressToHash160, _ = addressEncoder.AddressDecode(address, addressEncoder.QTUM_mainnetAddressP2PKH)
	}

	//fmt.Printf("addressToHash160: %s\n",hex.EncodeToString(addressToHash160))

	to32bytesArg := append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, addressToHash160[:]...)
	//fmt.Printf("to32bytesArg: %s\n",hex.EncodeToString(to32bytesArg))

	return to32bytesArg, nil
}

// GetQRC20Balance 获取qrc20余额
func (wm *WalletManager) GetQRC20Balance(token openwallet.SmartContract, address string, isTestNet bool) (decimal.Decimal, error) {
	if wm.config.RPCServerType == RPCServerExplorer {
		return wm.getAddressTokenBalanceByExplorer(token, address)
	} else {
		return wm.GetQRC20UnspentByAddress(token.Address, address, token.Decimals, isTestNet)
	}
}

func (wm *WalletManager)GetQRC20UnspentByAddress(contractAddress, address string, tokenDecimal uint64, isTestNet bool) (decimal.Decimal, error) {

	trimContractAddr := strings.TrimPrefix(contractAddress, "0x")

	to32bytesArg, err := AddressTo32bytesArg(address, isTestNet)
	if err != nil {
		return decimal.New(0,0), err
	}

	combineString := hex.EncodeToString(append([]byte{0x70, 0xa0, 0x82, 0x31}, to32bytesArg[:]...))
	//fmt.Printf("combineString: %s\n",combineString)

	request := []interface{}{
		trimContractAddr,
		combineString,
	}

	result, err := wm.walletClient.Call("callcontract", request)
	if err != nil {
		return decimal.New(0,0), err
	}

	//fmt.Printf("Callcontract result: %s\n", result.String())

	QRC20Utox := NewQRC20Unspent(result)

	sotashiUnspent, _ := strconv.ParseInt(QRC20Utox.Output,16,64)
	sotashiUnspentDecimal, _ := decimal.NewFromString(common.NewString(sotashiUnspent).String())
	unspent := sotashiUnspentDecimal.Div(decimal.New(1, int32(tokenDecimal)))

	return unspent, nil
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

func (wm *WalletManager)QRC20Transfer(contractAddress string, from string, to string, gasPrice string, amount decimal.Decimal, gasLimit int64, tokenDecimal uint64, isTestNet bool) (string, error){

	trimContractAddr := strings.TrimPrefix(contractAddress, "0x")

	amountDecimal := amount.Mul(decimal.New(1, int32(tokenDecimal)))
	sotashiAmount := amountDecimal.IntPart()

	amountToArg, err := AmountTo32bytesArg(sotashiAmount)
	if err != nil {
		return "", err
	}

	addressToArg, err := AddressTo32bytesArg(to, isTestNet)
	if err != nil {
		return "", err
	}

	combineString := hex.EncodeToString(append([]byte{0xa9, 0x05, 0x9c, 0xbb}, addressToArg[:]...))

	dataHex := combineString + amountToArg
	fmt.Printf("dataHex: %s\n",dataHex)

	request := []interface{}{
		trimContractAddr,
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
