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

package tron

import "fmt"

func (wm *WalletManager) GetWitnesses() ([]string, error) {

	var (
		addresses = make([]string, 0)
	)

	request := []interface{}{
		// walletID,
	}

	res, err := wm.WalletClient.Call("/wallet/listwitnesses", request)
	if err != nil {
		return nil, err
	}

	array := res.Get("witnesses").Array()
	for _, a := range array {
		fmt.Println("EEEEE = ", a)
		addresses = append(addresses, a.String())
	}

	return addresses, nil

}
