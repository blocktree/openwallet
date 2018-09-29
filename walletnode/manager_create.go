/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 * * The OpenWallet library is free software: you can redistribute it and/or modify
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
	"path/filepath"
	s "strings"
	"time"

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/mount"
	"docker.io/go-docker/api/types/network"
	"github.com/blocktree/OpenWallet/log"
	"github.com/docker/go-connections/nat"
)

// Check wallet container, create if not
//
// Pre-requirement
//		INI file exists!
//
// Workflow:
//		if 钱包容器存在:
//			return nil
//		else 钱包容器不存在 or 坏了，创建：
//			1> 初始化物理服务器目录
//			return {IP, Status}
//
func (wn *WalletnodeManager) CheckAdnCreateContainer(symbol string) error {

	if err := loadConfig(symbol); err != nil {
		return err
	}

	WNConfig.walletnodeIsEncrypted = "false"

	// Init docker client
	c, err := getDockerClient(symbol)
	if err != nil {
		log.Error(err)
		return err
	}

	cName, err := getCName(symbol) // container name
	if err != nil {
		return err
	}

	// Check if exist
	if res, err := c.ContainerInspect(context.Background(), cName); err == nil {
		// Exist, return
		fmt.Printf("%s walletnode exist: %s\n", s.ToUpper(symbol), res.State.Status)
		return nil
	}

	// None exist: Create action within client c
	if err = c.ContainerRemove(context.Background(), "temp", types.ContainerRemoveOptions{}); err == nil {
		fmt.Println(time.Duration(10))
		// return err
	}
	if _, err = c.ContainerCreate( // 创建目录
		context.Background(),
		&container.Config{
			Image: "ubuntu:latest",
			Cmd:   []string{"/bin/sh", "-c", fmt.Sprintf("mkdir -p %s/%s/data %s/%s/testdata", MountSrcPrefix, symbol, MountSrcPrefix, symbol)}},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: "/openwallet", Target: "/openwallet", ReadOnly: false, BindOptions: &mount.BindOptions{Propagation: "private"}},
			}},
		&network.NetworkingConfig{},
		"temp"); err != nil {
		log.Error(err)
		return err
	} else {
		// Start
		if err = c.ContainerStart(context.Background(), "temp", types.ContainerStartOptions{}); err != nil {
			log.Error(err)
		}
		if err = c.ContainerRemove(context.Background(), "temp", types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Error(err)
		}
	}

	// --------------------------------- Create Container -----------------------------------------
	var portBindings map[nat.Port][]nat.PortBinding
	// portBindings := map[nat.Port][]nat.PortBinding{}
	// var exposedPorts map[nat.Port]struct{}
	// exposedPorts := map[nat.Port]struct{}{}
	var RPCPort string
	// var Cmd []string
	var Env []string
	var MountSrcDir string

	ctn_config, ok := FullnodeContainerConfigs[s.ToLower(symbol)]
	if !ok {
		return nil
	}

	portBindings = map[nat.Port][]nat.PortBinding{}
	for _, v := range ctn_config.PORT {
		if WNConfig.isTestNet == "true" {
			portBindings[nat.Port(v[0])] = []nat.PortBinding{nat.PortBinding{HostIP: DockerAllowed, HostPort: v[2]}}
			//exposedPorts[nat.Port(v[0])] = struct{}{}
			if v[0] == ctn_config.APIPORT[0] {
				RPCPort = v[2]
			}
		} else {
			portBindings[nat.Port(v[0])] = []nat.PortBinding{nat.PortBinding{HostIP: DockerAllowed, HostPort: v[1]}}
			// exposedPorts[nat.Port(v[0])] = struct{}{}
			if v[0] == ctn_config.APIPORT[0] {
				RPCPort = v[1]
			}
		}
	}

	if WNConfig.isTestNet == "true" {
		Env = []string{"TESTNET=true"}
		MountSrcDir = filepath.Join(MountSrcPrefix, s.ToLower(symbol), "/testdata")
	} else {
		Env = []string{"TESTNET=false"}
		MountSrcDir = filepath.Join(MountSrcPrefix, s.ToLower(symbol), "/data")
	}

	cConfig := container.Config{
		// string to container name
		Hostname: cName,
		// string to container domainname
		// Domainname: fmt.Sprintf("%s.local.com", cName),
		// string, Command to run when starting the container
		// Cmd: []string,
		// string, List of environment variable to set in the container
		Env: Env,
		// string, Name of the image as it was passed by the operator(e.g. could be symbolic)
		Image: ctn_config.IMAGE,
		// nat.PortSet         `json:",omitempty"` , List of exposed ports
		// ExposedPorts: exposedPorts,
		// string, Current directory (PWD) in the command will be launched
		// WorkingDir: "/root",

		// // map[string]struct{}, List of volumes (mounts) used for the container
		// Volumes: map[string]struct{}{
		// 	// "type":        "volume",
		// 	//"Name":        "OpenWallet Fullnode Data",
		// 	// "Source": "/tmp",
		// 	// "Destination": "/tmp",
		// 	// "Driver":      "local",
		// 	// "Mode":        "ro,Z",
		// 	// "RW":          true,
		// 	// "Propagation": "",
		// 	// "vers":        "4,soft,timeo=180,bg,tcp,rw",
		// },
		// // strslice.StrSlice, Entrypoint to run when starting the container
		// Entrypoint,
		// // map[string]string, List of labels set to this container
		// Labels: map[string]string{},
	}

	hConfig := container.HostConfig{
		// Network mode to use for the container, default "bridge"
		NetworkMode: "bridge",
		// map[nat.Port][]nat.PortBinding
		PortBindings: portBindings,
		// Mount Volumes
		Mounts: []mount.Mount{
			{
				// "Driver":      "local",
				// "vers":        "4,soft,timeo=180,bg,tcp,rw",
				Type:        mount.TypeBind,
				Source:      MountSrcDir,
				Target:      MountDstDir,
				ReadOnly:    false,
				BindOptions: &mount.BindOptions{Propagation: "private"},
			},
		},
	}

	// endpointSetting := network.EndpointSettings{
	// 	// // Configurations
	// 	// IPAMConfig, // *EndpointIPAMConfig
	// 	// Links,      // []string
	// 	// Aliases,    // []string
	// 	// // Operational data
	// 	// NetworkID,           // string
	// 	// EndpointID,          // string
	// 	// Gateway,             // string
	// 	// IPAddress,           // string
	// 	// IPPrefixLen,         // int
	// 	// IPv6Gateway,         // string
	// 	// GlobalIPv6Address,   // string
	// 	// GlobalIPv6PrefixLen, // int
	// 	// MacAddress,          // string
	// 	// DriverOpts,          // map[string]string
	// }

	nConfig := network.NetworkingConfig{
		// map[string]*EndpointSettings // Endpoint configs for each connecting network
		// EndpointsConfig: map[string]*network.EndpointSettings{"endporint": &endpointSetting},
	}

	if ctn, err := c.ContainerCreate(context.Background(), &cConfig, &hConfig, &nConfig, cName); err != nil {
		log.Error(err)
		return err
	} else {
		fmt.Println(ctn)
		fmt.Printf("%s walletnode created in success!\n", symbol)
	}

	// // Get exposed port
	// apiPort := string("")
	// if res, err := c.ContainerInspect(context.Background(), cName); err != nil {
	// 	log.Error(err)
	// 	return err
	// } else {
	// 	fmt.Println(res.NetworkSettings)
	// 	fmt.Println(res.NetworkSettings.Ports)
	// 	if v, ok := res.NetworkSettings.Ports["18332/tcp"]; ok {
	// 		apiPort = v[0].HostPort
	// 		fmt.Println("apiPort = ", apiPort)
	// 	} else {
	// 		log.Error("No apiPort loaded!")
	// 		return errors.New("No apiPort loaded!")
	// 	}
	// }

	// Get info from docker inspect for fullnode API
	WNConfig.WalletURL = fmt.Sprintf("http://%s:%s", WNConfig.walletnodeServerAddr, RPCPort)
	return nil
}
