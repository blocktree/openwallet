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

package nebulasio

import (
	"testing"
)



func TestGetBlockHeight(t *testing.T) {

	height ,err := wm.GetBlockHeight()

	if err != nil {
		t.Errorf("GetBlockHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHeight height = %v\n", height)
}

func TestGetBlockHashByHeight(t *testing.T) {

	hash ,err := wm.GetBlockHashByHeight("8989")

	if err != nil {
		t.Errorf("GetBlockHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockHash hash = %v \n", hash)
}

func TestGetBlockByHeight(t *testing.T) {

	block ,err := wm.GetBlockByHeight("8989")

	if err != nil {
		t.Errorf("GetBlockByHeight failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockByHeight block = %v \n", block)
}

func TestGetBlockByHash(t *testing.T) {

	block ,err := wm.GetBlockByHash("95480cc637d0782c60f321b3600200074f468444c1399ae7bba0fc0f8007a410")

	if err != nil {
		t.Errorf("GetBlockByHash failed unexpected error: %v\n", err)
		return
	}
	t.Logf("GetBlockByHash block = %v \n", block)
}