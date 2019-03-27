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
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"math/big"
	"testing"
)

func TestWalletManager_TriggerSmartContract(t *testing.T) {

	contractAddr := "417EA07B5BE5A0FE26A64ACAF451C8D8653FDB56B6"
	function := "balanceOf(address)"
	tokenOwner := "41EFB6D8A02F4B639605D71FF8DC78C97329759D70"
	param, err := makeTransactionParameter("", []SolidityParam{
		SolidityParam{
			SOLIDITY_TYPE_ADDRESS,
			tokenOwner,
		},
	})
	//ownerAddr := "41BAF6BB7DE44E427FA11CE85EE59843E4DEFA114E"
	if err != nil {
		t.Errorf("makeTransactionParameter failed: %v\n", err)
		return
	}

	tx, err := tw.TriggerSmartContract(contractAddr, function, param, 100000, 0, tokenOwner)
	if err != nil {
		t.Errorf("TriggerSmartContract failed: %v\n", err)
		return
	}
	log.Infof("TriggerSmartContract: %+v", tx)
}

func TestTRC20TransferData(t *testing.T) {

	toAddrHex := "41EFB6D8A02F4B639605D71FF8DC78C97329759D70"
	amount := big.NewInt(1199900000000)

	var funcParams []SolidityParam
	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_ADDRESS,
		ParamValue: toAddrHex,
	})

	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_UINT256,
		ParamValue: amount,
	})

	//fmt.Println("make token transfer data, amount:", amount.String())
	dataHex, err := makeTransactionParameter(TRC20_TRANSFER_METHOD_ID, funcParams)
	if err != nil {
		t.Errorf("makeTransactionParameter failed: %v\n", err)
		return
	}
	log.Infof("makeTransactionParameter: %+v", dataHex)


	contractAddr := "417EA07B5BE5A0FE26A64ACAF451C8D8653FDB56B6"
	function := "transfer(address,uint256)"
	tokenOwner := "411561161E6AFF4A66BA660651D8E61428C50C57B8"
	if err != nil {
		t.Errorf("makeTransactionParameter failed: %v\n", err)
		return
	}

	triggHex, err := makeTransactionParameter("", funcParams)
	if err != nil {
		t.Errorf("makeTransactionParameter failed: %v\n", err)
		return
	}

	tx, err := tw.TriggerSmartContract(contractAddr, function, triggHex, 10000000, 0, tokenOwner)
	if err != nil {
		t.Errorf("TriggerSmartContract failed: %v\n", err)
		return
	}
	log.Infof("TriggerSmartContract: %+v", tx)
}

func TestWalletManager_GetContract(t *testing.T) {
	r, err := tw.GetContractInfo("TMWkPhsb1dnkAVNy8ej53KrFNGWy9BJrfu")
	if err != nil {
		t.Errorf("GetContract failed: %v\n", err)
		return
	}
	log.Infof("GetContract: %+v", r)
}

func TestWalletManager_GetTRC20Balance(t *testing.T) {

	balance, err := tw.GetTRC20Balance(
		"TXphYHMUvT2ptHt8QtQb5i9T9DWUtfBWha",
		"TMWkPhsb1dnkAVNy8ej53KrFNGWy9BJrfu")
	if err != nil {
		t.Errorf("GetTRC20Balance failed: %v\n", err)
		return
	}
	log.Infof("balance: %+v", balance)
}

func TestWalletManager_GetTRC10Balance(t *testing.T) {

	balance, err := tw.GetTRC10Balance(
		"TXphYHMUvT2ptHt8QtQb5i9T9DWUtfBWha",
		"1002000")
	if err != nil {
		t.Errorf("GetTRC10Balance failed: %v\n", err)
		return
	}
	log.Infof("balance: %+v", balance)
}

func TestWalletManager_GetTokenBalanceByAddress(t *testing.T) {

	trc10Contract := openwallet.SmartContract{
		Address:  "1002000",
		Protocol: TRC10,
		Decimals: 6,
	}

	balance10, err := tw.ContractDecoder.GetTokenBalanceByAddress(
		trc10Contract,
		"TXphYHMUvT2ptHt8QtQb5i9T9DWUtfBWha")
	if err != nil {
		t.Errorf("GetTRC10Balance failed: %v\n", err)
		return
	}
	log.Infof("GetTRC10Balance: %+v", balance10[0].Balance)

	trc20Contract := openwallet.SmartContract{
		Address:  "TMWkPhsb1dnkAVNy8ej53KrFNGWy9BJrfu",
		Protocol: TRC20,
		Decimals: 6,
	}

	balance20, err := tw.ContractDecoder.GetTokenBalanceByAddress(
		trc20Contract,
		"TXphYHMUvT2ptHt8QtQb5i9T9DWUtfBWha")
	if err != nil {
		t.Errorf("GetTRC20Balance failed: %v\n", err)
		return
	}
	log.Infof("GetTRC20Balance: %+v", balance20[0].Balance)
}

func TestEncodeName(t *testing.T) {
	nameBytes, _ := hex.DecodeString("31303031343636")
	fmt.Println(string(nameBytes))
}

func TestParseTransferEvent(t *testing.T) {
	data := "a9059cbb0000000000000000000000415bdf283199369adb124f39dda845ae02c5d3eb5d0000000000000000000000000000000000000000000000000000000001312d00"
	to, amount, err := ParseTransferEvent(data)
	if err != nil {
		t.Errorf("ParseTransferEvent failed: %v\n", err)
		return
	}
	log.Infof("%s: %d",to, amount)
}