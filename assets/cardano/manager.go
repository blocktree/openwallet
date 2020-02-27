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

package cardano

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/console"
	"github.com/bndr/gotabulate"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"github.com/tyler-smith/go-bip39"
	"log"
	"path/filepath"
	"time"
)

const (
	maxAddresNum = 10000000
)

var (
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
)

type WalletManager struct {
	WalletClient *Client            // 节点客户端
	Config       *WalletConfig      //钱包管理配置
	WalletsInSum map[string]*Wallet //参与汇总的钱包
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)

	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*Wallet)

	return &wm
}

//GetWalletInfo 获取钱包信息
//wid 钱包id，可选
func (wm *WalletManager) GetWalletInfo(wid ...string) ([]*Wallet, error) {

	var (
		err     error
		wallets = make([]*Wallet, 0)
	)

	//调用服务
	result := wm.WalletClient.callGetWalletAPI(wid...)
	if err = isError(result); err != nil {
		return nil, err
	}

	content := gjson.GetBytes(result, "data")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			wallets = append(wallets, NewWalletForV1(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		wallets = append(wallets, NewWalletForV1(content))
	}

	return wallets, err
}

//CreateNewWallet 创建新钱包
func (wm *WalletManager) CreateNewWallet(name, mnemonic, password string) error {

	var (
		err error
	)

	//密钥32
	h := common.NewString(password).SHA256()

	//调用服务创建钱包
	result := wm.WalletClient.callCreateWalletAPI(name, mnemonic, h, true)
	if err = isError(result); err != nil {
		return err
	}

	//log.Printf("新钱包助记词：%v", mnemonic)

	content := gjson.GetBytes(result, "data")
	wallet := NewWalletForV1(content)
	wallet.Password = password
	wallet.Mnemonic = mnemonic
	return wm.exportWalletToFile(wallet)
}

//CreateBatchAddress 批量创建地址
func (wm *WalletManager) CreateBatchAddress(wid string, aid int64, password string, count uint) ([]*Address, string, error) {

	var (
		err          error
		done         uint
		producerDone uint
		synCount     uint = 100
	)

	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405") + ".txt"
	filePath := filepath.Join(wm.Config.addressDir, filename)
	//生产通道
	producer := make(chan *Address)

	//消费通道
	worker := wm.createAddressSaveChan(filename)

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
				wm.getAddressWrok(wid, aid, password, producer, err)
			}()

		}
	} else {

		for i := uint(0); i < synCount; i++ {
			go func(runCount uint) {
				for i := uint(0); i < runCount; i++ {
					wm.getAddressWrok(wid, aid, password, producer, err)

				}
			}(runCount)
		}
		//余数不为0，泽直接开启线程运行余下数量
		if otherCount := count % synCount; otherCount != 0 {
			//fmt.Printf("余数为 %d ", otherCount)
			go func(otherCount uint) {
				for i := uint(0); i < otherCount; i++ {
					wm.getAddressWrok(wid, aid, password, producer, err)

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
		case activeWorker <- activeValue:
			values = values[1:]
			done++
			if done == count {
				fmt.Printf("All thread completed!")
				return addresses, filePath, nil
			}

		}
	}

	return addresses, filePath, nil
}

//http获取地址
func (wm *WalletManager) getAddressWrok(wid string, aid int64, passphrase string, producer chan *Address, err error) {
	result := wm.WalletClient.callCreateNewAddressAPI(wid, aid, passphrase)
	if err = isError(result); err != nil {
		log.Print(err)
		return
	}

	content := gjson.GetBytes(result, "data")
	a := NewAddressV1(content)
	fmt.Printf("Create：	%s\n", a.Address)
	producer <- a
}

//保存地址
func (wm *WalletManager) saveAddressWork(address chan *Address, filename string) {

	for a := range address {
		wm.exportAddressToFile(a, filename)
		fmt.Printf("Save:	%s\n", a.Address)
	}
	//return addrs, filename, nil
}

//保存地址通道
func (wm *WalletManager) createAddressSaveChan(filename string) chan<- *Address {
	address := make(chan *Address)
	go wm.saveAddressWork(address, filename)
	return address
}

//CreateNewAccount 根据钱包wid创建单个账户
func (wm *WalletManager) CreateNewAccount(name, wid, passphrase string) (*Account, error) {

	var (
		err error
	)

	//调用服务创建新账户
	result := wm.WalletClient.callCreateNewAccountAPI(name, wid, passphrase)
	if err = isError(result); err != nil {
		log.Print(err)
		return nil, err
	}
	content := gjson.GetBytes(result, "data")
	a := NewAccountV1(content)
	return a, err
}

//GetAccountInfo 获取用户信息
func (wm *WalletManager) GetAccountInfo(wid string, aid ...string) ([]*Account, error) {

	var (
		err      error
		accounts = make([]*Account, 0)
	)

	//调用服务
	result := wm.WalletClient.callGetAccountsAPI(wid, aid...)
	if err = isError(result); err != nil {
		return nil, err
	}

	content := gjson.GetBytes(result, "data")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			accounts = append(accounts, NewAccountV1(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		accounts = append(accounts, NewAccountV1(content))
	}

	return accounts, err
}

//GetAddressInfo 获取指定aid用户的地址组
func (wm *WalletManager) GetAddressInfo(wid, aid string) ([]*Address, error) {

	var (
		err     error
		address = make([]*Address, 0)
	)

	//调用服务
	result := wm.WalletClient.callGetAccountByIDAPI(wid, aid)
	if err = isError(result); err != nil {
		return nil, err
	}

	content := gjson.GetBytes(result, "data.addresses")
	if content.IsArray() {
		//解析如果是数组
		for _, obj := range content.Array() {
			address = append(address, NewAddressV1(obj))
		}
	} else if content.IsObject() {
		//解析如果是单个对象
		address = append(address, NewAddressV1(content))
	}

	return address, err
}

//SendTx 发送交易
func (wm *WalletManager) SendTx(wid string, aid int64, to string, amount uint64, password string) (*Transaction, error) {

	//输入密码
	//password, err := console.InputPassword(false)
	//h := common.NewString(password).SHA256()

	//调用服务创建新账户
	result, err := wm.WalletClient.callSendTxAPI(wid, aid, to, amount, password)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}
	content := gjson.GetBytes(result, "data")
	t := NewTransactionV1(content)
	return t, nil
}

//SummaryTxFlow 执行汇总流程
func (wm *WalletManager) SummaryWallets() {

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range wm.WalletsInSum {

		//统计钱包最新余额
		ws, err := wm.GetWalletInfo(wid)
		if err != nil {
			log.Printf("Can not find wallet information：%v\n", err)
			continue
		}
		if len(ws) > 0 {
			w := ws[0]
			balance, _ := decimal.NewFromString(common.NewString(w.Balance).String())
			//如果余额大于阀值，汇总的地址
			if balance.GreaterThan(wm.Config.Threshold) {
				//汇总所有有钱的账户
				accounts, err := wm.GetAccountInfo(w.WalletID)
				if err != nil {
					log.Printf("Can not find account information：%v\n", err)
					continue
				}

				for _, a := range accounts {
					//大于最小额度才转账
					sendAmount, _ := decimal.NewFromString(common.NewString(a.Amount).String())
					if sendAmount.GreaterThan(wm.Config.MinSendAmount) {
						log.Printf("Summary wallet [%s] - account[%d]  balance = %v \n", w.WalletID, a.Index, sendAmount.Div(coinDecimal))
						log.Printf("Summary wallet [%s] - account[%d]  Start Send Transaction\n", w.WalletID, a.Index)
						tx, err := wm.SendTx(wid, a.Index, wm.Config.SumAddress, uint64(sendAmount.Sub(wm.Config.MinFees).IntPart()), wallet.Password)
						if err != nil {
							log.Printf("Summary wallet [%s] - account[%d]   unexpected error: %v\n", w.WalletID, a.Index, err)
							continue
						} else {
							log.Printf("Summary wallet [%s] - account[%d]   successfully，Received Address[%s], TXID：%s\n", w.WalletID, a.Index, wm.Config.SumAddress, tx.TxID)
						}
					}
				}
			} else {
				log.Printf("Wallet Account[%s]-[%s]  Current Balance: %v，below threshold: %v\n", w.Name, w.WalletID, balance.Div(coinDecimal), wm.Config.Threshold.Div(coinDecimal))
			}
		}
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

func (wm *WalletManager) AddWalletInSummary(wid string, wallet *Wallet) {
	wm.WalletsInSum[wid] = wallet
}

//CreateAddress 给指定账户创建地址
func (wm *WalletManager) CreateAddress(wid string, aid int64, passphrase string) (*Address, error) {
	result := wm.WalletClient.callCreateNewAddressAPI(wid, aid, passphrase)
	err := isError(result)
	if err != nil {
		log.Printf("Create address failed! ")
		return nil, err
	}
	content := gjson.GetBytes(result, "data")
	a := NewAddressV1(content)
	return a, nil
}

//EstimateFees 计算预估手续费
func (wm *WalletManager) EstimateFees(wid string, aid int64, to string, amount uint64, passphrase string) (uint64, error) {

	result, _ := wm.WalletClient.callEstimateFeesAPI(wid, aid, to, amount, passphrase)
	err := isError(result)
	if err != nil {
		return 0, nil
	}

	fees := gjson.GetBytes(result, "data.estimatedAmount")

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
		//error 返回错误
		{
			"status": "error",
			"diagnostic": {},
			"message": ""
		}
	*/

	//V0的错误信息存放在Left上
	if gjson.GetBytes(result, "status").String() == "success" {
		return nil
	}

	err = errors.New(gjson.GetBytes(result, "diagnostic").String())

	return err
}

//exportAddressToFile 导出地址到文件中
func (wm *WalletManager) exportAddressToFile(a *Address, filename string) {
	file.MkdirAll(wm.Config.addressDir)
	filepath := filepath.Join(wm.Config.addressDir, filename)
	file.WriteFile(filepath, []byte(a.Address+"\n"), true)
}

//exportWalletToFile 导出钱包到文件
func (wm *WalletManager) exportWalletToFile(w *Wallet) error {

	var (
		err     error
		content []byte
	)

	filename := fmt.Sprintf("wallet-%s-%s.json", w.Name, w.WalletID)

	file.MkdirAll(wm.Config.keyDir)
	filepath := filepath.Join(wm.Config.keyDir, filename)

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
			log.Printf("unexpected error: %v\n", err)
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
			log.Printf("unexpected error: %v\n", err)
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
func (wm *WalletManager) LoadConfig() error {
	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	wm.Config.ServerAPI = c.String("apiUrl")
	wm.Config.Threshold = decimal.RequireFromString(c.String("threshold")).Mul(coinDecimal)
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.MinFees = decimal.RequireFromString(c.String("minFees")).Mul(coinDecimal)
	wm.Config.MinSendAmount = decimal.RequireFromString(c.String("minSendAmount")).Mul(coinDecimal)
	wm.Config.WalletDataPath = c.String("walletDataPath")
	if wm.Config.WalletDataPath == "" {
		return errors.New("walletDataPath is for backup wallet, so set it")
	}

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}

	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)

	wm.WalletClient = NewClient(wm.Config.ServerAPI, false)

	return nil
}

//打印钱包列表
func (wm *WalletManager) printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		fmt.Print(w.WalletID)
		balance, err := decimal.NewFromString(w.Balance)
		if err != nil {
			continue
		}
		balance = balance.Div(coinDecimal)
		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Name, balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "WID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}
