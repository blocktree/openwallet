package virtualeconomy

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func Test_getBlockHeight(t *testing.T) {
	c := NewClient("http://localhost:9922/", false)

	r, err := c.Call("blocks/height", nil, "GET")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}

	fmt.Println(gjson.Get(r.Raw, "height"))

	fmt.Println(r.Get("height").Uint())
}

func Test_getBlockByHash(t *testing.T) {
	hash := "3Uvb87ukKKwVeU6BFsZ21hy9sSbSd3Rd5QZTWbNop1d3TaY9ZzceJAT54vuY8XXQmw6nDx8ZViPV3cVznAHTtiVE"

	c := NewClient("http://localhost:9922/", false)

	r, err := c.Call("blocks/signature/"+hash, nil, "GET")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_getBlockHash(t *testing.T) {
	c := NewClient("http://localhost:9922/", false)

	height := 1352447
	path := fmt.Sprintf("blocks/at/%d", height)

	fmt.Println(path)

	r, err := c.Call(path, nil, "GET")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}

	fmt.Println(r.Get("signature").String())

}

func Test_getBalance(t *testing.T) {
	c := NewClient("http://47.106.102.2:10026/", false)

	address := "AREkgFxYhyCdtKD9JSSVhuGQomgGcacvQqM"
	path := "addresses/balance/details/" + address

	r, err := c.Call(path, nil, "GET")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}

	fmt.Println(r.Get("available"))

}

func Test_getTransaction(t *testing.T) {
	c := NewClient("http://localhost:9922/", false)
	txid := "FYNmFyt93EhS6XKgDwaJfqUwLgr95nwyPKH7M4Q7hFt7" //"9KBoALfTjvZLJ6CAuJCGyzRA1aWduiNFMvbqTchfBVpF"
	path := "transactions/info/" + txid

	r, err := c.Call(path, nil, "GET")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_convert(t *testing.T) {

	amount := uint64(5000000001)

	amountStr := fmt.Sprintf("%d", amount)

	fmt.Println(amountStr)

	d, _ := decimal.NewFromString(amountStr)

	w, _ := decimal.NewFromString("100000000")

	d = d.Div(w)

	fmt.Println(d.String())

	d = d.Mul(w)

	fmt.Println(d.String())

	r, _ := strconv.ParseInt(d.String(), 10, 64)

	fmt.Println(r)

	fmt.Println(time.Now().UnixNano())
}

func Test_getTransactionByAddresses(t *testing.T) {
	addrs := "ARAA8AnUYa4kWwWkiZTTyztG5C6S9MFTx11"

	c := NewClient("http://localhost:9922/", false)
	result, err := c.getMultiAddrTransactions(0, -1, addrs)

	if err != nil {
		t.Error("get transactions failed!")
	} else {
		for _, tx := range result {
			fmt.Println(tx.TxID)
		}
	}
}
