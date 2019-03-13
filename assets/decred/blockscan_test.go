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

package decred

import (
	"github.com/pborman/uuid"
	"testing"
	"encoding/base64"
	"github.com/blocktree/openwallet/openwallet"
	"path/filepath"
)

func TestGetBTCBlockHeight(t *testing.T) {
	height, err := tw.GetBlockHeight()
	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		t.Error(err.Error())
		return
	}
	t.Logf("GetBlockHeight height = %d \n", height)
}

func TestBTCBlockScanner_GetCurrentBlockHeight(t *testing.T) {
	bs := tw.blockscanner
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("GetCurrentBlockHeight height = %d \n", header.Height)
	t.Logf("GetCurrentBlockHeight hash = %v \n", header.Hash)
}

func TestGetBlockHeight(t *testing.T) {
	height, _ := tw.GetBlockHeight()
	t.Logf("GetBlockHeight height = %d \n", height)
}

func TestGetLocalNewBlock(t *testing.T) {
	height, hash := tw.GetLocalNewBlock()
	t.Logf("GetLocalBlockHeight height = %d \n", height)
	t.Logf("GetLocalBlockHeight hash = %v \n", hash)
}

func TestSaveLocalBlockHeight(t *testing.T) {
	bs := tw.blockscanner
	header, _ := bs.GetCurrentBlockHeader()
	t.Logf("SaveLocalBlockHeight height = %d \n", header.Height)
	t.Logf("GetLocalBlockHeight hash = %v \n", header.Hash)
	tw.SaveLocalNewBlock(header.Height, header.Hash)
}

func TestGetBlockHash(t *testing.T) {
	//height := GetLocalBlockHeight()
	hash, err := tw.GetBlockHash(9164)
	if err != nil {
		t.Errorf("GetBlockHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHash hash = %s \n", hash)
}

func TestGetBlock(t *testing.T) {
	raw, err := tw.GetBlock("00000000000034ad3fb85373c204c07158bdf6cb6ab1e4da7d697fe5628b85a3")
	if err != nil {
		t.Errorf("GetBlock failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlock = %v \n", raw)
}

func TestGetTransaction(t *testing.T) {
		raw, err := tw.GetTransaction("611081384086f2f7de1357858c03ce09391cef70918eac5295d646d7e78ffc87")
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTransaction = %v \n", raw)
}

func TestGetTxIDsInMemPool(t *testing.T) {
	txids, err := tw.GetTxIDsInMemPool()
	if err != nil {
		t.Errorf("GetTxIDsInMemPool failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTxIDsInMemPool = %v \n", txids)
}

func TestBTCBlockScanner_scanning(t *testing.T) {

	accountID := "hccharge"
	//address := "TsosMFZ2mwvRffkWY2fyyEqUiDeokDvCiek"

	wallet := &openwallet.Wallet{WalletID: accountID, DBFile:filepath.Join(tw.config.dbPath, accountID + ".db")}
	//wallet, err := tw.GetWalletInfo(accountID)
	//if err != nil {
	//	t.Errorf("BTCBlockScanner_scanning failed unexpected error: %v\n", err)
	//	return
	//}

	bs := tw.blockscanner

	bs.DropRechargeRecords(accountID)

	bs.SetRescanBlockHeight(3000)

	//bs.AddAddress(address, accountID, wallet)
	bs.AddWallet(accountID, wallet)

	bs.scanBlock()
}

func TestBTCBlockScanner_Run(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	accountID := "W7LEupZ2mdM29ay4oZgAoph4ESD8qu5faH"
	//address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	wallet, err := tw.GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
		return
	}

	bs := tw.blockscanner

	bs.AddWallet(accountID, wallet)

	bs.DropRechargeRecords(accountID)

	bs.SetRescanBlockHeight(10000)



	bs.Run()

	<-endRunning

}

func TestBTCBlockScanner_ScanBlock(t *testing.T) {

	accountID := "W7LEupZ2mdM29ay4oZgAoph4ESD8qu5faH"
	wallet, err := tw.GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
		return
	}

	bs := tw.blockscanner
	bs.AddWallet(accountID, wallet)
	bs.ScanBlock(9298)

}

func TestWallet_GetRecharges(t *testing.T) {
	//accountID := "WG4rn9R9rbr6xQyJCKYcQJCWNzcDR7WrXj"
	accountID := "W7LEupZ2mdM29ay4oZgAoph4ESD8qu5faH"
	//wallet := &openwallet.Wallet{WalletID: accountID, DBFile:filepath.Join(tw.config.dbPath, accountID + ".db")}

	wallet, err := tw.GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	recharges, err := wallet.GetRecharges(false)
	if err != nil {
		t.Errorf("GetRecharges failed unexpected error: %v\n", err)
		return
	}

	t.Logf("recharges.count = %v", len(recharges))
	for _, r := range recharges {
		t.Logf("rechanges = %v", r)
	}
}

func TestBTCBlockScanner_DropRechargeRecords(t *testing.T) {
	accountID := "W7LEupZ2mdM29ay4oZgAoph4ESD8qu5faH"
	bs := tw.blockscanner

	//address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	wallet, err := tw.GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
		return
	}

	bs.AddWallet(accountID, wallet)


	bs.DropRechargeRecords(accountID)
}

func TestGetUnscanRecords(t *testing.T) {
	list, err := tw.GetUnscanRecords()
	if err != nil {
		t.Errorf("GetUnscanRecords failed unexpected error: %v\n", err)
		return
	}

	for _, r := range list {
		t.Logf("GetUnscanRecords unscan: %v", r)
	}
}

func TestBTCBlockScanner_RescanFailedRecord(t *testing.T) {
	bs := tw.blockscanner

	accountID := "W7LEupZ2mdM29ay4oZgAoph4ESD8qu5faH"
	//address := "mpkUFiXonEZriywHUhig6PTDQXKzT6S5in"

	wallet, err := tw.GetWalletInfo(accountID)
	if err != nil {
		t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
		return
	}

	bs.AddWallet(accountID, wallet)

	bs.RescanFailedRecord()
}

func TestFullAddress(t *testing.T) {

	dic := make(map[string]string)
	for i := 0; i < 20000000; i++ {
		dic[uuid.NewUUID().String()] = uuid.NewUUID().String()
	}
}

func TestTransactionID(t *testing.T) {
	encode := "+mWTaTmBwGu08mDPY0HOrn8xs3E="
	bytes, _ := base64.StdEncoding.DecodeString(encode)
	t.Logf("id: %v", string(bytes))
	id := base64.StdEncoding.EncodeToString(bytes)
	t.Logf("encode: %v", id)
}