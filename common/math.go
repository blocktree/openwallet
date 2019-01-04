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

import "math/big"

//WIP: 还没测试

func FloatStringToBigInt(x string) *big.Int {
	rst := new(big.Int)
	if _, valid := rst.SetString(x, 10); !valid {
		return nil
	}
	return rst
}

func FloatStringPointShift(x string, y int64) string {
	rst := new(big.Int)
	if _, valid := rst.SetString(x, 10); !valid {
		return "0"
	}
	rst = rst.Exp(big.NewInt(10), big.NewInt(y), nil)
	return rst.String()
}

func BigIntPointShift(x *big.Int, y int64) *big.Int {
	x = x.Exp(big.NewInt(10), big.NewInt(y), nil)
	return x
}

func FloatStringPointShiftToBigInt(x string, y int64) *big.Int {
	bigInt := FloatStringToBigInt(x)
	if bigInt != nil{
		bigInt = BigIntPointShift(bigInt, y)
	}
	return bigInt
}