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
	// 钱包命令
	CmdMerchant = cli.Command{
		Name:      "node",
		Usage:     "Manage mqnode node",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Use mqnode commands to join mqnode server

`,
		Subcommands: []cli.Command{
			{
				//创建或查看本地通信密钥对
				Name:      "keychain",
				Usage:     "create or see node keychain",
				ArgsUsage: "<init>",
				Action:    getMerchantKeychain,
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
				//加入并连接到商户节点
				Name:      "server",
				Usage:     "server [-s symbol,i -init,-p path ,-debug,-logdir]",
				Action:    joinMerchantNode,
				ArgsUsage: "<init>",
				Category:  "mqnode COMMANDS",
				Description: `
	openw mqnode sercer

This command will connect and join mqnode node.

	`,
			},

			{
				//配置商户
				Name:      "config",
				Usage:     "Start a timer to sum wallet balance",
				ArgsUsage: "<init>",
				Action:    configMerchantNode,
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

func getMerchantKeychain(c *cli.Context) error {

	var (
		err error
	)

	isInit := c.Bool("init")

	if isInit {
		flag, err := console.Stdin.PromptConfirm("Create a new node will cover the existing node data and reinitialize a new one, please backup the existing node key first. Continue to create?")
		if err != nil {
			return err
		}
		if flag {
			err = mqnode.InitMerchantKeychainFlow()
		}
	} else {
		err = mqnode.GetMerchantKeychain()
	}

	if err != nil {
		log.Error("unexpected error: ", err)
	}

	return err
}

func joinMerchantNode(c *cli.Context) error {
	var (
		err error
	)


	logDir := c.GlobalString("logdir")
	debug := c.GlobalBool("debug")
	utils.SetupLog(logDir, "mqnode.log", debug)

	err = mqnode.JoinMerchantNodeFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

func configMerchantNode(c *cli.Context) error {
	var (
		err error
	)

	isInit := c.Bool("init")

	if isInit {
		err = mqnode.ConfigMerchantFlow()
	} else {
		err = mqnode.ShowMechantConfig()
	}

	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}


