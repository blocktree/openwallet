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

package manager

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"path/filepath"
	"testing"
)

type subscriber struct {
}

func init() {
	//tm.Init()
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriber) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Debug("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriber) BlockTxExtractDataNotify(account *openwallet.AssetsAccount, data *openwallet.TxExtractData) error {
	log.Debug("account:", account)
	log.Debug("data:", data)
	return nil
}

func TestSubscribe(t *testing.T) {
	tm := testInitWalletManager()
	var (
		endRunning = make(chan bool, 1)
	)

	sub := subscriber{}
	tm.AddObserver(&sub)
	//tm.SetRescanBlockHeight("QTUM", 236098)
	log.Debug("SupportAssets:", tm.cfg.SupportAssets)
	<-endRunning
}

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Notice("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	for i, input := range data.TxInputs {
		log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	}

	for i, output := range data.TxOutputs {
		log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	}

	log.Std.Notice("data.Transaction: %+v", data.Transaction)
	return nil
}

func TestSubscribeAddress_BTC(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "BTC"
		accountID  = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs      = map[string]string{
			"mh84vzGGA6T5s39BovBJe2dQ3r9KkWQPML":         accountID,
			"tb1qk7ya4ntkmnc6mrzj7a2ly60524hjue440egv5k": accountID,
			"mpS3FzHr1WsXWnVw15uiFzLQ9wRmo7rBJZ":         accountID,
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(1442265)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}

func TestSubscribeAddress_ETH(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "ETH"
		accountID  = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs      = map[string]string{
			"0x558ef7a2b56611ef352b2ecf5d2dd2bf548afecc": accountID,
			"0x95576498d5c2971bea1986ee92a3971de0747fc0": accountID,
			"0x73995d52f20d9d40cbc339d5d2772d9cde6b6858": accountID,
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(6693400)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}

func TestSubscribeAddress_QTUM(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "QTUM"
		accountID  = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs      = map[string]string{
			"Qf6t5Ww14ZWVbG3kpXKoTt4gXeKNVxM9QJ": accountID, //合约收币
			//"QWSTGRwdScLfdr6agUqR4G7ow4Mjc4e5re",	//合约发币
			//"QbTQBADMqSuHM6wJk2e8w1KckqK5RRYrQ6",	//主链转账
			//"QREUcesH46vMeF6frLy92aR1QC22tADNda", 	//主链转账
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	//scanner.SetRescanBlockHeight(255323)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}

func TestSubscribeAddress_LTC(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "LTC"
		accountID  = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs      = map[string]string{
			"mkSfFCHPAaHAyx9gBokXQMGWmyRtzpk4JK": accountID, //主链转账
			"my2S5LBREZ8YCcuAHZz1YChoZpGPZN28uw": accountID, //主链转账
			"mzZYMMuTNEHTQ1zqpSGkgpZvv8g3ZYWoGX": accountID,
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(	838844)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}

func TestSubscribeAddress_NAS(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "NAS"
		accountID  = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs      = map[string]string{
			"n1cT1JhXUQFxDSyBYwcn3ZBmd93nin5yrfB": accountID, //主链转账
			"n1J6e6RhZXuG1RpbcMYeeUShAo1e3QsJTXb": accountID, //主链转账
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(1215)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}
