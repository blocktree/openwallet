package walletnode

import (
	"testing"
)

var (
	src string
	dst string
)

func TestCopyFromContainer(t *testing.T) {

	src = "/usr/local/paicode/data/wallet.dat"
	dst = "./bk/wallet.dat"

	wn := WalletnodeManager{}
	if err := wn.CopyFromContainer("bopo", src, dst); err != nil {
		t.Errorf("Error: %v\n", err)
	}

}

func TestCopyToContainer(t *testing.T) {

	src = "./conf/BOPO.ini"
	dst = "/tmp"

	wn := WalletnodeManager{}
	if err := wn.CopyToContainer("bopo", src, dst); err != nil {
		t.Errorf("Error: %v\n", err)
	}

}
