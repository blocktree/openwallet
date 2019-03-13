/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package iota

import (
	"fmt"
	"log"
	"errors"
	"path/filepath"
	"github.com/iotaledger/giota"
	"github.com/blocktree/openwallet/console"
	"time"
)

type WalletManager struct{}

//func NewWalletManager() *WalletManager {
//	w := WalletManager{}
//	//w.blockscanner = NewBTCBlockScanner()
//	return &w
//}

//初始化配置流程
func (w *WalletManager) InitConfigFlow() error {

	file := filepath.Join(configFilePath, configFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's config.\n", file)
	return nil

	/*
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
		isTestNet  bool
	)

	for {

		fmt.Printf("[Start setup wallet config]\n")

		apiURL, err = console.InputText("Set node API url: ", true)
		if err != nil {
			return err
		}

		//删除末尾的/
		apiURL = strings.TrimSuffix(apiURL, "/")

		walletPath, err = console.InputText("Set wallet install path: ", false)
		if err != nil {
			return err
		}

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

		isTestNet, err = console.Stdin.PromptConfirm("The Network is TestNet? ")
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
		fmt.Printf("Wallet install path: %s\n", walletPath)
		fmt.Printf("Summary address: %s\n", sumAddress)
		fmt.Printf("Summary threshold: %s\n", threshold)
		fmt.Printf("Summary isTestNet: %v\n", isTestNet)
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

	_, filePath, err = newConfigFile(apiURL, walletPath, sumAddress, threshold, isTestNet)

	fmt.Printf("Config file create, file path: %s\n", filePath)

	return nil
	*/
}

//查看配置信息
func (w *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (w *WalletManager) CreateWalletFlow() error {

	var (
		//name     string
		err      error
	)

	//先加载是否有配置文件
	err = loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	//name, err = console.InputText("Enter wallet's name: ", true)

	seed := giota.NewSeed()

	fmt.Printf("\n")
	fmt.Printf("Wallet create successfully, seed: %s\n", seed)
	fmt.Printf("Please keep your seed in a safe place.\n")

	////备份seed
	//newBackupDir,err := BackupWallet(name,seed)
	//fmt.Printf("The seed has been backup in the path:%s\n", newBackupDir)

	return nil

}

//创建地址流程
func (w *WalletManager) CreateAddressFlow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//选择钱包
	seed, err := console.InputText("Please enter the seed of the wallet: ", true)
	if err != nil {
		return err
	}
	trytes,err:=giota.ToTrytes(seed)
	if err != nil{
		return err
	}

	// 输入地址数量
	start, err := console.InputStartNumber("Enter the start number of addresses: ", false)
	if err != nil {
		return err
	}
	startInt := int(start)

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}
	countInt := int(count)


	security := 2

	log.Printf("Start batch creation\n")
	log.Printf("-------------------------------------------------\n")

	backupFile,err := CreateAddresses(trytes,startInt,countInt,security)
	if err != nil{
		return err
	}

	log.Printf("-------------------------------------------------\n")
	log.Printf("All addresses have created, file path:%s\n", backupFile)
	return nil
}


// SummaryFollow 汇总流程
func (w *WalletManager) SummaryFollow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(sumAddress) == 0 {

		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	//输入要汇总的钱包的种子
	seed, err := console.InputText("Please enter the seed of the wallet which you want to sum: ", true)
	if err != nil {
		return err
	}

	// 输入tag
	note, err := console.InputText("Enter the tag (massage of this transaction): ", true)
	if err != nil {
		return err
	}
	tag,err := giota.ToTrytes(note)
	if err != nil {
		return err
	}

	//启动钱包汇总程序
	for {
		err := SummaryWallets(seed, giota.Address(sumAddress), tag)
		if err != nil {
			fmt.Printf("SummaryWallets expected err: %v\n", err)
			fmt.Printf("SummaryWallets will be started in 10 seconds.\n")
			time.Sleep(10000000000)
			continue
		}else {
			fmt.Printf("SummaryWallets will be started in 10 seconds.\n")
			time.Sleep(10000000000)
			continue
		}
	}

	return nil
}

//备份钱包流程
func (w *WalletManager) BackupWalletFlow() error {

	//var (
	//	err        error
	//	backupPath string
	//)
	//
	////先加载是否有配置文件
	//err = loadConfig()
	//if err != nil {
	//	return err
	//}
	//
	//list, err := GetWalletList()
	//if err != nil {
	//	return err
	//}
	//
	////打印钱包列表
	//printWalletList(list)
	//
	//fmt.Printf("[Please select a wallet to backup] \n")
	//
	////选择钱包
	//num, err := console.InputNumber("Enter wallet No. : ", true)
	//if err != nil {
	//	return err
	//}
	//
	//if int(num) >= len(list) {
	//	return errors.New("Input number is out of index! ")
	//}
	//
	//wallet := list[num]
	//
	//backupPath, err = BackupWallet(wallet.WalletID)
	//if err != nil {
	//	return err
	//}
	//
	////输出备份导出目录
	//fmt.Printf("Wallet backup file path: %s\n", backupPath)

	return nil

}

//SendTXFlow 发送交易
func (w *WalletManager) TransferFlow() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//输入钱包seed
	seed, err := console.InputText("Please enter the seed of your wallet: ", true)
	if err != nil {
		return err
	}

	// 输入发送地址
	addr, err := console.InputText("Enter the receive address: ", false)
	if err != nil {
		return err
	}
	address := giota.Address(addr)

	// 输入iota数量
	amount, err := console.InputNumber("Enter the IOTA amount: ", false)
	if err != nil {
		return err
	}
	value := int64(amount)

	// 输入tag
	note, err := console.InputText("Enter the tag (massage of this transaction): ", true)
	if err != nil {
		return err
	}
	tag,err := giota.ToTrytes(note)
	if err != nil {
		return err
	}

	err = SendTransaction(seed, address, value, tag)
	if err != nil{
		return err
	}

	log.Printf("SendTransaction() successfully.\n")

	return nil
}

//GetWalletList 获取钱包列表
func (w *WalletManager) GetWalletList() error {

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//选择钱包
	seed, err := console.InputText("Please enter the seed of the wallet: ", true)
	if err != nil {
		return err
	}

	if len(seed) != 81{
		fmt.Printf("The input seed is wrong.")
		return nil
	}

	fmt.Printf("Checking......please wait for a monent......\n")

	_,adrs,totalBalance,err := GetWalletInfo(seed)
	if err != nil {
		return err
	}

	if len(adrs)==0{
		fmt.Printf("There is no used addresses of this wallet.\n")
	}else {
		fmt.Printf("Used addresses of this wallet:\n")
		for i:=0;i<len(adrs);i++{
			fmt.Printf("%s\n",adrs[i])
		}
	}


	fmt.Printf("Total balance of this wallet: %v\n",totalBalance)

	return nil
}

//RestoreWalletFlow 恢复钱包流程
func (w *WalletManager) RestoreWalletFlow() error {

	//var (
	//	err      error
	//	keyFile  string
	//	dbFile   string
	//	datFile  string
	//	password string
	//)
	//
	////先加载是否有配置文件
	//err = loadConfig()
	//if err != nil {
	//	return err
	//}
	//
	////输入恢复文件路径
	//keyFile, err = console.InputText("Enter backup key file path: ", true)
	//if err != nil {
	//	return err
	//}
	//
	//dbFile, err = console.InputText("Enter backup db file path: ", true)
	//if err != nil {
	//	return err
	//}
	//
	//datFile, err = console.InputText("Enter backup wallet.dat file path: ", true)
	//if err != nil {
	//	return err
	//}
	//
	//password, err = console.InputPassword(false, 8)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Printf("Wallet restoring, please wait a moment...\n")
	//err = RestoreWallet(keyFile, dbFile, datFile, password)
	//if err != nil {
	//	return err
	//}
	//
	////输出备份导出目录
	//fmt.Printf("Restore wallet successfully.\n")

	return nil

}
/*
//InstallNode 安装节点
func (w *WalletManager) InstallNodeFlow() error {
	return errors.New("Install node is unsupport now. ")
}

//InitNodeConfig 初始化节点配置文件
func (w *WalletManager) InitNodeConfigFlow() error {
	return errors.New("Install node is unsupport now. ")
}

//StartNodeFlow 开启节点
func (w *WalletManager) StartNodeFlow() error {

	return startNode()
}

//StopNodeFlow 关闭节点
func (w *WalletManager) StopNodeFlow() error {

	return stopNode()
}

//RestartNodeFlow 重启节点
func (w *WalletManager) RestartNodeFlow() error {
	return errors.New("Install node is unsupport now. ")
}

//ShowNodeInfo 显示节点信息
func (w *WalletManager) ShowNodeInfo() error {
	return errors.New("Install node is unsupport now. ")
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
*/