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
	"path/filepath"
	"testing"

	"github.com/blocktree/OpenWallet/openwallet"
)

var (
	bst *FabricBlockScanner

	testAddress   string = "5ZaPXfJaLNrGnXuyXunFE4xKxakEzgTIZQ"
	testAccountID string = "simonluo"

	testBlockHeight uint64 = uint64(241234)
	testBlockHash   string = ""
)

func TestNewFabricBlockScanner(t *testing.T) {
	fabricBlockScanner := NewFabricBlockScanner(tw)
	fmt.Println(fabricBlockScanner)
}

func TestAddAddress(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	wallet := &openwallet.Wallet{WalletID: testAccountID, DBFile: filepath.Join(tw.config.dbPath, testAccountID+".db")}
	bst.AddAddress(testAddress, testAccountID, wallet)
}

func TestAddWallet(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	wallet := &openwallet.Wallet{WalletID: testAccountID, DBFile: filepath.Join(tw.config.dbPath, testAccountID+".db")}
	bst.AddWallet(testAccountID, wallet)
}

// func TestAddObserver(t *testing.T) {
// 	obj := openwallet.BlockScanNotificationObject{}
// 	bst.AddObserver(obj)
// }
// func TestRemoveObserver(t *testing.T) {
// 	obj := openwallet.BlockScanNotificationObject{}
// 	bst.RemoveObserver(obj)
// }

func TestClear(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	bst.Clear()
}

// func TestRun(t *testing.T) {
// 	bst.Run()
// }
// func TestStop(t *testing.T) {
// 	bst.Stop()
// }
// func TestPause(t *testing.T) {
// 	bst.Pause()
// }
// func TestRestart(t *testing.T) {
// 	bst.Restart()
// }

func TestSetRescanBlockHeight(t *testing.T) {
	bst = NewFabricBlockScanner(tw)
	if err := bst.SetRescanBlockHeight(testBlockHeight); err != nil {
		t.Errorf("TestSetRescanBlockHeight Failed: %v\n", err)
	}
}
func TestGetCurrentBlockHeader(t *testing.T) {
	bst = NewFabricBlockScanner(tw)
	if blockHeader, err := bst.GetCurrentBlockHeader(); err != nil {
		t.Errorf("TestGetCurrentBlockHeader Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetCurrentBlockHeader Results: %+v \n", blockHeader)
	}
}
func TestIsExistAddress(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	TestAddAddress(t)

	if bst.IsExistAddress(testAddress) != true {
		t.Errorf("TestIsExistAddress Failed: %v\n", "none")
	} else {
		fmt.Printf("TestIsExistAddress Results: %+v \n", "exist")
	}
}
func TestIsExistWallet(t *testing.T) {
	bst = NewFabricBlockScanner(tw)

	TestAddWallet(t)
	if bst.IsExistWallet(testAccountID) != true {
		t.Errorf("TestIsExistWallet Failed: %v\n", "none")
	} else {
		fmt.Printf("TestIsExistWallet Results: %+v \n", "exist")
	}
}
