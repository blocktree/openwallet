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

package utils

import (
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
	"os"
	"fmt"

)

const (
	VersionMajor = 1          // Major version component of the current release
	VersionMinor = 0          // Minor version component of the current release
	VersionPatch = 0          // Patch version component of the current release
	VersionMeta  = "unstable" // Version metadata to append to the version string
)

// Version holds the textual version string.
var Version = func() string {
	v := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()

var (
	CommandHelpTemplate = `{{.cmd.Name}}{{if .cmd.Subcommands}} command{{end}}{{if .cmd.Flags}} [command options]{{end}} [arguments...]
{{if .cmd.Description}}{{.cmd.Description}}
{{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.cmd.Name}}{{with .cmd.ShortName}}, {{.cmd}}{{end}}{{ "\t" }}{{.cmd.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = CommandHelpTemplate
}

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = Version
	if len(gitCommit) >= 8 {
		app.Version += "-" + gitCommit[:8]
	}
	app.Usage = usage
	return app
}


var (

	AppNameFlag = cli.StringFlag{
		Name: "name",
		Usage: "Application Name",
	}

	SymbolFlag = cli.StringFlag{
		Name: "symbol, s",
		Usage: "Currency symbol",
	}

	BatchFlag = cli.BoolFlag{
		Name: "batch",
		Usage: "Create address with batch",
	}

	InitFlag = cli.BoolFlag{
		Name: "i, init",
		Usage: "Init operate",
	}

	LogDirFlag = cli.StringFlag{
		Name: "logdir",
		Usage: "log files directory",
	}

	LogDebugFlag = cli.BoolFlag{
		Name: "debug",
		Usage: "print debug log info",
	}

	MQUrl = cli.StringFlag{
		Name: "mqu, mqurl",
		Usage: "config mq's url",
	}

	MQAccount = cli.StringFlag{
		Name: "mqa, mqaccount",
		Usage: "config mq's account",
	}

	MQPassword = cli.StringFlag{
		Name: "mqp, mqpassword",
		Usage: "config mq's password",
	}

	MQExchange = cli.StringFlag{
		Name: "mqe, mqexchange",
		Usage: "config mq's exchange",
	}

	EnableBlockScan = cli.StringFlag{
		Name: "EnableBlockScan",
		Usage: "start the block scan",
	}

	ISTestNet = cli.StringFlag{
		Name: "is_test_net",
		Usage: "start the test net",
	}
)