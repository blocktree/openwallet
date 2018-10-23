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

package tron

import (
	"fmt"
	"testing"
)

func TestGetNowBlock(t *testing.T) {

	if r, err := tw.GetNowBlock(); err != nil {
		t.Errorf("GetNowBlock failed: %v\n", err)
	} else {
		// t.Logf("GetNowBlock return: \n\t%+v\n", r.GetBlockHeader().GetRawData().GetNumber())
		// t.Logf("GetNowBlock return: \n\t%+v\n", r)

		printBlock(r)
	}
}

func TestGetNowBlockID(t *testing.T) {

	if r, err := tw.GetNowBlockID(); err != nil {
		t.Errorf("GetNowBlock failed: %v\n", err)
	} else {
		t.Logf("GetNowBlock return: \n\t%+v\n", r)
	}
}

func TestGetBlockByNum(t *testing.T) {
	var num uint64 = 3412368

	if r, err := tw.GetBlockByNum(num); err != nil {
		t.Errorf("GetBlockByNum failed: %v\n", err)
	} else {
		t.Logf("GetBlockByNum return: \n\t%+v\n", r)

		printBlock(r)
	}
}

func TestGetBlockByID(t *testing.T) {

	var blockID string
	blockID = "fdd9227774d977151d3c59ca9fb552c732aeed589f2dc807b56cd17cddb429d6"
	blockID = "0000000000341190edc6eb2c61e2efd0b6c45177962b43cd13c0cd32da62cc0e"

	if r, err := tw.GetBlockByID(blockID); err != nil {
		t.Errorf("GetBlockByID failed: %v\n", err)
	} else {
		t.Logf("GetBlockByID return: \n\t%+v\n", r)

		printBlock(r)
	}

	blockID = "0000000000341190edc6eb2c61e2efd0b6c45177962b43cd13c0cd32da62cc0a" // Error ID
	if r, err := tw.GetBlockByID(blockID); err != nil {
		t.Logf("GetBlockByID return: \n\t%+v\n", r)
	} else {
		t.Errorf("GetBlockByID failed: %v\n", err)
	}
}

func TestGetBlockByLimitNext(t *testing.T) {

	var startSum, endSum uint64 = 120 * 10000, 120*10000 + 3

	if r, err := tw.GetBlockByLimitNext(startSum, endSum); err != nil {
		t.Errorf("GetBlockByLimitNext failed: %v\n", err)
	} else {

		for _, v := range r.Block {
			printBlock(v)
		}
	}
}

func TestGetBlockByLatestNum(t *testing.T) {

	var num uint64 = 3

	if r, err := tw.GetBlockByLatestNum(num); err != nil {
		t.Errorf("GetBlockByLatestNum failed: %v\n", err)
	} else {
		t.Logf("GetBlockByLatestNum return: \n\t%+v\n", r)

		for _, v := range r.Block {
			printBlock(v)
		}
	}
}

func TestGenBlockHash(t *testing.T) {
	var blockHeight uint64
	var blockID string
	// var block *core.Block

	blockHeight = 3412368
	blockID = "0000000000341190edc6eb2c61e2efd0b6c45177962b43cd13c0cd32da62cc0e"
	if r, err := tw.GetBlockByNum(blockHeight); err != nil {
		t.Errorf("GenBlockHash failed: %v\n", err)
		return
	} else {
		pBlockID := tw.GenBlockID(r)
		fmt.Println("True to BlockID: ", blockID)
		fmt.Println("Predict BlockID: ", pBlockID)
	}

	blockHeight = 3412367
	blockID = "000000000034118fa1085f5f5e33f8c76b9479aae42d09ba6b30a47a5f788358"
	if r, err := tw.GetBlockByNum(blockHeight); err != nil {
		t.Errorf("GenBlockHash failed: %v\n", err)
		return
	} else {
		r.GetBlockHeader().GetRawData()
		pBlockID := tw.GenBlockID(r)
		fmt.Println("True to BlockID: ", blockID)
		fmt.Println("Predict BlockID: ", pBlockID)
	}

}
