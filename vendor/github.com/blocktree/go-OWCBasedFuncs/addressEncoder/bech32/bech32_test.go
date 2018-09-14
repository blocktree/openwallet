package bech32

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_bech32_address(t *testing.T) {
	address := "bc1qvgclzqz7smqr6haag9mknpwsjnxtdqkncr64kd"
	ret, err := Decode(address, "qpzry9x8gf2tvdw0s3jn54khce6mua7l")
	if err != nil {
		t.Error("decode error")
	} else {
		fmt.Println(hex.EncodeToString(ret))
	}

	addresschk := Encode("bc", "qpzry9x8gf2tvdw0s3jn54khce6mua7l", ret)
	if addresschk != address {
		t.Error("encode error")
	} else {
		fmt.Println(addresschk)
	}

}
