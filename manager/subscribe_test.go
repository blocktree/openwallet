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

func TestSubscribeAddress_ETH(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol = "ETH"
		accountID = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs = []string{
			"0x558ef7a2b56611ef352b2ecf5d2dd2bf548afecc",
			"0x95576498d5c2971bea1986ee92a3971de0747fc0",
			"0x73995d52f20d9d40cbc339d5d2772d9cde6b6858",
		}
	)

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol + ".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(4840986)


	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	for _, a := range addrs {
		scanner.AddAddress(a, accountID)
 	}


	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}


func TestSubscribeAddress_QTUM(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol = "QTUM"
		accountID = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs = []string{
			"Qf6t5Ww14ZWVbG3kpXKoTt4gXeKNVxM9QJ",	//合约收币
			//"QWSTGRwdScLfdr6agUqR4G7ow4Mjc4e5re",	//合约发币
			//"QbTQBADMqSuHM6wJk2e8w1KckqK5RRYrQ6",	//主链转账
			//"QREUcesH46vMeF6frLy92aR1QC22tADNda", 	//主链转账
		}
	)

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol + ".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(255323)


	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	for _, a := range addrs {
		scanner.AddAddress(a, accountID)
	}


	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}



func TestSubscribeAddress_LTC(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol = "LTC"
		accountID = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs = []string{
			"QZn9j1oWxcYCdL8VBKPfv1SAXNaAEYjoga",	//主链转账
			"QYkwzDhU7UyKd4hdX69c24unYjynVyYKot", 	//主链转账
		}
	)

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol + ".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(813080)


	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	for _, a := range addrs {
		scanner.AddAddress(a, accountID)
	}


	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}


func TestSubscribeAddress_NAS(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol = "NAS"
		accountID = "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
		addrs = []string{
			"n1cT1JhXUQFxDSyBYwcn3ZBmd93nin5yrfB",	//主链转账
			"n1J6e6RhZXuG1RpbcMYeeUShAo1e3QsJTXb", 	//主链转账
		}
	)

	assetsMgr, err := GetAssetsManager(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol + ".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(1215112)


	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	for _, a := range addrs {
		scanner.AddAddress(a, accountID)
	}


	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}