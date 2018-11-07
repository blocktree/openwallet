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

import "testing"

func TestRandomPrivateKey(t *testing.T) {
	key := RandomPrivateKey()
	t.Logf("key: %v", key)
}

func TestNewOWTPAuthWithCertificate(t *testing.T) {

	key := RandomPrivateKey()
	t.Logf("key: %v", key)

	cert, err := NewCertificate(key, "")
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	//t.Logf("cert: %v", cert)
	auth, err := NewOWTPAuthWithCertificate(cert)
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	//t.Logf("localPublicKey: %v", auth.localPublicKey)
	t.Logf("nodeID: %v", auth.LocalPID())
	t.Logf("auth header: %v", auth.HTTPAuthHeader())
}

func TestCertificate(t *testing.T) {
	cert, err := NewCertificate("CKv1vLjKGCp9uSGhLmXFgn2TLRdvEiiqRCfvmW6qwWJQ", "")
	if err != nil {
		t.Errorf("Dial failed unexpected error: %v", err)
		return
	}
	t.Logf("nodeID: %v", cert.ID())
}
