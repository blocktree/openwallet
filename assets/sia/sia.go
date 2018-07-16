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
	"log"
	"errors"
	"github.com/blocktree/OpenWallet/timer"
	"path/filepath"
	"github.com/shopspring/decimal"
)

const (
	//testAccount = "test-sign"
	maxAddressNum = 1000000
)

type WalletManager struct{}

//初始化配置流程
func (w *WalletManager) InitConfigFlow() error {

	file := filepath.Join(configFilePath, configFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's config.\n", file)

	return nil

}

//查看配置信息
func (w *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (w *WalletManager) CreateWalletFlow() error {

	var (
		password string
		//name      string
		err       error
		publicKey string
		filePath  string
		//testA    	*Account
	)

	flag, err := console.Stdin.PromptConfirm("Create a new wallet will cover the existing wallet data and reinitialize a new one, please backup the existing wallet first. Continue to create?")
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
		//name, err = console.InputText("Enter wallet's name: ", true)

		// 等待用户输入密码
		password, err = console.InputPassword(true, 8)

		publicKey, err = CreateNewWallet(password, true)
		if err != nil {
			return err
		}

		fmt.Printf("Please keep your primary seed in a safe place: %s\n", publicKey)

		filePath, err = BackupWallet()
		if err != nil {
			return err
		}

		//fmt.Printf("\n")
		//fmt.Printf("Wallet create successfully, first account: %s\n", name)
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

	wallets, err := GetWalletInfo()
	if err != nil {
		return err
	}
	if wallets[0].Rescanning{
		return errors.New(fmt.Sprint("Wallet is rescanning the block, please wait......"))
	}
	if !wallets[0].Unlocked {
		fmt.Println("The wallet is locked, please enter password to unlocked it.")
		password, err := console.InputPassword(false, 8)
		err = UnlockWallet(password)
		if err != nil {
			return errors.New(fmt.Sprintf("UnlockWallet failed unexpected error: %v\n", err))
		} else {
			log.Printf("UnlockWallet processing......\n")
		}
	}


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

	wallets, err := GetWalletInfo()
	if err != nil {
		return err
	}
	if wallets[0].Rescanning{
		return errors.New(fmt.Sprint("Wallet is rescanning the block, please wait......"))
	}
	if !wallets[0].Unlocked {
		fmt.Println("The wallet is locked, please enter password to unlocked it.")
		password, err := console.InputPassword(false, 8)
		err = UnlockWallet(password)
		if err != nil {
			return errors.New(fmt.Sprintf("UnlockWallet failed unexpected error: %v\n", err))
		} else {
			log.Printf("UnlockWallet processing......\n")
		}
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

	backupPath, err = BackupWallet()
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: %s", backupPath)

	return nil

}

//GetWalletList 获取钱包列表
func (w *WalletManager) GetWalletList() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	wallet, err := GetWalletInfo()
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(wallet)

	return nil
}

//SendTXFlow 发送交易
func (w *WalletManager) TransferFlow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	wallets, err := GetWalletInfo()
	if err != nil {
		return err
	}
	if wallets[0].Rescanning{
		return errors.New(fmt.Sprint("Wallet is rescanning the block, please wait......"))
	}
	if !wallets[0].Unlocked {
		fmt.Println("The wallet is locked, please enter password to unlocked it.")
		password, err := console.InputPassword(false, 8)
		err = UnlockWallet(password)
		if err != nil {
			return errors.New(fmt.Sprintf("UnlockWallet failed unexpected error: %v\n", err))
		} else {
			log.Printf("UnlockWallet processing......\n")
		}
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
	atculAmount, _ := decimal.NewFromString(amount)
	realAmount := atculAmount.Mul(coinDecimal)

	_, err = SendTransaction(realAmount.String(), receiver)
	if err != nil {
		return err
	}

	return nil
}

//RestoreWalletFlow 恢复钱包流程
func (w *WalletManager) RestoreWalletFlow() error {

	var (
		err      error
		filename string
	)

	//先加载是否有配置文件
	err = loadConfig()
	if err != nil {
		return err
	}

	//输入恢复文件路径
	filename, err = console.InputText("Enter backup file path: ", true)
	//keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	fmt.Printf("Wallet restoring, please wait a moment...\n")
	err = RestoreWallet(filename)
	if err != nil {
		return err
	}

	//输出备份导出目录

	return nil

}
