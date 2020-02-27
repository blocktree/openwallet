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
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/bndr/gotabulate"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"path/filepath"
	"time"
)

const (
	maxAddresNum = 10000
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	WalletClient *Client       // 节点客户端
	Config       *WalletConfig //钱包管理配置
	Log          *log.OWLogger //日志工具
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol)
	wm.Log = log.NewOWLogger(wm.Symbol())
	return &wm
}

//GetInfo
func (wm *WalletManager) GetInfo() (*gjson.Result, error) {

	result, err := wm.WalletClient.Call("getinfo", struct{}{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

//GetNewAddress
func (wm *WalletManager) GetNewAddress() (*Address, error) {

	result, err := wm.WalletClient.Call("getnewaddress", struct{}{})
	if err != nil {
		return nil, err
	}

	address := NewAddress(*result)

	return address, nil
}

//GetChangeAddress
func (wm *WalletManager) GetChangeAddress() (*Address, error) {

	result, err := wm.WalletClient.Call("getchangeaddress", struct{}{})
	if err != nil {
		return nil, err
	}

	address := NewAddress(*result)

	return address, nil
}

//GetBalance
func (wm *WalletManager) GetBalance() (*Balance, error) {

	result, err := wm.WalletClient.Call("getbalance", struct{}{})
	if err != nil {
		return nil, err
	}

	balance := NewBalance(result.Get("base"))

	return balance, nil
}

//SendToAddress
func (wm *WalletManager) SendToAddress(address string, amount int64) (string, error) {

	result, err := wm.WalletClient.Call("sendtoaddress", []interface{}{
		address,
		amount,
	})
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

//CreateBatchAddress 批量创建地址
func (wm *WalletManager) CreateBatchAddress(count uint64) (string, []*Address, error) {

	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//生成默认的找零地址
	_, err := wm.GetChangeAddress()
	if err != nil {
		return "", nil, err
	}

	timestamp := time.Now()
	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
	filePath := filepath.Join(wm.Config.addressDir, filename)

	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*Address)
	defer close(worker)

	//保存地址过程
	saveAddressWork := func(addresses chan []*Address, filename string) {

		for {
			//回收创建的地址
			getAddrs := <-addresses

			//导出一批地址
			wm.exportAddressToFile(getAddrs, filename)

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	/*	开启导出的线程，监听新地址，批量导出	*/

	go saveAddressWork(worker, filePath)

	/*	计算synCount个线程，内部运行的次数	*/

	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {

		for i := uint64(0); i < synCount; i++ {

			//开始创建地址
			wm.Log.Std.Info("Start create address thread[%d]", i)
			s := i * runCount
			e := (i + 1) * runCount
			go wm.createAddressWork(producer, s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		wm.Log.Std.Info("Start create address thread[REST]")
		s := count - otherCount
		e := count
		go wm.createAddressWork(producer, s, e)

		shouldDone++
	}

	values := make([][]*Address, 0)
	outputAddress := make([]*Address, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- []*Address
		var activeValue []*Address

		//当数据队列有数据时，释放顶部，激活消费
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
			outputAddress = append(outputAddress, pa...)
			//wm.Log.Std.Info("completed %d", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//wm.Log.Std.Info("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			wm.Log.Std.Info("All addresses have been created!")
			return filePath, outputAddress, nil
		}
	}

	return filePath, outputAddress, nil
}

//createAddressWork 创建地址过程
func (wm *WalletManager) createAddressWork(producer chan<- []*Address, start, end uint64) {

	runAddress := make([]*Address, 0)

	for i := start; i < end; i++ {
		// 生成地址
		address, err := wm.GetNewAddress()
		if err != nil {
			wm.Log.Std.Info("Create new privKey failed unexpected error: %v", err)
			continue
		}

		runAddress = append(runAddress, address)
	}

	//生成完成
	producer <- runAddress
}

//exportAddressToFile 导出地址到文件中
func (wm *WalletManager) exportAddressToFile(addrs []*Address, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		wm.Log.Std.Info("Export: %s ", a.Address)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(wm.Config.addressDir)
	file.WriteFile(filePath, []byte(content), true)
}

//BackupWallet 备份数据
func (wm *WalletManager) BackupWallet() (string, error) {

	//备份钱包，配置有密钥路径，采用配置文件的，没有采用默认的路径备份私钥。
	//默认备份路径 /root/.config/lux-headless/keys.json
	//默认备份路径 /root/.config/lux-headless/luxalpa.sqlite

	dataPath := ""
	if file.Exists(wm.Config.WalletDataPath) {
		dataPath = wm.Config.WalletDataPath
	} else {
		dataPath = filepath.Join("/", "root", ".config", "lux-headless")
	}

	//创建备份文件夹
	newBackupDir := filepath.Join(wm.Config.backupDir, common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)
	//keys.json

	keyFile := filepath.Join(dataPath, "keys.json")
	//备份助记词文件
	err := file.Copy(keyFile, newBackupDir)
	if err != nil {
		return "", err
	}

	confFile := filepath.Join(dataPath, "conf.json")
	//conf.json
	err = file.Copy(confFile, newBackupDir)
	if err != nil {
		return "", err
	}

	sqliteFile := filepath.Join(dataPath, "luxalpa.sqlite")
	//备份sqlite数据库
	err = file.Copy(sqliteFile, newBackupDir)
	if err != nil {
		return "", err
	}

	return newBackupDir, nil
}

//RestoreWallet 恢复钱包
func (wm *WalletManager) RestoreWallet(keyFile string, confFile string) error {

	//备份钱包，配置有密钥路径，采用配置文件的，没有采用默认的路径备份私钥。
	//默认备份路径 /root/.config/lux-headless/key.json

	//恢复key.json到钱包数据目录
	err := file.Copy(keyFile, wm.Config.WalletDataPath)
	if err != nil {
		return err
	}

	//恢复conf.json到钱包数据目录
	err = file.Copy(confFile, wm.Config.WalletDataPath)
	if err != nil {
		return err
	}

	log.Warningf("Please import [my_addresses] table data of your backup sqlite db into path: %s/luxalpa.sqlite", wm.Config.WalletDataPath)

	return nil
}

//SummaryWallets 执行汇总流程
func (wm *WalletManager) SummaryWallets() {

	wm.Log.Std.Info("[Summary Wallet Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	//统计钱包最新余额
	balance, err := wm.GetBalance()
	if err != nil {
		wm.Log.Std.Info("Summary wallet failed, unexpected error: %v ", err)
	} else {

		stable, _ := decimal.NewFromString(balance.Stable)
		pending, _ := decimal.NewFromString(balance.Pending)
		totalBalance := stable.Add(pending)
		totalBalance = totalBalance.Shift(-wm.Decimal())
		//如果余额大于阀值，汇总的地址
		if totalBalance.GreaterThan(wm.Config.Threshold) {

			wm.Log.Std.Info("Summary wallet balance = %v ", totalBalance.String())
			wm.Log.Std.Info("Summary wallet Start Send Transaction")
			fees, _ := decimal.NewFromString(wm.Config.MinFees)
			amount := totalBalance.Sub(fees)
			txID, err := wm.SendToAddress(wm.Config.SumAddress, amount.Shift(wm.Decimal()).IntPart())
			if err != nil {
				wm.Log.Std.Info("Summary wallet unexpected error: %v", err)
			} else {
				wm.Log.Std.Info("Summary wallet successfully，Received Address[%s], Send Amount: %s, TXID：%s", wm.Config.SumAddress, amount.String(), txID)
			}
		} else {
			wm.Log.Std.Info("Wallet wallet Current Balance: %v，below threshold: %v", totalBalance.String(), wm.Config.Threshold)
		}
	}

	wm.Log.Std.Info("[Summary Wallet end]------%s", common.TimeFormat("2006-01-02 15:04:05"))
}

//打印钱包列表
func (wm *WalletManager) printWalletList(balance *Balance) {

	tableInfo := make([][]interface{}, 0)

	stable, _ := decimal.NewFromString(balance.Stable)
	pending, _ := decimal.NewFromString(balance.Pending)
	totalBalance := stable.Add(pending)

	tableInfo = append(tableInfo, []interface{}{
		stable.Shift(-wm.Decimal()).String(),
		pending.Shift(-wm.Decimal()).String(),
		totalBalance.Shift(-wm.Decimal()).String(),
	})

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"Stable", "Pending", "Total Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}
