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
	testUrl = "ws://192.168.30.4:8083/websocket?a=dajosidjaiosjdioajsdioajsdiowhefi&t=1529669837&n=2&s=adisjdiasjdioajsdiojasidjioasjdojasd"
)

func init() {

}

func TestDial(t *testing.T) {

	client, err := Dial(testUrl, nil, nil)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	defer client.Close()

	//发送通道
	go client.writePump()

	//监听消息
	client.readPump()
}

func TestEncodeDataPacket(t *testing.T) {

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
