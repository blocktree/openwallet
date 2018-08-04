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
	// "fmt"
	"github.com/tidwall/gjson"
	// "github.com/pkg/errors"
	// "log"
	"github.com/blocktree/OpenWallet/walletnode"
)

func Backup(symbol string) error {
	Cname := symbol.ToLower()
	src := "/usr/local/paicode/data/wallet.dat"
	dst := fmt.Fprintf("./data/%s/key/", symbol)
	if err := walletnode.CopyFromContainer(c, Cname, src, dst); err != nil {
		return err
	}
	return nil
}

func Restore() error {
	if err := walletnode.CopyToContainer(); err != nil {
		return err
	}
	return nil
}
