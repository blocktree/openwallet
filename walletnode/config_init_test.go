package walletnode

import (
	"fmt"
	"testing"
)

func TestInitConfig(t *testing.T) {
	if err := initConfig("bopo"); err != nil {
		t.Errorf("\tTestInitConfig: %+v\n", err)
	} else {
		fmt.Println("\tTestInitConfig: Success!")
	}
}
