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

package owtp

import (
	"encoding/json"
	"fmt"
	"github.com/blocktree/OpenWallet/session"
	"testing"
)


func TestMem(t *testing.T) {
	config := `{"gclifetime":10}`
	conf := new(session.ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	globalSessions, _ := NewSessionManager("memory", conf)
	go globalSessions.GC()

	err := globalSessions.Put("123456", "username", "owtp")
	if err != nil {
		t.Fatal("set error,", err)
	}
	username := globalSessions.Get("123456", "username")
	if username != "owtp" {
		t.Fatal("get username error")
	}

	err = globalSessions.Put("qwer", "username", "good")
	if err != nil {
		t.Fatal("set error,", err)
	}
	username2 := globalSessions.Get("qwer", "username")
	if username2 != "good" {
		t.Fatal("get username error")
	}

	fmt.Printf("username = %s \n", username)
	fmt.Printf("username2 = %s \n", username2)
}
