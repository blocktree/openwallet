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

import (
	"net/http"
)

const (
	defaultAppName = "openw"
)

var (
	// BeeApp is an application instance
	OpenWalletApp *App
)

func init() {
	// create beego application
	OpenWalletApp = NewApp()
}

// App defines beego application with a new PatternServeMux.
type App struct {
	Handlers *AssetsRegister
	Server   *http.Server
}

// NewApp returns a new beego application.
func NewApp() *App {
	cr := NewAssetsRegister()
	app := &App{Handlers: cr, Server: &http.Server{}}
	return app
}

// Run beego application.
func (app *App) Run() {

	var (
		//err        error
		//l          net.Listener
		endRunning = make(chan bool, 1)
	)
	<-endRunning
}

// Router adds a patterned controller handler to OpenWallet.
// it's an alias method of App.Router.
// usage:
//  simple router
//  openw.Router("ethereum", &assets.EthereumAssets{})
func Router(name string, a AssetsInferface, mappingMethods ...string) *App {
	OpenWalletApp.Handlers.Add(name, a)
	return OpenWalletApp
}