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

import "gopkg.in/urfave/cli.v1"

var (
	CmdRun = cli.Command{
		Name:      "run",
		Usage:     "run [appname] [watchall] [-main=*.go] [-downdoc=true]  [-gendoc=true] [-vendor=true] [-e=folderToExclude] [-ex=extraPackageToWatch] [-tags=goBuildTags] [-runmode=BEEGO_RUNMODE]",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Run command will supervise the filesystem of the application for any changes, and recompile/restart it.

`,
		Action:  func(c *cli.Context) error {
			return nil
		},
	}

)