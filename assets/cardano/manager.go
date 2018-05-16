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
)

const (
	exportBackupDir = "./data/"
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

//InputNumber 输入地址数量
func InputNumber() int {

	var (
		count = 0 // 输入的创建数量
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("输入需要创建的地址数量: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return 0
		}
		count = common.NewString(line).Int()
		if count < 1 {
			log.Printf("输入地址数量必须大于0")
			continue
		}
		break
	}

	return count
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
func CreateBatchAddress(aid, passphrase string, count uint) ([]*Address, error) {

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
				getAddressWrok(aid,passphrase,producer,err)
			}()
		}
	}else{
		for i := uint(0); i < synCount; i++ {

			go func(runCount uint) {
				for i := uint(0); i < runCount; i++ {
						getAddressWrok(aid,passphrase,producer,err)

				}
			}(runCount)
		}
		//余数不为0，泽直接开启线程运行余下数量
		if otherCount := count%synCount;otherCount!=0{
			go func(otherCount uint) {
				for i := uint(0); i < otherCount; i++ {
						getAddressWrok(aid,passphrase,producer,err)

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
				return addresses, nil
			}

		}
	}
	return addresses, nil
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
}

//保存地址通道
func createAddressSaveChan (count uint,filename string)chan<- *Address{
	address := make(chan *Address)
	go saveAddressWork(address,count,filename)
	return address
}


//CreateNewAccount 根据钱包wid创建单个账户
func CreateNewAccount(name, wid, passphrase string) error {

	var (
		err error
	)

	//调用服务创建新账户
	result := callCreateNewAccountAPI(name, wid, passphrase)
	err = isError(result)

	return err
}

//GetAccountInfo 获取用户信息
func GetAccountInfo(aid ...string) ([]*Account, error) {

	var (
		err      error
		accounts = make([]*Account, 0)
	)

	//调用服务
	result := callGetAccounts(aid...)
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

//CreateAddressFlow
func CreateAddressFlow() error {

	var (
		err error
	)

	count := InputNumber()
	if count < 1 {
		err = errors.New("输入地址数量必须大于0")
		return err
	}
	//获取钱包所有账户
	_, err = GetWalletInfo()
	if err != nil {
		return err
	}

	return nil
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

	log.Printf("钱包创建成功，导出路径:%s\n",filepath)

	return nil
}

func WriteSomething() {
	content := "Hello, openwallet\n"
	filename := exportBackupDir + "testfile.txt"
	file.MkdirAll(exportBackupDir)
	file.WriteFile(filename, []byte(content), true)
	file.WriteFile(filename, []byte(content), true)
	file.WriteFile(filename, []byte(content), true)
}
