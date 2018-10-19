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

package openwallet

type SymbolInfo interface {

	//CurveType 曲线类型
	CurveType() uint32

	//FullName 币种全名
	FullName() string

	//Symbol 币种标识
	Symbol() string

	//小数位精度
	Decimal() int32
}

type SymbolInfoBase struct {
}

//CurveType 曲线类型
func (s *SymbolInfoBase) CurveType() uint32 {
	return 0
}

//FullName 币种全名
func (s *SymbolInfoBase) FullName() string {
	return ""
}

//Symbol 币种标识
func (s *SymbolInfoBase) Symbol() string {
	return ""
}

//小数位精度
func (s *SymbolInfoBase) Decimal() int32 {
	return 0
}
