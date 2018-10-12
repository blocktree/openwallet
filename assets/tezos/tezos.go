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

package tezos

import (
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/timer"
	"log"
	"path/filepath"
	"strings"
	"errors"
	"github.com/shopspring/decimal"
	"strconv"
	"github.com/blocktree/OpenWallet/openwallet"
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
		keyFile  string
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	name, err = console.InputText("Enter wallet's name: ", true)

	// 等待用户输入密码
	password, err = console.InputPassword(true, 3)

	_, keyFile, err = wm.CreateNewWallet(name, password)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
}

//创建地址流程
func (wm *WalletManager) CreateAddressFlow() error {
	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	//查询所有钱包信息
	wallets, err := wm.GetWallets()
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	wm.printWalletList(wallets, false)

	fmt.Printf("[Please select a wallet No to create address] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet number: ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	wallet := wallets[num]

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddresNum))
	}

	//输入密码
	password, err := console.InputPassword(false, 6)

	log.Printf("Start batch creation\n")
	log.Printf("-------------------------------------------------\n")

	filePath, _, err := wm.CreateBatchAddress(wallet.WalletID, password, count)
	if err != nil {
		return err
	}

	log.Printf("-------------------------------------------------\n")
	log.Printf("All addresses have created, file path:%s\n", filePath)

	return nil
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
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	//查询所有钱包信息
	wallets, err := wm.GetWallets()
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	wm.printWalletList(wallets, true)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.InputText("Enter the No. group: ", true)
	if err != nil {
		return err
	}

	//分隔数组
	wallet_array := strings.Split(nums, ",")

	for _, numIput := range wallet_array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]

				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Alias, w.WalletID)
				//输入钱包密码完成登记
				password, err := console.InputPassword(false, 6)
				if err != nil {
					return err
				}

				w.Password = password

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
	var err error
	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	list, err := wm.GetWallets()
	if err != nil {
		return err
	}

	//打印钱包列表
	wm.printWalletList(list, false)

	fmt.Printf("[Please select a wallet to backup] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]

	//创建备份文件夹
	newBackupDir := filepath.Join(wm.Config.backupDir, wallet.FileName()+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	// 备份种子文件
	file.Copy(wallet.KeyFile, newBackupDir)

	//备份地址数据库
	file.Copy(wallet.DBFile, newBackupDir)

	//输出备份导出目录
	log.Printf("Wallet backup file path: %s", newBackupDir)
	return nil
}

//SendTXFlow 发送交易
func (wm *WalletManager) TransferFlow() error {
	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	wallets, err := wm.GetWallets()
	if err != nil {
		return err
	}

	//打印钱包列表
	addrs := wm.printWalletList(wallets, true)

	fmt.Printf("[Please select a wallet to send transaction] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	wallet := wallets[num]
	addr := addrs[num]

	// 等待用户输入发送数量
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}

	atculAmount, _ := decimal.NewFromString(amount)
	atculAmount = atculAmount.Mul(coinDecimal)
	log.Printf("amount:%d", atculAmount.IntPart())

	// 等待用户输入发送地址
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	//输入密码解锁钱包
	password, err := console.InputPassword(false, 6)
	if err != nil {
		return err
	}

	haveEnoughBalance := false

	//加载钱包
	key, err := wallet.HDKey(password)
	if err != nil {
		return err
	}

	type sendStruct struct {
		sednKeys *Key
		fee      decimal.Decimal
		amount   decimal.Decimal
	}

	var sends []sendStruct
	var resultSub decimal.Decimal = atculAmount
	//新建发送地址列表以及验证余额是否足够
	for _, a := range addr {
		k, _ := wm.getKeys(key, a)

		//get balance
		decimal_balance := decimal.RequireFromString(a.Balance)

		//判断是否是reveal交易
		fee := wm.Config.MinFee
		isReverl := wm.isReverlKey(a.Address)
		if isReverl {
			//多了reveal操作后，fee * 2
			fee = wm.Config.MinFee.Mul(decimal.RequireFromString("2"))
		}
		// 将该地址多余额减去矿工费
		amount := decimal_balance.Sub(fee)
		//该地址预留一点币，否则交易会失败，暂定0.00002 tez
		amount = amount.Sub(decimal.RequireFromString("20"))
		if amount.IntPart() < 0 {
			continue
		}

		if resultSub.LessThanOrEqual(amount) {
			send := sendStruct{k, fee, resultSub}
			sends = append(sends, send)
			haveEnoughBalance = true
			break
		} else {
			send := sendStruct{k, fee, amount}
			sends = append(sends, send)
			//log.Printf("address:%s, amount:%d, resultSub:%d\n", k.Address, amount.IntPart(), resultSub.IntPart())
		}
		resultSub = resultSub.Sub(amount)
		//log.Printf("resultSub:%d\n", resultSub.IntPart())
	}

	if haveEnoughBalance {
		for _, send := range sends {
			txid, _ := wm.Transfer(*send.sednKeys, receiver, strconv.FormatInt(wm.Config.MinFee.IntPart(), 10), strconv.FormatInt(wm.Config.GasLimit.IntPart(), 10),
				strconv.FormatInt(wm.Config.StorageLimit.IntPart(), 10),strconv.FormatInt(send.amount.IntPart(), 10))
			log.Printf("transfer address:%s, to address:%s, amount:%d, txid:%s\n", send.sednKeys.Address, wm.Config.SumAddress, send.amount.IntPart(), txid)
		}
	} else {
		log.Printf("not enough balance\n")
		return errors.New("Wallet have not enough balance to transfer\n")
	}

	return nil
}

//GetWalletList 获取钱包列表
func (wm *WalletManager) GetWalletList() error {
	var err error

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	list, err := wm.GetWallets()
	if err != nil {
		return err
	}

	//打印钱包列表
	wm.printWalletList(list, false)

	fmt.Printf("Do you want to query address's public key ?\n0)No\n1)Yes\n")

	//选择0或则1
	num, err := console.InputNumber(":", true)
	if err != nil {
		return err
	}

	if num != 1 {
		return nil
	}

	fmt.Printf("[Please select the wallet ] \n")

	//选择钱包
	num, err = console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	// 等待用户输入查询地址
	address, err := console.InputText("Enter query address: ", true)
	if err != nil {
		return err
	}

	//输入密码解锁钱包
	password, err := console.InputPassword(false, 6)
	if err != nil {
		return err
	}

	db, err := list[num].OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	//加载钱包
	key, err := list[num].HDKey(password)
	if err != nil {
		return err
	}

	var addr *openwallet.Address
	got := false
	for _, a := range addrs {
		if a.Address == address {
			addr = a
			got = true
			break
		}
	}

	if got == false {
		fmt.Printf("There is no address：%s in wallet %s\n", address, list[num].Alias)
		return nil
	}

	k, _ := wm.getKeys(key, addr)
	fmt.Printf("Public key:%s\nPrivate key: %s\n", k.PublicKey, k.PrivateKey)

	return nil
}


//RestoreWalletFlow 恢复钱包
func (w *WalletManager) RestoreWalletFlow() error {

	var (
		err      error
		keyFile  string
		dbFile   string
		password string
	)

	//先加载是否有配置文件
	err = w.LoadConfig()
	if err != nil {
		return err
	}

	//输入恢复文件路径
	keyFile, err = console.InputText("Enter backup key file path: ", true)
	if err != nil {
		return err
	}

	dbFile, err = console.InputText("Enter backup db file path: ", true)
	if err != nil {
		return err
	}

	password, err = console.InputPassword(false, 3)
	if err != nil {
		return err
	}

	fmt.Printf("Wallet restoring, please wait a moment...\n")
	err = w.RestoreWallet(keyFile, dbFile, password)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Restore wallet successfully.\n")

	return nil
}

