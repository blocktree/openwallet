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
	// "docker.io/go-docker/api"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/network"
	"fmt"
	// "log"
	// "net/http"
	"strings"
	"time"
)

type NodeManagerStruct struct{}

func (w *NodeManagerStruct) GetNodeStatus(symbol string) error {
	// func(vals ...interface{}) {}(
	// 	context.Background, log.New, http.Client{}, time.Saturday,
	// 	docker.NewClient, types.ContainerListOptions{},
	// 	container.Config{}, network.NetworkingConfig{},
	// 	api.DefaultVersion, strings.ToLower("jij"),
	// ) // Delete before commit

	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Action within client
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	res, err := c.ContainerInspect(ctx, cName)
	if err != nil {
		return err
	}
	// Get results
	status := res.State.Status
	fmt.Printf("%s walletnode status: %s\n", symbol, status)
	return err
}

// Create a new container for wallet fullnode
// First: check if node is exist:
//		- yes to return "existing"
//		- or else to create new
func (w *NodeManagerStruct) CreateNodeFlow(symbol string) error {
	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Instantize parameters
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	// Check if exist
	if res, err := c.ContainerInspect(ctx, cName); err == nil {
		// Exist
		status := res.State.Status
		fmt.Printf("%s walletnode exist: %s\n", symbol, status)
		return nil
	}
	// Action within client
	containerConfig := container.Config{
		Hostname:   "bch",
		Domainname: "bch.local",
		// Cmd:        "ls",
		Image: "ubuntu:latest",
	}
	hostConfig := container.HostConfig{}
	networkingConfig := network.NetworkingConfig{}
	_, err = c.ContainerCreate(ctx, &containerConfig, &hostConfig, &networkingConfig, cName)
	if err == nil {
		fmt.Printf("%s walletnode created in success!\n", symbol)
	}
	return err
}

func (w *NodeManagerStruct) StartNodeFlow(symbol string) error {
	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Action within client
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	err = c.ContainerStart(ctx, cName, types.ContainerStartOptions{})
	if err == nil {
		fmt.Printf("%s walletnode start in success!\n", symbol)
	}

	return err
}

func (w *NodeManagerStruct) StopNodeFlow(symbol string) error {
	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Action within client
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	d := time.Duration(3000)
	err = c.ContainerStop(ctx, cName, &d)
	if err == nil {
		fmt.Printf("%s walletnode stop in success!\n", symbol)
	}
	return err
}

func (w *NodeManagerStruct) RestartNodeFlow(symbol string) error {
	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Action within client
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	err = c.ContainerRestart(ctx, cName, nil)
	if err == nil {
		fmt.Printf("%s walletnode stop in success!\n", symbol)
	}
	return err
}

func (w *NodeManagerStruct) RemoveNodeFlow(symbol string) error {
	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return (err)
	}
	// Action within client
	cName := strings.ToLower(symbol) // container name
	ctx := context.Background()      // nil
	err = c.ContainerRemove(ctx, cName, types.ContainerRemoveOptions{})
	if err == nil {
		fmt.Printf("%s walletnode remove in success!\n", symbol)
	}
	return err
}
