package tech

import (
	"testing"
	"fmt"
)

type enbedObj struct{
	i int
}

func (this *enbedObj)print(){
	fmt.Println("i:", this.i)
}

type outter struct{
	enbedObj
	i int
}

func TestTestBigInt(t *testing.T) {
    a := &outter{}
    a.i = 14
    a.enbedObj.i = 5
	a.print()
    fmt.Println("out i:", a.i)
}
