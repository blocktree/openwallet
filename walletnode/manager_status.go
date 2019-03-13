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
	"context"
)

// GetWalletnodeStatus get walletnode status
func (w *WalletnodeManager) GetWalletnodeStatus(symbol string) (status string, err error) {

	if err := loadConfig(symbol); err != nil {
		return "", err
	}

	// Init docker client
	c, err := getDockerClient(symbol)
	if err != nil {
		return "", err
	}

	// Instantize parameters
	cname, err := getCName(symbol) // container name
	if err != nil {
		return "", err
	}

	// Action within client
	res, err := c.ContainerInspect(context.Background(), cname)
	if err != nil {
		return "", err
	}

	// Get results
	status = res.State.Status
	return status, nil
}
