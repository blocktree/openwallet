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
	"testing"
)

func TestGetNowBlock(t *testing.T) {

	if r, err := tw.GetNowBlock(); err != nil {
		t.Errorf("GetNowBlock failed: %v\n", err)
	} else {
		// t.Logf("GetNowBlock return: \n\t%+v\n", r)

		printBlock(r)
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

	var blockID = "0000000000341190edc6eb2c61e2efd0b6c45177962b43cd13c0cd32da62cc0e"

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

		for _, v := range r {
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

		for _, v := range r {
			printBlock(v)
		}
	}
}
