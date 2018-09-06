package environment

import (
	"os"
)

func init() {
	//fmt.Println("tech init..")
	os.Chdir("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/testethereum/")
	os.Setenv("DEBUG_ENABLED", "1")
}
