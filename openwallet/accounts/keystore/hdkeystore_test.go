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

package keystore

import (
	"testing"
	"path/filepath"
	"github.com/blocktree/OpenWallet/assets/bitcoin"
)

func TestStoreHDKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	rootId, err := StoreHDKey(path, bitcoin.MasterKey, "hello", "12345678", StandardScryptN, StandardScryptP)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
}

func TestGetKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, bitcoin.MasterKey, StandardScryptN, StandardScryptP}

	key, err := ks.GetKey("W8LoZwG22AobHZgSAyTBDRbCKHtJh9bXEL",
		"wallet-hello-W8LoZwG22AobHZgSAyTBDRbCKHtJh9bXEL.json",
		"12345678")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.RootId)
	}
}