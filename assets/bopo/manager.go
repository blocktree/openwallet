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

package bopo

import (
	"path/filepath"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type WalletManager struct {
	config         *WalletConfig                 //钱包管理配置
	storage        *hdkeystore.HDKeystore        //秘钥存取
	blockscanner   *FabricBlockScanner           //区块扫描器
	fullnodeClient *Client                       // 全节点客户端
	walletClient   *Client                       // 节点客户端
	walletsInSum   map[string]*openwallet.Wallet //参与汇总的钱包
}

// func NewWalletManager() *WalletManager {

// 	wm := WalletManager{}
// 	wm.config = NewConfig()
// 	wm.storage = hdkeystore.NewHDKeystore(wm.config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
// 	wm.blockscanner = NewFabricBlockScanner(&wm)

// 	return &wm
// }

//loadConfig 读取配置
func (wm *WalletManager) loadConfig() error {

	var (
		c   config.Configer
		err error
	)

	wm.config = NewWalletConfig()
	wm.storage = hdkeystore.NewHDKeystore(wm.config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.blockscanner = NewFabricBlockScanner(wm)

	//读取配置
	absFile := filepath.Join(wm.config.configFilePath, wm.config.configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	wm.config.walletURL = c.String("walletURL")
	wm.config.threshold, _ = decimal.NewFromString(c.String("threshold"))
	wm.config.sumAddress = c.String("sumAddress")
	wm.config.rpcUser = c.String("rpcUser")
	wm.config.rpcPassword = c.String("rpcPassword")
	wm.config.nodeInstallPath = c.String("nodeInstallPath")
	wm.config.isTestNet, _ = c.Bool("isTestNet")
	if wm.config.isTestNet {
		wm.config.walletDataPath = c.String("testNetDataPath")
	} else {
		wm.config.walletDataPath = c.String("mainNetDataPath")
	}
	// wm.walletClient = NewClient(wm.config.walletAPI, false)
	wm.fullnodeClient = NewClient(wm.config.walletURL, false)
	return nil
}
