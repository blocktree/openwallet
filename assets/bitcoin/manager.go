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
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	storage *keystore.HDKeystore
)

func init() {
	storage = keystore.NewHDKeystore(keyDir, MasterKey, keystore.StandardScryptN, keystore.StandardScryptP)
}

func GetAddressesByAccount(account string) ([]string, error) {

	var (
		addresses = make([]string, 0)
	)

	request := []interface{}{
		account,
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

//ImportPrivKey 导入私钥
func ImportPrivKey(wif, name string) error {

	request := []interface{}{
		wif,
		name,
		false,
	}

	_, err := client.Call("importprivkey", request)

	if err != nil {
		return err
	}

	return err

}

//ImportMulti 批量导入地址和私钥
func ImportMulti(addresses, keys []string, name string) ([]int, error) {

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
				"address": a,
			},
			"keys":      []string{k},
			"label":     name,
			"timestamp": "now",
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
	producer := make(chan []string)
	defer close(producer)

	//消费通道
	worker := make(chan []string)
	defer close(worker)

	//创建地址过程
	createAddressWork := func(k *keystore.HDKey, alias string, index, start, end uint64) {

		runAddress := make([]string, 0)
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
		failed, errRun := ImportMulti(runAddress, runWIFs, name)
		if errRun != nil {
			producer <- make([]string, 0)
		}

		//删除导入失败的
		for _, fi := range failed {
			runAddress = append(runAddress[:fi], runAddress[fi+1:]...)
		}

		//生成完成
		producer <- runAddress
	}

	//保存地址过程
	saveAddressWork := func(addresses chan []string, filename string) {

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

	//解锁钱包
	err = UnlockWallet(password, 600)
	if err != nil {
		return "", err
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
			s := i * runCount
			e := (i + 1) * runCount
			go createAddressWork(key, name, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		s := count - otherCount
		e := count
		go createAddressWork(key, name, uint64(timestamp.Unix()), s, e)

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

//CreateNewWallet 创建钱包
func CreateNewWallet(name, password string) error {

	var (
		err     error
		wallets []*Wallet
	)

	//检查钱包名是否存在
	wallets, err = GetWalletKeys(keyDir)
	for _, w := range wallets {
		if w.Alias == name {
			return errors.New("The wallet's alias is duplicated!")
		}
	}

	fmt.Printf("Verify password in bitcoin-core wallet...\n")

	err = EncryptWallet(password)
	if err != nil {
		//钱包已经加密，解锁钱包1秒，检查密码
		err = UnlockWallet(password, 1)
		if err != nil {
			return errors.New("The wallet's password is not equal bitcoin-core wallet!\n")
		}
	}

	fmt.Printf("Create new wallet keystore...\n")

	keyFile, err := keystore.StoreHDKey(keyDir, MasterKey, name, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
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

//GetWalleList 获取钱包列表
func GetWalleList() ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		balance, _ := GetWalletBalance(w.Alias)
		w.Balance = balance
	}

	return wallets, nil
}

//GetWalletInfo 获取钱包列表
func GetWalletInfo(alias string) (*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.Alias == alias {
			balance, _ := GetWalletBalance(w.Alias)
			w.Balance = balance

			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWalletBalance 获取钱包余额
func GetWalletBalance(name string) (string, error) {

	request := []interface{}{
		name,
	}

	balance, err := client.Call("getbalance", request)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func CreateNewPrivateKey(key *keystore.HDKey, start, index uint64) (string, string, error) {

	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	childKey, err := key.DerivedKeyWithPath(derivedPath)

	privateKey, err := childKey.ECPrivKey()
	if err != nil {
		return "", "", err
	}

	cfg := chaincfg.MainNetParams
	if isTestNet {
		cfg = chaincfg.TestNet3Params
	}

	wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	if err != nil {
		return "", "", err
	}

	address, err := childKey.Address(&cfg)
	if err != nil {
		return "", "", err
	}

	return wif.String(), address.String(), err
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

//generateSeed 创建种子
func generateSeed() []byte {
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil
	}

	return seed
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
