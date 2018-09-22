/*
 * Copyright 2018 The OpenSubcribe Authors
 * This file is part of the OpenSubcribe library.
 *
 * The OpenSubcribe library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenSubcribe library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet

type SubcribeDBFile WrapperSourceFile

type SubcribeKeyFile string

// SubcribeWrapper 钱包包装器，扩展钱包功能
type SubcribeWrapper struct {
	*AppWrapper
}

func NewSubcribeWrapper(args ...interface{}) *SubcribeWrapper {

	wrapper := NewAppWrapper(args...)

	walletWrapper := SubcribeWrapper{AppWrapper: wrapper}

	for _, arg := range args {
		switch obj := arg.(type) {
		case SubcribeDBFile:
			walletWrapper.sourceFile = string(obj)
		}
	}

	return &walletWrapper
}

