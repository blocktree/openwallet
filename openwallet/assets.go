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

package openwallet

import (
	"github.com/blocktree/openwallet/log"
)

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
	//@required
	SymbolInfo

	//配置
	//@required
	AssetsConfig

	//GetAddressDecodeV2 地址解析器V2
	//如果实现了AddressDecoderV2，就无需实现AddressDecoder
	//@required
	GetAddressDecoderV2() AddressDecoderV2

	//GetAddressDecode 地址解析器
	//如果实现了AddressDecoderV2，就无需实现AddressDecoder
	//@required
	GetAddressDecode() AddressDecoder

	//GetTransactionDecoder 交易单解析器
	//@required
	GetTransactionDecoder() TransactionDecoder

	//GetBlockScanner 获取区块链扫描器
	//@required
	GetBlockScanner() BlockScanner

	//GetSmartContractDecoder 获取智能合约解析器
	//@optional
	GetSmartContractDecoder() SmartContractDecoder

	//GetAssetsLogger 获取资产日志工具
	//@optional
	GetAssetsLogger() *log.OWLogger

	//GetJsonRPCEndpoint 获取全节点服务的JSON-RPC客户端
	//@optional
	GetJsonRPCEndpoint() JsonRPCEndpoint
}

type AssetsAdapterBase struct {
	SymbolInfoBase
	AssetsConfigBase
}

//GetAddressDecode 地址解析器
func (a *AssetsAdapterBase) InitAssetsAdapter() error {
	return nil
}

//GetAddressDecode 地址解析器
//如果实现了AddressDecoderV2，就无需实现AddressDecoder
func (a *AssetsAdapterBase) GetAddressDecode() AddressDecoder {
	return nil
}

//GetAddressDecode 地址解析器
//如果实现了AddressDecoderV2，就无需实现AddressDecoder
func (a *AssetsAdapterBase) GetAddressDecoderV2() AddressDecoderV2 {
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

//GetAssetsLogger 获取资产账户日志工具
func (a *AssetsAdapterBase) GetAssetsLogger() *log.OWLogger {
	return nil
}

//GetJsonRPCEndpoint 获取全节点服务的JSON-RPC客户端
//@optional
func (a *AssetsAdapterBase) GetJsonRPCEndpoint() JsonRPCEndpoint {
	return nil
}