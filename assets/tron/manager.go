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
	"log"

	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/openwallet"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

type WalletManager struct {
	Config         *WalletConfig                 //钱包管理配置
	Storage        *hdkeystore.HDKeystore        //秘钥存取
	Blockscanner   *TronBlockScanner             //区块扫描器
	FullnodeClient *Client                       // 全节点客户端
	WalletClient   *Client                       // 节点客户端
	WalletsInSum   map[string]*openwallet.Wallet //参与汇总的钱包
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	wm.Storage = hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewTronBlockScanner(&wm)
	// wm.Decoder = &addressDecoder{}
	// wm.TxDecoder = NewTransactionDecoder(&wm)
	return &wm
}
