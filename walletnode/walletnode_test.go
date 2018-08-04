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

package walletnode

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var ( // Usecase Data
	symbols = []string{"btc", "bch", "eth", "sc"}
)

func init() {
	// Clear env

	// Init data to Usecase
	testIniData := `
	`
}

func TestCreate(t *testing.T) {
	if err := nil; err != nil {
		t.Errorf("GetBlockChainInfo failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetBlockChainInfo info: %v\n", b)
	}
}

func TestGet(t *testing.T) {
}

func TestStart(t *testing.T) {
}

func TestStop(t *testing.T) {
}

func TestRestart(t *testing.T) {
}

func TestRemove(t *testing.T) {
}
