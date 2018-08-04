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
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
)

func (w *NodeManagerStruct) StartNodeFlow(symbol string) error {

	if err := loadConfig(symbol); err != nil {
		return err
	}

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		return err
	}
	// Action within client
	cName, err := _GetCName(symbol) // container name
	if err != nil {
		return err
	}
	err = c.ContainerStart(context.Background(), cName, types.ContainerStartOptions{})
	if err == nil {
		fmt.Printf("%s walletnode start in success!\n", symbol)
	}
	return err
}
