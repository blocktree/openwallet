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
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/cmd/utils"
	"github.com/blocktree/OpenWallet/log"
	wn "github.com/blocktree/OpenWallet/walletnode"
	"gopkg.in/urfave/cli.v1"
)

var (
	// 全节点命令
	CmdNode = cli.Command{
		Name:        "node",
		Usage:       "Manage fullnode of wallet",
		ArgsUsage:   "",
		Category:    "Application COMMANDS",
		Description: `Manage fullnode`,
		Subcommands: []cli.Command{
			{
				//节点状态
				Name:     "status",
				Usage:    "get status of full node server",
				Action:   getNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node status -s <symbol>

				`,
			},
			{
				//创建容器
				Name:     "create",
				Usage:    "create new container",
				Action:   createNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node createContainer
			
			Create a container
			
				`,
			},
			{
				//启动节点
				Name:     "start",
				Usage:    "start full node server",
				Action:   startNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node start -s <symbol>

				`,
			},
			{
				//关闭节点
				Name:     "stop",
				Usage:    "stop full node server",
				Action:   stopNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node stop -s <symbol>
			
				`,
			},
			{
				//重启节点
				Name:     "restart",
				Usage:    "restart full node server",
				Action:   restartNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node restart -s <symbol>
			
				`,
			},
			{
				//移除容器
				Name:     "remove",
				Usage:    "remove fullnode server",
				Action:   removeNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node remove -s <symbol>
			
			Remove a container
			
				`,
			},
			{
				Name:     "logs",
				Usage:    "show logs of fullnode server",
				Action:   logsNode,
				Category: "FULLNODE COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
				wmd node logs -s <symbol>
			
			Remove a container
			
				`,
			},
			//			{
			//				//节点配置
			//				Name:     "initNodeConfig",
			//				Usage:    "init node configuration",
			//				Action:   initNodeConfig,
			//				Category: "FULLNODE COMMANDS",
			//				Flags: []cli.Flag{
			//					utils.SymbolFlag,
			//				},
			//				Description: `
			//	wmd node initNodeConfig
			//init node configuration
			//
			//	`,
			//			},
			//			{
			//				//创建镜像
			//				Name:     "buildImage",
			//				Usage:    "build docker image",
			//				Action:   buildImage,
			//				Category: "FULLNODE COMMANDS",
			//				Flags: []cli.Flag{
			//					utils.SymbolFlag,
			//				},
			//				Description: `
			//	wmd node buildImage
			//Build docker image
			//
			//	`,
			//			},
			//			{
			//				//运行节点
			//				Name:     "runFullNode",
			//				Usage:    "run a full node in container",
			//				Action:   runFullNode,
			//				Category: "FULLNODE COMMANDS",
			//				Flags: []cli.Flag{
			//					utils.SymbolFlag,
			//				},
			//				Description: `
			//	wmd node runFullNode
			//
			//Run a full node process in docker container
			//
			//	`,
			//			},
			//			{
			//				//登录容器
			//				Name:     "loginContainer",
			//				Usage:    "login a container",
			//				Action:   loginContainer,
			//				Category: "FULLNODE COMMANDS",
			//				Flags: []cli.Flag{
			//					utils.SymbolFlag,
			//				},
			//				Description: `
			//	wmd node loginContainer
			//
			//Login a created container by name
			//
			//	`,
			//			},
		},
	}
)

func getNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.GetNodeStatus(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func createNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.CreateNodeFlow(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func startNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.StartNodeFlow(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func stopNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.StopNodeFlow(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func restartNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.RestartNodeFlow(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func removeNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	err := m.RemoveNodeFlow(symbol)
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return nil
}

func logsNode(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m := assets.NodeManager(&wn.NodeManager{})
	if m == nil {
		log.Error(symbol, " walletnode manager did not load")
		return nil
	}
	if err := m.LogsNodeFlow(symbol); err != nil {
		log.Error("unexpected error: ", err)
	}
	return nil
}

//func initNodeConfig(c *cli.Context) (error) {
//	symbol := c.String("symbol")
//	if len(symbol) == 0 {
//		log.Error("Argument -s <symbol> is missing")
//	}
//	m := assets.GetWMD(symbol).(assets.NodeManager)
//	if m == nil {
//		log.Log.Errorf("%s wallet manager is not register\n", symbol)
//	}
//	//配置钱包
//	err := m.InitNodeConfig()
//	if err != nil {
//		log.Error("unexpected error: ", err)
//	}
//	return err
//}
//
//func buildImage(c *cli.Context) error {
//	symbol := c.String("symbol")
//	if len(symbol) == 0 {
//		log.Error("Argument -s <symbol> is missing")
//	}
//	m := assets.GetWMD(symbol).(assets.NodeManager)
//	if m == nil {
//		log.Log.Errorf("%s wallet manager is not register\n", symbol)
//	}
//	//配置钱包
//	err := m.BuildImage()
//	if err != nil {
//		log.Error("unexpected error: ", err)
//	}
//	return err
//}
//
//func loginContainer(c *cli.Context) error {
//	symbol := c.String("symbol")
//	if len(symbol) == 0 {
//		log.Error("Argument -s <symbol> is missing")
//	}
//	m := assets.GetWMD(symbol).(assets.NodeManager)
//	if m == nil {
//		log.Log.Errorf("%s wallet manager is not register\n", symbol)
//	}
//	//配置钱包
//	err := m.LoginContainer()
//	if err != nil {
//		log.Error("unexpected error: ", err)
//	}
//	return err
//}
