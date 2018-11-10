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

package openwallet

import "testing"

func TestGenContractID(t *testing.T) {
	address := "f2033ede578e17fa6231047265010445bca8cf1c"
	symbol := "qtum"
	ID := GenContractID(symbol, address)
	t.Logf("ID 1: %v", ID)
	ID = GenContractID(symbol, address)
	t.Logf("ID 2: %v", ID)
}
