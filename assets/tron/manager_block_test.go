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
		t.Logf("GetNowBlock return: \n\t%+v\n", r.Transactions[0])

		printBlock(r)
	}
}

func TestGetBlockByNum(t *testing.T) {
	var num uint64 = 1237157

	if r, err := tw.GetBlockByNum(num); err != nil {
		t.Errorf("GetBlockByNum failed: %v\n", err)
	} else {
		t.Logf("GetBlockByNum return: \n\t%+v\n", r)

		printBlock(r)
	}
}

func TestGetBlockByID(t *testing.T) {

	var blockID string = "0000000000038809c59ee8409a3b6c051e369ef1096603c7ee723c16e2376c73"

	if r, err := tw.GetBlockByID(blockID); err != nil {
		t.Errorf("GetBlockByID failed: %v\n", err)
	} else {
		t.Logf("GetBlockByID return: \n\t%+v\n", r)
	}
}

func TestGetBlockByLimitNext(t *testing.T) {

	var startSum, endSum uint64 = 100000, 100001

	if r, err := tw.GetBlockByLimitNext(startSum, endSum); err != nil {
		t.Errorf("GetBlockByLimitNext failed: %v\n", err)
	} else {
		t.Logf("GetBlockByLimitNext return: \n\t%+v\n", r)
	}
}

func TestGetBlockByLatestNum(t *testing.T) {

	var num uint64 = 3

	if r, err := tw.GetBlockByLatestNum(num); err != nil {
		t.Errorf("GetBlockByLatestNum failed: %v\n", err)
	} else {
		// t.Logf("GetBlockByLatestNum return: \n\t%+v\n", r)
		for _, v := range r.Block {
			// t.Logf("\tGetBlockByLatestNum return: \n\t%+v\n", v)
			printBlock(v)
			fmt.Println("")
		}
	}
}
