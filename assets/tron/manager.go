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
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

func init() {
	// log.SetFlags(log.Lshortfile | log.LstdFlags)
}

// WalletManager struct
type WalletManager struct {
	openwallet.AssetsAdapterBase

	Config         *WalletConfig          //钱包管理配置
	Storage        *hdkeystore.HDKeystore //秘钥存取
	FullnodeClient *Client                // 全节点客户端
	WalletClient   *Client                // 节点客户端
	Log            *log.OWLogger          //日志工具

	WalletsInSum map[string]*openwallet.Wallet //参与汇总的钱包

	Blockscanner    *TronBlockScanner               //区块扫描器
	AddrDecoder     openwallet.AddressDecoder       //地址编码器
	TxDecoder       openwallet.TransactionDecoder   //交易单编码器
	ContractDecoder openwallet.SmartContractDecoder //
}

// NewWalletManager create instance
func NewWalletManager() *WalletManager {

	wm := WalletManager{}
	wm.Config = NewConfig()
	wm.Storage = hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	wm.Blockscanner = NewTronBlockScanner(&wm)
	wm.AddrDecoder = NewAddressDecoder(&wm)
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.Log = log.NewOWLogger(wm.Symbol())
	wm.WalletClient = NewClient("http://192.168.27.124:18090", "", true)
	return &wm
}

//------------------------------------------------------------------------------------------------

//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}

//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "TRX"
}

//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}

//Decimal 小数位精度 *1000000
func (wm *WalletManager) Decimal() int32 {
	return 6
}

//GetAddressDecode 地址解析器
func (wm *WalletManager) GetAddressDecode() openwallet.AddressDecoder {
	return wm.AddrDecoder
}

//GetTransactionDecoder 交易单解析器
func (wm *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return wm.TxDecoder
}

//GetBlockScanner 获取区块链
func (wm *WalletManager) GetBlockScanner() openwallet.BlockScanner {
	return wm.Blockscanner
}

// func (this *WalletManager) GetSmartContractDecoder() openwallet.SmartContractDecoder {
// 	return this.ContractDecoder
// }
