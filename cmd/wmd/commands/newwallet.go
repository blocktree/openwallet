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
	"github.com/blocktree/OpenWallet/assets/cardano"
)

var (
	// 钱包命令
	CmdWallet = cli.Command{
		Name:      "wallet",
		Usage:     "Manage multi currency wallet",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
You create, import, restore wallet

`,
		Subcommands: []cli.Command{
			{
				//创建钱包
				Name:      "new",
				Usage:     "new a currency wallet",
				ArgsUsage: "<symbol>",
				Action:    createNewWallet,
				Category:  "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd wallet new -s <symbol>

This command will start the wallet node, and create new wallet.

	`,
			},
		},
	}

	// 写入命令
	CmdWrite = cli.Command{
		Name:      "write, w",
		Usage:     "write something",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
write something into file

`,
		Action:    writeSomething,
	}
)

//createNewWallet 创建新钱包
func createNewWallet(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		openwLogger.Log.Fatal("Argument -s <symbol> is missing")
	}

	//创建钱包
	err := cardano.CreateNewWalletFlow()
	if err != nil {
		openwLogger.Log.Fatalf("%v", err)
	}
	return err
}

func writeSomething() {
	cardano.WriteSomething()
}
