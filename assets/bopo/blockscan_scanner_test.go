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
	"testing"

	"github.com/blocktree/OpenWallet/openwallet"
)

func TestScanBlock(t *testing.T) {
	bs := NewFabricBlockScanner(tw)

	bs.scanBlock()
}

func TestScanBlock2(t *testing.T) {

	bs := NewFabricBlockScanner(tw)

	accountID := "simonluo"
	// wallet, err := tw.GetWalletInfo(accountID)
	// if err != nil {
	// 	t.Errorf("BTCBlockScanner_Run failed unexpected error: %v\n", err)
	// 	return
	// }
	wallet := &openwallet.Wallet{}

	bs.AddWallet(accountID, wallet)

	bs.ScanBlock(uint64(231234))

}
