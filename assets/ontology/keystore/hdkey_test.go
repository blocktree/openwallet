package keystore

import "testing"

func TestAddressDecoder_GenerateKeyPairWithHdKey(t *testing.T) {
	_,_, err := GenerateKeyPairWithHdKey(1)
	if err != nil{
		t.Errorf("generate key pair with hdkey failed, err=%v", err)
	}
}
