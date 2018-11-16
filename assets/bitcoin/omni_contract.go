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

package bitcoin

import (
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
)

func (wm *WalletManager) GetOmniBalance(propertyId uint64, address string) (decimal.Decimal, error) {
	request := []interface{}{
		address,
		propertyId,
	}

	result, err := wm.OnmiClient.Call("omni_getbalance", request)
	if err != nil {
		return decimal.Zero, err
	}

	balance, err := decimal.NewFromString(result.Get("balance").String())
	if err != nil {
		return decimal.Zero, err
	}

	return balance, nil
}

func (wm *WalletManager) GetOmniTransaction(txid string) (*OmniTransaction, error) {
	request := []interface{}{
		txid,
	}

	result, err := wm.OnmiClient.Call("omni_gettransaction", request)
	if err != nil {
		return nil, err
	}

	return NewOmniTx(result), nil
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

	var tokenBalanceList []*openwallet.TokenBalance

	for i:=0; i<len(address); i++ {
		propertyID := common.NewString(contract.Address).UInt64()
		balance, err := decoder.wm.GetOmniBalance(propertyID, address[i])
		if err != nil {
			decoder.wm.Log.Errorf("get address[%v] omni token balance failed, err: %v", address[i], err)
		}

		tokenBalance := &openwallet.TokenBalance{
			Contract: &contract,
			Balance: &openwallet.Balance{
				Address:          address[i],
				Symbol:           contract.Symbol,
				Balance:          balance.StringFixed(decoder.wm.Decimal()),
				ConfirmBalance:   balance.StringFixed(decoder.wm.Decimal()),
				UnconfirmBalance: "0",
			},
		}

		tokenBalanceList = append(tokenBalanceList, tokenBalance)
	}

	return tokenBalanceList, nil
}
