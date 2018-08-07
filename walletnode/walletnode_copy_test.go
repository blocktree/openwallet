package walletnode

import (
	"testing"
)

var (
	src string
	dst string
)

func init() {

	if err := loadConfig(symbol); err != nil {
		return err
	}

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		return err
	}

	// Action within client
	symbol := "bopo"
	cname, err := _GetCName(symbol) // container name
	if err != nil {
		return err
	}

}

func TestCopyFromContainer(t *testing.T) {

	src = "/usr/local/paicode/data/wallet.dat"
	dst = "./bk/wallet.dat"

	if err := CopyFromContainer(c, cname, src, dst); err != nil {
		t.Errorf("Error: %v\n", err)
	}

}

func TestCopyToContainer(t *testing.T) {

	src = "./conf/BOPO.ini"
	dst = "/tmp"

	if err := CopyToContainer(c, cname, src, dst); err != nil {
		t.Errorf("Error: %v\n", err)
	}

}
