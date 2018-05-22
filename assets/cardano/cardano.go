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
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/timer"
	"log"
	"path/filepath"
	"strings"
)

type WalletManager struct{}

//初始化配置流程
func (w *WalletManager) InitConfigFlow() error {

	var (
		err        error
		apiURL     string
		walletPath string
		//汇总阀值
		threshold float64
		//最小转账额度
		minSendAmount float64
		//最小矿工费
		minFees float64
		//汇总地址
		sumAddress string
		filePath   string
	)

	for {

		fmt.Printf("[开始进行初始化配置流程]\n")

		apiURL, err = console.InputText("设置钱包API地址: ", true)
		if err != nil {
			return err
		}

		walletPath, err = console.InputText("设置钱包主链文件目录: ", false)
		if err != nil {
			return err
		}

		sumAddress, err = console.InputText("设置汇总地址: ", false)
		if err != nil {
			return err
		}

		fmt.Printf("[请输入以%s为单位的个数，必须正实数]\n", Symbol)

		//threshold, err = console.InputNumber("设置汇总阀值: ")
		//if err != nil {
		//	return err
		//}

		threshold, err = console.InputRealNumber("设置汇总阀值: ", true)
		if err != nil {
			return err
		}

		minSendAmount, err = console.InputRealNumber("设置账户最小转账额度: ", true)
		if err != nil {
			return err
		}

		fmt.Printf("[汇总手续费建议不少于%f]\n", 0.3)

		minFees, err = console.InputRealNumber("设置转账矿工费: ", true)
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
		fmt.Printf("汇总阀值: %f\n", threshold)
		fmt.Printf("账户最小转账额度: %f\n", minSendAmount)
		fmt.Printf("转账矿工费: %f\n", minFees)
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

//查看配置信息
func (w *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (w *WalletManager) CreateWalletFlow() error {

	var (
		password string
		confirm  string
		name     string
		err      error
	)

	//先加载是否有配置文件
	err = loadConfig()
	if err != nil {
		return err
	}

	for {

		// 等待用户输入钱包名字
		name, err = console.Stdin.PromptInput("1.输入钱包名字: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpect error: %v", err)
			return err
		}

		if len(name) == 0 {
			log.Printf("钱包名字不能为空, 请重新输入")
			continue
		}

		break
	}

	for {

		// 等待用户输入密码
		password, err = console.Stdin.PromptPassword("2.输入钱包密码: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpect error: %v", err)
			return err
		}

		if len(password) < 8 {
			log.Printf("不合法的密码长度, 建议设置不小于8位的密码, 请重新输入")
			continue
		}

		confirm, err = console.Stdin.PromptPassword("3.再次确认钱包密码: ")

		if password != confirm {
			log.Printf("两次输入密码不一致, 请重新输入")
			continue
		}

		break
	}
	// 随机生成密钥
	words := genMnemonic()
	return CreateNewWallet(name, words, password)

}

//创建地址流程
func (w *WalletManager) CreateAddressFlow() error {

	var (
		newAccountName string
		selectAccount  *Account
	)

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//输入钱包ID
	wid := inputWID()

	//输入钱包地址，查询账户信息
	accounts, err := GetAccountInfo(wid)
	if err != nil {
		return err
	}

	// 输入地址数量
	count := inputNumber()
	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("创建地址数量不能超过%d\n", maxAddresNum))
	}

	//没有账户创建新的
	if len(accounts) > 0 {
		//选择一个账户创建地址
		for _, a := range accounts {
			//限制每个账户的最大地址数不超过maxAddresNum
			if a.AddressNumber+count > maxAddresNum {
				continue
			} else {
				selectAccount = a
			}
		}
	}

	//输入密码
	password, err := console.InputPassword(false)
	h := common.NewString(password).SHA256()

	//没有可选账户，需要创建新的
	if selectAccount == nil {

		newAccountName = fmt.Sprintf("account %d", len(accounts))

		log.Printf("没有可用账户，创新账户[%s]\n", newAccountName)

		selectAccount, err = CreateNewAccount(newAccountName, wid, h)
		if err != nil {
			return err
		}
	}

	log.Printf("开始批量创建地址\n")
	log.Printf("================================================\n")

	_, filePath, err := CreateBatchAddress(selectAccount.AcountID, h, uint(count))

	log.Printf("================================================\n")
	log.Printf("地址批量创建成功，导出路径:%s\n", filePath)

	return err
}

//汇总钱包流程

/*

汇总执行流程：
1. 执行启动汇总某个币种命令。
2. 列出该币种的全部可用钱包信息。
3. 输入需要汇总的钱包序号数组（以,号分隔）。
4. 输入每个汇总钱包的密码，完成汇总登记。
5. 工具启动定时器监听钱包，并输出日志到log文件夹。
6. 待已登记的汇总钱包达到阀值，发起账户汇总到配置下的地址。

*/

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
		return errors.New("汇总地址还没设置，请在./conf/ADA.json中配置")
	}

	//查询所有钱包信息
	wallets, err := GetWalletInfo()
	if err != nil {
		fmt.Printf("客户端没有创建任何钱包\n")
		return err
	}

	fmt.Printf("序号\tWID\t\t\t\t\t\t\t\t名字\n")
	fmt.Printf("-----------------------------------------------------------------------------------------\n")

	for i, w := range wallets {
		fmt.Printf("%d\t%s\t%s\n", i, w.WalletID, w.Name)
	}
	fmt.Printf("-----------------------------------------------------------------------------------------\n")

	fmt.Printf("[请选择需要汇总的钱包，输入序号组，以,分隔。例如: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.Stdin.PromptInput("输入需要汇总的钱包序号: ")
	if err != nil {
		//openwLogger.Log.Errorf("unexpect error: %v", err)
		return err
	}

	if len(nums) == 0 {
		return errors.New("输入不能为空")
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]

				fmt.Printf("登记汇总钱包[%s]-[%s]\n", w.Name, w.WalletID)
				//输入钱包密码完成登记
				password, err := console.InputPassword(false)
				if err != nil {
					//openwLogger.Log.Errorf("unexpect error: %v", err)
					return err
				}

				//配置钱包密码
				h := common.NewString(password).SHA256()

				// 创建一个地址用于验证密码是否可以,默认账户ID = 2147483648 = 0x80000000
				testAccountid := fmt.Sprintf("%s@2147483648", w.WalletID)
				_, err = CreateAddress(testAccountid, h)
				if err != nil {
					openwLogger.Log.Errorf("输入的钱包密码错误")
					continue
				}

				w.Password = h

				AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("输入的序号越界")
			}
		} else {
			return errors.New("输入的序号不是数字")
		}
	}

	if len(walletsInSum) == 0 {
		return errors.New("没有汇总钱包完成登记")
	}

	fmt.Printf("钱包汇总定时器开启，间隔%f秒运行一次\n", cycleSeconds.Seconds())

	//启动钱包汇总程序
	sumTimer := timer.NewTask(cycleSeconds, SummaryWallets)
	sumTimer.Start()

	<-endRunning

	return nil
}

//备份钱包流程
func (w *WalletManager) BackupWalletFlow() error {

	var (
		err error
	)

	//先加载是否有配置文件
	err = loadConfig()
	if err != nil {
		return err
	}

	//建立备份路径
	backupPath := filepath.Join(keyDir, "backup")
	file.MkdirAll(backupPath)

	//linux
	//备份secret.key, secret.key.lock, wallet-db
	err = file.Copy(filepath.Join(walletPath, "secret.key"), backupPath)
	if err != nil {
		return err
	}
	err = file.Copy(filepath.Join(walletPath, "secret.key.lock"), backupPath)
	if err != nil {
		return err
	}
	err = file.Copy(filepath.Join(walletPath, "wallet-db"), backupPath)
	if err != nil {
		return err
	}

	//macOS
	//备份secret.key, secret.key.lock, wallet-db
	//err = file.Copy(filepath.Join(walletPath, "Secrets-1.0"), filepath.Join(backupPath))
	//if err != nil {
	//	return err
	//}
	//err = file.Copy(filepath.Join(walletPath, "Wallet-1.0"), filepath.Join(backupPath))
	//if err != nil {
	//	return err
	//}

	//输出备份导出目录
	fmt.Printf("钱包文件备份路径: %s", backupPath)

	return nil

}

//func ShowWallets() error {
//
//}
