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

//Coin 币种信息
type Coin struct {
	Symbol     string        `json:"symbol"`
	IsContract bool          `json:"isContract"`
	ContractID string        `json:"contractID"`
	Contract   SmartContract `json:"contract"`
}

// AssetsAdapter 资产适配器接口
// 适配OpenWallet钱包体系的抽象接口
type AssetsAdapter interface {

	//币种信息
	SymbolInfo

	//GetAddressDecode 地址解析器
	GetAddressDecode() AddressDecoder

	//GetTransactionDecoder 交易单解析器
	GetTransactionDecoder() TransactionDecoder

	//GetBlockScanner 获取区块链扫描器
	GetBlockScanner() BlockScanner

	//GetSmartContractDecoder 获取智能合约解析器
	GetSmartContractDecoder() SmartContractDecoder
}

type AssetsAdapterBase struct {
	SymbolInfoBase
}

//GetAddressDecode 地址解析器
func (a *AssetsAdapterBase) GetAddressDecode() AddressDecoder {
	return nil
}

//GetTransactionDecoder 交易单解析器
func (a *AssetsAdapterBase) GetTransactionDecoder() TransactionDecoder {
	return nil
}

//GetBlockScanner 获取区块链扫描器
func (a *AssetsAdapterBase) GetBlockScanner() BlockScanner {
	return nil
}

//GetBlockScanner 获取智能合约解析器
func (a *AssetsAdapterBase) GetSmartContractDecoder() SmartContractDecoder {
	return nil
}
