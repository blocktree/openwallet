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

const (
	// VERSION represent openw web framework version.
	VERSION = "1.0.0"
	// DEV is for develop
	DEV = "dev"
	// PROD is for production
	PROD = "prod"
)

//hook function to run
type hookfunc func() error

var (
	hooks = make([]hookfunc, 0) //hook function slice to store the hookfunc
)

// AddAPPStartHook is used to register the hookfunc
// The hookfuncs will run in beego.Run()
// such as sessionInit, middlerware start, buildtemplate, admin start
func AddAPPStartHook(hf hookfunc) {
	hooks = append(hooks, hf)
}

//Run 运行应用
func Run(params ...string) {
	OpenWalletApp.Run()
}
