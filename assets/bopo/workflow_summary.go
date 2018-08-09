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
	"time"
)

func summaryWallets() {
	fmt.Printf("%s -> ", time.Now().Format("2006-01-02 15:04:05:000"))

	// List all wallets that have balance to summary (without summaryAddr)
	wallets, err := getWalletList()
	if err != nil {
		log.Println(err)
	}

	tmp := wallets[:0]
	for _, w := range wallets {
		if w.Balance != "" && w.Addr != sumAddress {
			tmp = append(tmp, w)
		}
	}
	wallets = tmp
	fmt.Printf("There are %d wallets will be summary······", len(wallets))
	// if len(wallets) > 0 {
	// 	printWalletList(wallets)		// 加上后，会转账失败！原因未知？留作提醒
	// }

	// Summary
	for _, w := range wallets {
		message := time.Now().UTC().Format(time.RFC850)

		if _, err := toTransfer(w.WalletID, sumAddress, w.Balance, message); err != nil {
			log.Println(err)
		}
	}
	time.Sleep(time.Second * 1)
	fmt.Printf("Done\n")

}
