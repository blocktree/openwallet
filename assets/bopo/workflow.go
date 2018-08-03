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

package bopo

import (
	// "bufio"
	// "encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	// "github.com/tidwall/gjson"
	// "github.com/blocktree/OpenWallet/common"
	// "github.com/blocktree/OpenWallet/common/file"
	// "github.com/btcsuite/btcd/chaincfg"
	// "github.com/btcsuite/btcutil"
	// "github.com/btcsuite/btcutil/hdkeychain"
	"github.com/bndr/gotabulate"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"path/filepath"
)

const (
	maxAddresNum = 10000
)

var (
	//秘钥存取
	// storage *keystore.HDKeystore
	// 节点客户端
	client *Client
)

func init() {
	// storage = keystore.NewHDKeystore(keyDir, keystore.StandardScryptN,
	// 	keystore.StandardScryptP)
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.Addr, w.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Alias", "Addr", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))
}

// loadConfig 读取配置
func loadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	sumAddress = c.String("sumAddress")
	rpcUser = c.String("rpcUser")
	rpcPassword = c.String("rpcPassword")
	nodeInstallPath = c.String("nodeInstallPath")
	isTestNet, _ = c.Bool("isTestNet")
	if isTestNet {
		walletDataPath = c.String("testNetDataPath")
	} else {
		walletDataPath = c.String("mainNetDataPath")
	}

	// token := basicAuth(rpcUser, rpcPassword)

	client = &Client{
		BaseURL: serverAPI,
		Debug:   false,
		// AccessToken: token,
	}
	return nil
}
