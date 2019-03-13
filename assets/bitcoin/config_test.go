/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitcoin

import (
	"github.com/astaxie/beego/config"
	"path/filepath"
	"testing"
)

func TestWalletManager_InitAssetsConfig(t *testing.T) {
	c, err := tw.InitAssetsConfig()
	if err != nil {
		t.Errorf("InitAssetsConfig failed unexpected error: %v\n", err)
		return
	}
	t.Logf("rpcServerType: %s", c.String("rpcServerType"))
}

func TestWalletManager_LoadAssetsConfig(t *testing.T) {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(tw.Config.configFilePath, tw.Config.configFileName)

	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return
	}

	err = tw.LoadAssetsConfig(c)
	if err != nil {
		t.Errorf("InitAssetsConfig failed unexpected error: %v\n", err)
		return
	}
	t.Logf("ServerAPI: %s", tw.Config.ServerAPI)
}