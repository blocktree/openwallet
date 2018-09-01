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

package owtp

import (
	"encoding/json"
	"testing"
	"time"
)

var (
	mqtestUrl = "amqp://aielves:aielves12301230@39.108.64.191:36672/"
)

func init() {

}

func TestMQDial(t *testing.T) {

	client, err := MQDial("hello", mqtestUrl, nil)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	defer client.Close()

	client.OpenPipe()
}

func TestMQEncodeDataPacket(t *testing.T) {

	//封装数据包
	packet := DataPacket{
		Method:    "subscribe",
		Req:       WSRequest,
		Nonce:     1,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"sss": "sdfsdf",
		},
	}

	respBytes, err := json.Marshal(packet)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	t.Logf("Send: %s \n", string(respBytes))

}
