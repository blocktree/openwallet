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

package sia

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"time"
	"github.com/imroc/req"
	"github.com/tyler-smith/go-bip39"
	"github.com/astaxie/beego/config"
	"path/filepath"
	"errors"
	"github.com/blocktree/OpenWallet/common"
	"log"
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
)

var (
	//钱包服务API
	serverAPI = "http://127.0.0.1:10031"
	//钱包主链私钥文件路径
	walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000000000000000000000)
	//参与汇总的钱包
	//walletsInSum = make(map[string]*AccountBalance)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(12).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(10).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(22600000000000000000000)
	//汇总地址
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
	// 节点客户端
	client *Client
)

//GetWalletInfo 获取钱包信息
func GetWalletInfo() ([]*Wallet, error) {

	var (
		wallets = make([]*Wallet, 0)
	)

	result, err := client.Call("wallet", "GET", nil)
	if err != nil {
		return nil, err
	}

	a := gjson.ParseBytes(result)
	wallets = append(wallets, NewWallet(a))

	return wallets, err

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
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	walletPath = c.String("walletPath")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	threshold = threshold.Mul(coinDecimal)
	//minSendAmount, _ = decimal.NewFromString(c.String("minSendAmount"))
	//minSendAmount = minSendAmount.Mul(coinDecimal)
	//minFees, _ = decimal.NewFromString(c.String("minFees"))
	//minFees = minFees.Mul(coinDecimal)
	sumAddress = c.String("sumAddress")

	client = &Client{
		BaseURL: serverAPI,
		Debug:   false,
	}

	return nil
}


//BackupWallet 备份钱包私钥数据
func BackupWallet(destination string) (string, error) {

	request := req.Param{
		"destination": destination,
	}

	_, err := client.Call("wallet/backup", "GET", request)
	if err != nil {
		return "", err
	}

	return destination, nil
}

//RestoreWallet 通过keystore恢复钱包
func RestoreWallet(keystore []byte) error {


	return nil
}

//UnlockWallet 解锁钱包
func UnlockWallet(password string) error {

	request := req.Param{
		"encryptionpassword": password,
	}

	_, err := client.Call("wallet/unlock", "POST", request)
	if err != nil {
		return err
	}

	return nil
}

//CreateNewWallet 创建钱包
func CreateNewWallet(password string, force bool) (string, error) {

	request := req.Param{
		"encryptionpassword": password,
		"force": force,
	}

	result, err := client.Call("wallet/init", "POST", request)
	if err != nil {
		return "", err
	}

	primaryseed := gjson.GetBytes(result, "seed").String()

	return primaryseed, err

}

//CreateAddress 创建钱包地址(慎用)
func CreateAddress() (string, error) {

	result, err := client.Call("wallet/address", "GET", nil)
	if err != nil {
		return "", err
	}

	address := gjson.GetBytes(result, "address").String()

	return address, err

}
//CreateBatchAddress 批量创建钱包地址
//func CreateBatchAddress(alias, accountID string, count uint64) (string, error) {
func CreateBatchAddress(count uint64) (string, error) {

	var (
		synCount   uint64 = 100
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405") + ".txt"
	filePath := filepath.Join(addressDir, filename)

	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*Address)
	defer close(worker)

	//创建地址过程
	createAddressWork := func(runCount uint64) {

		runAddress := make([]*Address, 0)

		for i := uint64(0); i < runCount; i++ {
			// 请求地址
			address, errRun := CreateReceiverAddress()
			if errRun != nil {
				continue
			}
			runAddress = append(runAddress, address)

		}
		//生成完成
		producer <- runAddress
	}

	//保存地址过程
	saveAddressWork := func(addresses chan []*Address, filename string) {

		for {
			//回收创建的地址
			getAddrs := <-addresses
			//log.Printf("Export %d", len(getAddrs))
			//导出一批地址
			exportAddressToFile(getAddrs, filename)

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
			log.Printf("Start create address thread[%d]\n", i)

			go createAddressWork(runCount)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		go createAddressWork(otherCount)

		shouldDone++
	}

	values := make([][]*Address, 0)

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
			//log.Printf("completed %d", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Printf("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			log.Printf("All addresses have been created!")
			return filePath, nil
		}
	}

	return filePath, nil
}

//CreateReceiverAddress 给指定账户创建地址
func CreateReceiverAddress() (*Address, error) {

	result, err := client.CallBatchAddress("wallet/address", "GET", nil)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	a := NewAddress(gjson.GetBytes(result, "data"))

	return a, err

}

//isError 是否报错
func isError(result []byte) error {

	var (
		err error
	)

	if gjson.GetBytes(result, "status").String() == "success" {
		return nil
	}

	errInfo := fmt.Sprintf("[%s]%s",
		gjson.GetBytes(result, "status").String(),
		gjson.GetBytes(result, "msg").String())
	err = errors.New(errInfo)

	return err
}

//exportAddressToFile 导出地址到文件中
func exportAddressToFile(addrs []*Address, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		log.Printf("Export: %s \n", a.Address)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(addressDir)
	file.WriteFile(filePath, []byte(content), true)
}


//CreateBatchAddress 批量创建钱包地址
//func CreateBatchAddress(count uint64) ([]byte, error) {
//
//	l := list.New()
//	var i uint64
//	for i=0;i < count ;i++  {
//		filePath, _ := client.Call("wallet/address", "GET", nil)
//		l.PushBack(filePath)
//	}
//	addresses, err := json.Marshal(l)
//	return addresses, err
//}

//GetAddressInfo 获取地址列表
func GetAddressInfo() ([]string, error) {

	result, err := client.Call("wallet/addresses", "GET", nil)
	if err != nil {
		return nil, err
	}

	content := gjson.GetBytes(result, "addresses").Array()

	addresses := make([]string, 0)
	for _, a := range content {
		addresses = append(addresses, a.String())
	}

	return addresses, err

}

func GetConsensus() error {

	_, err := client.Call("consensus", "GET", nil)
	if err != nil {
		return err
	}

	return nil
}

//genMnemonic 随机创建密钥
func genMnemonic() string {
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}

//SendTransaction 发送交易
func SendTransaction(amount string,destination string) (string, error) {

	request := req.Param{
		"amount": amount,
		"destination": destination,
	}

	_, err := client.Call("wallet/siacoins", "POST", request)
	if err != nil {
		return "", err
	}

	fmt.Printf("Send Transaction Successfully\n")

	return "", nil
}

//SummaryWallets 执行汇总流程
func SummaryWallets() error{

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

		//统计钱包最新余额
		ws, err := client.Call("wallet", "GET", nil)
		if err != nil {
			log.Printf("Can not find Account Balance：%v\n", err)
			return err
		}
		var (
			wallets = make([]*Wallet, 0)
		)
		a := gjson.ParseBytes(ws)
		wallets = append(wallets, NewWallet(a))

		if len(wallets) > 0 {
			//balance, _ := decimal.NewFromString(common.NewString(wallets[0].ConfirmBalance).String())
			balance, _ := decimal.NewFromString(common.NewString(wallets[0].ConfirmBalance).String())

			//如果余额大于阀值，汇总的地址
			if balance.GreaterThan(threshold) {

				log.Printf("Summary balance = %v \n", balance.Div(coinDecimal))
				log.Printf("Summary Start Send Transaction\n")

				//避免临界值的错误，减去1个

				//balance = balance.Sub(coinDecimal)

				//txID, err := SendTransaction(w.AccountID, sumAddress, assetsID_btm, uint64(balance.IntPart()), wallet.Password, false)
				_, err = SendTransaction(balance.Sub(minFees).String(),sumAddress)
				if err != nil {
					log.Printf("Summary unexpected error: %v\n", err)
					return err
				} else {
					log.Printf("Summary successfully，Received Address[%s]",  sumAddress)
				}
			} else {
				log.Printf("Wallet  Balance: %v，below threshold: %v\n",  balance.Div(coinDecimal), threshold.Div(coinDecimal))
			}
		} else {
			log.Printf("Wallet Current Balance: %v，below threshold: %v\n",  0, threshold.Div(coinDecimal))
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
	return nil
}
