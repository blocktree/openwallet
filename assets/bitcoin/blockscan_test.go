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
	"github.com/blocktree/OpenWallet/openwallet"
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
	height, hash, _ := bs.GetCurrentBlockHeight()
	t.Logf("GetCurrentBlockHeight height = %d \n", height)
	t.Logf("GetCurrentBlockHeight hash = %v \n", hash)
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
	height, hash, _ := bs.GetCurrentBlockHeight()
	t.Logf("SaveLocalBlockHeight height = %d \n", height)
	t.Logf("GetLocalBlockHeight hash = %v \n", hash)
	SaveLocalNewBlock(height, hash)
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
	raw, err := GetTransaction("c427232b92286f8e99be7ed70d644ab9066dbaeccfb83a27d1a2a51242aa6cfb")
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

func TestBTCBlockScanner_ExtractRechargeRecords(t *testing.T) {

	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	height := uint64(0)
	txid := "6209d3f5e6344b1ed808e2083324a84f36dbb32fb9a1db6f3f5ebe4e3cbef342"

	bs := NewBTCBlockScanner()
	bs.AddAddress(address, accountID)

	//extracting := make(chan bool, 1)
	//extracting <- true
	err := bs.ExtractRechargeRecords(height, txid)
	if err != nil {
		t.Errorf("ExtractRechargeRecords failed unexpected error: %v\n", err)
		return
	}

	wallet, err := GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("ExtractRechargeRecords failed unexpected error: %v\n", err)
		return
	}

	db, _ := wallet.OpenDB()
	defer db.Close()
	var recharges []*openwallet.Recharge
	db.All(&recharges)
	for _, r := range recharges {
		t.Logf("rechanges = %v", r)
	}

}

func TestBTCBlockScanner_scanning(t *testing.T) {

	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	bs := NewBTCBlockScanner()

	bs.DropRechargeRecords(accountID)

	SaveLocalNewBlock(1355030, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	bs.AddAddress(address, accountID)

	bs.scanning()
}

func TestBTCBlockScanner_Run(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	bs := NewBTCBlockScanner()

	bs.DropRechargeRecords(accountID)

	SaveLocalNewBlock(1355030, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	bs.AddAddress(address, accountID)

	bs.Run()

	<- endRunning

}

func TestWallet_GetRecharges(t *testing.T) {
	accountID := "W4ruoAyS5HdBMrEeeHQTBxo4XtaAixheXQ"
	wallet, err := GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	recharges, err := wallet.GetRecharges()
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