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
	"docker.io/go-docker"
	"fmt"
	"log"
	s "strings"
)

type NodeManagerStruct struct{}

func _GetClient() (*docker.Client, error) {
	var c *docker.Client
	var err error

	// Init docker client
	if _, ok := map[string]string{"localhost": "", "127.0.0.1": ""}[serverAddr]; ok {
		c, err = docker.NewEnvClient()
	} else {
		host := fmt.Sprintf("tcp://%s:%s", serverAddr, serverPort)
		c, err = docker.NewClient(host, "v1.37", nil, map[string]string{})
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return c, err
}

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
	cName, err := _GetCName(symbol) // container name
	if err != nil {
		return err
	}
	// Action within client
	res, err := c.ContainerInspect(context.Background(), cName)
	if err != nil {
		return err
	}
	// Get results
	status := res.State.Status
	fmt.Printf("%s walletnode status: %s\n", s.ToUpper(symbol), status)
	return nil
}
