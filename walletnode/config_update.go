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
	"fmt"
	"path/filepath"
	s "strings"

	bconfig "github.com/astaxie/beego/config"
	"github.com/pkg/errors"
)

// Update <Symbol>.ini file
func updateConfig(symbol string) error {
	if WNConfig == nil {
		return errors.New("getDockerClient: WalletnodeConfig does not initialized!")
	}

	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"
	absFile := filepath.Join(configFilePath, configFileName)

	c, err := bconfig.NewConfig("ini", absFile)
	if err != nil {
		return (errors.New(fmt.Sprintf("Load Config Failed: %s", err)))
	}

	err = c.Set("isTestNet", WNConfig.isTestNet)
	err = c.Set("rpcuser", WNConfig.rpcUser)
	err = c.Set("rpcpassword", WNConfig.rpcPassword)
	err = c.Set("apiurl", WNConfig.apiURL)
	err = c.Set("mainnetdatapath", WNConfig.mainNetDataPath)
	err = c.Set("testnetdatapath", WNConfig.testNetDataPath)

	err = c.Set("walletnode::walletnodeServerType", WNConfig.walletnodeServerType)
	err = c.Set("walletnode::walletnodeServerAddr", WNConfig.walletnodeServerAddr)
	err = c.Set("walletnode::walletnodeServerSocket", WNConfig.walletnodeServerSocket)

	if err := c.SaveConfigFile(absFile); err != nil {
		return err
	}

	return nil
}
