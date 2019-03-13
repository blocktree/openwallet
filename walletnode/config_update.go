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
		return errors.New("getDockerClient: WalletnodeConfig does not initialized")
	}

	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"
	absFile := filepath.Join(configFilePath, configFileName)

	c, err := bconfig.NewConfig("ini", absFile)
	if err != nil {
		return fmt.Errorf("Load Config Failed: %s", err)
	}

	err = c.Set("isTestNet", WNConfig.isTestNet)
	err = c.Set("rpcuser", WNConfig.RPCUser)
	err = c.Set("rpcpassword", WNConfig.RPCPassword)
	err = c.Set("WalletURL", WNConfig.WalletURL)

	err = c.Set("walletnode::ServerType", WNConfig.walletnodeServerType)
	err = c.Set("walletnode::ServerAddr", WNConfig.walletnodeServerAddr)
	err = c.Set("walletnode::ServerPort", WNConfig.walletnodeServerPort)
	// err = c.Set("walletnode::ServerSocket", WNConfig.walletnodeServerSocket)
	err = c.Set("walletnode::StartNodeCMD", WNConfig.walletnodeStartNodeCMD)
	err = c.Set("walletnode::StopNodeCMD", WNConfig.walletnodeStopNodeCMD)
	err = c.Set("walletnode::MainnetDataPath", WNConfig.walletnodeMainNetDataPath)
	err = c.Set("walletnode::TestnetDataPath", WNConfig.walletnodeTestNetDataPath)
	err = c.Set("walletnode::IsEncrypted", WNConfig.walletnodeIsEncrypted)

	if err := c.SaveConfigFile(absFile); err != nil {
		return err
	}

	return nil
}
