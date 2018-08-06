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
	"github.com/bndr/gotabulate"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
)

// 获取钱包信息
func getWalletB(addr string) (wallet *Wallet, err error) {

	// Get balance
	if d, err := client.Call(fmt.Sprintf("chain/%s", addr), "GET", nil); err != nil {
		// panic(err)
		return nil, err
	} else {
		if status, ok := gjson.ParseBytes(d).Map()["status"]; ok != true || status.String() != "ok" {
			log.Println("Bopo return data with 'status!=ok'!")
			return nil, errors.New("Bopo return data with 'status!=ok'!")
		}

		if data, ok := gjson.ParseBytes(d).Map()["data"]; !ok {
			return nil, nil
		} else {
			emp := data.Map()

			wallet = &Wallet{
				// Alias:
				Addr:    addr,
				Balance: emp["pais"].String(),
			}
		}
	}

	return wallet, nil
}

// // Insert inserts the value into the slice at the specified index,
// // which must be in range.
// // The slice must have room for the new element.
// func Insert(slice []string, index int, value string) []string {
// 	// Grow the slice by one element.
// 	slice = slice[0 : len(slice)+1]
// 	// Use copy to move the upper part of the slice out of the way and open a hole.
// 	copy(slice[index+1:], slice[index:])
// 	// Store the new value.
// 	slice[index] = value
// 	// Return the result.
// 	return slice
// }

// 打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		if ww, err := getWalletB(w.Addr); err == nil {
			bal := ww.Balance
			if bal != "" {
				cc, _ := decimal.NewFromString(bal)
				bal = cc.Div(coinDecimal).String()
				w.Balance = fmt.Sprintf("%s (%s coins)", ww.Balance, bal)
			}
		}
		tableInfo = append(tableInfo, []interface{}{
			i + 1, w.WalletID, w.Alias, w.Addr, w.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Alias", "Addr", "Balance(1 coin=10^8 pais)"})

	//打印信息
	fmt.Println(t.Render("simple"))
}
