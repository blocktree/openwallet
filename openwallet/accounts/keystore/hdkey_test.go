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
	"github.com/ethereum/go-ethereum/common"
	"testing"
	"github.com/btcsuite/btcutil/hdkeychain"
)

func TestEncodeStartPath(t *testing.T) {

	tests := []struct {
		input   string
		output 	string
	}{
		{
			input:   "m/44'/88'",
			output: "0300acd8",
		},
		{
			input:   "m/44'/88",
			output: "0300ac58",
		},
	}

	for i, test := range tests {

		path, err := encodeStartPath(test.input)

		if err != nil {

			t.Errorf("encodeStartPath #%d (%s): unexpected error: %v",
				i, test.input, err)
			continue
		}

		pathstr := common.Bytes2Hex(path)

		t.Logf("EncodeStartPath = %s", pathstr)
		if pathstr != test.output {
			t.Errorf("Path: %s not equals Output: %s", pathstr, test.output)
			continue
		}

	}

}

func TestGenerateSeed(t *testing.T) {
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		t.Fatalf("GenerateSeed failed unexpected error: %v",err)
	}
	t.Logf("Seed = %s", common.Bytes2Hex(seed))
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
		key, err := NewHDKey(common.Hex2Bytes(test.seed), test.startPath)
		if err != nil {
			t.Fatalf("NewHDKey failed unexpected error: %v", err)
		}
		t.Logf("Key[%d] address = %s", i, key.AccountId)
	}
}

