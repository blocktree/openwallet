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

package common

import (
	"testing"
)


func TestStringNumToBigIntWithDecimals(t *testing.T) {
	num := StringNumToBigIntWithExp("3.223", 8)
	t.Logf("num : %v\n", num)
	t.Logf("num int64 : %v\n", num.Int64())
	newnum := BigIntToDecimals(num, 8)
	t.Logf("newnum : %v\n", newnum.String())
}