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
	"fmt"
	"log"
	s "strings"
)

func (w *NodeManagerStruct) GetNodeStatus(symbol string) error {
	// func(vals ...interface{}) {}(
	// 	context.Background, log.New, time.Saturday,
	// 	docker.NewClient, types.ContainerListOptions{},
	// 	container.Config{}, network.NetworkingConfig{},
	// 	api.DefaultVersion, s.ToLower("jij"),
	// ) // Delete before commit

	if err := loadConfig(symbol); err != nil {
		log.Println(err)
		return err
	}

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		return err
	}

	// Instantize parameters
	cname, err := _GetCName(symbol) // container name
	if err != nil {
		return err
	}

	// dst := "/openwallet/data/wallet.dat"
	// src := "./eee.dat"
	// if err := CopyToContainer(cname, src, dst); err != nil {
	// 	return err
	// }

	// Action within client
	res, err := c.ContainerInspect(context.Background(), cname)
	if err != nil {
		return err
	}
	// Get results
	status := res.State.Status
	fmt.Printf("%s walletnode status: %s\n", s.ToUpper(symbol), status)
	return nil
}
