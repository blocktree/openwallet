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

package log

import (
	"github.com/astaxie/beego/logs"
	"testing"
)

func TestInfo(t *testing.T) {
	Debug("hello","world")
}

func TestCustomLogger(t *testing.T) {
	l := logs.GetLogger("BTC")
	l.Println("hello", "world")
}

func TestOWLogger(t *testing.T) {
	l := NewOWLogger("BTC")
	l.Info("hello", "world")
}