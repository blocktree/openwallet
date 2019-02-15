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

package openwallet

import (
	"encoding/base64"
	"fmt"
	"github.com/blocktree/OpenWallet/crypto"
	"strings"
)

type SmartContract struct {
	ContractID string `json:"contractID" storm:"id"` //计算ID：base64(sha256({symbol}_{address})) 主链symbol
	Symbol     string `json:"symbol"`	//主币的symbol
	Address    string `json:"address"`
	Token      string `json:"token"`    //合约的symbol
	Protocol   string `json:"protocol"`
	Name       string `json:"name"`
	Decimals   uint64 `json:"decimals"`
}

//GenContractID 合约ID
func GenContractID(symbol, address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("%v_%v", symbol, address))))
}

//SmartContractDecoder 智能合约解析器
type SmartContractDecoder interface {

	//GetTokenBalanceByAddress 查询地址token余额列表
	GetTokenBalanceByAddress(contract SmartContract, address ...string) ([]*TokenBalance, error)

	//GetSmartContractInfo 获取智能合约信息
	//GetSmartContractInfo(contractAddress string) (*SmartContract, error)

	//TODO:创建多重签名合约

	//TODO:通过交易单号查询部署的智能合约地址
}
