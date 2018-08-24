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
	"log"

	"github.com/bndr/gotabulate"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

//getWalletList 获取钱包列表
func (wm *WalletManager) getWalletList() ([]*Wallet, error) {
	var wallets = make([]*Wallet, 0)

	r, err := wm.fullnodeClient.Call("account", "GET", nil)
	if err != nil {
		return nil, err
	}
	data := gjson.ParseBytes(r).Map()

	for wid, a := range data {
		addr := a.String()
		w := &Wallet{Alias: wid, Addr: addr}
		wallets = append(wallets, w)
	}

	return wallets, nil
}

//CreateNewWallet 创建钱包
func (wm *WalletManager) createWallet(wid string) (*Wallet, error) {
	var wallet *Wallet

	if _, err := wm.fullnodeClient.Call("account", "POST", req.Param{"id": wid}); err != nil {
		return nil, err
	} else {
		if w, err := wm.getWalletInfo(wid); err != nil {
			wallet = &Wallet{}
		} else {
			wallet = w
		}
	}

	return wallet, nil
}

// -----------------------------------------------------------------------------
// Get one wallet info
func (wm *WalletManager) getWalletInfo(wid string) (*Wallet, error) {

	if r, err := wm.fullnodeClient.Call(fmt.Sprintf("account/%s", wid), "GET", nil); err != nil {
		return nil, err
	} else {
		data := gjson.ParseBytes(r).Map()
		return &Wallet{Alias: wid, Addr: data["address"].String()}, nil
	}
}

// 获取钱包信息
func (wm *WalletManager) getWalletB(addr string) (wallet *Wallet, err error) {

	// Get balance
	if d, err := wm.fullnodeClient.Call(fmt.Sprintf("chain/%s", addr), "GET", nil); err != nil {
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

// 打印钱包列表
func (wm *WalletManager) printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		if ww, err := wm.getWalletB(w.Addr); err == nil {
			bal := ww.Balance
			if bal != "" {
				cc, _ := decimal.NewFromString(bal)
				bal = cc.Div(wm.config.coinDecimal).String()
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
