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
)

var (
	blockHeight uint64 = uint64(241234)
	blockHash   string = "fweqfojoi23fjijojfafwqefqwefq="
)

func TestSaveLocalNewBlock(t *testing.T) {
	if err := tw.SaveLocalNewBlock(blockHeight, blockHash); err != nil {
		t.Errorf("TestSaveLocalNewBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestSaveLocalNewBlock: \n\t%+v, %+v\n", blockHeight, blockHash)
	}
}

func TestGetLocalNewBlock(t *testing.T) {
	height, hash := tw.GetLocalNewBlock()
	if height <= 0 {
		t.Errorf("TestGetLocalBlock Failed: %v\n", "height == 0")
	} else {
		fmt.Printf("TestGetLocalNewBlock: \n\t%+v, %+v\n", height, hash)
	}

}

func TestSaveLocalBlock(t *testing.T) {
	block := &Block{Height: blockHeight}
	if err := tw.SaveLocalBlock(block); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", block)
	}
}

func TestGetLocalBlock(t *testing.T) {
	block, err := tw.GetLocalBlock(blockHeight)
	if err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", block)
	}
}

func TestSaveTransaction(t *testing.T) {
	if err := tw.SaveTransaction(blockHeight); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", blockHeight)
	}
}

func TestSaveUnscanRecord(t *testing.T) {
	record := &UnscanRecord{ID: string(blockHeight)}
	if err := tw.SaveUnscanRecord(record); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", record)
	}
}

func TestGetUnscanRecords(t *testing.T) {

	if unscanRecords, err := tw.GetUnscanRecords(); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", unscanRecords)
	}

}

func TestDeleteUnscanRecord(t *testing.T) {
	if err := tw.DeleteUnscanRecord(blockHeight); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", blockHeight)
	}

}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func TestDeleteUnscanRecordNotFindTX(t *testing.T) {
	if err := tw.DeleteUnscanRecordNotFindTX(); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", "done")
	}
}

//DeleteUnscanRecordByTxID 删除未扫记录
func TestDeleteUnscanRecordByTxID(t *testing.T) {
	if err := tw.DeleteUnscanRecordByTxID(blockHeight, blockHash); err != nil {
		t.Errorf("TestGetLocalBlock Failed: %v\n", err)
	} else {
		fmt.Printf("TestGetLocalBlock: \n\t%+v\n", blockHeight)
	}
}
