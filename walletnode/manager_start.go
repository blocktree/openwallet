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
	"errors"
	"fmt"
	"log"
	"time"

	"docker.io/go-docker/api/types"
)

// StartWalletnode start walletnode
func (w *WalletnodeManager) StartWalletnode(symbol string) error {

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

	cnf := getFullnodeConfig(symbol)
	if cnf == nil {
		// Config not exist!
		err := errors.New("Wallet fullnode config no found")
		log.Println(err)
		return err
	}

	// Action within client
	if err := c.ContainerStart(context.Background(), cName, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Encrypt wallet fullnode if required
	//
	// Workflow:
	//	1. Encrypt
	// 	2. Start container
	if cnf.isEncrypted() && WNConfig.walletnodeIsEncrypted != "true" {
		fmt.Println("\nAttention! Wallet fullnode will be encrypted by default password same with habitus ...")

		time.Sleep(time.Second * 10)

		res, err := c.ContainerExecCreate(context.Background(), cName, types.ExecConfig{Cmd: cnf.ENCRYPT})
		if err != nil {
			log.Println(err)
			return err
		}

		if err := c.ContainerExecStart(context.Background(), res.ID, types.ExecStartCheck{}); err != nil {
			log.Println(err)
			return err
		}

		// update .ini to set isencrypted=true
		WNConfig.walletnodeIsEncrypted = "true"
		if updateConfig(symbol); err != nil {
			log.Println(err)
			return err
		}

		time.Sleep(time.Second * 3)

		if err := c.ContainerStart(context.Background(), cName, types.ContainerStartOptions{}); err != nil {
			return err
		}

		fmt.Printf("\t Encrypt success!\n\n")
	}

	return nil
}
