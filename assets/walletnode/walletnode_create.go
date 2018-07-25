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
	// "docker.io/go-docker"
	// "github.com/astaxie/beego"
	"errors"
	"log"
	// "docker.io/go-docker/api"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/mount"
	"docker.io/go-docker/api/types/network"
	"fmt"
	"github.com/blocktree/OpenWallet/console"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	s "strings"
	"time"
)

// Check <Symbol>.ini file, create new if not
//
// Workflow:
//		1> 当前目录没有 ini，是否创建？
//			1.1 存在，return nil
//		2> 询问是否设置为测试链？
//		3> 获取Master服务器IP地址和端口
func _CheckAndCreateConfig(symbol string) error {

	// Init
	// if err := FullnodeContainerPath.init(symbol); err != nil {
	// 	return err
	// }

	// Check <Symbol>.ini
	if err := loadConfig(symbol); err == nil {
		// <Symbol>.ini exist, return and go next
		return nil
	} else {
		fmt.Println("Current: ", err)
	}

	// Ask about whether create new
	dirname, _ := filepath.Abs("./")
	if isnew, err := console.InputText(fmt.Sprintf("Init new %s wallet fullnode in '%s/'(yes, no)[yes]: ", s.ToUpper(symbol), dirname), false); err != nil {
		log.Println(err)
		return err
	} else {
		switch isnew {
		case "", "yes":
		case "no":
			return errors.New("Init terminated!")
		default:
			return errors.New("Invalid!")
		}
	}

	// Ask about whether sync by testnet
	if istestnet, err := console.InputText("Within testnet('testnet','main')[testnet]: ", false); err != nil {
		return err
	} else {
		switch istestnet {
		case "testnet":
			isTestNet = "true"
		case "main":
			isTestNet = "false"
		case "":
			isTestNet = "true"
		default:
			return errors.New("Invalid!")
		}
	}

	// Ask about Docker master Address and Port
	if addr, err := console.InputText("Docker master server address [localhost]: ", false); err != nil {
		return err
	} else {
		if addr != "" {
			serverAddr = addr
		}
	}
	if port, err := console.InputText("Docker master server port[2375]: ", false); err != nil {
		return err
	} else {
		if port != "" {
			serverPort = port
		}
	}

	// Create new INI file, and update
	if err := initConfig(symbol); err != nil {
		log.Println(err)
		return err
	}
	if err := updateConfig(symbol); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

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
func _CheckAdnCreateContainer(symbol string) error {
	symbol = s.ToLower(symbol)

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		log.Println(err)
		return err
	}

	if err := loadConfig(symbol); err != nil {
		log.Println(err)
		return err
	}

	cName, err := _GetCName(symbol) // container name
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
			Cmd:   []string{"/bin/sh", "-c", fmt.Sprintf("mkdir -p /openwallet/data/%s", symbol)}},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: "/openwallet", Target: "/openwallet", ReadOnly: false, BindOptions: &mount.BindOptions{Propagation: "private"}},
			}},
		&network.NetworkingConfig{},
		"temp"); err != nil {
		log.Println(err)
		return err
	} else {
		// Start
		if err = c.ContainerStart(context.Background(), "temp", types.ContainerStartOptions{}); err != nil {
			log.Println(err)
		}
		// Stop
		d := time.Duration(3000)
		if err = c.ContainerStop(context.Background(), "temp", &d); err != nil {
			log.Println(err)
		}
		// Remove
		if err = c.ContainerRemove(context.Background(), "temp", types.ContainerRemoveOptions{}); err != nil {
			log.Println(err)
		}
	}

	ctn_config, ok := FullnodeContainerConfigs[s.ToLower(symbol)]
	if !ok {
		return nil
	}

	// var portBindings map[nat.Port][]nat.PortBinding
	portBindings := map[nat.Port][]nat.PortBinding{}
	// var exposedPorts map[nat.Port]struct{}
	exposedPorts := map[nat.Port]struct{}{}
	apiPort := string("")
	for _, v := range ctn_config.PORT {
		// portBindings[nat.Port(v[0])] = []nat.PortBinding{nat.PortBinding{HostPort: v[1]}}
		portBindings[nat.Port(v[0])] = []nat.PortBinding{nat.PortBinding{}}
		exposedPorts[nat.Port(v[0])] = struct{}{}
		if v[0] == ctn_config.APIPORT {
			apiPort = v[1]
		}
	}
	var Cmd []string
	if isTestNet == "true" {
		Cmd = ctn_config.CMD[1]
	} else {
		Cmd = ctn_config.CMD[0]
	}
	cConfig := container.Config{
		// string,
		Hostname: cName,
		// string,
		Domainname: fmt.Sprintf("%s.local.com", cName),
		// string, Command to run when starting the container
		Cmd: Cmd,
		// string, Name of the image as it was passed by the operator(e.g. could be symbolic)
		Image: ctn_config.IMAGE,
		// nat.PortSet         `json:",omitempty"` , List of exposed ports
		ExposedPorts: exposedPorts,
		// string, Current directory (PWD) in the command will be launched
		WorkingDir: "/root",

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
			// {
			// 	// "Driver":      "local",
			// 	// "vers":        "4,soft,timeo=180,bg,tcp,rw",
			// 	Type:        mount.TypeBind,
			// 	Source:      "/openwallet/exec/" + s.ToLower(symbol),
			// 	Target:      "/exec",
			// 	ReadOnly:    true,
			// 	BindOptions: &mount.BindOptions{Propagation: "private"}},
			{
				Type:        mount.TypeBind,
				Source:      "/openwallet/data/" + s.ToLower(symbol),
				Target:      "/data",
				ReadOnly:    false,
				BindOptions: &mount.BindOptions{Propagation: "private"}},
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
		log.Println(err)
		return err
	} else {
		fmt.Println(ctn)
		fmt.Printf("%s walletnode created in success!\n", symbol)
	}

	// Get info from docker inspect for fullnode API
	apiURL = fmt.Sprintf("http://%s:%s", serverAddr, apiPort)

	return nil
}

func _InitConfigFile(symbol string) error {
	// Update config
	if err := updateConfig(symbol); err != nil {
		return err
	}
	return nil
}

// Create a new container for wallet fullnode
//
// Workflow:
//		// 步骤一: 判定本地 .ini 文件是否存在
//		if  .ini 不存在，创建一个默认的 {
//			1> 询问用户配置参数
//			2> 创建初始 .ini 文件
//		} else {									// .ini 存在
//			1> 无操作，进入下一步
//		}
//
//		// 步骤二：判断是否需要创建节点容器
//		if 容器不存在 or 不正常 {
//			1> 删除后，或直接创建一个新的(需：)
//			2> 导出 container 数据(IP, status)
//		} else {									// 正常
//			1> 导出 container 数据(IP, status)
//		}
//
//		// 步骤三
//		1> 根据导出的 container 数据，更新配置文件中关于 container 的项（重复更新，方便用户改错后自动恢复）
func (w *NodeManagerStruct) CreateNodeFlow(symbol string) error {

	// 一:
	if err := _CheckAndCreateConfig(symbol); err != nil {
		return err
	}

	// 二:
	if err := _CheckAdnCreateContainer(symbol); err != nil {
		return err
	}

	// 三:
	if err := _InitConfigFile(symbol); err != nil {
		return err
	}

	return nil
}
