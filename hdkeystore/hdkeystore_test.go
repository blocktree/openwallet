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

func TestStoreHDKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	key, rootId, err := StoreHDKey(path, "sogosdfo", "123Test", StandardScryptN, StandardScryptP)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
	fmt.Println("seed:", hex.EncodeToString(key.Seed()))
}

func TestStoreHDKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	key, rootId, err := StoreHDKey(path, "sogosdfo", "123TestGCM", StandardScryptN, StandardScryptP, CipherAes256GCM)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
	fmt.Println("seed:", hex.EncodeToString(key.Seed()))
}

func TestGetKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP, CipherAes128CTR}

	key, err := ks.GetKey("WJQTDbPeYkHGXWDxySzbciBCgV7q9k5g8m", "sogosdfo-WJQTDbPeYkHGXWDxySzbciBCgV7q9k5g8m.key", "123Test")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	fmt.Println("seed:", hex.EncodeToString(key.Seed()))
}

func TestStoreHDKey2(t *testing.T) {
	path := filepath.Join(".", "keys")
	key, rootId, err := StoreHDKey(path, "sogosdfo123", "", StandardScryptN, StandardScryptP, CipherAes256CBC)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
	fmt.Println("seed: ", hex.EncodeToString(key.Seed()))
}

func TestGetKey2(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP, CipherAes256CBC}

	key, err := ks.GetKey("W52HTazdFjbzMMpXP1hk41ehb16ReXF22N", "sogosdfo123-W52HTazdFjbzMMpXP1hk41ehb16ReXF22N.key", "")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	fmt.Println("seed: ", hex.EncodeToString(key.Seed()))
}

func TestGetKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP, CipherAes256GCM}

	key, err := ks.GetKey("WFP5vVY1tNEedemEEre6CGgAa6U9w2xS9R", "sogosdfo-WFP5vVY1tNEedemEEre6CGgAa6U9w2xS9R.key", "123TestGCM")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	fmt.Println("seed: ", hex.EncodeToString(key.Seed()))
}
