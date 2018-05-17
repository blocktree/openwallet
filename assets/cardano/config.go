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
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"path/filepath"
	"fmt"
	"github.com/blocktree/OpenWallet/console"
	"errors"
	"github.com/blocktree/OpenWallet/common"
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
	exportBackupDir = "./data/ada/"
	//钥匙备份路径
	keyDir = "./data/ada/key/"
	//地址导出路径
	addressDir = "./data/ada/address/"
	//配置文件路径
	configFilePath = "./conf/"
	//配置文件名
	configFileName = "ADA.json"
	//币种
	coinSymbol = "ADA"
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
	threshold, minSendAmount, minFees uint64) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"apiURL":        apiURL,
		"walletPath":    walletPath,
		"sumAddress":    sumAddress,
		"threshold":     common.NewString(threshold).String(),
		"minSendAmount": common.NewString(minSendAmount).String(),
		"minFees":       common.NewString(minFees).String(),
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

//InitConfigFlow 初始化配置流程
func InitConfigFlow() error {

	var (
		err        error
		apiURL     string
		walletPath string
		//汇总阀值
		threshold uint64
		//最小转账额度
		minSendAmount uint64
		//最小矿工费
		minFees uint64
		//汇总地址
		sumAddress string
		filePath   string
	)

	for {

		fmt.Printf("[开始进行初始化配置流程]\n")

		apiURL, err = console.InputText("设置钱包API地址: ")
		if err != nil {
			return err
		}

		walletPath, err = console.InputText("设置钱包主链文件目录: ")
		if err != nil {
			return err
		}

		sumAddress, err = console.InputText("设置汇总地址: ")
		if err != nil {
			return err
		}

		fmt.Printf("[1个%s = %d，请输入整数*%d的数量]\n", coinSymbol, decimal, decimal)

		threshold, err = console.InputNumber("设置汇总阀值: ")
		if err != nil {
			return err
		}

		minSendAmount, err = console.InputNumber("设置账户最小转账额度: ")
		if err != nil {
			return err
		}

		fmt.Printf("[汇总手续费建议不少于%d]\n", uint64(0.3*float64(decimal)))

		minFees, err = console.InputNumber("设置转账矿工费: ")
		if err != nil {
			return err
		}

		//最小发送数量不能超过汇总阀值
		if minSendAmount > threshold {
			return errors.New("汇总阀值必须大于账户最小转账额度!")
		}

		if minFees > minSendAmount {
			return errors.New("账户最小转账额度必须大于手续费!")
		}

		//换两行
		fmt.Println()
		fmt.Println()

		//打印输入内容
		fmt.Printf("请检查以下内容是否正确?\n")
		fmt.Printf("-----------------------------------------------------------\n")
		fmt.Printf("钱包API地址: %s\n", apiURL)
		fmt.Printf("钱包主链文件目录: %s\n", walletPath)
		fmt.Printf("汇总地址: %s\n", sumAddress)
		fmt.Printf("汇总阀值: %d\n", threshold)
		fmt.Printf("账户最小转账额度: %d\n", minSendAmount)
		fmt.Printf("转账矿工费: %d\n", minFees)
		fmt.Printf("-----------------------------------------------------------\n")

		flag, err := console.Stdin.PromptConfirm("确认生成配置文件")
		if err != nil {
			return err
		}

		if !flag {
			continue
		} else {
			break
		}

	}

	//换两行
	fmt.Println()
	fmt.Println()

	_, filePath, err = newConfigFile(apiURL, walletPath, sumAddress, threshold, minSendAmount, minFees)

	fmt.Printf("配置已生成, 文件路径: %s\n", filePath)

	return nil
}

func printConfig() error {

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("配置文件未创建，请执行 wmd config -s <symbol> ")
	}

	apiURL := c.String("apiURL")
	walletPath := c.String("walletPath")
	threshold := c.String("threshold")
	minSendAmount := c.String("minSendAmount")
	minFees := c.String("minFees")
	sumAddress := c.String("sumAddress")

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("钱包API地址: %s\n", apiURL)
	fmt.Printf("钱包主链文件目录: %s\n", walletPath)
	fmt.Printf("汇总地址: %s\n", sumAddress)
	fmt.Printf("汇总阀值: %s\n", threshold)
	fmt.Printf("账户最小转账额度: %s\n", minSendAmount)
	fmt.Printf("转账矿工费: %s\n", minFees)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}

//ConfigSee 查看配置文件
func ConfigSee() error {
	return printConfig()
}