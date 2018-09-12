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
	"log"
	s "strings"

	docker "docker.io/go-docker"
)

var (
	Symbol                   string                              // Current fullnode wallet's symbol
	FullnodeContainerConfigs map[string]*FullnodeContainerConfig // All configs of fullnode wallet for Docker
	WNConfig                 *WalletnodeConfig
)

type WalletnodeManager struct{}

// Get docker client
func getDockerClient(symbol string) (c *docker.Client, err error) {

	symbol = s.ToLower(symbol)

	if WNConfig == nil {
		return nil, errors.New("getDockerClient: WalletnodeConfig does not initialized!")
	}

	// Init docker client
	//walletnodeServerType
	if WNConfig.walletnodeServerType == "docker" {

		if WNConfig.walletnodeServerAddr == "127.0.0.1" || WNConfig.walletnodeServerAddr == "localhost" {
			c, err = docker.NewEnvClient()
		} else {
			host := fmt.Sprintf("tcp://%s:%s", WNConfig.walletnodeServerAddr, WNConfig.walletnodeServerPort)
			c, err = docker.NewClient(host, "v1.37", nil, map[string]string{})
		}
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return c, err
}

// Private function, generate container name by <Symbol> and <isTestNet>
func getCName(symbol string) (string, error) {
	if WNConfig == nil {
		return "", errors.New("getCName: WalletnodeConfig does not initialized!")
	}

	// Within testnet, use "<Symbol>_testnet" as container name
	if WNConfig.isTestNet == "true" {
		return s.ToLower(symbol) + "_t", nil
	} else {
		return s.ToLower(symbol), nil
	}
}
