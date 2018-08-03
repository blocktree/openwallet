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
	// "github.com/asdine/storm"
	// "github.com/blocktree/OpenWallet/common/file"
	// "github.com/blocktree/OpenWallet/keystore"
	// "github.com/tidwall/gjson"
	// "path/filepath"
	"time"
)

//Wallet 钱包模型
type Wallet struct {
	WalletID string `json:"rootid"`
	Addr     string `json:"addr"`
	Balance  string `json:"balance"`
	Alias    string `json:"alias"`
	Password string `json:"password"`
	RootPub  string `json:"rootpub"`
	KeyFile  string
}

type BlockchainInfo struct {
	Chain                string `json:"chain"`
	Blocks               uint64 `json:"blocks"`
	Headers              uint64 `json:"headers"`
	Bestblockhash        string `json:"bestblockhash"`
	Difficulty           string `json:"difficulty"`
	Mediantime           uint64 `json:"mediantime"`
	Verificationprogress string `json:"verificationprogress"`
	Chainwork            string `json:"chainwork"`
	Pruned               bool   `json:"pruned"`
}

type Address struct {
	Address   string `json:"address" storm:"id"`
	Account   string `json:"account" storm:"index"`
	HDPath    string `json:"hdpath"`
	CreatedAt time.Time
}

type User struct {
	UserKey string `storm:"id"`     // primary key
	Group   string `storm:"index"`  // this field will be indexed
	Email   string `storm:"unique"` // this field will be indexed with a unique constraint
	Name    string // this field will not be indexed
	Age     int    `storm:"index"`
}
