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
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tyler-smith/go-bip39"
	"log"
	"strings"
)

const (
	exportBackupDir = "./data/"
)

var (
	//参与汇总的钱包
	walletsInSum = make(map[string]*Wallet)
	//汇总阀值
	threshold uint64 = 3 * 1000000
	//最小转账额度
	minSendAmount uint64 = 1 * 1000000
	//最小矿工费
	minFees uint64 = 0.3 * 1000000
	//汇总地址
	sumAddress = "DdzFFzCqrhsgCwX9VGWZAdAnPfeVwUzoPDAqQJfXE3DxKKFCTYmtGD9CrKjvGu7VjebGoCqPHN7DtkF1VhEvJbPhg2BrfhT5hkyBvjvZ"
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
	result := callCreateWalletAPI(name, mnemonic, h)
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
		err   error
		done uint
		producerDone uint
		synCount uint  = 50
	)

	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405") + ".txt"

	//生产通道
	producer := make(chan *Address)

	//消费通道
	worker := createAddressSaveChan(count,filename)

	values := make([]*Address, 0)
	addresses := make([]*Address, 0)

	//完成标记
	done = 0

	//生产完成标记
	producerDone = 0

	// 以下使用线程数量以及线程负载均衡

	//每个线程内循环的数量
	runCount := count/synCount


	if runCount == 0{
		for i := uint(0); i < count; i++ {

			go func() {
				// 请求地址
				getAddressWrok(aid,password,producer,err)
			}()
		}
	}else{
		for i := uint(0); i < synCount; i++ {

			go func(runCount uint) {
				for i := uint(0); i < runCount; i++ {
						getAddressWrok(aid,password,producer,err)

				}
			}(runCount)
		}
		//余数不为0，泽直接开启线程运行余下数量
		if otherCount := count%synCount;otherCount!=0{
			go func(otherCount uint) {
				for i := uint(0); i < otherCount; i++ {
						getAddressWrok(aid,password,producer,err)

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
			addresses = append(addresses,n)
			producerDone++
			log.Printf("生成 %d",done)
		case activeWorker <- activeValue:
			values = values[1:]
			done++
			log.Printf("完成多线程 %d",done)
			if done == count {
				log.Printf("完成多线程!")
				return addresses, filename, nil
			}

		}
	}
	return addresses, filename, nil
}

//http获取地址
func getAddressWrok(aid string,passphrase string,producer chan *Address,err error){
	result := callCreateNewAddressAPI(aid, passphrase)
	err = isError(result)
	if err != nil {
		log.Printf("生成地址发生错误")
		return
	}
	content := gjson.GetBytes(result, "Right")
	a := NewAddressV0(content)
	log.Printf("生成地址：	%s\n", a.Address)
	producer <- a
}

//保存地址
func saveAddressWork (address chan *Address,count uint,filename string){

	addrs := make([]*Address, 0)

	for a := range address{
		exportAddressToFile(a, filename)
		addrs = append(addrs, a)
		log.Printf("save	%s\n", a.Address)
	}
	//return addrs, filename, nil
}

//保存地址通道
func createAddressSaveChan (count uint,filename string)chan<- *Address{
	address := make(chan *Address)
	go saveAddressWork(address,count,filename)
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
		return errors.Errorf("创建地址数量不能超过%d\n", maxAddresNum)
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
			balance := common.NewString(w.Balance).UInt64()
			//如果余额大于阀值，汇总的地址
			if balance > threshold {
				//汇总所有有钱的账户
				accounts, err := GetAccountInfo(w.WalletID)
				if err != nil {
					log.Printf("无法查找账户信息：%v\n", err)
					continue
				}

				for _, a := range accounts {
					//大于最小额度才转账
					sendAmount := common.NewString(a.Amount).UInt64()
					if sendAmount > minSendAmount {
						log.Printf("汇总账户[%s]余额 = %d \n", a.AcountID, sendAmount)
						log.Printf("汇总账户[%s]开始发送交易\n", a.AcountID)
						tx, err := SendTx(a.AcountID, sumAddress, sendAmount - minFees, wallet.Password)
						if err != nil {
							log.Printf("汇总账户[%s]出错：%v\n", a.AcountID, err)
							continue
						} else {
							log.Printf("汇总账户[%s]成功，发送地址[%s], TXID：%s\n", a.AcountID, sumAddress, tx.TxID)
						}
					}
				}
			} else {
				log.Printf("钱包[%s]-[%s]当前余额%s，未达到阀值%d\n", w.Name, w.WalletID, w.Balance, threshold)
			}
		}
	}
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

	//查询所有钱包信息
	wallets, err := GetWalletInfo()
	if err != nil {
		log.Printf("客户端没有创建任何钱包\n")
		return err
	}

	log.Printf("序号\tWID\t\t\t\t\t\t\t\t名字\n")
	log.Printf("-----------------------------------------------------------------------------------------\n")

	for i, w := range wallets {
		log.Printf("%d\t%s\t%s\n", i, w.WalletID, w.Name)
	}
	log.Printf("-----------------------------------------------------------------------------------------\n")

	log.Printf("[请选择需要汇总的钱包，输入序号组，以,分隔。例如: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.Stdin.PromptInput("输入需要汇总的钱包序号: ")
	if err != nil {
		openwLogger.Log.Errorf("unexpect error: %v", err)
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
			if numInt < len(wallets)-1 {
				w := wallets[numInt]

				log.Printf("登记汇总钱包[%s]-[%s]\n", w.Name, w.WalletID)
				//输入钱包密码完成登记
				password, err := console.InputPassword(false)
				if err != nil {
					openwLogger.Log.Errorf("unexpect error: %v", err)
					return err
				}

				//配置钱包密码
				w.Password = common.NewString(password).SHA256()

				AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("输入的序号越界")
			}
		} else {
			return errors.New("输入的序号不是数字")
		}
	}

	//启动钱包汇总程序
	SummaryWallets()

	return nil
}

func AddWalletInSummary(wid string, wallet *Wallet) {
	walletsInSum[wid] = wallet
}

//genMnemonic 随机创建密钥
func genMnemonic() string {
	entropy, _ := bip39.NewEntropy(256)
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
	file.MkdirAll(exportBackupDir)
	filepath := exportBackupDir + filename
	file.WriteFile(filepath, []byte(a.Address+"\n"), true)
}

//exportWalletToFile 导出钱包到文件
func exportWalletToFile(w *Wallet) error {

	var (
		err     error
		content []byte
	)

	filename := fmt.Sprintf("wallet-%s-%s.json", w.Name, w.WalletID)

	file.MkdirAll(exportBackupDir)
	filepath := exportBackupDir + filename

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
