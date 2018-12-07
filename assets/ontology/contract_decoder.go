package ontology

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/blocktree/OpenWallet/log"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type AddrBalance struct {
	Address    string
	ONTBalance *big.Int
	ONGBalance *big.Int
	ONGUnbound *big.Int
	index      int
}

func newAddrBalance(data []string) *AddrBalance {
	// getbalance
	/*
		{
			"Action":"getbalance",
			"Desc":"SUCCESS",
			"Error":0,
			"Result":
			{
				"ont":"899999999",
				"ong":"62785074999999999"
			},
				"Version":"1.0.0"
			}
	*/
	//getunboundong
	/*
		{
			"Action":"getunboundong",
			"Desc":"SUCCESS",
			"Error":0,
			"Result":"1575449000000000","
			Version":"1.0.0"
		}
	*/

	ontBalance, err := strconv.ParseInt(gjson.Get(data[0], "ont").String(), 10, 64)
	if err != nil {
		return nil
	}

	ongBalance, err := strconv.ParseInt(gjson.Get(data[0], "ong").String(), 10, 64)
	if err != nil {
		return nil
	}

	ongUnbound, err := strconv.ParseInt(data[1][1:len(data[1])-1], 10, 64)
	if err != nil {
		return nil
	}
	return &AddrBalance{
		ONTBalance: big.NewInt(ontBalance),
		ONGBalance: big.NewInt(ongBalance),
		ONGUnbound: big.NewInt(ongUnbound),
	}
}

func convertFlostStringToBigInt(amount string) (*big.Int, error) {
	vDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert from string to decimal failed, err=", err)
		return nil, err
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < 9; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}
	d, _ := decimal.NewFromString(decimalInt.String())
	vDecimal = vDecimal.Mul(d)
	rst := new(big.Int)
	if _, valid := rst.SetString(vDecimal.String(), 10); !valid {
		log.Error("conver to big.int failed")
		return nil, errors.New("conver to big.int failed")
	}
	return rst, nil
}

func convertBigIntToFloatDecimal(amount string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert string to deciaml failed, err=", err)
		return d, err
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < 9; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}

	w, _ := decimal.NewFromString(decimalInt.String())
	d = d.Div(w)
	return d, nil
}

func convertIntStringToBigInt(amount string) (*big.Int, error) {
	vInt64, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		log.Error("convert from string to int failed, err=", err)
		return nil, err
	}

	return big.NewInt(vInt64), nil
}
