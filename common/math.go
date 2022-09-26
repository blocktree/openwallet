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
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
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

func StringValueToBigInt(value string, base int) (*big.Int, error) {
	bigvalue := new(big.Int)
	var success bool

	if value == "" {
		value = "0"
	}
	value = strings.TrimPrefix(value, "0x")

	_, success = bigvalue.SetString(value, base)
	if !success {
		return big.NewInt(0), fmt.Errorf("convert value [%v] to bigint failed, check the value and base passed through", value)
	}
	return bigvalue, nil
}

func BytesToDecimals(bits []byte, decimals int32) decimal.Decimal {
	if bits == nil {
		return decimal.Zero
	}
	amount := new(big.Int)
	amount.SetBytes(bits)
	return decimal.NewFromBigInt(amount, 0).Shift(-decimals)
}
