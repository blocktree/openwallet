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
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

var (
	Version   = ""
	GitRev    = ""
	BuildTime = ""
)

var (
	// 钱包命令
	CmdVersion = cli.Command{
		Name:      "version",
		Usage:     "Manage multi currency wallet",
		ArgsUsage: "",
		Action:    version,
		Category:  "VERSION COMMANDS",
	}
)

//walletConfig 钱包配置
func version(c *cli.Context) error {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("GitRev: %s\n", GitRev)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	return nil
}
