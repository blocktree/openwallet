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

import "fmt"

// JsonRPCEndpoint 全节点服务的JSON-RPC接口
type JsonRPCEndpoint interface {

	//SendRPCRequest 发起JSON-RPC请求
	//@optional
	SendRPCRequest(method string, request interface{}) ([]byte, error)

	//SupportJsonRPCEndpoint 是否开放客户端直接调用全节点的JSON-RPC方法
	//@optional
	SupportJsonRPCEndpoint() bool
}

type JsonRPCEndpointBase int

//SendRPCRequest 发起JSON-RPC请求
//@optional
func (base *JsonRPCEndpointBase) SendRPCRequest(method string, request interface{}) ([]byte, error) {
	return nil, fmt.Errorf("SendRPCRequest is not implemented")
}

//SupportJsonRPCEndpoint 是否开放客户端直接调用全节点的JSON-RPC方法
//@optional
func (base *JsonRPCEndpointBase) SupportJsonRPCEndpoint() bool {
	return false
}
