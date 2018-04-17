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

import "gopkg.in/urfave/cli.v1"

var (
	CmdRun = cli.Command{
		Name:      "run",
		Usage:     "Creates a OpenWallet application",
		ArgsUsage: "",
		Category:  "Application COMMANDS",
		Description: `
Creates a OpenWallet application for the given app name in the current directory.

  The command 'new' creates a folder named [appname] and generates the following structure:

            ├── main.go
            ├── {{"conf"|foldername}}
            │     └── app.conf
            ├── {{"controllers"|foldername}}
            │     └── default.go
            ├── {{"models"|foldername}}
            ├── {{"routers"|foldername}}
            │     └── router.go
            ├── {{"tests"|foldername}}
            │     └── default_test.go
            ├── {{"static"|foldername}}
            │     └── {{"js"|foldername}}
            │     └── {{"css"|foldername}}
            │     └── {{"img"|foldername}}
            └── {{"views"|foldername}}
                  └── index.tpl

`,
		Action:  func(c *cli.Context) error {
			return nil
		},
	}

)