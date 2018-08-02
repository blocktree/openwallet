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

package workflow

// import (
// 	"bufio"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/astaxie/beego/config"
// 	// "github.com/tidwall/gjson"
// 	// "github.com/blocktree/OpenWallet/common"
// 	// "github.com/blocktree/OpenWallet/common/file"
// 	"github.com/blocktree/OpenWallet/keystore"
// 	"github.com/bndr/gotabulate"
// 	// "github.com/btcsuite/btcd/chaincfg"
// 	// "github.com/btcsuite/btcutil"
// 	// "github.com/btcsuite/btcutil/hdkeychain"
// 	// "github.com/codeskyblue/go-sh"
// 	"github.com/pkg/errors"
// 	"github.com/shopspring/decimal"
// 	"io/ioutil"
// 	"log"
// 	// "math"
// 	"os"
// 	"path/filepath"
// 	// "sort"
// 	"strings"
// 	//"time"
// )

// func GetAddressesByAccount(walletID string) ([]string, error) {
// 	var (
// 		addresses = make([]string, 0)
// 	)
//
// 	// loadConfig()		// 500?
//
// 	request := []interface{}{
// 		walletID,
// 	}
//
// 	result, err := client.Call("getaddressesbyaccount", request)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	array := result.Array()
// 	for _, a := range array {
// 		addresses = append(addresses, a.String())
// 	}
//
// 	return addresses, nil
//
// }
//
// //GetAddressesFromLocalDB 从本地数据库
// func GetAddressesFromLocalDB(walletID string) ([]*Address, error) {
//
// 	wallet, err := GetWalletInfo(walletID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	db, err := wallet.OpenDB()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer db.Close()
//
// 	var addresses []*Address
// 	err = db.Find("Account", walletID, &addresses)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return addresses, nil
//
// }

// //CreateReceiverAddress 给指定账户创建地址
// func CreateReceiverAddress(account string) (string, error) {
//
// 	request := []interface{}{
// 		account,
// 	}
//
// 	result, err := client.Call("getnewaddress", request)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return result.String(), err
// }
//
// //CreateBatchAddress 批量创建地址
// func CreateBatchAddress(name, password string, count uint64) (string, error) {
//
// 	var (
// 		synCount   uint64 = 20
// 		quit              = make(chan struct{})
// 		done              = 0 //完成标记
// 		shouldDone        = 0 //需要完成的总数
// 	)
//
// 	//读取钱包
// 	w, err := GetWalletInfo(name)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	//加载钱包
// 	key, err := w.HDKey(password)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	timestamp := time.Now()
// 	//建立文件名，时间格式2006-01-02 15:04:05
// 	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
// 	filePath := filepath.Join(addressDir, filename)
//
// 	//生产通道
// 	producer := make(chan []*Address)
// 	defer close(producer)
//
// 	//消费通道
// 	worker := make(chan []*Address)
// 	defer close(worker)
//
// 	//保存地址过程
// 	saveAddressWork := func(addresses chan []*Address, filename string, wallet *Wallet) {
//
// 		var (
// 			saveErr error
// 		)
//
// 		for {
// 			//回收创建的地址
// 			getAddrs := <-addresses
//
// 			//批量写入数据库
// 			saveErr = saveAddressToDB(getAddrs, wallet)
// 			//数据保存成功才导出文件
// 			if saveErr == nil {
// 				//导出一批地址
// 				exportAddressToFile(getAddrs, filename)
// 			}
//
// 			//累计完成的线程数
// 			done++
// 			if done == shouldDone {
// 				close(quit) //关闭通道，等于给通道传入nil
// 			}
// 		}
// 	}
//
// 	//解锁钱包
// 	err = UnlockWallet(password, 3600)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	/*	开启导出的线程，监听新地址，批量导出	*/
//
// 	go saveAddressWork(worker, filePath, w)
//
// 	/*	计算synCount个线程，内部运行的次数	*/
//
// 	//每个线程内循环的数量，以synCount个线程并行处理
// 	runCount := count / synCount
// 	otherCount := count % synCount
//
// 	if runCount > 0 {
//
// 		for i := uint64(0); i < synCount; i++ {
//
// 			//开始创建地址
// 			log.Printf("Start create address thread[%d]\n", i)
// 			s := i * runCount
// 			e := (i + 1) * runCount
// 			go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)
//
// 			shouldDone++
// 		}
// 	}
//
// 	if otherCount > 0 {
//
// 		//开始创建地址
// 		log.Printf("Start create address thread[REST]\n")
// 		s := count - otherCount
// 		e := count
// 		go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)
//
// 		shouldDone++
// 	}
//
// 	values := make([][]*Address, 0)
//
// 	//以下使用生产消费模式
//
// 	for {
//
// 		var activeWorker chan<- []*Address
// 		var activeValue []*Address
//
// 		//当数据队列有数据时，释放顶部，激活消费
// 		if len(values) > 0 {
// 			activeWorker = worker
// 			activeValue = values[0]
//
// 		}
//
// 		select {
//
// 		//生成者不断生成数据，插入到数据队列尾部
// 		case pa := <-producer:
// 			values = append(values, pa)
// 			//log.Printf("completed %d", len(pa))
// 			//当激活消费者后，传输数据给消费者，并把顶部数据出队
// 		case activeWorker <- activeValue:
// 			//log.Printf("Get %d", len(activeValue))
// 			values = values[1:]
//
// 		case <-quit:
// 			//退出
// 			LockWallet()
// 			log.Printf("All addresses have been created!")
// 			return filePath, nil
// 		}
// 	}
//
// 	LockWallet()
//
// 	return filePath, nil
// }
//
//
// //saveAddressToDB 保存地址到数据库
// func saveAddressToDB(addrs []*Address, wallet *Wallet) error {
// 	db, err := wallet.OpenDB()
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
//
// 	tx, err := db.Begin(true)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()
//
// 	for _, a := range addrs {
// 		err = tx.Save(a)
// 		if err != nil {
// 			continue
// 		}
// 	}
// //exportAddressToFile 导出地址到文件中
// func exportAddressToFile(addrs []*Address, filePath string) {
//
// 	var (
// 		content string
// 	)
//
// 	for _, a := range addrs {
//
// 		log.Printf("Export: %s \n", a.Address)
//
// 		content = content + a.Address + "\n"
// 	}
//
// 	file.MkdirAll(addressDir)
// 	file.WriteFile(filePath, []byte(content), true)
// }
//
// 	return tx.Commit()
//
// }
// //createAddressWork 创建地址过程
// func createAddressWork(k *keystore.HDKey, producer chan<- []*Address, walletID string, index, start, end uint64) {
//
// 	runAddress := make([]*Address, 0)
// 	runWIFs := make([]string, 0)
//
// 	for i := start; i < end; i++ {
// 		// 生成地址
// 		wif, address, errRun := CreateNewPrivateKey(k, index, i)
// 		if errRun != nil {
// 			log.Printf("Create new privKey failed unexpected error: %v\n", errRun)
// 			continue
// 		}
//
// 		////导入私钥
// 		//errRun = ImportPrivKey(wif, alias)
// 		//if errRun != nil {
// 		//	log.Printf("Import privKey failed unexpected error: %v\n", errRun)
// 		//	continue
// 		//}
//
// 		runAddress = append(runAddress, address)
// 		runWIFs = append(runWIFs, wif)
// 	}
//
// 	//批量导入私钥
// 	failed, errRun := ImportMulti(runAddress, runWIFs, walletID, CoreWalletWatchOnly)
// 	if errRun != nil {
// 		producer <- make([]*Address, 0)
// 		return
// 	}
//
// 	//删除导入失败的
// 	for _, fi := range failed {
// 		runAddress = append(runAddress[:fi], runAddress[fi+1:]...)
// 	}
//
// 	//生成完成
// 	producer <- runAddress
// }
//
// //generateSeed 创建种子
// func generateSeed() []byte {
// 	seed, err := hdkeychain.GenerateSeed(32)
// 	if err != nil {
// 		return nil
// 	}
//
// 	return seed
// }
//
//
// //CreateChangeAddress 创建找零地址
// func CreateChangeAddress(walletID string, key *keystore.HDKey) (*Address, error) { //
// 	//生产通道
// 	producer := make(chan []*Address)
// 	defer close(producer)
//
// 	go createAddressWork(key, producer, walletID, uint64(time.Now().Unix()), 0, 1)
//
// 	//回收创建的地址
// 	getAddrs := <-producer
//
// 	if len(getAddrs) == 0 {
// 		return nil, errors.New("Change address creation failed!")
// 	}
//
// 	//批量写入数据库
// 	err := saveAddressToDB(getAddrs, &Wallet{Alias: key.Alias, WalletID: key.RootId})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return getAddrs[0], nil
// }
