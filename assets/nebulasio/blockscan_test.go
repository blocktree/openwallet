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

package nebulasio

import (
	"fmt"
	"github.com/pborman/uuid"
	"testing"
)

func TestGetBTCBlockHeight(t *testing.T) {
	height, err := wm.GetBlockHeight()
	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHeight height = %d \n", height)
}


func TestBTCBlockScanner_GetCurrentBlockHeight(t *testing.T) {
	bs := NewNASBlockScanner(wm)
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("GetCurrentBlockHeight height = %d \n", header.Height)
	t.Logf("GetCurrentBlockHeight hash = %v \n", header.Hash)
}

func TestGetLocalNewBlock(t *testing.T) {
	height, hash := wm.GetLocalNewBlock()
	t.Logf("GetLocalBlockHeight height = %d \n", height)
	t.Logf("GetLocalBlockHeight hash = %v \n", hash)
}

func TestSaveLocalBlockHeight(t *testing.T) {
	bs := NewNASBlockScanner(wm)
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("SaveLocalBlockHeight height = %d \n", header.Height)
	t.Logf("GetLocalBlockHeight hash = %v \n", header.Hash)
	wm.SaveLocalNewBlock(header.Height, header.Hash)
}

func TestGetTransaction(t *testing.T) {
	raw, err := wm.GetTransaction("36319dafe14641c59f0681f57e048561e34cbefe20ad3640c6d95f4a61cce87d",1222)
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTransaction = %v \n", raw)
}

/*
func TestGetTxOut(t *testing.T) {
	raw, err := wm.GetTxOut("7768a6436475ed804344a3711e90e7f10f7db42da8918580c8b669dd63d64cc3", 0)
	if err != nil {
		t.Errorf("GetTxOut failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTxOut = %v \n", raw)
}

func TestGetTxIDsInMemPool(t *testing.T) {
	txids, err := wm.GetTxIDsInMemPool()
	if err != nil {
		t.Errorf("GetTxIDsInMemPool failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTxIDsInMemPool = %v \n", txids)
}
*/

func TestBTCBlockScanner_scanning(t *testing.T) {

	//accountID := "wjq2"
	//address := "n1Prn7ZbZtd5CTN8Yrj4K9c3gD4u8tjFQzX"

	bs := NewNASBlockScanner(wm)

	//添加观察者
	sub := subscriber{}
	bs.AddObserver(&sub)

	//重置区块扫描器的扫描开始高度
	err := bs.SetRescanBlockHeight(39027)
	if err != nil {
		t.Errorf("SetRescanBlockHeight error: %v\n", err)
		return
	}
	//tw.SaveLocalNewBlock(1355030, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	//AddAddress 添加订阅地址
	//bs.AddAddress(address, accountID)

	bs.ScanBlockTask()
}

func TestBTCBlockScanner_Run(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	//accountID := "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
	//address := "n1NCdn2vo1vz2didNfnvxPaAPZbh634CLqM"

	bs := NewNASBlockScanner(wm)

	//添加观察者
	sub := subscriber{}
	bs.AddObserver(&sub)

	bs.SetRescanBlockHeight(1165601)

	//bs.AddAddress(address, accountID)

	bs.Run()

	<- endRunning

}

func TestBTCBlockScanner_ScanBlock(t *testing.T) {

	//accountID := "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
	//address := "n1a5JWwqVug7CjvsEbVAGRE5KQsKjK2Jy56"

	bs := NewNASBlockScanner(wm)
	//bs.AddAddress(address, accountID)
	bs.ScanBlock(15037)
}

func TestBTCBlockScanner_ExtractTransaction(t *testing.T) {

	//accountID := "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
	//address := "n1a5JWwqVug7CjvsEbVAGRE5KQsKjK2Jy56"

	//bs := NewNASBlockScanner(wm)
	//bs.AddAddress(address, accountID)

	//bs.ExtractTransaction(
	//	0,
	//	"",
	//	"26e6f69d472a53c93030ae6c262d6fa5219c715966933c15a9247abd5478184b")

}

//获取钱包相关的充值记录
func TestWallet_GetRecharges(t *testing.T) {
	accountID := "W4VUMN3wxQcwVEwsRvoyuhrJ95zhyc4zRW"
	wallet, err := wm.GetWalletByID(accountID)
	if err != nil {
		t.Errorf("GetWalletByID failed unexpected error: %v\n", err)
		return
	}

	recharges, err := wallet.GetRecharges(false)
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	t.Logf("recharges.count = %v", len(recharges))
}

func TestGetUnscanRecords(t *testing.T) {
	list, err := wm.GetUnscanRecords()
	if err != nil {
		t.Errorf("GetUnscanRecords failed unexpected error: %v\n", err)
		return
	}

	for _, r := range list {
		t.Logf("GetUnscanRecords unscan: %v", r)
	}
}

func TestBTCBlockScanner_RescanFailedRecord(t *testing.T) {
	bs := NewNASBlockScanner(wm)
	bs.RescanFailedRecord()
}

func TestFullAddress (t *testing.T) {

	dic := make(map[string]string)
	for i := 0;i<20000000;i++ {
		dic[uuid.NewUUID().String()] = uuid.NewUUID().String()
	}
}


func TestGetBlockHeight(t *testing.T) {

	height ,err := wm.GetBlockHeight()

	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHeight height = %v\n", height)
}

func TestGetBlockHashByHeight(t *testing.T) {

	hash ,err := wm.GetBlockHashByHeight(1111111111)

	if err != nil {
		t.Errorf("GetBlockHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHash hash = %v \n", hash)
}

func TestGetBlockByHeight(t *testing.T) {

	block ,err := wm.GetBlockByHeight("12063")

	if err != nil {
		t.Errorf("GetBlockByHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockByHeight block = %v \n", block)
}

func TestGetBlockByHash(t *testing.T) {

	block ,err := wm.GetBlockByHash("95480cc637d0782c60f321b3600200074f468444c1399ae7bba0fc0f8007a410")

	if err != nil {
		t.Errorf("GetBlockByHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockByHash block = %v \n", block)
}

func TestCheckIsContract(t *testing.T) {

	result := CheckIsContract("n1NDcf1uxtLLGNQPK2y3MzQBfUcWVBZKioZ")
	fmt.Printf("result=%v\n",result)
}


func TestAddObserver(t *testing.T) {

 	bs := NewNASBlockScanner(wm)
	sub := subscriber{}
 	bs.AddObserver(&sub)
}

func TestCallGetsubscribe(t *testing.T) {

	result,err := wm.WalletClient.CallGetsubscribe()
	if err !=nil{
		fmt.Printf("err=%v\n",err)
	}

	fmt.Printf("topic=%+v\n",result.Get( "topic").String())
}
