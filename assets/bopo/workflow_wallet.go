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

package bopo

import (
	"github.com/imroc/req"
	// "bufio"
	// "encoding/json"
	"fmt"
	// "github.com/blocktree/OpenWallet/assets/bopo"
	// "github.com/astaxie/beego/config"
	"github.com/tidwall/gjson"
	// "github.com/blocktree/OpenWallet/common"
	// "github.com/blocktree/OpenWallet/common/file"
	// "github.com/blocktree/OpenWallet/keystore"
	// "github.com/btcsuite/btcd/chaincfg"
	// "github.com/btcsuite/btcutil"
	// "github.com/btcsuite/btcutil/hdkeychain"
	// "github.com/codeskyblue/go-sh"
	// "github.com/pkg/errors"
	// "github.com/shopspring/decimal"
	// "io/ioutil"
	// "log"
	// "math"
	// "os"
	// "path/filepath"
	// "sort"
	// "strings"
	//"time"
)

// -----------------------------------------------------------------------------------
//getWalletList 获取钱包列表
func getWalletList() ([]*Wallet, error) {
	var wallets = make([]*Wallet, 0)

	r, err := client.Call("account", "GET", nil)
	if err != nil {
		return nil, err
	}
	data := gjson.ParseBytes(r).Map()

	for walletid, a := range data {
		addr := a.String()
		w := &Wallet{Alias: walletid, Addr: addr}

		// Get balance
		if d, err := client.Call(fmt.Sprintf("chain/%s", addr), "GET", nil); err == nil {
			w.Balance = gjson.ParseBytes(d).Map()["pais"].String()
		}

		wallets = append(wallets, w)
	}

	return wallets, nil
}

// Get one wallet info
func getWalletInfo(name string) (*Wallet, error) {

	if r, err := client.Call(fmt.Sprintf("account/%s", name), "GET", nil); err != nil {
		return nil, err
	} else {
		data := gjson.ParseBytes(r).Map()
		return &Wallet{Alias: name, Addr: data["address"].String()}, nil
	}
}

//CreateNewWallet 创建钱包
func createWallet(name string) (*Wallet, error) {
	var wallet *Wallet

	if _, err := client.Call("account", "POST", req.Param{"id": name}); err != nil {
		return nil, err
	} else {
		if w, err := getWalletInfo(name); err != nil {
			wallet = &Wallet{}
		} else {
			wallet = w
		}
	}

	return wallet, nil
}

// //BackupWalletData 备份钱包
// func BackupWalletData(dest string) error {
//
// 	request := []interface{}{
// 		dest,
// 	}
//
// 	_, err := client.Call("backupwallet", request)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
//
// }
//
// //BackupWallet 备份数据
// func BackupWallet(walletID string) (string, error) {
// 	w, err := GetWalletInfo(walletID)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	//创建备份文件夹
// 	newBackupDir := filepath.Join(backupDir, w.FileName()+"-"+common.TimeFormat("20060102150405"))
// 	file.MkdirAll(newBackupDir)
//
// 	//创建临时备份文件wallet.dat
// 	tmpWalletDat := fmt.Sprintf("tmp-walllet-%d.dat", time.Now().Unix())
// 	tmpWalletDat = filepath.Join(walletDataPath, tmpWalletDat)
//
// 	//1. 备份核心钱包的wallet.dat
// 	err = BackupWalletData(tmpWalletDat)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	//复制临时文件到备份文件夹
// 	file.Copy(tmpWalletDat, filepath.Join(newBackupDir, "wallet.dat"))
//
// 	//删除临时文件
// 	file.Delete(tmpWalletDat)
//
// 	//2. 备份种子文件
// 	file.Copy(filepath.Join(keyDir, w.FileName()+".key"), newBackupDir)
//
// 	//3. 备份地址数据库
// 	file.Copy(filepath.Join(dbPath, w.FileName()+".db"), newBackupDir)
//
// 	return newBackupDir, nil
// }
//
// //RestoreWallet 恢复钱包
// func RestoreWallet(keyFile, dbFile, datFile, password string) error {
//
// 	//根据流程，提供种子文件路径，wallet.dat文件的路径，钱包数据库文件的路径。
// 	//输入钱包密码。
// 	//先备份核心钱包原来的wallet.dat到临时文件夹。
// 	//关闭钱包节点，复制wallet.dat到钱包的data目录下。
// 	//启动钱包，通过GetCoreWalletinfo检查钱包是否启动了。
// 	//检查密码是否可以解析种子文件，是否可以解锁钱包。
// 	//如果密码错误，关闭钱包节点，恢复原钱包的wallet.dat。
// 	//重新启动钱包。
// 	//复制种子文件到data/bcc/key/。
// 	//复制钱包数据库文件到data/bcc/db/。
//
// 	var (
// 		restoreSuccess = false
// 		err            error
// 		key            *keystore.HDKey
// 		sleepTime      = 30 * time.Second
// 	)
//
// 	fmt.Printf("Validating key file... \n")
//
// 	//检查密码是否可以解析种子文件，是否可以解锁钱包。
// 	key, err = storage.GetKey("", keyFile, password)
// 	if err != nil {
// 		return errors.New("Passowrd is incorrect!")
// 	}
//
// 	//钱包当前的dat文件
// 	curretWDFile := filepath.Join(walletDataPath, "wallet.dat")
//
// 	//创建临时备份文件wallet.dat，备份
// 	tmpWalletDat := fmt.Sprintf("restore-walllet-%d.dat", time.Now().Unix())
// 	tmpWalletDat = filepath.Join(walletDataPath, tmpWalletDat)
//
// 	fmt.Printf("Backup current wallet.dat file... \n")
//
// 	err = BackupWalletData(tmpWalletDat)
// 	if err != nil {
// 		return err
// 	}
//
// 	//调试使用
// 	//file.Copy(curretWDFile, tmpWalletDat)
//
// 	fmt.Printf("Stop node server... \n")
//
// 	//关闭钱包节点
// 	stopNode()
// 	time.Sleep(sleepTime)
//
// 	fmt.Printf("Restore wallet.dat file... \n")
//
// 	//删除当前钱包文件
// 	file.Delete(curretWDFile)
//
// 	//恢复备份dat到钱包数据目录
// 	err = file.Copy(datFile, walletDataPath)
// 	if err != nil {
// 		return err
// 	}
//
// 	fmt.Printf("Start node server... \n")
//
// 	//重新启动钱包
// 	startNode()
// 	time.Sleep(sleepTime)
//
// 	fmt.Printf("Validating wallet password... \n")
//
// 	//检查wallet.dat是否可以解锁钱包
// 	err = UnlockWallet(password, 1)
// 	if err != nil {
// 		restoreSuccess = false
// 		err = errors.New("Password is incorrect!")
// 	} else {
// 		restoreSuccess = true
// 	}
//
// 	if restoreSuccess {
// 		/* 恢复成功 */
//
// 		fmt.Printf("Restore wallet key and datebase file... \n")
//
// 		//复制种子文件到data/bcc/key/
// 		file.MkdirAll(keyDir)
// 		file.Copy(keyFile, filepath.Join(keyDir, key.FileName()+".key"))
//
// 		//复制钱包数据库文件到data/bcc/db/
// 		file.MkdirAll(dbPath)
// 		file.Copy(dbFile, filepath.Join(dbPath, key.FileName()+".db"))
//
// 		fmt.Printf("Backup wallet has been restored. \n")
//
// 		err = nil
// 	} else {
// 		/* 恢复失败还远原来的文件 */
//
// 		fmt.Printf("Wallet unlock password is incorrect. \n")
//
// 		fmt.Printf("Stop node server... \n")
//
// 		//关闭钱包节点
// 		stopNode()
// 		time.Sleep(sleepTime)
//
// 		fmt.Printf("Restore original wallet.data... \n")
//
// 		//删除当前钱包文件
// 		file.Delete(curretWDFile)
//
// 		file.Copy(tmpWalletDat, curretWDFile)
//
// 		fmt.Printf("Start node server... \n")
//
// 		//重新启动钱包
// 		startNode()
// 		time.Sleep(sleepTime)
//
// 		fmt.Printf("Original wallet has been restored. \n")
//
// 	}
//
// 	//删除临时备份的dat文件
// 	file.Delete(tmpWalletDat)
//
// 	return err
// }
// //SummaryWallets 执行汇总流程
// func SummaryWallets() {
//
// 	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
//
// 	//读取参与汇总的钱包
// 	for wid, wallet := range walletsInSum {
//
// 		//重新加载utxo
// 		RebuildWalletUnspent(wid)
//
// 		//统计钱包最新余额
// 		wb := GetWalletBalance(wid)
//
// 		balance, _ := decimal.NewFromString(wb)
// 		//如果余额大于阀值，汇总的地址
// 		if balance.GreaterThan(threshold) {
//
// 			log.Printf("Summary account[%s]balance = %v \n", wallet.WalletID, balance)
// 			log.Printf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)
//
// 			txID, err := SendTransaction(wallet.WalletID, sumAddress, balance, wallet.Password, false)
// 			if err != nil {
// 				log.Printf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
// 				continue
// 			} else {
// 				log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, sumAddress, txID)
// 			}
// 		} else {
// 			log.Printf("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v\n", wallet.Alias, wallet.WalletID, balance, threshold)
// 		}
// 	}
//
// 	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
// }
//
// //AddWalletInSummary 添加汇总钱包账户
// func AddWalletInSummary(wid string, wallet *Wallet) {
// 	walletsInSum[wid] = wallet
// }
//
//
