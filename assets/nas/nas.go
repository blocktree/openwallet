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

package nas

import (
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/shopspring/decimal"
	"github.com/blocktree/OpenWallet/log"
	"path/filepath"
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

	log.Std.Info("Start batch creation")
	log.Std.Info("-------------------------------------------------")

	filePath, _, err := wm.CreateBatchAddress(wallet.WalletID, password, count)
	if err != nil {
		return err
	}

	log.Std.Info("-------------------------------------------------")
	log.Std.Info("All addresses have created, file path:%s", filePath)

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
	log.Std.Info("Wallet backup file path: %s", newBackupDir)
	return nil
}

//SendTXFlow 发送交易 1NAS = 10^18 Wei
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

	// 等待用户输入发送数量,单位为NAS
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}
	atculAmount, _ := decimal.NewFromString(amount)

	//将输入单位为NAS转为Wei
	amount_wei := atculAmount.Mul(coinDecimal)

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
	db, err := wallet.OpenDB()
	if err != nil {
		log.Std.Info("open db failed, err=", err)
		return err

	}
	defer db.Close()

	key, err := wallet.HDKey(password)
	if err != nil {
		return err
	}

	type sendStruct struct {
		sednKeys *Key
		amount   decimal.Decimal
	}

	var sends []sendStruct
	var resultSub decimal.Decimal = amount_wei
	//新建发送地址列表以及验证余额是否足够
	for _, a := range addr {
		k, _ := wm.getKeys(key, a)

		//get balance,单位为Wei
		balance_decimal := decimal.RequireFromString(a.Balance)

		//该地址预留一点币，否则交易会失败，暂定0.00001 NAS
		//fmt.Printf("a.Address=%s,balance_decimal=%v\n",a.Address,balance_decimal)
		balance_leave := decimal.RequireFromString("10000000000000")

		cmp_result := balance_decimal.Cmp(balance_leave)
		if  cmp_result==0 ||cmp_result==-1{
			continue
		}
		balance_safe := balance_decimal.Sub(balance_leave)

		//此处作用为当钱包中某个账户balance_safe小于发送值时，发送此钱包的balance_safe
/*		example :
					wallet_ID:	W7GdMT1hte6ckTLYMf5iJHyr4hGRaYSgVg
					addr1:		n1S8ojaa9Pz8TduXEm8vXrxBs6Kz5dyp7km   2NAS
					addr2:		n1H91CSmYdRRYCa9zxcApW9QmCTix1MBrs4   2NAS
					addr3:		n1Pf8mSMwne8hWXDPF3iEqvfTMyeXwdb9Uk	  2NAS
					发送:5NAS = addr1(2NAS)+addr2(2NAS)+addr3(1NAS)*/

		if resultSub.LessThanOrEqual(balance_safe) {
		//	send := sendStruct{k, fee, resultSub}
			send := sendStruct{k,  resultSub}
			sends = append(sends, send)
			haveEnoughBalance = true
			break
		} else {
			send := sendStruct{k,  balance_safe}
			sends = append(sends, send)
			//log.Printf("address:%s, amount:%d, resultSub:%d\n", k.Address, amount.IntPart(), resultSub.IntPart())
		}
		resultSub = resultSub.Sub(balance_safe)
		//log.Printf("resultSub:%d\n", resultSub.IntPart())
	}

	if haveEnoughBalance {
		for _, send := range sends {

			txid, err := wm.Transfer(send.sednKeys, receiver, "2000000",send.amount.String())
		//	log.Std.Info("transfer address:%s, to address:%s, amount:%s, txid:%s", send.sednKeys.Address, receiver, send.amount.String(), txid)
			if err != nil{
				log.Std.Info("Transfer Fail!",)
			}else{
					log.Std.Info("Transfer Success! txid=%s",txid)
					err := NotenonceInDB(send.sednKeys,db)
					if err != nil {
						log.Std.Info("NotenonceInDB error")
					}
			}
		}
	} else {
		log.Std.Info("not enough balance")
		return errors.New("Wallet have not enough balance to transfer")
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
	wm.printWalletList(list, true)

	fmt.Printf("Do you want to query address's public key ?\n0)No\n1)Yes\n")

	//选择钱包
	num, err := console.InputNumber(":", true)
	if err != nil {
		return err
	}

	if num == 0 {
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
	fmt.Printf("Public key:%v\n", k.PublicKey)
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

