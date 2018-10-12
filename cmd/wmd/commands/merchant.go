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
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/merchant"
	"gopkg.in/urfave/cli.v1"
)

var (
	// 钱包命令
	CmdMerchant = cli.Command{
		Name:      "merchant",
		Usage:     "Manage merchant node",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Use merchant commands to join merchant server

`,
		Subcommands: []cli.Command{
			{
				//创建或查看本地通信密钥对
				Name:      "keychain",
				Usage:     "create or see node keychain",
				ArgsUsage: "<init>",
				Action:    getMerchantKeychain,
				Category:  "MERCHANT COMMANDS",
				Flags: []cli.Flag{
					utils.InitFlag,
				},
				Description: `
	wmd merchant keychain [-i]

This command will show the local publicKey and merchant publicKey.

	`,
			},
			{
				//加入并连接到商户节点
				Name:     "join",
				Usage:    "Join and connect merchant node",
				Action:   joinMerchantNode,
				Category: "MERCHANT COMMANDS",
				Description: `
	wmd merchant join

This command will connect and join merchant node.

	`,
			},

			{
				//配置商户
				Name:      "config",
				Usage:     "Start a timer to sum wallet balance",
				ArgsUsage: "<init>",
				Action:    configMerchantNode,
				Category:  "MERCHANT COMMANDS",
				Flags: []cli.Flag{
					utils.InitFlag,
				},
				Description: `
	wmd wallet config [i]

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
		err = merchant.InitMerchantKeychainFlow()
	} else {
		err = merchant.GetMerchantKeychain()
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
	utils.SetupLog(logDir, "merchant.log", debug)

	err = merchant.JoinMerchantNodeFlow()
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
		err = merchant.ConfigMerchantFlow()
	} else {
		err = merchant.ShowMechantConfig()
	}

	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}
