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
	"reflect"
	"github.com/astaxie/beego/context/param"
	"sync"
)

//AssetsInfo 资产信息
type AssetsInfo struct {
	name string
	controllerType reflect.Type
	methods        map[string]string
	routerType     int
	methodParams   []*param.MethodParam
}

//AssetsRegister 资产注册器
type AssetsRegister struct {
	routers      map[string]*AssetsInfo
	pool         sync.Pool
}


// NewAssetsRegister returns a new AssetsRegister.
func NewAssetsRegister() *AssetsRegister {
	ar := &AssetsRegister{
		routers:  make(map[string]*AssetsInfo),
	}

	return ar
}

//Add 注册资产
func (p *AssetsRegister) Add(name string, a AssetsInferface) {

	reflectVal := reflect.ValueOf(a)
	t := reflect.Indirect(reflectVal).Type()

	route := &AssetsInfo{}
	route.name = name
	route.controllerType = t

	p.routers[name] = route
}

//FindAssets 寻找资产
func (p *AssetsRegister) FindAssetsInfo(name string) (routerInfo *AssetsInfo, isFind bool) {

	if t, ok := p.routers[name]; ok {
		return t, true
	}
	return nil, false
}