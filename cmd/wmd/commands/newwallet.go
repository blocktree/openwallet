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
	"gopkg.in/urfave/cli.v1"
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
				//初始化钱包
				Name:      "config",
				Usage:     "config a currency account",
				ArgsUsage: "<symbol>",
				Action:    walletConfig,
				Category:  "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
					utils.InitFlag,
				},
				Description: `
	wmd wallet config -s <symbol>

This command will init the wallet node.

	`,
			},
			{
				//获取钱包列表信息
				Name:     "list",
				Usage:    "Get all wallet information",
				Action:   getWalletList,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
			},
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
			{
				//批量创建地址
				Name:     "batchaddr",
				Usage:    "Create batch address for wallet",
				Action:   batchAddress,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
					utils.BatchFlag,
				},
				Description: `
	wmd wallet batchaddr -s <symbol>

This command will create batch address for your given wallet id.

	`,
			},

			{
				//启动定时器汇总钱包
				Name:     "startsum",
				Usage:    "Start a timer to sum wallet balance",
				Action:   startSummary,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
					utils.BatchFlag,
				},
				Description: `
	wmd wallet startsum -s ada

This command will Start a timer to sum wallet balance.
When the total balance over the threshold, wallet will send money
to a sum address.

	`,
			},
			{
				//发起交易单
				Name:     "transfer",
				Usage:    "Select a wallet transfer assets",
				Action:   sendTransaction,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd wallet transfer -s <symbol>

This command will transfer the coin.

	`,
			},
			{
				//备份钱包私钥
				Name:     "backup",
				Usage:    "Backup wallet key in filePath: ./data/<symbol>/key/",
				Action:   backupWalletKey,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
				Description: `
	wmd wallet backup -s ada

This command will Backup wallet key in filePath: ./data/<symbol>/key/.

	`,
			},
			{
				//恢复钱包
				Name:     "restore",
				Usage:    "Restore a wallet by backup data",
				Action:   restoreWallet,
				Category: "WALLET COMMANDS",
				Flags: []cli.Flag{
					utils.SymbolFlag,
				},
			},
		},
	}
)

//walletConfig 钱包配置
func walletConfig(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}

	isInit := c.Bool("init")

	if isInit {
		return m.InitConfigFlow()
	} else {
		return m.ShowConfig()
	}
}

//createNewWallet 创建新钱包
func createNewWallet(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}

	err := m.CreateWalletFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

//batchAddress 为钱包创建批量地址
func batchAddress(c *cli.Context) error {

	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	//为钱包创建批量地址
	err := m.CreateAddressFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

//startSummary 启动汇总定时器
func startSummary(c *cli.Context) error {

	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	err := m.SummaryFollow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err

}

//backupWalletKey 备份钱包
func backupWalletKey(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	err := m.BackupWalletFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

//getWalletList 获取钱包列表信息
func getWalletList(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	err := m.GetWalletList()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

//sendTransaction 发起交易单
func sendTransaction(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	err := m.TransferFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}

//restoreWallet 恢复钱包
func restoreWallet(c *cli.Context) error {
	symbol := c.String("symbol")
	if len(symbol) == 0 {
		log.Error("Argument -s <symbol> is missing")
		return nil
	}
	m, ok := assets.GetWMD(symbol).(assets.WalletManager)
	if !ok {
		log.Error(symbol, " wallet manager is not registered!")
		return nil
	}
	err := m.RestoreWalletFlow()
	if err != nil {
		log.Error("unexpected error: ", err)
	}
	return err
}
