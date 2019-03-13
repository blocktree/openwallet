package tech

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"

	"github.com/blocktree/openwallet/log"
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
	slice := make([]int, 8, 20) //{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println("slice:", slice, " cap:", cap(slice), " len:", len(slice))
	slice = slice[0:8:15]
	fmt.Println("slice:", slice, " cap:", cap(slice), " len:", len(slice))
	slice = slice[0:6:10]
	fmt.Println("slice:", slice, " cap:", cap(slice), " len:", len(slice))
}

func TestSlice2() {
	slice := []int{1, 2, 3, 4, 5, 6}
	newSlice := slice[3:4:4]
	fmt.Println("slice:", slice)
	fmt.Println("newslice:", newSlice, " cap:", cap(newSlice))
	_ = append(newSlice, 7)
	fmt.Println("after append slice:", slice)
	fmt.Println("after append new slice:", newSlice, " cap:", cap(newSlice))
}

func TestStringAndSlice() {
	str := "1234567890"
	slice := []byte(str)
	slice[0] = 'x'
	fmt.Println("str:", str)
	fmt.Println("slice:", string(slice))
}

func TestMap() {
	type t struct {
		a string
		b int
	}
	t1 := t{a: "haha", b: 0}
	t2 := t{a: "xixi", b: 1}
	tmap := map[string]t{
		"haha": t1,
		"xixi": t2,
	}

	fmt.Println(tmap)
	for k, v := range tmap {
		if k == "xixi" {
			v.b++
			//tmap[k].b++ will report a compile error
			tmap[k] = v
		}
	}

	fmt.Println(tmap)
}

func TestWalletLog() {
	log.Debugf("debug in TestWalletLog.")
	log.Debugf("debug testlog [%v] ", "testwallet")

	log.Infof("info in TestWalletLog.")
	log.Infof("info testlog [%v] ", "testwallet")

	log.Warningf("warning in TestWalletLog.")
	log.Warningf("warning testlog [%v] ", "testwallet")

	log.Errorf("error in TestWalletLog.")
	log.Errorf("error testlog [%v] ", "testwallet")
}
