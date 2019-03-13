/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package ethereum

import (
	"github.com/blocktree/openwallet/log"
	"testing"
)

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager()
	wm.Config.ServerAPI = "http://47.106.102.2:10001"
	//wm.Config.ServerAPI = "https://mainnet.nebulas.io"
	wm.WalletClient = &Client{BaseURL: wm.Config.ServerAPI, Debug: true}
	return wm
}

func TestWalletManager_GetErc20TokenEvent(t *testing.T) {
	wm := testNewWalletManager()
	txid := "0xb2356cbc9f9926c2cdbc54d9d1d2105fae1d94a08b436a99216f7ae072b1ccdd"
	txevent, err := wm.GetErc20TokenEvent(txid)
	if err != nil {
		t.Errorf("GetErc20TokenEvent error: %v", err)
		return
	}
	log.Infof("txevent: %+v", txevent)
}
