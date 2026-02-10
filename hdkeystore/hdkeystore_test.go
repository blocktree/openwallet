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
	"github.com/awnumar/memguard"
	"path/filepath"
	"testing"
)

func TestStoreHDKeyGCM(t *testing.T) {
	auth := memguard.NewBufferFromBytes([]byte("123TestGCM"))
	defer auth.Destroy()
	path := filepath.Join(".", "keys")
	rootID, err := StoreLockerHDKey(path, "sogosdfo456", auth)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootID)
	}
}

func TestGetKeyGCM(t *testing.T) {
	alias := "sogosdfo456"
	rootID := "W4HPZzRupuc5aRw1odC7QcYT5D1s8JRY9z"
	ks := &HDKeystore{}
	path := ks.JoinDirPath(filepath.Join(".", "keys"), fmt.Sprintf("%s-%s.key", alias, rootID))
	aad, _ := GetRandomSecure(32)
	aadCall := func(keyID string) ([]byte, error) {
		return aad, nil
	}
	auth := memguard.NewBufferFromBytes([]byte("123TestGCM"))
	defer auth.Destroy()
	key, err := ks.GetLockerKey(path, auth, aadCall)
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
