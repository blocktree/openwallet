package qtum

import (
	"testing"
	"encoding/hex"
)

func Test_addressTo32bytesArg(t *testing.T) {
	address := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"

	to32bytesArg, err := AddressToArg(address)
	if err != nil {
		t.Errorf("fail")
	}else {
		t.Logf("success")
	}

	t.Logf("This is to32bytesArg string for you to use: %s\n", hex.EncodeToString(to32bytesArg))
}

func Test_getUnspentByAddress(t *testing.T) {

}