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

package bitcoincash

import (
	"encoding/hex"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"testing"
)

func TestAddressDecoder_AddressToHash(t *testing.T) {
	hash, _ := hex.DecodeString("85537b4e0e2ecea63e4cec6573c319bf7925192c")

	cfg := addressEncoder.LTC_testnetAddressP2SH2

	addr :=addressEncoder.AddressEncode(hash, cfg)
	t.Logf("addr: %s", addr)
}

func TestAddressDecoder_HashToAddress(t *testing.T) {
	addr := "2N5QBsCqE6NvEUswatz7awNUMQMXxiePPF6"

	cfg := addressEncoder.LTC_testnetAddressP2SH

	hash, err :=addressEncoder.AddressDecode(addr, cfg)
	if err != nil {
		t.Errorf("AddressDecode failed unexpected error: %v\n", err)
		return
	}
	t.Logf("hash: %s", hex.EncodeToString(hash))
}