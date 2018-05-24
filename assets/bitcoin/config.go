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

package bitcoin

import (
	"path/filepath"
	"github.com/astaxie/beego/config"
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/shopspring/decimal"
	"encoding/json"
	"strings"
	"errors"
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
	//币种
	Symbol = "BTC"
)

var (
	//钥匙备份路径
	keyDir = filepath.Join("data", strings.ToLower(Symbol), "key")
	//地址导出路径
	addressDir = filepath.Join("data", strings.ToLower(Symbol), "address")
	//配置文件路径
	configFilePath = filepath.Join("conf")
	//配置文件名
	configFileName = Symbol + ".json"
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
	apiURL, walletPath, sumAddress string,
	threshold, minSendAmount, minFees float64) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":        apiURL,
		"walletPath":    walletPath,
		"sumAddress":    sumAddress,
		"threshold":     decimal.NewFromFloat(threshold).String(),
		"minSendAmount": decimal.NewFromFloat(minSendAmount).String(),
		"minFees":       decimal.NewFromFloat(minFees).StringFixed(6),
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

//printConfig Print config information
func printConfig() error {

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("config file not create，please run: wmd config -s <symbol> ")
	}

	apiURL := c.String("apiURL")
	walletPath := c.String("walletPath")
	threshold := c.String("threshold")
	minSendAmount := c.String("minSendAmount")
	minFees := c.String("minFees")
	sumAddress := c.String("sumAddress")

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("Wallet API URL: %s\n", apiURL)
	fmt.Printf("Wallet Data FilePath: %s\n", walletPath)
	fmt.Printf("Summary Address: %s\n", sumAddress)
	fmt.Printf("Summary Threshold: %s\n", threshold)
	fmt.Printf("Min Send Amount: %s\n", minSendAmount)
	fmt.Printf("Transfer Fees: %s\n", minFees)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}
