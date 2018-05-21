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
	"regexp"
)

const (
	// 中国大陆手机号码正则匹配, 不是那么太精细
	// 只要是 13,14,15,17,18 开头的 11 位数字就认为是中国手机号
	chinaMobilePattern = `^1[34578][0-9]{9}$`
	// 用户昵称的正则匹配, 合法的字符有 0-9, A-Z, a-z, _, 汉字
	// 字符 '_' 只能出现在中间且不能重复, 如 "__"
	nicknamePattern = `^[a-z0-9A-Z\p{Han}]+(_[a-z0-9A-Z\p{Han}]+)*$`
	// 用户名的正则匹配, 合法的字符有 0-9, A-Z, a-z, _
	// 第一个字母不能为 _, 0-9
	// 最后一个字母不能为 _, 且 _ 不能连续
	usernamePattern = `^[a-zA-Z][a-z0-9A-Z]*(_[a-z0-9A-Z]+)*$`
	// 电子邮箱的正则匹配, 考虑到各个网站的 mail 要求不一样, 这里匹配比较宽松
	// 邮箱用户名可以包含 0-9, A-Z, a-z, -, _, .
	// 开头字母不能是 -, _, .
	// 结尾字母不能是 -, _, .
	// -, _, . 这三个连接字母任意两个不能连续, 如不能出现 --, __, .., -_, -., _.
	// 邮箱的域名可以包含 0-9, A-Z, a-z, -
	// 连接字符 - 只能出现在中间, 不能连续, 如不能 --
	// 支持多级域名, x@y.z, x@y.z.w, x@x.y.z.w.e
	mailPattern = `^[a-z0-9A-Z]+([\-_\.][a-z0-9A-Z]+)*@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)*\.)+[a-zA-Z]{2,4}$`
	//是否整数值
	numberPattern = `^[0-9]*$`
	//验证实数
	realNumberPattern = `^-?[0-9]+\.{0,1}[0-9]+$`
)

var (
	chinaMobileRegexp = regexp.MustCompile(chinaMobilePattern)
	nicknameRegexp    = regexp.MustCompile(nicknamePattern)
	usernameRegexp    = regexp.MustCompile(usernamePattern)
	mailRegexp        = regexp.MustCompile(mailPattern)
	numberRegexp        = regexp.MustCompile(numberPattern)
	realNumberRegexp        = regexp.MustCompile(realNumberPattern)
)

// 检验是否为合法的中国手机号, 不是那么太精细
// 只要是 13,14,15,18 开头的 11 位数字就认为是中国手机号
func IsChinaMobile(b []byte) bool {
	if len(b) != 11 {
		return false
	}
	return chinaMobileRegexp.Match(b)
}

// 同 func IsChinaMobile(b []byte) bool
func IsChinaMobileString(str string) bool {
	if len(str) != 11 {
		return false
	}
	return chinaMobileRegexp.MatchString(str)
}

// 检验是否为合法的昵称, 合法的字符有 0-9, A-Z, a-z, _, 汉字
// 字符 '_' 只能出现在中间且不能重复, 如 "__"
func IsNickname(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	return nicknameRegexp.Match(b)
}

// 同 func IsNickname(b []byte) bool
func IsNicknameString(str string) bool {
	if len(str) == 0 {
		return false
	}
	return nicknameRegexp.MatchString(str)
}

// 检验是否为合法的用户名, 合法的字符有 0-9, A-Z, a-z, _
// 第一个字母不能为 _, 0-9
// 最后一个字母不能为 _, 且 _ 不能连续
func IsUserName(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	return usernameRegexp.Match(b)
}

// 同 func IsName(b []byte) bool
func IsUserNameString(str string) bool {
	if len(str) == 0 {
		return false
	}
	return usernameRegexp.MatchString(str)
}

// 检验是否为合法的电子邮箱, 考虑到各个网站的 mail 要求不一样, 这里匹配比较宽松
// 邮箱用户名可以包含 0-9, A-Z, a-z, -, _, .
// 开头字母不能是 -, _, .
// 结尾字母不能是 -, _, .
// -, _, . 这三个连接字母任意两个不能连续, 如不能出现 --, __, .., -_, -., _.
// 邮箱的域名可以包含 0-9, A-Z, a-z, -
// 连接字符 - 只能出现在中间, 不能连续, 如不能 --
// 支持多级域名, x@y.z, x@y.z.w, x@x.y.z.w.e
func IsMail(b []byte) bool {
	if len(b) < 6 { // x@x.xx
		return false
	}
	return mailRegexp.Match(b)
}

// 同 func IsMail(b []byte) bool
func IsMailString(str string) bool {
	if len(str) < 6 { // x@x.xx
		return false
	}
	return mailRegexp.MatchString(str)
}

//IsNumber 是否整数值
func IsNumber(b []byte) bool {
	if len(b) == 0 { // x@x.xx
		return false
	}
	return numberRegexp.Match(b)
}

// 同 func IsNumber(b []byte) bool
func IsNumberString(str string) bool {
	if len(str) == 0 { // 1
		return false
	}
	return numberRegexp.MatchString(str)
}


//IsRealNumber 是否实数值
func IsRealNumber(b []byte) bool {
	if len(b) == 0 { // x@x.xx
		return false
	}
	return realNumberRegexp.Match(b)
}

// 同 func IsRealNumber(b []byte) bool
func IsRealNumberString(str string) bool {
	if len(str) == 0 { // 1
		return false
	}
	return realNumberRegexp.MatchString(str)
}

