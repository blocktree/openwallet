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

package cardano

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"path/filepath"
)

/*
	工具可以读取各个币种钱包的默认的配置资料，
	币种钱包的配置资料放在conf/{symbol}.conf，例如：ADA.conf, BTC.conf，ETH.conf。
	执行wmd wallet -s <symbol> 命令会先检查是否存在该币种钱包的配置文件。
	没有：执行ConfigFlow，配置文件初始化。
	有：执行常规命令。
	使用者还可以通过wmd config -s 进行修改配置文件。
	或执行wmd config flow 重新进行一次配置初始化流程。

*/

const (
	//备份路径
	//exportBackupDir = "./data/ada/"
	//钥匙备份路径
	keyDir = "./data/ada/key/"
	//地址导出路径
	addressDir = "./data/ada/address/"
	//配置文件路径
	configFilePath = "./conf/"
	//配置文件名
	configFileName = "ADA.json"
	//币种
	Symbol = "ADA"
)

//isExistConfigFile 检查配置文件是否存在
func isExistConfigFile() bool {
	_, err := config.NewConfig("json",
		filepath.Join(configFilePath, configFileName))
	if err != nil {
		return false
	}
	return true
}

//newConfigFile 创建配置文件
func newConfigFile(
	apiURL, walletPath, sumAddress,
	threshold, minSendAmount, minFees string) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":        apiURL,
		"walletPath":    walletPath,
		"sumAddress":    sumAddress,
		"threshold":     threshold,
		"minSendAmount": minSendAmount,
		"minFees":       minFees,
	}

	filepath.Join()

	bytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, "", err
	}

	//实例化配置
	c, err := config.NewConfigData("json", bytes)
	if err != nil {
		return nil, "", err
	}

	//写入配置到文件
	file.MkdirAll(configFilePath)
	absFile := filepath.Join(configFilePath, configFileName)
	err = c.SaveConfigFile(absFile)
	if err != nil {
		return nil, "", err
	}

	return c, absFile, nil
}

func printConfig() error {

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	apiURL := c.String("apiURL")
	walletPath := c.String("walletPath")
	threshold := c.String("threshold")
	minSendAmount := c.String("minSendAmount")
	minFees := c.String("minFees")
	sumAddress := c.String("sumAddress")

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("Node API url: %s\n", apiURL)
	fmt.Printf("Wallet main net filePath: %s\n", walletPath)
	fmt.Printf("Summary address: %s\n", sumAddress)
	fmt.Printf("Summary threshold: %s\n", threshold)
	fmt.Printf("Minimum transfer amount: %s\n", minSendAmount)
	fmt.Printf("Transfer fees: %s\n", minFees)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}
