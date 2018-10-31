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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"github.com/tyler-smith/go-bip39"
	"log"
	"path/filepath"
	"time"
	"github.com/bndr/gotabulate"
)

const (
	maxAddresNum = 10000000
)

var (
	//钱包服务API
	serverAPI = "https://127.0.0.1:10026/api/"
	//钱包主链私钥文件路径
	walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
	//参与汇总的钱包
	walletsInSum = make(map[string]*Wallet)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(10000).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(100).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(0.3).Mul(coinDecimal)
	//汇总地址
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
)

//StartWalletProcess 启动钱包进程
func StartWalletProcess() {

}

//StopWalletProcess 停止钱包进程
func StopWalletProcess() {

}

//GetWalletInfo 获取钱包信息
//wid 钱包id，可选
func GetWalletInfo(wid ...string) ([]*Wallet, error) {

	var (
		err     error
		wallets = make([]*Wallet, 0)
	)

	//调用服务
	result := callGetWalletAPI(wid...)
	err = isError(result)

	content := gjson.GetBytes(result, "Right")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			wallets = append(wallets, NewWalletForV0(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		wallets = append(wallets, NewWalletForV0(content))
	}

	return wallets, err
}

//CreateNewWallet 创建新钱包
func CreateNewWallet(name, mnemonic, password string) error {

	var (
		err error
	)

	//密钥32
	h := common.NewString(password).SHA256()

	//调用服务创建钱包
	result := callCreateWalletAPI(name, mnemonic, h, true)
	err = isError(result)
	if err != nil {
		return err
	}

	//log.Printf("新钱包助记词：%v", mnemonic)

	content := gjson.GetBytes(result, "Right")
	wallet := NewWalletForV0(content)
	wallet.Password = password
	wallet.Mnemonic = mnemonic
	return exportWalletToFile(wallet)
}

//CreateBatchAddress 批量创建地址
func CreateBatchAddress(aid, password string, count uint) ([]*Address, string, error) {

	var (
		err          error
		done         uint
		producerDone uint
		synCount     uint = 100
	)

	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405") + ".txt"
	filePath := addressDir + filename
	//生产通道
	producer := make(chan *Address)

	//消费通道
	worker := createAddressSaveChan(filename)

	values := make([]*Address, 0)
	addresses := make([]*Address, 0)

	//完成标记
	done = 0

	//生产完成标记
	producerDone = 0

	// 以下使用线程数量以及线程负载均衡

	//每个线程内循环的数量
	runCount := count / synCount

	if runCount == 0 {
		//fmt.Printf("runCount 小于线程数")
		for i := uint(0); i < count; i++ {

			go func() {
				// 请求地址
				getAddressWrok(aid, password, producer, err)
			}()

		}
	} else {

		for i := uint(0); i < synCount; i++ {
			//fmt.Printf("runCount 启动线程 %d 共：%d ", i, runCount)
			go func(runCount uint) {
				for i := uint(0); i < runCount; i++ {
					getAddressWrok(aid, password, producer, err)

				}
			}(runCount)
		}
		//余数不为0，泽直接开启线程运行余下数量
		if otherCount := count % synCount; otherCount != 0 {
			//fmt.Printf("余数为 %d ", otherCount)
			go func(otherCount uint) {
				for i := uint(0); i < otherCount; i++ {
					//fmt.Printf("余数运行 %d ", i)
					getAddressWrok(aid, password, producer, err)

				}
			}(otherCount)
		}
	}

	//以下使用生产消费模式

	for {
		var activeWorker chan<- *Address
		var activeValue *Address
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]
		}

		select {
		case n := <-producer:
			values = append(values, n)
			addresses = append(addresses, n)
			producerDone++
			//log.Printf("生成 %d",done)
		case activeWorker <- activeValue:
			values = values[1:]
			done++
			//log.Printf("完成多线程 %d",done)
			if done == count {
				fmt.Printf("All thread completed!")
				return addresses, filePath, nil
			}

		}
	}

	return addresses, filePath, nil
}

//http获取地址
func getAddressWrok(aid string, passphrase string, producer chan *Address, err error) {
	result := callCreateNewAddressAPI(aid, passphrase)
	err = isError(result)
	if err != nil {
		log.Printf("Create address failed! ")
		return
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAddressV0(content)
	fmt.Printf("Create：	%s\n", a.Address)
	producer <- a
}

//保存地址
func saveAddressWork(address chan *Address, filename string) {

	for a := range address {
		exportAddressToFile(a, filename)
		fmt.Printf("Save:	%s\n", a.Address)
	}
	//return addrs, filename, nil
}

//保存地址通道
func createAddressSaveChan(filename string) chan<- *Address {
	address := make(chan *Address)
	go saveAddressWork(address, filename)
	return address
}

//CreateNewAccount 根据钱包wid创建单个账户
func CreateNewAccount(name, wid, passphrase string) (*Account, error) {

	var (
		err error
	)

	//调用服务创建新账户
	result := callCreateNewAccountAPI(name, wid, passphrase)
	err = isError(result)
	if err != nil {
		return nil, err
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAccountV0(content)
	return a, err
}

//GetAccountInfo 获取用户信息
func GetAccountInfo(aid ...string) ([]*Account, error) {

	var (
		err      error
		accounts = make([]*Account, 0)
	)

	//调用服务
	result := callGetAccountsAPI(aid...)
	err = isError(result)

	content := gjson.GetBytes(result, "Right")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			accounts = append(accounts, NewAccountV0(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		accounts = append(accounts, NewAccountV0(content))
	}

	return accounts, err
}

//GetAddressInfo 获取指定aid用户的地址组
func GetAddressInfo(aid string) ([]*Address, error) {

	var (
		err     error
		address = make([]*Address, 0)
	)

	//调用服务
	result := callGetAccountByIDAPI(aid)
	err = isError(result)

	content := gjson.GetBytes(result, "Right.caAddresses")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			address = append(address, NewAddressV0(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		address = append(address, NewAddressV0(content))
	}

	return address, err
}

//SendTx 发送交易
func SendTx(from, to string, amount uint64, password string) (*Transaction, error) {

	//输入密码
	//password, err := console.InputPassword(false)
	//h := common.NewString(password).SHA256()

	//调用服务创建新账户
	result, err := callSendTxAPI(from, to, amount, password)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}
	content := gjson.GetBytes(result, "Right")
	t := NewTransactionV0(content)
	return t, nil
}

//SummaryTxFlow 执行汇总流程
func SummaryWallets() {

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range walletsInSum {

		//统计钱包最新余额
		ws, err := GetWalletInfo(wid)
		if err != nil {
			log.Printf("Can not find wallet information：%v\n", err)
			continue
		}
		if len(ws) > 0 {
			w := ws[0]
			balance, _ := decimal.NewFromString(common.NewString(w.Balance).String())
			//如果余额大于阀值，汇总的地址
			if balance.GreaterThan(threshold) {
				//汇总所有有钱的账户
				accounts, err := GetAccountInfo(w.WalletID)
				if err != nil {
					log.Printf("Can not find account information：%v\n", err)
					continue
				}

				for _, a := range accounts {
					//大于最小额度才转账
					sendAmount, _ := decimal.NewFromString(common.NewString(a.Amount).String())
					if sendAmount.GreaterThan(minSendAmount) {
						log.Printf("Summary account[%s]balance = %v \n", a.AcountID, sendAmount.Div(coinDecimal))
						log.Printf("Summary account[%s]Start Send Transaction\n", a.AcountID)
						tx, err := SendTx(a.AcountID, sumAddress, uint64(sendAmount.Sub(minFees).IntPart()), wallet.Password)
						if err != nil {
							log.Printf("Summary account[%s]unexpected error: %v\n", a.AcountID, err)
							continue
						} else {
							log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", a.AcountID, sumAddress, tx.TxID)
						}
					}
				}
			} else {
				log.Printf("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v\n", w.Name, w.WalletID, balance.Div(coinDecimal), threshold.Div(coinDecimal))
			}
		}
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

func AddWalletInSummary(wid string, wallet *Wallet) {
	walletsInSum[wid] = wallet
}

//CreateAddress 给指定账户创建地址
func CreateAddress(aid string, passphrase string) (*Address, error) {
	result := callCreateNewAddressAPI(aid, passphrase)
	err := isError(result)
	if err != nil {
		log.Printf("Create address failed! ")
		return nil, err
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAddressV0(content)
	return a, nil
}

//EstimateFees 计算预估手续费
func EstimateFees(from, to string, amount uint64) (uint64, error) {

	result := callEstimateFeesAPI(from, to, amount)
	err := isError(result)
	if err != nil {
		return 0, nil
	}

	fees := gjson.GetBytes(result, "Right.getCCoin")

	return fees.Uint(), nil
}

//钱包恢复机制

//genMnemonic 随机创建密钥
func genMnemonic() string {
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}

//isError 是否报错
func isError(result []byte) error {
	var (
		err error
	)

	/*
		//failed 返回错误
		{
			"Left": {
				"tag": "RequestError",
				"contents": "Passphrase doesn't match"
			}
		}
	*/

	//V0的错误信息存放在Left上
	if !gjson.GetBytes(result, "Left").Exists() {
		return nil
	}

	err = errors.New(gjson.GetBytes(result, "Left.contents").String())

	return err
}

//exportAddressToFile 导出地址到文件中
func exportAddressToFile(a *Address, filename string) {
	file.MkdirAll(addressDir)
	filepath := addressDir + filename
	file.WriteFile(filepath, []byte(a.Address+"\n"), true)
}

//exportWalletToFile 导出钱包到文件
func exportWalletToFile(w *Wallet) error {

	var (
		err     error
		content []byte
	)

	filename := fmt.Sprintf("wallet-%s-%s.json", w.Name, w.WalletID)

	file.MkdirAll(keyDir)
	filepath := filepath.Join(keyDir, filename)

	//把钱包写入到文件进行备份
	content, err = json.MarshalIndent(w, "", "\t")
	if err != nil {
		return errors.New("Wallet key encode json failed! ")
	}

	if !file.WriteFile(filepath, content, true) {
		return errors.New("Wallet key write to file failed! ")
	}

	log.Printf("================================================\n")

	log.Printf("Wallet key backup successfully，file path: %s\n", filepath)

	return nil
}

//inputNumber 输入地址数量
func inputNumber() uint64 {

	var (
		count uint64 = 0 // 输入的创建数量
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("Enter the number of addresses you want: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return 0
		}
		count = common.NewString(line).UInt64()
		if count < 1 {
			log.Printf("Input number must be greater than 0!\n")
			continue
		}
		break
	}

	return count
}

//inputWID 输入钱包ID
func inputWID() string {

	var (
		wid string
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("Enter wallet ID: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return ""
		}
		if len(line) == 0 {
			log.Printf("Wallet ID is empty, please re-enter!\n")
			continue
		}
		wid = line
		break
	}

	return wid
}

//loadConfig 读取配置
func loadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("json", absFile)
	if err != nil {
		return fmt.Errorf("Config is not setup. Please run 'wmd config -s <symbol>', %v ", err)
	}

	serverAPI = c.String("apiURL")
	walletPath = c.String("walletPath")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	threshold = threshold.Mul(coinDecimal)
	minSendAmount, _ = decimal.NewFromString(c.String("minSendAmount"))
	minSendAmount = minSendAmount.Mul(coinDecimal)
	minFees, _ = decimal.NewFromString(c.String("minFees"))
	minFees = minFees.Mul(coinDecimal)
	sumAddress = c.String("sumAddress")

	return nil
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		balance, err := decimal.NewFromString(w.Balance)
		if err != nil {
			continue
		}
		balance = balance.Div(coinDecimal)
		tableInfo = append(tableInfo, []interface{}{
			i,w.WalletID,w.Name,balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "WID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}