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
				Name:     "initNodeConfig",
				Usage:    "init node configuration",
				Action:   initNodeConfig,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd node initNodeConfig 
init node configuration

	`,
			},
			{
				//创建镜像
				Name:     "buildImage",
				Usage:    "build docker image",
				Action:   buildImage,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd node buildImage 
Build docker image

	`,
			},
			{
				//创建容器
				Name:     "runContainer",
				Usage:    "run container",
				Action:   runContainer,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd node runContainer 

Run a container  

	`,
			},
			{
				//运行节点
				Name:     "runFullNode",
				Usage:    "run a full node in container",
				Action:   runFullNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd node runFullNode 

Run a full node process in docker container

	`,
			},
			{
				//登录容器
				Name:     "loginContainer",
				Usage:    "login a container",
				Action:   loginContainer,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd node loginContainer

Login a created container by name

	`,
			},
		},
	}
)

func initNodeConfig(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.InitNodeConfig()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}

func buildImage(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.BuildImage()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}

func runContainer(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.RunContainer()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}

func runFullNode(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.RunFullNode()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}

	return err
}

func loginContainer(c *cli.Context) (error) {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}
	m := assets.GetWMD(symbol).(assets.NodeManager)
	if m == nil {
		openwLogger.Log.Errorf("%s wallet manager is not register\n", symbol)
	}
	//配置钱包
	err := m.LoginContainer()
	if err != nil {
		openwLogger.Log.Errorf("%v", err)
	}
	return err
}