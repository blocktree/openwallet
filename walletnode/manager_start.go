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

	cnf := GetFullnodeConfig(symbol)
	if cnf == nil {
		// Config not exist!
		err := errors.New("Wallet fullnode config no found!")
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

		exec := types.ExecConfig{Cmd: cnf.ENCRYPT}
		if res, err := c.ContainerExecCreate(context.Background(), cName, exec); err != nil {
			log.Println(err)
			return err
		} else {
			if err := c.ContainerExecStart(context.Background(), res.ID, types.ExecStartCheck{}); err != nil {
				log.Println(err)
				return err
			} else {
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
			}
		}

		fmt.Printf("\t Encrypt success!\n\n")
	}

	return nil
}

// type ExecConfig struct {
// 	User         string   // User that will run the command
// 	Privileged   bool     // Is the container in privileged mode
// 	Tty          bool     // Attach standard streams to a tty.
// 	AttachStdin  bool     // Attach the standard input, makes possible user interaction
// 	AttachStderr bool     // Attach the standard error
// 	AttachStdout bool     // Attach the standard output
// 	Detach       bool     // Execute in detach mode
// 	DetachKeys   string   // Escape keys for detach
// 	Env          []string // Environment variables
// 	Cmd          []string // Execution commands and args
// }
