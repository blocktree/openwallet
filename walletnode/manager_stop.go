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

	cnf := getFullnodeConfig(symbol)
	if cnf == nil {
		return errors.New("Fullnode config no found!")
	}

	if cnf.STOPCMD == nil {
		fmt.Printf("\n> Stop container by Docker signal(Docker Signal)... \n\n")

		d := time.Duration(3000)
		err = c.ContainerStop(context.Background(), cName, &d)
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 3)

	} else {
		fmt.Printf("\n> Stop container by Command from Service(Stop Command)... \n\n")

		exec := types.ExecConfig{Cmd: cnf.STOPCMD, AttachStderr: true, AttachStdin: true, AttachStdout: true}
		if res, err := c.ContainerExecCreate(context.Background(), cName, exec); err != nil {
			log.Println(err)
			return err
		} else {
			if err := c.ContainerExecStart(context.Background(), res.ID, types.ExecStartCheck{}); err != nil {
				log.Println(err)
				return err
			}

			time.Sleep(time.Second * 10)

		}
	}

	if status, err := w.GetWalletnodeStatus(symbol); err != nil {
		log.Println(err)
	} else {
		fmt.Printf("\nStop container finished, check current container status: %s\n", status)
		if status == "running" {
			fmt.Printf("\n!!!May wait for more seconds, and please check <wmd node logs> returns to confirm finally!\n\n")
		}
	}

	return nil
}
