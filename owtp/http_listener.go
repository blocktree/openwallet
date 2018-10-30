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
	"context"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

//owtp监听器
type httpListener struct {
	net.Listener
	handler   PeerHandler
	closed    chan struct{}
	incoming  chan Peer
	laddr     string
	peerstore Peerstore //节点存储器
}

//serve 监听服务
func (l *httpListener) serve() error {

	if l.Listener == nil {
		return errors.New("listener is not setup.")
	}

	defer close(l.closed)
	//http.ListenAndServe(l.laddr, l)
	http.Serve(l.Listener, l)

	return nil
}

//ServeHTTP 实现HTTP服务监听
func (l *httpListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Debug("http url path:", r.URL.Path)

	//创建一个上下文通知，监控节点是否已经关闭
	ctx, cancel := context.WithCancel(context.Background())

	var cnCh <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		cnCh = cn.CloseNotify()
	}

	header := r.Header
	if header == nil {
		log.Debug("header is nil")
		http.Error(w, "Header is nil", 400)
		return
	}

	var err error

	peer, err := NewHTTPClientWithHeader(header, w, r, l.handler, cancel)
	if err != nil {
		log.Debug("NewClient unexpected error:", err)
		http.Error(w, "authorization not passed", 401)
		return
	}

	// Just to make sure.
	//defer peer.Close()

	log.Debug("NewClient successfully")

	select {
	case l.incoming <- peer:
	case <-l.closed:
		//peer.Close()
		return
	case <-cnCh:
		log.Debug("http CloseNotify")
		return
	}

	// wait until conn gets closed, otherwise the handler closes it early
	select {
	case <-ctx.Done(): //收到节点关闭的通知
		//log.Debug("peer 1:", peer.PID(), "closed")
		return
	case <-l.closed:
		//log.Debug("peer 2:", peer.PID(), "closed")
		peer.Close()
	case <-cnCh:
		log.Debug("http CloseNotify")
		return
	}

}

//Accept 接收新节点链接，线程阻塞
func (l *httpListener) Accept() (Peer, error) {
	select {
	case c, ok := <-l.incoming:
		if !ok {
			return nil, fmt.Errorf("listener is closed")
		}
		return c, nil
	case <-l.closed:
		return nil, fmt.Errorf("listener is closed")
	}
}

//ListenAddr 创建OWTP协议通信监听
func HttpListenAddr(addr string, handler PeerHandler) (*httpListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener := httpListener{
		Listener:  l,
		laddr:     addr,
		handler:   handler,
		incoming:  make(chan Peer),
		closed:    make(chan struct{}),
	}

	go listener.serve()

	return &listener, nil
}
