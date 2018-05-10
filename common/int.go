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

package common

import "strconv"

/***************** 数值类型转字符串 *****************/

//Int 强化int类型
type Int int

//IntProtocal 强化int的扩展方法
type IntProtocal interface {
	Int(def ...int) int
	Int8(def ...int8) int8
	Int16(def ...int16) int16
	Int32(def ...int32) int32
	Int64(def ...int64) int64
}

// ToString int转为string类型
func (v Int) String() string {
	string := strconv.FormatInt(int64(v), 10)
	return string
}

