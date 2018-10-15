package qtum

import (
	"testing"
	"encoding/hex"
	"strconv"
)


func Test_addressTo32bytesArg(t *testing.T) {
	address := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"

	to32bytesArg, err := AddressToArg(address)
	if err != nil {
		t.Errorf("To32bytesArg failed unexpected error: %v\n", err)
	}else {
		t.Logf("To32bytesArg success.")
	}

	t.Logf("This is to32bytesArg string for you to use: %s\n", hex.EncodeToString(to32bytesArg))
}


func Test_getUnspentByAddress(t *testing.T) {
	contractAddress := "91a6081095ef860d28874c9db613e7a4107b0281"
	address := "qVT4jAoQDJ6E4FbjW1HPcwgXuF2ZdM2CAP"

	QRC20Utox, err := tw.GetUnspentByAddress(contractAddress, address)
	if err != nil {
		t.Errorf("GetUnspentByAddress failed unexpected error: %v\n", err)
	}

	Unspent, err := strconv.ParseInt(QRC20Utox.Output,16,64)
	if err != nil {
		t.Errorf("strconv.ParseInt failed unexpected error: %v\n", err)
	}else {
		t.Logf("QRC20Unspent %s: %s = %d\n", QRC20Utox.Address, address, Unspent)
	}
}

