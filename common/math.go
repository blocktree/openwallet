/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package common

import (
	"github.com/shopspring/decimal"
	"math/big"
)

func StringNumToBigIntWithExp(amount string, exp int32) *big.Int {
	vDecimal, _ := decimal.NewFromString(amount)
	vDecimal = vDecimal.Shift(exp)
	bigInt, ok := new(big.Int).SetString(vDecimal.String(), 10)
	if !ok {
		return big.NewInt(0)
	}
	return bigInt
}


func IntToDecimals(amount int64, decimals int32) decimal.Decimal {
	return decimal.New(amount, 0).Shift(-decimals)
}

func BigIntToDecimals(amount *big.Int, decimals int32) decimal.Decimal {
	if amount == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(amount, 0).Shift(-decimals)
}