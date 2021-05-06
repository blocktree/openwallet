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

package openwallet

import (
	"fmt"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/log"
	"path/filepath"
	"testing"
)

var (
	testBlockchainLocal *BlockchainLocal
)

func init() {
	testBlockchainLocal = getTestDB()
}

func getTestDB() *BlockchainLocal {
	dbPath := filepath.Join("db")
	file.MkdirAll(dbPath)
	base, err := NewBlockchainLocal(filepath.Join(dbPath, "blockchain-btc.db"), false)
	if err != nil {
		log.Errorf("NewBlockchainLocal err: %v", err)
		return nil
	}
	return base
}

func TestBlockchainDAIBase_SaveCurrentBlockHead(t *testing.T) {

	base := testBlockchainLocal
	if base == nil {
		return
	}
	header := &BlockHeader{
		Hash:              "1111",
		Confirmations:     0,
		Merkleroot:        "",
		Previousblockhash: "",
		Height:            100,
		Version:           0,
		Time:              0,
		Fork:              false,
		Symbol:            "BTC",
	}
	err := base.SaveCurrentBlockHead(header)
	if err != nil {
		t.Errorf("LoadBlockchainDB err: %v", err)
		return
	}
	current, err := base.GetCurrentBlockHead("btc")
	if err != nil {
		t.Errorf("GetCurrentBlockHead err: %v", err)
		return
	}
	log.Infof("current height: %d, hash: %s", current.Height, current.Hash)
}

func TestBlockchainDAIBase_SaveLocalBlockHead(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	header := &BlockHeader{
		Hash:              "222",
		Confirmations:     0,
		Merkleroot:        "",
		Previousblockhash: "",
		Height:            222,
		Version:           0,
		Time:              0,
		Fork:              false,
		Symbol:            "BTC",
	}
	err := base.SaveLocalBlockHead(header)
	if err != nil {
		t.Errorf("SaveLocalBlockHead err: %v", err)
		return
	}

	current, err := base.GetLocalBlockHeadByHeight(222, "btc")
	if err != nil {
		t.Errorf("GetLocalBlockHeadByHeight err: %v", err)
		return
	}
	log.Infof("Get height: %d, hash: %s", current.Height, current.Hash)
}

func TestBlockchainDAIBase_SaveUnscanRecord(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	record := NewUnscanRecord(333, "333", "rpc call error", "BTC")
	err := base.SaveUnscanRecord(record)
	if err != nil {
		t.Errorf("SaveUnscanRecord err: %v", err)
		return
	}
}

func TestBlockchainDAIBase_GetUnscanRecords(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	list, err := base.GetUnscanRecords("")
	if err != nil {
		t.Errorf("GetUnscanRecords err: %v", err)
		return
	}
	for _, r := range list {
		log.Infof("unscan record: %+v", r)
	}
}

func TestBlockchainDAIBase_DeleteUnscanRecordByHeight(t *testing.T) {

	base := testBlockchainLocal
	if base == nil {
		return
	}

	err := base.DeleteUnscanRecordByHeight(333, "btc")
	if err != nil {
		t.Errorf("DeleteUnscanRecordByHeight err: %v", err)
		return
	}
}

func TestBlockchainDAIBase_DeleteUnscanRecordByID(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	id := "25a04faedff7f6f897ec047c252e6063dc1959c3d6027f91e28afa4caed715c2"
	err := base.DeleteUnscanRecordByID(id, "btc")
	if err != nil {
		t.Errorf("DeleteUnscanRecordByID err: %v", err)
		return
	}
}

func TestBlockchainDAIBase_BatchSaveLocalBlockHead(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}

	for i := 1; i < 1000000000; i++ {
		header := &BlockHeader{
			Hash:              fmt.Sprintf("hash_%d", i),
			Confirmations:     0,
			Merkleroot:        "",
			Previousblockhash: "",
			Height:            uint64(i),
			Version:           0,
			Time:              0,
			Fork:              false,
			Symbol:            "BTC",
		}
		err := base.SaveLocalBlockHead(header)
		if err != nil {
			t.Errorf("SaveLocalBlockHead err: %v", err)
			return
		}
	}
}

func TestBlockchainDAIBase_GetLocalBlockHeadByHeight(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}

	current, err := base.GetLocalBlockHeadByHeight(8525, "btc")
	if err != nil {
		t.Errorf("GetLocalBlockHeadByHeight err: %v", err)
		return
	}
	log.Infof("Get height: %d, hash: %s", current.Height, current.Hash)
}

func TestBlockchainBase_SaveTransaction(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	tx := &Transaction{
		WxID:        "w1234456",
		TxID:        "tx1234567",
		AccountID:   "myaccount",
		Coin:        Coin{Symbol:     "BTC"},
		Amount:      "123",
	}
	base.SaveTransaction(tx)
}

func TestBlockchainBase_GetTransactionsByTxID(t *testing.T) {
	base := testBlockchainLocal
	if base == nil {
		return
	}
	list, err := base.GetTransactionsByTxID("tx1234567", "")
	if err != nil {
		t.Errorf("GetTransactionsByTxID err: %v", err)
		return
	}
	for _, r := range list {
		log.Infof("transaction record: %+v", r)
	}
}
