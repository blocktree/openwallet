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
	"log"
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
			openwLogger.Log.Errorf("unexpect error: %v", err)
			return "", err
		}

		if len(password) < 8 {
			log.Printf("不合法的密码长度, 建议设置不小于8位的密码, 请重新输入")
			continue
		}

		// 二次确认密码
		if isConfirm {

			confirm, err = Stdin.PromptPassword("再次确认钱包密码: ")

			if password != confirm {
				log.Printf("两次输入密码不一致, 请重新输入")
				continue
			}

		}

		break
	}

	return password, nil
}
