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

import (
	"fmt"
	"testing"
)

// 测试明文AES加密
func TestAES(t *testing.T) {
	var (
		// appkey     = "2d68067484a20f1a346b3cf28a898ed7f5736f5bacf0fe60449da95efdb97ad4"
		appsecret  = "0dd1e322907ad7f55deaa35fec2aac97cae7931454d734364bc63f3e9b9f993a"
		planttext  = String("考几分就；辣椒粉李经理发吉林省；发；快递放假；介绍费sdfaf")
		ciphertext string
		err        error
	)

	if ciphertext, err = planttext.AES(appsecret); err != nil {
		fmt.Println("err =", err.Error())
	}

	fmt.Println("ciphertext = ", ciphertext)
}

//测试AES密文解密
func TestUnAES(t *testing.T) {

	var (
		// appkey     = "2d68067484a20f1a346b3cf28a898ed7f5736f5bacf0fe60449da95efdb97ad4"
		appsecret  = "0dd1e322907ad7f55deaa35fec2aac97cae7931454d734364bc63f3e9b9f993a"
		planttext  = new(String)
		ciphertext = "Pr5nzcrcWtoJBjx3JrmeUwgftz3tWrpM8+/9BCduLEqcZBjNslDDiqWbvI+hpXMEZNsjorj3zpX2Sbolj/gpoLUqsXHN2UZmrdDCZY2M3PM="
		err        error
		// result     interface{}
	)

	err = planttext.UnAES(ciphertext, appsecret)

	fmt.Println("planttext = ", planttext)
	if err != nil {
		fmt.Println("err = ", err.Error())
	}

}

// 测试各种基本类型转为String
func TestTypeToString(t *testing.T) {
	var (
		a  = int8(10)
		b  = int32(100)
		c  = int64(1000)
		d  = "sdfsf"
		f  = 100.2
		g  = 9999
		h  = false
		i  = []string{"12345", "asdfg"}
		j  = map[string]interface{}{"a": "12345", "b": "asdfg"}
		k = struct {
			Name string
			Age uint64
		}{Name: "john", Age: 90}
		kv = make(map[string]interface{}, 0)
	)

	kv["a"] = a
	kv["b"] = b
	kv["c"] = c
	kv["d"] = d
	kv["f"] = f
	kv["g"] = g
	kv["h"] = h
	kv["i"] = i
	kv["j"] = j
	kv["k"] = k

	for k, v := range kv {
		newString := NewString(v)
		fmt.Println(k, ":", newString)
	}

}
