package tech

import (
	"fmt"
	"strings"

	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/shopspring/decimal"
)

func TestFloat() {
	//i := big.NewInt(1)
	a, _ := decimal.NewFromString("5")
	di := decimal.NewFromFloat(1).Div(decimal.NewFromFloat(1000000000000000000)).Mul(a)
	dii := decimal.NewFromFloatWithExponent(1, -2)
	a, _ = decimal.NewFromString("1000000000000000000")
	//dii := di.Div(decimal.NewFromFloat(big.NewInt(1), 18))
	//j := big
	fmt.Println("di:", di.String())
	fmt.Println("dii:", dii.String())
	fmt.Println("a:", a.String())
	oneEth := "1,000,000,000,000,000,000"
	ETH, _ := decimal.NewFromString(strings.Replace(oneEth, ",", "", -1))
	fmt.Println("eth:", ETH)

	amount, err := ethereum.ConvertEthStringToWei(di.String())
	if err != nil {
		fmt.Println("convert eth string to wei, err=", err)
		return
	}
	fmt.Println("amount:", amount)

	de, err := ethereum.ConverWeiStringToEthDecimal(amount.String())
	if err != nil {
		fmt.Println("conver to eth decimal failed, err=", err)
		return
	}
	fmt.Println("de:", de)
}

func TestConvertEthStringToWei() {
	amount, err := ethereum.ConvertEthStringToWei("123.5678")
	if err != nil {
		fmt.Println("convert eth string to wei, err=", err)
		return
	}
	//fmt.Println("amount:", amount)
	fmt.Println("amount:", "0x"+amount.Text(16))
}

func TestConverWeiStringToEthDecimal() {
	de, err := ethereum.ConverWeiStringToEthDecimal("123567800000000000000")
	if err != nil {
		fmt.Println("conver to eth decimal failed, err=", err)
		return
	}
	fmt.Println("de:", de)
}
