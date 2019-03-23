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
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/tidwall/gjson"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	WalletClient    *Client                         // 节点客户端
	Config          *WalletConfig                   // 节点配置
	Decoder         openwallet.AddressDecoder       //地址编码器
	TxDecoder       openwallet.TransactionDecoder   //交易单编码器
	Log             *log.OWLogger                   //日志工具
	ContractDecoder openwallet.SmartContractDecoder //智能合约解析器
	Blockscanner    *BlockScanner                   //区块扫描器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol)
	//区块扫描器
	wm.Blockscanner = NewBlockScanner(&wm)
	wm.Decoder = NewAddressDecoder(&wm)
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.Log = log.NewOWLogger(wm.Symbol())
	wm.ContractDecoder = NewContractDecoder(&wm)
	return &wm
}

func (wm *WalletManager) GetInfo() (*gjson.Result, error) {
	result, err := wm.WalletClient.Call("chain/get_info", nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wm *WalletManager) GetAccount(name string) (*gjson.Result, error) {

	param := map[string]interface{}{
		"account_name": name,
	}

	result, err := wm.WalletClient.Call("chain/get_account", param)
	if err != nil {
		return nil, err
	}
	return result, nil
}