/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitcoin

import (
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
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


//GetOmniInfo
func (wm *WalletManager) GetOmniInfo() (*gjson.Result, error) {

	result, err := wm.OnmiClient.Call("omni_getinfo", nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//GetOmniProperty 获取Omni资产信息
func (wm *WalletManager)GetOmniProperty(propertyId uint64) (*gjson.Result, error) {

	request := []interface{}{
		propertyId,
	}

	result, err := wm.OnmiClient.Call("omni_getproperty", request)
	if err != nil {
		return nil, err
	}
	return result, nil
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
