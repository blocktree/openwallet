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
	"gopkg.in/urfave/cli.v1"
	"github.com/blocktree/OpenWallet/cmd/utils"
	"github.com/blocktree/OpenWallet/mqnode"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/console"
)

var (
	// 服务命令
	CmdServer = cli.Command{
		Name:      "server",
		Usage:     "Manage mqnode server",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Use mqnode commands to join mqnode server

`,
		Subcommands: []cli.Command{
			{
				//创建或查看本地通信密钥对
				Name:      "keychain",
				Usage:     "create or see server keychain",
				ArgsUsage: "<init>",
				Action:    serviceKeychain,
				Category:  "mqnode COMMANDS",
				Flags: []cli.Flag{
					utils.InitFlag,
				},
				Description: `
	openw mqnode keychain [-ri]

This command will show the local publicKey and mqnode publicKey.
	`,
			},

			{
				//配置服务
				Name:      "config",
				Usage:     "Start a timer to sum wallet balance",
				ArgsUsage: "<init>",
				Action:    configServiceNode,
				Category:  "mqnode COMMANDS",
				Flags: []cli.Flag{
					utils.InitFlag,
					utils.SymbolFlag,
					utils.MQUrl,
					utils.MQAccount,
					utils.MQPassword,
					utils.MQExchange,
					utils.EnableBlockScan,
					utils.ISTestNet,
				},
				Description: `
	openw wallet config [i]

	`,
			},

			{
				//启动服务
				Name:      "run",
				Usage:     "Start a timer to sum wallet balance",
				ArgsUsage: "<init>",
				Action:    run,
				Category:  "mqnode COMMANDS",
				Flags: []cli.Flag{
					utils.InitFlag,
				},
				Description: `
	openw wallet config [i]

	`,
			},
		},
	}
)

func serviceKeychain(c *cli.Context) error {

	var (
		err error
	)

	isInit := c.Bool("init")

	if isInit {
		flag, err := console.Stdin.PromptConfirm("Create a new server will cover the existing server data and reinitialize a new one, please backup the existing server key first. Continue to create?")
		if err != nil {
			return err
		}
		if flag {
			err = mqnode.InitServiceKeychain()
		}
	} else {
		err = mqnode.GetMerchantKeychain()
	}

	if err != nil {
		log.Error("unexpected error: ", err)
	}

	return err
}



func configServiceNode(c *cli.Context) error {
	var (
		err error
	)
	symbol := c.String("s")

	if len(symbol) != 0 {
		if symbol != "" {
			err = mqnode.SetSymbolAssests(symbol)
			if err != nil {
				log.Error("unexpected error: ", err)
			}
		} else {
			log.Error("symbol can't be null ")
		}
	}

	isInit := c.Bool("init")

	if isInit {
		err = mqnode.ConfigService()
	} else {
		err = mqnode.ShowServiceConfig()
	}

	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}



func run(c *cli.Context) error {
	var (
		err error
	)
	logDir := c.GlobalString("logdir")
	debug := c.GlobalBool("debug")
	utils.SetupLog(logDir, "openw.log", debug)
	err = mqnode.RunServer()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

