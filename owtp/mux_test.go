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

package owtp

import (
	"testing"
	"time"
)

func TestRequestReplayAttack(t *testing.T) {

	var (
		err    error
		status uint64
		msg    string
	)

	mux := NewServeMux(120)
	mux.requestNonceLimit = 3 * time.Second

	status, msg = mux.checkNonceReplayReason("1", 1)

	t.Logf("status = %d, msg = %s", status, msg)

	peer := &HTTPClient{
		pid: "1",
	}

	err = mux.AddRequest(peer, 1, time.Now().Unix(), "h", nil, nil, false)
	if err != nil {
		t.Errorf("RequestReplayAttack failed unexpected error: %v", err)
		return
	}

	status, msg = mux.checkNonceReplayReason("1", 1)
	t.Logf("status = %d, msg = %s", status, msg)

	time.Sleep(5 * time.Second)

	status, msg = mux.checkNonceReplayReason("1", 1)
	t.Logf("status = %d, msg = %s", status, msg)
}
