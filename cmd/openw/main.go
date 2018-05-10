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
	"github.com/blocktree/OpenWallet/cmd/utils"
	"sort"
	"gopkg.in/urfave/cli.v1"
	"github.com/blocktree/OpenWallet/cmd/openw/commands"
)


const (
	clientIdentifier = "openw" // Client identifier to advertise over the network
)


var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the openwallet command line interface")
	//flags that configure the node
	nodeFlags = []cli.Flag{
		utils.AppNameFlag,
	}

	//rpcFlags = []cli.Flag{
	//	utils.RPCEnabledFlag,
	//	utils.RPCListenAddrFlag,
	//	utils.RPCPortFlag,
	//	utils.RPCApiFlag,
	//	utils.WSEnabledFlag,
	//	utils.WSListenAddrFlag,
	//	utils.WSPortFlag,
	//	utils.WSApiFlag,
	//	utils.WSAllowedOriginsFlag,
	//	utils.IPCDisabledFlag,
	//	utils.IPCPathFlag,
	//}
)

func init() {
	// Initialize the CLI app and start openw
	app.Action = openw
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2018 The openwallet Authors"
	app.Commands = []cli.Command{
		commands.CmdNew,
		commands.CmdRun,
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	//app.Flags = append(app.Flags, rpcFlags...)

	//app.Before = func(ctx *cli.Context) error {
	//	runtime.GOMAXPROCS(runtime.NumCPU())
	//	if err := debug.Setup(ctx); err != nil {
	//		return err
	//	}
	//	// Start system runtime metrics collection
	//	go metrics.CollectProcessMetrics(3 * time.Second)
	//
	//	utils.SetupNetwork(ctx)
	//	return nil
	//}
	//
	//app.After = func(ctx *cli.Context) error {
	//	debug.Exit()
	//	console.Stdin.Close() // Resets terminal mode.
	//	return nil
	//}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//openw is the main entry point into the system if no special subcommand is ran.
//It creates a default node based on the command line arguments and runs it in
//blocking mode, waiting for it to be shut down.
func openw(ctx *cli.Context) error {
	//node := makeFullNode(ctx)
	//startNode(ctx, node)
	//node.Wait()
	return nil
}

