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

package walletnode

import (
	"errors"
	"fmt"
	"strings"

	// "docker.io/go-docker/api"

	sh "github.com/codeskyblue/go-sh"
)

// LogsWalletnode watch logs now
func (w *WalletnodeManager) LogsWalletnode(symbol string) error {

	if err := loadConfig(symbol); err != nil {
		return err
	}

	cName, err := getCName(symbol) // container name
	if err != nil {
		return err
	}

	cnf := getFullnodeConfig(symbol)
	if cnf == nil {
		return errors.New("Wallet fullnode configs can not found")
	}

	logfile := cnf.getLogFile()
	if logfile == "" {
		return errors.New("Logfile no found")
	}

	host := ""
	if WNConfig.walletnodeServerType == "docker" {
		host = fmt.Sprintf("-H %s:%s", WNConfig.walletnodeServerAddr, WNConfig.walletnodeServerPort)
	}

	cmd := fmt.Sprintf("docker %s exec %s tail -f /data/%s", host, cName, logfile)
	cmds := strings.Split(cmd, " ")
	session := sh.Command(cmds[0], cmds[1:])
	return session.Run()
}
