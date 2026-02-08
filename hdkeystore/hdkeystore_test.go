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
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"
)

func TestStoreHDKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	rootId, err := StoreLockerHDKey(path, "sogosdfo456", "123TestGCM")
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
}

func TestGetKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP}

	key, err := ks.GetKey("VzPvEVCRXM4EvkNwJFVmDRSepvXRjDJcXh", "sogosdfo456-VzPvEVCRXM4EvkNwJFVmDRSepvXRjDJcXh.key", "123TestGCM")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	_ = key.Seed(func(seed []byte) error {
		fmt.Println("seed: ", hex.EncodeToString(seed))
		return nil
	})
}
