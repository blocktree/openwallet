/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package main

import (
	"fmt"
	"github.com/blocktree/openwallet/v2/cmd/utils"
	"github.com/blocktree/openwallet/v2/cmd/wmd/commands"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)

const (
	clientIdentifier = "wmd" // Client identifier to advertise over the network
)

var (
	// The app that holds all commands and flags.
	app = utils.NewApp(commands.GitRev, "the Wallet Manager Driver command line interface")
)

func init() {
	// Initialize the CLI app and start openw
	app.Name = "wmd"
	app.Action = wmd
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2019 The openwallet Authors"
	app.Version = commands.Version
	app.Commands = []cli.Command{
		commands.CmdWallet,
		commands.CmdVersion,
		//commands.CmdNode,
		//commands.CmdConfig,
		//commands.CmdMerchant,
	}
	app.Flags = []cli.Flag{
		utils.AppNameFlag,
		utils.LogDirFlag,
		utils.LogDebugFlag,
	}

	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//wmd is a util to manager multi currency symbol wallet
func wmd(ctx *cli.Context) error {

	return nil
}
