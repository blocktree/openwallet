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

package tron

import (
	"github.com/blocktree/openwallet/log"
	"testing"
)

func TestEncodeAddress(t *testing.T) {
	addr, err := EncodeAddress("412b30841b1705b302b6c431182670e53d32c7b084", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("addr: %s", addr)
}
