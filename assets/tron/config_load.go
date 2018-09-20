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
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/astaxie/beego/config"
	"github.com/shopspring/decimal"
)

const (
	maxAddresNum = 10000
)

//loadConfig 读取配置
func (wm *WalletManager) LoadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	wm.Config.ServerAPI = c.String("serverAPI")
	wm.Config.Threshold, _ = decimal.NewFromString(c.String("threshold"))
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.RpcUser = c.String("rpcUser")
	wm.Config.RpcPassword = c.String("rpcPassword")
	wm.Config.NodeInstallPath = c.String("nodeInstallPath")
	wm.Config.IsTestNet, _ = c.Bool("isTestNet")
	if wm.Config.IsTestNet {
		wm.Config.WalletDataPath = c.String("testNetDataPath")
	} else {
		wm.Config.WalletDataPath = c.String("mainNetDataPath")
	}

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}

	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)

	// token := BasicAuth(wm.Config.RpcUser, wm.Config.RpcPassword)

	wm.WalletClient = NewClient(wm.Config.ServerAPI, "", false)

	return nil
}
