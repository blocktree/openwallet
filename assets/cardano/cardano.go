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
	"github.com/blocktree/OpenWallet/timer"
	"github.com/shopspring/decimal"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

//初始化配置流程
func (wm *WalletManager) InitConfigFlow() error {
	wm.Config.InitConfig()
	file := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's Config.\n", file)

	return nil
}

//查看配置信息
func (wm *WalletManager) ShowConfig() error {
	return wm.Config.PrintConfig()
}

//创建钱包流程
func (wm *WalletManager) CreateWalletFlow() error {

	var (
		password string
		name     string
		err      error
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	name, err = console.InputText("Enter wallet's name: ", true)

	// 等待用户输入密码
	password, err = console.InputPassword(true, 8)

	// 随机生成密钥
	words := genMnemonic()
	return wm.CreateNewWallet(name, words, password)

}

//创建地址流程
func (wm *WalletManager) CreateAddressFlow() error {

	var (
		newAccountName string
		selectAccount  *Account
	)

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	//输入钱包ID
	wid := inputWID()

	//输入钱包地址，查询账户信息
	accounts, err := wm.GetAccountInfo(wid)
	if err != nil {
		return err
	}

	// 输入地址数量
	count := inputNumber()
	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddresNum))
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
	password, err := console.InputPassword(false, 8)
	h := common.NewString(password).SHA256()

	//没有可选账户，需要创建新的
	if selectAccount == nil {

		newAccountName = fmt.Sprintf("account %d", len(accounts))

		log.Printf("No available accounts, create account[%s]\n", newAccountName)

		selectAccount, err = wm.CreateNewAccount(newAccountName, wid, h)
		if err != nil {
			return err
		}
	}

	log.Printf("Start batch creation\n")
	log.Printf("================================================\n")

	_, filePath, err := wm.CreateBatchAddress(wid, selectAccount.Index, h, uint(count))

	log.Printf("================================================\n")
	log.Printf("All addresses have created, file path:%s\n", filePath)

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
func (wm *WalletManager) SummaryFollow() error {

	var (
		endRunning = make(chan bool, 1)
	)

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(wm.Config.SumAddress) == 0 {
		return errors.New("Summary address is not set. Please set it in './conf/ADA.json' ")
	}

	//查询所有钱包信息
	wallets, err := wm.GetWalletInfo()
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	wm.printWalletList(wallets)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.Stdin.PromptInput("Enter the No. group: ")
	if err != nil {
		//openwLogger.Log.Errorf("unexpect error: %v", err)
		return err
	}

	if len(nums) == 0 {
		return errors.New("Input can not be empty! ")
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]

				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Name, w.WalletID)
				//输入钱包密码完成登记
				password, err := console.InputPassword(false, 8)
				if err != nil {
					//openwLogger.Log.Errorf("unexpect error: %v", err)
					return err
				}

				//配置钱包密码
				h := common.NewString(password).SHA256()

				// 创建一个地址用于验证密码是否可以,默认账户ID = 2147483648 = 0x80000000
				//testAccountid := fmt.Sprintf("%s@2147483648", w.WalletID)
				//_, err = wm.CreateAddress(w.WalletID, 1, h)
				//if err != nil {
				//	openwLogger.Log.Errorf("The password to unlock wallet is incorrect! ")
				//	continue
				//}

				w.Password = h

				wm.AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("The input No. out of index! ")
			}
		} else {
			return errors.New("The input No. is not numeric! ")
		}
	}

	if len(wm.WalletsInSum) == 0 {
		return errors.New("Not summary wallets to register! ")
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", wm.Config.CycleSeconds.Seconds())

	//启动钱包汇总程序
	sumTimer := timer.NewTask(wm.Config.CycleSeconds, wm.SummaryWallets)
	sumTimer.Start()

	<-endRunning

	return nil
}

//备份钱包流程
func (wm *WalletManager) BackupWalletFlow() error {

	var (
		err error
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	//建立备份路径
	backupPath := filepath.Join(wm.Config.keyDir, "backup")
	file.MkdirAll(backupPath)

	switch runtime.GOOS {
	case "darwin":
		//macOS
		//备份state-wallet-mainnet
		err = file.Copy(wm.Config.WalletDataPath, filepath.Join(backupPath))
		if err != nil {
			return err
		}

	case "linux":
		//linux
		//备份state-wallet-mainnet
		err = file.Copy(wm.Config.WalletDataPath, backupPath)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupport operation system")
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: [%s]\n", backupPath)

	return nil

}

//SendTXFlow 发送交易
func (wm *WalletManager) TransferFlow() error {

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	list, err := wm.GetWalletInfo()
	if err != nil {
		return err
	}

	//打印钱包列表
	wm.printWalletList(list)

	fmt.Printf("[Please select a wallet to send transaction] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]

	// 等待用户输入发送数量
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}

	atculAmount, _ := decimal.NewFromString(amount)
	atculAmount = atculAmount.Mul(coinDecimal)
	balance, err := decimal.NewFromString(wallet.Balance)

	if atculAmount.GreaterThan(balance) {
		return errors.New("Input amount is greater than balance! ")
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	//汇总所有有钱的账户
	accounts, err := wm.GetAccountInfo(wallet.WalletID)
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return errors.New("Wallet is empty account! ")
	}

	//计算预估手续费

	for _, a := range accounts {
		//大于最小额度才转账
		accountAmount, _ := decimal.NewFromString(a.Amount)
		if accountAmount.GreaterThan(atculAmount) && atculAmount.GreaterThan(wm.Config.MinFees) {

			fmt.Printf("-----------------------------------------------\n")
			fmt.Printf("From Wallet [%s] - Account: [%d]\n", wallet.WalletID, a.Index)
			fmt.Printf("To Address: %s\n", receiver)
			fmt.Printf("Send: %v\n", atculAmount.Div(coinDecimal))
			fmt.Printf("Fees: %v\n", wm.Config.MinFees.Div(coinDecimal))
			fmt.Printf("Receive: %v\n", atculAmount.Sub(wm.Config.MinFees).Div(coinDecimal))
			fmt.Printf("-----------------------------------------------\n")

			fmt.Printf("[Please unlock wallet to send transaction]\n")

			//输入密码解锁钱包
			password, err := console.InputPassword(false, 8)
			if err != nil {
				return err
			}

			//配置钱包密码
			h := common.NewString(password).SHA256()

			tx, err := wm.SendTx(wallet.WalletID, a.Index, receiver, uint64(atculAmount.Sub(wm.Config.MinFees).IntPart()), h)
			if err != nil {
				log.Printf("Send transaction failed, unexpected error：%v\n", err)
				continue
			} else {
				log.Printf("Send transaction successfully, TXID：%s\n", tx.TxID)
				return nil
			}
		} else {
			fmt.Printf("account [%d] Insufficient balance\n", a.Index)
		}
	}

	return nil
}

//GetWalletList 获取钱包列表
func (wm *WalletManager) GetWalletList() error {

	var (
		err error
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	list, err := wm.GetWalletInfo()
	if err != nil {
		return err
	}

	//打印钱包列表
	wm.printWalletList(list)

	return nil
}

//RestoreWalletFlow 恢复钱包
func (w *WalletManager) RestoreWalletFlow() error {

	fmt.Printf("Restore wallet is unavailable now.\n")

	return nil
}
