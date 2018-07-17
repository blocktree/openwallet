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
	// "errors"
	// "docker.io/go-docker/api"
	// "docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/network"
	"fmt"
	"github.com/blocktree/OpenWallet/console"
	"path/filepath"
	s "strings"
)

// Check <Symbol>.ini file, create new if not
//
// Workflow:
//		1> 当前目录没有 ini，是否创建？
//		2> 是否设置为测试链？
//		3> 服务器IP地址和端口
func _CheckAndCreateConfig(symbol string) error {
	var isNew bool

	// Check <Symbol>.ini
	if err := loadConfig(symbol); err == nil {
		// <Symbol>.ini exist, return and go next
		return nil
	} else {
		fmt.Println("Current: ", err)
	}

	// Ask about whether create new
	dirname, _ := filepath.Abs("./")
	if isnew, err := console.InputText(fmt.Sprintf("Init new %s wallet fullnode in '%s/' (1: Yes, 0: No) [default: 1]: ", s.ToUpper(symbol), dirname), false); err != nil {
		return err
	} else {
		switch isnew {
		case "1":
			isNew = true
		case "0":
			return nil
		default:
			isNew = true
			// return errors.New("Only accept '0'|'1' to setup!")
		}
	}

	// Ask about whether sync by testnet
	if istestnet, err := console.InputText("Within testnet (1: Yes, 0: No) [default: 1]: ", false); err != nil {
		return err
	} else {
		switch istestnet {
		case "1":
			isTestNet = "true"
		case "0":
			isTestNet = "false"
		default:
			isTestNet = "true"
			// return errors.New("Only accept '0'|'1' to setup!")
		}
	}

	// Ask about Docker master Address and Port
	if addr, err := console.InputText("Docker master server address [default: localhost]: ", false); err != nil {
		return err
	} else {
		if addr != "" {
			serverAddr = addr
		}
	}
	if port, err := console.InputText("Docker master server port [default: 2735]: ", false); err != nil {
		return err
	} else {
		if port != "" {
			serverPort = port
		}
	}

	// Creat config
	if isNew == true {
		if err := initConfig(symbol); err != nil {
			return err
		}
	}

	return nil
}

func _CheckAdnCreateContainer(symbol string) error {
	// Instantize parameters
	cName, err := _GetCName(symbol) // container name
	if err != nil {
		return err
	}

	// Init docker client
	c, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	ctx := context.Background() // nil
	// Check if exist
	if res, err := c.ContainerInspect(ctx, cName); err == nil {
		// Exist
		status := res.State.Status
		fmt.Printf("%s walletnode exist: %s\n", symbol, status)
		return nil
	}
	// Action within client
	// dataDir := fmt.Sprintf("/storage/openwallet/%s", s.ToLower(symbol))
	containerConfig := container.Config{
		// string,
		Hostname: cName,
		// string,
		Domainname: fmt.Sprintf("%s.local", s.ToLower(cName)),
		// string, Command to run when starting the container
		Cmd: []string{"/bin/sh", "-c", "while true; do ping 8.8.8.8; done"},
		// string, Name of the image as it was passed by the operator(e.g. could be symbolic)
		Image: "ubuntu:latest",
		// map[string]struct{}, List of volumes (mounts) used for the container
		Volumes: map[string]struct{}{
			// Volumes: struct{}{
			// "type": "volume",
			// "src":           dataDir,
			// "dst":           dataDir,
			// "volume-driver": "local",
			// "vers":          "4,soft,timeo=180,bg,tcp,rw",
			// }
		},
		// string, Current directory (PWD) in the command will be launched
		WorkingDir: "",
		// strslice.StrSlice, Entrypoint to run when starting the container
		// Entrypoint,
		// map[string]string, List of labels set to this container
		Labels: map[string]string{},
	}
	hostConfig := container.HostConfig{
		// // NetworkMode   // Network mode to use for the container
		// NetworkMode,
		// // nat.PortMap   // Port mapping between the exposed port (container) and the host
		// PortBindings,
	}
	endpointSetting := network.EndpointSettings{
		// // Configurations
		// IPAMConfig, // *EndpointIPAMConfig
		// Links,      // []string
		// Aliases,    // []string
		// // Operational data
		// NetworkID,           // string
		// EndpointID,          // string
		// Gateway,             // string
		// IPAddress,           // string
		// IPPrefixLen,         // int
		// IPv6Gateway,         // string
		// GlobalIPv6Address,   // string
		// GlobalIPv6PrefixLen, // int
		// MacAddress,          // string
		// DriverOpts,          // map[string]string
	}
	networkingConfig := network.NetworkingConfig{
		// map[string]*EndpointSettings // Endpoint configs for each connecting network
		EndpointsConfig: map[string]*network.EndpointSettings{"endporint": &endpointSetting},
	}
	_, err = c.ContainerCreate(ctx, &containerConfig, &hostConfig, &networkingConfig, cName)
	if err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode created in success!\n", symbol)
	}
	return nil
}

func _InitConfigFile(symbol string) error {
	mainNetDataPath = filepath.Join(DATAPATH, s.ToLower(Symbol)+"/data")
	testNetDataPath = filepath.Join(DATAPATH, s.ToLower(symbol)+"/testdata")
	apiURL = fmt.Sprintf("http://%s:%s", FullnodeAddr, FullnodePort)
	rpcUser = "wallet"
	rpcPassword = "walletPassword2017"

	// Update config
	if err := updateConfig(symbol); err != nil {
		return err
	}
	return nil
}

// Create a new container for wallet fullnode
//
// Workflow:
//		// 步骤一
//		if 判定 ini 文件是否存在 {	// .ini 不存在
//			1> 询问用户配置参数
//			2> 创建初始 .ini 文件
//		} else {					// .ini 存在
//			1> 无操作，进入下一步
//		}
//
//		// 步骤二
//		if 判断 container 是否存在和正常 {	// 不存在 or 不正常
//			1> 删除，或直接创建一个新的
//			2> 导出 container 数据
//		} else {							// 正常
//			1> 导出 container 数据
//		}
//
//		// 步骤三
//		1> 根据导出的 container 数据，更新配置文件中关于 container 的项（重复更新，方便用户改错后自动恢复）
func (w *NodeManagerStruct) CreateNodeFlow(symbol string) error {

	// 一: Check <Symbol>.ini config, create new if not
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
