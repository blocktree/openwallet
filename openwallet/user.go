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

type User struct {
	//应用
	App *App
	//名字
	Name string
	//用户key，私链唯一
	UserKey string
	//根公钥，由用户密码或短语构成的私钥推导，可以更新
	//PublicKey string

}

// NewUser 实例化用户
func NewUser(userKey, name string, app *App) *User {

	user := &User{
		app,
		name,
		userKey,
	}

	return user
}

//GetUsersFromKeydir 从keystore路径中加载客户端用户，使用命令行工具
//func GetUsersFromKeydir(keydir string, app *App) []*User {
//
//}

// NewUserByFile 通过keystore文件加载用户
//func NewUserByKeyFile(filename string, app *App) *User {
//
//	// Load the key from the keystore and decrypt its contents
//	keyjson, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return nil, err
//	}
//	key, err := DecryptKey(keyjson, auth)
//	if err != nil {
//		return nil, err
//	}
//	// Make sure we're really operating on the requested key (no swap attacks)
//	if key.Address != addr {
//		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
//	}
//	return key, nil
//
//	user := &User{
//		app,
//		"",
//		userKey,
//		NewUserAccount(userKey, keydir),
//	}
//
//	return user
//}