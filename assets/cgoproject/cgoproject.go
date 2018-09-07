package main

/*
#include "testc.h"
extern void HollyShit();

struct TestStruct
{
	int A;
	int B;
};

*/
import "C"
import (
	"fmt"
)

//export HollyShit
func HollyShit() {
	fmt.Println("Holly shit, mother fucker.")
}

func PrintText(text string) {
	fmt.Println(text)
}

func main() {
	//fmt.Println("Hello World...")
	C.MyPrint(C.CString("hello cgo...\n"))
	//	C.HollyShit()
	var testStruct C.struct_TestStruct
	testStruct.A = 1
	testStruct.B = 2
	fmt.Println("print struct, A:", testStruct.A, " B:", testStruct.B)
}
