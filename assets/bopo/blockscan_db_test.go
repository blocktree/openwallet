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

package bopo

import (
	"fmt"
	"testing"

	"github.com/blocktree/OpenWallet/openwallet"
)

func TestSaveLocalNewBlock(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if err := bst.SaveLocalNewBlock(testBlockHeight, testBlockHash); err != nil {
		t.Errorf("TestSaveLocalNewBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestSaveLocalNewBlock: \n\t%+v, %+v\n", testBlockHeight, testBlockHash)
	}
}

func TestGetLocalNewBlock(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	height, hash := bst.GetLocalNewBlock()
	if height <= 0 {
		t.Errorf("TestGetLocalBlock Failed: %v\n", "height == 0")
	} else {
		fmt.Printf("TestGetLocalNewBlock: \n\t%+v, %+v\n", height, hash)
	}

}

func TestSaveLocalBlock(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	block := &Block{Height: testBlockHeight}
	if err := bst.SaveLocalBlock(block); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", block)
	}
}

func TestGetLocalBlock(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	block, err := bst.GetLocalBlock(testBlockHeight)
	if err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", block)
	}
}

func TestSaveTransaction(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if err := bst.SaveTransaction(testBlockHeight); err != nil {
		t.Errorf("TestSaveTransaction Failed: %v\n", err)
	} else {
		fmt.Printf("TestSaveTransaction: \n\t%+v\n", testBlockHeight)
	}
}

func TestSaveUnscanRecord(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	record := &UnscanRecord{ID: string(testBlockHeight)}
	if err := bst.SaveUnscanRecord(record); err != nil {
		t.Errorf("TestSaveUnscanRecord Failed: %v\n", err)
	} else {
		fmt.Printf("TestSaveUnscanRecord: \n\t%+v\n", record)
	}
}

func TestGetUnscanRecords(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if unscanRecords, err := bst.GetUnscanRecords(); err != nil {
		t.Errorf("TestGetUnscanRecords Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetUnscanRecords: \n\t%+v\n", unscanRecords)
	}

}

func TestDeleteUnscanRecord(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	TestSaveUnscanRecord(t)

	if err := bst.DeleteUnscanRecord(testBlockHeight); err != nil {
		t.Errorf("DeleteUnscanRecord Failed: %v\n", err)
	} else {
		fmt.Printf("DeleteUnscanRecord: \n\t%+v\n", testBlockHeight)
	}

}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func TestDeleteUnscanRecordNotFindTX(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if err := bst.DeleteUnscanRecordNotFindTX(); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", "done")
	}
}

//DeleteUnscanRecordByTxID 删除未扫记录
func TestDeleteUnscanRecordByTxID(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if err := bst.DeleteUnscanRecordByTxID(testBlockHeight, testBlockHash); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", testBlockHeight)
	}
}

func TestGetWalletByAddress(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if wallet, exist := bst.GetWalletByAddress(testAddress); exist != true {
		t.Errorf("TestGetWalletByAddress Failed: %v\n", "none")
	} else {
		fmt.Printf("TestGetWalletByAddress: \n\t%+v\n", wallet)
	}
}

func TestSaveRechargeToWalletDB(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	list := []*openwallet.Recharge{}
	if err := bst.SaveRechargeToWalletDB(testBlockHeight, list); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", testBlockHeight)
	}
}

func TestDeleteRechargesByHeight(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	if err := bst.DeleteRechargesByHeight(testBlockHeight); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", testBlockHeight)
	}
}
