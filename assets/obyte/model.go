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

package obyte

import "github.com/tidwall/gjson"

type Address struct {
	Address string
}

func NewAddress(r gjson.Result) *Address {
	obj := &Address{}
	obj.Address = r.String()
	return obj
}

type Balance struct {
	Stable  string
	Pending string
}

func NewBalance(r gjson.Result) *Balance {
	obj := &Balance{}
	obj.Stable = r.Get("stable").String()
	obj.Pending = r.Get("pending").String()
	return obj
}
