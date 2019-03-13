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

package commands

import (
	"gopkg.in/urfave/cli.v1"
	"fmt"
	"github.com/blocktree/openwallet/cmd/utils"
)

var (
	CmdNew = cli.Command{
		Name:      "new",
		Usage:     "Creates a openwallet application",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Creates a openwallet application for the given app name in the current directory.

  The command 'new' creates a folder named [appname] and generates the following structure:

`,
		Flags: []cli.Flag{
			utils.AppNameFlag,
		},
		Action: createNewApp,
	}

)

//createNewApp 创建新应用
func createNewApp(c *cli.Context) error {
	if len(c.Args()) != 1 {
		//log.Log.Fatal("Argument [appname] is missing")
	}
	//读取第一个参数作为应用名
	fmt.Printf("New App %v\n", c.Args().Get(0))
	return nil
}
