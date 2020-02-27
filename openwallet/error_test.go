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
	"encoding/json"
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"testing"
)

func TestConvertError(t *testing.T) {
	err := Errorf(ErrAccountNotAddress, "[%s] have not addresses", "zzz")
	log.Infof("err: %v", err)
	owerr := ConvertError(err)
	log.Infof("owerr: %v", owerr)
	jerr, _ := json.Marshal(owerr)
	log.Infof("jerr: %v", string(jerr))

	err2 := fmt.Errorf("[%s] have not addresses", "zzz")
	log.Infof("err2: %v", err2)
	owerr2 := ConvertError(err2)
	log.Infof("owerr2: %v", owerr2)
	jerr2, _ := json.Marshal(owerr2)
	log.Infof("jerr2: %v", string(jerr2))
}
