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

import (
	"encoding/hex"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"testing"
)

func TestAddressDecoder_PublicKeyToAddress(t *testing.T) {
	addr := "tb1qk87tnszyj4528l6pa86zfqcl0d90c7vvkrt7j7rxlkxy9drvxqhsmwpg6q"

	cfg := addressEncoder.BTC_testnetAddressBech32V0

	hash, err :=addressEncoder.AddressDecode(addr, cfg)
	if err != nil {
		t.Errorf("AddressDecode failed unexpected error: %v\n", err)
		return
	}
	t.Logf("hash: %s", hex.EncodeToString(hash))
}

func TestAddressDecoder_ScriptPubKeyToBech32Address(t *testing.T) {

	scriptPubKey, _ := hex.DecodeString("00140513ef61e8bb24e4b82be1366d25c8fb78693ff8")

	addr, err := ScriptPubKeyToBech32Address(scriptPubKey, false)
	if err != nil {
		t.Errorf("ScriptPubKeyToBech32Address failed unexpected error: %v\n", err)
		return
	}
	t.Logf("addr: %s", addr)

}