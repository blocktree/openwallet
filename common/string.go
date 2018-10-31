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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/blocktree/OpenWallet/crypto"
)

// String AT的字符串库
type String string

const (
	// NilString 定义空串的关键字
	NilString String = "<nil>"
)

//StringProtocal 强化String的扩展方法
type StringProtocal interface {
	String() string
	Length() int
	IsNil() bool
	IsEmpty() bool
}

// String 转string
func (s String) String() string {
	return string(s)
}

// Length 字符串长度，支持中文字等等的计算
func (s String) Length() int {
	length := len([]rune(s.String()))
	return length
}

// IsNil 判断是否空指针
func (s String) IsNil() bool {

	if s == NilString {
		return true
	}

	return false
}

// IsEmpty 是否空值
func (s String) IsEmpty() bool {

	if len(s) == 0 {
		return true
	}

	return false
}

// Int String类型转为int类型
func (s String) Int(def ...int) int {
	var (
		value int
		err   error
	)
	if value, err = strconv.Atoi(s.String()); err != nil {
		if len(def) > 0 {
			value = def[0]
		}
	}
	return value
}

// UInt8 String类型转为uint8类型
func (s String) UInt8(def ...int) uint8 {
	value := s.UInt64(def...)
	return uint8(value)
}

// UInt16 String类型转为uint32类型
func (s String) UInt16(def ...int) uint16 {
	value := s.UInt64(def...)
	return uint16(value)
}

// UInt32 String类型转为uint32类型
func (s String) UInt32(def ...int) uint32 {
	value := s.UInt64(def...)
	return uint32(value)
}

// UInt64 String类型转为uint64类型
func (s String) UInt64(def ...int) uint64 {
	var (
		value uint64
		err   error
	)
	if value, err = strconv.ParseUint(s.String(), 10, 64); err != nil {
		if len(def) > 0 {
			value = uint64(def[0])
		}
	}
	return value
}

// Int8 String类型转为int8类型
func (s String) Int8(def ...int) int8 {
	value := s.Int64(def...)
	return int8(value)
}

// Int16 String类型转为int32类型
func (s String) Int16(def ...int) int16 {
	value := s.Int64(def...)
	return int16(value)
}

// Int32 String类型转为int32类型
func (s String) Int32(def ...int) int32 {
	value := s.Int64(def...)
	return int32(value)
}

// Int64 String类型转为int64类型
func (s String) Int64(def ...int) int64 {
	var (
		value int64
		err   error
	)
	if value, err = strconv.ParseInt(s.String(), 10, 64); err != nil {
		if len(def) > 0 {
			value = int64(def[0])
		}
	}
	return value
}

// Bool string转布尔型
func (s String) Bool(def ...bool) bool {
	var value = false
	if s.Int() == 1 {
		value = true
	} else if s.Int() == 0 {
		value = false
	} else {
		if len(def) > 0 {
			value = def[0]
		}
	}
	return value
}

// Float32 String转为float32
func (s String) Float32(def ...float32) float32 {
	value := s.Float64()
	return float32(value)
}

// Float64 String转为float64
func (s String) Float64(def ...float64) float64 {
	var (
		value float64
		err   error
	)
	if value, err = strconv.ParseFloat(s.String(), 32); err != nil {
		if len(def) > 0 {
			value = def[0]
		}
	}
	return value
}

// AES AES加密
// key 密钥key hex字符串
// return 密文
func (s String) AES(key string) (string, error) {
	var (
		plantext   []byte
		keybyte    []byte
		err        error
		ciphertext []byte
		result     string
	)
	plantext = []byte(s.String())
	if keybyte, err = hex.DecodeString(key); err != nil {
		return "", err
	}

	if ciphertext, err = crypto.AESEncrypt(plantext, keybyte); err != nil {
		return "", err
	}

	//转为base64
	result = base64.StdEncoding.EncodeToString(ciphertext)

	return result, err
}

// UnAES 通过base64编码的aes密文初始化一个字符串
// aesBase64string base64编码的aes密文
// key 密钥字符串
// return 明文
func (s *String) UnAES(aesBase64string string, key string) error {

	var (
		plantext   []byte
		keybyte    []byte
		err        error
		ciphertext []byte
		result     String
	)

	if keybyte, err = hex.DecodeString(key); err != nil {
		return err
	}

	if ciphertext, err = base64.StdEncoding.DecodeString(aesBase64string); err != nil {
		return err
	}

	if plantext, err = crypto.AESDecrypt(ciphertext, keybyte); err != nil {
		return err
	}

	result = String(plantext)
	*s = result

	return nil
}

// NewStringByInt 通过int初始化字符串
func NewStringByInt(v int64) String {
	str := strconv.FormatInt(v, 10)
	return String(str)
}

// NewStringByUInt 通过int初始化字符串
func NewStringByUInt(v uint64) String {
	str := strconv.FormatUint(v, 10)
	return String(str)
}

// NewStringByBool 通过bool初始化字符串
func NewStringByBool(v bool) String {
	str := strconv.FormatBool(v)
	return String(str)
}

// NewStringByFloat 通过float初始化字符串
func NewStringByFloat(v float64) String {
	str := strconv.FormatFloat(v, 'f', -1, 64)
	return String(str)
}

// NewString 初始化字符串，自动转型
func NewString(value interface{}, def ...String) String {

	val := reflect.ValueOf(value) //读取变量的值，可能是指针或值

	switch val.Kind() {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewStringByInt(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewStringByUInt(val.Uint())
	case reflect.Float32, reflect.Float64:
		return NewStringByFloat(val.Float())
	case reflect.Bool:
		return NewStringByBool(val.Bool())
	case reflect.String:
		return String(val.String())
	case reflect.Slice, reflect.Array, reflect.Map: //如果类型为数组，要继续迭代元素
		jsonstr, _ := json.Marshal(value)
		return String(jsonstr)
	default:
		if len(def) > 0 {
			return String(def[0])
		}

	}
	return ""
}

// MD5 字符串转为MD5后的hash hex
func (s String) MD5() string {
	hash := crypto.MD5([]byte(s))
	mdStr := hex.EncodeToString(hash)
	return mdStr
}

// SHA1 字符串转为SHA1后的hash hex
func (s String) SHA1() string {
	hash := crypto.SHA1([]byte(s))
	mdStr := hex.EncodeToString(hash)
	return mdStr
}

// SHA256 字符串转为SHA256后的hash hex
func (s String) SHA256() string {
	hash := crypto.SHA256([]byte(s))
	mdStr := hex.EncodeToString(hash)
	return mdStr
}

// HmacSHA1 字符串转为HmacSHA1后的hash hex
func (s String) HmacSHA1(secret string) string {
	hash := crypto.HmacSHA1(secret, []byte(s))
	mdStr := hex.EncodeToString(hash)
	return mdStr
}

// HmacMD5 字符串转为HmacMD5后的hash md5
func (s String) HmacMD5(secret string) string {
	hash := crypto.HmacMD5(secret, []byte(s))
	mdStr := hex.EncodeToString(hash)
	return mdStr
}

func Substr(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < 0 || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end:end])
}
