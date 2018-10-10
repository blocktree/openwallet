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

//http://192.168.32.107:20003/insight-api/

func TestGetBlockHeightByExplorer(t *testing.T) {
	height, err := tw.getBlockHeightByExplorer()
	if err != nil {
		t.Errorf("getBlockHeightByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getBlockHeightByExplorer height = %d \n", height)
}

func TestGetBlockHashByExplorer(t *testing.T) {
	hash, err := tw.getBlockHashByExplorer(1434016)
	if err != nil {
		t.Errorf("getBlockHashByExplorer failed unexpected error: %v\n", err)
		return
	}
	t.Logf("getBlockHashByExplorer hash = %s \n", hash)
}

func TestGetBlockByExplorer(t *testing.T) {
	block, err := tw.getBlockByExplorer("0000000000002bd2475d1baea1de4067ebb528523a8046d5f9d8ef1cb60460d3")
	if err != nil {
		t.Errorf("GetBlock failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlock = %v \n", block)
}

func TestListUnspentByExplorer(t *testing.T) {
	list, err := tw.listUnspentByExplorer("msHemmfSZ3au6h9S1annGcTGrTVryRbSFV")
	if err != nil {
		t.Errorf("listUnspentByExplorer failed unexpected error: %v\n", err)
		return
	}
	for i, unspent := range list {
		t.Logf("listUnspentByExplorer[%d] = %v \n", i, unspent)
	}

}

func TestGetTransactionByExplorer(t *testing.T) {
	raw, err := tw.getTransactionByExplorer("6595e0d9f21800849360837b85a7933aeec344a89f5c54cf5db97b79c803c462")
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetTransaction = %v \n", raw)
}

