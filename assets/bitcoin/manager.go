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

package bitcoin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/bndr/gotabulate"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxAddresNum = 10000
)

var (
	//秘钥存取
	storage *keystore.HDKeystore
	// 节点客户端
	client *Client
)

func init() {

	storage = keystore.NewHDKeystore(keyDir, MasterKey, keystore.StandardScryptN, keystore.StandardScryptP)

}

func GetAddressesByAccount(walletID string) ([]string, error) {

	var (
		addresses = make([]string, 0)
	)

	request := []interface{}{
		walletID,
	}

	result, err := client.Call("getaddressesbyaccount", request)
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		addresses = append(addresses, a.String())
	}

	return addresses, nil

}

//GetAddressesFromLocalDB 从本地数据库
func GetAddressesFromLocalDB(walletID string) ([]*Address, error) {

	wallet, err := GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var addresses []*Address
	err = db.Find("Account", walletID, &addresses)
	if err != nil {
		return nil, err
	}

	return addresses, nil

}

//ImportPrivKey 导入私钥
func ImportPrivKey(wif, walletID string) error {

	request := []interface{}{
		wif,
		walletID,
		false,
	}

	_, err := client.Call("importprivkey", request)

	if err != nil {
		return err
	}

	return err

}

//ImportMulti 批量导入地址和私钥
func ImportMulti(addresses []*Address, keys []string, walletID string, watchOnly bool) ([]int, error) {

	/*
		[
		{
			"scriptPubKey" : { "address": "1NL9w5fP9kX2D9ToNZPxaiwFJCngNYEYJo" },
			"timestamp" : 0,
			"label" : "Personal"
		},
		{
			"scriptPubKey" : "76a9149e857da0a5b397559c78c98c9d3f7f655d19c68688ac",
			"timestamp" : 1493912405,
			"label" : "TestFailure"
		}
		]' '{ "rescan": true }'
	*/

	var (
		request     []interface{}
		imports     = make([]interface{}, 0)
		failedIndex = make([]int, 0)
	)

	if len(addresses) != len(keys) {
		return nil, errors.New("Import addresses is not equal keys count!")
	}

	for i, a := range addresses {
		k := keys[i]
		obj := map[string]interface{}{
			"scriptPubKey": map[string]interface{}{
				"address": a.Address,
			},
			"label":     walletID,
			"timestamp": "now",
			"watchonly": watchOnly,
		}

		if !watchOnly {
			obj["keys"] = []string{k}
		}

		imports = append(imports, obj)
	}

	request = []interface{}{
		imports,
		map[string]interface{}{
			"rescan": false,
		},
	}

	result, err := client.Call("importmulti", request)
	if err != nil {
		return nil, err
	}

	for i, r := range result.Array() {
		if !r.Get("success").Bool() {
			failedIndex = append(failedIndex, i)
		}
	}

	return failedIndex, err

}

//GetCoreWalletinfo 获取核心钱包节点信息
func GetCoreWalletinfo() error {

	_, err := client.Call("getwalletinfo", nil)

	if err != nil {
		return err
	}

	return err

}

//UnlockWallet 解锁钱包
func UnlockWallet(passphrase string, seconds int) error {

	request := []interface{}{
		passphrase,
		seconds,
	}

	_, err := client.Call("walletpassphrase", request)
	if err != nil {
		return err
	}

	return err

}

//LockWallet 锁钱包
func LockWallet() error {

	_, err := client.Call("walletlock", nil)
	if err != nil {
		return err
	}

	return err
}

//GetNetworkInfo 获取网络信息
func GetNetworkInfo() error {

	_, err := client.Call("getnetworkinfo", nil)
	if err != nil {
		return err
	}

	return err
}

//KeyPoolRefill 重新填充私钥池
func KeyPoolRefill(keyPoolSize uint64) error {

	request := []interface{}{
		keyPoolSize,
	}

	_, err := client.Call("keypoolrefill", request)
	if err != nil {
		return err
	}

	return nil
}

//CreateReceiverAddress 给指定账户创建地址
func CreateReceiverAddress(account string) (string, error) {

	request := []interface{}{
		account,
	}

	result, err := client.Call("getnewaddress", request)
	if err != nil {
		return "", err
	}

	return result.String(), err

}

//CreateBatchAddress 批量创建地址
func CreateBatchAddress(name, password string, count uint64) (string, error) {

	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//读取钱包
	w, err := GetWalletInfo(name)
	if err != nil {
		return "", err
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return "", err
	}

	timestamp := time.Now()
	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
	filePath := filepath.Join(addressDir, filename)

	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*Address)
	defer close(worker)

	//保存地址过程
	saveAddressWork := func(addresses chan []*Address, filename string, wallet *Wallet) {

		var (
			saveErr error
		)

		for {
			//回收创建的地址
			getAddrs := <-addresses

			//批量写入数据库
			saveErr = saveAddressToDB(getAddrs, wallet)
			//数据保存成功才导出文件
			if saveErr == nil {
				//导出一批地址
				exportAddressToFile(getAddrs, filename)
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//解锁钱包
	err = UnlockWallet(password, 3600)
	if err != nil {
		return "", err
	}

	/*	开启导出的线程，监听新地址，批量导出	*/

	go saveAddressWork(worker, filePath, w)

	/*	计算synCount个线程，内部运行的次数	*/

	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {

		for i := uint64(0); i < synCount; i++ {

			//开始创建地址
			log.Printf("Start create address thread[%d]\n", i)
			s := i * runCount
			e := (i + 1) * runCount
			go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		s := count - otherCount
		e := count
		go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

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
			LockWallet()
			log.Printf("All addresses have been created!")
			return filePath, nil
		}
	}

	LockWallet()

	return filePath, nil
}

//CreateNewWallet 创建钱包
func CreateNewWallet(name, password string) (string, error) {

	var (
		err     error
		wallets []*Wallet
	)

	//检查钱包名是否存在
	wallets, err = GetWalletKeys(keyDir)
	for _, w := range wallets {
		if w.Alias == name {
			return "", errors.New("The wallet's alias is duplicated!")
		}
	}

	fmt.Printf("Verify password in bitcoin-core wallet...\n")

	err = EncryptWallet(password)
	if err != nil {
		//钱包已经加密，解锁钱包1秒，检查密码
		err = UnlockWallet(password, 1)
		if err != nil {
			return "", errors.New("The wallet's password is not equal bitcoin-core wallet!\n")
		}
	} else {
		//加密钱包后，需要10秒后重启bitcoin core
		fmt.Printf("Start node server... \n")
		time.Sleep(10 * time.Second)
		startNode()
	}

	fmt.Printf("Create new wallet keystore...\n")

	keyFile, err := keystore.StoreHDKey(keyDir, MasterKey, name, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return "", err
	}

	return keyFile, nil
}

//EncryptWallet 通过密码加密钱包，只在第一次加密码时才有效
func EncryptWallet(password string) error {

	request := []interface{}{
		password,
	}

	_, err := client.Call("encryptwallet", request)
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

//GetWalletList 获取钱包列表
func GetWalletList() ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		balance := GetWalletBalance(w.WalletID)
		w.Balance = balance
	}

	return wallets, nil
}

//GetWalletInfo 获取钱包列表
func GetWalletInfo(walletID string) (*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			balance := GetWalletBalance(w.WalletID)
			w.Balance = balance

			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWalletBalance 获取钱包余额
func GetWalletBalance(name string) string {

	//request := []interface{}{
	//	name,
	//	1,
	//	true,
	//}

	utxos, err := ListUnspent(1)
	if err != nil {
		return "0"
	}
	balance := decimal.New(0, 0)

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, utxo := range utxos {
		if name == utxo.Account {
			amount, _ := decimal.NewFromString(utxo.Amount)
			balance = balance.Add(amount)
		}
	}

	//for _, u := range utxos {
	//	amount, _ := decimal.NewFromString(u.Amount)
	//	balance = balance.Add(amount)
	//}

	//balance, err := client.Call("getbalance", request)
	//if err != nil {
	//	return "", err
	//}

	return balance.StringFixed(8)
}

//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func CreateNewPrivateKey(key *keystore.HDKey, start, index uint64) (string, *Address, error) {

	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	childKey, err := key.DerivedKeyWithPath(derivedPath)

	privateKey, err := childKey.ECPrivKey()
	if err != nil {
		return "", nil, err
	}

	cfg := chaincfg.MainNetParams
	if isTestNet {
		cfg = chaincfg.TestNet3Params
	}

	wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	if err != nil {
		return "", nil, err
	}

	address, err := childKey.Address(&cfg)
	if err != nil {
		return "", nil, err
	}

	addr := Address{
		Address:   address.String(),
		Account:   key.RootId,
		HDPath:    derivedPath,
		CreatedAt: time.Now(),
	}

	return wif.String(), &addr, err
}

//CreateBatchPrivateKey
func CreateBatchPrivateKey(key *keystore.HDKey, count uint64) ([]string, error) {

	var (
		wifs = make([]string, 0)
	)

	start := time.Now().Unix()
	for i := uint64(0); i < count; i++ {
		wif, _, err := CreateNewPrivateKey(key, uint64(start), i)
		if err != nil {
			continue
		}
		wifs = append(wifs, wif)
	}

	return wifs, nil

}

//BackupWalletData 备份钱包
func BackupWalletData(dest string) error {

	request := []interface{}{
		dest,
	}

	_, err := client.Call("backupwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//BackupWallet 备份数据
func BackupWallet(walletID string) (string, error) {
	w, err := GetWalletInfo(walletID)
	if err != nil {
		return "", err
	}

	//创建备份文件夹
	newBackupDir := filepath.Join(backupDir, w.FileName()+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	//创建临时备份文件wallet.dat
	tmpWalletDat := fmt.Sprintf("tmp-walllet-%d.dat", time.Now().Unix())
	tmpWalletDat = filepath.Join(walletDataPath, tmpWalletDat)

	//1. 备份核心钱包的wallet.dat
	err = BackupWalletData(tmpWalletDat)
	if err != nil {
		return "", err
	}

	//复制临时文件到备份文件夹
	file.Copy(tmpWalletDat, filepath.Join(newBackupDir, "wallet.dat"))

	//删除临时文件
	file.Delete(tmpWalletDat)

	//2. 备份种子文件
	file.Copy(filepath.Join(keyDir, w.FileName()+".key"), newBackupDir)

	//3. 备份地址数据库
	file.Copy(filepath.Join(dbPath, w.FileName()+".db"), newBackupDir)

	return newBackupDir, nil
}

//RestoreWallet 恢复钱包
func RestoreWallet(keyFile, dbFile, datFile, password string) error {

	//根据流程，提供种子文件路径，wallet.dat文件的路径，钱包数据库文件的路径。
	//输入钱包密码。
	//先备份核心钱包原来的wallet.dat到临时文件夹。
	//关闭钱包节点，复制wallet.dat到钱包的data目录下。
	//启动钱包，通过GetCoreWalletinfo检查钱包是否启动了。
	//检查密码是否可以解析种子文件，是否可以解锁钱包。
	//如果密码错误，关闭钱包节点，恢复原钱包的wallet.dat。
	//重新启动钱包。
	//复制种子文件到data/btc/key/。
	//复制钱包数据库文件到data/btc/db/。

	var (
		restoreSuccess = false
		err            error
		key            *keystore.HDKey
		sleepTime = 30 * time.Second
	)

	fmt.Printf("Validating key file... \n")

	//检查密码是否可以解析种子文件，是否可以解锁钱包。
	key, err = storage.GetKey("", keyFile, password)
	if err != nil {
		return errors.New("Passowrd is incorrect!")
	}

	//钱包当前的dat文件
	curretWDFile := filepath.Join(walletDataPath, "wallet.dat")

	//创建临时备份文件wallet.dat，备份
	tmpWalletDat := fmt.Sprintf("restore-walllet-%d.dat", time.Now().Unix())
	tmpWalletDat = filepath.Join(walletDataPath, tmpWalletDat)


	fmt.Printf("Backup current wallet.dat file... \n")

	err = BackupWalletData(tmpWalletDat)
	if err != nil {
		return err
	}

	//调试使用
	//file.Copy(curretWDFile, tmpWalletDat)

	fmt.Printf("Stop node server... \n")

	//关闭钱包节点
	stopNode()
	time.Sleep(sleepTime)

	fmt.Printf("Restore wallet.dat file... \n")

	//删除当前钱包文件
	file.Delete(curretWDFile)

	//恢复备份dat到钱包数据目录
	err = file.Copy(datFile, walletDataPath)
	if err != nil {
		return err
	}

	fmt.Printf("Start node server... \n")

	//重新启动钱包
	startNode()
	time.Sleep(sleepTime)

	fmt.Printf("Validating wallet password... \n")

	//检查wallet.dat是否可以解锁钱包
	err = UnlockWallet(password, 1)
	if err != nil {
		restoreSuccess = false
		err = errors.New("Password is incorrect!")
	} else {
		restoreSuccess = true
	}

	if restoreSuccess {
		/* 恢复成功 */

		fmt.Printf("Restore wallet key and datebase file... \n")

		//复制种子文件到data/btc/key/
		file.MkdirAll(keyDir)
		file.Copy(keyFile, filepath.Join(keyDir, key.FileName()+".key"))

		//复制钱包数据库文件到data/btc/db/
		file.MkdirAll(dbPath)
		file.Copy(dbFile, filepath.Join(dbPath, key.FileName()+".db"))

		fmt.Printf("Backup wallet has been restored. \n")

		err = nil
	} else {
		/* 恢复失败还远原来的文件 */

		fmt.Printf("Wallet unlock password is incorrect. \n")

		fmt.Printf("Stop node server... \n")

		//关闭钱包节点
		stopNode()
		time.Sleep(sleepTime)

		fmt.Printf("Restore original wallet.data... \n")

		//删除当前钱包文件
		file.Delete(curretWDFile)

		file.Copy(tmpWalletDat, curretWDFile)

		fmt.Printf("Start node server... \n")

		//重新启动钱包
		startNode()
		time.Sleep(sleepTime)

		fmt.Printf("Original wallet has been restored. \n")

	}

	//删除临时备份的dat文件
	file.Delete(tmpWalletDat)

	return err
}

//DumpWallet 导出钱包所有私钥文件
func DumpWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := client.Call("dumpwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//ImportWallet 导入钱包私钥文件
func ImportWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := client.Call("importwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//GetBlockChainInfo 获取钱包区块链信息
func GetBlockChainInfo() (*BlockchainInfo, error) {

	result, err := client.Call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}

	blockchain := NewBlockchainInfo(result)

	return blockchain, nil

}

//ListUnspent 获取未花记录
func ListUnspent(min uint64) ([]*Unspent, error) {

	var (
		utxos = make([]*Unspent, 0)
	)

	request := []interface{}{
		min,
	}

	result, err := client.Call("listunspent", request)
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		utxos = append(utxos, NewUnspent(&a))
	}

	return utxos, nil

}

//RebuildWalletUnspent 批量插入未花记录到本地
func RebuildWalletUnspent(walletID string) error {

	wallet, err := GetWalletInfo(walletID)
	if err != nil {
		return err
	}

	//查找核心钱包确认数大于1的
	utxos, err := ListUnspent(1)
	if err != nil {
		return err
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	//清空历史的UTXO
	db.Drop("Unspent")

	//开始事务
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, utxo := range utxos {
		var addr Address
		err = db.One("Address", utxo.Address, &addr)
		utxo.Account = addr.Account
		utxo.HDAddress = addr
		key := common.NewString(fmt.Sprintf("%s_%d_%s", utxo.TxID, utxo.Vout, utxo.Address)).SHA256()
		utxo.Key = key

		err = tx.Save(utxo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

//ListUnspentFromLocalDB 查询本地数据库的未花记录
func ListUnspentFromLocalDB(walletID string) ([]*Unspent, error) {

	var (
		wallet *Wallet
	)

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			wallet = w
			break
		}
	}

	if wallet == nil {
		return nil, errors.New("The wallet that your given name is not exist!")
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var utxos []*Unspent
	err = db.Find("Account", walletID, &utxos)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

//BuildTransaction 构建交易单
func BuildTransaction(utxos []*Unspent, to, change string, amount, fees decimal.Decimal) (string, decimal.Decimal, error) {

	var (
		inputs      = make([]interface{}, 0)
		outputs     = make(map[string]interface{})
		totalAmount = decimal.New(0, 0)
	)

	for _, u := range utxos {

		if u.Spendable {
			ua, _ := decimal.NewFromString(u.Amount)
			totalAmount = totalAmount.Add(ua)

			inputs = append(inputs, map[string]interface{}{
				"txid": u.TxID,
				"vout": u.Vout,
			})
		}

	}

	if totalAmount.LessThan(amount) {
		return "", decimal.New(0, 0), errors.New("The balance is not enough!")
	}

	changeAmount := totalAmount.Sub(amount).Sub(fees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		//ca, _ := changeAmount.Float64()
		outputs[change] = changeAmount.StringFixed(8)

		fmt.Printf("Create change address for receiving %s coin.\n", outputs[change])
	}

	//ta, _ := amount.Float64()
	outputs[to] = amount.StringFixed(8)

	request := []interface{}{
		inputs,
		outputs,
	}

	rawTx, err := client.Call("createrawtransaction", request)
	if err != nil {
		return "", decimal.New(0, 0), err
	}

	return rawTx.String(), changeAmount, nil
}

//SignRawTransaction 钱包交易单
func SignRawTransaction(txHex, walletID string, key *keystore.HDKey, utxos []*Unspent) (string, error) {

	var (
		wifs = make([]string, 0)
	)

	//查找未花签名需要的私钥
	for _, u := range utxos {

		childKey, err := key.DerivedKeyWithPath(u.HDAddress.HDPath)

		privateKey, err := childKey.ECPrivKey()
		if err != nil {
			return "", err
		}

		cfg := chaincfg.MainNetParams
		if isTestNet {
			cfg = chaincfg.TestNet3Params
		}

		wif, err := btcutil.NewWIF(privateKey, &cfg, true)
		if err != nil {
			return "", err
		}

		wifs = append(wifs, wif.String())

	}

	request := []interface{}{
		txHex,
		utxos,
		wifs,
	}

	result, err := client.Call("signrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.Get("hex").String(), nil

}

//SendRawTransaction 广播交易
func SendRawTransaction(txHex string) (string, error) {

	request := []interface{}{
		txHex,
	}

	result, err := client.Call("sendrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil

}

//SendTransaction 发送交易
func SendTransaction(walletID, to string, amount decimal.Decimal, password string, feesInSender bool) ([]string, error) {

	var (
		usedUTXO   []*Unspent
		balance    = decimal.New(0, 0)
		totalSend  = amount
		actualFees = decimal.New(0, 0)
		sendTime   = 1
		txIDs      = make([]string, 0)
	)

	utxos, err := ListUnspentFromLocalDB(walletID)
	if err != nil {
		return nil, err
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{utxos, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//读取钱包
	w, err := GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	totalBalance, _ := decimal.NewFromString(w.Balance)
	if totalBalance.LessThanOrEqual(amount) && feesInSender {
		return nil, errors.New("The wallet's balance is not enough!")
	} else if totalBalance.LessThan(amount) && !feesInSender {
		return nil, errors.New("The wallet's balance is not enough!")
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return nil, err
	}

	//解锁钱包
	err = UnlockWallet(password, 120)
	if err != nil {
		return nil, err
	}

	//创建找零地址
	changeAddr, err := CreateChangeAddress(walletID, key)
	if err != nil {
		return nil, err
	}

	feesRate, err := EstimateFeeRate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Calculating wallet unspent record to build transaction...\n")

	//循环的计算余额是否足够支付发送数额+手续费
	for {

		usedUTXO = make([]*Unspent, 0)
		balance = decimal.New(0, 0)

		//计算一个可用于支付的余额
		for _, u := range utxos {

			if u.Spendable {
				ua, _ := decimal.NewFromString(u.Amount)
				balance = balance.Add(ua)
				usedUTXO = append(usedUTXO, u)
				if balance.GreaterThanOrEqual(totalSend) {
					break
				}
			}
		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		fees, err := EstimateFee(int64(len(usedUTXO)), 2, feesRate)
		if err != nil {
			return nil, err
		}

		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
		if feesInSender {
			//总共要发送的
			totalSend = amount.Add(fees)
			if totalSend.GreaterThan(balance) {
				continue
			}
			totalSend = amount
		} else {

			if fees.GreaterThanOrEqual(amount) {
				return nil, errors.New("The sent amount is not enough for fees!")
			}

			totalSend = amount.Sub(fees)
		}

		actualFees = fees

		break

	}

	changeAmount := balance.Sub(totalSend).Sub(actualFees)

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("From WalletID: %s\n", walletID)
	fmt.Printf("To Address: %s\n", to)
	fmt.Printf("Use: %v\n", balance.StringFixed(8))
	fmt.Printf("Fees: %v\n", actualFees.StringFixed(8))
	fmt.Printf("Receive: %v\n", totalSend.StringFixed(8))
	fmt.Printf("Change: %v\n", changeAmount.StringFixed(8))
	fmt.Printf("-----------------------------------------------\n")

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > maxTxInputs {
		sendTime = int(math.Ceil(float64(len(usedUTXO)) / float64(maxTxInputs)))
	}

	for i := 0; i < sendTime; i++ {

		var sendUxto []*Unspent
		var pieceOfSend = decimal.New(0, 0)

		s := i * maxTxInputs

		//最后一个，计算余数
		if i == sendTime-1 {
			sendUxto = usedUTXO[s:]

			pieceOfSend = totalSend
		} else {
			sendUxto = usedUTXO[s : s+maxTxInputs]

			for _, u := range sendUxto {
				ua, _ := decimal.NewFromString(u.Amount)
				pieceOfSend = pieceOfSend.Add(ua)
			}

		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		piecefees, err := EstimateFee(int64(len(sendUxto)), 2, feesRate)
		if piecefees.LessThan(decimal.NewFromFloat(0.00001)) {
			piecefees = decimal.NewFromFloat(0.00001)
		}
		if err != nil {
			return nil, err
		}

		//解锁钱包
		err = UnlockWallet(password, 120)
		if err != nil {
			return nil, err
		}

		//log.Printf("pieceOfSend = %s \n", pieceOfSend.StringFixed(8))
		//log.Printf("piecefees = %s \n", piecefees.StringFixed(8))
		//log.Printf("feesRate = %s \n", feesRate.StringFixed(8))

		//创建交易
		txRaw, _, err := BuildTransaction(sendUxto, to, changeAddr.Address, pieceOfSend, piecefees)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Build Transaction Successfully\n")

		//签名交易
		signedHex, err := SignRawTransaction(txRaw, walletID, key, sendUxto)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Sign Transaction Successfully\n")

		txid, err := SendRawTransaction(signedHex)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Submit Transaction Successfully\n")

		txIDs = append(txIDs, txid)
		//txIDs = append(txIDs, signedHex)

		//减去已发送的
		totalSend = totalSend.Sub(pieceOfSend)

	}

	//发送成功后，删除已使用的UTXO
	clearUnspends(usedUTXO, w)

	LockWallet()

	return txIDs, nil
}

//CreateChangeAddress 创建找零地址
func CreateChangeAddress(walletID string, key *keystore.HDKey) (*Address, error) {

	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	go createAddressWork(key, producer, walletID, uint64(time.Now().Unix()), 0, 1)

	//回收创建的地址
	getAddrs := <-producer

	if len(getAddrs) == 0 {
		return nil, errors.New("Change address creation failed!")
	}

	//批量写入数据库
	err := saveAddressToDB(getAddrs, &Wallet{Alias: key.Alias, WalletID: key.RootId})
	if err != nil {
		return nil, err
	}

	return getAddrs[0], nil
}

//EstimateFee 预估手续费
func EstimateFee(inputs, outputs int64, feeRate decimal.Decimal) (decimal.Decimal, error) {

	var piece int64 = 1

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if inputs > int64(maxTxInputs) {
		piece = int64(math.Ceil(float64(inputs) / float64(maxTxInputs)))
	}

	//计算公式如下：148 * 输入数额 + 34 * 输出数额 + 10
	trx_bytes := decimal.New(inputs*148+outputs*34+piece*10, 0)
	trx_fee := trx_bytes.Div(decimal.New(1000, 0)).Mul(feeRate)

	return trx_fee, nil
}

//EstimateFeeRate 预估的没KB手续费率
func EstimateFeeRate() (decimal.Decimal, error) {

	defaultRate, _ := decimal.NewFromString("0.0001")

	//估算交易大小 手续费
	request := []interface{}{
		2,
	}

	result, err := client.Call("estimatefee", request)
	if err != nil {
		return decimal.New(0, 0), err
	}

	feeRate, _ := decimal.NewFromString(result.String())

	if feeRate.LessThan(defaultRate) {
		feeRate = defaultRate
	}


	return feeRate, nil
}

//SummaryWallets 执行汇总流程
func SummaryWallets() {

	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range walletsInSum {

		//重新加载utxo
		RebuildWalletUnspent(wid)

		//统计钱包最新余额
		wb := GetWalletBalance(wid)

		balance, _ := decimal.NewFromString(wb)
		//如果余额大于阀值，汇总的地址
		if balance.GreaterThan(threshold) {

			log.Printf("Summary account[%s]balance = %v \n", wallet.WalletID, balance)
			log.Printf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)

			txID, err := SendTransaction(wallet.WalletID, sumAddress, balance, wallet.Password, false)
			if err != nil {
				log.Printf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
				continue
			} else {
				log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, sumAddress, txID)
			}
		} else {
			log.Printf("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v\n", wallet.Alias, wallet.WalletID, balance, threshold)
		}
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

//AddWalletInSummary 添加汇总钱包账户
func AddWalletInSummary(wid string, wallet *Wallet) {
	walletsInSum[wid] = wallet
}

//clearUnspends 清楚已使用的UTXO
func clearUnspends(utxos []*Unspent, wallet *Wallet) {
	db, err := wallet.OpenDB()
	if err != nil {
		return
	}
	defer db.Close()

	//开始事务
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, utxo := range utxos {
		tx.DeleteStruct(utxo)
	}

	tx.Commit()
}

//createAddressWork 创建地址过程
func createAddressWork(k *keystore.HDKey, producer chan<- []*Address, walletID string, index, start, end uint64) {

	runAddress := make([]*Address, 0)
	runWIFs := make([]string, 0)

	for i := start; i < end; i++ {
		// 生成地址
		wif, address, errRun := CreateNewPrivateKey(k, index, i)
		if errRun != nil {
			log.Printf("Create new privKey failed unexpected error: %v\n", errRun)
			continue
		}

		////导入私钥
		//errRun = ImportPrivKey(wif, alias)
		//if errRun != nil {
		//	log.Printf("Import privKey failed unexpected error: %v\n", errRun)
		//	continue
		//}

		runAddress = append(runAddress, address)
		runWIFs = append(runWIFs, wif)
	}

	//批量导入私钥
	failed, errRun := ImportMulti(runAddress, runWIFs, walletID, CoreWalletWatchOnly)
	if errRun != nil {
		producer <- make([]*Address, 0)
		return
	}

	//删除导入失败的
	for _, fi := range failed {
		runAddress = append(runAddress[:fi], runAddress[fi+1:]...)
	}

	//生成完成
	producer <- runAddress
}

//generateSeed 创建种子
func generateSeed() []byte {
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil
	}

	return seed
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

//saveAddressToDB 保存地址到数据库
func saveAddressToDB(addrs []*Address, wallet *Wallet) error {
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range addrs {
		err = tx.Save(a)
		if err != nil {
			continue
		}
	}

	return tx.Commit()

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
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	sumAddress = c.String("sumAddress")
	rpcUser = c.String("rpcUser")
	rpcPassword = c.String("rpcPassword")
	nodeInstallPath = c.String("nodeInstallPath")
	isTestNet, _ = c.Bool("isTestNet")
	if isTestNet {
		walletDataPath = c.String("testNetDataPath")
	} else {
		walletDataPath = c.String("mainNetDataPath")
	}


	token := basicAuth(rpcUser, rpcPassword)

	client = &Client{
		BaseURL:     serverAPI,
		Debug:       false,
		AccessToken: token,
	}

	return nil
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

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
