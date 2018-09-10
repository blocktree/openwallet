package tech

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
)

//big.Int不能复制
func TestBigInt() {
	num1 := big.NewInt(100)
	fmt.Println("num1:", num1)
	num2 := new(big.Int)
	num2.Sub(num1, big.NewInt(50))

	fmt.Println("num1:", num1, " num2:", num2)
	fmt.Println("num1:", num1, " num2:", num2)
	num2 = num2.Add(num1, big.NewInt(200))
	fmt.Println("num1:", num1, " num2:", num2)
}

func TestDiffer() {
	func() {
		defer func() {
			fmt.Println("in defer.")
		}()
	}()
	fmt.Println("out of parenthesis ")
}

func Log(args ...interface{}) {
	pc, filename, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Caller not okay.")
	} else {
		funcName := runtime.FuncForPC(pc).Name()
		tokens := strings.Split(funcName, ".")
		fmt.Println("Func Name=" + tokens[len(tokens)-1])
		tokens = strings.Split(filename, "/")
		fmt.Printf("file: %s    line=%d\n", tokens[len(tokens)-1], line)
	}
}

func TestGetFuncAndFileName() {
	Log()
}

func TestSlice() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println(slice[0:8:8])
	fmt.Println(slice[0:6:8])
}
