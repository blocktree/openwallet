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
	ws "github.com/gorilla/websocket"
	"net"
	"net/http"
)

// Default gorilla upgrader
var upgrader = ws.Upgrader{}

type listener struct {
	net.Listener
	auth     Authorization
	closed   chan struct{}
	incoming chan *Client
}

func (l *listener) serve() {
	defer close(l.closed)
	http.Serve(l.Listener, l)
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade websocket", 400)
		return
	}

	ctx, _ := context.WithCancel(context.Background())

	var cnCh <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		cnCh = cn.CloseNotify()
	}

	wscon := &Client{
		ws:      c,
		handler: nil,
		send:    make(chan []byte, MaxMessageSize),
		close:   make(chan struct{}),
	}

	// Just to make sure.
	defer wscon.Close()

	select {
	case l.incoming <- wscon:
	case <-l.closed:
		c.Close()
		return
	case <-cnCh:
		return
	}

	// wait until conn gets closed, otherwise the handler closes it early
	select {
	case <-ctx.Done():
	case <-l.closed:
		c.Close()
		return
	case <-cnCh:
		return
	}
}

func (l *listener) Accept() (*Client, error) {
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
