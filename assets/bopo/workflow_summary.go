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
	"io/ioutil"
	"log"
	"time"
)

func summaryWallets() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05:000"))

	if err := ioutil.WriteFile("/tmp/abcd.txt", []byte("AAAA"), 0600); err != nil {
		log.Println(err)
	}
}
