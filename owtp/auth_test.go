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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/log"
	"github.com/mr-tron/base58/base58"
	"testing"
)

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
	auth, err := NewOWTPAuthWithCertificate(cert, false)
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

func TestBase58(t *testing.T) {
	encode := "6f1yErtapWhSGwQmXQUeM7dTnaGgQb2m8hv13hXYsPW6q5EvkZXWdPa3HjYo4yYV6YocGZerwth6qyNdAaT9MrV7B2wJZeugTn5CrvbuWC4cqQkKXWWhXa76LNfK8fDcDjhJvKMeGEce2R2j2fadQHXyH5QbSieWUuhxMUUi8dRY7wP"
	pub, _ := base58.Decode(encode)
	log.Info("pub decode:", hex.EncodeToString(pub))

	pubbs, _ := hex.DecodeString("038EA1AF7C58F2FC1BF1BF7C04075513D9B6FFF4517D84F53B2D04E179D1054810F4C27C24AB9A54F962E2892A067B4E83BEF7C53A145C2EDC6BFB83852EDD77")

	log.Info("pub encode:", base58.Encode(pubbs))
}


func TestEncryptData(t *testing.T) {
	nonce := 1181076977457565696
	log.Infof("nonce: %d", nonce)
	b58 := "CKv1vLjKGCp9uSGhLmXFgn2TLRdvEiiqRCfvmW6qwWJQ"
	key, _ := base58.Decode(b58)
	log.Infof("key: %s", hex.EncodeToString(key))
	//把nonce作为salt
	nonceBit := []byte(common.NewString(nonce).String())
	h := hmac.New(sha256.New, nonceBit)
	h.Write(key)
	md := h.Sum(nil)
	newkey := md
	log.Infof("key: %s", hex.EncodeToString(newkey))
}