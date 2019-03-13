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

package openwallet

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/go-owcdrivers/owkeychain"
)

type AddressCreateResult struct {
	Success bool
	Err     error
	Address *Address
}

// BatchCreateAddressByAccount 批量创建地址
// @account 账户
// @decoder 地址编码器
// @conf 环境配置
// @count 连续创建数量
// @workerSize 并行线程数。建议20条，并行执行5000条大约8.22秒。
func BatchCreateAddressByAccount(account *AssetsAccount, decoder AddressDecoder, count int64, workerSize int) ([]*Address, error) {

	var (
		quit         = make(chan struct{})
		done         = int64(0) //完成标记
		failed       = 0
		shouldDone   = count //需要完成的总数
		addressArr   = make([]*Address, 0)
		workPermitCH = make(chan struct{}, workerSize) //工作令牌
		producer     = make(chan AddressCreateResult)  //生产通道
		worker       = make(chan AddressCreateResult)  //消费通道
	)

	defer func() {
		close(workPermitCH)
		close(producer)
		close(worker)
	}()

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	//消费工作
	consumeWork := func(result chan AddressCreateResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				addressArr = append(addressArr, gets.Address)
			} else {
				failed++ //标记生成失败数
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//生产工作
	produceWork := func(eAccount *AssetsAccount, eDecoder AddressDecoder, eCount int64, eProducer chan AddressCreateResult) {
		addrIndex := eAccount.AddressIndex
		for i := uint64(0); i < uint64(eCount); i++ {
			workPermitCH <- struct{}{}
			addrIndex++
			go func(mAccount *AssetsAccount, mDecoder AddressDecoder, newIndex int, end chan struct{}, mProducer chan<- AddressCreateResult) {

				//生成地址
				mProducer <- CreateAddressByAccountWithIndex(mAccount, mDecoder, newIndex, 0)
				//释放
				<-end

			}(eAccount, eDecoder, addrIndex, workPermitCH, eProducer)
		}
	}

	//独立线程运行消费
	go consumeWork(worker)

	//独立线程运行生产
	go produceWork(account, decoder, count, producer)

	//以下使用生产消费模式
	batchCreateAddressRuntime(producer, worker, quit)

	if failed > 0 {
		return nil, fmt.Errorf("create address failed")
	} else {
		return addressArr, nil
	}
}

//batchCreateAddressRuntime 运行时
func batchCreateAddressRuntime(producer chan AddressCreateResult, worker chan AddressCreateResult, quit chan struct{}) {

	var (
		values = make([]AddressCreateResult, 0)
	)

	for {

		var activeWorker chan<- AddressCreateResult
		var activeValue AddressCreateResult

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}

}

func CreateAddressByAccountWithIndex(account *AssetsAccount, decoder AddressDecoder, addrIndex int, addrIsChange int64) AddressCreateResult {

	result := AddressCreateResult{
		Success: true,
	}

	if len(account.HDPath) == 0 {
		result.Success = false
		result.Err = fmt.Errorf("hdPath is empty")
		return result
	}
	hdPath := fmt.Sprintf("%s/%d/%d", account.HDPath, addrIsChange, addrIndex)
	var newKeys = make([][]byte, 0) //通过多个拥有者公钥生成地址
	for _, pub := range account.OwnerKeys {
		if len(pub) == 0 {
			continue
		}
		pubkey, err := owkeychain.OWDecode(pub)
		if err != nil {
			result.Success = false
			result.Err = err
			return result
		}
		start, err := pubkey.GenPublicChild(uint32(addrIsChange))
		newKey, err := start.GenPublicChild(uint32(addrIndex))
		newKeys = append(newKeys, newKey.GetPublicKeyBytes())
	}
	var err error
	var address, publicKey string
	if len(newKeys) > 1 {
		address, err = decoder.RedeemScriptToAddress(newKeys, uint64(account.Required), false)
		if err != nil {
			result.Success = false
			result.Err = err
			return result
		}
	} else {
		address, err = decoder.PublicKeyToAddress(newKeys[0], false)
		if err != nil {
			result.Success = false
			result.Err = err
			return result
		}
		publicKey = hex.EncodeToString(newKeys[0])
	}
	if len(address) == 0 {
		result.Success = false
		result.Err = fmt.Errorf("create address content error")
		return result
	}
	newAddr := &Address{
		AccountID: account.AccountID,
		Symbol:    account.Symbol,
		Index: uint64(addrIndex),
		Address:   address,
		Balance:   "0",
		WatchOnly: false,
		PublicKey: publicKey,
		Alias:     "",
		Tag:       "",
		HDPath:    hdPath,
		IsChange:  common.NewString(addrIsChange).Bool(),
	}

	result.Success = true
	result.Address = newAddr

	return result
}
