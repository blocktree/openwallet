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
	"github.com/blocktree/OpenWallet/timer"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"github.com/tyler-smith/go-bip39"
	"log"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxAddresNum = 1000000
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

//CreateNewWalletFlow 创建钱包流程
func CreateNewWalletFlow() error {

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
					fmt.Printf("余数运行 %d ", i)
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
				fmt.Printf("完成多线程!")
				return addresses, filename, nil
			}

		}
	}
	filePath := addressDir + filename
	return addresses, filePath, nil
}

//http获取地址
func getAddressWrok(aid string, passphrase string, producer chan *Address, err error) {
	result := callCreateNewAddressAPI(aid, passphrase)
	err = isError(result)
	if err != nil {
		log.Printf("生成地址发生错误")
		return
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAddressV0(content)
	fmt.Printf("生成地址：	%s\n", a.Address)
	producer <- a
}

//保存地址
func saveAddressWork(address chan *Address, filename string) {

	for a := range address {
		exportAddressToFile(a, filename)
		fmt.Printf("保存地址:	%s\n", a.Address)
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

//CreateAddressFlow 创建地址流程
func CreateAddressFlow() error {

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

//SendTx 发送交易
func SendTx(from, to string, amount uint64, password string) (*Transaction, error) {

	var (
		err error
	)
	//输入密码
	//password, err := console.InputPassword(false)
	//h := common.NewString(password).SHA256()

	//调用服务创建新账户
	result := callSendTxAPI(from, to, amount, password)
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

	log.Printf("[钱包汇总计算开始]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range walletsInSum {

		//统计钱包最新余额
		ws, err := GetWalletInfo(wid)
		if err != nil {
			log.Printf("无法查找钱包信息：%v\n", err)
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
					log.Printf("无法查找账户信息：%v\n", err)
					continue
				}

				for _, a := range accounts {
					//大于最小额度才转账
					sendAmount, _ := decimal.NewFromString(common.NewString(a.Amount).String())
					if sendAmount.GreaterThan(minSendAmount) {
						log.Printf("汇总账户[%s]余额 = %v \n", a.AcountID, sendAmount.Div(coinDecimal))
						log.Printf("汇总账户[%s]开始发送交易\n", a.AcountID)
						tx, err := SendTx(a.AcountID, sumAddress, uint64(sendAmount.Sub(minFees).IntPart()), wallet.Password)
						if err != nil {
							log.Printf("汇总账户[%s]出错：%v\n", a.AcountID, err)
							continue
						} else {
							log.Printf("汇总账户[%s]成功，发送地址[%s], TXID：%s\n", a.AcountID, sumAddress, tx.TxID)
						}
					}
				}
			} else {
				log.Printf("钱包[%s]-[%s]当前余额%v，未达到阀值%v\n", w.Name, w.WalletID, balance.Div(coinDecimal), threshold.Div(coinDecimal))
			}
		}
	}

	log.Printf("[钱包汇总计算结束]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

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
func SummaryFollow() error {

	var (
		endRunning = make(chan bool, 1)
	)

	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
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

func AddWalletInSummary(wid string, wallet *Wallet) {
	walletsInSum[wid] = wallet
}

//BackupWalletkey 备份钱包密钥文件
func BackupWalletkey() error {

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

//CreateAddress 给指定账户创建地址
func CreateAddress(aid string, passphrase string) (*Address, error) {
	result := callCreateNewAddressAPI(aid, passphrase)
	err := isError(result)
	if err != nil {
		log.Printf("生成地址发生错误")
		return nil, err
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAddressV0(content)
	return a, nil
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
		return errors.New("钱包信息序列化json失败")
	}

	if !file.WriteFile(filepath, content, true) {
		return errors.New("钱包密钥信息写入文件失败")
	}

	log.Printf("================================================\n")

	log.Printf("钱包创建成功，导出路径:%s\n", filepath)

	return nil
}

//inputNumber 输入地址数量
func inputNumber() uint64 {

	var (
		count uint64 = 0 // 输入的创建数量
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("输入需要创建的地址数量: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return 0
		}
		count = common.NewString(line).UInt64()
		if count < 1 {
			log.Printf("输入地址数量必须大于0")
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
		line, err := console.Stdin.PromptInput("输入钱包WID: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return ""
		}
		if len(line) == 0 {
			log.Printf("钱包WID不能为空，请重新输入")
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
		return errors.New("配置文件未创建，请执行 wmd config -s <symbol> ")
	}

	serverAPI = c.String("apiURL")
	walletPath = c.String("walletPath")
	thresholdTemp := common.NewString(c.String("threshold")).Float64()
	threshold = decimal.NewFromFloat(thresholdTemp).Mul(coinDecimal)
	minSendAmountTemp := common.NewString(c.String("minSendAmount")).Float64()
	minSendAmount = decimal.NewFromFloat(minSendAmountTemp).Mul(coinDecimal)
	minFeesTemp := common.NewString(c.String("minFees")).Float64()
	minFees = decimal.NewFromFloat(minFeesTemp).Mul(coinDecimal)
	sumAddress = c.String("sumAddress")

	return nil
}
