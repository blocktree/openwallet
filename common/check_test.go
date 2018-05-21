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

import "testing"

func TestIsRealNumberString(t *testing.T) {

	tests := []struct {
		text string
		tag string
	}{
		{
			text: "-90000000",
			tag: "zheng shu",
		},
		{
			text: "-1212",
			tag: "zheng shu",
		},
		{
			text: "1.232",
			tag: "xiao shu",
		},
		{
			text: "-1.232",
			tag: "xiao shu",
		},
		{
			text: "dfsfsdfsdf",
			tag: "charactor",
		},
	}

	for i, test := range tests {

		t.Logf("TestIsRealNumber[%d] = %v", i, IsRealNumberString(test.text))

	}

}
