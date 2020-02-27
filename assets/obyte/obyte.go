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

package obyte

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/console"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/timer"
	"github.com/shopspring/decimal"
	"path/filepath"
	"strings"
	"time"
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

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	_, err = wm.GetInfo()
	if err != nil {
		return err
	}

	log.Info("Wallet has been created, only one wallet can work")

	return nil

}

//创建地址流程
func (wm *WalletManager) CreateAddressFlow() error {

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddresNum {
		return fmt.Errorf(fmt.Sprintf("The number of addresses can not exceed %d", maxAddresNum))
	}

	log.Std.Info("Start batch creation")
	log.Std.Info("-------------------------------------------------")

	filePath, _, err := wm.CreateBatchAddress(count)
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

		return fmt.Errorf(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	if wm.Config.CycleSeconds == 0 {
		wm.Config.CycleSeconds = 30 * 1000
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
		err        error
		backupPath string
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	backupPath, err = wm.BackupWallet()
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: %s\n", backupPath)

	return nil

}

//SendTXFlow 发送交易
func (wm *WalletManager) TransferFlow() error {

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入发送数量
	amount, err := console.InputRealNumber("Enter amount to send: ", true)
	if err != nil {
		return err
	}

	atculAmount, _ := decimal.NewFromString(amount)
	balance, err := wm.GetBalance()
	if err != nil {
		return err
	}

	stable, _ := decimal.NewFromString(balance.Stable)
	pending, _ := decimal.NewFromString(balance.Pending)
	totalBalance := stable.Add(pending)

	if atculAmount.GreaterThan(totalBalance) {
		return fmt.Errorf("Input amount is greater than balance! ")
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	sendAmount := atculAmount.Shift(wm.Decimal()).IntPart()

	//建立交易单
	txID, err := wm.SendToAddress(receiver, sendAmount)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully, TXID：%s\n", txID)

	return nil
}

//GetWalletList 获取钱包列表
func (wm *WalletManager) GetWalletList() error {

	//先加载是否有配置文件
	err := wm.LoadConfig()
	if err != nil {
		return err
	}

	balance, err := wm.GetBalance()
	if err != nil {
		return err
	}

	//打印钱包列表
	wm.printWalletList(balance)

	return nil
}

//RestoreWalletFlow 恢复钱包流程
func (wm *WalletManager) RestoreWalletFlow() error {

	var (
		err      error
		keyFile  string
		confFile string
	)

	//先加载是否有配置文件
	err = wm.LoadConfig()
	if err != nil {
		return err
	}

	//输入恢复文件路径
	keyFile, err = console.InputText("Enter backup key file path: ", true)
	if err != nil {
		return err
	}

	//输入恢复文件路径
	confFile, err = console.InputText("Enter backup conf file path: ", true)
	if err != nil {
		return err
	}

	err = wm.RestoreWallet(keyFile, confFile)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Restore wallet successfully.\n")

	return nil

}

//loadConfig 读取配置
func (wm *WalletManager) LoadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)

	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return fmt.Errorf("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	wm.LoadAssetsConfig(c)

	return nil
}

//LoadAssetsConfig 加载外部配置
func (wm *WalletManager) LoadAssetsConfig(c config.Configer) error {

	//wm.Config.Symbol = c.String("symbol")
	wm.Config.ServerAPI = c.String("serverAPI")
	wm.Config.Threshold, _ = decimal.NewFromString(c.String("threshold"))
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.WalletDataPath = c.String("walletDataPath")
	wm.Config.CurveType = uint32(c.DefaultInt64("curveType", int64(CurveType)))
	wm.Config.addressDir = filepath.Join("data", strings.ToLower(wm.Config.Symbol), "address")
	//配置文件路径
	wm.Config.configFilePath = filepath.Join("conf")
	//配置文件名
	wm.Config.configFileName = wm.Config.Symbol + ".ini"
	//备份路径
	wm.Config.backupDir = filepath.Join("data", strings.ToLower(wm.Config.Symbol), "backup")
	wm.Config.CoinDecimals = int32(c.DefaultInt64("coinDecimals", 6))
	cyclesec := c.String("cycleSeconds")
	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)
	wm.Config.MinFees = c.String("minFees")
	wm.WalletClient = NewClient(wm.Config.ServerAPI, "", false)
	return nil
}

//InitAssetsConfig 初始化默认配置
func (wm *WalletManager) InitAssetsConfig() (config.Configer, error) {
	return config.NewConfigData("ini", []byte(wm.Config.DefaultConfig))
}

//GetAssetsLogger 获取资产账户日志工具
func (wm *WalletManager) GetAssetsLogger() *log.OWLogger {
	return wm.Log
}

//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}

//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "Obyte"
}

//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}

//小数位精度
func (wm *WalletManager) Decimal() int32 {
	return wm.Config.CoinDecimals
}
