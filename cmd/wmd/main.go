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

package main

import (
	"os"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"sort"
	"github.com/blocktree/OpenWallet/cmd/utils"
	"github.com/blocktree/OpenWallet/cmd/wmd/commands"
)


const (
	clientIdentifier = "wmd" // Client identifier to advertise over the network
	version = "0.1.1"
)


var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the Wallet Manager Driver command line interface")
	//flags that configure the node
	nodeFlags = []cli.Flag{
		utils.AppNameFlag,
	}

)

func init() {
	// Initialize the CLI app and start openw
	app.Action = wmd
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2018 The OpenWallet Authors"
	app.Version = version
	app.Commands = []cli.Command{
		commands.CmdWallet,
		commands.CmdConfig,
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
