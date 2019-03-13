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

package nebulasio

import (
	"testing"
)


var decoder *addressDecoder


func TestPrivateKeyToWIF(t *testing.T) {

	pri := []byte{0x57,0xa5,0x27,0xaa,0x5e,0xc8,0xdd,0xa9,0x6e,0x4c,0xbe,0x16,0x29,0xcf,0xf1,0x08,0x6d,0x80,0xd4,0xd3,0xaf,0x67,0x14,0x58,0x97,0x45,0x39,0x64,0x16,0xce,0x7f,0xfd}

	WIF,err := decoder.PrivateKeyToWIF(pri, false)
	if err != nil {
		t.Errorf("PrivateKeyToWIF failed unexpected error: %v\n", err)
		return
	}
	t.Logf("WIF: %s", WIF)
}

func TestPublicKeyToAddress(t *testing.T) {

	pub := []byte{0x02,0x43,0xe7,0x98,0x6f,0x5a,0xb0,0xcd,0xdb,0x95,0xb2,0xf9,0x4a,0x12,0x5e,0x8a,0x67,0x0e,0x49,0x94,0x0f,0xa7,0xf9,0xff,0x65,0x7c,0xb0,0xbb,0xaf,0x15,0xcb,0x3c,0x57}

	addr,err := decoder.PublicKeyToAddress(pub,false)
	if err != nil {
		t.Errorf("PublicKeyToAddress failed unexpected error: %v\n", err)
		return
	}
	t.Logf("addr: %s", addr)
}

func TestWIFToPrivateKey(t *testing.T) {

	Wif := ""

	pri,err := decoder.WIFToPrivateKey(Wif,false)
	if err != nil {
		t.Errorf("WIFToPrivateKey failed unexpected error: %v\n", err)
		return
	}
	t.Logf("pri: %x", pri)
}