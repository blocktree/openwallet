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
	// 全节点命令
	CmdNode = cli.Command{
		Name:      "node",
		Usage:     "Manage full node",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Manage full node

`,
		Subcommands: []cli.Command{
			{
				//创建镜像
				Name:     "install",
				Usage:    "install full node",
				Action:   installFullNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.DatadirFlag,
					utils.SymbolFlag,
				},
				Description: `
	wmd node install -s <symbol> -p <datadir>
install full node 

	`,
			},
			{
				//创建镜像
				Name:     "start",
				Usage:    "start full node",
				Action:   startFullNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.DatadirFlag,
					utils.SymbolFlag,
				},
				Description: `
	wmd node start
start full node

	`,
			},
			{
				//创建容器
				Name:     "stop",
				Usage:    "stop full node",
				Action:   stopFullNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.DatadirFlag,
					utils.SymbolFlag,
				},
				Description: `
	wmd node stop

stop full node

	`,
			},
			{
				//运行节点
				Name:     "restart",
				Usage:    "restart a full node",
				Action:   restartFullNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.DatadirFlag,
					utils.SymbolFlag,
				},
				Description: `
	wmd node restart 

restart full node
	`,
			},
			{
				//登录容器
				Name:     "info",
				Usage:    "full node infomations",
				Action:   fullNodeInfo,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.DatadirFlag,
					utils.SymbolFlag,
				},
				Description: `
	wmd node info

get full node infomation
	`,
			},
		},
	}
)

func installFullNode(c *cli.Context) (error) {
	symbol := c.String("symbol")
	datadir := c.String("datadir")

	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//安装全节点
	err := m.InstallFullNode(datadir)
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}

func startFullNode(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//运行全节点
	err := m.StartFullNode()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}

func stopFullNode(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//停止全节点
	err := m.StopFullNode()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}

func restartFullNode(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//重启全节点
	err := m.RestartFullNode()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}

func fullNodeInfo(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.FullNodeInfo()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}