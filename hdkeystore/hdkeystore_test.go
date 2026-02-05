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

// 重点：强烈建议不再使用该CTR模式，不再安全
func TestStoreHDKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	key, rootId, err := StoreHDKey(path, "sogosdfo789", "123Test", StandardScryptN, StandardScryptP)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
	fmt.Println("seed:", hex.EncodeToString(key.seed))
}

func TestStoreHDKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	key, rootId, err := StoreHDKey(path, "sogosdfo456", "123TestGCM", StandardScryptN, StandardScryptP, CipherAes256GCM)
	if err != nil {
		t.Errorf("StoreHDKey failed unexpected error: %v", err)
	} else {
		t.Logf("StoreHDKey root id = %s", rootId)
	}
	fmt.Println("seed:", hex.EncodeToString(key.seed)) // 最终显示应该是清空的种子
}

func TestGetKey(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP, CipherAes128CTR}

	key, err := ks.GetKey("Vyhrnx73eeTtj6ZyjBAZVLupHML81x3FAZ", "sogosdfo789-Vyhrnx73eeTtj6ZyjBAZVLupHML81x3FAZ.key", "123Test")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	fmt.Println("seed:", hex.EncodeToString(key.Seed()))
}

func TestGetKeyGCM(t *testing.T) {
	path := filepath.Join(".", "keys")
	ks := &HDKeystore{path, StandardScryptN, StandardScryptP, CipherAes256GCM}

	key, err := ks.GetKey("W69vJsxzAQwztBLtaRno3FKZZw1zMvsH93", "sogosdfo456-W69vJsxzAQwztBLtaRno3FKZZw1zMvsH93.key", "123TestGCM")

	if err != nil {
		t.Errorf("GetKey failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetKey root id = %s", key.KeyID)
	}
	fmt.Println("seed: ", hex.EncodeToString(key.Seed()))
}
