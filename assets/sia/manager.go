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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NebulousLabs/entropy-mnemonics"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/codeskyblue/go-sh"
	"github.com/imroc/req"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/bndr/gotabulate"
)

var (
	//钱包服务API
	serverAPI = "http://127.0.0.1:10031"
	//授权密码
	//Auth = ""
	////备份文件地址
	//restorePath =""
	//钱包主链私钥文件路径
	//walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.New(1, 24)
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
	//秘钥存取
	storage *keystore.HDKeystore
)

func init() {
	storage = keystore.NewHDKeystore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

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
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	//restorePath = c.String("restorePath")
	rpcPassword = c.String("rpcPassword")
	walletDataPath = c.String("walletDataPath")

	//walletPath = c.String("walletPath")
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
		Auth:    rpcPassword,
	}

	return nil
}

//BackupWallet 备份钱包私钥数据
func BackupWallet() (string, error) {

	//创建备份文件夹
	newBackupDir := filepath.Join(backupDir, "wallet-backup"+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	//复制临时文件到备份文件夹
	file.Copy(filepath.Join(walletDataPath, "wallet.db"), newBackupDir)

	return newBackupDir, nil
}

//RestoreWallet 通过keystore恢复钱包
func RestoreWallet(dbFile string) error {

	//钱包当前的db文件
	currentWDFile := filepath.Join(walletDataPath, "wallet.db")

	//创建临时备份文件夹
	tmpWalletData := filepath.Join(backupDir, "restore-wallet-backup"+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(tmpWalletData)

	//备份
	file.Copy(currentWDFile, tmpWalletData)

	fmt.Printf("Restore wallet.db file... \n")

	//删除当前钱包文件
	err2 := file.Delete(currentWDFile)
	if !err2 {
		log.Printf("Restore wallet unsuccessfully...please copy the backup file to wallet data path manually. \n")
	} else {
		//恢复备份dat到钱包数据目录
		err := file.Copy(dbFile, walletDataPath)

		if err != nil {

			fmt.Printf("Restore wallet unsuccessfully...please copy the backup file to wallet data path manually.  \n")
			//删除当前钱包文件
			//file.Delete(currentWDFile)
			fmt.Printf("Restore original wallet.data... \n")
			tmpData := filepath.Join(tmpWalletData,"wallet.db")
			file.Copy(tmpData, walletDataPath)
			return err

		} else {
			//删除临时备份的dat文件
			file.Delete(tmpWalletData)
		}
	}
	return nil
}

//startNode 开启节点
func startNode() error {

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup! ")
	}

	startNodeCMD := c.String("startNodeCMD")
	return cmdCall(startNodeCMD, false)
}

//stopNode 关闭节点
func stopNode() error {
	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup! ")
	}

	stopNodeCMD := c.String("stopNodeCMD")
	return cmdCall(stopNodeCMD, true)
}

//cmdCall 执行命令
func cmdCall(cmd string, wait bool) error {

	var (
		cmdName string
		args    []string
	)

	cmds := strings.Split(cmd, " ")
	if len(cmds) > 0 {
		cmdName = cmds[0]
		args = cmds[1:]
	} else {
		return errors.New("command not found ")
	}
	session := sh.Command(cmdName, args)
	if wait {
		return session.Run()
	} else {
		return session.Start()
	}
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

	//检查钱包名是否存在
	//wallets, err := GetWalletKeys(keyDir)
	//for _, w := range wallets {
	//	if w.Alias == name {
	//		return "", errors.New("The wallet's alias is duplicated! ")
	//	}
	//}

	request := req.Param{
		"encryptionpassword": password,
		"force":              force,
	}

	result, err := client.Call("wallet/init", "POST", request)
	//if err != nil {
	//	return "", err
	//}

	//primaryseed := gjson.GetBytes(result, "primaryseed").String()
	//
	//storeNewKey(name, primaryseed, password)

	return string(result), err

}

func CreateNewWalletKey(alias, password string) (string, error) {

	//检查钱包名是否存在
	wallets, err := GetWalletKeys(keyDir)
	for _, w := range wallets {
		if w.Alias == alias {
			return "", errors.New("The wallet's alias is duplicated! ")
		}
	}

	fmt.Printf("Create new wallet keystore...\n")

	_, keyFile, err := keystore.StoreHDKey(keyDir, alias, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return "", err
	}

	return keyFile, nil

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
	filename := "addresses-" + common.TimeFormat("20060102150405") + ".txt"
	filePath := filepath.Join(addressDir, filename)

	//生产通道
	producer := make(chan []string)
	defer close(producer)

	//消费通道
	worker := make(chan []string)
	defer close(worker)

	//创建地址过程
	createAddressWork := func(runCount uint64, prod chan []string) {

		runAddress := make([]string, 0)

		for i := uint64(0); i < runCount; i++ {
			// 请求地址
			address, errRun := CreateReceiverAddress()
			if errRun != nil {
				continue
			}
			onlyAddress := gjson.GetBytes(address,"address").String()
			runAddress = append(runAddress, onlyAddress)

		}
		//生成完成
		prod <- runAddress
	}

	//保存地址过程
	saveAddressWork := func(addresses chan []string, filename string) {
		//fmt.Println(addresses)
		for {
			//回收创建的地址
			getAddrs := <-addresses
			//log.Printf("Export %d", len(getAddrs))
			//导出一批地址
			exportAddressToFile(getAddrs, filePath)

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

			go createAddressWork(runCount, producer)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		go createAddressWork(otherCount, producer)

		shouldDone++
	}

	values := make([][]string, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- []string
		var activeValue []string

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
func CreateReceiverAddress() ([]byte, error) {

	result, err := client.Call("wallet/address", "GET", nil)

	return result, err

}

//exportAddressToFile 导出地址到文件中
func exportAddressToFile(addrs []string, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		log.Printf("Export: %s \n", a)

		content = content + a + "\n"
	}

	file.MkdirAll(addressDir)
	file.WriteFile(filePath, []byte(content), true)
}

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

//GetWalletKeys 通过给定的文件路径加载keystore文件得到钱包列表
func GetWalletKeys(dir string) ([]*Wallet, error) {

	var (
		buf = new(bufio.Reader)
		key struct {
			Alias  string `json:"alias"`
			RootId string `json:"rootid"`
		}

		wallets = make([]*Wallet, 0)
	)

	//加载文件，实例化钱包
	readWallet := func(path string) *Wallet {

		fd, err := os.Open(path)
		defer fd.Close()
		if err != nil {
			return nil
		}

		buf.Reset(fd)
		// Parse the address.
		key.Alias = ""
		key.RootId = ""
		err = json.NewDecoder(buf).Decode(&key)
		if err != nil {
			return nil
		}

		return &Wallet{WalletID: key.RootId, Alias: key.Alias}
	}

	//扫描key目录的所有钱包
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return wallets, err
	}

	for _, fi := range files {
		// Skip any non-key files from the folder
		if skipKeyFile(fi) {
			continue
		}
		if fi.IsDir() {
			continue
		}

		path := filepath.Join(keyDir, fi.Name())

		w := readWallet(path)
		w.KeyFile = fi.Name()
		wallets = append(wallets, w)

	}

	return wallets, nil

}

//SendTransaction 发送交易
func SendTransaction(amount string, destination string) (string, error) {

	request := req.Param{
		"amount":      amount,
		"destination": destination,
	}

	txIDs, err := client.Call("wallet/siacoins", "POST", request)
	if err != nil {
		return "", err
	}

	fmt.Printf("Send Transaction Successfully\n")

	txID := gjson.GetBytes(txIDs,"transactionids").String()
	return txID, nil
}

//SummaryWallets 执行汇总流程
func SummaryWallets() {

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//统计钱包最新余额
	ws, err := client.Call("wallet", "GET", nil)
	if err != nil {
		log.Printf("Can not find Account Balance：%v\n", err)
	}
	var (
		wallets = make([]*Wallet, 0)
	)
	a := gjson.ParseBytes(ws)
	wallets = append(wallets, NewWallet(a))

	if len(wallets) > 0 {
		//balance, _ := decimal.NewFromString(common.NewString(wallets[0].ConfirmBalance).String())
		balance, _ := decimal.NewFromString(wallets[0].ConfirmBalance)

		//如果余额大于阀值，汇总的地址
		if balance.GreaterThan(threshold) {
			//如果钱包正在执行转账，需要等待转账完毕才能继续进行汇总
			if wallets[0].OutgoingSC == "0"{
				log.Printf("Summary balance = %v \n", balance.Div(coinDecimal))
				log.Printf("Summary Start Send Transaction\n")

				//避免临界值的错误，减去1个

				//balance = balance.Sub(coinDecimal)

				//txID, err := SendTransaction(w.AccountID, sumAddress, assetsID_btm, uint64(balance.IntPart()), wallet.Password, false)
				txID, err := SendTransaction(balance.Sub(minFees).String(), sumAddress)
				if err != nil {
					log.Printf("Summary unexpected error: %v\n", err)
				} else {
					log.Printf("Summary successfully，Received Address[%s], Transaction ID:%s", sumAddress,txID)
				}
			}else {
				log.Printf("wallet has coins spent in incomplete transaction, summaryWallets will begin after it finish...")
			}
		} else {
			log.Printf("Wallet  Balance: %v，below threshold: %v\n", balance.Div(coinDecimal), threshold.Div(coinDecimal))
		}
	} else {
		log.Printf("Wallet Current Balance: %v，below threshold: %v\n", 0, threshold.Div(coinDecimal))
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

//exportWalletToFile 导出钱包到文件
func exportKeystoreToFile(content []byte) error {

	filename := fmt.Sprintf("wallet-%s.json", common.TimeFormat("20060102150405"))

	file.MkdirAll(keyDir)
	filePath := filepath.Join(keyDir, filename)

	//把钱包写入到文件进行备份
	if !file.WriteFile(filePath, content, true) {
		return errors.New("Keystore write to file failed! ")
	}

	return nil
}

// skipKeyFile ignores editor backups, hidden files and folders/symlinks.
func skipKeyFile(fi os.FileInfo) bool {
	// Skip editor backups and UNIX-style hidden files.
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	// Skip misc special files, directories (yes, symlinks too).
	if fi.IsDir() || fi.Mode()&os.ModeType != 0 {
		return true
	}
	return false
}

//storeNewKey 把钱包的助记词以hdkeystore形式保存
func storeNewKey(wallet *Wallet, words, auth string) (*keystore.HDKey, string, error) {

	seed, err := mnemonics.FromString(words, mnemonics.English)
	if err != nil {
		return nil, "", err
	}

	key, err := keystore.NewHDKey(seed, wallet.Alias, "")
	if err != nil {
		return nil, "", err
	}
	wallet.WalletID = key.RootId
	filePath := storage.JoinPath(wallet.FileName() + ".key")
	storage.StoreKey(filePath, key, auth)

	wallet.KeyFile = filePath

	return key, filePath, err
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		balance, _ := decimal.NewFromString(w.ConfirmBalance)
		balance = balance.Div(coinDecimal)
		tableInfo = append(tableInfo, []interface{}{
			i, w.Rescanning, w.Unlocked, balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "Rescanning", "Unlocked", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}