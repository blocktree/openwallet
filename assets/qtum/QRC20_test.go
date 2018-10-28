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
	"testing"
	"encoding/hex"
	"strconv"
	"github.com/shopspring/decimal"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
)

func Test_addressTo32bytesArg(t *testing.T) {
	address := "qdphfFinfJutJFvtnr2UaCwNAMxC3HbVxa"

	to32bytesArg, err := AddressTo32bytesArg(address)
	if err != nil {
		t.Errorf("To32bytesArg failed unexpected error: %v\n", err)
	}else {
		t.Logf("To32bytesArg success.")
	}

	t.Logf("This is to32bytesArg string for you to use: %s\n", hex.EncodeToString(to32bytesArg))
}


func Test_getUnspentByAddress(t *testing.T) {
	contractAddress := "91a6081095ef860d28874c9db613e7a4107b0281"
	address := "qakZE2dYZU7VF1m5RDRBFVr1fFMFKhTNWA"

	QRC20Utox, err := tw.GetQRC20UnspentByAddress(contractAddress, address)
	if err != nil {
		t.Errorf("GetUnspentByAddress failed unexpected error: %v\n", err)
	}

	sotashiUnspent, _ := strconv.ParseInt(QRC20Utox.Output,16,64)
	sotashiUnspentDecimal, _ := decimal.NewFromString(common.NewString(sotashiUnspent).String())
	unspent := sotashiUnspentDecimal.Div(coinDecimal)

	if err != nil {
		t.Errorf("strconv.ParseInt failed unexpected error: %v\n", err)
	}else {
		t.Logf("QRC20Unspent %s: %s = %v\n", QRC20Utox.Address, address, unspent)
	}
}

func Test_AmountTo32bytesArg(t *testing.T){
	var amount int64= 100000000
	bytesArg, err := AmountTo32bytesArg(amount)
	if err != nil {
		t.Errorf("strconv.ParseInt failed unexpected error: %v\n", err)
	}else {
		t.Logf("hexAmount = %s\n", bytesArg)
	}
}

func Test_QRC20Transfer(t *testing.T) {
	contractAddress := "91a6081095ef860d28874c9db613e7a4107b0281"
	from := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"
	to := "qJ2HTPYoMF1DPBhgURjRqemun5WimD57Hy"
	gasPrice := "0.00000040"
	var gasLimit int64 = 250000
	var amount decimal.Decimal = decimal.NewFromFloat(10)

	result, err := tw.QRC20Transfer(contractAddress, from, to, gasPrice, amount, gasLimit)
	if err != nil {
		t.Errorf("QRC20Transfer failed unexpected error: %v\n", err)
	}else {
		t.Logf("QRC20Transfer = %s\n", result)
	}
}

func Test_GetTokenBalanceByAddress(t *testing.T) {
	contract := openwallet.SmartContract{
		Address: "91a6081095ef860d28874c9db613e7a4107b0281",
	}
	addrs := []string{
		"qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP",
		"qQLYQn7vCAU8irPEeqjZ3rhFGLnS5vxVy8",
		"qMXS1YFtA5qr2UfhcDMthTCK6hWhJnzC47",
		"qJq5GbHeaaNbi6Bs5QCbuCZsZRXVWPoG1k",
		"qP1VPw7RYm5qRuqcAvtiZ1cpurQpVWREu8",
		"qdphfFinfJutJFvtnr2UaCwNAMxC3HbVxa",
	}
	balanceList, err := tw.GetTokenBalanceByAddress(contract, addrs...)
	if err != nil {
		t.Errorf("get token balance by address failed, err=%v", err)
		return
	}

	//输出json格式
	//objStr, _ := json.MarshalIndent(balanceList, "", " ")
	//t.Logf("balance list:%v", string(objStr))

	for i:=0; i<len(balanceList); i++ {
		t.Logf("%s: %s\n",addrs[i], balanceList[i].Balance.ConfirmBalance)
	}
}