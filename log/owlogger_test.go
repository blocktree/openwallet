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

package log

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/blocktree/openwallet/v2/common/file"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

//SetupLog 配置日志
func getNewLogger(prefix string) *OWLogger {
	l := NewOWLogger(prefix)
	logDir := filepath.Join("logs")
	//记录日志
	file.MkdirAll(logDir)
	logFile := filepath.Join("logs", "test.log")
	logConfig := fmt.Sprintf(`{"filename":"%s"}`, logFile)
	//log.Println(logConfig)
	l.SetLogger(logs.AdapterFile, logConfig)
	l.SetLogger(logs.AdapterConsole, logConfig)
	return l
}

func TestMultiLogger(t *testing.T) {
	log_btc := getNewLogger("BTC")
	log_ltc := getNewLogger("LTC")
	log_eth := getNewLogger("ETH")
	log_qtum := getNewLogger("QTUM")
	log_nas := getNewLogger("NAS")

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {

		for i := 0; i < 10000; i++ {
			log_btc.Info("BTC log out:", i)
			time.Sleep(800)
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {

		for i := 0; i < 10000; i++ {
			log_ltc.Info("LTC log out:", i)
			time.Sleep(800)
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {

		for i := 0; i < 10000; i++ {
			log_eth.Info("ETH log out:", i)
			time.Sleep(800)
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {

		for i := 0; i < 10000; i++ {
			log_qtum.Info("QTUM log out:", i)
			time.Sleep(800)
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {

		for i := 0; i < 10000; i++ {
			log_nas.Info("NAS log out:", i)
			time.Sleep(800)
		}
		wait.Done()
	}()
	wait.Wait()
}
