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

package bytom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/bndr/gotabulate"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"path/filepath"
	"time"
)

const (
	maxAddresNum = 1000000
)

var (
	//钱包服务API
	serverAPI = "http://127.0.0.1:10031"
	//钱包主链私钥文件路径
	walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
	//参与汇总的钱包
	walletsInSum = make(map[string]*AccountBalance)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(12).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(10).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(0.005).Mul(coinDecimal)
	//汇总地址
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
	// 节点客户端
	client *Client
)

//CreateNewWallet 创建钱包
func CreateNewWallet(alias, password string) (*Wallet, error) {

	request := struct {
		Alias    string `json:"alias"`
		Password string `json:"password"`
	}{alias, password}

	result, err := client.Call("create-key", request)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	w := NewWallet(gjson.GetBytes(result, "data"))

	return w, err

}

//GetWalletInfo 获取钱包信息
func GetWalletInfo() ([]*Wallet, error) {

	var (
		wallets = make([]*Wallet, 0)
	)

	result, err := client.Call("list-keys", nil)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	array := gjson.GetBytes(result, "data").Array()
	for _, a := range array {
		wallets = append(wallets, NewWallet(a))
	}

	return wallets, err

}

//CreateNormalAccount 创建一个单签普通账户
func CreateNormalAccount(xpub, alias string) (*Account, error) {

	/*
		{
			"root_xpubs": ["2d6c07cb1ff7800b0793e300cd62b6ec5c0943d308799427615be451ef09c0304bee5dd492c6b13aaa854d303dc4f1dcb229f9578786e19c52d860803efa3b9a"],
			"quorum": 1,
			"alias": "alice"
		}
	*/

	request := struct {
		RootXpubs []string `json:"root_xpubs"`
		Quorum    int      `json:"quorum"`
		Alias     string   `json:"alias"`
	}{[]string{xpub}, 1, alias}

	result, err := client.Call("create-account", request)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	a := NewAccount(gjson.GetBytes(result, "data"))

	return a, err

}

//GetAccountInfo 获取账户信息
func GetAccountInfo() ([]*Account, error) {

	var (
		accounts = make([]*Account, 0)
	)

	result, err := client.Call("list-accounts", nil)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	array := gjson.GetBytes(result, "data").Array()
	for _, a := range array {
		accounts = append(accounts, NewAccount(a))
	}

	return accounts, err

}

//GetAccountBalance 获取账户资产
func GetAccountBalance(accountID string, assetsID string) ([]*AccountBalance, error) {

	var (
		accounts = make([]*AccountBalance, 0)
	)

	result, err := client.Call("list-balances", nil)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	array := gjson.GetBytes(result, "data").Array()
	for _, a := range array {

		account := NewAccountBalance(a)
		if len(assetsID) > 0 {
			if account.AssetID != assetsID {
				continue
			}
		}

		if len(accountID) > 0 {
			if account.AccountID != accountID {
				continue
			}
		}

		accounts = append(accounts, account)
	}

	return accounts, err

}

//CreateReceiverAddress 给指定账户创建地址
func CreateReceiverAddress(alias, accountID string) (*Address, error) {

	/*
		{
			"account_alias": "alice",
			"account_id": "0BDQARM800A02"
		}
	*/

	request := struct {
		Account_alias string `json:"account_alias"`
		Account_id    string `json:"account_id"`
	}{alias, accountID}

	result, err := client.Call("create-account-receiver", request)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	a := NewAddress(accountID, alias, gjson.GetBytes(result, "data"))

	return a, err

}

//CreateBatchAddress 批量创建地址
func CreateBatchAddress(alias, accountID string, count uint64) (string, error) {

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
			address, errRun := CreateReceiverAddress(alias, accountID)
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

//GetAddressInfo 获取指定账户的所有地址
func GetAddressInfo(alias, accountID string) ([]*Address, error) {

	var (
		addresses = make([]*Address, 0)
	)

	request := struct {
		Account_alias string `json:"account_alias"`
		Account_id    string `json:"account_id"`
	}{alias, accountID}

	result, err := client.Call("list-addresses", request)
	if err != nil {
		return nil, err
	}

	err = isError(result)
	if err != nil {
		return nil, err
	}

	array := gjson.GetBytes(result, "data").Array()
	for _, a := range array {
		addresses = append(addresses, NewAddress(accountID, alias, a))
	}

	return addresses, err

}

//SummaryWallets 执行汇总流程
func SummaryWallets() {

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range walletsInSum {

		//统计钱包最新余额
		ws, err := GetAccountBalance(wid, assetsID_btm)
		if err != nil {
			log.Printf("Can not find Account Balance：%v\n", err)
			continue
		}
		if len(ws) > 0 {
			w := ws[0]
			balance, _ := decimal.NewFromString(common.NewString(w.Amount).String())
			//如果余额大于阀值，汇总的地址
			if balance.GreaterThan(threshold) {

				log.Printf("Summary account[%s]balance = %v \n", w.AccountID, balance.Div(coinDecimal))
				log.Printf("Summary account[%s]Start Send Transaction\n", w.AccountID)

				//避免临界值的错误，减去1个
				//balance = balance.Sub(coinDecimal)

				txID, err := SendTransaction(w.AccountID, sumAddress, assetsID_btm, uint64(balance.IntPart()), wallet.Password, false)
				if err != nil {
					log.Printf("Summary account[%s]unexpected error: %v\n", w.AccountID, err)
					continue
				} else {
					log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", w.AccountID, sumAddress, txID)
				}
			} else {
				log.Printf("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v\n", w.Alias, w.AccountID, balance.Div(coinDecimal), threshold.Div(coinDecimal))
			}
		} else {
			log.Printf("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v\n", wallet.Alias, wallet.AccountID, 0, threshold.Div(coinDecimal))
		}
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

//AddWalletInSummary 添加汇总钱包账户
func AddWalletInSummary(wid string, wallet *AccountBalance) {
	walletsInSum[wid] = wallet
}

//CreateReceiverAddress 给指定账户创建地址
func BuildTransaction(from, to, assetsID string, amount, fees uint64) (string, error) {

	/*
		{
			"base_transaction": null,
			"actions": [{
				"account_id": "0BF63M2U00A04",
				"amount": 20000000,
				"asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				"type": "spend_account"
			}, {
				"account_id": "0BF63M2U00A04",
				"amount": 99,
				"asset_id": "3152a15da72be51b330e1c0f8e1c0db669269809da4f16443ff266e07cc43680",
				"type": "spend_account"
			}, {
				"amount": 99,
				"asset_id": "3152a15da72be51b330e1c0f8e1c0db669269809da4f16443ff266e07cc43680",
				"receiver": {
					"control_address": "0014a3f9111f3b0ee96cbd119a3ea5c60058f506fb19"
				},
				"type": "control_address"
			}],
			"ttl": 0,
			"time_range": 1521625823
		}
	*/

	request := struct {
		BaseTransaction interface{}              `json:"base_transaction"`
		Actions         []map[string]interface{} `json:"actions"`
		TTL             uint64                   `json:"ttl"`
		TimeRange       uint64                   `json:"time_range"`
	}{nil,
		[]map[string]interface{}{
			map[string]interface{}{
				"account_id": from,
				"amount":     amount + fees,
				"asset_id":   assetsID,
				"type":       "spend_account",
			},
			map[string]interface{}{
				"amount":   amount,
				"asset_id": assetsID,
				"address":  to,
				//"receiver": map[string]interface{}{
				//	"control_address": to,
				//},
				"type": "control_address",
			},
		}, 1, 0}

	result, err := client.Call("build-transaction", request)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	tx := gjson.GetBytes(result, "data").Raw

	return tx, err

}

//SignTransaction 签名交易
func SignTransaction(txForm, password string) (string, error) {

	tx := gjson.Parse(txForm).Value().(map[string]interface{})

	request := struct {
		Transaction map[string]interface{} `json:"transaction"`
		Password    string                 `json:"password"`
	}{tx, password}

	result, err := client.Call("sign-transaction", request)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	singedTx := gjson.GetBytes(result, "data").Raw

	return singedTx, err

}

//SubmitTransaction 提交新交易单
func SubmitTransaction(txRaw string) (string, error) {

	request := struct {
		RawTransaction string `json:"raw_transaction"`
	}{txRaw}

	result, err := client.Call("submit-transaction", request)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	txID := gjson.GetBytes(result, "data.tx_id").String()

	return txID, err

}

//GetTransactions 获取交易列表
func GetTransactions(accountID string) (string, error) {

	request := struct {
		Account_id string `json:"account_id"`
		Detail     bool   `json:"detail"`
	}{accountID, true}

	result, err := client.Call("list-transactions", request)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	content := gjson.GetBytes(result, "data")

	return content.Raw, nil

}

//SendTransaction 发送交易
func SendTransaction(accountID, to, assetsID string, amount uint64, password string, feesInSender bool) (string, error) {

	//建立交易单
	tx, err := BuildTransaction(accountID, to, assetsID, amount, 0)
	if err != nil {
		return "", err
	}

	totalFees, err := EstimateTransactionGas(tx)

	if !feesInSender {
		amount = amount - totalFees
	}

	//添加手续重新建立交易单
	txAddFees, err := BuildTransaction(accountID, to, assetsID, amount, totalFees)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	fmt.Printf("Build Transaction Successfully\n")

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("From AccountID: %s\n", accountID)
	fmt.Printf("To Address: %s\n", to)
	fmt.Printf("Send: %v\n", decimal.New(int64(amount+totalFees), 0).Div(coinDecimal))
	fmt.Printf("Fees: %v\n", decimal.New(int64(totalFees), 0).Div(coinDecimal))
	fmt.Printf("Receive: %v\n", decimal.New(int64(amount), 0).Div(coinDecimal))
	fmt.Printf("-----------------------------------------------\n")

	//签名交易单
	signTx, err := SignTransaction(txAddFees, password)
	if err != nil {
		return "", err
	}

	fmt.Printf("Sign Transaction Successfully\n")

	//广播交易单
	txRaw := gjson.Get(signTx, "transaction.raw_transaction").String()
	txID, err := SubmitTransaction(txRaw)
	if err != nil {
		return "", err
	}

	fmt.Printf("Submit Transaction Successfully\n")

	return txID, nil
}

//GetTransactions 获取交易列表
func EstimateTransactionGas(txForm string) (uint64, error) {

	tx := gjson.Parse(txForm).Value().(map[string]interface{})

	request := struct {
		Transaction map[string]interface{} `json:"transaction_template"`
	}{tx}

	result, err := client.Call("estimate-transaction-gas", request)
	if err != nil {
		return 0, err
	}

	err = isError(result)
	if err != nil {
		return 0, err
	}

	totalNeu := gjson.GetBytes(result, "data.total_neu").Uint()

	return totalNeu, nil

}

//BackupWallet 备份钱包私钥数据
func BackupWallet() (string, error) {

	result, err := client.Call("backup-wallet", nil)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	content := gjson.GetBytes(result, "data")

	var buf bytes.Buffer
	err = json.Indent(&buf, []byte(content.Raw), "", "\t")
	if err != nil {
		return "", err
	}

	return exportKeystoreToFile(buf.Bytes())
}

//RestoreWallet 通过keystore恢复钱包
func RestoreWallet(keystore []byte) error {

	request, ok := gjson.ParseBytes(keystore).Value().(map[string]interface{})
	if !ok {
		return errors.New("Can not parse keystore file! ")
	}

	result, err := client.Call("restore-wallet", request)
	if err != nil {
		return err
	}

	err = isError(result)
	if err != nil {
		return err
	}

	return nil
}

//GetWalletList 获取钱包资产信息列表
func GetWalletList(assetsID string) ([]*AccountBalance, error) {

	accounts, err := GetAccountInfo()

	balances := make([]*AccountBalance, 0)
	accMap := make(map[string]*AccountBalance)

	//收集资产信息
	for _, a := range accounts {
		ab := &AccountBalance{}
		ab.AccountID = a.ID
		ab.Alias = a.Alias

		balances = append(balances, ab)
		accMap[a.ID] = ab
	}

	//查询钱包资产
	list, err := GetAccountBalance("", assetsID)
	if err != nil {
		return nil, err
	}

	//合并资产
	for _, a := range list {
		ac := accMap[a.AccountID]
		ac.AssetAlias = a.AssetAlias
		ac.AssetID = a.AssetID
		ac.Amount = a.Amount
	}

	return balances, nil
}

//SignMessage 消息签名
func SignMessage(address, message, password string) (string, error) {

	request := struct {
		Address  string `json:"address"`
		Message  string `json:"message"`
		Password string `json:"password"`
	}{address, message, password}

	result, err := client.Call("sign-message", request)
	if err != nil {
		return "", err
	}

	err = isError(result)
	if err != nil {
		return "", err
	}

	signature := gjson.GetBytes(result, "data.signature").String()

	return signature, nil
}

//exportWalletToFile 导出钱包到文件
func exportKeystoreToFile(content []byte) (string, error) {

	filename := fmt.Sprintf("wallet-%s.json", common.TimeFormat("20060102150405"))

	file.MkdirAll(keyDir)
	filePath := filepath.Join(keyDir, filename)

	//把钱包写入到文件进行备份
	if !file.WriteFile(filePath, content, true) {
		return "", errors.New("Keystore write to file failed! ")
	}

	return filePath, nil
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

//打印钱包列表
func printWalletList(list []*AccountBalance) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		balance := decimal.New(int64(w.Amount), 0)
		balance = balance.Div(coinDecimal)
		tableInfo = append(tableInfo, []interface{}{
			i, w.AccountID, w.Alias, balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}
