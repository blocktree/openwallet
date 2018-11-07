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

package ontology

import (
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"path/filepath"
	"errors"
)

type WalletManager struct {

	openwallet.AssetsAdapterBase

	Storage        *hdkeystore.HDKeystore        //秘钥存取
	WalletClient   *Client                       // 节点客户端
	ExplorerClient *Explorer                     // 浏览器API客户端
	Config         *WalletConfig                 //钱包管理配置
	WalletsInSum   map[string]*openwallet.Wallet //参与汇总的钱包
	Blockscanner   *BTCBlockScanner              //区块扫描器
	Decoder        openwallet.AddressDecoder     //地址编码器
	TxDecoder      openwallet.TransactionDecoder //交易单编码器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewBTCBlockScanner(&wm)
	//wm.Decoder = NewAddressDecoder(&wm)
	//wm.TxDecoder = NewTransactionDecoder(&wm)
	return &wm
}

//GetWalletInfo 获取钱包列表
func (wm *WalletManager) GetWalletInfo(walletID string) (*openwallet.Wallet, error) {

	wallets, err := wm.GetWallets()
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWallets 获取钱包列表
func (wm *WalletManager) GetWallets() ([]*openwallet.Wallet, error) {

	wallets, err := openwallet.GetWalletsByKeyDir(wm.Config.keyDir)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		w.DBFile = filepath.Join(wm.Config.dbPath, w.FileName()+".db")
	}

	return wallets, nil

}