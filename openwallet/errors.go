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

package openwallet

import "fmt"

const (
	/* 交易类别 */
	ErrInsufficientBalanceOfAccount = 2001 //账户余额不足
	ErrInsufficientBalanceOfAddress = 2002 //地址余额不足
	ErrInsufficientFees             = 2003 //手续费不足
	ErrDustLimit                    = 2004 //限制粉尘攻击
)

type OWError struct {
	Code int64
	Err  string
}

func (err *OWError) Error() string {
	return fmt.Sprintf("[%d]%s", err.Code, err.Err)
}
