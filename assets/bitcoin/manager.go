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
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/blocktree/OpenWallet/common"
	"path/filepath"
	"log"
	"github.com/blocktree/OpenWallet/common/file"
)

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

//GetWalletinfo 获取节点信息
func GetWalletinfo() error {

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

	request := []interface{} {
		account,
	}

	result, err := client.Call("getnewaddress", request)
	if err != nil {
		return "", err
	}

	return result.String(), err

}

//CreateBatchAddress 批量创建地址
func CreateBatchAddress(account string, count uint64) (string, error) {

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
	producer := make(chan []string)
	defer close(producer)

	//消费通道
	worker := make(chan []string)
	defer close(worker)

	//创建地址过程
	createAddressWork := func(runCount uint64) {

		runAddress := make([]string, 0)

		for i := uint64(0); i < runCount; i++ {
			// 请求地址
			address, errRun := CreateReceiverAddress(account)
			if errRun != nil {
				continue
			}
			runAddress = append(runAddress, address)

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