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

package bitcoin

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

func GetAddressesByAccount(account string) ([]string, error) {

	var (
		err error
	)

	request := struct {
		Account string `json:"account"`
	}{account}

	result := client.Call("getaddressesbyaccount", request)

	err = isError(result)

	addresses, ok := gjson.GetBytes(result, "result").Value().([]string)
	if !ok {
		return nil, errors.New("no array")
	}

	return addresses, err

}

//isError 是否报错
func isError(result []byte) error {
	var (
		err error
	)

	/*
		//failed 返回错误
		{
			"result": null,
			"error": {
				"code": -8,
				"message": "Block height out of range"
			},
			"id": "foo"
		}
	*/

	if !gjson.GetBytes(result, "error").Exists() {
		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		gjson.GetBytes(result, "error.code").Int(),
		gjson.GetBytes(result, "error.message").String())
	err = errors.New(errInfo)

	return err
}
