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

package eosio

import (
	"github.com/blocktree/openwallet/openwallet"
)

type ContractDecoder struct {
	openwallet.SmartContractDecoderBase
	wm *WalletManager
}

//NewContractDecoder 智能合约解析器
func NewContractDecoder(wm *WalletManager) *ContractDecoder {
	decoder := ContractDecoder{}
	decoder.wm = wm
	return &decoder
}

//func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {
//
//	codeAccount := contract.Address
//	tokenCoin := contract.Token
//	//tokenDecimals := rawTx.Coin.Contract.Decimals
//
//	//获取wallet
//	account, err := wrapper.GetAssetsAccountInfo(accountID)
//	if err != nil {
//		return err
//	}
//
//	accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(account.Alias), tokenCoin, eos.AccountName(codeAccount))
//	if len(accountAssets) == 0 {
//		return fmt.Errorf("eos account balance is not enough")
//	}
//
//	accountBalance = accountAssets[0]
//
//}