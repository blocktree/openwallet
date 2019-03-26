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

package tron

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"math/big"
	"strings"
)

/*

在TRON中检测TRX或TRC10事务涉及4种类型的合同：

TransferContract（系统合同类型：TRX转账）
TransferAssetContract（系统合同类型：TRC10转账）
CreateSmartContract（智能合约类型）
TriggerSmartContract（智能合约类型：TRC20转账）
Transaction，TransactionInfo 和 Block 的数据包含所有智能合约交易信息。

技术细节
https://cn.developers.tron.network/docs/%E4%BA%A4%E6%8D%A2%E4%B8%AD%E7%9A%84trc10%E5%92%8Ctrx%E8%BD%AC%E7%A7%BB

TRX转账示例
https://tronscan.org/#/transaction/f8f8ac5b4b0df34dad410147231061806c9fa8c207e7f3107cadc6d00925ccbc

TRC10转账示例
https://tronscan.org/#/transaction/c0edfc83e3535700b46598444f2425696686d20566101d8b5b2aa95c0915a2a0

TRC20转账示例
https://tronscan.org/#/transaction/a5614f60e7d3b9d8859abe89968d81007c321c5ad83cb9c7abaa736a20401a11

*/

const (
	TRX_GET_TOKEN_BALANCE_METHOD      = "0x70a08231"
	TRX_TRANSFER_TOKEN_BALANCE_METHOD = "0xa9059cbb"
	TRX_TRANSFER_EVENT_ID             = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

const (
	SOLIDITY_TYPE_ADDRESS = "address"
	SOLIDITY_TYPE_UINT256 = "uint256"
	SOLIDITY_TYPE_UINT160 = "uint160"
)

type SolidityParam struct {
	ParamType  string
	ParamValue interface{}
}

func makeRepeatString(c string, count uint) string {
	cs := make([]string, 0)
	for i := 0; i < int(count); i++ {
		cs = append(cs, c)
	}
	return strings.Join(cs, "")
}

func makeTransactionParameter(params []SolidityParam) (string, error) {

	data := ""
	for i, _ := range params {
		var param string
		if params[i].ParamType == SOLIDITY_TYPE_ADDRESS {
			param = strings.ToLower(params[i].ParamValue.(string))
			param = strings.TrimPrefix(param, "0x")
			if len(param) != 42 {
				return "", fmt.Errorf("length of address error.")
			}
			param = makeRepeatString("0", 22) + param
		} else if params[i].ParamType == SOLIDITY_TYPE_UINT256 {
			intParam := params[i].ParamValue.(*big.Int)
			param = intParam.Text(16)
			l := len(param)
			if l > 64 {
				return "", fmt.Errorf("integer overflow.")
			}
			param = makeRepeatString("0", uint(64-l)) + param
			//fmt.Println("makeTransactionData intParam:", intParam.String(), " param:", param)
		} else {
			return "", fmt.Errorf("not support solidity type")
		}

		data += param
	}
	return data, nil
}

//TriggerSmartContract 初始智能合约方法
func (wm *WalletManager) TriggerSmartContract(
	contractAddress string,
	function string,
	parameter string,
	feeLimit uint64,
	callValue uint64,
	ownerAddress string) (*gjson.Result, error) {
	params := req.Param{
		"contract_address":  contractAddress,
		"function_selector": function,
		"parameter":         parameter,
		"fee_limit":         feeLimit,
		"call_value":        callValue,
		"owner_address":     ownerAddress,
	}
	r, err := wm.WalletClient.Call("/wallet/triggersmartcontract", params)
	if err != nil {
		return nil, err
	}
	return r, nil
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

//func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {
//
//	var tokenBalanceList []*openwallet.TokenBalance
//
//	for i:=0; i<len(address); i++ {
//		propertyID := common.NewString(contract.Address).UInt64()
//		balance, err := decoder.wm.GetOmniBalance(propertyID, address[i])
//		if err != nil {
//			decoder.wm.Log.Errorf("get address[%v] omni token balance failed, err: %v", address[i], err)
//		}
//
//		tokenBalance := &openwallet.TokenBalance{
//			Contract: &contract,
//			Balance: &openwallet.Balance{
//				Address:          address[i],
//				Symbol:           contract.Symbol,
//				Balance:          balance.StringFixed(decoder.wm.Decimal()),
//				ConfirmBalance:   balance.StringFixed(decoder.wm.Decimal()),
//				UnconfirmBalance: "0",
//			},
//		}
//
//		tokenBalanceList = append(tokenBalanceList, tokenBalance)
//	}
//
//	return tokenBalanceList, nil
//}
