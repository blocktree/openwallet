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
	"errors"
	"path/filepath"

	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/openwallet"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	Storage         *hdkeystore.HDKeystore        //秘钥存取
	RPCClient       *RpcClient                    // RPC API
	Config          *WalletConfig                 //钱包管理配置
	WalletsInSum    map[string]*openwallet.Wallet //参与汇总的钱包
	Blockscanner    *ONTBlockScanner              //区块扫描器
	Decoder         openwallet.AddressDecoder     //地址编码器
	TxDecoder       openwallet.TransactionDecoder //交易单编码器
	ContractDecoder *ContractDecoder              //智能合约解析器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewONTBlockScanner(&wm)
	wm.Decoder = NewAddressDecoder(&wm)
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.ContractDecoder = NewContractDecoder(&wm)

	//	wm.RPCClient = NewRpcClient("http://localhost:20336/")
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

//SendRawTransaction 广播交易
func (wm *WalletManager) SendRawTransaction(txHex string) (string, error) {

	return wm.sendRawTransactionByRest(txHex)
}

func (wm *WalletManager) sendRawTransactionByRest(txHex string) (string, error) {
	params := []interface{}{txHex}

	txid, err := wm.RPCClient.sendRpcRequest("0", "sendrawtransaction", params)
	if err != nil {
		return "", err
	}
	return string(txid), nil
}
