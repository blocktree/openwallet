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
	"testing"
	"github.com/pborman/uuid"
)

func TestGetBTCBlockHeight(t *testing.T) {
	height, err := GetBlockHeight()
	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHeight height = %d \n", height)
}


func TestBTCBlockScanner_GetCurrentBlockHeight(t *testing.T) {
	bs := NewBTCBlockScanner()
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("GetCurrentBlockHeight height = %d \n", header.Height)
	t.Logf("GetCurrentBlockHeight hash = %v \n", header.Hash)
}

func TestGetBlockHeight(t *testing.T) {
	height, _ := GetBlockHeight()
	t.Logf("GetBlockHeight height = %d \n", height)
}

func TestGetLocalNewBlock(t *testing.T) {
	height, hash := GetLocalNewBlock()
	t.Logf("GetLocalBlockHeight height = %d \n", height)
	t.Logf("GetLocalBlockHeight hash = %v \n", hash)
}

func TestSaveLocalBlockHeight(t *testing.T) {
	bs := NewBTCBlockScanner()
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("SaveLocalBlockHeight height = %d \n", header.Height)
	t.Logf("GetLocalBlockHeight hash = %v \n", header.Hash)
	SaveLocalNewBlock(header.Height, header.Hash)
}

func TestGetBlockHash(t *testing.T) {
	//height := GetLocalBlockHeight()
	hash, err := GetBlockHash(1354918)
	if err != nil {
		t.Errorf("GetBlockHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHash hash = %s \n", hash)
}

func TestGetBlock(t *testing.T) {
	raw, err := GetBlock("000000000000000127454a8c91e74cf93ad76752cceb7eb3bcff0c398ba84b1f")
	if err != nil {
		t.Errorf("GetBlock failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlock = %v \n", raw)
}

func TestGetTransaction(t *testing.T) {
	raw, err := GetTransaction("6ecc5bc424d6b7680c8154c99d42b3caead7380dfd65998de62c758cfe52f9c8")
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTransaction = %v \n", raw)
}

func TestGetTxIDsInMemPool(t *testing.T) {
	txids, err := GetTxIDsInMemPool()
	if err != nil {
		t.Errorf("GetTxIDsInMemPool failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTxIDsInMemPool = %v \n", txids)
}

func TestBTCBlockScanner_scanning(t *testing.T) {

	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	wallet, err := GetWallet(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_scanning failed unexpected error: %v\n", err)
		return
	}

	bs := NewBTCBlockScanner()

	bs.DropRechargeRecords(accountID)

	SaveLocalNewBlock(1355030, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	bs.AddAddress(address, accountID, wallet.ToOpenWallet())

	bs.scanBlock()
}

func TestBTCBlockScanner_Run(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	wallet, err := GetWallet(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
		return
	}

	bs := NewBTCBlockScanner()

	bs.DropRechargeRecords(accountID)

	SaveLocalNewBlock(1355359, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	bs.AddAddress(address, accountID, wallet.ToOpenWallet())

	bs.Run()

	<- endRunning

}

func TestWallet_GetRecharges(t *testing.T) {
	accountID := "WFvvr5q83WxWp1neUMiTaNuH7ZbaxJFpWu"
	wallet, err := GetWallet(accountID)
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	recharges, err := wallet.ToOpenWallet().GetRecharges()
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	t.Logf("recharges.count = %v", len(recharges))
	//for _, r := range recharges {
	//	t.Logf("rechanges.count = %v", len(r))
	//}
}

func TestBTCBlockScanner_DropRechargeRecords(t *testing.T) {
	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	bs := NewBTCBlockScanner()
	bs.DropRechargeRecords(accountID)
}

func TestGetUnscanRecords(t *testing.T) {
	list, err := GetUnscanRecords()
	if err != nil {
		t.Errorf("GetUnscanRecords failed unexpected error: %v\n", err)
		return
	}

	for _, r := range list {
		t.Logf("GetUnscanRecords unscan: %v", r)
	}
}

func TestBTCBlockScanner_RescanFailedRecord(t *testing.T) {
	bs := NewBTCBlockScanner()
	bs.RescanFailedRecord()
}

func TestFullAddress (t *testing.T) {

	dic := make(map[string]string)
	for i := 0;i<20000000;i++ {
		dic[uuid.NewUUID().String()] = uuid.NewUUID().String()
	}
}