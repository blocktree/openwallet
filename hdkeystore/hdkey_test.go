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

package hdkeystore

import (
	"github.com/blocktree/go-owcrypt"
	"testing"
	"encoding/hex"
)

func TestGenerateSeed(t *testing.T) {

	for i:= 0; i<30;i++ {
		seed, err := GenerateSeed(32)
		if err != nil {
			t.Fatalf("GenerateSeed failed unexpected error: %v",err)
			return
		}

		key, err := NewHDKey(seed, "hello", OpenwCoinTypePath)
		if err != nil {
			t.Fatalf("GenerateSeed failed unexpected error: %v", err)
		}

		t.Logf("[%d] Seed = %s", i, hex.EncodeToString(seed))
		//t.Logf("[%d] Mnemonic = %s", i, key.Mnemonic())
		t.Logf("[%d] KeyID = %s", i, key.KeyID)
	}
}

func TestNewHDKey(t *testing.T) {

	tests := []struct {
		accountId string
		seed   string
		startPath 	string
	}{
		{
			accountId:   "hello",
			seed: "4b68b20a5d3ac671a61e6e94b4de309530a12439b7c3ee548d20966674696656",
			startPath:   "m/44'/88",
		},
		{
			accountId:   "hello2",
			seed: "dfac3098fd7c3cd9b9cdf44e3e1ae912e3d2ce05795a857a53ebff6111b1580b",
			startPath:   "m/44'/88'",
		},
	}

	for i, test := range tests {
		seed, _ := hex.DecodeString(test.seed)
		key, err := NewHDKey(seed, "hello", test.startPath)
		if err != nil {
			t.Fatalf("NewHDKey failed unexpected error: %v", err)
		}
		//t.Logf("Key[%d] Mnemonic = %s", i, key.Mnemonic())
		t.Logf("Key[%d] address = %s", i, key.KeyID)
		t.Logf("Key[%d] seed = %s", i, hex.EncodeToString(key.Seed()))
	}
}

func TestHDKey_DerivedKeyWithPath(t *testing.T) {
	seed, _ := GenerateSeed(32)
	key, _ := NewHDKey(seed, "hello", OpenwCoinTypePath)
	childKey, err := key.DerivedKeyWithPath("m/44'/88'/1581919647/0", owcrypt.ECC_CURVE_ED25519_NORMAL)
	if err != nil {
		t.Fatalf("DerivedKeyWithPath failed unexpected error: %v", err)
		return
	}
	t.Logf("child key: %s", hex.EncodeToString(childKey.GetPublicKeyBytes()))
}