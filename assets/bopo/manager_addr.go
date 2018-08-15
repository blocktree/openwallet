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

package bopo

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	//"log"
)

// -----------------------------------------------------------------------------
// Verify address
func verifyAddr(addr string) error {
	if r, err := client.Call(fmt.Sprintf("account/verify/%s", addr), "GET", nil); err != nil {
		return err
	} else {
		if data := gjson.GetBytes(r, "verifyState").Bool(); data != true {
			return errors.New("Address invalid!")
		}
	}
	return nil
}
