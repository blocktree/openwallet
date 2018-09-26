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

package tron

import (
	"testing"
)

var (
	pTxRaw    = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e"
	pTxSigned = "0a7d0a02d0762208239cf236e19b41cf40e887c8a7e12c5a66080112620a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412310a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518904e124123218536461fff784bd32e206a4c1a7adc05b455aa37cd624724cb3a2826119d434317d121bda9b5352bf0aaa61326471c5167d14376a96a466317d041696c6801"
)

func TestCreateTransactionRef(t *testing.T) {

	var to_address, owner_address string
	var amount int64

	// predictTxRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d"
	// to_address, owner_address, amount = "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf", "TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax", uint64(1)
	to_address, owner_address, amount = TOADDRESS, OWNERADDRESS, AMOUNT

	if r, err := tw.CreateTransactionRef(to_address, owner_address, amount); err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("TestCreateTransaction return: \n\t%+v\n", r)
		// fmt.Println("P_Raw: ", predictTxRaw)
		// fmt.Println("R_Raw: ", r)
	}
}

func TestGetTransactoinSignRef(t *testing.T) {

	// predictTxRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d"
	// txRawhex := "0a7d0a02761f220840379a940e63492e4090a4bc86e12c5a66080212621260306131353239393966656530326531656530313138396263343161363865393036396237393139656632616438323132313532396236633161626639666233316339303737646662336332353436396536653934336666626661376131383031"
	pk := []byte(PRIVATEKEY)
	txRaw := "0a5f0a02b82818a8f0960122082c12208785a1741440e3c4d4dff8d2fcab155a360802123212300a1541e11973395042ba3c0b52b4cdf4e15ea77818f27512154199fee02e1ee01189bc41a68e9069b7919ef2ad82180170fbbcd4dff8d2fcab15"

	if r, err := tw.GetTransactionSignRef(txRaw, pk); err != nil {
		t.Errorf("GetTransactionSignRef failed: %v\n", err)
	} else {
		// if strings.Join(r[:], "") != RAW_expect {
		// 	t.Errorf("TestCreateTransaction return invalid RAW!")
		// }
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", r)
	}
}

func TestValidSignedTransactionRef(t *testing.T) {
	// UnsignedTX
	// txRaw := "0a7e0a02761f220840379a940e63492e4090a4bc86e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d"
	// UnsignedTX
	// txRaw := "0a7540e8075a6608021262126030613135323939396665653032653165653031313839626334316136386539303639623739313965663261643832313231353239623663316162663966623331633930373764666233633235343639653665393433666662666137613138303170b9f8d2f9a29be9ab15"
	// SignedTX from Client
	// txRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d1240720cb95a8eba333fa5de71d823a285b13ecde4768d02f312fce2bb2455018e5470a99b6aeaf6e35e67595747ba14a1086c887504a19ce12d2a86f5b1a139631e"
	// txRaw := "0a7e0a021031220816b0c1a29ce3387c40e099ad83e02c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a12154199fee02e1ee01189bc41a68e9069b7919ef2ad8218c0843d124067d47e7f668e9460bf34972671b89ea2f246d833ae87593be6be8e0545db45bfacdab7b7a1d9ef99d0e4ed859578c4c5b1190199beaee85dca4f6624aaaf4b96"

	var txSignedRaw string
	txSignedRaw = TXSIGNED
	// txSignedRaw = "0a5f0a02b82818a8f0960122082c12208785a1741440b6c6898cba97fcab155a360802123212300a1541e11973395042ba3c0b52b4cdf4e15ea77818f27512154199fee02e1ee01189bc41a68e9069b7919ef2ad82180170cebe898cba97fcab1512413a964a02b0ac98aba92f87e66fe735cf305fb19f4c1187353bf2a54a8faed5290dae5639e5edc04337bd5f4af6c33ee2acbe403c316b3f2dbd4710a344090dbe00"
	txSignedRaw = "0a5f0a02b82818a8f0960122082c12208785a1741440b6c6898cba97fcab155a360802123212300a1541e11973395042ba3c0b52b4cdf4e15ea77818f27512154199fee02e1ee01189bc41a68e9069b7919ef2ad82180170cebe898cba97fcab151241a6190d00501c9a999c746f9560d048827f2e0bc36902c1c35654398f0961d58348881cc561b823707c7f236d106fe53327c7c335e28bf90cd0baf0165039ce0200"
	txSignedRaw = "0a5f0a02b82818a8f0960122082c12208785a1741440e3c4d4dff8d2fcab155a360802123212300a1541e11973395042ba3c0b52b4cdf4e15ea77818f27512154199fee02e1ee01189bc41a68e9069b7919ef2ad82180170fbbcd4dff8d2fcab1512417113348120d01fac161eee8e2c929ad7f88432ac4dbac60f87d7f188edadc641191a76faa77c86055a8c08af81ecc6b1d020aea7079ff829a78546b3ff4d2b5001"
	if err := tw.ValidSignedTransactionRef(txSignedRaw); err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	} else {
		t.Logf("GetTransactionSignRef return: \n\t%+v\n", "Success!")
	}
}

func TestBroadcastTransactionRef(t *testing.T) {
}

func TestSuiteTx(t *testing.T) {

	txRaw, err := tw.CreateTransactionRef(TOADDRESS, OWNERADDRESS, AMOUNT)
	if err != nil {
		t.Errorf("TestCreateTransaction failed: %v\n", err)
	}

	pk := []byte(PRIVATEKEY)
	txSignedRaw, err := tw.GetTransactionSignRef(txRaw, pk)
	if err != nil {
		t.Errorf("GetTransactionSignRef failed: %v\n", err)
	}

	err = tw.ValidSignedTransactionRef(txSignedRaw)
	if err != nil {
		t.Errorf("ValidSignedTransactionRef: %v\n", err)
	}
}
