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

package commands

import (
	"github.com/blocktree/OpenWallet/cmd/utils"
	"github.com/blocktree/OpenWallet/logger"
	"gopkg.in/urfave/cli.v1"
	"github.com/blocktree/OpenWallet/assets"
)

var (
	// 钱包命令
	CmdConfig = cli.Command{
		Name:      "config",
		Usage:     "Manage wallet config",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Manage wallet config

`,
		Subcommands: []cli.Command{
			{
				//查看配置文件信息
				Name:     "init",
				Usage:    "Init config flow",
				Action:   configWallet,
				Category: "CONFIG COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd config init -s ada

Init config flow.

	`,
			},
			{
				//查看配置文件信息
				Name:     "see",
				Usage:    "See Wallet config info",
				Action:   configSee,
				Category: "CONFIG COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd config see -s ada

See Wallet config info.

	`,
			},
		},
	}
)


//configWallet 配置钱包流程
func configWallet(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.InitConfigFlow()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}

//configSee 查看钱包配置信息
func configSee(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	err := m.ShowConfig()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}
