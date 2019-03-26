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
	"math/big"
	"testing"
)

func TestWalletManager_TriggerSmartContract(t *testing.T) {

	contractAddr := "417EA07B5BE5A0FE26A64ACAF451C8D8653FDB56B6"
	function := "balanceOf(address)"
	tokenOwner := "41EFB6D8A02F4B639605D71FF8DC78C97329759D70"
	param, err := makeTransactionParameter([]SolidityParam{
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

	r, err := tw.TriggerSmartContract(contractAddr, function, param, 100000000, 0, tokenOwner)
	if err != nil {
		t.Errorf("TriggerSmartContract failed: %v\n", err)
		return
	}
	if r.Get("result.message").Exists() {
		msg, _ := hex.DecodeString(r.Get("result.message").String())
		t.Errorf("TriggerSmartContract failed: %v\n", string(msg))
		return
	}
	constant_result := r.Get("constant_result").Array()[0].String()
	balance, err := ConvertToBigInt(constant_result, 16)

	log.Infof("TriggerSmartContract: %+v", balance.String())
}

func ConvertToBigInt(value string, base int) (*big.Int, error) {
	bigvalue := new(big.Int)
	var success bool

	if value == "" {
		value = "0"
	}

	_, success = bigvalue.SetString(value, base)
	if !success {
		errInfo := fmt.Sprintf("convert value [%v] to bigint failed, check the value and base passed through\n", value)
		log.Errorf(errInfo)
		return big.NewInt(0), fmt.Errorf(errInfo)
	}
	return bigvalue, nil
}