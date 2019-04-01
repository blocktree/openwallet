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
	ErrCreateRawTransactionFailed   = 2005 //创建原始交易单失败
	ErrSignRawTransactionFailed     = 2006 //签名原始交易单失败
	ErrVerifyRawTransactionFailed   = 2007 //验证原始交易单失败
	ErrSubmitRawTransactionFailed   = 2008 //广播原始交易单失败

	/* 账户类别 */
	ErrAccountNotFound    = 3001 //账户不存在
	ErrAddressNotFound    = 3002 //地址不存在
	ErrContractNotFound   = 3003 //合约不存在
	ErrAdressEncodeFailed = 3004 //地址编码失败
	ErrAdressDecodeFailed = 3006 //地址解码失败
	ErrNonceInvaild       = 3007 //Nonce不正确
	ErrAccountNotAddress  = 3008 //账户没有地址

	/* 网络类型 */
	ErrCallFullNodeAPIFailed = 4001 //全节点API无法访问

	/* 其他 */
	ErrUnknownException = 9001 //未知异常情况
)

type Error struct {
	Code int64
	Err  string
}

//Error 错误信息
func (err *Error) Error() string {
	return fmt.Sprintf("[%d]%s", err.Code, err.Err)
}

//ConvertError error转OWError
func ConvertError(err error) *Error {
	owerr, ok := err.(*Error)
	if !ok {
		return &Error{Code: ErrUnknownException, Err: err.Error()}
	}
	return owerr
}

//Errorf 生成OWError
func Errorf(code int64, format string, a ...interface{}) error {
	err := &Error{
		Code: code,
		Err:  fmt.Sprintf(format, a...),
	}
	return err
}

//NewError 生成OWError
func NewError(code int64, text string) error {
	err := &Error{
		Code: code,
		Err:  text,
	}
	return err
}
