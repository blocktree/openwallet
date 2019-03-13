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

package bopo

import (
	"path/filepath"

	"github.com/blocktree/openwallet/openwallet"
)

var (
	testAddress   string = "5ZaPXfJaLNrGnXuyXunFE4xKxakEzgTIZQ"
	testAccountID string = "simonluo"
	testWalletID  string = "walletid"

	testBlockHeight uint64 = uint64(336451)
	testBlockHash   string = "CLAYB+gzOjmJ2L6FIjpd5T9QQrGjzIuuulTSkni6mAlmgP+VTekkyjHQA6RqJ9nuxaj+9U90GBCrYtYww+NgaQ=="

	testWallet *openwallet.Wallet
	tw         *WalletManager
)

func init() {

	tw = &WalletManager{}
	tw.config = NewWalletConfig()
	tw.config.walletURL = "http://192.168.2.194:10021"
	tw.fullnodeClient = NewClient(tw.config.walletURL, true)

	testWallet = &openwallet.Wallet{
		WalletID: testWalletID,
		DBFile:   filepath.Join(tw.config.dbPath, testWalletID+".db")}
}
