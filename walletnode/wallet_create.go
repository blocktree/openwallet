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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	s "strings"

	"github.com/blocktree/OpenWallet/console"
)

// Check <Symbol>.ini file, create new if not
//
// Workflow:
//		1> 当前目录没有 ini，是否创建？
//			1.1 存在，return nil
//		2> 询问是否设置为测试链？
//		3> 获取Master服务器IP地址和端口
func CheckAndCreateConfig(symbol string) error {

	// Check <Symbol>.ini
	if err := loadConfig(symbol); err != nil {
		// <Symbol>.ini exist, return and go next
		// return nil
	} else {
		fmt.Printf("Config file <%s.ini> existed!\n", s.ToUpper(symbol))
	}

	// Ask about whether create new
	dirname, _ := filepath.Abs("./")
	fmt.Printf("\nInit new %s wallet fullnode in '%s/'( \n  yes:\t to create config file and docker, \n  no:\t just to create docker, \n  quit:\t exit now!)", s.ToUpper(symbol), dirname)
	if isnew, err := console.InputText("[yes]: ", false); err != nil {
		log.Println(err)
		return err
	} else {
		switch isnew {
		case "", "yes":
		case "no":
			return nil
		case "quit":
			os.Exit(0)
		default:
			return errors.New("Invalid!")
		}
	}

	// Ask about whether sync by testnet
	if istestnet, err := console.InputText("Within testnet('testnet','main')[main]: ", false); err != nil {
		return err
	} else {
		switch istestnet {
		case "main", "":
			WNConfig.isTestNet = "false"
		case "testnet":
			WNConfig.isTestNet = "true"
		default:
			return errors.New("Invalid!")
		}
	}

	// Ask about Docker master
	if x, err := console.InputText("Where to run Walletnode: local/docker [docker]: ", false); err != nil {
		return err
	} else {
		if x == "" {
			WNConfig.walletnodeServerType = "docker"
		} else {
			if _, ok := map[string]string{"local": "", "docker": ""}[x]; !ok {
				return errors.New("Invalid!")
			}
			WNConfig.walletnodeServerType = x
		}
	}
	//if WNConfig.walletnodeServerType == "localdocker" {

	//	if x, err := console.InputText("Docker master server socket [/var/run/docker.socket]: ", false); err != nil {
	//		return err
	//	} else {
	//		if x != "" {
	//			WNConfig.walletnodeServerSocket = x
	//		} else {
	//			WNConfig.walletnodeServerSocket = "/var/run/docker.socket"
	//		}
	//	}

	//} else if WNConfig.walletnodeServerType == "docker" {

	if WNConfig.walletnodeServerType == "docker" {

		if x, err := console.InputText("Docker master server addr [127.0.0.1]: ", false); err != nil {
			return err
		} else {
			if x != "" {
				WNConfig.walletnodeServerAddr = x
			} else {
				WNConfig.walletnodeServerAddr = "127.0.0.1"
			}
		}

		if x, err := console.InputText("Docker master server port [2375]: ", false); err != nil {
			return err
		} else {
			if x != "" {
				WNConfig.walletnodeServerPort = x
			} else {
				WNConfig.walletnodeServerPort = "2375"
			}
		}

	} else if WNConfig.walletnodeServerType == "local" {
		if x, err := console.InputText("Start walletnode command: ", false); err != nil {
			return err
		} else {
			WNConfig.walletnodeStartNodeCMD = x
		}

		if x, err := console.InputText("Stop walletnode command: ", false); err != nil {
			return err
		} else {
			WNConfig.walletnodeStopNodeCMD = x
		}
		// console.InputText("Please edit <stopnodecmd/startnodecmd> in Symbol.ini before use wallet [yes]: ", false)
	}

	if cnf := getFullnodeConfig(symbol); cnf != nil {
		if cnf.isEncrypted() {
			fmt.Println("** Wallet fullnode need to be encrypted, and will encrypt within starting! **")
		}
	}

	// ---------------------- Create container success and update SYMBOL.ini -----------------
	fmt.Println("Start to create/update config file...")
	// Create new INI file, and update
	if err := initConfig(symbol); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("\t create success!")

	if err := updateConfig(symbol); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("\t update success!")

	return nil
}
