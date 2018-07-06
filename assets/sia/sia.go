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

package sia

import (
	"fmt"
	"github.com/blocktree/OpenWallet/console"
	"strings"
	"log"
	"errors"
	"github.com/blocktree/OpenWallet/timer"
)

const (
	//testAccount = "test-sign"
	maxAddressNum = 1000000
)

type WalletManager struct{}

//初始化配置流程
func (w *WalletManager) InitConfigFlow() error {

	var (
		err        error
		apiURL     string
		walletPath string
		//汇总阀值
		threshold string
		//最小转账额度
		//minSendAmount float64
		//最小矿工费
		//minFees float64
		//汇总地址
		sumAddress string
		filePath   string
	)

	for {

		fmt.Printf("[Start setup wallet config]\n")

		apiURL, err = console.InputText("Set node API url: ", true)
		if err != nil {
			return err
		}

		//删除末尾的/
		apiURL = strings.TrimSuffix(apiURL, "/")

		//walletPath, err = console.InputText("Set wallet main net filePath: ", false)
		//if err != nil {
		//	return err
		//}

		sumAddress, err = console.InputText("Set summary address: ", false)
		if err != nil {
			return err
		}

		fmt.Printf("[Please enter the amount of %s and must be numbers]\n", Symbol)

		//threshold, err = console.InputNumber("设置汇总阀值: ")
		//if err != nil {
		//	return err
		//}

		threshold, err = console.InputRealNumber("Set summary threshold: ", true)
		if err != nil {
			return err
		}

		//换两行
		fmt.Println()
		fmt.Println()

		//打印输入内容
		fmt.Printf("Please check the following setups is correct?\n")
		fmt.Printf("-----------------------------------------------------------\n")
		fmt.Printf("Node API url: %s\n", apiURL)
		//fmt.Printf("Wallet main net filePath: %s\n", walletPath)
		fmt.Printf("Summary address: %s\n", sumAddress)
		fmt.Printf("Summary threshold: %s\n", threshold)
		fmt.Printf("-----------------------------------------------------------\n")

		flag, err := console.Stdin.PromptConfirm("Confirm to save the setups?")
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

	_, filePath, err = newConfigFile(apiURL, walletPath, sumAddress, threshold)

	fmt.Printf("Config file create, file path: %s\n", filePath)

	return nil

}

//查看配置信息
func (w *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (w *WalletManager) CreateWalletFlow() error {

	var (
		password  string
		name      string
		err       error
		publicKey string
		filePath  string
		//testA    	*Account
		destination string
	)

	flag, err := console.Stdin.PromptConfirm("Create a new wallet will destroy the existing wallet and reinitialize a new one, continue?")
	if err != nil {
		return err
	}

	if flag {

		//先加载是否有配置文件
		err = loadConfig()
		if err != nil {
			return err
		}

		// 等待用户输入钱包名字
		name, err = console.InputText("Enter wallet's name: ", true)

		// 等待用户输入密码
		password, err = console.InputPassword(true)

		publicKey, err = CreateNewWallet(password, true)
		if err != nil {
			return err
		}

		fmt.Printf("Please keep your public key in a safe place: %s\n", publicKey)

		////创建钱包第一个账户
		//_, err = CreateNormalAccount(publicKey, name)
		//if err != nil {
		//	return err
		//}
		//
		////同事创建一个测试账户和测试地址，用于验证消息签名
		//testA, err = CreateNormalAccount(publicKey, name+"_"+testAccount)
		//if err != nil {
		//	return err
		//}
		////生成一个测试地址，用于通过消息签名验证密码是否正确
		//_, err = CreateReceiverAddress(testA.Alias, testA.ID)
		//if err != nil {
		//	return err
		//}

		//每创建一次钱包，备份一次

		// 等待用户输入钱包备份路径
		destination, err = console.InputText("Enter wallet's backup path: ", true)

		err = BackupWallet(destination)
		if err != nil {
			return err
		}

		fmt.Printf("\n")
		fmt.Printf("Wallet create successfully, first account: %s\n", name)
		fmt.Printf("Keystore backup successfully, file path: %s\n", filePath)
	}
	return nil
}

//创建地址流程
func (w *WalletManager) CreateAddressFlow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//查询所有钱包信息
	//wallets, err := GetWalletList(assetsID_btm)
	//if err != nil {
	//	fmt.Printf("The node did not create any wallet!\n")
	//	return err
	//}

	//打印钱包
	//printWalletList(wallets)

	//fmt.Printf("[Please select a wallet account to create address] \n")

	//选择钱包
	//num, err := console.InputNumber("Enter wallet number: ", true)
	//if err != nil {
	//	return err
	//}

	//if int(num) >= len(wallets) {
	//	return errors.New("Input number is out of index! ")
	//}

	//account := wallets[num]

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddressNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddressNum))
	}

	log.Printf("Start batch creation\n")
	log.Printf("-------------------------------------------------\n")

	filePath, err := CreateBatchAddress(count)

	log.Printf("-------------------------------------------------\n")
	log.Printf("All addresses have created, file path:%s\n", filePath)

	return err
}

// SummaryFollow 汇总流程
func (w *WalletManager) SummaryFollow() error {

	var (
		endRunning = make(chan bool, 1)
	)

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(sumAddress) == 0 {

		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.json' \n", Symbol))
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", cycleSeconds.Seconds())

	//启动钱包汇总程序
	sumTimer := timer.NewTask(cycleSeconds, SummaryWallets)
	sumTimer.Start()

	<-endRunning

	return nil
}

//备份钱包流程
func (w *WalletManager) BackupWalletFlow() error {

	var (
		err        error
		backupPath string
	)

	//先加载是否有配置文件
	err = loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入备份地址
	backupPath, err = console.InputText("Enter backup path: ", true)
	if err != nil {
		return err
	}

	err = BackupWallet(backupPath)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: %s", backupPath)

	return nil

}

//GetWalletList 获取钱包列表
func (w *WalletManager) GetWalletList() error {
	return nil
}

//SendTXFlow 发送交易
func (w *WalletManager) TransferFlow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入发送数量
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}

	// 等待用户输入发送地址
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	//建立交易单
	_, err = SendTransaction(amount, receiver)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully.\n")

	return nil
}

//RestoreWalletFlow 恢复钱包流程
func (w *WalletManager) RestoreWalletFlow() error {

	//var (
	//	err        error
	//	filename string
	//)
	//
	////先加载是否有配置文件
	//err = loadConfig()
	//if err != nil {
	//	return err
	//}
	//
	////输入恢复文件路径
	//filename, err = console.InputText("Enter backup file path: ", true)
	//keyjson, err := ioutil.ReadFile(filename)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Printf("Wallet restoring, please wait a moment...\n")
	//err = RestoreWallet(keyjson)
	//if err != nil {
	//	return err
	//}
	//
	////输出备份导出目录
	//fmt.Printf("Restore wallet successfully.\n")

	return nil

}

//SetConfigFlow 初始化配置流程
func (w *WalletManager) SetConfigFlow(subCmd string) error {
	file := configFilePath + configFileName
	fmt.Printf("You can run 'vim %s' to edit %s config.\n", file, subCmd)
	return nil
}

//ShowConfigInfo 查看配置信息
func (w *WalletManager) ShowConfigInfo(subCmd string) error {
	printConfig()
	return nil
}
