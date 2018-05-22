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

package console

import (
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/common"
	"fmt"
)

//PasswordPrompt 提示输入密码
//@param 是否二次确认
func InputPassword(isConfirm bool) (string, error) {

	var (
		confirm  string
		password string
		err      error
	)

	for {

		// 等待用户输入密码
		password, err = Stdin.PromptPassword("输入钱包密码: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpect error: %v\n", err)
			return "", err
		}

		if len(password) < 8 {
			fmt.Printf("不合法的密码长度, 建议设置不小于8位的密码, 请重新输入\n")
			continue
		}

		// 二次确认密码
		if isConfirm {

			confirm, err = Stdin.PromptPassword("再次确认钱包密码: ")

			if password != confirm {
				fmt.Printf("两次输入密码不一致, 请重新输入\n")
				continue
			}

		}

		break
	}

	return password, nil
}

//InputText 输入文本
func InputText(prompt string, required bool) (string, error) {

	var (
		text  string
		err      error
	)

	for {

		// 等待用户输入
		text, err = Stdin.PromptInput(prompt)
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v\n", err)
			return "", err
		}

		if len(text) == 0 && required {
			fmt.Printf("内容不能为空\n")
			continue
		}

		break
	}

	return text, nil
}


//InputNumber 输入数值
func InputNumber(prompt string) (uint64, error) {

	var (
		num  uint64
	)

	for {
		// 等待用户输入参数
		line, err := Stdin.PromptInput(prompt)
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v\n", err)
			return 0, err
		}
		num = common.NewString(line).UInt64()

		if num <= 0 {
			fmt.Printf("内容必须数字，而且必须大于0\n")
			continue
		}

		break
	}

	return num, nil
}


//InputRealNumber 输入实数值
func InputRealNumber(prompt string, p bool) (float64, error) {

	var (
		num  float64
	)

	for {
		// 等待用户输入参数
		line, err := Stdin.PromptInput(prompt)
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v\n", err)
			return 0, err
		}
		num = common.NewString(line).Float64()

		if p && num <= 0 {
			fmt.Printf("内容必须数字，而且必须大于0\n")
			continue
		}

		break
	}

	return num, nil
}