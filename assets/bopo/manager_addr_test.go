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
	"testing"
)

var ()

func init() {
	serverAPI = "http://192.168.2.194:17280"
	isTestNet = false
	client = &Client{
		BaseURL: serverAPI,
		Debug:   true,
	}
}

func TestVerifyAddr(t *testing.T) {
	// Invalid addr, return err
	if err := verifyAddr("failureaddr"); err == nil {
		t.Errorf("TestVerifyAddr: %v\n", err)
	}
	// Verified addr, return nil
	if err := verifyAddr("5SJrzpTvUjMoTi2KvjM9pKvrh04_LoTYwg"); err != nil {
		t.Errorf("TestVerifyAddr: %v\n", err)
	}
}
