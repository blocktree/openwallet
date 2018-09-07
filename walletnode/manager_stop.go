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
	"time"
)

func (w *WalletnodeManager) StopWalletnode(symbol string) error {

	if err := loadConfig(symbol); err != nil {
		return err
	}

	// Init docker client
	c, err := getDockerClient(symbol)
	if err != nil {
		return err
	}

	cName, err := getCName(symbol) // container name
	if err != nil {
		return err
	}

	d := time.Duration(3000)
	err = c.ContainerStop(context.Background(), cName, &d)
	if err != nil {

		return err
	}
	return nil
}
