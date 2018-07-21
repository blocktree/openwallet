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

import "testing"

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
	height, _ := bs.GetCurrentBlockHeight()
	t.Logf("GetCurrentBlockHeight height = %d \n", height)
}

func TestGetBlockHeight(t *testing.T) {
	height, _ := GetBlockHeight()
	t.Logf("GetBlockHeight height = %d \n", height)
}

func TestGetLocalBlockHeight(t *testing.T) {
	height := GetLocalBlockHeight()
	t.Logf("GetLocalBlockHeight height = %d \n", height)
}

func TestSaveLocalBlockHeight(t *testing.T) {
	bs := NewBTCBlockScanner()
	height, _ := bs.GetCurrentBlockHeight()
	t.Logf("SaveLocalBlockHeight height = %d \n", height)
	SaveLocalBlockHeight(height)
}

func TestGetBlockHash(t *testing.T) {
	height := GetLocalBlockHeight()
	hash, _ := GetBlockHash(height)
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
	raw, err := GetTransaction("ab2a602af7024c802fd79574aeb10f8d93985ddb8d9d2b95b1905d87e1f50171")
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTransaction = %v \n", raw)
}