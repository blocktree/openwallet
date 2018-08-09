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

package utils

import (
	"path/filepath"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/log"
)

//SetupLog 配置日志
func SetupLog(logDir, logFile string, debug bool) {

	//记录日志
	logLevel := log.LevelInformational
	if debug {
		logLevel = log.LevelDebug
	}

	if len(logDir) > 0 {
		file.MkdirAll(logDir)
		logFile := filepath.Join(logDir, logFile)
		logConfig := fmt.Sprintf(`{"filename":"%s","level":%d,"daily":true,"maxdays":7,"maxsize":0}`, logFile, logLevel)
		//log.Println(logConfig)
		log.SetLogger(logs.AdapterFile, logConfig)
		log.SetLogger(logs.AdapterConsole, logConfig)
	} else {
		log.SetLevel(logLevel)
	}
}
